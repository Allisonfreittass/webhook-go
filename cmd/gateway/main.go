package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

type server struct {
	db *pgxpool.Pool
}

func main() {
	ctx := context.Background()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://localhost:5432/webhook_gateway"
	}
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatal(err)
	}
	log.Println("connected to database")

	srv := &server{db: pool}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", healthHandler)

	// rota do webhook
	mux.HandleFunc("POST /webhooks/{source}", srv.ingestHandler)

	addr := ":8080"
	log.Printf("Starting server on %s", addr)

	log.Fatal(http.ListenAndServe(addr, mux))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *server) ingestHandler(w http.ResponseWriter, r *http.Request) {

	source := r.PathValue("source")

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	body, err := io.ReadAll(r.Body)

	if err != nil {
		// limite de range do body
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			http.Error(w, "body too large", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "could not read body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	headers, err := json.Marshal(r.Header)
	if err != nil {
		log.Printf("marshal headers: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		log.Printf("begin transaction: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	defer tx.Rollback(ctx)

	var eventID string
	err = tx.QueryRow(ctx,
		`INSERT INTO events (source, headers, body) VALUES ($1, $2, $3) RETURNING id`, source, headers, body,
	).Scan(&eventID)
	if err != nil {
		log.Printf("insert event: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO deliveries (event_id) VALUES ($1)`, eventID,
	)
	if err != nil {
		log.Printf("insert delivery: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		log.Printf("commit transaction: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	log.Printf("stored event %s from %s (%d bytes)", eventID, source, len(body))
	w.WriteHeader(http.StatusAccepted)
}

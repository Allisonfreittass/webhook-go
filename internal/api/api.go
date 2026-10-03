package api

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/Allisonfreittass/webhook-go/internal/store"
)

type Server struct {
	store *store.Store
}

func NewServer(st *store.Store) *Server {
	return &Server{store: st}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", healthHandler)

	// rota do webhook
	mux.HandleFunc("POST /webhooks/{source}", s.ingestHandler)

	return mux
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Server) ingestHandler(w http.ResponseWriter, r *http.Request) {

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

	eventID, err := s.store.CreateEvent(ctx, source, headers, body)
	if err != nil {
		log.Printf("create event: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	log.Printf("stored event %s from %s (%d bytes)", eventID, source, len(body))
	w.WriteHeader(http.StatusAccepted)
}

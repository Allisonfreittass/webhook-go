package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Allisonfreittass/webhook-go/internal/api"
	"github.com/Allisonfreittass/webhook-go/internal/config"
	"github.com/Allisonfreittass/webhook-go/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()

	cfg := config.Load()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatal(err)
	}
	log.Println("connected to database")
	st := store.New(pool)
	srv := api.NewServer(st)

	server := &http.Server{
		Addr:              cfg.Addr,
		Handler:           srv.Routes(),
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("Starting server on %s", cfg.Addr)

	log.Fatal(server.ListenAndServe())
}

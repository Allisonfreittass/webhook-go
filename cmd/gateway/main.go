package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/Allisonfreittass/webhook-go/internal/api"
	"github.com/Allisonfreittass/webhook-go/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

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
	st := store.New(pool)
	srv := api.NewServer(st)

	addr := ":8080"
	log.Printf("Starting server on %s", addr)

	log.Fatal(http.ListenAndServe(addr, srv.Routes()))
}

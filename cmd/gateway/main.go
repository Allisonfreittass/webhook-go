package main

import (
	"errors"
	"io"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", healthHandler)

	// rota do webhook
	mux.HandleFunc("POST /webhooks/{source}", ingestHandler)

	addr := ":8080"
	log.Printf("Starting server on %s", addr)

	log.Fatal(http.ListenAndServe(addr, mux))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func ingestHandler(w http.ResponseWriter, r *http.Request) {

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

	log.Printf("received webhook: source=%s content-type=%s size=%d", source, r.Header.Get("Content-Type"), len(body))
	w.WriteHeader(http.StatusAccepted)
}

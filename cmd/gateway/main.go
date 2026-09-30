package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", healthHandler)

	addr := ":8080"
	log.Printf("Starting server on %s", addr)

	log.Fatal(http.ListenAndServe(addr, mux))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

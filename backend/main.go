package main

import (
	"log"
	"net/http"
	"os"

	"github.com/favxlaw/mini-brimble/config"
	"github.com/favxlaw/mini-brimble/db"
	"github.com/favxlaw/mini-brimble/handlers"
	"github.com/favxlaw/mini-brimble/pipeline"
)

func main() {
	cfg := config.Load()

	if err := os.MkdirAll("./data", 0755); err != nil {
		log.Fatalf("create data dir: %v", err)
	}

	store, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer store.Close()

	broadcaster := pipeline.NewLogBroadcaster()
	h := handlers.New(store, broadcaster, cfg)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /deployments", h.CreateDeployment)
	mux.HandleFunc("GET /deployments", h.ListDeployments)
	mux.HandleFunc("GET /deployments/{id}", h.GetDeployment)
	mux.HandleFunc("GET /deployments/{id}/logs", h.StreamLogs)

	log.Printf("server running on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, cors(mux)); err != nil {
		log.Fatalf("server: %v", err)
	}
}

// cors wraps the mux to allow requests from the frontend dev server
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

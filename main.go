package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"test-server/controller"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "18080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		host, _ := os.Hostname()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
			"source": controller.CONFIG_ORIGIN,
			"host":   host,
		})
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("dev-ship test server\n"))
	})

	log.Printf("test server listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

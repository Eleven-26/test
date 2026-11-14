package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
)

// envMap 把进程环境变量摊平成 key → value，供 /health 在浏览器里直接查看。
func envMap() map[string]string {
	envs := make(map[string]string)
	for _, kv := range os.Environ() {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			continue
		}
		envs[k] = v
	}
	return envs
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "18080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		host, _ := os.Hostname()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "ok",
			"source": "",
			"host":   host,
			"env":    envMap(),
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

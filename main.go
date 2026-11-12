package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"test-server/controller"
)

type envCheckResult struct {
	Target     string `json:"target"`
	Reachable  bool   `json:"reachable"`
	StatusCode int    `json:"status_code,omitempty"`
	LatencyMs  int64  `json:"latency_ms"`
	Error      string `json:"error,omitempty"`
}

func envList() map[string]string {
	env := make(map[string]string)
	for _, kv := range os.Environ() {
		k, v, ok := strings.Cut(kv, "=")
		if ok {
			env[k] = v
		}
	}
	return env
}

func logEnv() {
	env := envList()
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	log.Printf("environment variables (%d):", len(keys))
	for _, k := range keys {
		log.Printf("  %s=%s", k, env[k])
	}
}

// checkTarget 发起一次 HTTP GET，判断目标地址是否可达。
func checkTarget(target string) envCheckResult {
	res := envCheckResult{Target: target}
	start := time.Now()
	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(target)
	res.LatencyMs = time.Since(start).Milliseconds()
	if err != nil {
		res.Error = err.Error()
		log.Printf("test environment check failed: target=%s latency=%dms error=%v", target, res.LatencyMs, err)
		return res
	}
	defer resp.Body.Close()
	res.StatusCode = resp.StatusCode
	res.Reachable = resp.StatusCode >= 200 && resp.StatusCode < 400
	log.Printf("test environment check: target=%s reachable=%t status=%d latency=%dms",
		target, res.Reachable, res.StatusCode, res.LatencyMs)
	return res
}

func envCheckTarget(port string) string {
	if target := os.Getenv("TEST_ENV_URL"); target != "" {
		return target
	}
	return fmt.Sprintf("http://127.0.0.1:%s/health", port)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "18080"
	}
	target := envCheckTarget(port)

	logEnv()

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
	mux.HandleFunc("/env", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(envList())
	})
	mux.HandleFunc("/env-check", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(checkTarget(target))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("dev-ship test server\n"))
	})

	// 服务监听前端口还没起来，延迟做一次自检，避免启动即失败。
	go func() {
		time.Sleep(500 * time.Millisecond)
		checkTarget(target)
	}()

	log.Printf("test server listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

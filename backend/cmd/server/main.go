package main

import (
	"log"
	"net/http"
	"os"
	"time"

	handlerhttp "github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/handler/http"
	"github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/infrastructure/persistence"
	messageuc "github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/usecase/message"
)

func main() {
	csvPath := envOrDefault("DATA_CSV_PATH", "data/messages.csv")
	addr := envOrDefault("ADDR", ":8080")

	repo := persistence.NewMessageCSVRepository(csvPath)
	uc := messageuc.NewUseCase(repo)
	h := handlerhttp.NewMessageHandler(uc)
	xssH := handlerhttp.NewXSSExperimentHandler()

	mux := http.NewServeMux()
	mux.HandleFunc("/", h.GetRoot)
	// HTML 埋め込み時のエスケープ（XSS 防止）を示す実験・教育用エンドポイント。
	mux.HandleFunc("/experiment/greet", xssH.Greet)

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("listening on %s (csv=%s)", addr, csvPath)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

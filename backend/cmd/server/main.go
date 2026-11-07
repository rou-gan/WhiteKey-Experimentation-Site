package main

import (
	"log"
	"net/http"
	"os"

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

	mux := http.NewServeMux()
	mux.HandleFunc("/", h.GetRoot)

	log.Printf("listening on %s (csv=%s)", addr, csvPath)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

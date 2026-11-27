package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	handlerhttp "github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/handler/http"
	"github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/infrastructure/persistence"
	memberuc "github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/usecase/member"
	messageuc "github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/usecase/message"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	addr := envOrDefault("ADDR", ":8080")

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		log.Fatalf("ping database: %v", err)
	}
	if err := persistence.Migrate(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	repo := persistence.NewMessagePostgresRepository(db)
	uc := messageuc.NewUseCase(repo)
	h := handlerhttp.NewMessageHandler(uc)

	memberRepo := persistence.NewMemberPostgresRepository(db)
	challenges := persistence.NewMemberChallengePostgresStore(db)
	sessions := persistence.NewMemberSessionPostgresStore(db)
	memberUC := memberuc.NewUseCase(memberRepo, challenges, sessions)
	memberH := handlerhttp.NewMemberHandler(memberUC)

	xssH := handlerhttp.NewXSSExperimentHandler()
	cmdH := handlerhttp.NewCmdInjectExperimentHandler()

	mux := http.NewServeMux()
	mux.HandleFunc("/", h.GetRoot)
	mux.HandleFunc("/members", memberH.Register)
	mux.HandleFunc("/members/auth/challenge", memberH.IssueChallenge)
	mux.HandleFunc("/members/auth/verify", memberH.VerifyResponse)
	mux.HandleFunc("/members/me", memberH.DeleteMe)
	// HTML 埋め込み時のエスケープ（XSS 防止）を示す実験・教育用エンドポイント。
	mux.HandleFunc("/experiment/greet", xssH.Greet)
	// OS コマンドインジェクション対策の確認用（検証済み argv で ping）。
	mux.HandleFunc("/experiment/ping", cmdH.Ping)

	server := &http.Server{
		Addr:              addr,
		Handler:           handlerhttp.SecurityHeaders(handlerhttp.CORS(mux)),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("listening on %s", addr)
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

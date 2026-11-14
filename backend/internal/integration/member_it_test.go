package integration_test

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/crypto/argon2"

	handlerhttp "github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/handler/http"
	"github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/infrastructure/persistence"
	memberuc "github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/usecase/member"
)

// EDS-MEM-1 / usecase と同じ Argon2id 最低ライン。
const (
	argonMemory      = 19_456
	argonIterations  = 2
	argonParallelism = 1
	argonKeyLen      = 32
)

type itEnv struct {
	db  *sql.DB
	mux http.Handler
}

func openIT(t *testing.T) *itEnv {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping: %v", err)
	}
	if err := persistence.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	truncateMemberTables(t, db)

	memberRepo := persistence.NewMemberPostgresRepository(db)
	challenges := persistence.NewMemberChallengePostgresStore(db)
	sessions := persistence.NewMemberSessionPostgresStore(db)
	uc := memberuc.NewUseCase(memberRepo, challenges, sessions)
	h := handlerhttp.NewMemberHandler(uc)

	mux := http.NewServeMux()
	mux.HandleFunc("/members", h.Register)
	mux.HandleFunc("/members/auth/challenge", h.IssueChallenge)
	mux.HandleFunc("/members/auth/verify", h.VerifyResponse)
	mux.HandleFunc("/members/me", h.DeleteMe)

	return &itEnv{db: db, mux: mux}
}

func truncateMemberTables(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`
		TRUNCATE member_sessions, member_auth_challenges, members RESTART IDENTITY CASCADE
	`)
	if err != nil {
		t.Fatalf("truncate: %v", err)
	}
}

func uniqueLogin(t *testing.T) string {
	t.Helper()
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		t.Fatalf("rand: %v", err)
	}
	return fmt.Sprintf("mem%02x%02x%02x", b[0], b[1], b[2])
}

func doJSON(t *testing.T, mux http.Handler, method, path string, body any, headers map[string]string) (int, map[string]any, []byte) {
	t.Helper()
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		rdr = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, rdr)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	raw := rec.Body.Bytes()
	var parsed map[string]any
	if len(raw) > 0 && json.Valid(raw) {
		if err := json.Unmarshal(raw, &parsed); err != nil {
			t.Fatalf("unmarshal %q: %v", string(raw), err)
		}
	}
	return rec.Code, parsed, raw
}

func clientResponse(password string, saltB64, challengeB64 string) string {
	salt, err := base64.StdEncoding.DecodeString(saltB64)
	if err != nil {
		panic(err)
	}
	challenge, err := base64.StdEncoding.DecodeString(challengeB64)
	if err != nil {
		panic(err)
	}
	secret := argon2.IDKey([]byte(password), salt, argonIterations, argonMemory, argonParallelism, argonKeyLen)
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write(challenge)
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func countActiveMembers(t *testing.T, db *sql.DB) int {
	t.Helper()
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM members WHERE status = 'active'`).Scan(&n); err != nil {
		t.Fatalf("count active: %v", err)
	}
	return n
}

func memberRowByLogin(t *testing.T, db *sql.DB, loginID string) (status, storedLogin, nickname string, found bool) {
	t.Helper()
	err := db.QueryRow(`
		SELECT status, login_id, nickname FROM members WHERE login_id = $1
	`, loginID).Scan(&status, &storedLogin, &nickname)
	if err == sql.ErrNoRows {
		return "", "", "", false
	}
	if err != nil {
		t.Fatalf("query member: %v", err)
	}
	return status, storedLogin, nickname, true
}

func memberStatusByOriginalID(t *testing.T, db *sql.DB, id string) (status, loginID, nickname string) {
	t.Helper()
	err := db.QueryRow(`
		SELECT status, login_id, nickname FROM members WHERE id = $1
	`, id).Scan(&status, &loginID, &nickname)
	if err != nil {
		t.Fatalf("query by id: %v", err)
	}
	return status, loginID, nickname
}

func findMemberIDByLogin(t *testing.T, db *sql.DB, loginID string) string {
	t.Helper()
	var id string
	if err := db.QueryRow(`SELECT id FROM members WHERE login_id = $1`, loginID).Scan(&id); err != nil {
		t.Fatalf("find id: %v", err)
	}
	return id
}

func assertGenericConflict(t *testing.T, body map[string]any, raw []byte) {
	t.Helper()
	msg, _ := body["message"].(string)
	if msg == "" {
		t.Fatalf("expected generic message, body=%s", string(raw))
	}
	lower := strings.ToLower(msg)
	if strings.Contains(lower, "login") || strings.Contains(lower, "nick") {
		t.Fatalf("conflict message distinguishes field: %q", msg)
	}
}

// シナリオ: IT-MEM-1-S01 / 目的: 登録から認証まで通してセッションを得ること。
func TestIT_MEM_1_S01_RegisterChallengeVerify(t *testing.T) {
	env := openIT(t)
	loginID := uniqueLogin(t)
	password := "Abcd123!"
	nickname := "nick" + loginID

	code, reg, _ := doJSON(t, env.mux, http.MethodPost, "/members", map[string]string{
		"loginId": loginID, "password": password, "nickname": nickname,
	}, nil)
	if code != http.StatusCreated {
		t.Fatalf("register status=%d", code)
	}
	if reg["loginId"] != loginID || reg["nickname"] != nickname {
		t.Fatalf("register body=%#v", reg)
	}
	if _, ok := reg["password"]; ok {
		t.Fatalf("password leaked: %#v", reg)
	}

	code, chal, _ := doJSON(t, env.mux, http.MethodPost, "/members/auth/challenge", map[string]string{
		"loginId": loginID,
	}, nil)
	if code != http.StatusOK {
		t.Fatalf("challenge status=%d", code)
	}
	chalID, _ := chal["challengeId"].(string)
	salt, _ := chal["salt"].(string)
	challenge, _ := chal["challenge"].(string)
	if chalID == "" || salt == "" || challenge == "" {
		t.Fatalf("challenge body=%#v", chal)
	}

	resp := clientResponse(password, salt, challenge)
	code, ver, _ := doJSON(t, env.mux, http.MethodPost, "/members/auth/verify", map[string]string{
		"challengeId": chalID, "response": resp,
	}, nil)
	if code != http.StatusOK {
		t.Fatalf("verify status=%d body=%#v", code, ver)
	}
	if tok, _ := ver["sessionToken"].(string); tok == "" {
		t.Fatalf("empty sessionToken: %#v", ver)
	}

	st, _, _, found := memberRowByLogin(t, env.db, loginID)
	if !found || st != "active" {
		t.Fatalf("db status=%q found=%v", st, found)
	}
}

// シナリオ: IT-MEM-1-S02 / 目的: 削除後に旧パスワードで認証できないこと。
func TestIT_MEM_1_S02_DeleteThenAuthFails(t *testing.T) {
	env := openIT(t)
	loginID := uniqueLogin(t)
	password := "Abcd123!"
	nickname := "nick" + loginID

	if code, _, _ := doJSON(t, env.mux, http.MethodPost, "/members", map[string]string{
		"loginId": loginID, "password": password, "nickname": nickname,
	}, nil); code != http.StatusCreated {
		t.Fatalf("register status=%d", code)
	}
	memberID := findMemberIDByLogin(t, env.db, loginID)

	_, chal, _ := doJSON(t, env.mux, http.MethodPost, "/members/auth/challenge", map[string]string{
		"loginId": loginID,
	}, nil)
	resp := clientResponse(password, chal["salt"].(string), chal["challenge"].(string))
	_, ver, _ := doJSON(t, env.mux, http.MethodPost, "/members/auth/verify", map[string]string{
		"challengeId": chal["challengeId"].(string), "response": resp,
	}, nil)
	token := ver["sessionToken"].(string)

	code, _, raw := doJSON(t, env.mux, http.MethodDelete, "/members/me", nil, map[string]string{
		"Authorization": "Bearer " + token,
	})
	if code != http.StatusNoContent {
		t.Fatalf("delete status=%d body=%s", code, string(raw))
	}

	st, newLogin, newNick := memberStatusByOriginalID(t, env.db, memberID)
	if st != "deleted" {
		t.Fatalf("status=%q, want deleted", st)
	}
	if newLogin == loginID || newNick == nickname {
		t.Fatalf("mask incomplete: login=%q nick=%q", newLogin, newNick)
	}

	code, chal2, _ := doJSON(t, env.mux, http.MethodPost, "/members/auth/challenge", map[string]string{
		"loginId": loginID,
	}, nil)
	if code != http.StatusOK {
		t.Fatalf("post-delete challenge status=%d", code)
	}
	resp2 := clientResponse(password, chal2["salt"].(string), chal2["challenge"].(string))
	code, _, _ = doJSON(t, env.mux, http.MethodPost, "/members/auth/verify", map[string]string{
		"challengeId": chal2["challengeId"].(string), "response": resp2,
	}, nil)
	if code != http.StatusUnauthorized {
		t.Fatalf("verify after delete status=%d, want 401", code)
	}

	code, _, _ = doJSON(t, env.mux, http.MethodDelete, "/members/me", nil, map[string]string{
		"Authorization": "Bearer " + token,
	})
	if code != http.StatusUnauthorized {
		t.Fatalf("old bearer delete status=%d, want 401", code)
	}
}

// シナリオ: IT-MEM-1-S03 / 目的: 未登録 loginId でも挑戦は成功形だが検証は失敗すること。
func TestIT_MEM_1_S03_UnknownChallengeThenVerifyFails(t *testing.T) {
	env := openIT(t)
	loginID := uniqueLogin(t) // 未登録のまま使う

	code, chal, _ := doJSON(t, env.mux, http.MethodPost, "/members/auth/challenge", map[string]string{
		"loginId": loginID,
	}, nil)
	if code != http.StatusOK {
		t.Fatalf("challenge status=%d, want 200 (not 404)", code)
	}
	if chal["challengeId"] == "" || chal["salt"] == "" || chal["challenge"] == "" {
		t.Fatalf("challenge body=%#v", chal)
	}

	resp := clientResponse("Abcd123!", chal["salt"].(string), chal["challenge"].(string))
	code, _, _ = doJSON(t, env.mux, http.MethodPost, "/members/auth/verify", map[string]string{
		"challengeId": chal["challengeId"].(string), "response": resp,
	}, nil)
	if code != http.StatusUnauthorized {
		t.Fatalf("verify status=%d, want 401", code)
	}
}

// シナリオ: IT-MEM-1-S04 / 目的: 既に使われている loginId で再登録できないこと。
func TestIT_MEM_1_S04_DuplicateLoginID(t *testing.T) {
	env := openIT(t)
	loginID := uniqueLogin(t)
	nickA := "nickA" + loginID[3:]
	nickB := "nickB" + loginID[3:]

	if code, _, _ := doJSON(t, env.mux, http.MethodPost, "/members", map[string]string{
		"loginId": loginID, "password": "Abcd123!", "nickname": nickA,
	}, nil); code != http.StatusCreated {
		t.Fatalf("first register status=%d", code)
	}
	before := countActiveMembers(t, env.db)

	code, body, raw := doJSON(t, env.mux, http.MethodPost, "/members", map[string]string{
		"loginId": loginID, "password": "Abcd123!", "nickname": nickB,
	}, nil)
	if code != http.StatusConflict {
		t.Fatalf("second register status=%d body=%s", code, string(raw))
	}
	assertGenericConflict(t, body, raw)

	if after := countActiveMembers(t, env.db); after != before {
		t.Fatalf("active count %d -> %d", before, after)
	}
}

// シナリオ: IT-MEM-1-S05 / 目的: 既に使われている nickname で再登録できないこと。
func TestIT_MEM_1_S05_DuplicateNickname(t *testing.T) {
	env := openIT(t)
	loginA := uniqueLogin(t)
	loginB := uniqueLogin(t)
	nickname := "shared" + loginA[3:]

	if code, _, _ := doJSON(t, env.mux, http.MethodPost, "/members", map[string]string{
		"loginId": loginA, "password": "Abcd123!", "nickname": nickname,
	}, nil); code != http.StatusCreated {
		t.Fatalf("first register status=%d", code)
	}
	before := countActiveMembers(t, env.db)

	code, body, raw := doJSON(t, env.mux, http.MethodPost, "/members", map[string]string{
		"loginId": loginB, "password": "Abcd123!", "nickname": nickname,
	}, nil)
	if code != http.StatusConflict {
		t.Fatalf("second register status=%d body=%s", code, string(raw))
	}
	assertGenericConflict(t, body, raw)

	if after := countActiveMembers(t, env.db); after != before {
		t.Fatalf("active count %d -> %d", before, after)
	}
}

package http_test

import (
	"html"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	handlerhttp "github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/handler/http"
)

// 目的: クエリ name が HTML 応答へ埋め込まれてもエスケープされ、反射型 XSS にならないこと。
func TestXSSExperimentHandler_Greet_EscapesQueryName(t *testing.T) {
	t.Parallel()

	probe := `<script>alert(1)</script>`
	h := handlerhttp.NewXSSExperimentHandler()

	req := httptest.NewRequest(http.MethodGet, "/experiment/greet?name="+url.QueryEscape(probe), nil)
	rec := httptest.NewRecorder()
	h.Greet(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Fatalf("content-type = %q", ct)
	}

	body := rec.Body.String()
	if strings.Contains(body, probe) {
		t.Fatalf("raw probe reflected in HTML body (XSS risk): %q", body)
	}

	want := "<h1>Hello, " + html.EscapeString(probe) + "</h1>"
	if body != want {
		t.Fatalf("body = %q, want %q", body, want)
	}
}

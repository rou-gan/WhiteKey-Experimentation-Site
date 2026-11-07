package http

import (
	"net/http"

	"github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/experiment/xssrender"
)

// XSSExperimentHandler exposes an intentional reflected-XSS surface for
// security-gate experiments. Do not merge to main without remediation.
type XSSExperimentHandler struct{}

func NewXSSExperimentHandler() *XSSExperimentHandler {
	return &XSSExperimentHandler{}
}

// Greet reflects ?name= into HTML via xssrender.RenderGreeting.
func (h *XSSExperimentHandler) Greet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Query().Get("name")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(xssrender.RenderGreeting(name)))
}

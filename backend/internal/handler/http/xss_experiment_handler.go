package http

import (
	"net/http"

	"github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/experiment/xssrender"
)

// XSSExperimentHandler serves HTML greetings used for XSS-escape drills
// and security unit tests. User input is escaped before HTML embedding.
type XSSExperimentHandler struct{}

func NewXSSExperimentHandler() *XSSExperimentHandler {
	return &XSSExperimentHandler{}
}

// Greet embeds ?name= into HTML via xssrender.RenderGreeting (escaped).
func (h *XSSExperimentHandler) Greet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Query().Get("name")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// RenderGreeting HTML-escapes name before embedding (reflected XSS drill).
	_, _ = w.Write([]byte(xssrender.RenderGreeting(name))) // #nosec G705 -- escaped by xssrender.RenderGreeting
}

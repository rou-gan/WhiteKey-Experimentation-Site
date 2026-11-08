package http

import (
	"encoding/json"
	"net/http"

	"github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/experiment/cmdinject"
)

// CmdInjectExperimentHandler serves OS command injection drills.
type CmdInjectExperimentHandler struct{}

func NewCmdInjectExperimentHandler() *CmdInjectExperimentHandler {
	return &CmdInjectExperimentHandler{}
}

type pingExperimentResponse struct {
	Host    string   `json:"host"`
	Args    []string `json:"args"`
	Output  string   `json:"output,omitempty"`
	Warning string   `json:"warning"`
}

// Ping builds (and runs) a ping command from ?host=.
// INTENTIONAL: uses shell concatenation via cmdinject.PingArgs.
func (h *CmdInjectExperimentHandler) Ping(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	host := r.URL.Query().Get("host")
	args, err := cmdinject.PingArgs(host)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	out, runErr := cmdinject.RunPing(host)
	resp := pingExperimentResponse{
		Host:    host,
		Args:    args,
		Warning: "experiment endpoint: intentional OS command injection sink; do not expose in production",
	}
	if runErr != nil {
		resp.Output = string(out) + "\n" + runErr.Error()
	} else {
		resp.Output = string(out)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(resp)
}

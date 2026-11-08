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
// host は検証され、シェルを介さず argv で実行する。
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
		Warning: "experiment endpoint: educational ping with validated argv; do not expose broadly in production",
	}
	if runErr != nil {
		// 実行失敗時も内部エラー詳細はクライアントへ出さない。
		resp.Output = string(out)
		if resp.Output == "" {
			resp.Output = "ping failed"
		}
	} else {
		resp.Output = string(out)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(resp)
}

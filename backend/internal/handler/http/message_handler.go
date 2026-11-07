package http

import (
	"encoding/json"
	"net/http"

	messageuc "github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/usecase/message"
)

type MessageHandler struct {
	uc *messageuc.UseCase
}

func NewMessageHandler(uc *messageuc.UseCase) *MessageHandler {
	return &MessageHandler{uc: uc}
}

type messageResponse struct {
	Message string `json:"message"`
}

func (h *MessageHandler) GetRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	msg, err := h.uc.GetByID(r.Context(), "1")
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(messageResponse{Message: msg.Content()})
}

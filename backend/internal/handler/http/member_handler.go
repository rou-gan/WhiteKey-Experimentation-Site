package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	domain "github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/domain/member"
	memberuc "github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/usecase/member"
)

// memberService is the usecase surface the HTTP layer maps to status/JSON.
type memberService interface {
	Register(ctx context.Context, loginID, password, nickname string) (*memberuc.RegisterResult, error)
	IssueChallenge(ctx context.Context, loginID string) (*memberuc.ChallengeResult, error)
	VerifyResponse(ctx context.Context, challengeID, responseB64 string) (*memberuc.VerifyResult, error)
	DeleteMe(ctx context.Context, sessionToken string) error
}

// MemberHandler maps member HTTP contracts to the usecase.
type MemberHandler struct {
	uc memberService
}

// NewMemberHandler wires the member HTTP handlers.
func NewMemberHandler(uc memberService) *MemberHandler {
	return &MemberHandler{uc: uc}
}

type genericErrorBody struct {
	Message string `json:"message"`
}

type registerRequest struct {
	LoginID  string `json:"loginId"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}

type registerResponse struct {
	LoginID  string `json:"loginId"`
	Nickname string `json:"nickname"`
}

type challengeRequest struct {
	LoginID string `json:"loginId"`
}

type challengeResponse struct {
	ChallengeID string `json:"challengeId"`
	Salt        string `json:"salt"`
	Challenge   string `json:"challenge"`
}

type verifyRequest struct {
	ChallengeID string `json:"challengeId"`
	Response    string `json:"response"`
}

type verifyResponse struct {
	SessionToken string `json:"sessionToken"`
	Nickname     string `json:"nickname"`
	ExpiresAt    string `json:"expiresAt"`
}

// Register handles POST /members.
func (h *MemberHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeGenericJSON(w, http.StatusBadRequest, "request failed")
		return
	}

	res, err := h.uc.Register(r.Context(), req.LoginID, req.Password, req.Nickname)
	if err != nil {
		switch {
		case errors.Is(err, memberuc.ErrConflict):
			writeGenericJSON(w, http.StatusConflict, "conflict")
		case errors.Is(err, domain.ErrInvalidLoginID),
			errors.Is(err, domain.ErrInvalidPassword),
			errors.Is(err, domain.ErrInvalidNickname):
			writeGenericJSON(w, http.StatusBadRequest, "request failed")
		default:
			writeGenericJSON(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusCreated, registerResponse{
		LoginID:  res.LoginID,
		Nickname: res.Nickname,
	})
}

// IssueChallenge handles POST /members/auth/challenge.
func (h *MemberHandler) IssueChallenge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req challengeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeGenericJSON(w, http.StatusBadRequest, "request failed")
		return
	}

	res, err := h.uc.IssueChallenge(r.Context(), req.LoginID)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidLoginID) {
			writeGenericJSON(w, http.StatusBadRequest, "request failed")
			return
		}
		writeGenericJSON(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, challengeResponse{
		ChallengeID: res.ChallengeID,
		Salt:        res.Salt,
		Challenge:   res.Challenge,
	})
}

// VerifyResponse handles POST /members/auth/verify.
func (h *MemberHandler) VerifyResponse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req verifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeGenericJSON(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	res, err := h.uc.VerifyResponse(r.Context(), req.ChallengeID, req.Response)
	if err != nil {
		if errors.Is(err, memberuc.ErrAuthFailed) {
			writeGenericJSON(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		writeGenericJSON(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, verifyResponse{
		SessionToken: res.SessionToken,
		Nickname:     res.Nickname,
		ExpiresAt:    res.ExpiresAt.UTC().Format(time.RFC3339),
	})
}

// DeleteMe handles DELETE /members/me.
func (h *MemberHandler) DeleteMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token, ok := bearerToken(r.Header.Get("Authorization"))
	if !ok {
		writeGenericJSON(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.uc.DeleteMe(r.Context(), token); err != nil {
		switch {
		case errors.Is(err, memberuc.ErrUnauthorized):
			writeGenericJSON(w, http.StatusUnauthorized, "unauthorized")
		case errors.Is(err, memberuc.ErrNotFound):
			writeGenericJSON(w, http.StatusNotFound, "not found")
		default:
			writeGenericJSON(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func bearerToken(header string) (string, bool) {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	if token == "" {
		return "", false
	}
	return token, true
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeGenericJSON(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, genericErrorBody{Message: message})
}

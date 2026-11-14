package member

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/argon2"
)

// EDS-MEM-1 付録 A 最低ライン。
const (
	argonMemory      = 19_456
	argonIterations  = 2
	argonParallelism = 1
	argonKeyLen      = 32
	saltLen          = 16
	challengeLen     = 32
	sessionTokenLen  = 32
)

func deriveSecret(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, argonIterations, argonMemory, argonParallelism, argonKeyLen)
}

func newSalt() ([]byte, error) {
	b := make([]byte, saltLen)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("salt: %w", err)
	}
	return b, nil
}

func newChallengeBytes() ([]byte, error) {
	b := make([]byte, challengeLen)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("challenge: %w", err)
	}
	return b, nil
}

func newChallengeID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("challenge id: %w", err)
	}
	return "chal_" + hex.EncodeToString(b), nil
}

func computeResponse(secret, challenge []byte) string {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write(challenge)
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func responsesEqual(gotB64 string, secret, challenge []byte) bool {
	want := computeResponse(secret, challenge)
	return hmac.Equal([]byte(gotB64), []byte(want))
}

func newSessionToken() (token string, hash []byte, err error) {
	b := make([]byte, sessionTokenLen)
	if _, err = rand.Read(b); err != nil {
		return "", nil, fmt.Errorf("session token: %w", err)
	}
	token = base64.RawURLEncoding.EncodeToString(b)
	sum := sha256.Sum256([]byte(token))
	return token, sum[:], nil
}

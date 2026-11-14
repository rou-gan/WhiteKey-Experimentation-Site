package member

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
	"unicode"
)

// Status is the lifecycle state of a member.
type Status string

const (
	StatusActive  Status = "active"
	StatusDeleted Status = "deleted"
)

// EDS-MEM-1 §6 パスワード許可記号（空白除く ASCII 印刷可能記号 32 文字）。
const allowedPasswordSymbols = `!"#$%&'()*+,-./:;<=>?@[\]^_` + "`" + `{|}~`

var (
	ErrInvalidLoginID  = errors.New("invalid login id")
	ErrInvalidPassword = errors.New("invalid password")
	ErrInvalidNickname = errors.New("invalid nickname")
	ErrNotActive       = errors.New("member is not active")
)

// Member is a membership entity (no infrastructure dependencies).
type Member struct {
	id               string
	loginID          string
	nickname         string
	status           Status
	salt             []byte
	secret           []byte
	pendingPassword  string // validated plaintext; cleared after SetCredentials
	deletedAt        *time.Time
}

// NewActiveMember validates credentials and creates an active member.
// Plaintext password is held only as pending until SetCredentials.
func NewActiveMember(loginID, password, nickname string) (*Member, error) {
	if err := validateLoginID(loginID); err != nil {
		return nil, err
	}
	if err := validatePassword(password); err != nil {
		return nil, err
	}
	if err := validateNickname(nickname); err != nil {
		return nil, err
	}
	id, err := newID()
	if err != nil {
		return nil, err
	}
	return &Member{
		id:              id,
		loginID:         loginID,
		nickname:        nickname,
		status:          StatusActive,
		pendingPassword: password,
	}, nil
}

// MaskForLogicalDelete replaces identifying fields and clears secrets (EDS-MEM-1).
func (m *Member) MaskForLogicalDelete() error {
	if m == nil {
		return errors.New("member is nil")
	}
	if m.status != StatusActive {
		return ErrNotActive
	}
	maskLogin, err := randomMask("del_")
	if err != nil {
		return err
	}
	maskNick, err := randomMask("dn_")
	if err != nil {
		return err
	}
	m.loginID = maskLogin
	m.nickname = maskNick
	m.salt = make([]byte, 16)
	m.secret = make([]byte, 32)
	if _, err := rand.Read(m.salt); err != nil {
		return fmt.Errorf("mask salt: %w", err)
	}
	if _, err := rand.Read(m.secret); err != nil {
		return fmt.Errorf("mask secret: %w", err)
	}
	m.pendingPassword = ""
	now := time.Now().UTC()
	m.deletedAt = &now
	m.status = StatusDeleted
	return nil
}

// SetCredentials stores KDF outputs and clears the pending plaintext password.
func (m *Member) SetCredentials(salt, secret []byte) error {
	if m == nil {
		return errors.New("member is nil")
	}
	if len(salt) == 0 || len(secret) == 0 {
		return errors.New("salt and secret are required")
	}
	m.salt = append([]byte(nil), salt...)
	m.secret = append([]byte(nil), secret...)
	m.pendingPassword = ""
	return nil
}

func (m *Member) ID() string              { return m.id }
func (m *Member) LoginID() string         { return m.loginID }
func (m *Member) Nickname() string        { return m.nickname }
func (m *Member) Status() Status          { return m.status }
func (m *Member) Salt() []byte            { return append([]byte(nil), m.salt...) }
func (m *Member) Secret() []byte          { return append([]byte(nil), m.secret...) }
func (m *Member) PendingPassword() string { return m.pendingPassword }
func (m *Member) DeletedAt() *time.Time {
	if m.deletedAt == nil {
		return nil
	}
	t := *m.deletedAt
	return &t
}

// ValidateLoginID exposes login ID rules for challenge issuance.
func ValidateLoginID(loginID string) error {
	return validateLoginID(loginID)
}

// Reconstitute rebuilds a member from persistence (no validation).
func Reconstitute(id, loginID, nickname string, status Status, salt, secret []byte, deletedAt *time.Time) *Member {
	var del *time.Time
	if deletedAt != nil {
		t := *deletedAt
		del = &t
	}
	return &Member{
		id:        id,
		loginID:   loginID,
		nickname:  nickname,
		status:    status,
		salt:      append([]byte(nil), salt...),
		secret:    append([]byte(nil), secret...),
		deletedAt: del,
	}
}

func validateLoginID(loginID string) error {
	n := len(loginID)
	if n < 8 || n > 15 {
		return ErrInvalidLoginID
	}
	for _, r := range loginID {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') {
			return ErrInvalidLoginID
		}
	}
	return nil
}

func validatePassword(password string) error {
	n := len(password)
	if n < 8 || n > 15 {
		return ErrInvalidPassword
	}
	var hasLetter, hasDigit, hasSymbol bool
	for i := 0; i < len(password); i++ {
		c := password[i]
		if c > unicode.MaxASCII || c == ' ' {
			return ErrInvalidPassword
		}
		switch {
		case (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z'):
			hasLetter = true
		case c >= '0' && c <= '9':
			hasDigit = true
		default:
			if !isAllowedPasswordSymbol(c) {
				return ErrInvalidPassword
			}
			hasSymbol = true
		}
	}
	if !hasLetter || !hasDigit || !hasSymbol {
		return ErrInvalidPassword
	}
	return nil
}

func isAllowedPasswordSymbol(c byte) bool {
	for i := 0; i < len(allowedPasswordSymbols); i++ {
		if allowedPasswordSymbols[i] == c {
			return true
		}
	}
	return false
}

func validateNickname(nickname string) error {
	runes := []rune(nickname)
	if len(runes) < 1 || len(runes) > 50 {
		return ErrInvalidNickname
	}
	if trimAllSpace(runes) {
		return ErrInvalidNickname
	}
	return nil
}

func trimAllSpace(runes []rune) bool {
	for _, r := range runes {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

func newID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("new id: %w", err)
	}
	// UUID v4 variant bits
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}

func randomMask(prefix string) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("mask: %w", err)
	}
	return prefix + hex.EncodeToString(b), nil
}

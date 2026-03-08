package services

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"time"
)

const defaultAdminSessionTTL = 24 * time.Hour

var (
	ErrAdminUsernameRequired = errors.New("admin username is required")
	ErrAdminPasswordRequired = errors.New("admin password is required")
	ErrAdminPasswordHash     = errors.New("admin password hash is invalid")
)

type AdminAuthService struct {
	username     string
	usernameHash [32]byte
	passwordHash [32]byte
	sessionTTL   time.Duration

	mu       sync.Mutex
	sessions map[string]time.Time
}

func NewAdminAuthService(username, password string, sessionTTL time.Duration) (*AdminAuthService, error) {
	normalizedUsername, err := normalizeAdminUsername(username)
	if err != nil {
		return nil, err
	}
	if err := validateAdminPassword(password); err != nil {
		return nil, err
	}
	if sessionTTL <= 0 {
		sessionTTL = defaultAdminSessionTTL
	}

	return &AdminAuthService{
		username:     normalizedUsername,
		usernameHash: sha256.Sum256([]byte(normalizedUsername)),
		passwordHash: sha256.Sum256([]byte(password)),
		sessionTTL:   sessionTTL,
		sessions:     make(map[string]time.Time),
	}, nil
}

func NewAdminAuthServiceWithPasswordHash(username, passwordHashHex string, sessionTTL time.Duration) (*AdminAuthService, error) {
	normalizedUsername, err := normalizeAdminUsername(username)
	if err != nil {
		return nil, err
	}

	passwordHashHex = strings.TrimSpace(passwordHashHex)
	passwordHashBytes, err := hex.DecodeString(passwordHashHex)
	if err != nil || len(passwordHashBytes) != sha256.Size {
		return nil, ErrAdminPasswordHash
	}
	if sessionTTL <= 0 {
		sessionTTL = defaultAdminSessionTTL
	}

	var passwordHash [sha256.Size]byte
	copy(passwordHash[:], passwordHashBytes)

	return &AdminAuthService{
		username:     normalizedUsername,
		usernameHash: sha256.Sum256([]byte(normalizedUsername)),
		passwordHash: passwordHash,
		sessionTTL:   sessionTTL,
		sessions:     make(map[string]time.Time),
	}, nil
}

func (s *AdminAuthService) AuthenticateCredentials(username, password string) bool {
	usernameHash := sha256.Sum256([]byte(strings.TrimSpace(username)))
	passwordHash := sha256.Sum256([]byte(password))

	s.mu.Lock()
	defer s.mu.Unlock()

	return subtle.ConstantTimeCompare(s.usernameHash[:], usernameHash[:]) == 1 &&
		subtle.ConstantTimeCompare(s.passwordHash[:], passwordHash[:]) == 1
}

func (s *AdminAuthService) CurrentUsername() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.username
}

func (s *AdminAuthService) PasswordHashHex() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return hex.EncodeToString(s.passwordHash[:])
}

func (s *AdminAuthService) SetUsername(username string) error {
	normalizedUsername, err := normalizeAdminUsername(username)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.username == normalizedUsername {
		return nil
	}

	s.username = normalizedUsername
	s.usernameHash = sha256.Sum256([]byte(normalizedUsername))
	s.sessions = make(map[string]time.Time)
	return nil
}

func (s *AdminAuthService) SetPassword(password string) error {
	if err := validateAdminPassword(password); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.passwordHash = sha256.Sum256([]byte(password))
	s.sessions = make(map[string]time.Time)
	return nil
}

func (s *AdminAuthService) IssueToken(now time.Time) (string, time.Time, error) {
	if now.IsZero() {
		now = time.Now()
	}
	token, err := generateAdminSessionToken()
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt := now.Add(s.sessionTTL)

	s.mu.Lock()
	s.cleanupExpiredSessionsLocked(now)
	s.sessions[token] = expiresAt
	s.mu.Unlock()

	return token, expiresAt, nil
}

func (s *AdminAuthService) AuthenticateToken(token string, now time.Time) bool {
	token = strings.TrimSpace(token)
	if token == "" {
		return false
	}
	if now.IsZero() {
		now = time.Now()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	expiresAt, ok := s.sessions[token]
	if !ok {
		return false
	}
	if !expiresAt.After(now) {
		delete(s.sessions, token)
		return false
	}
	return true
}

func (s *AdminAuthService) RevokeToken(token string) {
	token = strings.TrimSpace(token)
	if token == "" {
		return
	}
	s.mu.Lock()
	delete(s.sessions, token)
	s.mu.Unlock()
}

func (s *AdminAuthService) cleanupExpiredSessionsLocked(now time.Time) {
	for token, expiresAt := range s.sessions {
		if !expiresAt.After(now) {
			delete(s.sessions, token)
		}
	}
}

func generateAdminSessionToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func normalizeAdminUsername(username string) (string, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return "", ErrAdminUsernameRequired
	}
	return username, nil
}

func validateAdminPassword(password string) error {
	if strings.TrimSpace(password) == "" {
		return ErrAdminPasswordRequired
	}
	return nil
}

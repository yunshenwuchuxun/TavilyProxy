package services

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"log/slog"
	"strings"
	"sync"

	"tavily-proxy/server/internal/models"

	"gorm.io/gorm"
)

const masterKeySettingKey = "master_key"

var ErrMasterKeyRequired = errors.New("master key is required")

type MasterKeyService struct {
	db     *gorm.DB
	logger *slog.Logger

	mu  sync.RWMutex
	key string
}

func NewMasterKeyService(db *gorm.DB, logger *slog.Logger) *MasterKeyService {
	return &MasterKeyService{db: db, logger: logger}
}

func (s *MasterKeyService) LoadOrCreate(ctx context.Context) error {
	var setting models.Setting
	err := s.db.WithContext(ctx).First(&setting, "key = ?", masterKeySettingKey).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		newKey, err := generateSecret(32)
		if err != nil {
			return err
		}
		setting = models.Setting{Key: masterKeySettingKey, Value: newKey}
		if err := s.db.WithContext(ctx).Create(&setting).Error; err != nil {
			return err
		}
		s.logger.Info("generated master key", "master_key", newKey)
	}

	s.mu.Lock()
	s.key = setting.Value
	s.mu.Unlock()
	return nil
}

func (s *MasterKeyService) Get() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.key
}

func (s *MasterKeyService) Authenticate(token string) bool {
	current := s.Get()
	if current == "" || token == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(current), []byte(token)) == 1
}

func (s *MasterKeyService) Reset(ctx context.Context) (string, error) {
	newKey, err := generateSecret(32)
	if err != nil {
		return "", err
	}
	if err := s.db.WithContext(ctx).Save(&models.Setting{Key: masterKeySettingKey, Value: newKey}).Error; err != nil {
		return "", err
	}

	s.mu.Lock()
	s.key = newKey
	s.mu.Unlock()
	return newKey, nil
}

func (s *MasterKeyService) Set(ctx context.Context, key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return ErrMasterKeyRequired
	}
	if err := s.db.WithContext(ctx).Save(&models.Setting{Key: masterKeySettingKey, Value: key}).Error; err != nil {
		return err
	}

	s.mu.Lock()
	s.key = key
	s.mu.Unlock()
	return nil
}

func generateSecret(bytes int) (string, error) {
	if bytes <= 0 {
		return "", errors.New("invalid secret length")
	}
	buf := make([]byte, bytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}


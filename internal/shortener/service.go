package shortener

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/url"
	"strings"
	"time"

	"URL_shortener/internal/storage"
)

const (
	idLength    = 7
	maxAttempts = 5
)

var (
	ErrInvalidURL = errors.New("invalid url")
)

type Service struct {
	store storage.Storage
}

func NewService(store storage.Storage) *Service {
	return &Service{store: store}
}

func (s *Service) Shorten(ctx context.Context, original string) (string, error) {
	normalized, err := normalizeURL(original)
	if err != nil {
		return "", err
	}

	for i := 0; i < maxAttempts; i++ {
		id := generateID()
		if _, ok, _ := s.store.Get(ctx, id); ok {
			continue
		}
		if err := s.store.Save(ctx, id, normalized); err != nil {
			return "", err
		}
		return id, nil
	}
	return "", errors.New("failed to generate unique id")
}

func (s *Service) Resolve(ctx context.Context, id string) (string, bool, error) {
	return s.store.Get(ctx, id)
}

func normalizeURL(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", ErrInvalidURL
	}
	if !strings.HasPrefix(trimmed, "http://") && !strings.HasPrefix(trimmed, "https://") {
		trimmed = "https://" + trimmed
	}
	u, err := url.Parse(trimmed)
	if err != nil || u.Host == "" {
		return "", ErrInvalidURL
	}
	u.Host = strings.ToLower(u.Host)
	return u.String(), nil
}

func generateID() string {
	b := make([]byte, 5)
	_, _ = rand.Read(b)
	s := base64.RawURLEncoding.EncodeToString(b)
	if len(s) > idLength {
		return s[:idLength]
	}
	return (s + base64.RawURLEncoding.EncodeToString([]byte(time.Now().Format("150405"))))[:idLength]
}

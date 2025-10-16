package storage

import "context"

type Storage interface {
	Save(ctx context.Context, id string, originalURL string) error
	Get(ctx context.Context, id string) (string, bool, error)
}

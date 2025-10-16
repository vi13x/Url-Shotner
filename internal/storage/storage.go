package storage

import "context"

// Storage описывает интерфейс для сохранения и извлечения URL по короткому идентификатору.
type Storage interface {
	Save(ctx context.Context, id string, originalURL string) error
	Get(ctx context.Context, id string) (string, bool, error)
}

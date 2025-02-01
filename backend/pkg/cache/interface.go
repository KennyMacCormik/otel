package cache

import "context"

type Interface interface {
	Get(ctx context.Context, key string) (any, error)
	Set(ctx context.Context, key string, value any) error
	Delete(ctx context.Context, key string) error
	Close(ctx context.Context) error
	GetKeys(ctx context.Context) ([]string, error)
	GetLength() (int, error)
}

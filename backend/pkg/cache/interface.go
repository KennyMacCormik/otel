package cache

import "context"

type CacheInterface interface {
	Get(ctx context.Context, key string) (any, error)
	Set(ctx context.Context, key string, value any) (int, error) // 201 Created; 200 OK; 204 No Content
	Delete(ctx context.Context, key string) error
	Close(ctx context.Context) error
	GetKeys(ctx context.Context) ([]string, error)
	GetLength() (int64, error)
}

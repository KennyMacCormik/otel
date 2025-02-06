package client

import "context"

type BackendClientInterface interface {
	Get(ctx context.Context, key, requestId string) (any, error)
	Set(ctx context.Context, key string, value any, requestId string) (int, error)
	Delete(ctx context.Context, key, requestId string) error
}

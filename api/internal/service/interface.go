package service

import (
	"context"
	"log/slog"
)

type ServiceInterface interface {
	Get(ctx context.Context, key, requestId string, lg *slog.Logger) (any, error)
	Set(ctx context.Context, key string, value any, requestId string, lg *slog.Logger) (int, error)
	Delete(ctx context.Context, key, requestId string, lg *slog.Logger) error
}

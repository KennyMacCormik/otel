package storage

import (
	"github.com/KennyMacCormik/otel/backend/pkg/cache"
	"github.com/KennyMacCormik/otel/backend/pkg/cache/impl/syncMap"
)

func NewStorage() cache.Interface {
	return syncMap.NewSyncMapCache()
}

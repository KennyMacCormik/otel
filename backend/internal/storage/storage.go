package storage

import (
	"github.com/KennyMacCormik/HerdMaster/pkg/cache"
	"github.com/KennyMacCormik/HerdMaster/pkg/cache/impl/syncMap"
)

func NewStorage() cache.Interface {
	return syncMap.NewSyncMapCache()
}

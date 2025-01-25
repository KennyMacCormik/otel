package cache

import (
	"github.com/KennyMacCormik/HerdMaster/pkg/cache"
	"github.com/KennyMacCormik/HerdMaster/pkg/cache/impl/syncMap"
	"github.com/KennyMacCormik/HerdMaster/pkg/cache/wrappers/shardedCache"
	"github.com/KennyMacCormik/HerdMaster/pkg/cache/wrappers/ttlCache"
	"log/slog"
)

func NewCache(lg *slog.Logger) (cache.Interface, error) {
	fn := func() cache.Interface {
		c, _ := ttlCache.NewTtlCache(syncMap.NewSyncMapCache(), ttlCache.WithLogger(lg))
		return c
	}
	c, err := shardedCache.NewShardedCache(fn)
	if err != nil {
		return nil, err
	}
	return c, nil
}

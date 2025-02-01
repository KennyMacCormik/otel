package ttl_cache

import "time"

type TtlCacheEntry struct {
	Value     any
	ExpiresAt time.Time
}

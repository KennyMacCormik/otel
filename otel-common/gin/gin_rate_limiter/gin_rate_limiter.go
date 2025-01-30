package gin_rate_limiter

import (
	"github.com/KennyMacCormik/common/log"
	"github.com/KennyMacCormik/otel/otel-common/gin/gin_request_id"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"strconv"
	"sync/atomic"
)

const (
	defaultMaxRunning = 100
	defaultMaxWait    = 100
	defaultRetryAfter = 1 // in seconds
)

// RateLimiter struct represents rate-limiting gin-specific middleware
type RateLimiter struct {
	running chan struct{}

	maxRunning, maxWaiting, retryAfter                                       int64
	runningRequests, totalRequests, timedOutWaiting, rejectedTooManyRequests atomic.Int64
}

// NewRateLimiter returns initialized RateLimiter
func NewRateLimiter(maxRunning, maxWait, retryAfter int64) *RateLimiter {
	maxRunning, maxWait, retryAfter = normalizeParams(maxRunning, maxWait, retryAfter)

	rm := &RateLimiter{running: make(chan struct{}, maxRunning),
		maxRunning: maxRunning, maxWaiting: maxWait, retryAfter: retryAfter}

	return rm
}

// GetRateLimiter returns gin-compatible rate-limiting middleware
func (rm *RateLimiter) GetRateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		rm.totalRequests.Add(1)
		defer rm.totalRequests.Add(-1)

		requestID, err := gin_request_id.GetRequestIDFromCtx(c)
		if err != nil {
			return
		}

		lg := log.CopyLogger().With("requestID", requestID)

		if rm.rejectIfTooManyRequests(c, lg) {
			return
		}

		// wait or run
		select {
		case rm.running <- struct{}{}:
			rm.runRequest(c)
		case <-c.Request.Context().Done():
			// reject with timeout
			rm.timedOutWaiting.Add(1)

			lg.Warn("request rejected: context canceled",
				"Retry-After", rm.retryAfter,
			)

			c.Header("Retry-After", strconv.Itoa(int(rm.retryAfter)))
			c.AbortWithStatus(http.StatusTooManyRequests)
		}
	}
}

// runRequest executes a request
func (rm *RateLimiter) runRequest(c *gin.Context) {
	rm.runningRequests.Add(1)
	defer rm.runningRequests.Add(-1)

	defer func() { <-rm.running }()

	c.Next()
}

// rejectIfTooManyRequests if totalRequests exceeds maxWaiting + maxRunning request will be rejected
// as both queues are full
func (rm *RateLimiter) rejectIfTooManyRequests(c *gin.Context, lg *slog.Logger) bool {
	t := rm.totalRequests.Load()
	if t >= rm.maxWaiting+rm.maxRunning {
		rm.rejectedTooManyRequests.Add(1)

		lg.Warn("request rejected: too many requests",
			"totalRequests", t,
			"maxRunning", rm.maxRunning,
			"maxWaiting", rm.maxWaiting,
		)

		c.Header("Retry-After", strconv.Itoa(int(rm.retryAfter)))
		c.AbortWithStatus(http.StatusTooManyRequests)

		return true
	}

	return false
}

// normalizeParams validates maxRunning, maxWaiting, retryAfter
// and replaces them with default values if validation fails
func normalizeParams(maxRunning, maxWaiting, retryAfter int64) (int64, int64, int64) {
	if maxRunning < 1 {
		log.Warn("maxRunning should be > 1: replacing with defaultMaxRunning",
			"maxRunning", maxRunning, "defaultMaxRunning", defaultMaxRunning)
		maxRunning = defaultMaxRunning
	}

	if maxWaiting < 1 {
		log.Warn("maxWaiting should be > 1: replacing with defaultMaxWait",
			"maxRunning", maxRunning, "defaultMaxWait", defaultMaxWait)
		maxWaiting = defaultMaxWait
	}

	if retryAfter < 1 {
		log.Warn("retryAfter should be > 1: replacing with defaultRetryAfter",
			"maxRunning", maxRunning, "defaultRetryAfter", defaultRetryAfter)
		retryAfter = defaultRetryAfter
	}

	return maxRunning, maxWaiting, retryAfter
}

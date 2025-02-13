package gin_rate_limiter

import (
	"log/slog"
	"net/http"
	"strconv"
	"sync/atomic"

	"github.com/KennyMacCormik/common/log"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/KennyMacCormik/otel/backend/pkg/gin/gin_request_id"
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

	metricRunningPlusWaitingRequests prometheus.Gauge
	metricRunningRequests            prometheus.Gauge
	metricRejected                   prometheus.Counter
	metricTimeouts                   prometheus.Counter
	metricTotalRequests              prometheus.Counter
}

// NewRateLimiter returns initialized RateLimiter
func NewRateLimiter(maxRunning, maxWait, retryAfter int64) *RateLimiter {
	maxRunning, maxWait, retryAfter = normalizeParams(maxRunning, maxWait, retryAfter)

	rm := &RateLimiter{running: make(chan struct{}, maxRunning),
		maxRunning: maxRunning, maxWaiting: maxWait, retryAfter: retryAfter}

	rm.metricRunningPlusWaitingRequests = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "rate_limiter_running_plus_waiting_requests",
		Help: "Total number of queued + running requests",
	})
	rm.metricRunningRequests = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "rate_limiter_running_requests",
		Help: "Number of currently running requests",
	})
	rm.metricRejected = promauto.NewCounter(prometheus.CounterOpts{
		Name: "rate_limiter_rejected_requests",
		Help: "Total number of requests rejected due to rate limits",
	})
	rm.metricTimeouts = promauto.NewCounter(prometheus.CounterOpts{
		Name: "rate_limiter_timeout_requests",
		Help: "Total number of requests that timed out waiting",
	})
	rm.metricTotalRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "rate_limiter_total_requests_static",
		Help: "Total number of requests",
	})

	return rm
}

func (rm *RateLimiter) GetRateLimiterMetricsEndpoint() func(*gin.Engine) {
	return func(router *gin.Engine) {
		router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}
}

// GetRateLimiter returns gin-compatible rate-limiting middleware
func (rm *RateLimiter) GetRateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		rm.totalRequests.Add(1)
		rm.metricRunningPlusWaitingRequests.Inc()
		rm.metricTotalRequests.Inc()
		defer func() {
			rm.totalRequests.Add(-1)
			rm.metricRunningPlusWaitingRequests.Dec()
		}()

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
			rm.metricTimeouts.Inc()

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
	rm.metricRunningRequests.Inc()
	defer func() {
		rm.runningRequests.Add(-1)
		rm.metricRunningRequests.Dec()
	}()

	defer func() { <-rm.running }()

	c.Next()
}

// rejectIfTooManyRequests if totalRequests exceeds maxWaiting + maxRunning request will be rejected
// as both queues are full
func (rm *RateLimiter) rejectIfTooManyRequests(c *gin.Context, lg *slog.Logger) bool {
	t := rm.totalRequests.Load()
	if t >= rm.maxWaiting+rm.maxRunning {
		rm.rejectedTooManyRequests.Add(1)
		rm.metricRejected.Inc()

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
			"maxWaiting", maxWaiting, "defaultMaxWait", defaultMaxWait)
		maxWaiting = defaultMaxWait
	}

	if retryAfter < 1 {
		log.Warn("retryAfter should be > 1: replacing with defaultRetryAfter",
			"retryAfter", retryAfter, "defaultRetryAfter", defaultRetryAfter)
		retryAfter = defaultRetryAfter
	}

	return maxRunning, maxWaiting, retryAfter
}

package handlers

import (
	"api/internal/compute"
	"errors"
	"fmt"
	"github.com/KennyMacCormik/HerdMaster/pkg/cache"
	"github.com/KennyMacCormik/HerdMaster/pkg/cache/wrappers/ttlCache"
	"github.com/KennyMacCormik/HerdMaster/pkg/gin/middleware"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"log/slog"
	"net/http"
)

var (
	errMalformedRequest = errorMsg{Err: "malformed request"}
	errInternalServer   = errorMsg{Err: "internal server error"}
)

type body struct {
	Key string `json:"key"`
	Val string `json:"value"`
}

type errorMsg struct {
	Err string `json:"err"`
}

func NewStorageHandlers(comp compute.Interface, lg *slog.Logger) func(*gin.Engine) {
	return func(router *gin.Engine) {
		router.GET("/storage/:key", get(comp, lg))
		router.PUT("/storage", set(comp, lg))
		router.POST("/storage", set(comp, lg))
		router.DELETE("/storage/:key", del(comp, lg))
	}
}

func get(comp compute.Interface, lg *slog.Logger) func(c *gin.Context) {
	return func(c *gin.Context) {
		const (
			traceName = "api.http.get"
			spanName  = "http.get"
		)
		// tracing
		span := startSpan(c, traceName, spanName)
		defer span.End()
		// prep
		requestId, err := middleware.GetRequestIDFromCtx(c)
		newLg := middleware.LogReq(c, requestId, lg, false)
		if err != nil {
			logErrorAndTraceEvent(err, "cannot get request id from context", span, newLg)
		}
		// get key
		key, err := getKey(c)
		if err != nil {
			logErrorAndTraceEvent(err, "failed to get key", span, newLg)
			c.JSON(http.StatusBadRequest, errMalformedRequest)
			return
		}
		span.SetAttributes(attribute.String("decoded key", key))
		newLg.Debug("decoded key", "key", key)
		// invoke request
		val, err := comp.Get(c.Request.Context(), key, requestId)
		if err != nil {
			if errors.Is(err, cache.ErrNotFound) || errors.Is(err, &ttlCache.ErrTimeout{}) {
				logErrorAndTraceEvent(err, "key not found", span, newLg)
				c.JSON(http.StatusNotFound, "not found")
				return
			}
			logErrorAndTraceEvent(err, "error accessing storage", span, newLg)
			c.JSON(http.StatusInternalServerError, errInternalServer)
			return
		}
		// send response
		result := body{Key: key, Val: fmt.Sprintf("%v", val)}
		c.JSON(http.StatusOK, result)
	}
}

func set(comp compute.Interface, lg *slog.Logger) func(c *gin.Context) {
	return func(c *gin.Context) {
		const (
			traceName = "api.http.set"
			spanName  = "http.set"
		)
		// tracing
		span := startSpan(c, traceName, spanName)
		defer span.End()
		// prep
		requestId, err := middleware.GetRequestIDFromCtx(c)
		newLg := middleware.LogReq(c, requestId, lg, false)
		if err != nil {
			logErrorAndTraceEvent(err, "cannot get request id from context", span, newLg)
		}
		b := &body{}
		// get body
		err = c.ShouldBindJSON(&b)
		if err != nil {
			logErrorAndTraceEvent(err, "cannot get body from request", span, newLg)
			c.JSON(http.StatusBadRequest, errMalformedRequest)
			return
		}
		span.SetAttributes(attribute.String("key", b.Key), attribute.String("value", b.Val))
		newLg.Debug("request body", "key", b.Key, "value", b.Val)
		// invoke request
		err = comp.Set(c.Request.Context(), b.Key, b.Val, requestId)
		if err != nil {
			logErrorAndTraceEvent(err, "error accessing storage", span, newLg)
			c.JSON(http.StatusInternalServerError, errInternalServer)
			return
		}
		// send response
		c.JSON(http.StatusOK, "ok")
	}
}

func del(comp compute.Interface, lg *slog.Logger) func(c *gin.Context) {
	return func(c *gin.Context) {
		const (
			traceName = "api.http.set"
			spanName  = "http.set"
		)
		// tracing
		span := startSpan(c, traceName, spanName)
		defer span.End()
		// prep
		requestId, err := middleware.GetRequestIDFromCtx(c)
		newLg := middleware.LogReq(c, requestId, lg, false)
		if err != nil {
			logErrorAndTraceEvent(err, "cannot get request id from context", span, newLg)
		}
		// get key
		key, err := getKey(c)
		if err != nil {
			logErrorAndTraceEvent(err, "failed to get key", span, newLg)
			c.JSON(http.StatusBadRequest, "failed to get key")
			return
		}
		span.SetAttributes(attribute.String("key", key))
		newLg.Debug("decoded key", "key", key)
		// invoke request
		err = comp.Delete(c.Request.Context(), key, requestId)
		if err != nil {
			logErrorAndTraceEvent(err, "error accessing storage", span, newLg)
			c.JSON(http.StatusInternalServerError, errInternalServer)
			return
		}
		// send response
		c.JSON(http.StatusOK, "ok")
	}
}

func startSpan(c *gin.Context, traceName, spanName string) trace.Span {
	tracer := otel.Tracer(traceName)
	ctx, span := tracer.Start(c.Request.Context(), spanName)
	c.Request = c.Request.WithContext(ctx)
	return span
}

func logErrorAndTraceEvent(err error, msg string, span trace.Span, lg *slog.Logger) {
	span.AddEvent(msg, trace.WithAttributes(attribute.String("error", err.Error())))
	lg.Error(msg, "error", err)
}

func getKey(c *gin.Context) (string, error) {
	key := c.Param("key")
	if key == "" {
		return "", fmt.Errorf("no key provided")
	}
	return key, nil
}

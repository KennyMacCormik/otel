package storage

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/KennyMacCormik/common/log"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/KennyMacCormik/otel/backend/pkg/cache"
	"github.com/KennyMacCormik/otel/backend/pkg/gin/gin_request_id"
	cacheErrors "github.com/KennyMacCormik/otel/backend/pkg/models/errors/cache"
	httpModels "github.com/KennyMacCormik/otel/backend/pkg/models/http"
	otelHelpers "github.com/KennyMacCormik/otel/backend/pkg/otel/helpers"
)

type StorageHandler struct {
	st cache.CacheInterface
}

func NewStorageHandler(st cache.CacheInterface) *StorageHandler {
	return &StorageHandler{st: st}
}

func (s *StorageHandler) GetGinStorageHandler() func(*gin.Engine) {
	return func(router *gin.Engine) {
		router.GET("/storage/:key", s.ginGet())
		router.PUT("/storage", s.ginSet())
		router.DELETE("/storage/:key", s.ginDel())
	}
}

func (s *StorageHandler) ginGet() func(c *gin.Context) {
	return func(c *gin.Context) {
		const (
			spanName = "http.get"
		)

		lg, span := getLogAndSpan(c, spanName)
		defer span.End()
		defer lg.Info("request finished")

		key, err := getKey(c)
		if err != nil {
			otelHelpers.SetSpanErr(span, err)
			lg.Error("malformed request", "error", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		val, err := s.st.Get(c.Request.Context(), key)
		if err != nil {
			if errors.Is(err, cacheErrors.ErrNotFound) {
				lg.Warn("key not found", "key", key)
				c.Status(http.StatusNotFound)
				return
			}
			otelHelpers.SetSpanErr(span, err)
			lg.Error("failed to get value", "key", key, "error", err.Error())
			c.Status(http.StatusInternalServerError)
			return
		}

		value := fmt.Sprintf("%v", val)
		result := httpModels.Body{Key: key, Val: value}
		log.Debug("got value", "key", key, "value", value)
		c.JSON(http.StatusOK, result)
	}
}

func (s *StorageHandler) ginSet() func(c *gin.Context) {
	return func(c *gin.Context) {
		const (
			spanName = "http.set"
		)

		lg, span := getLogAndSpan(c, spanName)
		defer span.End()
		defer lg.Info("request finished")

		b := &httpModels.Body{}

		err := c.ShouldBindJSON(&b)
		if err != nil {
			otelHelpers.SetSpanErr(span, err)
			lg.Error("failed read request body", "error", err.Error())
			c.Status(http.StatusBadRequest)
			return
		}

		lg.Info("request key", "key", b.Key)
		lg.Debug("request value", "value", b.Val)

		code, err := s.st.Set(c.Request.Context(), b.Key, b.Val)
		if err != nil {
			otelHelpers.SetSpanErr(span, err)
			lg.Error("failed to set value", "key", b.Key, "value", b.Val, "error", err.Error())
			c.Status(http.StatusInternalServerError)
			return
		}

		c.Status(code)
	}
}

func (s *StorageHandler) ginDel() func(c *gin.Context) {
	return func(c *gin.Context) {
		const (
			spanName = "http.delete"
		)

		lg, span := getLogAndSpan(c, spanName)
		defer span.End()
		defer lg.Info("request finished")

		key, err := getKey(c)
		if err != nil {
			otelHelpers.SetSpanErr(span, err)
			lg.Error("malformed request", "error", err.Error())
			c.Status(http.StatusBadRequest)
			return
		}

		err = s.st.Delete(c.Request.Context(), key)
		if err != nil {
			otelHelpers.SetSpanErr(span, err)
			lg.Error("failed to delete value", "key", key, "error", err.Error())
			c.Status(http.StatusInternalServerError)
			return
		}

		c.Status(http.StatusNoContent)
	}
}

func getKey(c *gin.Context) (string, error) {
	key := c.Param("key")
	if key == "" {
		return "", errors.New("no key provided")
	}

	if !isUrlEncoded(strings.TrimPrefix(c.Request.RequestURI, "/storage/"), key) {
		return "", errors.New("key must be URL-encoded")
	}

	return key, nil
}

func isUrlEncoded(rawKey, ginDecodedKey string) bool {
	rawEscaped := url.QueryEscape(rawKey)

	// if rawEscaped == rawKey,
	// it indicated that rawKey is safe by itself
	if rawEscaped == rawKey {
		return true
	}

	// if decoding results in ginDecodedKey,
	// it indicates that the raw string contained unencoded values
	if s, err := url.QueryUnescape(rawEscaped); err == nil && s == ginDecodedKey {
		return false
	}

	return true
}

func getLogAndSpan(c *gin.Context, spanName string) (*slog.Logger, trace.Span) {
	var lg *slog.Logger

	span := otelHelpers.StartSpanWithGinCtx(c, spanName, spanName)

	defer func() {
		lg.Debug("request received",
			"ClientIP", c.ClientIP(),
			"Proto", c.Request.Proto,
			"Header", c.Request.Header,
			"RemoteAddr", c.Request.RemoteAddr,
			"RequestURI", c.Request.RequestURI,
			"ContentLength", c.Request.ContentLength,
			"Host", c.Request.Host,
		)
		lg.Info("request trace ID", "trace_id", span.SpanContext().TraceID().String())
	}()

	reqId, err := gin_request_id.GetRequestIDFromCtx(c)
	if err != nil {
		span.SetAttributes(attribute.String("request_id", "N/A"))
		log.Error("no request ID in context")
		lg = log.CopyLogger().With("Method", c.Request.Method, "UrlPath", c.Request.URL.Path)
		return lg, span
	}

	span.SetAttributes(attribute.String("request_id", reqId))
	lg = log.CopyLogger().With("request_id", reqId, "Method", c.Request.Method, "UrlPath", c.Request.URL.Path)

	return lg, span
}

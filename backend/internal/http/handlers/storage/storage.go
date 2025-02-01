package storage

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/KennyMacCormik/otel/backend/pkg/cache"
)

var (
	errStatusNotFound       = errorMsg{Err: "not found"}
	errStatusBadRequest     = errorMsg{Err: "malformed request"}
	errStatusInternalServer = errorMsg{Err: "internal server error"}
)

type body struct {
	Key string `json:"key"`
	Val string `json:"value"`
}

type errorMsg struct {
	Err string `json:"err"`
}

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
		// const (
		// 	traceName = "backend/http/get"
		// 	spanName  = "http/get"
		// )
		// tracing
		// span := startSpan(c, traceName, spanName)
		// defer span.End()
		// prep
		// requestId, err := middleware.GetRequestIDFromCtx(c)
		// newLg := middleware.LogReq(c, requestId, lg, false)
		// if err != nil {
		//	logErrorAndTraceEvent(err, "cannot get request id from context", span, newLg)
		// }
		// get key
		key, err := getKey(c)
		if err != nil {
			// logErrorAndTraceEvent(err, "failed to get key", span, newLg)
			c.JSON(http.StatusBadRequest, errStatusBadRequest)
			return
		}
		// span.SetAttributes(attribute.String("key", key))
		// newLg.Debug("decoded key", "key", key)
		// invoke request
		val, err := s.st.Get(c.Request.Context(), key)
		if err != nil {
			if errors.Is(err, cache.ErrNotFound) {
				// logErrorAndTraceEvent(err, "key not found", span, newLg)
				c.JSON(http.StatusNotFound, errStatusNotFound)
				return
			}
			// logErrorAndTraceEvent(err, "error accessing storage", span, newLg)
			c.JSON(http.StatusInternalServerError, errStatusInternalServer)
			return
		}

		result := body{Key: key, Val: fmt.Sprintf("%v", val)}
		c.JSON(http.StatusOK, result)
	}
}

func (s *StorageHandler) ginSet() func(c *gin.Context) {
	return func(c *gin.Context) {
		// const (
		// 	traceName = "backend/http/set"
		// 	spanName  = "http/set"
		// )
		// tracing
		// span := startSpan(c, traceName, spanName)
		// defer span.End()
		// prep
		// id, err := middleware.GetRequestIDFromCtx(c)
		// newLg := middleware.LogReq(c, id, lg, false)
		// if err != nil {
		//	logErrorAndTraceEvent(err, "cannot get request id from context", span, newLg)
		// }
		b := &body{}
		err := c.ShouldBindJSON(&b)
		if err != nil {
			// logErrorAndTraceEvent(err, "cannot get body from request", span, newLg)
			c.JSON(http.StatusBadRequest, errStatusBadRequest)
			return
		}
		// span.SetAttributes(attribute.String("key", b.Key), attribute.String("value", b.Val))
		// newLg.Debug("request body", "key", b.Key, "value", b.Val)
		// invoke request
		code, err := s.st.Set(c.Request.Context(), b.Key, b.Val)
		if err != nil {
			// logErrorAndTraceEvent(err, "error accessing storage", span, newLg)
			c.JSON(http.StatusInternalServerError, errStatusInternalServer)
			return
		}

		c.JSON(int(code), "ok")
	}
}

func (s *StorageHandler) ginDel() func(c *gin.Context) {
	return func(c *gin.Context) {
		// const (
		// 	traceName = "backend/http/set"
		// 	spanName  = "http/set"
		// )
		// tracing
		// span := startSpan(c, traceName, spanName)
		// defer span.End()
		// prep
		// requestId, err := middleware.GetRequestIDFromCtx(c)
		// newLg := middleware.LogReq(c, requestId, lg, false)
		// if err != nil {
		//	logErrorAndTraceEvent(err, "cannot get request id from context", span, newLg)
		// }
		// get key
		key, err := getKey(c)
		if err != nil {
			// logErrorAndTraceEvent(err, "failed to get key", span, newLg)
			c.JSON(http.StatusBadRequest, errStatusBadRequest)
			return
		}
		// span.SetAttributes(attribute.String("key", key))
		// newLg.Debug("decoded key", "key", key)
		// invoke request
		err = s.st.Delete(c.Request.Context(), key)
		if err != nil {
			// logErrorAndTraceEvent(err, "error accessing storage", span, newLg)
			c.JSON(http.StatusInternalServerError, errStatusInternalServer)
			return
		}

		c.JSON(http.StatusNoContent, "ok")
	}
}

func getKey(c *gin.Context) (string, error) {
	key := c.Param("key")
	if key == "" {
		return "", fmt.Errorf("no key provided")
	}

	return key, nil
}

//	func logErrorAndTraceEvent(err error, msg string, span trace.Span, lg *slog.Logger) {
//		span.AddEvent(msg, trace.WithAttributes(attribute.String("error", err.Error())))
//		lg.Error(msg, "error", err)
//	}

// func startSpan(c *gin.Context, traceName, spanName string) trace.Span {
// 	tracer := otel.Tracer(traceName)
// 	ctx, span := tracer.Start(c.Request.Context(), spanName)
// 	c.Request = c.Request.WithContext(ctx)
// 	return span
// }

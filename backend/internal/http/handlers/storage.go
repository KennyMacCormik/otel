package handlers

import (
	"errors"
	"fmt"
	"github.com/KennyMacCormik/HerdMaster/pkg/cache"
	"github.com/KennyMacCormik/HerdMaster/pkg/gin/middleware"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"log/slog"
	"net/http"
)

type body struct {
	Key string `json:"key"`
	Val string `json:"val,omitempty"`
}

type errorMsg struct {
	Err string `json:"err"`
}

func logErrorAndTraceEvent(err error, msg string, span trace.Span, lg *slog.Logger) {
	span.AddEvent(msg, trace.WithAttributes(attribute.String("error", err.Error())))
	lg.Error(msg, "error", err)
}

func NewStorageHandlers(st cache.Interface, lg *slog.Logger) func(*gin.Engine) {
	return func(router *gin.Engine) {
		router.GET("/storage", get(st, lg))
		router.PUT("/storage", set(st, lg))
		router.POST("/storage", set(st, lg))
		router.DELETE("/storage", del(st, lg))
	}
}

func startSpan(c *gin.Context) trace.Span {
	tracer := otel.Tracer("backend/get")
	ctx, span := tracer.Start(c.Request.Context(), "get")
	c.Request = c.Request.WithContext(ctx)
	return span
}

func get(st cache.Interface, lg *slog.Logger) func(c *gin.Context) {
	return func(c *gin.Context) {
		// tracing
		span := startSpan(c)
		defer span.End()
		// prep
		id, err := middleware.GetRequestIDFromCtx(c)
		newLg := middleware.LogReq(c, id, lg, false)
		if err != nil {
			logErrorAndTraceEvent(err, "cannot get request id from context", span, newLg)
		}
		b := &body{}
		// get body
		err = c.ShouldBindJSON(&b)
		if err != nil {
			logErrorAndTraceEvent(err, "cannot get body from request", span, newLg)
			c.JSON(http.StatusBadRequest, errorMsg{Err: "malformed body"})
			return
		}
		newLg.Debug("request body", "body", b)
		// invoke request
		span.SetAttributes(attribute.String("Key", b.Key))
		val, err := st.Get(c.Request.Context(), b.Key)
		if err != nil {
			if errors.Is(err, cache.ErrNotFound) {
				logErrorAndTraceEvent(err, "key not found", span, newLg)
				c.JSON(http.StatusNotFound, "not found")
				return
			}
			logErrorAndTraceEvent(err, "error accessing storage", span, newLg)
			c.JSON(http.StatusInternalServerError, errorMsg{Err: "server-side error"})
			return
		}
		// send response
		result := body{Key: b.Key, Val: fmt.Sprintf("%v", val)}
		c.JSON(http.StatusOK, result)
	}
}

func set(st cache.Interface, lg *slog.Logger) func(c *gin.Context) {
	return func(c *gin.Context) {
		// tracing
		span := startSpan(c)
		defer span.End()
		// prep
		id, err := middleware.GetRequestIDFromCtx(c)
		newLg := middleware.LogReq(c, id, lg, false)
		if err != nil {
			logErrorAndTraceEvent(err, "cannot get request id from context", span, newLg)
		}
		b := &body{}
		// get body
		err = c.ShouldBindJSON(&b)
		if err != nil {
			logErrorAndTraceEvent(err, "cannot get body from request", span, newLg)
			c.JSON(http.StatusBadRequest, errorMsg{Err: "malformed body"})
			return
		}
		newLg.Debug("request body", "body", b)
		// invoke request
		span.SetAttributes(
			attribute.String("Key", b.Key),
			attribute.String("Val", b.Val),
		)
		err = st.Set(c.Request.Context(), b.Key, b.Val)
		if err != nil {
			logErrorAndTraceEvent(err, "error accessing storage", span, newLg)
			c.JSON(http.StatusInternalServerError, errorMsg{Err: "server-side error"})
			return
		}
		// send response
		c.JSON(http.StatusOK, "ok")
	}
}

func del(st cache.Interface, lg *slog.Logger) func(c *gin.Context) {
	return func(c *gin.Context) {
		// tracing
		span := startSpan(c)
		defer span.End()
		// prep
		id, err := middleware.GetRequestIDFromCtx(c)
		newLg := middleware.LogReq(c, id, lg, false)
		if err != nil {
			logErrorAndTraceEvent(err, "cannot get request id from context", span, newLg)
		}
		b := &body{}
		// get body
		err = c.ShouldBindJSON(&b)
		if err != nil {
			logErrorAndTraceEvent(err, "cannot get body from request", span, newLg)
			c.JSON(http.StatusBadRequest, errorMsg{Err: "malformed body"})
			return
		}
		newLg.Debug("request body", "body", b)
		// invoke request
		span.SetAttributes(attribute.String("Key", b.Key))
		err = st.Delete(c.Request.Context(), b.Key)
		if err != nil {
			logErrorAndTraceEvent(err, "error accessing storage", span, newLg)
			c.JSON(http.StatusInternalServerError, errorMsg{Err: "server-side error"})
			return
		}
		// send response
		c.JSON(http.StatusOK, "ok")
	}
}

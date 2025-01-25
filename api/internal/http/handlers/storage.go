package handlers

import (
	"api/internal/compute"
	"errors"
	"fmt"
	"github.com/KennyMacCormik/HerdMaster/pkg/cache"
	"github.com/KennyMacCormik/HerdMaster/pkg/cache/wrappers/ttlCache"
	"github.com/KennyMacCormik/HerdMaster/pkg/gin/middleware"
	"github.com/gin-gonic/gin"
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

func NewStorageHandlers(comp compute.Interface, lg *slog.Logger) func(*gin.Engine) {
	return func(router *gin.Engine) {
		router.GET("/storage", get(comp, lg))
		router.PUT("/storage", set(comp, lg))
		router.POST("/storage", set(comp, lg))
		router.DELETE("/storage", del(comp, lg))
	}
}
func get(comp compute.Interface, lg *slog.Logger) func(c *gin.Context) {
	return func(c *gin.Context) {
		// prep
		requestId, err := middleware.GetRequestIDFromCtx(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorMsg{Err: "server-side error"})
			return
		}
		ctx := c.Request.Context()
		b := &body{}
		// get body
		err = c.ShouldBindJSON(&b)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorMsg{Err: "malformed body"})
			return
		}
		// invoke request
		val, err := comp.Get(ctx, b.Key, requestId)
		if err != nil {
			if errors.Is(err, cache.ErrNotFound) || errors.Is(err, &ttlCache.ErrTimeout{}) {
				c.JSON(http.StatusNotFound, "not found")
				return
			}
			c.JSON(http.StatusInternalServerError, errorMsg{Err: "server-side error"})
			return
		}
		// send response
		result := body{Key: b.Key, Val: fmt.Sprintf("%v", val)}
		c.JSON(http.StatusOK, result)
	}
}

func set(comp compute.Interface, lg *slog.Logger) func(c *gin.Context) {
	return func(c *gin.Context) {
		// prep
		ctx := c.Request.Context()
		b := &body{}
		// get body
		err := c.ShouldBindJSON(&b)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorMsg{Err: "malformed body"})
			return
		}
		// invoke request
		err = comp.Set(ctx, b.Key, b.Val)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorMsg{Err: "server-side error"})
			return
		}
		// send response
		c.JSON(http.StatusOK, "ok")
	}
}

func del(comp compute.Interface, lg *slog.Logger) func(c *gin.Context) {
	return func(c *gin.Context) {
		// prep
		ctx := c.Request.Context()
		b := &body{}
		// get body
		err := c.ShouldBindJSON(&b)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorMsg{Err: "malformed body"})
			return
		}
		// invoke request
		err = comp.Delete(ctx, b.Key)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorMsg{Err: "server-side error"})
			return
		}
		// send response
		c.JSON(http.StatusOK, "ok")
	}
}

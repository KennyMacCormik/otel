package storage

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/KennyMacCormik/otel/backend/pkg/cache"
	cacheErrors "github.com/KennyMacCormik/otel/backend/pkg/models/errors/cache"
	httpModels "github.com/KennyMacCormik/otel/backend/pkg/models/http"
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
		key, err := getKey(c)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		val, err := s.st.Get(c.Request.Context(), key)
		if err != nil {
			if errors.Is(err, cacheErrors.ErrNotFound) {
				c.Status(http.StatusNotFound)
				return
			}
			c.Status(http.StatusInternalServerError)
			return
		}

		result := httpModels.Body{Key: key, Val: fmt.Sprintf("%v", val)}
		c.JSON(http.StatusOK, result)
	}
}

func (s *StorageHandler) ginSet() func(c *gin.Context) {
	return func(c *gin.Context) {
		b := &httpModels.Body{}

		err := c.ShouldBindJSON(&b)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		code, err := s.st.Set(c.Request.Context(), b.Key, b.Val)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		c.Status(code)
	}
}

func (s *StorageHandler) ginDel() func(c *gin.Context) {
	return func(c *gin.Context) {
		key, err := getKey(c)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		err = s.st.Delete(c.Request.Context(), key)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		c.Status(http.StatusNoContent)
	}
}

func getKey(c *gin.Context) (string, error) {
	key := c.Param("key")
	if key == "" {
		return "", fmt.Errorf("no key provided")
	}

	return key, nil
}

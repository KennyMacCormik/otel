package storage

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

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
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	if !isUrlEncoded(strings.TrimPrefix(c.Request.RequestURI, "/storage/"), key) {
		return "", fmt.Errorf("key must be URL-encoded")
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

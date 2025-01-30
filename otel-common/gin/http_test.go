package http

import (
	"github.com/KennyMacCormik/HerdMaster/pkg/gin/router"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func TestNewHttpServer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	gf := router.NewGinFactory()
	gf.AddHandlers(func(r *gin.Engine) {
		r.GET("/ping", func(c *gin.Context) {
			c.String(http.StatusOK, "pong")
		})
	})

	server := NewHttpServer(
		"127.0.0.1:8080",
		gf,
		10*time.Second,
		10*time.Second,
		10*time.Second,
	)

	assert.NotNil(t, server, "HttpServer instance should not be nil")
	assert.Equal(t, "127.0.0.1:8080", server.svr.Addr, "HttpServer should use the correct address")
	assert.Equal(t, 10*time.Second, server.svr.ReadTimeout, "HttpServer should set the correct ReadTimeout")
	assert.Equal(t, 10*time.Second, server.svr.WriteTimeout, "HttpServer should set the correct WriteTimeout")
	assert.Equal(t, 10*time.Second, server.svr.IdleTimeout, "HttpServer should set the correct IdleTimeout")
}

func TestHttpServer_StartAndClose(t *testing.T) {
	gin.SetMode(gin.TestMode)
	gf := router.NewGinFactory()
	gf.AddHandlers(func(r *gin.Engine) {
		r.GET("/ping", func(c *gin.Context) {
			c.String(http.StatusOK, "pong")
		})
	})

	server := NewHttpServer(
		"127.0.0.1:8080",
		gf,
		10*time.Second,
		10*time.Second,
		10*time.Second,
	)

	// Start the server in a separate goroutine
	go func() {
		err := server.Start()
		assert.ErrorIs(t, err, http.ErrServerClosed, "Server should be closed gracefully")
	}()

	// Simulate a client request
	client := &http.Client{}
	req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1:8080/ping", nil)
	time.Sleep(100 * time.Millisecond) // Allow server to start
	resp, err := client.Do(req)
	assert.NoError(t, err, "Client request should not return an error")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Response status should be 200 OK")

	// Close the server
	err = server.Close(5 * time.Second)
	assert.NoError(t, err, "Server should close without errors")
}

func TestHttpServer_CloseTimeout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	gf := router.NewGinFactory()

	// create channels for sync
	serverStarted, requestInProgress := make(chan struct{}), make(chan struct{})

	// Add a handler that simulates a long-running request
	gf.AddHandlers(func(r *gin.Engine) {
		r.GET("/long", func(c *gin.Context) {
			close(requestInProgress)
			time.Sleep(2 * time.Second) // Simulate a long-running task
			c.String(http.StatusOK, "done")
		})
	})

	server := NewHttpServer(
		"127.0.0.1:8081",
		gf,
		10*time.Second,
		10*time.Second,
		10*time.Second,
	)

	// Start the server in a separate goroutine
	go func() {
		close(serverStarted)
		_ = server.Start()
	}()
	<-serverStarted // Ensure the server is fully started

	// Invoke long call
	go func() {
		client := &http.Client{}
		_, _ = client.Get("http://127.0.0.1:8081/long")
	}()

	// Simulate a context timeout
	<-requestInProgress                      // Ensure the request is actively blocking
	err := server.Close(1 * time.Nanosecond) // Set a very short timeout

	// Assert that an error occurs due to timeout
	assert.Error(t, err, "Server should return an error when shutdown times out")
}

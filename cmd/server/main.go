package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/8fs/8fs/internal/config"
	"github.com/8fs/8fs/internal/container"
	"github.com/8fs/8fs/internal/transport/http/middleware"
	"github.com/8fs/8fs/internal/transport/http/router"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize container with dependencies
	c, err := container.NewContainer(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize container: %v", err)
	}

	c.Logger.Info("Starting 8fs server",
		"version", "0.2.0",
		"storage_driver", cfg.Storage.Driver,
		"auth_driver", cfg.Auth.Driver,
	)

	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Create HTTP server
	server := &http.Server{
		Addr:         cfg.Address(),
		Handler:      setupRouter(c),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		c.Logger.Info("Server listening", "address", cfg.Address())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			c.Logger.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	c.Logger.Info("Shutting down server...")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server gracefully
	if err := server.Shutdown(ctx); err != nil {
		c.Logger.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	c.Logger.Info("Server exited")
}

// setupRouter configures and returns the HTTP router
func setupRouter(c *container.Container) *gin.Engine {
	r := gin.New()

	// Global middleware
	r.Use(gin.Recovery())
	r.Use(middleware.Logger(c.Logger))
	r.Use(middleware.RequestID())

	if c.Config.Audit.Enabled {
		r.Use(middleware.Audit(c.AuditLogger))
	}

	if c.Config.Metrics.Enabled {
		r.Use(middleware.Metrics())
	}

	// Setup routes
	router.SetupRoutes(r, c)

	return r
}

package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zellis-rameesn/go-ecommerce/internal/config"
	"github.com/zellis-rameesn/go-ecommerce/internal/database"
	"github.com/zellis-rameesn/go-ecommerce/internal/logger"
	"github.com/zellis-rameesn/go-ecommerce/internal/server"
	"github.com/zellis-rameesn/go-ecommerce/internal/services"
)

func main() {

	cfg := config.Load()

	log := logger.New(cfg.Server.GinMode)

	db, err := database.New(&cfg.Database)

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}

	mainDB, err := db.DB()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get database connection")
	}
	defer mainDB.Close()
	gin.SetMode(cfg.Server.GinMode)

	authService := services.NewAuthService(db, cfg)
	userService := services.NewUserService(db)
	productService := services.NewProductService(db)

	srv := server.New(cfg, &log, db, authService, userService, productService)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      srv.SetupRoutes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("Server stopped unexpectedly")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to shutdown server")
	}

	log.Info().Msg("Shutting down database")
}

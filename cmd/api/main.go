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
	"github.com/zellis-rameesn/go-ecommerce/internal/events"
	"github.com/zellis-rameesn/go-ecommerce/internal/interfaces"
	"github.com/zellis-rameesn/go-ecommerce/internal/logger"
	"github.com/zellis-rameesn/go-ecommerce/internal/providers"
	"github.com/zellis-rameesn/go-ecommerce/internal/server"
	"github.com/zellis-rameesn/go-ecommerce/internal/services"
)

func main() {
	// @title           ECommerce Rest API
	// @version         1.0
	// @description     This is an ecommerce rest api
	// @termsOfService  http://swagger.io/terms/

	// @contact.name   API Support
	// @contact.url    http://www.swagger.io/support
	// @contact.email  support@swagger.io

	// @license.name  Apache 2.0
	// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

	// @host      localhost:8080
	// @BasePath  /api/v1

	// @securityDefinitions.apikey BearerAuth
	// @in header
	// @name Authorization
	// @description Type "Bearer" followed by a space and JWT token.

	// @externalDocs.description  OpenAPI
	// @externalDocs.url          https://swagger.io/resources/open-api/
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

	ctx := context.Background()
	publisher, err := events.NewEventPublisher(ctx, &cfg.AWS)
	if err != nil {
		log.Error().Msg("Failed to create publisher")
		return
	}

	gin.SetMode(cfg.Server.GinMode)

	authService := services.NewAuthService(db, cfg, publisher)
	userService := services.NewUserService(db)
	productService := services.NewProductService(db)
	cartService := services.NewCartService(db)
	orderService := services.NewOrderService(db)

	var uploadProvider interfaces.UploadInterface
	if cfg.Upload.Provider == "s3" {
		uploadProvider = providers.NewS3Provider(&cfg.AWS)
	} else {
		uploadProvider = providers.NewLocalProvider(cfg.Upload.Path)
	}
	uploadService := services.NewUploadService(uploadProvider)

	srv := server.New(cfg, &log, db, authService, userService, productService, uploadService, cartService, orderService)

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

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/zellis-rameesn/go-ecommerce/internal/config"
	"github.com/zellis-rameesn/go-ecommerce/internal/database"
	"github.com/zellis-rameesn/go-ecommerce/internal/logger"
)

func main() {

	cfg := config.Load()

	log := logger.New(cfg.Server.GinMode)

	db, err := database.New(cfg.Database)

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}

	mainDB, err := db.DB()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get database connection")
	}
	defer mainDB.Close()
	gin.SetMode(cfg.Server.GinMode)
	log.Info().Msg("Starting server")
}

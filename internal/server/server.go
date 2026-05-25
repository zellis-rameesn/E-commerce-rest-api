package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/zellis-rameesn/go-ecommerce/internal/config"
	"gorm.io/gorm"
)

type Server struct {
	Config *config.Config
	Logger *zerolog.Logger
	DB     *gorm.DB
}

func New(cfg *config.Config, logger *zerolog.Logger, db *gorm.DB) *Server {
	return &Server{
		Config: cfg,
		Logger: logger,
		DB:     db,
	}
}

func (s *Server) SetupRoutes() *gin.Engine {
	router := gin.Default()
	router.Use(s.corsMiddleware)

	router.GET("/health", s.healthCheck)

	api := router.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{ //nolint:gocritic // I need this for readability
			auth.POST("/register", s.register)
			auth.POST("/login", s.login)
			auth.POST("/refresh", s.refreshToken)
			auth.POST("/logout", s.logout)
		}

		protected := api.Group("/")
		protected.Use(s.authMiddleware)
		{
			user := protected.Group("/user")
			{ //nolint:gocritic // I need this for readability
				user.GET("/profile", s.getProfile)
				user.PUT("/update-profile", s.updateProfile)
			}
		}
	}

	return router
}

func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

func (s *Server) corsMiddleware(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "http://localhost:3000")
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(204)
		return
	}
	c.Next()
}

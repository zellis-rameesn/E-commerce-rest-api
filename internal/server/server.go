package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/zellis-rameesn/go-ecommerce/internal/config"
	"github.com/zellis-rameesn/go-ecommerce/internal/services"
	"gorm.io/gorm"
)

type Server struct {
	Config         *config.Config
	Logger         *zerolog.Logger
	DB             *gorm.DB
	AuthService    *services.AuthService
	UserService    *services.UserService
	ProductService *services.ProductService
	UploadService  *services.UploadService
}

func New(cfg *config.Config, logger *zerolog.Logger, db *gorm.DB, authService *services.AuthService, userService *services.UserService, productService *services.ProductService, uploadService *services.UploadService) *Server {
	return &Server{
		Config:         cfg,
		Logger:         logger,
		DB:             db,
		AuthService:    authService,
		UserService:    userService,
		ProductService: productService,
		UploadService:  uploadService,
	}
}

func (s *Server) SetupRoutes() *gin.Engine {
	router := gin.Default()
	router.Use(s.corsMiddleware)

	router.Static("/uploads", "./uploads")

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
		{
			category := protected.Group("/categories")
			category.Use(s.adminMiddleware)
			{ //nolint:gocritic // I need this for readability
				category.POST("/", s.createCategory)
				category.PUT("/:id", s.updateCategory)
				category.DELETE("/:id", s.deleteCategory)
			}
		}
		{
			product := protected.Group("/products")
			product.Use(s.adminMiddleware)
			{ //nolint:gocritic // I need this for readability
				product.POST("/", s.createProduct)
				product.PUT("/:id", s.updateProduct)
				product.DELETE("/:id", s.deleteProduct)
				product.POST("/:id/image", s.uploadImage)
			}
		}
	}

	// public routes
	api.GET("/categories", s.getCategories)
	api.GET("/products", s.getProducts)
	api.GET("/product/:id", s.getProduct)

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

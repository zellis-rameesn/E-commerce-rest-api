package server

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/zellis-rameesn/go-ecommerce/docs"
	"github.com/zellis-rameesn/go-ecommerce/internal/config"
	"github.com/zellis-rameesn/go-ecommerce/internal/ratelimit"
	"github.com/zellis-rameesn/go-ecommerce/internal/services"
	"github.com/zellis-rameesn/go-ecommerce/internal/utils"
	"gorm.io/gorm"
)

type Server struct {
	Config         *config.Config
	Logger         *zerolog.Logger
	DB             *gorm.DB
	RateLimiter    *ratelimit.RedisTokenBucket
	AuthService    *services.AuthService
	UserService    *services.UserService
	ProductService *services.ProductService
	UploadService  *services.UploadService
	CartService    *services.CartService
	OrderService   *services.OrderService
}

func New(cfg *config.Config, logger *zerolog.Logger, db *gorm.DB, rateLimiter *ratelimit.RedisTokenBucket, authService *services.AuthService, userService *services.UserService, productService *services.ProductService, uploadService *services.UploadService, cartService *services.CartService, orderService *services.OrderService) *Server {
	return &Server{
		Config:         cfg,
		Logger:         logger,
		DB:             db,
		RateLimiter:    rateLimiter,
		AuthService:    authService,
		UserService:    userService,
		ProductService: productService,
		UploadService:  uploadService,
		CartService:    cartService,
		OrderService:   orderService,
	}
}

func (s *Server) SetupRoutes() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(s.customRecovery())
	router.Use(s.corsMiddleware)

	router.Static("/uploads", "./uploads")

	router.GET("/health", s.healthCheck)

	// Add documentation routes
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.StaticFile("/api-docs", "./docs/rapidoc.html")

	api := router.Group("/api/v1")
	api.Use(s.RateLimiterMiddleware(s.RateLimiter, RateLimitConfig{
		KeyExtractor:   IPKeyExtractor(),
		ErrorMessage:   "Rate limit exceeded. Please try again later",
		IncludeHeaders: true,
		Limit:          s.Config.RateLimit.Capacity,
		RefillRate:     s.Config.RateLimit.RefillRate,
		FailOpen:       true,
	}))
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
		{
			cart := protected.Group("/cart")
			{ //nolint:gocritic // I need this for readability
				cart.GET("/", s.getCart)
				cart.POST("/", s.addCart)
				cart.PUT("/item/:itemID", s.updateCart)
				cart.DELETE("/item/:itemID", s.removeCartItem)
			}
		}
		{
			order := protected.Group("/orders")
			{ //nolint:gocritic // I need this for readability
				order.GET("/", s.getOrders)
				order.GET("/:id", s.getOrder)
				order.POST("/", s.createOrder)
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

func (s *Server) customRecovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, err any) {
		log.Printf("PANIC: %v", err)
		utils.AbortResponse(c, "An unexpected error has occurred")
	})
}

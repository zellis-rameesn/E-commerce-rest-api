package server

import (
	"github.com/gin-gonic/gin"
	"github.com/zellis-rameesn/go-ecommerce/internal/dto"
	"github.com/zellis-rameesn/go-ecommerce/internal/services"
	"github.com/zellis-rameesn/go-ecommerce/internal/utils"
)

func (s *Server) register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
	}
	authService := services.NewAuthService(s.DB, s.Config)

	authResponse, err := authService.Register(&req)
	if err != nil {
		utils.BadRequestResponse(c, "Registration failed", err)
	}
	utils.CreatedResponse(c, "Registration successful", authResponse)
}

func (s *Server) login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
	}
	authService := services.NewAuthService(s.DB, s.Config)

	authResponse, err := authService.Login(&req)
	if err != nil {
		utils.UnauthorizedResponse(c, "Login failed")
	}
	utils.SuccessResponse(c, "Login successful", authResponse)
}

func (s *Server) refreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
	}
	authService := services.NewAuthService(s.DB, s.Config)

	authResponse, err := authService.RefreshToken(&req)
	if err != nil {
		utils.BadRequestResponse(c, "Failed to generate token", err)
	}
	utils.SuccessResponse(c, "Token refresh successful", authResponse)

}

func (s *Server) logout(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
	}
	authService := services.NewAuthService(s.DB, s.Config)

	err := authService.Logout(req.RefreshToken)
	if err != nil {
		utils.BadRequestResponse(c, "Logout failed", err)
	}
	utils.SuccessResponse(c, "Logout successful", nil)
}

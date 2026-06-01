package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/zellis-rameesn/go-ecommerce/internal/config"
	"github.com/zellis-rameesn/go-ecommerce/internal/dto"
	"github.com/zellis-rameesn/go-ecommerce/internal/models"
	"github.com/zellis-rameesn/go-ecommerce/internal/utils"
	"gorm.io/gorm"
)

type AuthService struct {
	db     *gorm.DB
	config *config.Config
}

func NewAuthService(db *gorm.DB, cfg *config.Config) *AuthService {
	return &AuthService{
		db:     db,
		config: cfg,
	}
}

func (a *AuthService) Register(req *dto.RegisterRequest) (*dto.AuthResponse, error) {
	// Check if user exists
	var existingUser models.User
	if err := a.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, errors.New("user already present")
	}
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := models.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  hashedPassword,
		Phone:     req.Phone,
		Role:      models.UserRoleCustomer,
	}
	if err := a.db.Create(&user).Error; err != nil {
		return nil, err
	}
	cart := models.Cart{
		UserID: user.ID,
	}
	if err = a.db.Create(&cart).Error; err != nil {
		fmt.Println("Unable to create cart")
	}

	return a.GenerateAuthResponse(&user)
}

func (a *AuthService) Login(req *dto.LoginRequest) (*dto.AuthResponse, error) {
	var user models.User
	if err := a.db.Where("email = ? AND is_active = ?", req.Email, true).First(&user).Error; err != nil {
		return nil, errors.New("user not found")
	}
	if isValidPassword := utils.CheckPassword(user.Password, req.Password); !isValidPassword {
		return nil, errors.New("invalid Password")
	}

	return a.GenerateAuthResponse(&user)
}

func (a *AuthService) RefreshToken(req *dto.RefreshTokenRequest) (*dto.AuthResponse, error) {
	claims, err := utils.ValidateToken(req.RefreshToken, a.config.JWT.Secret)
	if err != nil {
		return nil, errors.New("invalid token")
	}
	var refreshToken models.RefreshToken
	if err = a.db.Where("token = ? AND expires_at > ?", req.RefreshToken, time.Now()).First(&refreshToken).Error; err != nil {
		return nil, errors.New("refresh token not found or expired")
	}
	var user models.User
	if err := a.db.First(&user, claims.UserID).Error; err != nil {
		return nil, errors.New("user not found")
	}

	if err := a.db.Delete(&refreshToken).Error; err != nil {
		return nil, errors.New("failed to refresh token")
	}

	return a.GenerateAuthResponse(&user)
}

func (a *AuthService) Logout(refreshToken string) error {
	return a.db.Where("token = ?", refreshToken).Delete(&models.RefreshToken{}).Error
}

func (a *AuthService) GenerateAuthResponse(user *models.User) (*dto.AuthResponse, error) {
	accessToken, refreshToken, err := utils.GenerateTokens(&a.config.JWT, user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	refreshTokenModel := &models.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(a.config.JWT.RefreshTokenExpires),
	}
	if err := a.db.Create(refreshTokenModel).Error; err != nil {
		return nil, errors.New("failed to update refresh token")
	}

	authReponse := &dto.AuthResponse{
		User: dto.UserResponse{
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Phone:     user.Phone,
			Role:      user.Role,
			IsActive:  user.IsActive,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return authReponse, nil
}

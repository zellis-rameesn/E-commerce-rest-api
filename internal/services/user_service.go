package services

import (
	"github.com/zellis-rameesn/go-ecommerce/internal/dto"
	"github.com/zellis-rameesn/go-ecommerce/internal/models"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		db: db,
	}
}

func (u *UserService) GetProfile(userId uint) (*dto.UserResponse, error) {
	var user models.User
	if err := u.db.First(&user, userId).Error; err != nil {
		return nil, err
	}

	return &dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
		Phone:     user.Phone,
		IsActive:  user.IsActive,
	}, nil
}

func (u *UserService) UpdateProfile(userId uint, req *dto.UpdateProfileRequest) (*dto.UserResponse, error) {
	var user models.User
	if err := u.db.First(&user, userId).Error; err != nil {
		return nil, err
	}

	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.Phone = req.Phone

	if err := u.db.Save(&user).Error; err != nil {
		return nil, err
	}

	return u.GetProfile(userId)
}

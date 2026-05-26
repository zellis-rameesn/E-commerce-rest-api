package services

import (
	"github.com/zellis-rameesn/go-ecommerce/internal/dto"
	"github.com/zellis-rameesn/go-ecommerce/internal/models"
	"gorm.io/gorm"
)

type ProductService struct {
	db *gorm.DB
}

func NewProductService(db *gorm.DB) *ProductService {
	return &ProductService{
		db: db,
	}
}

func (p *ProductService) CreateCategory(req *dto.CreateCategoryRequest) (*dto.CategoryResponse, error) {
	category := &models.Category{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := p.db.Create(&category).Error; err != nil {
		return nil, err
	}

	return &dto.CategoryResponse{
		ID:          category.ID,
		Name:        category.Name,
		Description: category.Description,
		IsActive:    category.IsActive,
	}, nil
}

func (p *ProductService) GetCategories() ([]dto.CategoryResponse, error) {
	var categories []models.Category
	if err := p.db.Where("is_active = ?", true).Find(&categories).Error; err != nil {
		return nil, err
	}
	reponse := make([]dto.CategoryResponse, len(categories))
	for i := range categories {
		reponse = append(reponse, dto.CategoryResponse{
			ID:          categories[i].ID,
			Name:        categories[i].Name,
			Description: categories[i].Description,
			IsActive:    categories[i].IsActive,
		})
	}

	return reponse, nil
}

func (p *ProductService) UpdateCategory(id uint, req *dto.UpdateCategoryRequest) (*dto.CategoryResponse, error) {
	var category models.Category
	if err := p.db.First(&category, id).Error; err != nil {
		return nil, err
	}
	category.Name = req.Name
	category.Description = req.Description
	// cannot dereference a nil pointer, if client does not send isActive this can panic, hence the check
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}

	if err := p.db.Save(&category).Error; err != nil {
		return nil, err
	}

	return &dto.CategoryResponse{
		ID:          category.ID,
		Name:        category.Name,
		Description: category.Description,
		IsActive:    category.IsActive,
	}, nil
}

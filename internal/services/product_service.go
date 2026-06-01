package services

import (
	"math"

	"github.com/zellis-rameesn/go-ecommerce/internal/dto"
	"github.com/zellis-rameesn/go-ecommerce/internal/models"
	"github.com/zellis-rameesn/go-ecommerce/internal/utils"
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
	response := make([]dto.CategoryResponse, len(categories))
	for i := range categories {
		response[i] = dto.CategoryResponse{
			ID:          categories[i].ID,
			Name:        categories[i].Name,
			Description: categories[i].Description,
			IsActive:    categories[i].IsActive,
		}
	}

	return response, nil
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

func (p *ProductService) DeleteCategory(id uint) error {
	return p.db.Delete(&models.Category{}, id).Error
}

func (p *ProductService) GetProducts(page, limit int) ([]*dto.ProductResponse, *utils.PaginationMeta, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	var total int64

	p.db.Model(&models.Product{}).Where("is_active = ?", true).Count(&total)

	var products []models.Product
	if err := p.db.Preload("Category").Preload("Images").Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return nil, nil, err
	}

	response := make([]*dto.ProductResponse, len(products))
	for i := range products {
		// append may cause reallocations
		// response = append(response, p.CreateProductResponse(&products[i]))

		response[i] = p.CreateProductResponse(&products[i]) // no reallocations
	}

	totalProducts := math.Ceil(float64(total) / float64(limit))

	meta := &utils.PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: int(totalProducts),
	}

	return response, meta, nil
}

func (p *ProductService) GetProduct(id uint) (*dto.ProductResponse, error) {
	var product models.Product
	if err := p.db.Preload("Category").Preload("Images").First(&product, id).Error; err != nil {
		return nil, err
	}
	return p.CreateProductResponse(&product), nil
}

func (p *ProductService) CreateProduct(req *dto.CreateProductRequest) (*dto.ProductResponse, error) {
	product := &models.Product{
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		SKU:         req.SKU,
	}

	if err := p.db.Create(product).Error; err != nil {
		return nil, err
	}

	return p.CreateProductResponse(product), nil
}

func (p *ProductService) UpdateProduct(id uint, req *dto.UpdateProductRequest) (*dto.ProductResponse, error) {
	var product models.Product
	if err := p.db.First(&product, id).Error; err != nil {
		return nil, err
	}
	product.Name = req.Name
	product.Description = req.Description
	product.Price = req.Price
	product.Stock = req.Stock
	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}
	if err := p.db.Save(&product).Error; err != nil {
		return nil, err
	}

	return p.GetProduct(id)
}

func (p *ProductService) DeleteProduct(id uint) error {
	return p.db.Delete(&models.Product{}, id).Error
}

func (p *ProductService) CreateProductResponse(product *models.Product) *dto.ProductResponse {

	productImages := make([]dto.ProductImageResponse, len(product.Images))

	for i := range product.Images {
		productImages[i] = dto.ProductImageResponse{
			ID:        product.Images[i].ID,
			URL:       product.Images[i].URL,
			AltText:   product.Images[i].AltText,
			IsPrimary: product.Images[i].IsPrimary,
		}
	}

	return &dto.ProductResponse{
		ID:          product.ID,
		CategoryID:  product.CategoryID,
		Name:        product.Name,
		Description: product.Description,
		SKU:         product.SKU,
		Stock:       product.Stock,
		Price:       product.Price,
		IsActive:    product.IsActive,
		Images:      productImages,
		Category: dto.CategoryResponse{
			ID:          product.Category.ID,
			Name:        product.Category.Description,
			Description: product.Category.Description,
			IsActive:    product.Category.IsActive,
		},
	}
}

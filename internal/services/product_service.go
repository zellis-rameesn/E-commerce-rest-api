package services

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zellis-rameesn/go-ecommerce/internal/cache"
	"github.com/zellis-rameesn/go-ecommerce/internal/dto"
	"github.com/zellis-rameesn/go-ecommerce/internal/models"
	"github.com/zellis-rameesn/go-ecommerce/internal/utils"
	"gorm.io/gorm"
)

type ProductService struct {
	db *gorm.DB
	rd *cache.RedisClient
}

type ProductCache struct {
	Response []*dto.ProductResponse `json:"response"`
	Meta     *utils.PaginationMeta  `json:"meta"`
}

func NewProductService(db *gorm.DB, rd *cache.RedisClient) *ProductService {
	return &ProductService{
		db: db,
		rd: rd,
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

func (p *ProductService) getProductsCacheVersion(ctx context.Context) int64 {
	var version int64
	if err := p.rd.Get(ctx, "version", &version); err == redis.Nil {
		log.Printf("Cache miss for version %d", version)
		_ = p.rd.Set(ctx, "version", 1, 0)
		version = 1
	} else if err != nil {
		log.Printf("failed to fetch version from cache: %s", err.Error())
		return 0
	}
	log.Printf("Cache hit for version %d", version)
	return version
}

func (p *ProductService) GetProducts(ctx context.Context, page, limit int) ([]*dto.ProductResponse, *utils.PaginationMeta, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	version := p.getProductsCacheVersion(ctx)
	key := fmt.Sprintf("products:version:%d:page:%d:limit:%d", version, page, limit)

	if version > 0 {
		var productsCache ProductCache
		if err := p.rd.Get(ctx, key, &productsCache); err != nil {
			log.Printf("Failed to fetch from cache: %s", err.Error())
		} else {
			log.Printf("Cache hit for key %s", key)
			return productsCache.Response, productsCache.Meta, nil
		}
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

	cache := ProductCache{
		Response: response,
		Meta:     meta,
	}

	if err := p.rd.Set(ctx, key, cache, 5*time.Minute); err != nil {
		log.Printf("Failed to store %s in cache: %s", key, err.Error())
	}

	return response, meta, nil
}

func (p *ProductService) GetProduct(ctx context.Context, id uint) (*dto.ProductResponse, error) {
	key := fmt.Sprintf("product:%d", id)
	var cachedProduct dto.ProductResponse

	if err := p.rd.Get(ctx, key, &cachedProduct); err != nil {
		log.Printf("Failed to fetch from cache: %s", err.Error())
	} else {
		log.Printf("Cache hit for key %s", key)
		return &cachedProduct, nil
	}

	var product models.Product
	if err := p.db.Preload("Category").Preload("Images").First(&product, id).Error; err != nil {
		return nil, err
	}

	productResponse := p.CreateProductResponse(&product)

	if err := p.rd.Set(ctx, key, productResponse, 5*time.Minute); err != nil {
		log.Printf("Failed to store %s in cache: %s", key, err.Error())
	}

	return productResponse, nil
}

func (p *ProductService) CreateProduct(ctx context.Context, req *dto.CreateProductRequest) (*dto.ProductResponse, error) {
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

	_ = p.rd.Increment(ctx, "version")

	return p.CreateProductResponse(product), nil
}

func (p *ProductService) UpdateProduct(ctx context.Context, id uint, req *dto.UpdateProductRequest) (*dto.ProductResponse, error) {
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

	key := fmt.Sprintf("product:%d", id)
	if err := p.rd.Delete(ctx, key); err != nil {
		log.Printf("Failed to delete %s from cache: %s", key, err.Error())
	}

	_ = p.rd.Increment(ctx, "version")

	return p.GetProduct(ctx, id)
}

func (p *ProductService) DeleteProduct(ctx context.Context, id uint) error {
	if err := p.db.Delete(&models.Product{}, id).Error; err != nil {
		return err
	}

	key := fmt.Sprintf("product:%d", id)
	if err := p.rd.Delete(ctx, key); err != nil {
		log.Printf("Failed to delete %s from cache: %s", key, err.Error())
	}
	_ = p.rd.Increment(ctx, "version")

	return nil
}

func (p *ProductService) AddProductImage(id uint, url, altText string) error {
	var count int64
	p.db.Model(&models.ProductImage{}).Where("product_id = ?", id).Count(&count)

	image := &models.ProductImage{
		ProductID: id,
		URL:       url,
		AltText:   altText,
		IsPrimary: count == 0, // First image is primary
	}
	return p.db.Create(image).Error
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

package server

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zellis-rameesn/go-ecommerce/internal/dto"
	"github.com/zellis-rameesn/go-ecommerce/internal/services"
	"github.com/zellis-rameesn/go-ecommerce/internal/utils"
)

func (s *Server) CreateCategory(c *gin.Context) {
	var req dto.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid data!", err)
		return
	}
	productService := services.NewProductService(s.DB)
	category, err := productService.CreateCategory(&req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create category", err)
		return
	}
	utils.CreatedResponse(c, "Category created successfully", category)
}

func (s *Server) UpdateCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid category id!", err)
		return
	}
	var req dto.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid data!", err)
		return
	}
	productService := services.NewProductService(s.DB)
	category, err := productService.UpdateCategory(uint(id), &req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update category", err)
		return
	}
	utils.SuccessResponse(c, "Category updated successfully", category)
}

func (s *Server) DeleteCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid category id!", err)
		return
	}
	productService := services.NewProductService(s.DB)
	if err := productService.DeleteCategory(uint(id)); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update category", err)
		return
	}
	utils.SuccessResponse(c, "Category deleted successfully", nil)
}

func (s *Server) GetCategories(c *gin.Context) {
	productService := services.NewProductService(s.DB)
	categories, err := productService.GetCategories()
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get category", err)
		return
	}
	utils.SuccessResponse(c, "Categories fetched successfully", categories)
}

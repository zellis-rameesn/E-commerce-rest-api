package server

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zellis-rameesn/go-ecommerce/internal/dto"
	"github.com/zellis-rameesn/go-ecommerce/internal/utils"
)

func (s *Server) getCart(c *gin.Context) {
	userID := c.GetUint("user_id")
	cartResponse, err := s.CartService.GetCart(userID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get cart", err)
		return
	}
	utils.SuccessResponse(c, "Cart fetched", cartResponse)
}

func (s *Server) addCart(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req dto.AddToCartRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid data!", err)
		return
	}

	cartResponse, err := s.CartService.AddToCart(userID, &req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to add to cart", err)
		return
	}
	utils.CreatedResponse(c, "Added to cart successfully", cartResponse)
}

func (s *Server) updateCart(c *gin.Context) {
	userID := c.GetUint("user_id")

	itemID, err := strconv.ParseUint(c.Param("itemID"), 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid product id!", err)
		return
	}

	var req dto.UpdateCartItemRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid data!", err)
		return
	}

	cartResponse, err := s.CartService.UpdateCartItem(userID, uint(itemID), &req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update cart", err)
		return
	}
	utils.CreatedResponse(c, "Updated cart successfully", cartResponse)
}

func (s *Server) removeCartItem(c *gin.Context) {
	userID := c.GetUint("user_id")

	itemID, err := strconv.ParseUint(c.Param("itemID"), 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid product id!", err)
		return
	}

	err = s.CartService.RemoveCartItem(userID, uint(itemID))
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to delete cart item", err)
		return
	}
	utils.CreatedResponse(c, "Deleted cart item successfully", nil)
}

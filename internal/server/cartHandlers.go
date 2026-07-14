package server

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zellis-rameesn/go-ecommerce/internal/dto"
	"github.com/zellis-rameesn/go-ecommerce/internal/utils"
)

// @Summary Get user's cart
// @Description Retrieve current user's shopping cart with all items
// @Tags Cart
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=dto.CartResponse} "Cart retrieved successfully"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Failure 404 {object} utils.Response "Cart not found"
// @Router /cart [get]
func (s *Server) getCart(c *gin.Context) {
	userID := c.GetUint("user_id")
	cartResponse, err := s.CartService.GetCart(userID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get cart", err)
		return
	}
	utils.SuccessResponse(c, "Cart fetched", cartResponse)
}

// @Summary Add item to cart
// @Description Add a product to the user's shopping cart
// @Tags Cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.AddToCartRequest true "Item to add to cart"
// @Success 200 {object} utils.Response{data=dto.CartResponse} "Item added to cart successfully"
// @Failure 400 {object} utils.Response "Invalid request data or insufficient stock"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Router /cart [post]
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

// @Summary Update cart item quantity
// @Description Update the quantity of an item in the user's cart
// @Tags Cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Cart Item ID"
// @Param request body dto.UpdateCartItemRequest true "New quantity"
// @Success 200 {object} utils.Response{data=dto.CartResponse} "Cart item updated successfully"
// @Failure 400 {object} utils.Response "Invalid request data or insufficient stock"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Router /cart/item/{id} [put]
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

// @Summary Remove item from cart
// @Description Remove an item from the user's shopping cart
// @Tags Cart
// @Security BearerAuth
// @Param id path int true "Cart Item ID"
// @Success 200 {object} utils.Response "Item removed from cart successfully"
// @Failure 400 {object} utils.Response "Invalid cart item ID"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Router /cart/item/{id} [delete]
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

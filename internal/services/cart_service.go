package services

import (
	"errors"
	"fmt"

	"github.com/zellis-rameesn/go-ecommerce/internal/dto"
	"github.com/zellis-rameesn/go-ecommerce/internal/models"
	"gorm.io/gorm"
)

type CartService struct {
	db *gorm.DB
}

func NewCartService(db *gorm.DB) *CartService {
	return &CartService{
		db: db,
	}
}

func (c *CartService) GetCart(userID uint) (*dto.CartResponse, error) {
	var cart models.Cart
	if err := c.db.Preload("CartItems.Product.Category").Where("user_id = ?", userID).First(&cart).Error; err != nil {
		return nil, err
	}
	return c.convertToCartResponse(&cart), nil
}

func (c *CartService) AddToCart(userID uint, req *dto.AddToCartRequest) (*dto.CartResponse, error) {
	var product models.Product
	if err := c.db.First(&product, req.ProductID).Error; err != nil {
		return nil, errors.New("product not found")
	}

	if product.Stock < req.Quantity {
		return nil, errors.New("insufficient stock")
	}

	var cart models.Cart
	if err := c.db.Where("user_id = ?", userID).First(&cart).Error; err != nil {
		cart := &models.Cart{
			UserID: userID,
		}
		if err := c.db.Create(cart).Error; err != nil {
			return nil, err
		}
	}

	var cartItem models.CartItem
	if err := c.db.Unscoped().Where("cart_id = ? AND product_id = ?", cart.ID, product.ID).First(&cartItem).Error; err != nil {
		cartItem = models.CartItem{
			CartID:    cart.ID,
			ProductID: product.ID,
			Quantity:  req.Quantity,
		}
		if err := c.db.Create(&cartItem).Error; err != nil {
			return nil, err
		}
	} else {
		if cartItem.DeletedAt.Valid {
			cartItem.DeletedAt = gorm.DeletedAt{}
			cartItem.Quantity = req.Quantity
			if err := c.db.Unscoped().Save(&cartItem).Error; err != nil {
				return nil, err
			}
			return c.GetCart(userID)
		}
		cartItem.Quantity += req.Quantity

		if cartItem.Quantity > product.Stock {
			return nil, errors.New("insufficient stock")
		}
		if err := c.db.Create(&cartItem).Error; err != nil {
			return nil, err
		}
	}

	return c.GetCart(userID)
}

func (c *CartService) UpdateCartItem(userID, itemID uint, req *dto.UpdateCartItemRequest) (*dto.CartResponse, error) {
	// // preloading here will fetch all the cart items, which is not really needed since we only need the specific cart item
	// // we cannot filter by product id since we haven't joined that table
	// // hence using join is the better approach

	var cartItem models.CartItem
	if err := c.db.Joins(`LEFT JOIN carts c on c.id = cart_items.cart_id`).Where(`c.user_id = ? AND cart_items.id = ?`, userID, itemID).First(&cartItem).Error; err != nil {
		return nil, errors.New("cart item not found")
	}

	var product models.Product
	if err := c.db.First(&product, cartItem.ProductID).Error; err != nil {
		return nil, errors.New("product not found")
	}
	if product.Stock < req.Quantity {
		return nil, errors.New("insufficient stock")
	}
	fmt.Println("cartItem", cartItem)
	cartItem.Quantity = req.Quantity
	if err := c.db.Save(&cartItem).Error; err != nil {
		return nil, err
	}

	return c.GetCart(userID)
}

func (c *CartService) RemoveCartItem(userID, itemID uint) error {
	var cartItem models.CartItem
	return c.db.Joins(`LEFT JOIN carts c on c.id = cart_items.cart_id`).Where(`c.user_id = ? AND cart_items.id = ?`, userID, itemID).Delete(&cartItem).Error
}

func (c *CartService) convertToCartResponse(cart *models.Cart) *dto.CartResponse {
	cartItems := make([]dto.CartItemResponse, len(cart.CartItems))
	var total float64
	for i := range cart.CartItems {
		subTotal := float64(cart.CartItems[i].Quantity) * cart.CartItems[i].Product.Price
		total += subTotal
		cartItems[i] = dto.CartItemResponse{
			ID:       cart.CartItems[i].ID,
			Quantity: cart.CartItems[i].Quantity,
			Subtotal: subTotal,
			Product: dto.ProductResponse{
				ID:          cart.CartItems[i].Product.ID,
				CategoryID:  cart.CartItems[i].Product.CategoryID,
				Name:        cart.CartItems[i].Product.Name,
				Description: cart.CartItems[i].Product.Description,
				Price:       cart.CartItems[i].Product.Price,
				Stock:       cart.CartItems[i].Product.Stock,
				SKU:         cart.CartItems[i].Product.SKU,
				IsActive:    cart.CartItems[i].Product.IsActive,
				Category: dto.CategoryResponse{
					ID:          cart.CartItems[i].Product.Category.ID,
					Name:        cart.CartItems[i].Product.Category.Name,
					Description: cart.CartItems[i].Product.Category.Description,
					IsActive:    cart.CartItems[i].Product.Category.IsActive,
				},
			},
		}
	}

	return &dto.CartResponse{
		CartItems: cartItems,
		Total:     total,
		UserID:    cart.UserID,
		ID:        cart.ID,
	}
}

package services

import (
	"errors"
	"math"

	"github.com/zellis-rameesn/go-ecommerce/internal/dto"
	"github.com/zellis-rameesn/go-ecommerce/internal/models"
	"github.com/zellis-rameesn/go-ecommerce/internal/utils"
	"gorm.io/gorm"
)

type OrderService struct {
	db *gorm.DB
}

func NewOrderService(db *gorm.DB) *OrderService {
	return &OrderService{
		db: db,
	}
}

func (o *OrderService) CreateOrder(userID uint) (dto.OrderResponse, error) {
	var orderResponse dto.OrderResponse
	err := o.db.Transaction(func(tx *gorm.DB) error {
		var cart models.Cart
		if err := tx.Preload("CartItems.Product.Category").Where("user_id = ?", userID).First(&cart).Error; err != nil {
			return err
		}
		if len(cart.CartItems) == 0 {
			return errors.New("no items in cart")
		}

		var orderItems []models.OrderItem
		var orderTotal float64

		for i := range cart.CartItems {
			cartItem := cart.CartItems[i]
			if cartItem.Quantity > cartItem.Product.Stock {
				return errors.New("insufficient stock")
			}
			cartItemTotal := float64(cartItem.Quantity) * cartItem.Product.Price
			orderTotal += cartItemTotal
			orderItems = append(orderItems, models.OrderItem{
				ProductID: cartItem.ProductID,
				Quantity:  cartItem.Quantity,
				Price:     cartItemTotal,
			})

			cartItem.Product.Stock -= cartItem.Quantity
			if err := tx.Save(&cartItem).Error; err != nil {
				return err
			}
		}
		order := models.Order{
			UserID:      userID,
			TotalAmount: orderTotal,
			OrderItems:  orderItems,
		}
		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		if err := tx.Where("cart_id = ?", cart.ID).Delete(&models.CartItem{}).Error; err != nil {
			return err
		}

		response, err := o.getOrderResponse(tx, order.ID)
		if err != nil {
			return err
		}
		orderResponse = response

		return nil
	})

	if err != nil {
		return dto.OrderResponse{}, err
	}

	return orderResponse, nil
}

func (o *OrderService) getOrderResponse(tx *gorm.DB, orderID uint) (dto.OrderResponse, error) {
	var order models.Order
	if err := tx.Preload("OrderItems.Product.Category").First(&order, orderID).Error; err != nil {
		return dto.OrderResponse{}, err
	}
	orderDto := o.convertToOrderResponse(&order)
	return orderDto, nil
}

func (o *OrderService) GetOrders(userID uint, limit, page int) (*[]dto.OrderResponse, utils.PaginationMeta, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	offset := limit * (page - 1)
	var count int64
	o.db.Model(models.Order{}).Where("user_id = ?", userID).Count(&count)

	totalPages := math.Ceil(float64(count) / float64(limit))

	var orders []models.Order
	if err := o.db.Preload("OrderItems.Product.Category").Where("user_id = ?", userID).Order("created_at DESC").Offset(offset).Limit(limit).Find(&orders).Error; err != nil {
		return nil, utils.PaginationMeta{}, err
	}

	meta := utils.PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      count,
		TotalPages: int(totalPages),
	}
	ordersSlice := make([]dto.OrderResponse, len(orders))
	for i := range orders {
		order := orders[i]
		ordersSlice[i] = o.convertToOrderResponse(&order)

	}
	return &ordersSlice, meta, nil
}

func (o *OrderService) GetOrder(userID, orderID uint) (dto.OrderResponse, error) {
	var order models.Order
	if err := o.db.Preload("OrderItems.Product.Category").Where("user_id = ? AND id = ?", userID, orderID).First(&order).Error; err != nil {
		return dto.OrderResponse{}, err
	}

	return o.convertToOrderResponse(&order), nil
}

func (o *OrderService) convertToOrderResponse(order *models.Order) dto.OrderResponse {
	orderItems := make([]dto.OrderItemResponse, len(order.OrderItems))

	for i := range order.OrderItems {
		item := order.OrderItems[i]
		orderItems[i] = dto.OrderItemResponse{
			ID:       item.ID,
			Price:    item.Price,
			Quantity: item.Quantity,
			Product: dto.ProductResponse{
				ID:          item.Product.ID,
				CategoryID:  item.Product.CategoryID,
				Name:        item.Product.Name,
				Description: item.Product.Description,
				Price:       item.Product.Price,
				Stock:       item.Product.Stock,
				SKU:         item.Product.SKU,
				IsActive:    item.Product.IsActive,
				Category: dto.CategoryResponse{
					ID:          item.Product.Category.ID,
					Name:        item.Product.Category.Name,
					Description: item.Product.Category.Description,
					IsActive:    item.Product.Category.IsActive,
				},
			},
		}
	}

	return dto.OrderResponse{
		ID:          order.ID,
		UserID:      order.UserID,
		Status:      order.Status,
		TotalAmount: order.TotalAmount,
		OrderItems:  orderItems,
		CreatedAt:   order.CreatedAt.String(),
	}
}

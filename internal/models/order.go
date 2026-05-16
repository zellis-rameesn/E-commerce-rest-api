package models

import "gorm.io/gorm"

type Order struct {
	gorm.Model
	UserID      uint        `json:"user_id" gorm:"not null"`
	Status      OrderStatus `json:"status" gorm:"default:pending"`
	TotalAmount float64     `json:"total_amount" gorm:"not null"`
	// Relationships
	OrderItems []OrderItem `json:"order_items"`
}

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type OrderItem struct {
	gorm.Model
	OrderID   uint `json:"order_id" gorm:"not null"`
	ProductID uint `json:"product_id" gorm:"not null"`
	Price     uint `json:"price" gorm:"not null"`

	// Relationships
	Order   Order   `json:"-"`
	Product Product `json:"-"`
}

type Cart struct {
	gorm.Model
	UserID uint `json:"user_id" gorm:"not null"`

	// Relationships
	CartItems []CartItem `json:"cart_items"`
}

type CartItem struct {
	gorm.Model
	CartID    uint `json:"cart_id" gorm:"not null"`
	ProductID uint `json:"product_id" gorm:"not null"`
	Quantity  uint `json:"quantity" gorm:"not null"`
	// Relationships
	Product Product
}

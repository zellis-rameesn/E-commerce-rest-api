package server

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zellis-rameesn/go-ecommerce/internal/utils"
)

func (s *Server) createOrder(c *gin.Context) {
	userID := c.GetUint("user_id")

	orderResponse, err := s.OrderService.CreateOrder(userID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create order", err)
		return
	}
	utils.CreatedResponse(c, "Order created", orderResponse)
}

func (s *Server) getOrders(c *gin.Context) {
	userID := c.GetUint("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	orderResponse, meta, err := s.OrderService.GetOrders(userID, limit, page)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to fetch orders", err)
		return
	}
	utils.PaginationResponse(c, "Fetched orders", orderResponse, meta)
}

func (s *Server) getOrder(c *gin.Context) {
	userID := c.GetUint("user_id")
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid order id!", err)
		return
	}

	orderResponse, err := s.OrderService.GetOrder(userID, uint(orderID))
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to fetch orders", err)
		return
	}
	utils.SuccessResponse(c, "Fetched order", orderResponse)
}

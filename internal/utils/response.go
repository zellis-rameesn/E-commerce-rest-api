package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	Error   string `json:"error"`
}

type PaginationReponse struct {
	Response
	Meta PaginationMeta
}

type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

func SuccessResponse(c *gin.Context, msg string, data any) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: msg,
		Data:    data,
	})
}

func CreatedResponse(c *gin.Context, msg string, data any) {
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Message: msg,
		Data:    data,
	})
}

func PaginationResponse(c *gin.Context, msg string, data any, meta PaginationMeta) {
	c.JSON(http.StatusOK, PaginationReponse{
		Response: Response{
			Success: true,
			Message: msg,
			Data:    data,
		},
		Meta: meta,
	})
}

func ErrorResponse(c *gin.Context, statusCode int, msg string, err error) {
	response := Response{
		Success: false,
		Message: msg,
	}
	if err != nil {
		response.Error = err.Error()
	}
	c.JSON(statusCode, response)
}

func BadRequestResponse(c *gin.Context, statusCode int, msg string, err error) {
	ErrorResponse(c, http.StatusBadRequest, msg, err)
}

func UnauthorizedResponse(c *gin.Context, statusCode int, msg string, err error) {
	ErrorResponse(c, http.StatusUnauthorized, msg, err)
}

func ForbiddenResponse(c *gin.Context, statusCode int, msg string, err error) {
	ErrorResponse(c, http.StatusForbidden, msg, err)
}

func NotFoundResponse(c *gin.Context, statusCode int, msg string, err error) {
	ErrorResponse(c, http.StatusNotFound, msg, err)
}

func InternalServerErrorResponse(c *gin.Context, statusCode int, msg string, err error) {
	ErrorResponse(c, http.StatusInternalServerError, msg, err)
}

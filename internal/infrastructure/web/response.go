// Package web provides HTTP response utilities for OpenHost.
// It includes helpers for JSON responses, error handling, and common
// response patterns used throughout the application.
package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response codes for standardized API responses
const (
	CodeSuccess       = 0
	CodeBadRequest    = 400
	CodeUnauthorized  = 401
	CodeForbidden     = 403
	CodeNotFound      = 404
	CodeConflict      = 409
	CodeValidation    = 422
	CodeTooMany       = 429
	CodeServerError   = 500
	CodeServiceUnavailable = 503
)

// APIResponse represents a standard API response structure
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
	Meta    *MetaData   `json:"meta,omitempty"`
}

// MetaData contains pagination and additional response metadata
type MetaData struct {
	Page       int   `json:"page,omitempty"`
	PerPage    int   `json:"per_page,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
	HasMore    bool  `json:"has_more,omitempty"`
}

// PaginatedData wraps data with pagination metadata
type PaginatedData struct {
	Items      interface{} `json:"items"`
	Pagination *MetaData   `json:"pagination"`
}

// Responder provides standardized HTTP response methods
type Responder struct{}

// NewResponder creates a new Responder instance
func NewResponder() *Responder {
	return &Responder{}
}

// DefaultResponder is the global responder instance
var DefaultResponder = NewResponder()

// Success sends a successful JSON response
func (r *Responder) Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

// SuccessWithMessage sends a successful JSON response with custom message
func (r *Responder) SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	})
}

// Created sends a 201 Created response
func (r *Responder) Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, APIResponse{
		Code:    CodeSuccess,
		Message: "created",
		Data:    data,
	})
}

// NoContent sends a 204 No Content response
func (r *Responder) NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Paginated sends a paginated JSON response
func (r *Responder) Paginated(c *gin.Context, items interface{}, page, perPage int, total int64) {
	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, APIResponse{
		Code:    CodeSuccess,
		Message: "success",
		Data:    items,
		Meta: &MetaData{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
			HasMore:    page < totalPages,
		},
	})
}

// Error sends an error response with the specified status code
func (r *Responder) Error(c *gin.Context, statusCode int, code int, message string) {
	c.JSON(statusCode, APIResponse{
		Code:    code,
		Message: message,
	})
}

// ErrorWithDetails sends an error response with additional details
func (r *Responder) ErrorWithDetails(c *gin.Context, statusCode int, code int, message string, errors interface{}) {
	c.JSON(statusCode, APIResponse{
		Code:    code,
		Message: message,
		Errors:  errors,
	})
}

// BadRequest sends a 400 Bad Request response
func (r *Responder) BadRequest(c *gin.Context, message string) {
	r.Error(c, http.StatusBadRequest, CodeBadRequest, message)
}

// BadRequestWithErrors sends a 400 Bad Request response with validation errors
func (r *Responder) BadRequestWithErrors(c *gin.Context, message string, errors interface{}) {
	r.ErrorWithDetails(c, http.StatusBadRequest, CodeBadRequest, message, errors)
}

// Unauthorized sends a 401 Unauthorized response
func (r *Responder) Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "unauthorized"
	}
	r.Error(c, http.StatusUnauthorized, CodeUnauthorized, message)
}

// Forbidden sends a 403 Forbidden response
func (r *Responder) Forbidden(c *gin.Context, message string) {
	if message == "" {
		message = "forbidden"
	}
	r.Error(c, http.StatusForbidden, CodeForbidden, message)
}

// NotFound sends a 404 Not Found response
func (r *Responder) NotFound(c *gin.Context, message string) {
	if message == "" {
		message = "not found"
	}
	r.Error(c, http.StatusNotFound, CodeNotFound, message)
}

// Conflict sends a 409 Conflict response
func (r *Responder) Conflict(c *gin.Context, message string) {
	r.Error(c, http.StatusConflict, CodeConflict, message)
}

// ValidationError sends a 422 Unprocessable Entity response
func (r *Responder) ValidationError(c *gin.Context, errors interface{}) {
	r.ErrorWithDetails(c, http.StatusUnprocessableEntity, CodeValidation, "validation failed", errors)
}

// TooManyRequests sends a 429 Too Many Requests response
func (r *Responder) TooManyRequests(c *gin.Context, message string) {
	if message == "" {
		message = "too many requests"
	}
	r.Error(c, http.StatusTooManyRequests, CodeTooMany, message)
}

// ServerError sends a 500 Internal Server Error response
func (r *Responder) ServerError(c *gin.Context, message string) {
	if message == "" {
		message = "internal server error"
	}
	r.Error(c, http.StatusInternalServerError, CodeServerError, message)
}

// ServiceUnavailable sends a 503 Service Unavailable response
func (r *Responder) ServiceUnavailable(c *gin.Context, message string) {
	if message == "" {
		message = "service unavailable"
	}
	r.Error(c, http.StatusServiceUnavailable, CodeServiceUnavailable, message)
}

// Redirect sends a redirect response
func (r *Responder) Redirect(c *gin.Context, location string) {
	c.Redirect(http.StatusFound, location)
}

// RedirectPermanent sends a permanent redirect response
func (r *Responder) RedirectPermanent(c *gin.Context, location string) {
	c.Redirect(http.StatusMovedPermanently, location)
}

// File sends a file download response
func (r *Responder) File(c *gin.Context, filepath string) {
	c.File(filepath)
}

// FileAttachment sends a file as attachment (download)
func (r *Responder) FileAttachment(c *gin.Context, filepath, filename string) {
	c.FileAttachment(filepath, filename)
}

// Global convenience functions

// Success sends a successful JSON response using the default responder
func Success(c *gin.Context, data interface{}) {
	DefaultResponder.Success(c, data)
}

// SuccessWithMessage sends a successful JSON response with message
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	DefaultResponder.SuccessWithMessage(c, message, data)
}

// Created sends a 201 Created response using the default responder
func Created(c *gin.Context, data interface{}) {
	DefaultResponder.Created(c, data)
}

// NoContent sends a 204 No Content response
func NoContent(c *gin.Context) {
	DefaultResponder.NoContent(c)
}

// Paginated sends a paginated JSON response
func Paginated(c *gin.Context, items interface{}, page, perPage int, total int64) {
	DefaultResponder.Paginated(c, items, page, perPage, total)
}

// BadRequest sends a 400 Bad Request response
func BadRequest(c *gin.Context, message string) {
	DefaultResponder.BadRequest(c, message)
}

// BadRequestWithErrors sends a 400 Bad Request with validation errors
func BadRequestWithErrors(c *gin.Context, message string, errors interface{}) {
	DefaultResponder.BadRequestWithErrors(c, message, errors)
}

// Unauthorized sends a 401 Unauthorized response
func Unauthorized(c *gin.Context, message string) {
	DefaultResponder.Unauthorized(c, message)
}

// Forbidden sends a 403 Forbidden response
func Forbidden(c *gin.Context, message string) {
	DefaultResponder.Forbidden(c, message)
}

// NotFound sends a 404 Not Found response
func NotFound(c *gin.Context, message string) {
	DefaultResponder.NotFound(c, message)
}

// Conflict sends a 409 Conflict response
func Conflict(c *gin.Context, message string) {
	DefaultResponder.Conflict(c, message)
}

// ValidationError sends a 422 Unprocessable Entity response
func ValidationError(c *gin.Context, errors interface{}) {
	DefaultResponder.ValidationError(c, errors)
}

// TooManyRequests sends a 429 Too Many Requests response
func TooManyRequests(c *gin.Context, message string) {
	DefaultResponder.TooManyRequests(c, message)
}

// ServerError sends a 500 Internal Server Error response
func ServerError(c *gin.Context, message string) {
	DefaultResponder.ServerError(c, message)
}

// ServiceUnavailable sends a 503 Service Unavailable response
func ServiceUnavailable(c *gin.Context, message string) {
	DefaultResponder.ServiceUnavailable(c, message)
}

// Redirect sends a redirect response
func Redirect(c *gin.Context, location string) {
	DefaultResponder.Redirect(c, location)
}

// RedirectPermanent sends a permanent redirect response
func RedirectPermanent(c *gin.Context, location string) {
	DefaultResponder.RedirectPermanent(c, location)
}

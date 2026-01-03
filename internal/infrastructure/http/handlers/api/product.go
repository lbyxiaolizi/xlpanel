package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/openhost/openhost/internal/core/service/product"
)

// ProductHandler handles product API endpoints
type ProductHandler struct {
	productService *product.Service
}

// NewProductHandler creates a new product handler
func NewProductHandler(productService *product.Service) *ProductHandler {
	return &ProductHandler{productService: productService}
}

// ListProductGroups godoc
// @Summary List product groups
// @Description Returns all product groups
// @Tags products
// @Produce json
// @Param active query bool false "Filter by active status"
// @Success 200 {array} ProductGroupResponse
// @Router /api/v1/products/groups [get]
func (h *ProductHandler) ListProductGroups(c *gin.Context) {
	activeOnly := c.Query("active") == "true"

	groups, err := h.productService.ListProductGroups(activeOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch product groups"})
		return
	}

	var response []ProductGroupResponse
	for _, group := range groups {
		response = append(response, ProductGroupResponse{
			ID:          group.ID,
			Name:        group.Name,
			Slug:        group.Slug,
			Description: group.Description,
			Active:      group.Active,
		})
	}

	c.JSON(http.StatusOK, response)
}

// GetProductGroup godoc
// @Summary Get product group
// @Description Returns a product group with its products
// @Tags products
// @Produce json
// @Param slug path string true "Group slug"
// @Success 200 {object} ProductGroupDetailResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/products/groups/{slug} [get]
func (h *ProductHandler) GetProductGroup(c *gin.Context) {
	slug := c.Param("slug")

	group, err := h.productService.GetProductGroupBySlug(slug)
	if err != nil {
		if err == product.ErrProductGroupNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Product group not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch product group"})
		return
	}

	var products []ProductResponse
	for _, p := range group.Products {
		if !p.Active {
			continue
		}
		products = append(products, ProductResponse{
			ID:          p.ID,
			Name:        p.Name,
			Slug:        p.Slug,
			Description: p.Description,
		})
	}

	c.JSON(http.StatusOK, ProductGroupDetailResponse{
		ID:          group.ID,
		Name:        group.Name,
		Slug:        group.Slug,
		Description: group.Description,
		Active:      group.Active,
		Products:    products,
	})
}

// ListProducts godoc
// @Summary List products
// @Description Returns products with optional filtering
// @Tags products
// @Produce json
// @Param group query int false "Filter by group ID"
// @Param active query bool false "Filter by active status"
// @Param limit query int false "Number of results per page" default(20)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} PaginatedResponse
// @Router /api/v1/products [get]
func (h *ProductHandler) ListProducts(c *gin.Context) {
	limit, offset := PaginationParams(c)
	activeOnly := c.Query("active") != "false"

	var groupID *uint64
	if gid := c.Query("group"); gid != "" {
		if parsed, err := strconv.ParseUint(gid, 10, 64); err == nil {
			groupID = &parsed
		}
	}

	products, total, err := h.productService.ListProducts(groupID, activeOnly, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch products"})
		return
	}

	var response []ProductResponse
	for _, p := range products {
		response = append(response, ProductResponse{
			ID:          p.ID,
			Name:        p.Name,
			Slug:        p.Slug,
			Description: p.Description,
		})
	}

	c.JSON(http.StatusOK, NewPaginatedResponse(response, total, limit, offset))
}

// GetProduct godoc
// @Summary Get product details
// @Description Returns detailed product information including configuration options
// @Tags products
// @Produce json
// @Param slug path string true "Product slug"
// @Success 200 {object} ProductDetailResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/products/{slug} [get]
func (h *ProductHandler) GetProduct(c *gin.Context) {
	slug := c.Param("slug")

	p, err := h.productService.GetProductBySlug(slug)
	if err != nil {
		if err == product.ErrProductNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch product"})
		return
	}

	var configGroups []ConfigGroupResponse
	for _, cg := range p.ConfigGroups {
		var options []ConfigOptionResponse
		for _, opt := range cg.Options {
			var subOptions []ConfigSubOptionResponse
			for _, sub := range opt.SubOptions {
				subOptions = append(subOptions, ConfigSubOptionResponse{
					ID:   sub.ID,
					Name: sub.Name,
					Pricing: PricingResponse{
						SetupFee:    sub.Pricing.SetupFee.String(),
						Monthly:     sub.Pricing.Monthly.String(),
						Quarterly:   sub.Pricing.Quarterly.String(),
						Yearly:      sub.Pricing.Yearly.String(),
						Triennially: sub.Pricing.Triennially.String(),
					},
				})
			}
			options = append(options, ConfigOptionResponse{
				ID:         opt.ID,
				Name:       opt.Name,
				InputType:  opt.InputType,
				Required:   opt.Required,
				SubOptions: subOptions,
			})
		}
		configGroups = append(configGroups, ConfigGroupResponse{
			ID:          cg.ID,
			Name:        cg.Name,
			Description: cg.Description,
			Options:     options,
		})
	}

	c.JSON(http.StatusOK, ProductDetailResponse{
		ID:           p.ID,
		Name:         p.Name,
		Slug:         p.Slug,
		Description:  p.Description,
		ModuleName:   p.ModuleName,
		ConfigGroups: configGroups,
	})
}

// GetProductPricing godoc
// @Summary Calculate product pricing
// @Description Calculates pricing for a product with selected options
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Param request body PricingCalculationRequest true "Pricing parameters"
// @Success 200 {object} PricingCalculationResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/products/{id}/pricing [post]
func (h *ProductHandler) GetProductPricing(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid product ID"})
		return
	}

	var req PricingCalculationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	result, err := h.productService.GetProductPricing(productID, req.BillingCycle, req.SelectedOptions)
	if err != nil {
		if err == product.ErrProductNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to calculate pricing"})
		return
	}

	c.JSON(http.StatusOK, PricingCalculationResponse{
		ProductID:    result.ProductID,
		ProductName:  result.ProductName,
		BillingCycle: result.BillingCycle,
		SetupFee:     result.SetupFee.String(),
		RecurringFee: result.RecurringFee.String(),
		Total:        result.Total.String(),
	})
}

// Admin endpoints

// CreateProductGroup godoc
// @Summary Create product group (Admin)
// @Description Creates a new product group
// @Tags admin/products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateProductGroupRequest true "Product group data"
// @Success 201 {object} ProductGroupResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /api/v1/admin/products/groups [post]
func (h *ProductHandler) CreateProductGroup(c *gin.Context) {
	var req CreateProductGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	group, err := h.productService.CreateProductGroup(req.Name, req.Slug, req.Description, req.SortOrder, req.Active)
	if err != nil {
		if err == product.ErrSlugExists {
			c.JSON(http.StatusConflict, ErrorResponse{Error: "Slug already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create product group"})
		return
	}

	c.JSON(http.StatusCreated, ProductGroupResponse{
		ID:          group.ID,
		Name:        group.Name,
		Slug:        group.Slug,
		Description: group.Description,
		Active:      group.Active,
	})
}

// CreateProduct godoc
// @Summary Create product (Admin)
// @Description Creates a new product
// @Tags admin/products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateProductRequest true "Product data"
// @Success 201 {object} ProductResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /api/v1/admin/products [post]
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	p, err := h.productService.CreateProduct(req.GroupID, req.Name, req.Slug, req.Description, req.ModuleName, req.Active)
	if err != nil {
		if err == product.ErrSlugExists {
			c.JSON(http.StatusConflict, ErrorResponse{Error: "Slug already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create product"})
		return
	}

	c.JSON(http.StatusCreated, ProductResponse{
		ID:          p.ID,
		Name:        p.Name,
		Slug:        p.Slug,
		Description: p.Description,
	})
}

// UpdateProduct godoc
// @Summary Update product (Admin)
// @Description Updates an existing product
// @Tags admin/products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Product ID"
// @Param request body UpdateProductRequest true "Product data"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/admin/products/{id} [put]
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid product ID"})
		return
	}

	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if err := h.productService.UpdateProduct(productID, req.Name, req.Description, req.ModuleName, req.Active); err != nil {
		if err == product.ErrProductNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update product"})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "Product updated successfully"})
}

// DeleteProduct godoc
// @Summary Delete product (Admin)
// @Description Deletes a product
// @Tags admin/products
// @Produce json
// @Security BearerAuth
// @Param id path int true "Product ID"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/admin/products/{id} [delete]
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid product ID"})
		return
	}

	if err := h.productService.DeleteProduct(productID); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "Product deleted successfully"})
}

// Response types

type ProductGroupResponse struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Active      bool   `json:"active"`
}

type ProductGroupDetailResponse struct {
	ID          uint64            `json:"id"`
	Name        string            `json:"name"`
	Slug        string            `json:"slug"`
	Description string            `json:"description"`
	Active      bool              `json:"active"`
	Products    []ProductResponse `json:"products"`
}

type ProductResponse struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

type ProductDetailResponse struct {
	ID           uint64                `json:"id"`
	Name         string                `json:"name"`
	Slug         string                `json:"slug"`
	Description  string                `json:"description"`
	ModuleName   string                `json:"module_name"`
	ConfigGroups []ConfigGroupResponse `json:"config_groups"`
}

type ConfigGroupResponse struct {
	ID          uint64                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Options     []ConfigOptionResponse `json:"options"`
}

type ConfigOptionResponse struct {
	ID         uint64                   `json:"id"`
	Name       string                   `json:"name"`
	InputType  string                   `json:"input_type"`
	Required   bool                     `json:"required"`
	SubOptions []ConfigSubOptionResponse `json:"sub_options"`
}

type ConfigSubOptionResponse struct {
	ID      uint64          `json:"id"`
	Name    string          `json:"name"`
	Pricing PricingResponse `json:"pricing"`
}

type PricingResponse struct {
	SetupFee    string `json:"setup_fee"`
	Monthly     string `json:"monthly"`
	Quarterly   string `json:"quarterly"`
	Yearly      string `json:"yearly"`
	Triennially string `json:"triennially"`
}

type PricingCalculationRequest struct {
	BillingCycle    string            `json:"billing_cycle" binding:"required"`
	SelectedOptions map[uint64]uint64 `json:"selected_options"`
}

type PricingCalculationResponse struct {
	ProductID    uint64 `json:"product_id"`
	ProductName  string `json:"product_name"`
	BillingCycle string `json:"billing_cycle"`
	SetupFee     string `json:"setup_fee"`
	RecurringFee string `json:"recurring_fee"`
	Total        string `json:"total"`
}

type CreateProductGroupRequest struct {
	Name        string `json:"name" binding:"required"`
	Slug        string `json:"slug" binding:"required"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
	Active      bool   `json:"active"`
}

type CreateProductRequest struct {
	GroupID     uint64 `json:"group_id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Slug        string `json:"slug" binding:"required"`
	Description string `json:"description"`
	ModuleName  string `json:"module_name" binding:"required"`
	Active      bool   `json:"active"`
}

type UpdateProductRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	ModuleName  string `json:"module_name" binding:"required"`
	Active      bool   `json:"active"`
}

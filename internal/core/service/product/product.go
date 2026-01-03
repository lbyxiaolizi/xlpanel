package product

import (
	"errors"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/openhost/openhost/internal/core/domain"
)

var (
	ErrProductNotFound      = errors.New("product not found")
	ErrProductGroupNotFound = errors.New("product group not found")
	ErrConfigGroupNotFound  = errors.New("config group not found")
	ErrSlugExists           = errors.New("slug already exists")
)

// Service provides product management operations
type Service struct {
	db *gorm.DB
}

// NewService creates a new product service
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// CreateProductGroup creates a new product group
func (s *Service) CreateProductGroup(name, slug, description string, sortOrder int, active bool) (*domain.ProductGroup, error) {
	// Check if slug already exists
	var existing domain.ProductGroup
	if err := s.db.Where("slug = ?", slug).First(&existing).Error; err == nil {
		return nil, ErrSlugExists
	}

	group := &domain.ProductGroup{
		Name:        name,
		Slug:        slug,
		Description: description,
		SortOrder:   sortOrder,
		Active:      active,
	}

	if err := s.db.Create(group).Error; err != nil {
		return nil, err
	}

	return group, nil
}

// GetProductGroup retrieves a product group by ID
func (s *Service) GetProductGroup(id uint64) (*domain.ProductGroup, error) {
	var group domain.ProductGroup
	if err := s.db.Preload("Products").First(&group, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductGroupNotFound
		}
		return nil, err
	}
	return &group, nil
}

// GetProductGroupBySlug retrieves a product group by slug
func (s *Service) GetProductGroupBySlug(slug string) (*domain.ProductGroup, error) {
	var group domain.ProductGroup
	if err := s.db.Preload("Products").Where("slug = ?", slug).First(&group).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductGroupNotFound
		}
		return nil, err
	}
	return &group, nil
}

// ListProductGroups returns all product groups
func (s *Service) ListProductGroups(activeOnly bool) ([]domain.ProductGroup, error) {
	var groups []domain.ProductGroup
	query := s.db.Order("sort_order ASC, name ASC")
	if activeOnly {
		query = query.Where("active = ?", true)
	}
	if err := query.Find(&groups).Error; err != nil {
		return nil, err
	}
	return groups, nil
}

// UpdateProductGroup updates a product group
func (s *Service) UpdateProductGroup(id uint64, name, description string, sortOrder int, active bool) error {
	updates := map[string]interface{}{
		"name":        name,
		"description": description,
		"sort_order":  sortOrder,
		"active":      active,
	}
	return s.db.Model(&domain.ProductGroup{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteProductGroup deletes a product group
func (s *Service) DeleteProductGroup(id uint64) error {
	// Check if group has products
	var count int64
	s.db.Model(&domain.Product{}).Where("product_group_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("cannot delete group with products")
	}
	return s.db.Delete(&domain.ProductGroup{}, id).Error
}

// CreateProduct creates a new product
func (s *Service) CreateProduct(groupID uint64, name, slug, description, moduleName string, active bool) (*domain.Product, error) {
	// Check if slug already exists
	var existing domain.Product
	if err := s.db.Where("slug = ?", slug).First(&existing).Error; err == nil {
		return nil, ErrSlugExists
	}

	product := &domain.Product{
		ProductGroupID: groupID,
		Name:           name,
		Slug:           slug,
		Description:    description,
		ModuleName:     moduleName,
		Active:         active,
	}

	if err := s.db.Create(product).Error; err != nil {
		return nil, err
	}

	return product, nil
}

// GetProduct retrieves a product by ID
func (s *Service) GetProduct(id uint64) (*domain.Product, error) {
	var product domain.Product
	if err := s.db.Preload("ConfigGroups.Options.SubOptions").First(&product, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	return &product, nil
}

// GetProductBySlug retrieves a product by slug
func (s *Service) GetProductBySlug(slug string) (*domain.Product, error) {
	var product domain.Product
	if err := s.db.Preload("ConfigGroups.Options.SubOptions").
		Where("slug = ?", slug).First(&product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	return &product, nil
}

// ListProducts returns products with optional filters
func (s *Service) ListProducts(groupID *uint64, activeOnly bool, limit, offset int) ([]domain.Product, int64, error) {
	var products []domain.Product
	var total int64

	query := s.db.Model(&domain.Product{})
	if groupID != nil {
		query = query.Where("product_group_id = ?", *groupID)
	}
	if activeOnly {
		query = query.Where("active = ?", true)
	}
	query.Count(&total)

	if err := query.Order("name ASC").Limit(limit).Offset(offset).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// UpdateProduct updates a product
func (s *Service) UpdateProduct(id uint64, name, description, moduleName string, active bool) error {
	updates := map[string]interface{}{
		"name":        name,
		"description": description,
		"module_name": moduleName,
		"active":      active,
	}
	return s.db.Model(&domain.Product{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteProduct deletes a product
func (s *Service) DeleteProduct(id uint64) error {
	// Check if product has active services
	var count int64
	s.db.Model(&domain.Service{}).Where("product_id = ? AND status != ?", id, domain.ServiceStatusTerminated).Count(&count)
	if count > 0 {
		return errors.New("cannot delete product with active services")
	}
	return s.db.Delete(&domain.Product{}, id).Error
}

// CreateConfigGroup creates a new configuration group
func (s *Service) CreateConfigGroup(name, description string) (*domain.ConfigGroup, error) {
	group := &domain.ConfigGroup{
		Name:        name,
		Description: description,
	}

	if err := s.db.Create(group).Error; err != nil {
		return nil, err
	}

	return group, nil
}

// AssignConfigGroupToProduct assigns a config group to a product
func (s *Service) AssignConfigGroupToProduct(productID, configGroupID uint64) error {
	pcg := &domain.ProductConfigGroup{
		ProductID:     productID,
		ConfigGroupID: configGroupID,
	}
	return s.db.Create(pcg).Error
}

// RemoveConfigGroupFromProduct removes a config group from a product
func (s *Service) RemoveConfigGroupFromProduct(productID, configGroupID uint64) error {
	return s.db.Where("product_id = ? AND config_group_id = ?", productID, configGroupID).
		Delete(&domain.ProductConfigGroup{}).Error
}

// CreateConfigOption creates a configuration option
func (s *Service) CreateConfigOption(groupID uint64, name, inputType string, required bool, sortOrder int) (*domain.ConfigOption, error) {
	option := &domain.ConfigOption{
		ConfigGroupID: groupID,
		Name:          name,
		InputType:     inputType,
		Required:      required,
		SortOrder:     sortOrder,
	}

	if err := s.db.Create(option).Error; err != nil {
		return nil, err
	}

	return option, nil
}

// CreateConfigSubOption creates a sub-option with pricing
func (s *Service) CreateConfigSubOption(optionID uint64, name string, sortOrder int, pricing PricingRequest) (*domain.ConfigSubOption, error) {
	subOption := &domain.ConfigSubOption{
		ConfigOptionID: optionID,
		Name:           name,
		SortOrder:      sortOrder,
		Pricing: domain.Pricing{
			SetupFee:    pricing.SetupFee,
			Monthly:     pricing.Monthly,
			Quarterly:   pricing.Quarterly,
			Yearly:      pricing.Yearly,
			Triennially: pricing.Triennially,
		},
	}

	if err := s.db.Create(subOption).Error; err != nil {
		return nil, err
	}

	return subOption, nil
}

// GetProductPricing returns pricing for a product based on selected options
func (s *Service) GetProductPricing(productID uint64, billingCycle string, selectedOptions map[uint64]uint64) (*ProductPricingResult, error) {
	product, err := s.GetProduct(productID)
	if err != nil {
		return nil, err
	}

	result := &ProductPricingResult{
		ProductID:    productID,
		ProductName:  product.Name,
		BillingCycle: billingCycle,
		SetupFee:     decimal.Zero,
		RecurringFee: decimal.Zero,
	}

	// Calculate pricing from selected options
	for _, configGroup := range product.ConfigGroups {
		for _, option := range configGroup.Options {
			selectedSubOptionID, ok := selectedOptions[option.ID]
			if !ok {
				continue
			}

			for _, subOption := range option.SubOptions {
				if subOption.ID == selectedSubOptionID {
					result.SetupFee = result.SetupFee.Add(subOption.Pricing.SetupFee)

					switch billingCycle {
					case "monthly":
						result.RecurringFee = result.RecurringFee.Add(subOption.Pricing.Monthly)
					case "quarterly":
						result.RecurringFee = result.RecurringFee.Add(subOption.Pricing.Quarterly)
					case "yearly", "annually":
						result.RecurringFee = result.RecurringFee.Add(subOption.Pricing.Yearly)
					case "triennially":
						result.RecurringFee = result.RecurringFee.Add(subOption.Pricing.Triennially)
					default:
						result.RecurringFee = result.RecurringFee.Add(subOption.Pricing.Monthly)
					}

					result.SelectedOptions = append(result.SelectedOptions, SelectedOptionDetail{
						OptionID:      option.ID,
						OptionName:    option.Name,
						SubOptionID:   subOption.ID,
						SubOptionName: subOption.Name,
					})
				}
			}
		}
	}

	result.Total = result.SetupFee.Add(result.RecurringFee)
	return result, nil
}

// GetFeaturedProducts returns featured/popular products
func (s *Service) GetFeaturedProducts(limit int) ([]domain.Product, error) {
	var products []domain.Product
	// For now, just return active products. Could add a "featured" flag or use order count
	if err := s.db.Where("active = ?", true).Limit(limit).Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

// PricingRequest represents a pricing request
type PricingRequest struct {
	SetupFee    decimal.Decimal
	Monthly     decimal.Decimal
	Quarterly   decimal.Decimal
	Yearly      decimal.Decimal
	Triennially decimal.Decimal
}

// ProductPricingResult represents the calculated pricing for a product
type ProductPricingResult struct {
	ProductID       uint64                 `json:"product_id"`
	ProductName     string                 `json:"product_name"`
	BillingCycle    string                 `json:"billing_cycle"`
	SetupFee        decimal.Decimal        `json:"setup_fee"`
	RecurringFee    decimal.Decimal        `json:"recurring_fee"`
	Total           decimal.Decimal        `json:"total"`
	SelectedOptions []SelectedOptionDetail `json:"selected_options"`
}

// SelectedOptionDetail represents a selected configuration option
type SelectedOptionDetail struct {
	OptionID      uint64 `json:"option_id"`
	OptionName    string `json:"option_name"`
	SubOptionID   uint64 `json:"sub_option_id"`
	SubOptionName string `json:"sub_option_name"`
}

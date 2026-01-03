package service

import (
	"errors"
	"time"

	"xlpanel/internal/core"
	"xlpanel/internal/domain"
	"xlpanel/internal/infra"
)

var (
	ErrCartNotFound     = errors.New("cart not found")
	ErrProductNotFound  = errors.New("product not found")
	ErrInvalidQuantity  = errors.New("invalid quantity")
)

type CartService struct {
	carts    *infra.Repository[domain.Cart]
	products *infra.Repository[domain.Product]
}

func NewCartService(carts *infra.Repository[domain.Cart], products *infra.Repository[domain.Product]) *CartService {
	return &CartService{
		carts:    carts,
		products: products,
	}
}

// GetOrCreateCart gets an existing cart or creates a new one
func (s *CartService) GetOrCreateCart(customerID, tenantID string) *domain.Cart {
	// Try to find existing cart
	for _, cart := range s.carts.List() {
		if cart.CustomerID == customerID && cart.TenantID == tenantID {
			return &cart
		}
	}

	// Create new cart
	cart := domain.Cart{
		ID:         core.NewID(),
		CustomerID: customerID,
		TenantID:   tenantID,
		Items:      []domain.CartItem{},
		UpdatedAt:  time.Now().UTC(),
	}
	s.carts.Add(cart)
	return &cart
}

// AddItem adds an item to the cart
func (s *CartService) AddItem(cartID, productCode string, quantity int) (*domain.Cart, error) {
	if quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	cart, ok := s.carts.Get(cartID)
	if !ok {
		return nil, ErrCartNotFound
	}

	// Find product to get price
	product, ok := s.products.Get(productCode)
	if !ok {
		return nil, ErrProductNotFound
	}

	// Check if item already exists in cart
	found := false
	for i, item := range cart.Items {
		if item.ProductCode == productCode {
			cart.Items[i].Quantity += quantity
			found = true
			break
		}
	}

	// Add new item if not found
	if !found {
		cart.Items = append(cart.Items, domain.CartItem{
			ProductCode: productCode,
			Quantity:    quantity,
			UnitPrice:   product.UnitPrice,
		})
	}

	cart.UpdatedAt = time.Now().UTC()
	s.carts.Add(cart)
	return &cart, nil
}

// UpdateItem updates the quantity of an item in the cart
func (s *CartService) UpdateItem(cartID, productCode string, quantity int) (*domain.Cart, error) {
	if quantity < 0 {
		return nil, ErrInvalidQuantity
	}

	cart, ok := s.carts.Get(cartID)
	if !ok {
		return nil, ErrCartNotFound
	}

	// If quantity is 0, remove the item
	if quantity == 0 {
		return s.RemoveItem(cartID, productCode)
	}

	// Update quantity
	found := false
	for i, item := range cart.Items {
		if item.ProductCode == productCode {
			cart.Items[i].Quantity = quantity
			found = true
			break
		}
	}

	if !found {
		return nil, ErrProductNotFound
	}

	cart.UpdatedAt = time.Now().UTC()
	s.carts.Add(cart)
	return &cart, nil
}

// RemoveItem removes an item from the cart
func (s *CartService) RemoveItem(cartID, productCode string) (*domain.Cart, error) {
	cart, ok := s.carts.Get(cartID)
	if !ok {
		return nil, ErrCartNotFound
	}

	// Remove item
	newItems := []domain.CartItem{}
	for _, item := range cart.Items {
		if item.ProductCode != productCode {
			newItems = append(newItems, item)
		}
	}

	cart.Items = newItems
	cart.UpdatedAt = time.Now().UTC()
	s.carts.Add(cart)
	return &cart, nil
}

// ClearCart removes all items from the cart
func (s *CartService) ClearCart(cartID string) error {
	cart, ok := s.carts.Get(cartID)
	if !ok {
		return ErrCartNotFound
	}

	cart.Items = []domain.CartItem{}
	cart.UpdatedAt = time.Now().UTC()
	s.carts.Add(cart)
	return nil
}

// GetCart retrieves a cart by ID
func (s *CartService) GetCart(cartID string) (*domain.Cart, error) {
	cart, ok := s.carts.Get(cartID)
	if !ok {
		return nil, ErrCartNotFound
	}
	return &cart, nil
}

// CalculateTotal calculates the total price of items in the cart
func (s *CartService) CalculateTotal(cartID string) (float64, error) {
	cart, ok := s.carts.Get(cartID)
	if !ok {
		return 0, ErrCartNotFound
	}

	total := 0.0
	for _, item := range cart.Items {
		total += item.UnitPrice * float64(item.Quantity)
	}

	return total, nil
}

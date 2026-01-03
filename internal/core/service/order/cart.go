package order

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/openhost/openhost/internal/core/domain"
)

var (
	ErrCartNotFound    = errors.New("cart not found")
	ErrCartItemNotFound = errors.New("cart item not found")
)

const CartExpiration = 7 * 24 * time.Hour // 7 days

// CartService provides shopping cart operations
type CartService struct {
	db *gorm.DB
}

// NewCartService creates a new cart service
func NewCartService(db *gorm.DB) *CartService {
	return &CartService{db: db}
}

// GetOrCreateCart gets an existing cart or creates a new one
func (s *CartService) GetOrCreateCart(customerID *uint64, sessionID string) (*domain.Cart, error) {
	var cart domain.Cart
	var err error

	if customerID != nil {
		err = s.db.Preload("Items.Product").Where("customer_id = ?", *customerID).First(&cart).Error
	} else if sessionID != "" {
		err = s.db.Preload("Items.Product").Where("session_id = ?", sessionID).First(&cart).Error
	} else {
		return nil, errors.New("customer ID or session ID required")
	}

	if err == nil {
		// Check if cart is expired
		if time.Now().After(cart.ExpiresAt) {
			s.db.Delete(&domain.CartItem{}, "cart_id = ?", cart.ID)
			s.db.Delete(&cart)
			return s.createCart(customerID, sessionID)
		}
		return &cart, nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return s.createCart(customerID, sessionID)
	}

	return nil, err
}

// createCart creates a new cart
func (s *CartService) createCart(customerID *uint64, sessionID string) (*domain.Cart, error) {
	cart := &domain.Cart{
		CustomerID: customerID,
		SessionID:  sessionID,
		Currency:   "USD",
		ExpiresAt:  time.Now().Add(CartExpiration),
	}

	if err := s.db.Create(cart).Error; err != nil {
		return nil, err
	}

	return cart, nil
}

// AddItem adds a product to the cart
func (s *CartService) AddItem(cartID, productID uint64, quantity int, billingCycle, domainName, hostname string, configOptions domain.JSONMap) (*domain.CartItem, error) {
	if quantity <= 0 {
		quantity = 1
	}

	// Get product with pricing
	var product domain.Product
	if err := s.db.First(&product, productID).Error; err != nil {
		return nil, ErrProductNotFound
	}

	// Get pricing for the selected billing cycle
	// TODO: Get pricing from ConfigSubOption based on configOptions
	setupFee := decimal.NewFromFloat(0)
	recurringFee := decimal.NewFromFloat(29.99) // Default pricing

	// Check if item already exists in cart
	var existingItem domain.CartItem
	if err := s.db.Where("cart_id = ? AND product_id = ?", cartID, productID).First(&existingItem).Error; err == nil {
		// Update existing item
		existingItem.Quantity += quantity
		existingItem.Total = existingItem.SetupFee.Add(existingItem.RecurringFee.Mul(decimal.NewFromInt(int64(existingItem.Quantity))))
		if err := s.db.Save(&existingItem).Error; err != nil {
			return nil, err
		}
		return &existingItem, nil
	}

	// Create new cart item
	total := setupFee.Add(recurringFee.Mul(decimal.NewFromInt(int64(quantity))))

	item := &domain.CartItem{
		CartID:        cartID,
		ProductID:     productID,
		Quantity:      quantity,
		BillingCycle:  billingCycle,
		ConfigOptions: configOptions,
		Domain:        domainName,
		Hostname:      hostname,
		SetupFee:      setupFee,
		RecurringFee:  recurringFee,
		Discount:      decimal.Zero,
		Total:         total,
	}

	if err := s.db.Create(item).Error; err != nil {
		return nil, err
	}

	// Update cart expiration
	s.db.Model(&domain.Cart{}).Where("id = ?", cartID).Update("expires_at", time.Now().Add(CartExpiration))

	return item, nil
}

// UpdateItem updates a cart item
func (s *CartService) UpdateItem(cartItemID uint64, quantity int) (*domain.CartItem, error) {
	var item domain.CartItem
	if err := s.db.First(&item, cartItemID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCartItemNotFound
		}
		return nil, err
	}

	if quantity <= 0 {
		// Remove item if quantity is 0 or less
		return nil, s.RemoveItem(cartItemID)
	}

	item.Quantity = quantity
	item.Total = item.SetupFee.Add(item.RecurringFee.Mul(decimal.NewFromInt(int64(quantity)))).Sub(item.Discount)

	if err := s.db.Save(&item).Error; err != nil {
		return nil, err
	}

	return &item, nil
}

// RemoveItem removes an item from the cart
func (s *CartService) RemoveItem(cartItemID uint64) error {
	return s.db.Delete(&domain.CartItem{}, cartItemID).Error
}

// ApplyCoupon applies a coupon to the cart
func (s *CartService) ApplyCoupon(cartID uint64, couponCode string) error {
	var coupon domain.Coupon
	if err := s.db.Where("code = ?", couponCode).First(&coupon).Error; err != nil {
		return ErrInvalidCoupon
	}

	if !coupon.IsValid() {
		return ErrInvalidCoupon
	}

	// Update cart with coupon
	if err := s.db.Model(&domain.Cart{}).Where("id = ?", cartID).Update("coupon_id", coupon.ID).Error; err != nil {
		return err
	}

	// Recalculate item discounts
	return s.recalculateCartDiscounts(cartID, &coupon)
}

// RemoveCoupon removes the coupon from the cart
func (s *CartService) RemoveCoupon(cartID uint64) error {
	if err := s.db.Model(&domain.Cart{}).Where("id = ?", cartID).Update("coupon_id", nil).Error; err != nil {
		return err
	}

	// Remove discounts from all items
	return s.db.Model(&domain.CartItem{}).Where("cart_id = ?", cartID).
		Updates(map[string]interface{}{
			"discount": decimal.Zero,
		}).Error
}

// recalculateCartDiscounts recalculates discounts for all cart items
func (s *CartService) recalculateCartDiscounts(cartID uint64, coupon *domain.Coupon) error {
	var items []domain.CartItem
	if err := s.db.Where("cart_id = ?", cartID).Find(&items).Error; err != nil {
		return err
	}

	for _, item := range items {
		discount := decimal.Zero
		itemSubtotal := item.SetupFee.Add(item.RecurringFee.Mul(decimal.NewFromInt(int64(item.Quantity))))

		switch coupon.Type {
		case domain.CouponTypePercentage:
			discount = itemSubtotal.Mul(coupon.Amount).Div(decimal.NewFromInt(100))
		case domain.CouponTypeFixed:
			discount = coupon.Amount
		case domain.CouponTypeFreeSetup:
			discount = item.SetupFee
		}

		// Cap discount at item total
		if discount.GreaterThan(itemSubtotal) {
			discount = itemSubtotal
		}

		item.Discount = discount
		item.Total = itemSubtotal.Sub(discount)
		s.db.Save(&item)
	}

	return nil
}

// GetCartSummary returns a summary of the cart
func (s *CartService) GetCartSummary(cartID uint64) (*CartSummary, error) {
	var cart domain.Cart
	if err := s.db.Preload("Items.Product").Preload("Coupon").First(&cart, cartID).Error; err != nil {
		return nil, ErrCartNotFound
	}

	summary := &CartSummary{
		CartID:   cart.ID,
		Currency: cart.Currency,
		Items:    make([]CartItemSummary, 0, len(cart.Items)),
	}

	for _, item := range cart.Items {
		summary.Items = append(summary.Items, CartItemSummary{
			ID:           item.ID,
			ProductID:    item.ProductID,
			ProductName:  item.Product.Name,
			Quantity:     item.Quantity,
			BillingCycle: item.BillingCycle,
			SetupFee:     item.SetupFee,
			RecurringFee: item.RecurringFee,
			Discount:     item.Discount,
			Total:        item.Total,
		})
		summary.Subtotal = summary.Subtotal.Add(item.SetupFee.Add(item.RecurringFee.Mul(decimal.NewFromInt(int64(item.Quantity)))))
		summary.TotalDiscount = summary.TotalDiscount.Add(item.Discount)
		summary.Total = summary.Total.Add(item.Total)
	}

	if cart.Coupon != nil {
		summary.CouponCode = cart.Coupon.Code
	}

	return summary, nil
}

// ClearCart removes all items from a cart
func (s *CartService) ClearCart(cartID uint64) error {
	return s.db.Delete(&domain.CartItem{}, "cart_id = ?", cartID).Error
}

// MergeCart merges a guest cart into a user's cart
func (s *CartService) MergeCart(sessionID string, customerID uint64) error {
	// Get guest cart
	var guestCart domain.Cart
	if err := s.db.Where("session_id = ?", sessionID).First(&guestCart).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // No guest cart to merge
		}
		return err
	}

	// Get or create user cart
	userCart, err := s.GetOrCreateCart(&customerID, "")
	if err != nil {
		return err
	}

	// Move items from guest cart to user cart
	if err := s.db.Model(&domain.CartItem{}).Where("cart_id = ?", guestCart.ID).
		Update("cart_id", userCart.ID).Error; err != nil {
		return err
	}

	// Delete guest cart
	return s.db.Delete(&guestCart).Error
}

// CleanupExpiredCarts removes expired carts
func (s *CartService) CleanupExpiredCarts() error {
	var carts []domain.Cart
	if err := s.db.Where("expires_at < ?", time.Now()).Find(&carts).Error; err != nil {
		return err
	}

	for _, cart := range carts {
		s.db.Delete(&domain.CartItem{}, "cart_id = ?", cart.ID)
		s.db.Delete(&cart)
	}

	return nil
}

// CartSummary represents a summary of cart contents
type CartSummary struct {
	CartID        uint64            `json:"cart_id"`
	Currency      string            `json:"currency"`
	Items         []CartItemSummary `json:"items"`
	Subtotal      decimal.Decimal   `json:"subtotal"`
	TotalDiscount decimal.Decimal   `json:"total_discount"`
	Tax           decimal.Decimal   `json:"tax"`
	Total         decimal.Decimal   `json:"total"`
	CouponCode    string            `json:"coupon_code,omitempty"`
}

// CartItemSummary represents a summary of a cart item
type CartItemSummary struct {
	ID           uint64          `json:"id"`
	ProductID    uint64          `json:"product_id"`
	ProductName  string          `json:"product_name"`
	Quantity     int             `json:"quantity"`
	BillingCycle string          `json:"billing_cycle"`
	SetupFee     decimal.Decimal `json:"setup_fee"`
	RecurringFee decimal.Decimal `json:"recurring_fee"`
	Discount     decimal.Decimal `json:"discount"`
	Total        decimal.Decimal `json:"total"`
}

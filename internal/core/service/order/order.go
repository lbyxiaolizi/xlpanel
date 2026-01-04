package order

import (
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/openhost/openhost/internal/core/domain"
	"github.com/openhost/openhost/internal/core/service/tax"
)

var (
	ErrOrderNotFound   = errors.New("order not found")
	ErrServiceNotFound = errors.New("service not found")
	ErrProductNotFound = errors.New("product not found")
	ErrInvalidQuantity = errors.New("quantity must be greater than 0")
	ErrCartEmpty       = errors.New("cart is empty")
	ErrInvalidCoupon   = errors.New("invalid or expired coupon")
)

// Service provides order management operations
type Service struct {
	db *gorm.DB
}

// NewService creates a new order service
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// CreateOrder creates a new order from cart items
func (s *Service) CreateOrder(customerID uint64, cartID uint64, ipAddress string) (*domain.Order, error) {
	var cart domain.Cart
	if err := s.db.Preload("Items.Product").Preload("Coupon").First(&cart, cartID).Error; err != nil {
		return nil, err
	}

	if len(cart.Items) == 0 {
		return nil, ErrCartEmpty
	}

	// Calculate totals
	subtotal := decimal.Zero
	discount := decimal.Zero
	var orderItems []domain.OrderItem

	for _, item := range cart.Items {
		itemTotal := item.SetupFee.Add(item.RecurringFee).Mul(decimal.NewFromInt(int64(item.Quantity)))
		subtotal = subtotal.Add(itemTotal)
		discount = discount.Add(item.Discount)

		orderItems = append(orderItems, domain.OrderItem{
			ProductID:     item.ProductID,
			Description:   item.Product.Name,
			Quantity:      item.Quantity,
			BillingCycle:  item.BillingCycle,
			SetupFee:      item.SetupFee,
			RecurringFee:  item.RecurringFee,
			Discount:      item.Discount,
			Total:         itemTotal.Sub(item.Discount),
			ConfigOptions: item.ConfigOptions,
			Domain:        item.Domain,
			Hostname:      item.Hostname,
		})
	}

	taxableAmount := subtotal.Sub(discount)
	taxAmount, err := tax.NewCalculator(s.db).CalculateForCustomer(customerID, taxableAmount)
	if err != nil {
		return nil, err
	}

	total := taxableAmount.Add(taxAmount)

	// Generate order number
	orderNumber := s.generateOrderNumber()

	order := &domain.Order{
		OrderNumber: orderNumber,
		CustomerID:  customerID,
		Status:      domain.OrderStatusPending,
		Currency:    cart.Currency,
		Subtotal:    subtotal,
		Discount:    discount,
		TaxAmount:   taxAmount,
		Total:       total,
		CouponID:    cart.CouponID,
		IPAddress:   ipAddress,
		Items:       orderItems,
	}

	if err := s.db.Create(order).Error; err != nil {
		return nil, err
	}

	// Clear cart
	s.db.Delete(&domain.CartItem{}, "cart_id = ?", cartID)
	s.db.Delete(&cart)

	return order, nil
}

// GetOrder retrieves an order by ID
func (s *Service) GetOrder(id uint64) (*domain.Order, error) {
	var order domain.Order
	if err := s.db.Preload("Items.Product").Preload("Customer").Preload("Invoice").
		First(&order, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	return &order, nil
}

// GetOrderByNumber retrieves an order by order number
func (s *Service) GetOrderByNumber(orderNumber string) (*domain.Order, error) {
	var order domain.Order
	if err := s.db.Preload("Items.Product").Preload("Customer").
		Where("order_number = ?", orderNumber).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	return &order, nil
}

// ListOrders returns orders for a customer
func (s *Service) ListOrders(customerID uint64, limit, offset int) ([]domain.Order, int64, error) {
	var orders []domain.Order
	var total int64

	query := s.db.Model(&domain.Order{}).Where("customer_id = ?", customerID)
	query.Count(&total)

	if err := query.Preload("Items").Order("created_at DESC").
		Limit(limit).Offset(offset).Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// ListAllOrders returns all orders in the system (admin only)
func (s *Service) ListAllOrders(status domain.OrderStatus, limit, offset int) ([]domain.Order, int64, error) {
	var orders []domain.Order
	var total int64

	query := s.db.Model(&domain.Order{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)

	if err := query.Preload("Items").Preload("Customer").Order("created_at DESC").
		Limit(limit).Offset(offset).Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// UpdateOrderStatus updates the status of an order
func (s *Service) UpdateOrderStatus(orderID uint64, status domain.OrderStatus) error {
	return s.db.Model(&domain.Order{}).Where("id = ?", orderID).
		Update("status", status).Error
}

// ActivateOrder activates an order and creates services
func (s *Service) ActivateOrder(orderID uint64) error {
	var order domain.Order
	if err := s.db.Preload("Items").First(&order, orderID).Error; err != nil {
		return ErrOrderNotFound
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		for i, item := range order.Items {
			// Create service for each order item
			service := &domain.Service{
				CustomerID:       order.CustomerID,
				ProductID:        item.ProductID,
				OrderID:          &order.ID,
				Status:           domain.ServiceStatusPending,
				Domain:           item.Domain,
				Hostname:         item.Hostname,
				BillingCycle:     item.BillingCycle,
				Currency:         order.Currency,
				RecurringAmount:  item.RecurringFee,
				NextDueDate:      s.calculateNextDueDate(item.BillingCycle),
				RegistrationDate: time.Now(),
				ConfigSelection:  item.ConfigOptions,
			}

			if err := tx.Create(service).Error; err != nil {
				return err
			}

			// Update order item with service ID
			order.Items[i].ServiceID = &service.ID
			if err := tx.Model(&order.Items[i]).Update("service_id", service.ID).Error; err != nil {
				return err
			}
		}

		// Update order status
		return tx.Model(&order).Update("status", domain.OrderStatusActive).Error
	})
}

// CancelOrder cancels an order
func (s *Service) CancelOrder(orderID uint64, reason string) error {
	var order domain.Order
	if err := s.db.First(&order, orderID).Error; err != nil {
		return ErrOrderNotFound
	}

	return s.db.Model(&order).Updates(map[string]interface{}{
		"status":      domain.OrderStatusCancelled,
		"admin_notes": reason,
	}).Error
}

// GetService retrieves a service by ID
func (s *Service) GetService(id uint64) (*domain.Service, error) {
	var service domain.Service
	if err := s.db.Preload("Product").Preload("Customer").Preload("Server").Preload("IPAddress").
		First(&service, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrServiceNotFound
		}
		return nil, err
	}
	return &service, nil
}

// ListServices returns services for a customer
func (s *Service) ListServices(customerID uint64, status domain.ServiceStatus, limit, offset int) ([]domain.Service, int64, error) {
	var services []domain.Service
	var total int64

	query := s.db.Model(&domain.Service{}).Where("customer_id = ?", customerID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)

	if err := query.Preload("Product").Order("created_at DESC").
		Limit(limit).Offset(offset).Find(&services).Error; err != nil {
		return nil, 0, err
	}

	return services, total, nil
}

// SuspendService suspends a service
func (s *Service) SuspendService(serviceID uint64, reason string) error {
	return s.db.Model(&domain.Service{}).Where("id = ?", serviceID).
		Updates(map[string]interface{}{
			"status":            domain.ServiceStatusSuspended,
			"suspension_reason": reason,
		}).Error
}

// UnsuspendService unsuspends a service
func (s *Service) UnsuspendService(serviceID uint64) error {
	return s.db.Model(&domain.Service{}).Where("id = ?", serviceID).
		Updates(map[string]interface{}{
			"status":            domain.ServiceStatusActive,
			"suspension_reason": "",
		}).Error
}

// TerminateService terminates a service
func (s *Service) TerminateService(serviceID uint64) error {
	now := time.Now()
	return s.db.Model(&domain.Service{}).Where("id = ?", serviceID).
		Updates(map[string]interface{}{
			"status":           domain.ServiceStatusTerminated,
			"termination_date": &now,
		}).Error
}

// RenewService extends the next due date for a service
func (s *Service) RenewService(serviceID uint64) error {
	var service domain.Service
	if err := s.db.First(&service, serviceID).Error; err != nil {
		return ErrServiceNotFound
	}

	nextDueDate := s.calculateNextDueDate(service.BillingCycle)
	if service.NextDueDate.After(time.Now()) {
		// If not yet due, extend from current due date
		nextDueDate = s.addBillingPeriod(service.NextDueDate, service.BillingCycle)
	}

	return s.db.Model(&service).Update("next_due_date", nextDueDate).Error
}

// GetDueServices returns services due for renewal before the given date
func (s *Service) GetDueServices(beforeDate time.Time, limit int) ([]domain.Service, error) {
	var services []domain.Service
	if err := s.db.Where("status = ? AND next_due_date <= ?", domain.ServiceStatusActive, beforeDate).
		Preload("Product").Preload("Customer").
		Limit(limit).Find(&services).Error; err != nil {
		return nil, err
	}
	return services, nil
}

// generateOrderNumber generates a unique order number
func (s *Service) generateOrderNumber() string {
	return fmt.Sprintf("ORD-%d", time.Now().UnixNano())
}

// calculateNextDueDate calculates the next due date based on billing cycle
func (s *Service) calculateNextDueDate(billingCycle string) time.Time {
	return s.addBillingPeriod(time.Now(), billingCycle)
}

// addBillingPeriod adds a billing period to a date
func (s *Service) addBillingPeriod(from time.Time, billingCycle string) time.Time {
	switch billingCycle {
	case "monthly":
		return from.AddDate(0, 1, 0)
	case "quarterly":
		return from.AddDate(0, 3, 0)
	case "semi-annually", "semiannually":
		return from.AddDate(0, 6, 0)
	case "annually", "yearly":
		return from.AddDate(1, 0, 0)
	case "biennially":
		return from.AddDate(2, 0, 0)
	case "triennially":
		return from.AddDate(3, 0, 0)
	default:
		return from.AddDate(0, 1, 0) // Default to monthly
	}
}

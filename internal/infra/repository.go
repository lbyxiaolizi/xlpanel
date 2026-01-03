package infra

import "sync"

type Repository[T any] struct {
	mu    sync.RWMutex
	items map[string]T
	idFn  func(T) string
}

func NewRepository[T any](idFn func(T) string) *Repository[T] {
	return &Repository[T]{
		items: make(map[string]T),
		idFn:  idFn,
	}
}

func (r *Repository[T]) Add(item T) T {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[r.idFn(item)] = item
	return item
}

func (r *Repository[T]) Get(id string) (T, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, ok := r.items[id]
	return item, ok
}

func (r *Repository[T]) List() []T {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]T, 0, len(r.items))
	for _, item := range r.items {
		result = append(result, item)
	}
	return result
}

func (r *Repository[T]) Remove(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.items, id)
}

type CouponRepository struct {
	mu      sync.RWMutex
	coupons map[string]any
}

func NewCouponRepository() *CouponRepository {
	return &CouponRepository{coupons: make(map[string]any)}
}

func (r *CouponRepository) Add(code string, coupon any) any {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.coupons[code] = coupon
	return coupon
}

func (r *CouponRepository) Get(code string) (any, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	coupon, ok := r.coupons[code]
	return coupon, ok
}

func (r *CouponRepository) List() []any {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]any, 0, len(r.coupons))
	for _, coupon := range r.coupons {
		result = append(result, coupon)
	}
	return result
}

func (r *CouponRepository) Remove(code string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.coupons, code)
}

type MetricsRegistry struct {
	mu     sync.RWMutex
	counts map[string]int
}

func NewMetricsRegistry() *MetricsRegistry {
	return &MetricsRegistry{counts: make(map[string]int)}
}

func (m *MetricsRegistry) Increment(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counts[key]++
}

func (m *MetricsRegistry) Snapshot() map[string]int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	snapshot := make(map[string]int, len(m.counts))
	for k, v := range m.counts {
		snapshot[k] = v
	}
	return snapshot
}

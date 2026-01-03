package service

import (
	"xlpanel/internal/domain"
	"xlpanel/internal/infra"
)

type CatalogService struct {
	products *infra.Repository[domain.Product]
}

func NewCatalogService(products *infra.Repository[domain.Product]) *CatalogService {
	return &CatalogService{products: products}
}

func (s *CatalogService) AddProduct(product domain.Product) domain.Product {
	return s.products.Add(product)
}

func (s *CatalogService) ListProducts() []domain.Product {
	return s.products.List()
}

func (s *CatalogService) GetProduct(code string) (domain.Product, bool) {
	return s.products.Get(code)
}

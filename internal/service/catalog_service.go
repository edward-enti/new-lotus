// Package service holds application logic over the repositories.
package service

import (
	"sonlotus/internal/domain"
	"sonlotus/internal/repository"
)

// CatalogService serves public catalog reads and the calculator.
type CatalogService struct {
	Categories *repository.CategoryRepo
	Products   *repository.ProductRepo
}

func (s *CatalogService) ActiveCategories() ([]domain.Category, error) {
	return s.Categories.ListActive()
}

func (s *CatalogService) FeaturedProducts() ([]domain.Product, error) {
	return s.Products.ListFeatured()
}

func (s *CatalogService) ProductsByFilter(f repository.ProductFilter) ([]domain.Product, error) {
	return s.Products.ListFiltered(f)
}

func (s *CatalogService) ProductsByCategory(slug string) ([]domain.Product, error) {
	return s.Products.ListByCategorySlug(slug)
}

func (s *CatalogService) ProductBySlug(slug string) (domain.Product, error) {
	return s.Products.GetBySlugActive(slug)
}

func (s *CatalogService) RelatedProducts(p domain.Product) ([]domain.Product, error) {
	return s.Products.Related(p.CategoryID, p.ID, 4)
}

// Calculate wraps the pure domain calculator (PLAN §7 Gate 1).
func (s *CatalogService) Calculate(areaM2 float64, coats int, coverageRate float64) (float64, domain.PackagingMix, error) {
	liters, err := domain.CalculateLiters(areaM2, coats, coverageRate)
	if err != nil {
		return 0, domain.PackagingMix{}, err
	}
	return liters, domain.PackagingMixFor(liters), nil
}

package service

import (
	"sonlotus/internal/domain"
	"sonlotus/internal/repository"
)

// ContentService serves public articles and FTS search.
type ContentService struct {
	Articles *repository.ArticleRepo
	Products *repository.ProductRepo
}

func (s *ContentService) ArticlesByType(articleType string) ([]domain.Article, error) {
	return s.Articles.ListByTypePublished(articleType)
}

func (s *ContentService) LatestArticles(limit int) ([]domain.Article, error) {
	all, err := s.Articles.ListPublished()
	if err != nil {
		return nil, err
	}
	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

func (s *ContentService) ArticleBySlug(slug string) (domain.Article, error) {
	return s.Articles.GetBySlugPublished(slug)
}

// SearchResult is one FTS hit for the live-search partial.
type SearchResult struct {
	Kind  string // "product" or "article"
	Title string
	Slug  string
	URL   string
}

// Search runs one FTS query across products and articles, public-filtered
// by the repositories (active / published only).
func (s *ContentService) Search(q string) ([]SearchResult, error) {
	prods, err := s.Products.SearchFTS(q)
	if err != nil {
		return nil, err
	}
	arts, err := s.Articles.SearchFTS(q)
	if err != nil {
		return nil, err
	}
	var out []SearchResult
	for _, p := range prods {
		out = append(out, SearchResult{Kind: "product", Title: p.Name, Slug: p.Slug, URL: "/san-pham/" + p.Slug})
	}
	for _, a := range arts {
		out = append(out, SearchResult{Kind: "article", Title: a.Title, Slug: a.Slug, URL: "/bai-viet/" + a.Slug})
	}
	return out, nil
}

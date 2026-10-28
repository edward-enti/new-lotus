package service

import (
	"encoding/json"
	"encoding/xml"
	"html/template"

	"sonlotus/internal/domain"
)

// SEOService builds PageSEO structs and sitemap XML. JSON-LD is produced
// ONLY by encoding/json from marshalled structs (never string
// concatenation); encoding/json escapes '<' as \u003c, so a </script>
// payload in a name or description cannot break out of the JSON-LD block.
type SEOService struct {
	BaseURL string // e.g. https://sonlotus.com; empty in tests/local
}

func marshalJSONLD(v any) template.JS {
	b, err := json.Marshal(v)
	if err != nil {
		return template.JS("{}")
	}
	return template.JS(b)
}

func (s *SEOService) canonical(path string) string {
	if s.BaseURL == "" {
		return path
	}
	return s.BaseURL + path
}

// Canonical returns the absolute URL for path when BaseURL is set,
// otherwise the path unchanged (local-dev relative fallback).
func (s *SEOService) Canonical(path string) string { return s.canonical(path) }

// SitemapURL returns the sitemap URL (absolute when BaseURL is set),
// for the robots.txt Sitemap line.
func (s *SEOService) SitemapURL() string { return s.canonical("/sitemap.xml") }

func (s *SEOService) simple(title, desc, path string) domain.PageSEO {
	return domain.PageSEO{
		Title:        title,
		Description:  desc,
		CanonicalURL: s.canonical(path),
		OGType:       "website",
	}
}

// HomeSEO emits Organization + LocalBusiness via @graph.
func (s *SEOService) HomeSEO() domain.PageSEO {
	seo := s.simple("Sơn Lotus — Sơn Hệ Nước Cao Cấp", "Sơn Lotus: sơn gỗ, sơn giả gỗ, sơn kim loại, sơn ngói, sơn loa và chống thấm hệ nước, không mùi, an toàn cho sức khỏe.", "/")
	type graph struct {
		Context string `json:"@context"`
		Graph   []any  `json:"@graph"`
	}
	seo.JSONLD = marshalJSONLD(graph{
		Context: "https://schema.org",
		Graph: []any{
			domain.OrganizationJSONLD{Context: "https://schema.org", Type: "Organization", Name: "Sơn Lotus", URL: s.canonical("/")},
			domain.LocalBusinessJSONLD{Context: "https://schema.org", Type: "LocalBusiness", Name: "Sơn Lotus — Công ty TNHH Bích Trang", URL: s.canonical("/")},
		},
	})
	return seo
}

// ProductSEO emits schema.org/Product JSON-LD.
func (s *SEOService) ProductSEO(p domain.Product) domain.PageSEO {
	title := p.SEOTitle
	if title == "" {
		title = p.Name + " — Sơn Lotus"
	}
	desc := p.SEODescription
	if desc == "" {
		desc = p.ShortSummary
	}
	seo := s.simple(title, desc, "/san-pham/"+p.Slug)
	seo.OGType = "product"
	seo.OGImage = p.ThumbnailURL
	seo.JSONLD = marshalJSONLD(domain.ProductJSONLD{
		Context: "https://schema.org", Type: "Product",
		Name: p.Name, Description: p.ShortSummary,
		URL: s.canonical("/san-pham/" + p.Slug), Image: p.ThumbnailURL,
		Brand: domain.Brand{Type: "Brand", Name: "Sơn Lotus"},
	})
	return seo
}

// ArticleSEO emits schema.org/Article JSON-LD.
func (s *SEOService) ArticleSEO(a domain.Article) domain.PageSEO {
	title := a.SEOTitle
	if title == "" {
		title = a.Title + " — Sơn Lotus"
	}
	desc := a.SEODescription
	if desc == "" {
		desc = a.Excerpt
	}
	seo := s.simple(title, desc, "/bai-viet/"+a.Slug)
	seo.OGType = "article"
	seo.OGImage = a.ThumbnailURL
	seo.JSONLD = marshalJSONLD(domain.ArticleJSONLD{
		Context: "https://schema.org", Type: "Article",
		Headline: a.Title, Description: a.Excerpt,
		URL: s.canonical("/bai-viet/" + a.Slug), Image: a.ThumbnailURL,
		Author:        domain.Brand{Type: "Organization", Name: "Sơn Lotus"},
		DatePublished: a.PublishedAt.Format("2006-01-02"),
	})
	return seo
}

// CategorySEO is a simple SEO block for category pages.
func (s *SEOService) CategorySEO(c domain.Category) domain.PageSEO {
	title := c.SEOTitle
	if title == "" {
		title = c.Name + " — Sơn Lotus"
	}
	desc := c.SEODescription
	if desc == "" {
		desc = c.ShortDescription
	}
	return s.simple(title, desc, "/danh-muc/"+c.Slug)
}

// SitemapURL is one entry of sitemap.xml.
type SitemapURL struct {
	XMLName xml.Name `xml:"url"`
	Loc     string   `xml:"loc"`
}

type sitemapDoc struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URLs    []SitemapURL `xml:"url"`
}

// SitemapXML renders sitemap.xml for public-filtered content only:
// static pages, active categories, active products, published articles.
func (s *SEOService) SitemapXML(cats []domain.Category, prods []domain.Product, arts []domain.Article) ([]byte, error) {
	var urls []SitemapURL
	add := func(path string) { urls = append(urls, SitemapURL{Loc: s.canonical(path)}) }
	for _, p := range []string{"/", "/san-pham", "/cong-cu/tinh-luong-son", "/du-an", "/ho-tro-ky-thuat", "/giai-phap", "/gioi-thieu", "/lien-he"} {
		add(p)
	}
	for _, c := range cats {
		if c.IsActive {
			add("/danh-muc/" + c.Slug)
		}
	}
	for _, p := range prods {
		if p.IsActive {
			add("/san-pham/" + p.Slug)
		}
	}
	for _, a := range arts {
		if a.IsPublished {
			add("/bai-viet/" + a.Slug)
		}
	}
	out, err := xml.MarshalIndent(sitemapDoc{XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9", URLs: urls}, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte(xml.Header), out...), nil
}

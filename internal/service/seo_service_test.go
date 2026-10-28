package service_test

import (
	"strings"
	"testing"

	"sonlotus/internal/domain"
	"sonlotus/internal/service"
)

// A </script> payload in a product name must not break out of the
// JSON-LD block: encoding/json escapes '<' as \u003c.
func TestProductJSONLDBreakout(t *testing.T) {
	svc := &service.SEOService{}
	p := domain.Product{
		Name: "Test </script><script>alert(1)</script>", Slug: "test-breakout",
		ShortSummary: "Mô tả </script> độc hại",
	}
	seo := svc.ProductSEO(p)
	payload := string(seo.JSONLD)
	if strings.Contains(payload, "</script>") {
		t.Fatalf("JSON-LD payload contains literal </script>: %s", payload)
	}
	if !strings.Contains(payload, "Product") {
		t.Fatalf("JSON-LD payload missing Product type: %s", payload)
	}
}

func TestSitemapFiltersInactiveAndUnpublished(t *testing.T) {
	svc := &service.SEOService{}
	cats := []domain.Category{{Slug: "son-go-he-nuoc", IsActive: true}}
	prods := []domain.Product{
		{Slug: "active-product", IsActive: true},
		{Slug: "inactive-product", IsActive: false},
	}
	arts := []domain.Article{
		{Slug: "published-article", IsPublished: true},
		{Slug: "draft-article", IsPublished: false},
	}
	body, err := svc.SitemapXML(cats, prods, arts)
	if err != nil {
		t.Fatal(err)
	}
	s := string(body)
	for _, want := range []string{"/san-pham/active-product", "/bai-viet/published-article", "/danh-muc/son-go-he-nuoc"} {
		if !strings.Contains(s, want) {
			t.Errorf("sitemap missing %s", want)
		}
	}
	for _, notWant := range []string{"inactive-product", "draft-article"} {
		if strings.Contains(s, notWant) {
			t.Errorf("sitemap contains filtered slug %s", notWant)
		}
	}
}

// With BaseURL set, sitemap <loc>, canonical and JSON-LD urls must
// all be absolute; with it empty the relative fallback remains.
func TestSitemapAbsoluteURLsWhenBaseURLSet(t *testing.T) {
	svc := &service.SEOService{BaseURL: "https://sonlotus.com"}
	body, err := svc.SitemapXML(
		[]domain.Category{{Slug: "son-go-he-nuoc", IsActive: true}},
		[]domain.Product{{Slug: "active-product", IsActive: true}},
		[]domain.Article{{Slug: "published-article", IsPublished: true}},
	)
	if err != nil {
		t.Fatal(err)
	}
	s := string(body)
	for _, want := range []string{
		"<loc>https://sonlotus.com/san-pham/active-product</loc>",
		"<loc>https://sonlotus.com/bai-viet/published-article</loc>",
		"<loc>https://sonlotus.com/danh-muc/son-go-he-nuoc</loc>",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("sitemap missing %s", want)
		}
	}
	if strings.Contains(s, "<loc>/") {
		t.Errorf("sitemap contains relative <loc> despite BaseURL: %s", s)
	}
	seo := svc.ProductSEO(domain.Product{Slug: "active-product", Name: "P", ShortSummary: "S"})
	if seo.CanonicalURL != "https://sonlotus.com/san-pham/active-product" {
		t.Errorf("canonical = %q", seo.CanonicalURL)
	}
	if !strings.Contains(string(seo.JSONLD), "https://sonlotus.com/san-pham/active-product") {
		t.Errorf("JSON-LD lacks absolute url: %s", seo.JSONLD)
	}
	if svc.SitemapURL() != "https://sonlotus.com/sitemap.xml" {
		t.Errorf("SitemapURL = %q", svc.SitemapURL())
	}
}

func TestMediaSanitizeFilename(t *testing.T) {
	if got := service.SanitizeFilename("../../etc/passwd"); got != "passwd" {
		t.Errorf("SanitizeFilename traversal = %q, want passwd", got)
	}
	if got := service.SanitizeFilename("my photo (1).png"); got != "my_photo__1_.png" {
		t.Errorf("SanitizeFilename spaces/parens = %q", got)
	}
	if service.IsAllowedUploadExtension(".html") || service.IsAllowedUploadExtension(".svg") {
		t.Error("html/svg must not be allowed upload extensions")
	}
	if !service.IsAllowedUploadExtension(".pdf") || !service.IsAllowedUploadExtension(".png") {
		t.Error("pdf/png must be allowed upload extensions")
	}
}

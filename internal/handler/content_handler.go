package handler

import (
	"database/sql"
	"errors"
	"net/http"

	"sonlotus/internal/domain"
)

// ArticlesData is the data for articles_list.html.
type ArticlesData struct {
	Title    string
	Articles []domain.Article
}

// ArticleData is the data for article_detail.html.
type ArticleData struct {
	Article domain.Article
	Related []domain.Product
}

func (h *Handlers) listArticles(w http.ResponseWriter, articleType, title string) {
	arts, err := h.Content.ArticlesByType(articleType)
	if err != nil {
		http.Error(w, "content error", http.StatusInternalServerError)
		return
	}
	seo := h.SEO.HomeSEO()
	seo.Title = title + " — Sơn Lotus"
	seo.CanonicalURL = h.SEO.Canonical(map[string]string{
		"project": "/du-an", "tech_support": "/ho-tro-ky-thuat", "solution": "/giai-phap",
	}[articleType])
	h.Render.Render(w, "articles_list", seo, ArticlesData{Title: title, Articles: arts})
}

// ProjectsList renders /du-an (type = project).
func (h *Handlers) ProjectsList(w http.ResponseWriter, r *http.Request) {
	h.listArticles(w, domain.ArticleProject, "Dự án đã thi công")
}

// TechSupportList renders /ho-tro-ky-thuat (type = tech_support).
func (h *Handlers) TechSupportList(w http.ResponseWriter, r *http.Request) {
	h.listArticles(w, domain.ArticleTechSupport, "Hỗ trợ kỹ thuật")
}

// SolutionsList renders /giai-phap (type = solution).
func (h *Handlers) SolutionsList(w http.ResponseWriter, r *http.Request) {
	h.listArticles(w, domain.ArticleSolution, "Giải pháp ngành")
}

// ArticleDetail renders /bai-viet/{slug}; unpublished slugs are 404.
func (h *Handlers) ArticleDetail(w http.ResponseWriter, r *http.Request) {
	a, err := h.Content.ArticleBySlug(r.PathValue("slug"))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w)
			return
		}
		http.Error(w, "content error", http.StatusInternalServerError)
		return
	}
	var related []domain.Product
	if a.RelatedCategoryID != nil {
		// Related products come from the article's category, via a
		// category-scoped list filtered in the repository layer below.
		all, err := h.Products.ListActive()
		if err != nil {
			http.Error(w, "catalog error", http.StatusInternalServerError)
			return
		}
		for _, p := range all {
			if p.CategoryID == *a.RelatedCategoryID {
				related = append(related, p)
				if len(related) >= 4 {
					break
				}
			}
		}
	}
	h.Render.Render(w, "article_detail", h.SEO.ArticleSEO(a), ArticleData{Article: a, Related: related})
}

// LiveSearchHTMX renders the search_results partial for ?q=... (FTS5).
func (h *Handlers) LiveSearchHTMX(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	results, err := h.Content.Search(q)
	if err != nil {
		http.Error(w, "search error", http.StatusInternalServerError)
		return
	}
	h.Render.RenderPartial(w, "search_results", results)
}

// SitemapXML renders sitemap.xml for public-filtered content only.
func (h *Handlers) SitemapXML(w http.ResponseWriter, r *http.Request) {
	cats, err := h.Catalog.ActiveCategories()
	if err != nil {
		http.Error(w, "catalog error", http.StatusInternalServerError)
		return
	}
	prods, err := h.Products.ListActive()
	if err != nil {
		http.Error(w, "catalog error", http.StatusInternalServerError)
		return
	}
	arts, err := h.Content.LatestArticles(1000)
	if err != nil {
		http.Error(w, "content error", http.StatusInternalServerError)
		return
	}
	body, err := h.SEO.SitemapXML(cats, prods, arts)
	if err != nil {
		http.Error(w, "sitemap error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	_, _ = w.Write(body)
}

// RobotsTxt renders robots.txt pointing to the sitemap, disallowing /admin.
func (h *Handlers) RobotsTxt(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("User-agent: *\nDisallow: /admin\nSitemap: " + h.SEO.SitemapURL() + "\n"))
}

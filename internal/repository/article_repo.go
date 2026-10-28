package repository

import (
	"database/sql"

	"sonlotus/internal/database"
	"sonlotus/internal/domain"
)

// ArticleRepo reads and writes articles.
type ArticleRepo struct{ DB *sql.DB }

const articleCols = `id, type, title, slug, excerpt, content_html, thumbnail_url, author_name, related_category_id, is_featured, is_published, published_at, seo_title, seo_description, created_at, updated_at`

func scanArticle(row interface{ Scan(...any) error }) (domain.Article, error) {
	var a domain.Article
	var featured, published int
	err := row.Scan(&a.ID, &a.Type, &a.Title, &a.Slug, &a.Excerpt, &a.ContentHTML, &a.ThumbnailURL, &a.AuthorName,
		&a.RelatedCategoryID, &featured, &published, &a.PublishedAt, &a.SEOTitle, &a.SEODescription, &a.CreatedAt, &a.UpdatedAt)
	a.IsFeatured = featured != 0
	a.IsPublished = published != 0
	return a, err
}

// Create inserts an article (FTS sync via triggers) and returns its id.
func (r *ArticleRepo) Create(a domain.Article) (int64, error) {
	res, err := r.DB.Exec(`INSERT INTO articles(type, title, slug, excerpt, content_html, thumbnail_url, author_name, related_category_id, is_featured, is_published, seo_title, seo_description)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		a.Type, a.Title, a.Slug, a.Excerpt, a.ContentHTML, a.ThumbnailURL, a.AuthorName, a.RelatedCategoryID, boolInt(a.IsFeatured), boolInt(a.IsPublished), a.SEOTitle, a.SEODescription)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetByID returns an article by id regardless of published state (admin use).
func (r *ArticleRepo) GetByID(id int64) (domain.Article, error) {
	row := r.DB.QueryRow(`SELECT `+articleCols+` FROM articles WHERE id = ?`, id)
	return scanArticle(row)
}

// ListAll returns all articles regardless of published state (admin use).
func (r *ArticleRepo) ListAll() ([]domain.Article, error) {
	rows, err := r.DB.Query(`SELECT ` + articleCols + ` FROM articles ORDER BY created_at DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Article
	for rows.Next() {
		a, err := scanArticle(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// Update overwrites an article's editable fields (admin CRUD).
func (r *ArticleRepo) Update(a domain.Article) error {
	_, err := r.DB.Exec(`UPDATE articles SET type = ?, title = ?, slug = ?, excerpt = ?, content_html = ?, thumbnail_url = ?, author_name = ?, related_category_id = ?, is_featured = ?, is_published = ?, seo_title = ?, seo_description = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		a.Type, a.Title, a.Slug, a.Excerpt, a.ContentHTML, a.ThumbnailURL, a.AuthorName, a.RelatedCategoryID, boolInt(a.IsFeatured), boolInt(a.IsPublished), a.SEOTitle, a.SEODescription, a.ID)
	return err
}

// ListByTypePublished returns published articles of one type, newest first.
func (r *ArticleRepo) ListByTypePublished(articleType string) ([]domain.Article, error) {
	rows, err := r.DB.Query(`SELECT `+articleCols+` FROM articles WHERE type = ? AND is_published = 1 ORDER BY published_at DESC, id DESC`, articleType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Article
	for rows.Next() {
		a, err := scanArticle(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// ListPublished returns all published articles, newest first.
func (r *ArticleRepo) ListPublished() ([]domain.Article, error) {
	rows, err := r.DB.Query(`SELECT ` + articleCols + ` FROM articles WHERE is_published = 1 ORDER BY published_at DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Article
	for rows.Next() {
		a, err := scanArticle(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// GetBySlugPublished returns a published article by slug, or sql.ErrNoRows.
func (r *ArticleRepo) GetBySlugPublished(slug string) (domain.Article, error) {
	row := r.DB.QueryRow(`SELECT `+articleCols+` FROM articles WHERE slug = ? AND is_published = 1`, slug)
	return scanArticle(row)
}

// SearchFTS returns published articles matching a sanitised FTS5 query.
func (r *ArticleRepo) SearchFTS(raw string) ([]domain.Article, error) {
	match := database.FTSQuery(raw)
	if match == "" {
		return []domain.Article{}, nil
	}
	rows, err := r.DB.Query(`SELECT a.id, a.type, a.title, a.slug, a.excerpt, a.content_html, a.thumbnail_url, a.author_name, a.related_category_id, a.is_featured, a.is_published, a.published_at, a.seo_title, a.seo_description, a.created_at, a.updated_at
		FROM articles a JOIN search_index s ON s.entity_type = 'article' AND s.entity_id = a.id
		WHERE a.is_published = 1 AND s.search_index MATCH ? ORDER BY rank`, match)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Article
	for rows.Next() {
		a, err := scanArticle(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

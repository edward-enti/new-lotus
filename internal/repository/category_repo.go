package repository

import (
	"database/sql"

	"sonlotus/internal/domain"
)

// CategoryRepo reads categories. Public reads filter is_active (PLAN §3).
type CategoryRepo struct{ DB *sql.DB }

func scanCategory(row interface{ Scan(...any) error }) (domain.Category, error) {
	var c domain.Category
	var active int
	err := row.Scan(&c.ID, &c.Name, &c.Slug, &c.ShortDescription, &c.DescriptionHTML,
		&c.SurfaceType, &c.IconName, &c.ThumbnailURL, &c.SortOrder, &active,
		&c.SEOTitle, &c.SEODescription, &c.CreatedAt, &c.UpdatedAt)
	c.IsActive = active != 0
	return c, err
}

const categoryCols = `id, name, slug, short_description, description_html, surface_type, icon_name, thumbnail_url, sort_order, is_active, seo_title, seo_description, created_at, updated_at`

// Create inserts a category and returns its id.
func (r *CategoryRepo) Create(c domain.Category) (int64, error) {
	res, err := r.DB.Exec(`INSERT INTO categories(name, slug, short_description, description_html, surface_type, icon_name, thumbnail_url, sort_order, is_active, seo_title, seo_description)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.Name, c.Slug, c.ShortDescription, c.DescriptionHTML, c.SurfaceType, c.IconName, c.ThumbnailURL, c.SortOrder, boolInt(c.IsActive), c.SEOTitle, c.SEODescription)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ListActive returns active categories in sort order.
func (r *CategoryRepo) ListActive() ([]domain.Category, error) {
	rows, err := r.DB.Query(`SELECT ` + categoryCols + ` FROM categories WHERE is_active = 1 ORDER BY sort_order ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Category
	for rows.Next() {
		c, err := scanCategory(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// ListAll returns all categories regardless of active state (admin use).
func (r *CategoryRepo) ListAll() ([]domain.Category, error) {
	rows, err := r.DB.Query(`SELECT ` + categoryCols + ` FROM categories ORDER BY sort_order ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Category
	for rows.Next() {
		c, err := scanCategory(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// GetBySlugActive returns an active category by slug, or sql.ErrNoRows.
func (r *CategoryRepo) GetBySlugActive(slug string) (domain.Category, error) {
	row := r.DB.QueryRow(`SELECT `+categoryCols+` FROM categories WHERE slug = ? AND is_active = 1`, slug)
	return scanCategory(row)
}

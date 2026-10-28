package repository

import (
	"database/sql"
	"encoding/json"
	"strings"

	"sonlotus/internal/database"
	"sonlotus/internal/domain"
)

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// ProductRepo reads and writes products.
type ProductRepo struct{ DB *sql.DB }

const productCols = `p.id, p.category_id, p.sku, p.name, p.slug, p.coating_role, p.finish_type, p.usage_area,
	p.short_summary, p.description_html, p.application_guide_html, p.features_json, p.specs_json, p.packaging_json,
	p.colors_json, p.images_json, p.thumbnail_url, p.tds_file_url, p.msds_file_url, p.coverage_rate,
	p.is_featured, p.is_active, p.seo_title, p.seo_description, p.created_at, p.updated_at,
	c.slug, c.name`

func scanProduct(row interface{ Scan(...any) error }) (domain.Product, error) {
	var p domain.Product
	var feat, specs, pack, colors, images string
	var featured, active int
	err := row.Scan(&p.ID, &p.CategoryID, &p.SKU, &p.Name, &p.Slug, &p.CoatingRole, &p.FinishType, &p.UsageArea,
		&p.ShortSummary, &p.DescriptionHTML, &p.ApplicationGuideHTML, &feat, &specs, &pack,
		&colors, &images, &p.ThumbnailURL, &p.TDSFileURL, &p.MSDSFileURL, &p.CoverageRate,
		&featured, &active, &p.SEOTitle, &p.SEODescription, &p.CreatedAt, &p.UpdatedAt,
		&p.CategorySlug, &p.CategoryName)
	if err != nil {
		return p, err
	}
	p.IsFeatured = featured != 0
	p.IsActive = active != 0
	_ = json.Unmarshal([]byte(feat), &p.Features)
	_ = json.Unmarshal([]byte(specs), &p.Specs)
	_ = json.Unmarshal([]byte(pack), &p.Packaging)
	_ = json.Unmarshal([]byte(colors), &p.Colors)
	_ = json.Unmarshal([]byte(images), &p.Images)
	if p.Features == nil {
		p.Features = []string{}
	}
	if p.Packaging == nil {
		p.Packaging = []string{}
	}
	if p.Colors == nil {
		p.Colors = []domain.ColorSwatch{}
	}
	if p.Images == nil {
		p.Images = []string{}
	}
	return p, nil
}

func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "[]"
	}
	return string(b)
}

// Create inserts a product (FTS sync happens via triggers) and returns its id.
func (r *ProductRepo) Create(p domain.Product) (int64, error) {
	res, err := r.DB.Exec(`INSERT INTO products(category_id, sku, name, slug, coating_role, finish_type, usage_area,
		short_summary, description_html, application_guide_html, features_json, specs_json, packaging_json, colors_json, images_json,
		thumbnail_url, tds_file_url, msds_file_url, coverage_rate, is_featured, is_active, seo_title, seo_description)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.CategoryID, p.SKU, p.Name, p.Slug, p.CoatingRole, p.FinishType, p.UsageArea,
		p.ShortSummary, p.DescriptionHTML, p.ApplicationGuideHTML, mustJSON(p.Features), mustJSON(p.Specs), mustJSON(p.Packaging), mustJSON(p.Colors), mustJSON(p.Images),
		p.ThumbnailURL, p.TDSFileURL, p.MSDSFileURL, p.CoverageRate, boolInt(p.IsFeatured), boolInt(p.IsActive), p.SEOTitle, p.SEODescription)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *ProductRepo) queryList(query string, args ...any) ([]domain.Product, error) {
	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Product
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

const productJoin = ` FROM products p JOIN categories c ON c.id = p.category_id `

// ListActive returns all active products.
func (r *ProductRepo) ListActive() ([]domain.Product, error) {
	return r.queryList(`SELECT ` + productCols + productJoin + ` WHERE p.is_active = 1 ORDER BY p.id ASC`)
}

// ListByCategorySlug returns active products of one active category.
func (r *ProductRepo) ListByCategorySlug(slug string) ([]domain.Product, error) {
	return r.queryList(`SELECT `+productCols+productJoin+` WHERE p.is_active = 1 AND c.slug = ? AND c.is_active = 1 ORDER BY p.id ASC`, slug)
}

// ListFeatured returns active featured products.
func (r *ProductRepo) ListFeatured() ([]domain.Product, error) {
	return r.queryList(`SELECT ` + productCols + productJoin + ` WHERE p.is_active = 1 AND p.is_featured = 1 ORDER BY p.id ASC`)
}

// ProductFilter is the catalog sidebar filter set (all fields optional).
type ProductFilter struct {
	CategorySlug string
	CoatingRole  string
	FinishType   string
	UsageArea    string
}

// ListFiltered returns active products matching the filter.
func (r *ProductRepo) ListFiltered(f ProductFilter) ([]domain.Product, error) {
	var conds []string
	var args []any
	conds = append(conds, "p.is_active = 1")
	if f.CategorySlug != "" {
		conds = append(conds, "c.slug = ?")
		args = append(args, f.CategorySlug)
	}
	if f.CoatingRole != "" {
		conds = append(conds, "p.coating_role = ?")
		args = append(args, f.CoatingRole)
	}
	if f.FinishType != "" {
		conds = append(conds, "p.finish_type = ?")
		args = append(args, f.FinishType)
	}
	if f.UsageArea != "" {
		conds = append(conds, "p.usage_area = ?")
		args = append(args, f.UsageArea)
	}
	return r.queryList(`SELECT `+productCols+productJoin+` WHERE `+strings.Join(conds, " AND ")+` ORDER BY p.id ASC`, args...)
}

// GetBySlugActive returns an active product by slug, or sql.ErrNoRows.
// An inactive product is indistinguishable from a missing one: callers
// must return 404 for both (PLAN §3).
func (r *ProductRepo) GetBySlugActive(slug string) (domain.Product, error) {
	row := r.DB.QueryRow(`SELECT `+productCols+productJoin+` WHERE p.slug = ? AND p.is_active = 1`, slug)
	return scanProduct(row)
}

// GetByID returns a product by id regardless of active state (admin use).
func (r *ProductRepo) GetByID(id int64) (domain.Product, error) {
	row := r.DB.QueryRow(`SELECT `+productCols+productJoin+` WHERE p.id = ?`, id)
	return scanProduct(row)
}

// ListAll returns all products regardless of active state (admin use).
func (r *ProductRepo) ListAll() ([]domain.Product, error) {
	return r.queryList(`SELECT ` + productCols + productJoin + ` ORDER BY p.id ASC`)
}

// Update overwrites a product's editable fields (admin CRUD).
func (r *ProductRepo) Update(p domain.Product) error {
	_, err := r.DB.Exec(`UPDATE products SET category_id = ?, sku = ?, name = ?, slug = ?, coating_role = ?, finish_type = ?, usage_area = ?,
		short_summary = ?, description_html = ?, application_guide_html = ?, features_json = ?, specs_json = ?, packaging_json = ?, colors_json = ?, images_json = ?,
		thumbnail_url = ?, tds_file_url = ?, msds_file_url = ?, coverage_rate = ?, is_featured = ?, is_active = ?, seo_title = ?, seo_description = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		p.CategoryID, p.SKU, p.Name, p.Slug, p.CoatingRole, p.FinishType, p.UsageArea,
		p.ShortSummary, p.DescriptionHTML, p.ApplicationGuideHTML, mustJSON(p.Features), mustJSON(p.Specs), mustJSON(p.Packaging), mustJSON(p.Colors), mustJSON(p.Images),
		p.ThumbnailURL, p.TDSFileURL, p.MSDSFileURL, p.CoverageRate, boolInt(p.IsFeatured), boolInt(p.IsActive), p.SEOTitle, p.SEODescription, p.ID)
	return err
}

// Related returns up to limit active products in the same category,
// excluding the given product id.
func (r *ProductRepo) Related(categoryID, excludeID int64, limit int) ([]domain.Product, error) {
	return r.queryList(`SELECT `+productCols+productJoin+` WHERE p.is_active = 1 AND p.category_id = ? AND p.id != ? ORDER BY p.id ASC LIMIT ?`, categoryID, excludeID, limit)
}

// SearchFTS returns active products matching a diacritic-insensitive
// FTS5 query (raw user input is sanitised by database.FTSQuery first).
func (r *ProductRepo) SearchFTS(raw string) ([]domain.Product, error) {
	match := database.FTSQuery(raw)
	if match == "" {
		return []domain.Product{}, nil
	}
	return r.queryList(`SELECT `+productCols+productJoin+`
		JOIN search_index s ON s.entity_type = 'product' AND s.entity_id = p.id
		WHERE p.is_active = 1 AND s.search_index MATCH ? ORDER BY rank`, match)
}

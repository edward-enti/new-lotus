package domain

import "time"

// Category is a top-level product category (SPEC §4.1).
type Category struct {
	ID               int64
	Name             string
	Slug             string
	ShortDescription string
	DescriptionHTML  string
	SurfaceType      string
	IconName         string
	ThumbnailURL     string
	SortOrder        int
	IsActive         bool
	SEOTitle         string
	SEODescription   string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// ColorSwatch is one entry of a product's colors_json array.
type ColorSwatch struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	Hex       string `json:"hex"`
	SwatchURL string `json:"swatch_url"`
}

// ProductSpecs is the decoded form of a product's specs_json object.
type ProductSpecs struct {
	CoverageM2PerLiter float64 `json:"coverage_m2_per_liter"`
	RecommendedCoats   int     `json:"recommended_coats"`
	DryTimeTouchMins   int     `json:"dry_time_touch_mins"`
	DryTimeRecoatHours float64 `json:"dry_time_recoat_hours"`
	Dilution           string  `json:"dilution"`
}

// Product is a water-based paint product (SPEC §4.1).
type Product struct {
	ID                   int64
	CategoryID           int64
	CategorySlug         string // filled by list queries via JOIN, empty on create
	CategoryName         string // filled by list queries via JOIN, empty on create
	SKU                  string
	Name                 string
	Slug                 string
	CoatingRole          string
	FinishType           string
	UsageArea            string
	ShortSummary         string
	DescriptionHTML      string
	ApplicationGuideHTML string
	Features             []string
	Specs                ProductSpecs
	Packaging            []string
	Colors               []ColorSwatch
	Images               []string
	ThumbnailURL         string
	TDSFileURL           string
	MSDSFileURL          string
	CoverageRate         float64
	IsFeatured           bool
	IsActive             bool
	SEOTitle             string
	SEODescription       string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

package domain

import "time"

// Article types (SPEC §4.1 CHECK constraint).
const (
	ArticleProject     = "project"
	ArticleTechSupport = "tech_support"
	ArticleSolution    = "solution"
	ArticleNews        = "news"
)

// Article is a project, tech-support guide, solution or news post.
type Article struct {
	ID                int64
	Type              string
	Title             string
	Slug              string
	Excerpt           string
	ContentHTML       string
	ThumbnailURL      string
	AuthorName        string
	RelatedCategoryID *int64
	IsFeatured        bool
	IsPublished       bool
	PublishedAt       time.Time
	SEOTitle          string
	SEODescription    string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

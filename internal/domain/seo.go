package domain

import "html/template"

// PageSEO is passed to base.html on every public page (SPEC §8.2).
// JSONLD is produced ONLY by encoding/json from a marshalled struct in
// the seo service and is emitted as template.JS; never build it by
// string concatenation.
type PageSEO struct {
	Title        string
	Description  string
	CanonicalURL string
	OGImage      string
	OGType       string
	JSONLD       template.JS
}

// ProductJSONLD is the schema.org/Product payload for product pages.
type ProductJSONLD struct {
	Context     string `json:"@context"`
	Type        string `json:"@type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	URL         string `json:"url,omitempty"`
	Image       string `json:"image,omitempty"`
	Brand       Brand  `json:"brand,omitempty"`
}

// ArticleJSONLD is the schema.org/Article payload for article pages.
type ArticleJSONLD struct {
	Context       string `json:"@context"`
	Type          string `json:"@type"`
	Headline      string `json:"headline"`
	Description   string `json:"description"`
	URL           string `json:"url,omitempty"`
	Image         string `json:"image,omitempty"`
	Author        Brand  `json:"author,omitempty"`
	DatePublished string `json:"datePublished,omitempty"`
}

// OrganizationJSONLD is the schema.org/Organization payload for the home page.
type OrganizationJSONLD struct {
	Context string `json:"@context"`
	Type    string `json:"@type"`
	Name    string `json:"name"`
	URL     string `json:"url,omitempty"`
	Logo    string `json:"logo,omitempty"`
}

// LocalBusinessJSONLD is the schema.org/LocalBusiness payload for the home page.
type LocalBusinessJSONLD struct {
	Context   string `json:"@context"`
	Type      string `json:"@type"`
	Name      string `json:"name"`
	URL       string `json:"url,omitempty"`
	Telephone string `json:"telephone,omitempty"`
}

// Brand is a reusable schema.org brand/author object.
type Brand struct {
	Type string `json:"@type"`
	Name string `json:"name"`
}

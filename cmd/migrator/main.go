// Command migrator seeds the database and (operator-invoked only)
// scrapes the legacy sonlotus.com sitemap into products, articles and
// url_redirects.
//
// Safety (PLAN §5, SPEC §7): fetch ONLY sonlotus.com hosts, max 2
// requests/second, 10s timeout, User-Agent SonLotus-Migrator/1.0,
// response-size and redirect caps, TLS verification NEVER skipped,
// slug collision = upsert-or-skip (existing slug is never duplicated),
// scraped HTML sanitised (script/style/iframe/form removed, event
// handlers and javascript: URLs stripped) before storage, and
// --dry-run creates NO database file and opens NO writable database.
// Scraped rows are drafts (products inactive, articles unpublished)
// pending human review in the admin CMS before publish — scraped
// products are assigned the first available category as a placeholder
// (not an inferred category mapping), and legacy content is never
// published directly.
//
// Do NOT run a live scrape during development, tests, or gates; the
// scrape path is delivered as code plus a --dry-run demonstration only.
package main

import (
	"database/sql"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"sonlotus/internal/database"
	"sonlotus/internal/domain"
	"sonlotus/internal/repository"
)

const (
	userAgent   = "SonLotus-Migrator/1.0"
	maxBody     = 10 << 20 // 10 MB response-size cap
	maxRedirect = 5
	maxSitemaps = 50
	maxURLs     = 1000
)

func allowedHost(host string) bool {
	return host == "sonlotus.com" || host == "www.sonlotus.com"
}

type scraper struct {
	client *http.Client
	tick   <-chan time.Time
}

func newScraper() *scraper {
	ticker := time.NewTicker(500 * time.Millisecond) // 2 req/s
	return &scraper{
		client: &http.Client{
			Timeout: 10 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= maxRedirect {
					return fmt.Errorf("too many redirects")
				}
				if !allowedHost(req.URL.Host) {
					return fmt.Errorf("redirect leaves sonlotus.com allowlist: %s", req.URL.Host)
				}
				return nil
			},
		},
		tick: ticker.C,
	}
}

// fetch returns the body of an allowlisted URL, size-capped. TLS
// verification is the default transport's: never skipped.
func (s *scraper) fetch(raw string) ([]byte, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	if !allowedHost(u.Host) {
		return nil, fmt.Errorf("host not in allowlist: %s", u.Host)
	}
	<-s.tick
	req, err := http.NewRequest(http.MethodGet, raw, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: status %d", raw, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))
	if err != nil {
		return nil, err
	}
	if len(body) > maxBody {
		return nil, fmt.Errorf("response exceeds %d bytes: %s", maxBody, raw)
	}
	return body, nil
}

type sitemapIndex struct {
	XMLName  xml.Name `xml:"sitemapindex"`
	Sitemaps []struct {
		Loc string `xml:"loc"`
	} `xml:"sitemap"`
}

type urlSet struct {
	XMLName xml.Name `xml:"urlset"`
	URLs    []struct {
		Loc string `xml:"loc"`
	} `xml:"url"`
}

func (s *scraper) discover() ([]string, error) {
	for _, candidate := range []string{
		"https://sonlotus.com/sitemap_index.xml",
		"https://sonlotus.com/sitemap.xml",
	} {
		body, err := s.fetch(candidate)
		if err != nil {
			log.Printf("sitemap candidate %s failed: %v", candidate, err)
			continue
		}
		var idx sitemapIndex
		if err := xml.Unmarshal(body, &idx); err == nil && len(idx.Sitemaps) > 0 {
			var out []string
			for i, sm := range idx.Sitemaps {
				if i >= maxSitemaps {
					break
				}
				child, err := s.fetch(sm.Loc)
				if err != nil {
					log.Printf("child sitemap %s failed: %v", sm.Loc, err)
					continue
				}
				var set urlSet
				if err := xml.Unmarshal(child, &set); err != nil {
					continue
				}
				for _, u := range set.URLs {
					out = append(out, u.Loc)
				}
			}
			return dedupe(out), nil
		}
		var set urlSet
		if err := xml.Unmarshal(body, &set); err == nil && len(set.URLs) > 0 {
			var out []string
			for _, u := range set.URLs {
				out = append(out, u.Loc)
			}
			return dedupe(out), nil
		}
	}
	return nil, fmt.Errorf("no usable sitemap found at sonlotus.com")
}

func dedupe(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, v := range in {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

type page struct {
	URL, Title, Description, H1, ContentHTML, Image string
}

// sanitizeHTML removes executable content (script/style/iframe/form,
// event-handler attributes, javascript: URLs) from scraped HTML. The
// repo's escape-by-default policy means stored HTML is never rendered
// raw, so this is a second layer, not the only protection.
func sanitizeHTML(html string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return ""
	}
	doc.Find("script, style, iframe, form, noscript").Remove()
	doc.Find("*").Each(func(_ int, sel *goquery.Selection) {
		for _, attr := range []string{"onclick", "onload", "onerror", "onmouseover"} {
			sel.RemoveAttr(attr)
		}
		for _, attr := range []string{"href", "src"} {
			if v, ok := sel.Attr(attr); ok && strings.HasPrefix(strings.ToLower(strings.TrimSpace(v)), "javascript:") {
				sel.RemoveAttr(attr)
			}
		}
		// Drop any other on* attribute generically.
		if len(sel.Nodes) > 0 {
			kept := sel.Nodes[0].Attr[:0]
			for _, a := range sel.Nodes[0].Attr {
				if !strings.HasPrefix(strings.ToLower(a.Key), "on") {
					kept = append(kept, a)
				}
			}
			sel.Nodes[0].Attr = kept
		}
	})
	clean, _ := doc.Find("body").Html()
	return clean
}

func slugFromURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	p := strings.Trim(u.Path, "/")
	if p == "" {
		return ""
	}
	parts := strings.Split(p, "/")
	slug := parts[len(parts)-1]
	slug = strings.ToLower(slug)
	var b strings.Builder
	for _, r := range slug {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	out := strings.Trim(b.String(), "-")
	if len(out) > 100 {
		out = out[:100]
	}
	return out
}

func (s *scraper) extract(raw string) (page, error) {
	body, err := s.fetch(raw)
	if err != nil {
		return page{}, err
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return page{}, err
	}
	p := page{URL: raw}
	p.Title = strings.TrimSpace(doc.Find("title").First().Text())
	p.Description, _ = doc.Find(`meta[name="description"]`).Attr("content")
	p.H1 = strings.TrimSpace(doc.Find("h1").First().Text())
	p.Image, _ = doc.Find(`meta[property="og:image"]`).Attr("content")
	sel := doc.Find("main, article, .entry-content").First()
	if sel.Length() > 0 {
		inner, _ := sel.Html()
		p.ContentHTML = sanitizeHTML(inner)
	}
	if p.Image == "" {
		p.Image, _ = doc.Find("img").First().Attr("src")
	}
	return p, nil
}

func slugExists(db *sql.DB, slug string) bool {
	var one int
	if err := db.QueryRow(`SELECT 1 FROM products WHERE slug = ? UNION SELECT 1 FROM articles WHERE slug = ? LIMIT 1`, slug, slug).Scan(&one); err == nil {
		return true
	}
	return false
}

func isProductURL(raw string) bool {
	return strings.Contains(raw, "/san-pham/") || strings.Contains(raw, "/product/")
}

func articleType(raw string) string {
	switch {
	case strings.Contains(raw, "/du-an/"):
		return domain.ArticleProject
	case strings.Contains(raw, "/ho-tro-"):
		return domain.ArticleTechSupport
	case strings.Contains(raw, "/giai-phap/"):
		return domain.ArticleSolution
	default:
		return domain.ArticleNews
	}
}

func newPathFor(raw, slug string) string {
	if isProductURL(raw) {
		return "/san-pham/" + slug
	}
	return "/bai-viet/" + slug
}

func main() {
	seedOnly := flag.Bool("seed-only", false, "run migrations and seed data")
	scrape := flag.Bool("scrape-sitemap", false, "scrape sonlotus.com sitemap into the database")
	dryRun := flag.Bool("dry-run", false, "log discovered URLs and parsed structs, no database writes")
	dbPath := flag.String("db", "", "database path (default: $DB_PATH or data/sonlotus.db)")
	flag.Parse()
	if !*seedOnly && !*scrape && !*dryRun {
		fmt.Fprintln(os.Stderr, "migrator: one of --seed-only, --scrape-sitemap, or --dry-run is required")
		flag.Usage()
		os.Exit(2)
	}
	if *dryRun && !*scrape {
		// Seed dry run: describe only, touch no database file.
		log.Println("dry-run: would run migrations, seed categories/products (idempotent, no-op if categories exist), and FTS backfill")
		log.Println("dry-run: no database file created, no writable database opened")
		return
	}
	path := *dbPath
	if path == "" {
		path = os.Getenv("DB_PATH")
	}
	if path == "" {
		path = "data/sonlotus.db"
	}
	if *scrape && *dryRun {
		// Dry-run scrape: network reads only, NO database opened at all.
		s := newScraper()
		urls, err := s.discover()
		if err != nil {
			log.Fatalf("discover: %v", err)
		}
		log.Printf("dry-run: discovered %d URLs (capped at %d)", len(urls), maxURLs)
		for i, u := range urls {
			if i >= maxURLs {
				break
			}
			p, err := s.extract(u)
			if err != nil {
				log.Printf("dry-run: extract %s failed: %v", u, err)
				continue
			}
			log.Printf("dry-run: url=%s slug=%s title=%q h1=%q image=%s content_bytes=%d",
				p.URL, slugFromURL(p.URL), p.Title, p.H1, p.Image, len(p.ContentHTML))
		}
		log.Println("dry-run: no database file created, no writable database opened")
		return
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		log.Fatalf("data dir: %v", err)
	}
	db, err := database.Open(path)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()
	if err := database.Migrate(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	if *seedOnly {
		if err := database.Seed(db); err != nil {
			log.Fatalf("seed: %v", err)
		}
		log.Println("seed-only: migrations + seed complete")
		return
	}
	// Live scrape (operator-invoked only, never in tests/gates).
	if err := database.Seed(db); err != nil {
		log.Fatalf("seed before scrape: %v", err)
	}
	prods := &repository.ProductRepo{DB: db}
	arts := &repository.ArticleRepo{DB: db}
	cats := &repository.CategoryRepo{DB: db}
	var firstCat int64
	if list, err := cats.ListAll(); err == nil && len(list) > 0 {
		firstCat = list[0].ID
	}
	s := newScraper()
	urls, err := s.discover()
	if err != nil {
		log.Fatalf("discover: %v", err)
	}
	inserted, skipped := 0, 0
	for i, u := range urls {
		if i >= maxURLs {
			break
		}
		slug := slugFromURL(u)
		if slug == "" || slugExists(db, slug) {
			skipped++
			continue
		}
		p, err := s.extract(u)
		if err != nil {
			log.Printf("extract %s failed: %v", u, err)
			continue
		}
		title := p.Title
		if title == "" {
			title = p.H1
		}
		if title == "" {
			skipped++
			continue
		}
		oldPath := ""
		if parsed, err := url.Parse(u); err == nil {
			oldPath = parsed.Path
		}
		if isProductURL(u) {
			if firstCat == 0 {
				log.Printf("skip product %s: no category available", u)
				skipped++
				continue
			}
			sku := "scraped-" + slug
			if len(sku) > 64 {
				sku = sku[:64]
			}
			_, err = prods.Create(domain.Product{
				CategoryID: firstCat, SKU: sku, Name: title, Slug: slug,
				ShortSummary: p.Description, DescriptionHTML: p.ContentHTML,
				ThumbnailURL: p.Image, SEOTitle: title, SEODescription: p.Description,
				IsActive: false, // draft: never publish legacy content directly
			})
		} else {
			_, err = arts.Create(domain.Article{
				Type: articleType(u), Title: title, Slug: slug,
				Excerpt: p.Description, ContentHTML: p.ContentHTML,
				ThumbnailURL: p.Image, SEOTitle: title, SEODescription: p.Description,
				IsPublished: false, // draft: never publish legacy content directly
			})
		}
		if err != nil {
			log.Printf("insert %s failed: %v", u, err)
			continue
		}
		newPath := newPathFor(u, slug)
		if oldPath != "" && oldPath != newPath {
			_, _ = db.Exec(`INSERT OR IGNORE INTO url_redirects(old_path, new_path, status_code) VALUES (?, ?, 301)`, oldPath, newPath)
		}
		inserted++
	}
	if err := database.BackfillFTS(db); err != nil {
		log.Fatalf("backfill: %v", err)
	}
	log.Printf("scrape complete: inserted=%d skipped=%d urls=%d", inserted, skipped, len(urls))
}

package repository_test

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"sonlotus/internal/database"
	"sonlotus/internal/domain"
	"sonlotus/internal/repository"
)

// testDB creates a fresh file-backed DB (WAL + FK need a real file, not
// :memory:), runs migrations and seeds it. Cleanup removes the DB files.
func testDB(t *testing.T) *sql.DB {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")
	db, err := database.Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := database.Seed(db); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return db
}

func TestPragmasAreOn(t *testing.T) {
	db := testDB(t)
	var fk int
	var jm string
	if err := db.QueryRow(`PRAGMA foreign_keys`).Scan(&fk); err != nil {
		t.Fatal(err)
	}
	if fk != 1 {
		t.Fatalf("foreign_keys = %d, want 1", fk)
	}
	if err := db.QueryRow(`PRAGMA journal_mode`).Scan(&jm); err != nil {
		t.Fatal(err)
	}
	if jm != "wal" {
		t.Fatalf("journal_mode = %q, want wal", jm)
	}
}

func TestSeedCountsAndIdempotency(t *testing.T) {
	db := testDB(t)
	var cats, prods int
	db.QueryRow(`SELECT COUNT(*) FROM categories`).Scan(&cats)
	db.QueryRow(`SELECT COUNT(*) FROM products`).Scan(&prods)
	if cats != 7 {
		t.Fatalf("categories = %d, want 7", cats)
	}
	if prods != 8 {
		t.Fatalf("products = %d, want 8", prods)
	}
	if err := database.Seed(db); err != nil {
		t.Fatalf("second seed: %v", err)
	}
	db.QueryRow(`SELECT COUNT(*) FROM categories`).Scan(&cats)
	db.QueryRow(`SELECT COUNT(*) FROM products`).Scan(&prods)
	if cats != 7 || prods != 8 {
		t.Fatalf("after second seed: categories=%d products=%d, want 7/8", cats, prods)
	}
	// Exact seed slugs from PLAN §3.
	for _, slug := range []string{"son-go-he-nuoc", "son-gia-go", "son-kim-loai-he-nuoc", "son-ngoi-be-tong", "son-loa", "son-composite-nhua", "tram-tret-chong-tham"} {
		var n int
		db.QueryRow(`SELECT COUNT(*) FROM categories WHERE slug = ?`, slug).Scan(&n)
		if n != 1 {
			t.Errorf("category slug %s count = %d, want 1", slug, n)
		}
	}
	for _, slug := range []string{
		"son-phu-go-he-nuoc-lotus-acrylic-lacquer", "son-lot-go-he-nuoc-lotus-wood-primer",
		"bot-tret-go-he-nuoc-lotus-wood-filler", "son-gia-go-lotus-sat-ma-kem-xi-mang",
		"son-kim-loai-ma-kem-he-nuoc-lotus", "son-loa-tao-gai-san-lotus-speaker-paint",
		"son-ngoi-cao-cap-he-nuoc-lotus", "keo-bong-nuoc-chong-tham-lotus-clear-sealer"} {
		var n int
		db.QueryRow(`SELECT COUNT(*) FROM products WHERE slug = ?`, slug).Scan(&n)
		if n != 1 {
			t.Errorf("product slug %s count = %d, want 1", slug, n)
		}
	}
}

func TestProductCreation(t *testing.T) {
	db := testDB(t)
	cats := &repository.CategoryRepo{DB: db}
	catID, err := cats.Create(domain.Category{Name: "Test Category", Slug: "test-category", IsActive: true})
	if err != nil {
		t.Fatalf("create category: %v", err)
	}
	prods := &repository.ProductRepo{DB: db}
	id, err := prods.Create(domain.Product{
		CategoryID: catID, SKU: "TEST-001", Name: "Sơn Test Mới", Slug: "son-test-moi",
		CoatingRole: "topcoat", FinishType: "gloss", UsageArea: "both",
		ShortSummary: "Sản phẩm kiểm thử", CoverageRate: 10, IsActive: true,
		Features: []string{"Không mùi"}, Packaging: []string{"Lon 0.8L"},
	})
	if err != nil {
		t.Fatalf("create product: %v", err)
	}
	got, err := prods.GetBySlugActive("son-test-moi")
	if err != nil {
		t.Fatalf("get by slug: %v", err)
	}
	if got.ID != id || got.Name != "Sơn Test Mới" || got.CoverageRate != 10 {
		t.Fatalf("got %+v, want id=%d name Sơn Test Mới coverage 10", got, id)
	}
	if len(got.Features) != 1 || got.Features[0] != "Không mùi" {
		t.Fatalf("features = %v", got.Features)
	}
}

func TestCategoryFiltering(t *testing.T) {
	db := testDB(t)
	prods := &repository.ProductRepo{DB: db}
	got, err := prods.ListByCategorySlug("son-gia-go")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Slug != "son-gia-go-lotus-sat-ma-kem-xi-mang" {
		t.Fatalf("son-gia-go products = %v, want exactly the faux-wood product", got)
	}
	filtered, err := prods.ListFiltered(repository.ProductFilter{CoatingRole: "primer"})
	if err != nil {
		t.Fatal(err)
	}
	if len(filtered) != 1 || filtered[0].Slug != "son-lot-go-he-nuoc-lotus-wood-primer" {
		t.Fatalf("primer filter = %v, want exactly the wood primer", filtered)
	}
}

func TestVietnameseFTSWithoutDiacritics(t *testing.T) {
	db := testDB(t)
	prods := &repository.ProductRepo{DB: db}
	// SPEC Gate 1 example: unaccented query must find the accented name.
	got, err := prods.SearchFTS("son gia go")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, p := range got {
		if p.Slug == "son-gia-go-lotus-sat-ma-kem-xi-mang" {
			found = true
		}
	}
	if !found {
		t.Fatalf("search 'son gia go' = %v, want Sơn Giả Gỗ Lotus product", got)
	}
	// Additional accented/unaccented pair (PLAN §8): both forms find wood paint.
	for _, q := range []string{"sơn gỗ", "son go"} {
		got, err := prods.SearchFTS(q)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) == 0 {
			t.Fatalf("search %q returned nothing, want wood products", q)
		}
	}
	// Empty / operator-only queries must not error and must not match all.
	empty, err := prods.SearchFTS(`" * :`)
	if err != nil {
		t.Fatal(err)
	}
	if len(empty) != 0 {
		t.Fatalf("operator-only search returned %d rows, want 0", len(empty))
	}
}

func TestFTSBackfillAfterDirectInsert(t *testing.T) {
	db := testDB(t)
	// Bypass the repository layer with raw SQL (trigger still fires, but
	// backfill is the safety net): delete the FTS row, backfill, search.
	if _, err := db.Exec(`DELETE FROM search_index WHERE entity_type = 'product'`); err != nil {
		t.Fatal(err)
	}
	if err := database.BackfillFTS(db); err != nil {
		t.Fatalf("backfill: %v", err)
	}
	prods := &repository.ProductRepo{DB: db}
	got, err := prods.SearchFTS("son gia go")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Fatal("after backfill, search 'son gia go' returned nothing")
	}
}

func TestForeignKeyFailure(t *testing.T) {
	db := testDB(t)
	prods := &repository.ProductRepo{DB: db}
	_, err := prods.Create(domain.Product{
		CategoryID: 999999, SKU: "TEST-FK", Name: "Bad Product", Slug: "bad-product",
		CoatingRole: "topcoat", CoverageRate: 8, IsActive: true,
	})
	if err == nil {
		t.Fatal("insert with nonexistent category_id succeeded, want FK failure")
	}
}

func TestInactiveProductIsMissing(t *testing.T) {
	db := testDB(t)
	prods := &repository.ProductRepo{DB: db}
	// Deactivate a seeded product directly, then it must be invisible.
	if _, err := db.Exec(`UPDATE products SET is_active = 0 WHERE slug = 'son-loa-tao-gai-san-lotus-speaker-paint'`); err != nil {
		t.Fatal(err)
	}
	if _, err := prods.GetBySlugActive("son-loa-tao-gai-san-lotus-speaker-paint"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("inactive product lookup err = %v, want sql.ErrNoRows", err)
	}
}

func TestCalculatorMath(t *testing.T) {
	// SPEC §6.1: Total = Area x Coats / Coverage x 1.05.
	cases := []struct {
		area, coverage float64
		coats          int
		want           float64
		wantErr        bool
	}{
		{100, 10, 2, 21.0, false},
		{50, 8, 1, 6.5625, false},
		{0, 10, 2, 0, false},
		{100, 10, 0, 0, false},
		{-1, 10, 2, 0, true},
		{100, 0, 2, 0, true},
		{100, -5, 2, 0, true},
		{100, 10, -1, 0, true},
	}
	for _, tc := range cases {
		got, err := domain.CalculateLiters(tc.area, tc.coats, tc.coverage)
		if tc.wantErr && err == nil {
			t.Errorf("CalculateLiters(%v, %d, %v) err = nil, want error", tc.area, tc.coats, tc.coverage)
			continue
		}
		if !tc.wantErr {
			if err != nil {
				t.Errorf("CalculateLiters(%v, %d, %v) err = %v", tc.area, tc.coats, tc.coverage, err)
				continue
			}
			if diff := got - tc.want; diff > 1e-9 || diff < -1e-9 {
				t.Errorf("CalculateLiters(%v, %d, %v) = %v, want %v", tc.area, tc.coats, tc.coverage, got, tc.want)
			}
		}
	}
	// Packaging mix edge cases.
	if mix := domain.PackagingMixFor(0); mix != (domain.PackagingMix{}) {
		t.Errorf("PackagingMixFor(0) = %+v, want empty", mix)
	}
	mix := domain.PackagingMixFor(21.0)
	if mix.Pack18L != 1 || mix.Pack38L != 0 || mix.Pack08L != 4 {
		// 21L = 18L + 3L remainder -> 4 x 0.8L = 3.2L cover, total 21.2L.
		t.Errorf("PackagingMixFor(21) = %+v, want {1 0 4}", mix)
	}
	mix = domain.PackagingMixFor(20.0)
	if mix.Pack18L != 1 || mix.Pack08L != 3 {
		t.Errorf("PackagingMixFor(20) = %+v, want 18L x1 + 0.8L x3", mix)
	}
}

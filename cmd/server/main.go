// Command server runs the SonLotus web platform: public routes, admin
// CMS, uploads, sitemap/robots and 301 redirects, behind the fixed
// middleware chain (PLAN §6): recovery → logging → security headers →
// redirect → rate limit → admin auth → CSRF. It fails closed when
// ADMIN_USER / ADMIN_PASSWORD are missing or empty (approved security
// note). Graceful shutdown closes the database; no explicit WAL
// checkpoint is a gate (PLAN §11, Emily L4 = overcaution).
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"sonlotus/internal/config"
	"sonlotus/internal/database"
	"sonlotus/internal/handler"
	"sonlotus/internal/middleware"
	"sonlotus/internal/repository"
	"sonlotus/internal/service"
)

func main() {
	cfg := config.Load()
	production := cfg.AppEnv == "production"
	var logger *slog.Logger
	if production {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}
	slog.SetDefault(logger)
	if !cfg.AdminConfigured() {
		// Fail closed: never start with a missing/empty admin credential.
		logger.Error("server refuses to start: ADMIN_USER or ADMIN_PASSWORD missing or empty")
		os.Exit(1)
	}
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		logger.Error("create data dir", "error", err)
		os.Exit(1)
	}
	uploadDir := filepath.Join(cfg.DataDir, "uploads")
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		logger.Error("create upload dir", "error", err)
		os.Exit(1)
	}
	db, err := database.Open(cfg.DBPath)
	if err != nil {
		logger.Error("open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	if err := database.Migrate(db); err != nil {
		logger.Error("migrate", "error", err)
		os.Exit(1)
	}
	if err := database.BackfillFTS(db); err != nil {
		logger.Error("backfill FTS", "error", err)
		os.Exit(1)
	}
	// Shared seed function (PLAN §3): one seed path for server and
	// migrator --seed-only. Seed is a no-op when categories exist.
	if err := database.Seed(db); err != nil {
		logger.Error("seed", "error", err)
		os.Exit(1)
	}

	cats := &repository.CategoryRepo{DB: db}
	prods := &repository.ProductRepo{DB: db}
	arts := &repository.ArticleRepo{DB: db}
	inqs := &repository.InquiryRepo{DB: db}
	reds := &repository.RedirectRepo{DB: db}
	sessions := &repository.SessionRepo{DB: db}

	rend, err := handler.NewRenderer()
	if err != nil {
		logger.Error("public renderer", "error", err)
		os.Exit(1)
	}
	adminRend, err := handler.NewAdminRenderer()
	if err != nil {
		logger.Error("admin renderer", "error", err)
		os.Exit(1)
	}
	auth, err := middleware.NewAdminAuth(cfg.AdminUser, cfg.AdminPassword, sessions)
	if err != nil {
		logger.Error("admin auth", "error", err)
		os.Exit(1)
	}
	csrf := middleware.NewCSRFStore()

	mux := handler.NewMux(&handler.Deps{
		Categories: cats, Products: prods, Articles: arts, Inquiries: inqs, Redirects: reds,
		Catalog: &service.CatalogService{Categories: cats, Products: prods},
		Content: &service.ContentService{Articles: arts, Products: prods},
		SEO:     &service.SEOService{BaseURL: cfg.SiteURL},
		Render:  rend,
	})
	handler.RegisterAdmin(mux, &handler.AdminDeps{
		Categories: cats, Products: prods, Articles: arts, Inquiries: inqs, Redirects: reds,
		Auth: auth, CSRF: csrf, Render: adminRend, UploadDir: uploadDir, Production: production,
	})

	// Coarse abuse guard for the public inquiry POST only. It must
	// NOT share a bucket with admin login: inquiry spam from one IP
	// must never 429 admin login from that IP. The login failure
	// rule (5 failed / 15 min / IP) lives in the login handler
	// itself and reuses the same Limiter mechanism.
	chain := middleware.Chain(mux, middleware.ChainConfig{
		Logger:       logger,
		Production:   production,
		Redirects:    reds,
		IsRealRoute:  handler.IsRealRoute,
		Auth:         auth,
		CSRF:         csrf,
		ChainLimiter: middleware.NewLimiter(100, 15*time.Minute),
	})

	srv := &http.Server{Addr: cfg.Addr, Handler: chain, ReadHeaderTimeout: 10 * time.Second}
	go func() {
		logger.Info("listening", "addr", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown", "error", err)
	}
	// Deferred db.Close() runs on return: shutdown closes the database.
	logger.Info("shutdown complete")
}

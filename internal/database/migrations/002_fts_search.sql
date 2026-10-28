-- 002_fts_search.sql — SPEC §4.2 virtual table verbatim, plus the
-- PLAN §3 FTS synchronisation triggers (plan addition, not SPEC text).
-- Full-Text Search virtual table for instant catalog & technical article search
CREATE VIRTUAL TABLE IF NOT EXISTS search_index USING fts5(
    entity_type, -- 'product' or 'article'
    entity_id UNINDEXED,
    title,
    slug UNINDEXED,
    thumbnail_url UNINDEXED,
    summary,
    body_text,
    tokenize = 'unicode61 remove_diacritics 2'
);

-- Keep search_index in sync on product writes.
CREATE TRIGGER IF NOT EXISTS trg_products_fts_insert AFTER INSERT ON products BEGIN
    INSERT INTO search_index(entity_type, entity_id, title, slug, thumbnail_url, summary, body_text)
    VALUES ('product', NEW.id, NEW.name, NEW.slug, NEW.thumbnail_url, NEW.short_summary, NEW.description_html);
END;
CREATE TRIGGER IF NOT EXISTS trg_products_fts_update AFTER UPDATE ON products BEGIN
    DELETE FROM search_index WHERE entity_type = 'product' AND entity_id = OLD.id;
    INSERT INTO search_index(entity_type, entity_id, title, slug, thumbnail_url, summary, body_text)
    VALUES ('product', NEW.id, NEW.name, NEW.slug, NEW.thumbnail_url, NEW.short_summary, NEW.description_html);
END;
CREATE TRIGGER IF NOT EXISTS trg_products_fts_delete AFTER DELETE ON products BEGIN
    DELETE FROM search_index WHERE entity_type = 'product' AND entity_id = OLD.id;
END;

-- Keep search_index in sync on article writes.
CREATE TRIGGER IF NOT EXISTS trg_articles_fts_insert AFTER INSERT ON articles BEGIN
    INSERT INTO search_index(entity_type, entity_id, title, slug, thumbnail_url, summary, body_text)
    VALUES ('article', NEW.id, NEW.title, NEW.slug, NEW.thumbnail_url, NEW.excerpt, NEW.content_html);
END;
CREATE TRIGGER IF NOT EXISTS trg_articles_fts_update AFTER UPDATE ON articles BEGIN
    DELETE FROM search_index WHERE entity_type = 'article' AND entity_id = OLD.id;
    INSERT INTO search_index(entity_type, entity_id, title, slug, thumbnail_url, summary, body_text)
    VALUES ('article', NEW.id, NEW.title, NEW.slug, NEW.thumbnail_url, NEW.excerpt, NEW.content_html);
END;
CREATE TRIGGER IF NOT EXISTS trg_articles_fts_delete AFTER DELETE ON articles BEGIN
    DELETE FROM search_index WHERE entity_type = 'article' AND entity_id = OLD.id;
END;

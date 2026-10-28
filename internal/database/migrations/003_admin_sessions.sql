-- 003_admin_sessions.sql — plan-required addition BEYOND the SPEC §4
-- product/article schema (see Supervisor clarification 1): server-side
-- admin sessions for Phase 3. This is NOT a change to the SPEC §4
-- tables; it is a new table the approved security note requires
-- (token hash, expiry, created_at only — no plaintext token, no
-- password, no extra session payload is stored here).
CREATE TABLE IF NOT EXISTS admin_sessions (
    token_hash TEXT PRIMARY KEY,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_sessions_expiry ON admin_sessions(expires_at);

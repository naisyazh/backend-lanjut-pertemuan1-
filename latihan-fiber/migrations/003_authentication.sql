-- Migration untuk Authentication & Security
-- Tugas Mandiri Modul 5

-- Tambah column role pada table users
ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(20) NOT NULL DEFAULT 'user';

-- Buat table refresh_tokens
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index untuk performa
CREATE INDEX IF NOT EXISTS refresh_tokens_user_id_idx ON refresh_tokens (user_id);
CREATE INDEX IF NOT EXISTS refresh_tokens_token_hash_idx ON refresh_tokens (token_hash);
CREATE INDEX IF NOT EXISTS refresh_tokens_expires_at_idx ON refresh_tokens (expires_at);

-- Tambah index untuk role pada users
CREATE INDEX IF NOT EXISTS users_role_idx ON users (role);

-- Sample data untuk testing
INSERT INTO users (username, email, password, role, is_active) 
SELECT 'admin', 'admin@example.com', '$2a$12$K7Qx8q2Y1z3r4t5y6u7i8Wp9Qa1Ws2Ed3Rf4Tg5Yh6Uj7Ik8Ol9P', 'admin', true
WHERE NOT EXISTS (SELECT 1 FROM users WHERE username = 'admin');
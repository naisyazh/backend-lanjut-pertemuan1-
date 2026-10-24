-- Migration 007: tabel pembimbing tugas akhir
-- psql -U postgres -d praktikum_backend -f migrations/007_supervisors.sql

CREATE TABLE IF NOT EXISTS supervisors (
    id               SERIAL       PRIMARY KEY,
    name             VARCHAR(120) NOT NULL,
    quota            INTEGER      NOT NULL CHECK (quota > 0),
    supervised_count INTEGER      NOT NULL DEFAULT 0,
    version          INTEGER      NOT NULL DEFAULT 1,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    -- Jaring pengaman: supervised_count tidak boleh melebihi quota.
    -- Bila terpicu, artinya ada bug di kode aplikasi.
    CONSTRAINT supervisors_count_within_quota
        CHECK (supervised_count >= 0 AND supervised_count <= quota)
);

CREATE TABLE IF NOT EXISTS supervisions (
    id            BIGSERIAL   PRIMARY KEY,
    student_id    INTEGER     NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    supervisor_id INTEGER     NOT NULL REFERENCES supervisors(id) ON DELETE CASCADE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Satu mahasiswa hanya boleh memiliki SATU pembimbing.
    -- Constraint ini dijaga di level database sehingga dua request
    -- serentak tidak mungkin sama-sama berhasil mendaftarkan mahasiswa
    -- yang sama ke dua pembimbing berbeda.
    CONSTRAINT supervisions_student_unique UNIQUE (student_id)
);

CREATE INDEX IF NOT EXISTS supervisions_supervisor_idx ON supervisions (supervisor_id);

-- Data awal: beberapa pembimbing untuk uji beban
INSERT INTO supervisors (name, quota) VALUES
    ('Dr. Anugrah Anang Prasetyo', 5),
    ('Dr. Budi Santoso',           5),
    ('Dr. Citra Dewi',             5),
    ('Dr. Dimas Arya',             5),
    ('Dr. Eka Fitriana',           5)
ON CONFLICT DO NOTHING;

-- Permission baru untuk fitur pembimbing
INSERT INTO permissions (name, description) VALUES
    ('supervision:manage', 'Menetapkan dan mengganti pembimbing tugas akhir')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('admin', 'supervision:manage')
ON CONFLICT DO NOTHING;

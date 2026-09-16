CREATE TABLE IF NOT EXISTS permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS role_permissions (
    id SERIAL PRIMARY KEY,
    role_name VARCHAR(20) NOT NULL,
    permission_name VARCHAR(100) NOT NULL REFERENCES permissions(name) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(role_name, permission_name)
);

CREATE INDEX IF NOT EXISTS role_permissions_role_idx ON role_permissions (role_name);
CREATE INDEX IF NOT EXISTS role_permissions_permission_idx ON role_permissions (permission_name);

INSERT INTO permissions (name, description) VALUES
 ('student:list', 'Melihat daftar seluruh student'),
 ('student:read:any', 'Melihat data student mana pun'),
 ('student:create', 'Membuat data student baru'),
 ('student:update:any', 'Mengubah data student mana pun'),
 ('student:delete', 'Menghapus student')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_name, permission_name) VALUES
 -- admin: semua permission
 ('admin', 'student:list'),
 ('admin', 'student:read:any'),
 ('admin', 'student:create'),
 ('admin', 'student:update:any'),
 ('admin', 'student:delete'),
 ('staff', 'student:list'),
 ('staff', 'student:read:any'),
 ('staff', 'student:create')
ON CONFLICT DO NOTHING;

UPDATE students SET nim = COALESCE(nim, 'LEGACY') WHERE nim IS NULL OR nim = '';

ALTER TABLE students ADD COLUMN IF NOT EXISTS owner_id INTEGER;

UPDATE students SET owner_id = (
    SELECT id FROM users WHERE role = 'admin' LIMIT 1
) WHERE owner_id IS NULL;

ALTER TABLE students 
DROP CONSTRAINT IF EXISTS students_owner_id_fkey;

ALTER TABLE students 
ADD CONSTRAINT students_owner_id_fkey 
FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE students ALTER COLUMN owner_id SET NOT NULL;

CREATE INDEX IF NOT EXISTS students_owner_id_idx ON students (owner_id);
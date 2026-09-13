CREATE TABLE IF NOT EXISTS nilai (
    id SERIAL PRIMARY KEY,
    nama_mata_kuliah VARCHAR(100) NOT NULL,
    nilai DECIMAL(4,2) NOT NULL CHECK (nilai >= 0 AND nilai <= 4),
    student_id INTEGER NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS nilai_student_id_idx ON nilai (student_id);
CREATE INDEX IF NOT EXISTS nilai_mata_kuliah_idx ON nilai (LOWER(nama_mata_kuliah));

-- Sample data untuk testing
INSERT INTO nilai (nama_mata_kuliah, nilai, student_id) VALUES
('Praktikum A', 3.50, 1),
('Database', 3.75, 1),
('Algoritma', 3.25, 1),
('Praktikum A', 2.80, 2),
('Database', 3.90, 2),
('Web Programming', 3.60, 2);
-- Index untuk cursor pagination students.
-- Urutan column HARUS sama persis dengan ORDER BY pada query.
CREATE INDEX IF NOT EXISTS students_created_at_id_desc_idx
    ON students (created_at DESC, id DESC);

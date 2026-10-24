-- reset_loadtest.sql
-- Kembalikan state database ke kondisi awal sebelum setiap uji beban.
-- Jalankan: psql -U postgres -d praktikum_backend -f migrations/reset_loadtest.sql

-- Hapus semua enrollment akun uji beban (mhs001-mhs060)
DELETE FROM enrollments
WHERE user_id IN (SELECT id FROM users WHERE username LIKE 'mhs%');

-- Hitung ulang enrolled_count dari baris enrollments yang sebenarnya
UPDATE courses c
SET enrolled_count = (SELECT COUNT(*) FROM enrollments e WHERE e.course_id = c.id);

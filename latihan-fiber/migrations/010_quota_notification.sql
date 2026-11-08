-- Migration 010: Ganti idempotency key untuk mendukung notifikasi broadcast ke seluruh peserta (1 event outbox -> N notifikasi)
-- psql -U postgres -d praktikum_backend -f migrations/010_quota_notification.sql

-- 1. Hapus constraint lama yang hanya mengizinkan 1 notifikasi per event outbox
ALTER TABLE notifications DROP CONSTRAINT IF EXISTS notifications_outbox_unique;

-- 2. Tambahkan constraint baru yang mengombinasikan (outbox_id, user_id)
--    Dengan composite key ini, satu event outbox dapat dikirimkan ke banyak user,
--    namun setiap user dijamin HANYA menerima tepat 1 email/notifikasi untuk event tersebut.
ALTER TABLE notifications DROP CONSTRAINT IF EXISTS notifications_outbox_user_unique;
ALTER TABLE notifications ADD CONSTRAINT notifications_outbox_user_unique UNIQUE (outbox_id, user_id);

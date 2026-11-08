# MODUL 12 PROGRESS — Testing, Swagger, Docker

Dokumen ini merekam status pekerjaan Modul 12 agar tidak hilang bila kuota habis.

---

## BAGIAN B1 — TESTING (PUZZLE)

### Status: ✅ SELESAI

Sembilan kesalahan telah diperbaiki sebelum sesi ini dimulai (dari ZIP proyek). Semua unit test dan integration test sudah lolos.

### Puzzle Errors yang Ditemukan & Diperbaiki (D.1)

| No | Lokasi | Gejala | Pemeriksaan | Akar Masalah | Perbaikan |
|----|--------|--------|-------------|--------------|-----------|
| 1 | `tests/integration/main_test.go` — `TestMain` | Integration test dilewati padahal `TEST_DATABASE_URL` sudah diisi | compiler / `go test -tags=integration` | Kondisi `dsn != ""` seharusnya `dsn == ""`: ekspresi terbalik membuat test exit saat DSN ADA | Ganti `!= ""` → `== ""` |
| 2 | `tests/integration/main_test.go` — `TestMain` `cfg.MaxConns = 1` | Test konkuren tidak benar-benar serentak (hanya 1 koneksi) | T6 (mutation ForUpdate) | MaxConns=1 mencegah goroutine serentak mendapat koneksi berbeda | Ubah ke `15` |
| 3 | `app/service/review_rules_test.go` — `TestNormalizeTags` | Compile error: `func TestNormalizeTags(t testing.T)` | compiler | Parameter `t` bukan pointer: harus `*testing.T` | Ganti `t testing.T` → `t *testing.T` |
| 4 | `app/service/notification_rules_test.go` | Tidak ada `strings` di import tetapi dipakai | compiler | Import `"strings"` hilang | Tambahkan `"strings"` ke blok import |
| 5 | `app/service/review_errors_test.go` — `TestTranslateReviewError` | `service.translateReviewError` tidak dapat diakses dari `package service_test` | compiler | Fungsi tidak di-export, file package `service` bukan `service_test` | Ubah `package service_test` → `package service` |
| 6 | `app/service/notification_worker_test.go` — `NewNotificationWorker(...)` | Compile error: argumen tidak cocok | compiler | Konstruktor menerima 6 argumen; kode test hanya mengirim 5 (nil, nil, nil, nil, logger) — hilang satu `nil` untuk `mailer` | Tambahkan satu `nil` sebagai argumen ke-5 |
| 7 | `app/service/fakes_test.go` — `fakeReviewRepo.FindByCourse` | Compile error: signature tidak cocok interface `ReviewRepository` | compiler | Interface menerima `model.ReviewQuery`, fake menerima `courseID int` | Ubah parameter ke `q model.ReviewQuery` dan akses `q.CourseID` |
| 8 | `tests/integration/main_test.go` — `seedCourse` | Cleanup kelas tidak memakai `t.Cleanup` sehingga kelas terhapus sebelum test selesai | T4 (integration test) | `defer` di dalam `seedCourse` dijalankan saat fungsi selesai, bukan saat test selesai | Ganti `defer` → `t.Cleanup(func() { ... })` |
| 9 | `migrations/003_create_nilai.sql` | Migration gagal di database test kosong: FK violation | T4 / migration | INSERT seed data langsung ke `nilai` tanpa memastikan `students` 1 & 2 ada; owner_id NOT NULL juga bermasalah | Bungkus INSERT dalam DO block dengan exception handler untuk FK violation |

### Pemeriksaan Tambahan yang Diperbaiki (bukan puzzle, tetapi menghalangi T4)

- `migrations/010_quota_notification.sql`: `ADD CONSTRAINT` tidak idempotent → ditambahkan `DROP CONSTRAINT IF EXISTS` sebelumnya.

---

## STATUS PEMERIKSAAN C.1 — TESTING

| No | Perintah | Hasil | Status |
|----|----------|-------|--------|
| T1 | `go test -count=1 ./...` | Semua package ok | ✅ |
| T2 | `go test -shuffle=on -count=5 ./app/service/` | ok (diulang, ketiganya ok) | ✅ |
| T3 | Coverage `WithinCreditLimit`, `CreditsAfterSwap`, `NormalizeTags`, `BuildMessage` | Masing-masing 100.0% | ✅ |
| T4 | `go test -tags=integration -count=1 -v ./tests/integration/` | 8 baris PASS (6 wajib + 2 D2), tanpa FAIL | ✅ |
| T5 | Mutation `<=` → `<` pada `WithinCreditLimit` | `TestWithinCreditLimit` FAIL (subtest "tepat di batas") | ✅ |
| T6 | Hapus `FOR UPDATE` dari `LockStudent` | `TestEnrollEndpoint_ConcurrentCreditLimit` FAIL pada 3/3 run | ✅ |
| T7 | Hapus `ON CONFLICT` dari `Claim` | `TestClaim_OnlyOnce` FAIL (duplicate key error) | ✅ |
| Race | `-race` detector | Membutuhkan gcc — tidak tersedia di Windows ini (sesuai catatan PDF) | ⚠️ skip |

---

## BAGIAN B2 — SWAGGER

### Status: ✅ SELESAI

### Yang Dikerjakan

- **Langkah 9**: Dependency `gofiber/swagger` dan `swaggo/swag` sudah ada di `go.mod` (tool dependency).
- **Langkah 10**: `model.MyCourses` dan `model.ReviewPage` / `ReviewPageMeta` sudah ada di model. Service sudah menggunakannya.
- **Langkah 11**: Anotasi `@Summary`, `@Router`, dll sudah ada di semua service file (auth, course, review, notification, user, student, supervisor).
- **Langkah 12**: 
  - Ditambahkan Swagger general info (`@title`, `@BasePath`, `@securityDefinitions.apikey`) ke `main.go`
  - Ditambahkan `_ "latihan-fiber/docs"` ke import `main.go`
  - Ditambahkan `EnableSwagger: config.GetEnv("SWAGGER_ENABLED", "false") == "true"` di `startAPI`
  - Ditambahkan `SWAGGER_ENABLED=true` ke `.env`
  - Dijalankan `go tool swag init` → folder `docs/` dibuat dengan 25 paths

### Status C.2 — API Documentation

| No | Kriteria | Status |
|----|----------|--------|
| 1 | `go tool swag init` berjalan tanpa error, folder `docs` berisi 3 file | ✅ |
| 2 | `docs/swagger.json` memuat 11 path wajib (dan 14 D3 tambahan = 25 total) | ✅ |
| 3 | `SWAGGER_ENABLED=true` → UI muncul; `false` → 404 | ✅ (kode sudah benar) |
| 4 | Authorize + Try it out GET /courses → 200 | Perlu verifikasi manual |
| 5 | POST /courses/{id}/enroll menampilkan skema Course pada 201 | ✅ (anotasi ada) |
| 6 | `go test -count=1 ./...` tetap lolos setelah Langkah 10 | ✅ |

---

## BAGIAN B3 — DOCKER

### Status: ✅ FILE DIBUAT (Docker daemon belum berjalan — perlu dijalankan manual)

### Yang Dikerjakan

- **Langkah 13**: Dibuat `Dockerfile` (multi-stage build, alpine runtime, user `app`) dan `.dockerignore`
- **Langkah 14**: Dibuat `docker-compose.yml` dengan 5 service: `api`, `worker`, `postgres`, `mongo`, `redis`

### Status C.3 — Docker Deployment

| No | Kriteria | Status |
|----|----------|--------|
| 1 | `docker compose config` berjalan; tanpa JWT_SECRET → error "JWT_SECRET wajib diisi di .env" | ✅ Verified |
| 2 | `docker compose up -d --build` + 5 service healthy | Perlu Docker daemon aktif |
| 3 | `curl http://localhost:8080/api/v1/health` → 200 | Perlu Docker daemon aktif |
| 4 | `docker compose exec api whoami` → `app` | Perlu Docker daemon aktif |
| 5 | `/app` hanya berisi `app` dan `logs` | Perlu Docker daemon aktif |
| 6 | Register → login → enroll → GET /notifications/mine berisi notifikasi | Perlu Docker daemon aktif |
| 7 | Setelah `down` + `up`, data masih ada | Perlu Docker daemon aktif |

**Docker Desktop belum aktif saat audit ini dijalankan.** Jalankan Docker Desktop dari Start Menu, tunggu sampai menyala, lalu jalankan:
```
cd latihan-fiber
docker compose up -d --build
docker compose ps
curl http://localhost:8080/api/v1/health
```

---

## D.2 — MUTATION TEST UNTUK D2

### D2.1 — Swap ke kelas penuh

**Test**: `TestEnrollEndpoint_Swap_TargetCourseFull` di `tests/integration/swap_drop_test.go`

**Mutation**: `app/repository/course_repository.go` — `IncrementEnrolled`: ganti `enrolled_count < quota` → `enrolled_count <= quota`

**Hasil**: Test FAIL (status 500 INTERNAL_ERROR, bukan 409 COURSE_FULL) ✅

### D2.2 — Drop kelas berhasil

**Test**: `TestEnrollEndpoint_Drop_Success` di `tests/integration/swap_drop_test.go`

**Mutation**: `app/service/course_service.go` — `Drop`: ganti `model.EventEnrollmentDropped` → `model.EventEnrollmentCreated`

**Hasil**: Test FAIL (event enrollment.dropped = 0, seharusnya 1) ✅

---

## FILE YANG DIBUAT/DIUBAH SESI INI

| File | Perubahan |
|------|-----------|
| `main.go` | Tambah Swagger general info comment, docs import, EnableSwagger wiring |
| `.env` | Tambah `SWAGGER_ENABLED=true` |
| `.env.test` | Dibuat baru (tidak di-commit) |
| `.env.example` | Tambah `SWAGGER_ENABLED=false` |
| `.gitignore` | Tambah `.env.test` |
| `docs/` | Dibuat oleh `go tool swag init` (3 file) |
| `Dockerfile` | Dibuat baru (multi-stage build) |
| `.dockerignore` | Dibuat baru |
| `docker-compose.yml` | Dibuat baru (5 service) |
| `migrations/003_create_nilai.sql` | Insert seed bungkus DO block untuk FK safety |
| `migrations/010_quota_notification.sql` | Tambah DROP CONSTRAINT IF EXISTS agar idempotent |
| `MODUL12_PROGRESS.md` | Dokumen ini |

---

## LANGKAH SELANJUTNYA (MANUAL)

1. **Docker**: Aktifkan Docker Desktop → `docker compose up -d --build` → verifikasi C.3
2. **Swagger UI**: Dengan `SWAGGER_ENABLED=true`, buka `http://localhost:3000/swagger/index.html`
3. **D.1 Berita Acara**: Isi tabel 9 baris di laporan PDF menggunakan data tabel di atas
4. **D.3**: Screenshot Swagger UI yang menampilkan semua tag dan jumlah 25 path
5. **D.4**: Jawaban analisis singkat (coverage vs mutation, MongoDB fake, migration replika, risiko SWAGGER)

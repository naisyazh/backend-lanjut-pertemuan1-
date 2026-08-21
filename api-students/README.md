# API Students

REST API untuk manajemen data mahasiswa (CRUD).

## Endpoint

| Method | URL | Deskripsi |
|--------|-----|-----------|
| GET | /students | List semua students |
| GET | /students/:id | Detail student |
| POST | /students | Tambah student |
| PUT | /students/:id | Update student (semua field) |
| PATCH | /students/:id | Update student (sebagian field) |
| DELETE | /students/:id | Hapus student |

## Query String (GET /students)
- `page` & `limit` - pagination
- `search` - cari berdasarkan nama
- `aktif` - filter aktif/tidak

## Jalanin
```bash
go run .
```

Server: `http://localhost:3000`

---
Naisya (434241068)

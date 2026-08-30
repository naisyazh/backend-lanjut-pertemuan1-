# API Students & Users

REST API untuk manage data mahasiswa dan user pake PostgreSQL.

## Setup

1. Install dependencies
```bash
go mod tidy
```

2. Buat database
```bash
psql -U postgres -c "CREATE DATABASE praktikum_backend;"
psql -U postgres -d praktikum_backend -f migrations/001_create_users.sql
psql -U postgres -d praktikum_backend -f migrations/002_create_students.sql
```

3. Copy `.env.example` jadi `.env` terus isi password database

4. Jalankan
```bash
go run .
```

## Environment Variables

```
APP_PORT=3001
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=praktikum_backend
DB_SSLMODE=disable
DB_MAX_CONNS=10
```

## Endpoints

### Users
- `GET /api/v1/users` - List users
- `GET /api/v1/users/:id` - Get user
- `POST /api/v1/users` - Create user
- `PUT /api/v1/users/:id` - Update user
- `PATCH /api/v1/users/:id` - Update sebagian
- `DELETE /api/v1/users/:id` - Delete user

### Students
- `GET /api/v1/students` - List students
- `GET /api/v1/students/:id` - Get student
- `POST /api/v1/students` - Create student
- `PUT /api/v1/students/:id` - Update student
- `PATCH /api/v1/students/:id` - Update sebagian
- `DELETE /api/v1/students/:id` - Delete student

## Contoh

Create student:
```bash
curl -X POST http://localhost:3001/api/v1/students \
  -H "Content-Type: application/json" \
  -d '{"nim":"123456789","name":"Naisya","grade":"A"}'
```

Get all students:
```bash
curl http://localhost:3001/api/v1/students
```

Search:
```bash
curl "http://localhost:3001/api/v1/students?search=naisya"
```

Pagination:
```bash
curl "http://localhost:3001/api/v1/students?page=1&limit=10"
```

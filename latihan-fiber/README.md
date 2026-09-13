# Backend Programming Lanjut - Module 5: Authentication & Security

Backend API dengan sistem authentication menggunakan JWT dan bcrypt untuk keamanan password.

## 🚀 Features

### Authentication System
- ✅ **JWT Authentication** - Access token (15 min) & Refresh token (7 days)
- ✅ **Password Security** - bcrypt hashing dengan cost 12
- ✅ **Rate Limiting** - Protection terhadap brute force attacks
- ✅ **Mass Assignment Prevention** - Role ditentukan server
- ✅ **Token Rotation** - Refresh token security

### API Endpoints

#### Public Endpoints
- `POST /api/v1/auth/register` - Registrasi user baru
- `POST /api/v1/auth/login` - Login user (dengan rate limiting)
- `POST /api/v1/auth/refresh` - Refresh access token
- `POST /api/v1/auth/logout` - Logout user
- `GET /api/v1/health` - Health check

#### Protected Endpoints (Requires Authentication)
- `GET /api/v1/auth/me` - Get user profile
- `GET /api/v1/students` - List students
- `POST /api/v1/students` - Create student
- `GET /api/v1/nilai` - List grades
- `POST /api/v1/nilai` - Create grade

## 🛠 Technology Stack

- **Backend**: Go (Golang) dengan Fiber framework
- **Database**: PostgreSQL
- **Authentication**: JWT dengan HMAC SHA-256
- **Password Hashing**: bcrypt
- **Testing**: Go testing framework

## 🔧 Setup & Installation

### 1. Clone Repository
```bash
git clone [repository-url]
cd latihan-fiber
```

### 2. Environment Configuration
Copy `.env.example` ke `.env` dan sesuaikan:
```env
APP_PORT=3002
DB_HOST=127.0.0.1
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=praktikum_backend
JWT_SECRET=your-secret-key
JWT_ISSUER=praktikum-backend
```

### 3. Database Migration
```bash
psql -h 127.0.0.1 -U postgres -d praktikum_backend -f migrations/001_create_users.sql
psql -h 127.0.0.1 -U postgres -d praktikum_backend -f migrations/002_create_students.sql
psql -h 127.0.0.1 -U postgres -d praktikum_backend -f migrations/003_authentication.sql
```

### 4. Install Dependencies
```bash
go mod tidy
```

### 5. Run Application
```bash
go run .
```

Server akan berjalan di `http://localhost:3002`

## 🧪 Testing

### Unit Tests
```bash
go test ./app/service/... -v
```

### API Testing Examples

#### Register
```bash
curl -X POST http://localhost:3002/api/v1/auth/register \
-H "Content-Type: application/json" \
-d '{"username":"testuser","email":"test@example.com","password":"Test123!@"}'
```

#### Login
```bash
curl -X POST http://localhost:3002/api/v1/auth/login \
-H "Content-Type: application/json" \
-d '{"username":"testuser","password":"Test123!@"}'
```

#### Access Protected Endpoint
```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
http://localhost:3002/api/v1/students
```

## 🔐 Security Features

### Password Security
- **bcrypt hashing** dengan cost 12
- **Password strength validation**:
  - Minimal 8 karakter
  - Harus mengandung huruf dan angka
  - Tidak boleh password umum/lemah
- **Timing attack protection** untuk login

### JWT Security
- **Short-lived access tokens** (15 menit)
- **Refresh token rotation**
- **Proper algorithm validation** (HMAC SHA-256)
- **Issuer validation**
- **Expiration validation**

### API Security
- **Authentication middleware** untuk endpoint protection
- **Rate limiting** (5 percobaan login per menit per IP)
- **Mass assignment prevention**
- **CORS configuration**
- **Proper error handling** tanpa information disclosure

## 📁 Project Structure

```
latihan-fiber/
├── app/
│   ├── model/          # Data models dan structs
│   ├── repository/     # Database access layer
│   └── service/        # Business logic layer
├── config/             # Application configuration
├── database/           # Database connection
├── helper/             # Utility functions (JWT, security, etc.)
├── middleware/         # HTTP middleware (auth, logging, etc.)
├── migrations/         # Database migrations
├── route/              # Route definitions
├── logs/               # Application logs
├── .env                # Environment variables
├── go.mod              # Go module dependencies
└── main.go             # Application entry point
```

## 📊 Test Results

### Unit Tests Coverage
- ✅ **Password Validation**: 7 test cases
- ✅ **Register Validation**: 5 test cases
- ✅ **Login Validation**: 4 test cases
- ✅ **Student Validation**: 4+ test cases
- ✅ **Total**: 20+ test cases - **ALL PASSED**

### Security Testing
- ✅ **Authentication Protection**: Protected endpoints block unauthorized access
- ✅ **JWT Validation**: Invalid/expired tokens properly rejected
- ✅ **Rate Limiting**: Login brute force protection working
- ✅ **Password Hashing**: bcrypt implementation secure

## 📚 Documentation

Laporan lengkap implementasi tersedia di: `LAPORAN_MODUL5_AUTHENTICATION.md`

## 👨‍💻 Author

**Mata Kuliah**: Backend Programming Lanjut  
**Module**: 5 - Authentication & Security
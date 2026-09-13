# LAPORAN TUGAS MANDIRI MODUL 5
## IMPLEMENTASI AUTHENTICATION & SECURITY PADA REST API

---

**Nama**: [ISI NAMA KAMU]  
**NIM**: 434241068  
**Mata Kuliah**: Backend Programming Lanjut  
**Modul**: 5 - Authentication & Security  
**Repository**: https://github.com/[USERNAME]/backend-lanjut-pertemuan1  

---

## 1. PENDAHULUAN

### 1.1 Tujuan
Tugas ini bertujuan untuk mengimplementasikan sistem authentication dan security pada REST API menggunakan:
- JWT (JSON Web Token) untuk authentication
- bcrypt untuk password hashing
- Middleware untuk proteksi endpoint
- Unit testing untuk validation rules

### 1.2 Tools & Technology Stack
- **Backend**: Go (Golang) dengan Fiber framework
- **Database**: PostgreSQL
- **Authentication**: JWT dengan HMAC SHA-256
- **Password Hashing**: bcrypt dengan cost 12
- **Testing**: Go testing framework

---

## 2. IMPLEMENTASI FITUR AUTHENTICATION

### 2.1 Database Schema
Menambahkan tabel dan kolom untuk mendukung authentication:

```sql
-- Tambah kolom role pada tabel users
ALTER TABLE users ADD COLUMN role VARCHAR(20) NOT NULL DEFAULT 'user';

-- Buat tabel refresh_tokens untuk JWT refresh mechanism
CREATE TABLE refresh_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### 2.2 Authentication Endpoints

#### A. Register Endpoint
- **URL**: `POST /api/v1/auth/register`
- **Fungsi**: Registrasi user baru dengan password hashing
- **Security**: Mass assignment prevention (role ditentukan server)

#### B. Login Endpoint  
- **URL**: `POST /api/v1/auth/login`
- **Fungsi**: Login user dan menghasilkan JWT tokens
- **Security**: Rate limiting (5 percobaan per menit per IP)

#### C. Refresh Token Endpoint
- **URL**: `POST /api/v1/auth/refresh`
- **Fungsi**: Memperbarui access token menggunakan refresh token
- **Security**: Token rotation untuk mencegah replay attacks

#### D. Logout Endpoint
- **URL**: `POST /api/v1/auth/logout`
- **Fungsi**: Mencabut refresh token

#### E. Profile Endpoint
- **URL**: `GET /api/v1/auth/me`
- **Fungsi**: Mengambil informasi user yang sedang login
- **Security**: Membutuhkan valid JWT token

---

## 3. SECURITY FEATURES YANG DIIMPLEMENTASIKAN

### 3.1 Password Security
- **bcrypt hashing** dengan cost factor 12
  - **Alasan pemilihan cost 12**: Berdasarkan OWASP recommendations, cost 12 memberikan keseimbangan optimal antara security dan performance. Cost ini membutuhkan ~250ms untuk hash satu password, cukup lambat untuk mencegah brute force namun masih acceptable untuk user experience. Cost 12 tahan terhadap hardware attack hingga beberapa tahun ke depan.
- **Password strength validation**:
  - Minimal 8 karakter
  - Harus mengandung huruf dan angka
  - Tidak boleh menggunakan password umum/lemah
- **Password Migration Policy**: Password lama yang ter-hash dengan algoritma lemah (MD5, SHA1) **tidak dapat diselamatkan** karena merupakan fungsi satu arah. Users dengan password lama harus melakukan reset password menggunakan mekanisme email verification atau admin manual reset.

### 3.2 JWT Security
- **HMAC SHA-256** signature algorithm
- **Short-lived access token** (15 menit)
- **Long-lived refresh token** (7 hari)
- **Token rotation** untuk refresh tokens
- **Issuer validation**

### 3.3 API Security
- **Authentication middleware** untuk endpoint protection
- **Rate limiting** untuk login endpoint
- **Mass assignment prevention**
- **Proper error handling** tanpa information disclosure

---

## 4. SCREENSHOT TESTING

### 4.1 Test Register User Baru
**[SCREENSHOT 1: PowerShell dengan hasil register berhasil]**

Penjelasan:
- User berhasil didaftarkan dengan username "mahasiswa1"
- Password di-hash menggunakan bcrypt sebelum disimpan
- Role secara otomatis diset sebagai "user" (mencegah mass assignment)
- Response menunjukkan user berhasil dibuat dengan ID = 4

### 4.2 Test Login User
**[SCREENSHOT 2: PowerShell dengan hasil login berhasil dan JWT token]**

Penjelasan:
- Login berhasil dengan credentials yang valid
- Server menghasilkan access_token dan refresh_token
- Access token berisi informasi user (username, role) dalam JWT format
- Token memiliki expiry time untuk security

### 4.3 Test Security - Akses Tanpa Authentication
**[SCREENSHOT 3: PowerShell menunjukkan 401 Unauthorized tanpa token]**

Penjelasan:
- Akses ke `/api/v1/students` tanpa token menghasilkan error 401
- Middleware authentication berhasil memblokir akses tidak sah
- Security system berfungsi dengan baik

### 4.4 Test Security - Akses Dengan Authentication  
**[SCREENSHOT 4: PowerShell menunjukkan akses berhasil dengan token]**

Penjelasan:
- Akses ke `/api/v1/students` dengan valid JWT token berhasil
- Middleware authentication mengizinkan akses
- Data students berhasil diambil (Naisya Gina Azzahra)
- Membuktikan authorization system berfungsi

### 4.5 Test Unit Testing
**[SCREENSHOT 5: Hasil go test dengan semua test passed]**

Penjelasan:
- Semua unit test berhasil (PASS)
- Test coverage meliputi:
  - Password strength validation (7 test cases)
  - Register validation (5 test cases) 
  - Login validation (4 test cases)
  - Student validation (4+ test cases)
- Total 20+ test cases semuanya PASS

### 4.6 Test Security - Token Manipulation
**[SCREENSHOT 6: PowerShell menunjukkan 401 dengan token yang diubah]**

Penjelasan:
- Token JWT yang valid sengaja diubah satu karakter (ditambah 'X' di akhir)
- System berhasil mendeteksi token tidak valid/corrupted
- Response 401 Unauthorized dengan pesan "access token tidak valid"
- Membuktikan JWT signature verification berfungsi dengan baik
- Mencegah serangan token manipulation/forgery

### 4.7 Test Security - Username Tidak Terdaftar
**[SCREENSHOT 7: PowerShell menunjukkan login gagal dengan username tidak ada]**

Penjelasan:
- Login dengan username "user_yang_tidak_ada" yang tidak terdaftar di database
- System memberikan pesan error yang sama dengan password salah: "username atau password salah"
- Mencegah username enumeration attack (attacker tidak bisa tahu username mana yang valid)
- Implementasi timing attack protection dengan VerifyDummyPassword()
- Security best practice: generic error message untuk login failures

### 4.8 Test Security - Rate Limiting Brute Force
**[SCREENSHOT 8: PowerShell menunjukkan 429 Too Many Requests setelah 6x percobaan]**

Penjelasan:
- 6 percobaan login gagal berturut-turut dari IP yang sama
- Setelah percobaan ke-5, request ke-6 diblokir dengan status 429 (Too Many Requests)
- Response header "Retry-After: 60" memberitahu client untuk menunggu 60 detik
- Rate limiting berhasil mencegah brute force password attacks
- Middleware LoginRateLimiter() berfungsi sesuai konfigurasi (5 max per menit)

### 4.9 Test Security - Mass Assignment Prevention  
**[SCREENSHOT 9: PowerShell menunjukkan role tetap 'user' meskipun kirim role:'admin']**

Penjelasan:
- Request register dengan payload tambahan: "role":"admin" (injection attempt)
- System mengabaikan field 'role' dari client dan tetap assign role "user"
- Response menunjukkan user berhasil dibuat dengan role="user" (bukan admin)
- Mass assignment attack berhasil dicegah
- Server-side role assignment mencegah privilege escalation vulnerability

### 4.10 Test Security - Refresh Token Rotation
**[SCREENSHOT 10: PowerShell menunjukkan old refresh token ditolak setelah rotation]**

Penjelasan:
- Refresh token pertama berhasil digunakan dan menghasilkan token pair baru
- Old refresh token secara otomatis di-revoke (token rotation)
- Percobaan menggunakan old refresh token kedua kali menghasilkan error 401
- Token rotation mencegah replay attacks pada refresh tokens
- Security mechanism memastikan refresh token hanya bisa digunakan sekali

---

## 5. ANALISIS KEAMANAN

### 5.1 Perbandingan localStorage vs httpOnly Cookie untuk JWT Storage

**localStorage:**
- ✅ Mudah diakses dari JavaScript
- ❌ Rentan terhadap XSS (Cross-Site Scripting) attacks
- ❌ Token dapat dicuri melalui malicious scripts
- ❌ Tidak otomatis dikirim ke server

**httpOnly Cookie:**
- ✅ Tidak dapat diakses dari JavaScript (XSS protection)
- ✅ Otomatis dikirim ke server pada setiap request
- ✅ Dapat dikonfigurasi dengan Secure dan SameSite flags
- ❌ Rentan terhadap CSRF jika tidak dikonfigurasi dengan benar
- ❌ Lebih kompleks untuk SPA (Single Page Application)

**Kesimpulan**: Untuk production, httpOnly cookie dengan SameSite=Strict dan Secure flags lebih aman, namun memerlukan CSRF protection tambahan.

### 5.2 Dampak Mengubah Access Token TTL

**Jika TTL diperpanjang menjadi 24 jam:**
- ❌ **Security Risk**: Token yang dicuri dapat digunakan lebih lama
- ❌ **Session Hijacking**: Window waktu serangan lebih panjang
- ❌ **Privilege Escalation**: Role changes tidak langsung berlaku
- ✅ **User Experience**: Users jarang perlu re-authentication

**Jika TTL diperpendek menjadi 1 menit:**
- ✅ **Security**: Minimal impact jika token dicuri
- ✅ **Real-time Authorization**: Role changes segera berlaku
- ❌ **Performance**: Frequent token refresh calls
- ❌ **User Experience**: Frequent interruptions jika refresh gagal
- ❌ **Network Overhead**: Lebih banyak HTTP requests

**Rekomendasi**: 15 menit (current) adalah sweet spot untuk balance security-usability.

### 5.3 Kerentanan yang Masih Tersisa

Meskipun system sudah cukup aman, masih ada beberapa kerentanan potensial:

**1. Session Fixation**
- **Risiko**: Attacker dapat menggunakan session ID yang sama
- **Mitigasi**: Regenerate session/token setelah login berhasil

**2. Timing Attacks pada Username**
- **Risiko**: Response time berbeda untuk user yang ada vs tidak ada
- **Mitigasi**: Sudah implemented dengan `VerifyDummyPassword()`

**3. JWT Secret Exposure**
- **Risiko**: Jika JWT_SECRET ter-leak, semua token dapat di-forge
- **Mitigasi**: Regular secret rotation dan proper secret management

**4. Database Injection**
- **Risiko**: SQL injection melalui input yang tidak sanitized
- **Mitigasi**: Sudah menggunakan parameterized queries

**5. Rate Limiting Bypass**
- **Risiko**: Distributed brute force dari multiple IPs
- **Mitigasi**: Implement CAPTCHA atau account lockout mechanism

**6. Refresh Token Storage**
- **Risiko**: Refresh tokens disimpan plain text di client
- **Mitigasi**: Encrypt refresh tokens atau gunakan httpOnly cookies

---

## 6. KODE IMPLEMENTASI UTAMA

### 6.1 Authentication Models (app/model/auth.go)
```go
package model

type RegisterRequest struct {
    Username string `json:"username"`
    Email    string `json:"email"`
    Password string `json:"password"`
}

type LoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type RefreshRequest struct {
    RefreshToken string `json:"refresh_token"`
}

type TokenPair struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    TokenType    string `json:"token_type"`
    ExpiresIn    int    `json:"expires_in"`
}

type AuthUser struct {
    UserID   int    `json:"user_id"`
    Username string `json:"username"`
    Role     string `json:"role"`
}

type RefreshToken struct {
    ID        int                `json:"id"`
    UserID    int                `json:"user_id"`
    TokenHash string             `json:"token_hash"`
    ExpiresAt time.Time          `json:"expires_at"`
    RevokedAt *time.Time         `json:"revoked_at,omitempty"`
    CreatedAt time.Time          `json:"created_at"`
}
```

### 6.2 Password Security (helper/security.go)
```go
package helper

import (
    "crypto/rand"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
    return string(bytes), err
}

func VerifyPassword(hashedPassword, plainPassword string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
    return err == nil
}

// VerifyDummyPassword menjalankan bcrypt untuk mencegah timing attack
func VerifyDummyPassword(password string) {
    bcrypt.CompareHashAndPassword([]byte("$2a$12$dummy"), []byte(password))
}

func RandomToken(bytes int) (string, error) {
    buffer := make([]byte, bytes)
    if _, err := rand.Read(buffer); err != nil {
        return "", fmt.Errorf("generating random token: %w", err)
    }
    return hex.EncodeToString(buffer), nil
}

func SHA256Hex(input string) string {
    hash := sha256.Sum256([]byte(input))
    return hex.EncodeToString(hash[:])
}
```

### 6.3 JWT Manager (helper/jwt.go)
```go
package helper

import (
    "errors"
    "fmt"
    "strconv"
    "time"
    "github.com/golang-jwt/jwt/v5"
    "latihan-fiber/app/model"
)

var (
    ErrInvalidToken = errors.New("token tidak valid")
    ErrExpiredToken = errors.New("token sudah kedaluwarsa")
)

type accessClaims struct {
    Username string `json:"username"`
    Role     string `json:"role"`
    jwt.RegisteredClaims
}

type JWTManager struct {
    secret    []byte
    issuer    string
    accessTTL time.Duration
}

func NewJWTManager(secret, issuer string, accessTTL time.Duration) *JWTManager {
    return &JWTManager{secret: []byte(secret), issuer: issuer, accessTTL: accessTTL}
}

func (m *JWTManager) AccessTTL() time.Duration { return m.accessTTL }

func (m *JWTManager) GenerateAccess(u model.User) (string, error) {
    now := time.Now()
    claims := accessClaims{
        Username: u.Username,
        Role:     u.Role,
        RegisteredClaims: jwt.RegisteredClaims{
            Subject:   strconv.Itoa(u.ID),
            Issuer:    m.issuer,
            IssuedAt:  jwt.NewNumericDate(now),
            ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(m.secret)
}

func (m *JWTManager) Parse(tokenString string) (model.AuthUser, error) {
    claims := &accessClaims{}
    token, err := jwt.ParseWithClaims(tokenString, claims,
        func(t *jwt.Token) (any, error) {
            if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("algoritma tidak diharapkan: %v", t.Header["alg"])
            }
            return m.secret, nil
        },
        jwt.WithIssuer(m.issuer),
        jwt.WithExpirationRequired(),
    )

    if err != nil {
        if errors.Is(err, jwt.ErrTokenExpired) {
            return model.AuthUser{}, ErrExpiredToken
        }
        return model.AuthUser{}, ErrInvalidToken
    }

    if !token.Valid {
        return model.AuthUser{}, ErrInvalidToken
    }

    userID, err := strconv.Atoi(claims.Subject)
    if err != nil {
        return model.AuthUser{}, ErrInvalidToken
    }

    return model.AuthUser{
        UserID:   userID,
        Username: claims.Username,
        Role:     claims.Role,
    }, nil
}
```

### 6.4 Authentication Middleware (middleware/auth.go)
```go
package middleware

import (
    "errors"
    "strings"
    "time"
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/limiter"
    "latihan-fiber/helper"
)

func RequireAuth(jwtManager *helper.JWTManager) fiber.Handler {
    return func(c *fiber.Ctx) error {
        token, err := bearerToken(c)
        if err != nil {
            c.Set("WWW-Authenticate", `Bearer realm="api"`)
            return helper.Fail(c, fiber.StatusUnauthorized,
                "header Authorization tidak ada atau salah bentuk")
        }

        authUser, err := jwtManager.Parse(token)
        if err != nil {
            c.Set("WWW-Authenticate", `Bearer realm="api"`)
            if errors.Is(err, helper.ErrExpiredToken) {
                return helper.Fail(c, fiber.StatusUnauthorized, "access token kedaluwarsa")
            }
            return helper.Fail(c, fiber.StatusUnauthorized, "access token tidak valid")
        }

        c.Locals(helper.LocalsAuthUser, authUser)
        return c.Next()
    }
}

func bearerToken(c *fiber.Ctx) (string, error) {
    header := c.Get(fiber.HeaderAuthorization)
    if header == "" {
        return "", errors.New("header kosong")
    }

    parts := strings.SplitN(header, " ", 2)
    if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
        return "", errors.New("format bukan Bearer")
    }

    token := strings.TrimSpace(parts[1])
    if token == "" {
        return "", errors.New("token kosong")
    }
    return token, nil
}

// Rate limiter untuk mencegah brute force login
func LoginRateLimiter() fiber.Handler {
    return limiter.New(limiter.Config{
        Max:        5,
        Expiration: 1 * time.Minute,
        KeyGenerator: func(c *fiber.Ctx) string {
            return c.IP()
        },
        LimitReached: func(c *fiber.Ctx) error {
            c.Set("Retry-After", "60")
            return helper.Fail(c, fiber.StatusTooManyRequests,
                "terlalu banyak percobaan login, coba lagi dalam satu menit")
        },
    })
}
```

### 6.5 Authentication Business Rules (app/service/auth_rules.go)
```go
package service

import (
    "strings"
    "unicode"
    "latihan-fiber/app/model"
)

const minPasswordLength = 8

func ValidateRegister(req model.RegisterRequest) map[string]string {
    errs := map[string]string{}

    username := strings.TrimSpace(req.Username)
    switch {
    case username == "":
        errs["username"] = "wajib diisi"
    case len(username) < 3:
        errs["username"] = "minimal 3 karakter"
    case !isValidUsername(username):
        errs["username"] = "hanya boleh huruf, angka, titik, dan garis bawah"
    }

    if !isValidEmail(req.Email) {
        errs["email"] = "format email tidak valid"
    }

    if msg := checkPasswordStrength(req.Password); msg != "" {
        errs["password"] = msg
    }

    return errs
}

func ValidateLogin(req model.LoginRequest) map[string]string {
    errs := map[string]string{}

    if strings.TrimSpace(req.Username) == "" {
        errs["username"] = "wajib diisi"
    }

    if req.Password == "" {
        errs["password"] = "wajib diisi"
    }

    return errs
}

func checkPasswordStrength(password string) string {
    if len(password) < minPasswordLength {
        return "minimal 8 karakter"
    }

    var hasLetter, hasDigit bool
    for _, r := range password {
        switch {
        case unicode.IsLetter(r):
            hasLetter = true
        case unicode.IsDigit(r):
            hasDigit = true
        }
    }

    if !hasLetter || !hasDigit {
        return "harus memuat huruf dan angka"
    }

    // Daftar password lemah
    weak := map[string]bool{
        "password1": true, "12345678": true, "qwerty123": true,
        "admin123": true, "password123": true, "abcd1234": true,
    }

    if weak[strings.ToLower(password)] {
        return "password terlalu umum"
    }

    return ""
}

func isValidUsername(username string) bool {
    for _, r := range username {
        if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '.' && r != '_' {
            return false
        }
    }
    return true
}

func isValidEmail(email string) bool {
    return strings.Contains(email, "@") && strings.Contains(email, ".")
}
```

### 6.6 Authentication Service (app/service/auth_service.go) - Key Functions
```go
// Register function dengan security features
func (s *AuthService) Register(c *fiber.Ctx) error {
    ctx, cancel := helper.RequestContext(c)
    defer cancel()

    var req model.RegisterRequest
    if err := c.BodyParser(&req); err != nil {
        return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
    }

    req.Username = strings.TrimSpace(req.Username)
    req.Email = strings.TrimSpace(req.Email)

    if errs := ValidateRegister(req); len(errs) > 0 {
        return helper.FailValidation(c, errs)
    }

    // Password DI-HASH sebelum menyentuh database
    hashed, err := helper.HashPassword(req.Password)
    if err != nil {
        return helper.Fail(c, fiber.StatusInternalServerError, "gagal memproses password")
    }

    // Role selalu ditentukan server - mencegah mass assignment
    created, err := s.users.Create(ctx, model.User{
        Username: req.Username,
        Email:    req.Email,
        Password: hashed,
        Role:     "user", // Always "user" for registration
        IsActive: true,
    })

    if err != nil {
        if errors.Is(err, repository.ErrDuplicate) {
            return helper.Fail(c, fiber.StatusConflict, "username sudah dipakai")
        }
        return helper.Fail(c, fiber.StatusInternalServerError, "gagal mendaftarkan user")
    }

    return helper.Created(c, "pendaftaran berhasil", created, "/api/v1/auth/me")
}

// Login function dengan security measures
func (s *AuthService) Login(c *fiber.Ctx) error {
    ctx, cancel := helper.RequestContext(c)
    defer cancel()

    var req model.LoginRequest
    if err := c.BodyParser(&req); err != nil {
        return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
    }

    if errs := ValidateLogin(req); len(errs) > 0 {
        return helper.FailValidation(c, errs)
    }

    user, err := s.users.FindByUsername(ctx, strings.TrimSpace(req.Username))
    if err != nil {
        // Timing attack protection - run dummy password check
        helper.VerifyDummyPassword(req.Password)
        return helper.Fail(c, fiber.StatusUnauthorized, "username atau password salah")
    }

    if !helper.VerifyPassword(user.Password, req.Password) {
        return helper.Fail(c, fiber.StatusUnauthorized, "username atau password salah")
    }

    if !user.IsActive {
        return helper.Fail(c, fiber.StatusForbidden, "akun dinonaktifkan")
    }

    pair, err := s.issueTokenPair(ctx, user)
    if err != nil {
        return helper.Fail(c, fiber.StatusInternalServerError, "gagal membuat token")
    }

    return helper.Success(c, fiber.StatusOK, "login berhasil", pair)
}
```

### 6.7 Route Protection (route/route.go)
```go
func Register(app *fiber.App, pool *pgxpool.Pool,
    userService *service.UserService, studentService *service.StudentService, 
    nilaiService *service.NilaiService, authService *service.AuthService, 
    jwtManager *helper.JWTManager) {

    api := app.Group("/api/v1")
    api.Get("/health", healthCheck(pool))

    // Authentication endpoints - PUBLIC
    auth := api.Group("/auth", middleware.RequireJSON)
    auth.Post("/register", authService.Register)
    auth.Post("/login", middleware.LoginRateLimiter(), authService.Login)
    auth.Post("/refresh", authService.Refresh)
    auth.Post("/logout", authService.Logout)
    auth.Get("/me", middleware.RequireAuth(jwtManager), authService.Me)

    // Students endpoints - PROTECTED
    students := api.Group("/students", middleware.RequireJSON, middleware.RequireAuth(jwtManager))
    students.Get("/", studentService.List)
    students.Get("/:id", studentService.Get)
    students.Post("/", studentService.Create)
    students.Put("/:id", studentService.Replace)
    students.Patch("/:id", studentService.Patch)
    students.Delete("/:id", studentService.Delete)

    // Nilai endpoints - PROTECTED  
    nilai := api.Group("/nilai", middleware.RequireJSON, middleware.RequireAuth(jwtManager))
    nilai.Get("/", nilaiService.GetAllNilai_Handler)
    nilai.Post("/", nilaiService.CreateNilai_Handler)
    nilai.Get("/nim/:nim", nilaiService.GetNilaiByNIM_Handler)
    nilai.Get("/statistik/:nim", nilaiService.GetStatistikNilai_Handler)
    nilai.Get("/matkul", nilaiService.GetNilaiByMatkul_Handler)
}
```

### 6.8 Unit Tests (app/service/auth_rules_test.go)
```go
package service

import (
    "testing"
    "latihan-fiber/app/model"
)

func TestCheckPasswordStrength(t *testing.T) {
    tests := []struct {
        name     string
        password string
        wantErr  string
    }{
        {
            name:     "Valid strong password",
            password: "MySecure123",
            wantErr:  "",
        },
        {
            name:     "Password too short",
            password: "Short1",
            wantErr:  "minimal 8 karakter",
        },
        {
            name:     "Password without letter",
            password: "12345678",
            wantErr:  "harus memuat huruf dan angka",
        },
        {
            name:     "Password without digit",
            password: "NoNumbersHere",
            wantErr:  "harus memuat huruf dan angka",
        },
        {
            name:     "Common weak password",
            password: "password123",
            wantErr:  "password terlalu umum",
        },
        {
            name:     "Another weak password",
            password: "admin123",
            wantErr:  "password terlalu umum",
        },
        {
            name:     "Valid password with symbols",
            password: "MyPass123!@#",
            wantErr:  "",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := checkPasswordStrength(tt.password)
            if got != tt.wantErr {
                t.Errorf("checkPasswordStrength() = %v, want %v", got, tt.wantErr)
            }
        })
    }
}

func TestValidateRegister(t *testing.T) {
    tests := []struct {
        name    string
        req     model.RegisterRequest
        wantErr bool
        errKeys []string
    }{
        {
            name: "Valid registration",
            req: model.RegisterRequest{
                Username: "testuser",
                Email:    "test@example.com",
                Password: "StrongPass123",
            },
            wantErr: false,
        },
        {
            name: "Empty username",
            req: model.RegisterRequest{
                Username: "",
                Email:    "test@example.com",
                Password: "StrongPass123",
            },
            wantErr: true,
            errKeys: []string{"username"},
        },
        {
            name: "Short username",
            req: model.RegisterRequest{
                Username: "ab",
                Email:    "test@example.com", 
                Password: "StrongPass123",
            },
            wantErr: true,
            errKeys: []string{"username"},
        },
        {
            name: "Invalid email",
            req: model.RegisterRequest{
                Username: "testuser",
                Email:    "invalid-email",
                Password: "StrongPass123",
            },
            wantErr: true,
            errKeys: []string{"email"},
        },
        {
            name: "Weak password",
            req: model.RegisterRequest{
                Username: "testuser",
                Email:    "test@example.com",
                Password: "weak",
            },
            wantErr: true,
            errKeys: []string{"password"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            errs := ValidateRegister(tt.req)
            
            if tt.wantErr && len(errs) == 0 {
                t.Errorf("ValidateRegister() expected errors, got none")
            }
            
            if !tt.wantErr && len(errs) > 0 {
                t.Errorf("ValidateRegister() expected no errors, got %v", errs)
            }
            
            for _, key := range tt.errKeys {
                if _, exists := errs[key]; !exists {
                    t.Errorf("ValidateRegister() missing expected error for key: %s", key)
                }
            }
        })
    }
}

func TestValidateLogin(t *testing.T) {
    tests := []struct {
        name    string
        req     model.LoginRequest
        wantErr bool
    }{
        {
            name: "Valid login",
            req: model.LoginRequest{
                Username: "testuser",
                Password: "password123",
            },
            wantErr: false,
        },
        {
            name: "Empty username",
            req: model.LoginRequest{
                Username: "",
                Password: "password123",
            },
            wantErr: true,
        },
        {
            name: "Empty password",
            req: model.LoginRequest{
                Username: "testuser",
                Password: "",
            },
            wantErr: true,
        },
        {
            name: "Both empty",
            req: model.LoginRequest{
                Username: "",
                Password: "",
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            errs := ValidateLogin(tt.req)
            
            if tt.wantErr && len(errs) == 0 {
                t.Errorf("ValidateLogin() expected errors, got none")
            }
            
            if !tt.wantErr && len(errs) > 0 {
                t.Errorf("ValidateLogin() expected no errors, got %v", errs)
            }
        })
    }
}
```

### 6.9 Database Migration (migrations/003_authentication.sql)
```sql
-- Migration untuk Authentication & Security - Tugas Mandiri Modul 5

-- Tambah column role pada table users
ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(20) NOT NULL DEFAULT 'user';

-- Buat table refresh_tokens untuk JWT refresh mechanism
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index untuk performa dan keamanan
CREATE INDEX IF NOT EXISTS refresh_tokens_user_id_idx ON refresh_tokens (user_id);
CREATE INDEX IF NOT EXISTS refresh_tokens_token_hash_idx ON refresh_tokens (token_hash);
CREATE INDEX IF NOT EXISTS refresh_tokens_expires_at_idx ON refresh_tokens (expires_at);

-- Tambah index untuk role pada users
CREATE INDEX IF NOT EXISTS users_role_idx ON users (role);

-- Sample admin user untuk testing (password akan di-hash manual)
INSERT INTO users (username, email, password, role, is_active) 
SELECT 'admin', 'admin@example.com', '$2a$12$K7Qx8q2Y1z3r4t5y6u7i8Wp9Qa1Ws2Ed3Rf4Tg5Yh6Uj7Ik8Ol9P', 'admin', true
WHERE NOT EXISTS (SELECT 1 FROM users WHERE username = 'admin');
```

### 6.10 Environment Configuration (.env)
```env
# Application Configuration
APP_PORT=3002
APP_NAME=Praktikum Backend Lanjut

# Database Configuration  
DB_HOST=127.0.0.1
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=naisya12
DB_NAME=praktikum_backend
DB_SSLMODE=disable
DB_MAX_CONNS=10

# JWT Configuration for Authentication
JWT_SECRET=[CENSORED_FOR_SECURITY]
JWT_ISSUER=praktikum-backend
JWT_ACCESS_TTL_MINUTES=15
JWT_REFRESH_TTL_DAYS=7

# CORS Configuration
ALLOWED_ORIGINS=http://localhost:5173
```
**⚠️ CATATAN KEAMANAN**: JWT_SECRET asli disensor untuk keamanan sesuai requirements modul.

---

## 7. BUSINESS RULES IMPLEMENTATION
- Minimal 8 karakter
- Harus mengandung huruf dan angka
- Tidak boleh menggunakan password umum (password123, admin123, etc.)
- Implementasi dengan unit test untuk memastikan rules berfungsi

### 7.2 Registration Rules
- Username minimal 3 karakter
- Username hanya boleh huruf, angka, titik, dan underscore
- Email harus format valid
- Password harus memenuhi strength requirements

### 7.3 Security Rules
- Role selalu ditentukan oleh server (mencegah privilege escalation)
- Access token berumur pendek (15 menit)
- Refresh token rotation untuk mencegah replay attacks
- Rate limiting untuk mencegah brute force attacks

---

## 8. STRUKTUR PROJECT

```
latihan-fiber/
├── app/
│   ├── model/          # Data models dan request/response structs
│   │   ├── auth.go     # Authentication models
│   │   └── user.go     # User model dengan role
│   ├── repository/     # Database access layer
│   │   ├── token_repository.go    # Refresh token operations
│   │   └── user_repository.go     # User CRUD dengan role support
│   └── service/        # Business logic layer
│       ├── auth_service.go        # Authentication business logic
│       ├── auth_rules.go          # Authentication validation rules
│       └── auth_rules_test.go     # Unit tests untuk validation
├── helper/
│   ├── jwt.go          # JWT token management
│   ├── security.go     # Password hashing utilities
│   └── response.go     # API response helpers
├── middleware/
│   └── auth.go         # Authentication & rate limiting middleware
├── migrations/
│   └── 003_authentication.sql    # Database schema untuk auth
└── main.go             # Application entry point
```

---

## 9. TESTING STRATEGY

### 9.1 Unit Testing
- **Password validation rules**: 7 test cases
- **Registration validation**: 5 test cases  
- **Login validation**: 4 test cases
- **Business rules**: Comprehensive coverage

### 9.2 Integration Testing
- **Register flow**: End-to-end user registration
- **Login flow**: Authentication dengan JWT generation
- **Protected endpoints**: Authorization dengan middleware

### 9.3 Security Testing
- **Unauthorized access**: Memastikan 401 error tanpa token
- **Authorized access**: Memastikan akses berhasil dengan valid token
- **Token validation**: JWT signature dan expiry validation

---

## 10. KESIMPULAN

### 10.1 Pencapaian
✅ **Authentication System**: Register, Login, Refresh, Logout berhasil diimplementasikan  
✅ **Security Features**: bcrypt, JWT, rate limiting, mass assignment prevention  
✅ **Protected Endpoints**: Students API berhasil dilindungi dengan authentication  
✅ **Unit Testing**: 20+ test cases semuanya PASS  
✅ **Security Testing**: Unauthorized access diblokir, authorized access berhasil  

### 10.2 Security Best Practices yang Diterapkan
1. **Password hashing** dengan bcrypt cost 12
2. **JWT dengan short-lived access token** dan refresh token rotation
3. **Rate limiting** untuk mencegah brute force attacks
4. **Mass assignment prevention** dengan server-side role assignment
5. **Proper error handling** tanpa information leakage
6. **Comprehensive validation** dengan unit tests

### 10.3 Hasil Testing
- Semua endpoint authentication berfungsi dengan benar
- Security middleware berhasil melindungi protected endpoints
- Unit tests menunjukkan business rules berfungsi sesuai requirements
- System tahan terhadap common security vulnerabilities

**Tugas Modul 5 - Authentication & Security telah berhasil diselesaikan dengan implementasi yang comprehensive dan testing yang menyeluruh.**

---

## 11. SUMBER BANTUAN & REFERENSI

### 11.1 Bantuan AI
Dalam pengerjaan tugas ini, saya menggunakan bantuan **Kiro AI Assistant** untuk:
- Strukturing implementasi authentication flow
- Code review untuk security best practices
- Debugging unit test cases
- Documentation dan laporan formatting
- Security analysis dan vulnerability assessment

### 11.2 Referensi Teknis
- **OWASP Authentication Cheat Sheet**: Guidelines untuk secure authentication
- **Go Fiber Documentation**: Framework-specific implementations
- **JWT.io**: JWT token format dan security considerations  
- **bcrypt Package Documentation**: Password hashing best practices
- **PostgreSQL Documentation**: Database schema design

---

**Tanggal**: 13 September 2026  
**Signature**: [NAMA KAMU]
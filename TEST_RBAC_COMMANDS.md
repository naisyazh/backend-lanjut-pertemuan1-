# Testing Commands untuk Module 6 RBAC

## 🎯 Setup Test Data

### 1. Buat Test Users dengan Role Berbeda

```bash
# 1. Register Admin User
curl -X POST http://127.0.0.1:3002/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin_test",
    "email": "admin@test.com", 
    "password": "password123"
  }'

# 2. Update role admin_test menjadi admin (manual di database)
psql -U postgres -d praktikum_backend -c "UPDATE users SET role = 'admin' WHERE username = 'admin_test';"

# 3. Register Staff User  
curl -X POST http://127.0.0.1:3002/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "staff_test",
    "email": "staff@test.com",
    "password": "password123"
  }'

# 4. Update role staff_test menjadi staff
psql -U postgres -d praktikum_backend -c "UPDATE users SET role = 'staff' WHERE username = 'staff_test';"

# 5. Register User Biasa
curl -X POST http://127.0.0.1:3002/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "user_test",
    "email": "user@test.com",
    "password": "password123"
  }'
```

### 2. Login dan Ambil Token

```bash
# Login sebagai Admin
curl -X POST http://127.0.0.1:3002/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin_test",
    "password": "password123"
  }'
# Simpan access_token dari response untuk ADMIN_TOKEN

# Login sebagai Staff  
curl -X POST http://127.0.0.1:3002/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "staff_test", 
    "password": "password123"
  }'
# Simpan access_token dari response untuk STAFF_TOKEN

# Login sebagai User
curl -X POST http://127.0.0.1:3002/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "user_test",
    "password": "password123"
  }'
# Simpan access_token dari response untuk USER_TOKEN
```

## 🧪 Test Cases sesuai Permission Matrix

### ADMIN TESTS (Semua akses ✅)

```bash
# Set token admin
ADMIN_TOKEN="your_admin_access_token_here"

# 1. GET /students - Should SUCCESS (200)
curl -X GET http://127.0.0.1:3002/api/v1/students \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json"

# 2. GET /students/:id - Should SUCCESS (200)
curl -X GET http://127.0.0.1:3002/api/v1/students/1 \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json"

# 3. POST /students - Should SUCCESS (201)
curl -X POST http://127.0.0.1:3002/api/v1/students \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "nim": "ADMIN001",
    "name": "Student by Admin",
    "grade": "A"
  }'

# 4. PUT /students/:id - Should SUCCESS (200)
curl -X PUT http://127.0.0.1:3002/api/v1/students/1 \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "nim": "UPDATED001", 
    "name": "Updated by Admin",
    "grade": "A+",
    "is_active": true
  }'

# 5. PATCH /students/:id - Should SUCCESS (200)
curl -X PATCH http://127.0.0.1:3002/api/v1/students/1 \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Partially Updated by Admin"
  }'

# 6. DELETE /students/:id - Should SUCCESS (204)
curl -X DELETE http://127.0.0.1:3002/api/v1/students/1 \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

### STAFF TESTS (Read + Create, No Update/Delete)

```bash
# Set token staff
STAFF_TOKEN="your_staff_access_token_here"

# 1. GET /students - Should SUCCESS (200) ✅
curl -X GET http://127.0.0.1:3002/api/v1/students \
  -H "Authorization: Bearer $STAFF_TOKEN" \
  -H "Content-Type: application/json"

# 2. GET /students/:id - Should SUCCESS (200) ✅
curl -X GET http://127.0.0.1:3002/api/v1/students/2 \
  -H "Authorization: Bearer $STAFF_TOKEN" \
  -H "Content-Type: application/json"

# 3. POST /students - Should SUCCESS (201) ✅
curl -X POST http://127.0.0.1:3002/api/v1/students \
  -H "Authorization: Bearer $STAFF_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "nim": "STAFF001",
    "name": "Student by Staff", 
    "grade": "B"
  }'

# 4. PUT /students/:id - Should FAIL (403) ❌
curl -X PUT http://127.0.0.1:3002/api/v1/students/2 \
  -H "Authorization: Bearer $STAFF_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "nim": "FAIL001",
    "name": "Should Fail",
    "grade": "F",
    "is_active": true
  }'

# 5. PATCH /students/:id - Should FAIL (403) ❌
curl -X PATCH http://127.0.0.1:3002/api/v1/students/2 \
  -H "Authorization: Bearer $STAFF_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Should Fail Patch"
  }'

# 6. DELETE /students/:id - Should FAIL (403) ❌
curl -X DELETE http://127.0.0.1:3002/api/v1/students/2 \
  -H "Authorization: Bearer $STAFF_TOKEN"
```

### USER TESTS (Ownership-based Access)

```bash
# Set token user
USER_TOKEN="your_user_access_token_here"

# 1. GET /students - Should FAIL (403) ❌
curl -X GET http://127.0.0.1:3002/api/v1/students \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json"

# 2. POST /students - Should SUCCESS (201) ✅
curl -X POST http://127.0.0.1:3002/api/v1/students \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "nim": "USER001",
    "name": "Student by User",
    "grade": "B+"
  }'
# Simpan ID student yang dibuat untuk test ownership

# 3. GET /students/:id (milik sendiri) - Should SUCCESS (200) ✅
curl -X GET http://127.0.0.1:3002/api/v1/students/[ID_STUDENT_MILIK_USER] \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json"

# 4. GET /students/:id (milik orang lain) - Should FAIL (403) ❌
curl -X GET http://127.0.0.1:3002/api/v1/students/[ID_STUDENT_MILIK_ADMIN] \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json"

# 5. PUT /students/:id (milik sendiri) - Should SUCCESS (200) ✅
curl -X PUT http://127.0.0.1:3002/api/v1/students/[ID_STUDENT_MILIK_USER] \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "nim": "USER001_UPDATED",
    "name": "Updated by Owner",
    "grade": "A",
    "is_active": true
  }'

# 6. PUT /students/:id (milik orang lain) - Should FAIL (403) ❌
curl -X PUT http://127.0.0.1:3002/api/v1/students/[ID_STUDENT_MILIK_ADMIN] \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "nim": "SHOULD_FAIL",
    "name": "Should Fail Update",
    "grade": "F",
    "is_active": false
  }'

# 7. PATCH /students/:id (milik sendiri) - Should SUCCESS (200) ✅
curl -X PATCH http://127.0.0.1:3002/api/v1/students/[ID_STUDENT_MILIK_USER] \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Partially Updated by Owner"
  }'

# 8. PATCH /students/:id (milik orang lain) - Should FAIL (403) ❌
curl -X PATCH http://127.0.0.1:3002/api/v1/students/[ID_STUDENT_MILIK_ADMIN] \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Should Fail Patch"
  }'

# 9. DELETE /students/:id - Should FAIL (403) ❌ (user tidak boleh delete apapun)
curl -X DELETE http://127.0.0.1:3002/api/v1/students/[ID_STUDENT_MILIK_USER] \
  -H "Authorization: Bearer $USER_TOKEN"
```

## 🔍 Expected Results Summary

### ✅ Success Cases (HTTP 200/201/204):
- **Admin**: Semua operasi berhasil
- **Staff**: GET, POST berhasil  
- **User**: POST, GET/PUT/PATCH data milik sendiri berhasil

### ❌ Forbidden Cases (HTTP 403):
- **Staff**: PUT, PATCH, DELETE gagal
- **User**: GET list, GET/PUT/PATCH data orang lain, DELETE gagal

### 🔒 Security Verification:
- User tidak bisa akses data dengan owner_id ≠ user_id mereka
- Staff tidak bisa update/delete
- Auto-assignment owner_id saat POST oleh user biasa
- Token validation di semua endpoint

## 📊 Test Matrix Verification

| Role    | GET /students | GET /students/:id | POST /students | PUT/PATCH /students/:id | DELETE /students/:id |
|---------|---------------|-------------------|----------------|------------------------|---------------------|
| admin   | ✅ 200         | ✅ 200            | ✅ 201         | ✅ 200                  | ✅ 204               |
| staff   | ✅ 200         | ✅ 200            | ✅ 201         | ❌ 403                  | ❌ 403               |
| user    | ❌ 403         | ✅ 200 / ❌ 403*   | ✅ 201         | ✅ 200 / ❌ 403*        | ❌ 403               |

*) ✅ jika owner_id = user_id, ❌ jika bukan
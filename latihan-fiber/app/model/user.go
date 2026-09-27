package model

import "time"

// User adalah entitas pengguna di sistem.
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // NEVER expose password in JSON response
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateUserRequest adalah request body untuk POST /users
// Mulai pertemuan ini, aturan validasi ditulis sebagai tag pada struct.
type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=30,alphanum"`
	Email    string `json:"email"    validate:"required,email,max=120"`
	Password string `json:"password" validate:"required,min=8,max=72,nospace"`
}

// ReplaceUserRequest adalah request body untuk PUT /users/:id
type ReplaceUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=30,alphanum"`
	Email    string `json:"email"    validate:"required,email,max=120"`
	IsActive bool   `json:"is_active"`
}

// PatchUserRequest adalah request body untuk PATCH /users/:id
// Pada PATCH, pointer membedakan "tidak dikirim" (nil) dari "dikirim
// bernilai kosong". omitnil dipilih karena ia menyatakan maksud yang
// sebenarnya: lewati hanya bila nil.
type PatchUserRequest struct {
	Username *string `json:"username,omitempty" validate:"omitnil,min=3,max=30,alphanum"`
	Email    *string `json:"email,omitempty"    validate:"omitnil,email,max=120"`
	IsActive *bool   `json:"is_active,omitempty"`
}

// ListQuery adalah parameter query untuk endpoint daftar (offset-based).
type ListQuery struct {
	Page     int
	Limit    int
	Search   string
	Sort     string
	Order    string
	IsActive *bool
}

// Offset menghitung berapa baris yang dilewati untuk halaman ini.
func (q ListQuery) Offset() int {
	return (q.Page - 1) * q.Limit
}

// Cursor adalah penanda posisi untuk keyset pagination.
type Cursor struct {
	CreatedAt time.Time
	ID        int
}

// CursorQuery adalah parameter untuk endpoint daftar berbasis cursor.
type CursorQuery struct {
	Limit    int
	Search   string
	IsActive *bool
	After    *Cursor
}

// WebResponse adalah struktur standar untuk semua respons JSON.
type WebResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// Meta adalah metadata untuk respons daftar berbasis offset.
type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// CursorMeta menggantikan Meta pada endpoint yang memakai cursor.
// Perhatikan tidak adanya Total dan TotalPages — keduanya tidak dapat
// disediakan tanpa COUNT(*) atas seluruh tabel, persis biaya yang ingin
// dihindari oleh pagination berbasis cursor.
type CursorMeta struct {
	Limit      int    `json:"limit"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

// ErrorResponse adalah struktur standar untuk respons error.
type ErrorResponse struct {
	Success   bool              `json:"success"`
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Fields    map[string]string `json:"fields,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
}

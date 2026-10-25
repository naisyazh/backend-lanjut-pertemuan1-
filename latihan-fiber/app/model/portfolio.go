package model

import "time"

// Portfolio adalah dokumen portofolio mahasiswa di MongoDB.
// Setiap mahasiswa memiliki tepat SATU dokumen portofolio.
// Data mahasiswanya sendiri tetap di PostgreSQL (student_id sebagai referensi).
type Portfolio struct {
	ID           string        `json:"id"`
	StudentID    int           `json:"student_id"`
	Headline     string        `json:"headline"`
	Bio          string        `json:"bio"`
	Projects     []Project     `json:"projects"`
	Certificates []Certificate `json:"certificates"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

// Project disarangkan (embedded) di dalam dokumen Portfolio karena:
// 1. Selalu dibaca bersama portofolio induknya
// 2. Jumlahnya terbatas (maks 20)
// 3. Tidak dipakai bersama dokumen lain
type Project struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Technologies []string  `json:"technologies"`
	Link         string    `json:"link"`
	Year         int       `json:"year"`
	CreatedAt    time.Time `json:"created_at"`
}

// Certificate juga disarangkan karena pola penggunaan yang sama dengan Project.
type Certificate struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Issuer    string    `json:"issuer"`
	IssuedAt  time.Time `json:"issued_at"`
	CreatedAt time.Time `json:"created_at"`
}

// UpsertPortfolioRequest dipakai untuk PUT /students/:id/portfolio.
// PUT mengganti seluruh portofolio, bukan sebagian.
type UpsertPortfolioRequest struct {
	Headline     string   `json:"headline"     validate:"required,min=3,max=200"`
	Bio          string   `json:"bio"          validate:"required,min=10,max=2000"`
}

// AddProjectRequest dipakai untuk POST /students/:id/portfolio/projects.
type AddProjectRequest struct {
	Title        string   `json:"title"        validate:"required,min=2,max=200"`
	Technologies []string `json:"technologies" validate:"required,min=1,max=20,dive,min=1,max=50"`
	Link         string   `json:"link"         validate:"required,url"`
	Year         int      `json:"year"         validate:"required,min=2000,max=2030"`
}

// TechCount dipakai untuk endpoint statistik teknologi.
type TechCount struct {
	Technology string `json:"technology"`
	Total      int    `json:"total"`
}

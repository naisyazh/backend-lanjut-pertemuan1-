package model

import "time"

type Nilai struct {
	ID              int       `json:"id" db:"id"`
	NamaMataKuliah  string    `json:"nama_mata_kuliah" db:"nama_mata_kuliah"`
	Nilai           float64   `json:"nilai" db:"nilai"`
	StudentID       int       `json:"student_id" db:"student_id"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

type NilaiWithStudent struct {
	ID              int       `json:"id"`
	NamaMataKuliah  string    `json:"nama_mata_kuliah"`
	Nilai           float64   `json:"nilai"`
	StudentID       int       `json:"student_id"`
	StudentNIM      string    `json:"student_nim"`
	StudentName     string    `json:"student_name"`
	CreatedAt       time.Time `json:"created_at"`
}

type CreateNilaiRequest struct {
	NamaMataKuliah string  `json:"nama_mata_kuliah" validate:"required"`
	Nilai          float64 `json:"nilai" validate:"required,min=0,max=4"`
	StudentNIM     string  `json:"student_nim" validate:"required"`
}
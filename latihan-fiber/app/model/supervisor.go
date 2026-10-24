package model

import "time"

// Supervisor adalah dosen pembimbing tugas akhir.
type Supervisor struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	Quota           int       `json:"quota"`
	SupervisedCount int       `json:"supervised_count"`
	Version         int       `json:"version"`
	CreatedAt       time.Time `json:"created_at"`
}

// Supervision adalah relasi antara mahasiswa dan pembimbingnya.
type Supervision struct {
	ID           int64     `json:"id"`
	StudentID    int       `json:"student_id"`
	SupervisorID int       `json:"supervisor_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// AssignRequest adalah body untuk POST /supervisors/:id/assign.
type AssignRequest struct {
	StudentID int `json:"student_id" validate:"required,min=1"`
}

// SwitchRequest adalah body untuk POST /supervisions/switch.
type SwitchRequest struct {
	StudentID       int `json:"student_id"        validate:"required,min=1"`
	NewSupervisorID int `json:"new_supervisor_id" validate:"required,min=1"`
}

// UpdateSupervisorQuotaRequest adalah body untuk PATCH /supervisors/:id.
// Memakai optimistic locking — version wajib dikirim agar server tahu
// versi mana yang sedang diedit client.
type UpdateSupervisorQuotaRequest struct {
	Quota   int `json:"quota"   validate:"required,min=1,max=100"`
	Version int `json:"version" validate:"required,min=1"`
}

package repository

import (
	"context"
	"fmt"
	"latihan-fiber/app/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type NilaiRepository struct {
	pool *pgxpool.Pool
}

func NewNilaiRepository(pool *pgxpool.Pool) *NilaiRepository {
	return &NilaiRepository{pool: pool}
}

// GetNilaiByNIM - Ambil semua nilai berdasarkan NIM mahasiswa
func (r *NilaiRepository) GetNilaiByNIM(nim string) ([]model.NilaiWithStudent, error) {
	query := `
		SELECT 
			n.id, 
			n.nama_mata_kuliah, 
			n.nilai, 
			n.student_id,
			s.nim as student_nim,
			s.name as student_name,
			n.created_at
		FROM nilai n
		JOIN students s ON n.student_id = s.id
		WHERE LOWER(s.nim) = LOWER($1)
		ORDER BY n.created_at DESC
	`

	rows, err := r.pool.Query(context.Background(), query, nim)
	if err != nil {
		return nil, fmt.Errorf("error querying nilai: %v", err)
	}
	defer rows.Close()

	var nilaiList []model.NilaiWithStudent
	for rows.Next() {
		var n model.NilaiWithStudent
		err := rows.Scan(
			&n.ID,
			&n.NamaMataKuliah,
			&n.Nilai,
			&n.StudentID,
			&n.StudentNIM,
			&n.StudentName,
			&n.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning nilai: %v", err)
		}
		nilaiList = append(nilaiList, n)
	}

	return nilaiList, nil
}

// CreateNilai - Tambah nilai baru
func (r *NilaiRepository) CreateNilai(req model.CreateNilaiRequest) (*model.Nilai, error) {
	// Cari student berdasarkan NIM
	var studentID int
	err := r.pool.QueryRow(context.Background(), "SELECT id FROM students WHERE LOWER(nim) = LOWER($1)", req.StudentNIM).Scan(&studentID)
	if err != nil {
		return nil, fmt.Errorf("student with NIM %s not found", req.StudentNIM)
	}

	// Insert nilai baru
	query := `
		INSERT INTO nilai (nama_mata_kuliah, nilai, student_id) 
		VALUES ($1, $2, $3) 
		RETURNING id, nama_mata_kuliah, nilai, student_id, created_at
	`

	var nilai model.Nilai
	err = r.pool.QueryRow(context.Background(), query, req.NamaMataKuliah, req.Nilai, studentID).Scan(
		&nilai.ID,
		&nilai.NamaMataKuliah,
		&nilai.Nilai,
		&nilai.StudentID,
		&nilai.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating nilai: %v", err)
	}

	return &nilai, nil
}

// GetAllNilai - Ambil semua nilai dengan informasi mahasiswa
func (r *NilaiRepository) GetAllNilai() ([]model.NilaiWithStudent, error) {
	query := `
		SELECT 
			n.id, 
			n.nama_mata_kuliah, 
			n.nilai, 
			n.student_id,
			s.nim as student_nim,
			s.name as student_name,
			n.created_at
		FROM nilai n
		JOIN students s ON n.student_id = s.id
		ORDER BY n.created_at DESC
	`

	rows, err := r.pool.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("error querying all nilai: %v", err)
	}
	defer rows.Close()

	var nilaiList []model.NilaiWithStudent
	for rows.Next() {
		var n model.NilaiWithStudent
		err := rows.Scan(
			&n.ID,
			&n.NamaMataKuliah,
			&n.Nilai,
			&n.StudentID,
			&n.StudentNIM,
			&n.StudentName,
			&n.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning nilai: %v", err)
		}
		nilaiList = append(nilaiList, n)
	}

	return nilaiList, nil
}
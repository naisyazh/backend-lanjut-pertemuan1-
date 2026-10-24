package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"latihan-fiber/app/model"
)

// Error khusus fitur pembimbing tugas akhir.
var (
	ErrSupervisorFull       = errors.New("kuota pembimbing sudah penuh")
	ErrAlreadyHasSupervisor = errors.New("mahasiswa sudah memiliki pembimbing")
	ErrNoSupervisor         = errors.New("mahasiswa belum memiliki pembimbing")
	ErrSupervisorVersion    = errors.New("data pembimbing sudah diubah pihak lain")
	ErrQuotaBelowSupervised = errors.New("kuota lebih kecil dari jumlah mahasiswa saat ini")
)

// SupervisorRepository mendefinisikan seluruh operasi database
// yang dibutuhkan fitur pembimbing tugas akhir.
type SupervisorRepository interface {
	// WithTx mengembalikan salinan repository yang berjalan di dalam tx.
	WithTx(tx DBTX) SupervisorRepository

	FindAll(ctx context.Context) ([]model.Supervisor, error)
	FindByID(ctx context.Context, id int) (model.Supervisor, error)

	// LockStudent mengunci baris students milik mahasiswa sampai
	// transaction selesai — mencegah satu mahasiswa ditetapkan ke
	// dua pembimbing secara serentak.
	LockStudent(ctx context.Context, studentID int) error

	// LockSupervisors mengunci beberapa baris supervisors SELALU
	// dalam urutan id menaik — mencegah deadlock saat switch pembimbing.
	LockSupervisors(ctx context.Context, ids ...int) error

	// IncrementSupervised mengambil satu slot pembimbing secara ATOMIK.
	// Pemeriksaan kuota dan penambahan terjadi dalam satu UPDATE.
	IncrementSupervised(ctx context.Context, supervisorID int) (model.Supervisor, error)

	// DecrementSupervised melepas satu slot pembimbing secara ATOMIK.
	DecrementSupervised(ctx context.Context, supervisorID int) (model.Supervisor, error)

	// InsertSupervision mencatat relasi mahasiswa–pembimbing.
	// UNIQUE (student_id) di database menjamin tidak ada duplikat.
	InsertSupervision(ctx context.Context, studentID, supervisorID int) error

	// DeleteSupervision menghapus relasi pembimbingan.
	// Mengembalikan ErrNoSupervisor bila relasi memang tidak ada.
	DeleteSupervision(ctx context.Context, studentID int) error

	// UpdateQuota menerapkan optimistic locking pada perubahan kuota.
	UpdateQuota(ctx context.Context, id, quota, version int) (model.Supervisor, error)
}

const supervisorColumns = `id, name, quota, supervised_count, version, created_at`

type supervisorPostgresRepository struct{ db DBTX }

// NewSupervisorRepository menerima DBTX — sama seperti CourseRepository,
// sehingga dapat dipakai di luar maupun di dalam transaction.
func NewSupervisorRepository(db DBTX) SupervisorRepository {
	return &supervisorPostgresRepository{db: db}
}

func (r *supervisorPostgresRepository) WithTx(tx DBTX) SupervisorRepository {
	return &supervisorPostgresRepository{db: tx}
}

func scanSupervisor(row pgx.Row) (model.Supervisor, error) {
	var s model.Supervisor
	err := row.Scan(&s.ID, &s.Name, &s.Quota, &s.SupervisedCount, &s.Version, &s.CreatedAt)
	return s, err
}

func (r *supervisorPostgresRepository) collect(rows pgx.Rows, err error) ([]model.Supervisor, error) {
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar pembimbing: %w", err)
	}
	defer rows.Close()
	result := []model.Supervisor{}
	for rows.Next() {
		s, err := scanSupervisor(rows)
		if err != nil {
			return nil, fmt.Errorf("membaca row pembimbing: %w", err)
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

func (r *supervisorPostgresRepository) FindAll(ctx context.Context) ([]model.Supervisor, error) {
	return r.collect(r.db.Query(ctx,
		"SELECT "+supervisorColumns+" FROM supervisors ORDER BY name"))
}

func (r *supervisorPostgresRepository) FindByID(ctx context.Context, id int) (model.Supervisor, error) {
	s, err := scanSupervisor(r.db.QueryRow(ctx,
		"SELECT "+supervisorColumns+" FROM supervisors WHERE id = $1", id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Supervisor{}, ErrNotFound
		}
		return model.Supervisor{}, fmt.Errorf("mengambil pembimbing: %w", err)
	}
	return s, nil
}

// LockStudent mengunci baris mahasiswa sampai transaction selesai.
//
// Row students dipakai sebagai "gembok penetapan": selama satu transaction
// memegang kunci ini, transaction lain untuk mahasiswa YANG SAMA harus antre.
// Tanpa FOR UPDATE, dua request serentak bisa sama-sama lolos pemeriksaan
// "mahasiswa belum punya pembimbing" dan menghasilkan dua pembimbing.
func (r *supervisorPostgresRepository) LockStudent(ctx context.Context, studentID int) error {
	var id int
	err := r.db.QueryRow(ctx,
		"SELECT id FROM students WHERE id = $1 FOR UPDATE", studentID).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("mengunci data mahasiswa: %w", err)
	}
	return nil
}

// LockSupervisors mengunci beberapa pembimbing sekaligus, SELALU dalam
// urutan id menaik — mencegah deadlock saat switch pembimbing.
//
// Tanpa urutan seragam: mahasiswa A switch dari sup1→sup2 dan mahasiswa B
// switch dari sup2→sup1 pada saat yang sama akan saling menunggu kunci
// milik satu sama lain (deadlock sesuai A.6).
func (r *supervisorPostgresRepository) LockSupervisors(ctx context.Context, ids ...int) error {
	rows, err := r.db.Query(ctx,
		"SELECT id FROM supervisors WHERE id = ANY($1) ORDER BY id FOR UPDATE", ids)
	if err != nil {
		return fmt.Errorf("mengunci pembimbing: %w", err)
	}
	defer rows.Close()
	found := 0
	for rows.Next() {
		found++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("mengunci pembimbing: %w", err)
	}
	if found != len(ids) {
		return ErrNotFound
	}
	return nil
}

// IncrementSupervised mengambil satu slot pembimbing secara ATOMIK.
//
// Pemeriksaan kuota dan penambahan supervised_count terjadi dalam satu
// pernyataan UPDATE. PostgreSQL mengunci baris tersebut selama UPDATE
// berjalan sehingga dua request yang datang bersamaan tidak mungkin
// sama-sama melihat "masih ada slot kosong".
func (r *supervisorPostgresRepository) IncrementSupervised(ctx context.Context, supervisorID int) (model.Supervisor, error) {
	s, err := scanSupervisor(r.db.QueryRow(ctx,
		`UPDATE supervisors SET supervised_count = supervised_count + 1
		 WHERE id = $1 AND supervised_count < quota
		 RETURNING `+supervisorColumns, supervisorID))
	if err == nil {
		return s, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return model.Supervisor{}, fmt.Errorf("menambah mahasiswa bimbingan: %w", err)
	}
	// Tidak ada row berubah — bedakan: tidak ada vs sudah penuh
	if _, findErr := r.FindByID(ctx, supervisorID); findErr != nil {
		return model.Supervisor{}, findErr
	}
	return model.Supervisor{}, ErrSupervisorFull
}

func (r *supervisorPostgresRepository) DecrementSupervised(ctx context.Context, supervisorID int) (model.Supervisor, error) {
	s, err := scanSupervisor(r.db.QueryRow(ctx,
		`UPDATE supervisors SET supervised_count = supervised_count - 1
		 WHERE id = $1
		 RETURNING `+supervisorColumns, supervisorID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Supervisor{}, ErrNotFound
		}
		return model.Supervisor{}, fmt.Errorf("mengurangi mahasiswa bimbingan: %w", err)
	}
	return s, nil
}

// InsertSupervision mencatat relasi mahasiswa–pembimbing.
// UNIQUE (student_id) memastikan tidak ada mahasiswa yang punya dua pembimbing.
func (r *supervisorPostgresRepository) InsertSupervision(ctx context.Context, studentID, supervisorID int) error {
	_, err := r.db.Exec(ctx,
		"INSERT INTO supervisions (student_id, supervisor_id) VALUES ($1, $2)",
		studentID, supervisorID)
	if err != nil {
		if pgErrorCode(err) == codeUniqueViolation {
			return ErrAlreadyHasSupervisor
		}
		return fmt.Errorf("mencatat pembimbingan: %w", err)
	}
	return nil
}

// DeleteSupervision menghapus relasi pembimbingan.
// DELETE yang tidak menemukan row bukan error menurut PostgreSQL — ia sukses
// dengan nol row. Bila tidak diperiksa, kode sesudahnya akan mengurangi
// supervised_count untuk relasi yang sebenarnya tidak pernah ada.
func (r *supervisorPostgresRepository) DeleteSupervision(ctx context.Context, studentID int) error {
	tag, err := r.db.Exec(ctx,
		"DELETE FROM supervisions WHERE student_id = $1", studentID)
	if err != nil {
		return fmt.Errorf("menghapus pembimbingan: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNoSupervisor
	}
	return nil
}

// UpdateQuota menerapkan optimistic locking.
//
// Tidak ada kunci yang dipegang selama admin mengisi form. UPDATE hanya
// berhasil bila version di database masih sama dengan yang dikirim client.
// Bila sudah berubah (orang lain menyimpan lebih dulu), perubahan ditolak
// dengan ErrSupervisorVersion — admin diminta muat ulang.
func (r *supervisorPostgresRepository) UpdateQuota(ctx context.Context, id, quota, version int) (model.Supervisor, error) {
	s, err := scanSupervisor(r.db.QueryRow(ctx,
		`UPDATE supervisors SET quota = $1, version = version + 1
		 WHERE id = $2 AND version = $3 AND $1 >= supervised_count
		 RETURNING `+supervisorColumns, quota, id, version))
	if err == nil {
		return s, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return model.Supervisor{}, fmt.Errorf("mengubah kuota pembimbing: %w", err)
	}
	current, findErr := r.FindByID(ctx, id)
	if findErr != nil {
		return model.Supervisor{}, findErr
	}
	if current.Version != version {
		return model.Supervisor{}, ErrSupervisorVersion
	}
	return model.Supervisor{}, ErrQuotaBelowSupervised
}

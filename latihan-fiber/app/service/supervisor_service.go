package service

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"

	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
	"latihan-fiber/helper"
)

// SupervisorService menangani seluruh logika bisnis penetapan
// dosen pembimbing tugas akhir.
type SupervisorService struct {
	supervisors repository.SupervisorRepository
	tx          *repository.TxManager
}

func NewSupervisorService(
	supervisors repository.SupervisorRepository,
	tx *repository.TxManager,
) *SupervisorService {
	return &SupervisorService{supervisors: supervisors, tx: tx}
}

// ---------- GET /supervisors ----------

func (s *SupervisorService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	list, err := s.supervisors.FindAll(ctx)
	if err != nil {
		return helper.Internal(err)
	}
	return helper.Success(c, fiber.StatusOK, "daftar pembimbing berhasil diambil", list)
}

// ---------- GET /supervisors/:id ----------

func (s *SupervisorService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	sup, err := s.supervisors.FindByID(ctx, id)
	if err != nil {
		return translateSupervisorError(err)
	}
	return helper.Success(c, fiber.StatusOK, "pembimbing ditemukan", sup)
}

// ---------- POST /supervisors/:id/assign ----------
//
// Menetapkan mahasiswa ke pembimbing. Dijaga permission supervision:manage.
//
// Urutan dalam transaction:
//  1. LockStudent (pessimistic) — mencegah satu mahasiswa ditetapkan
//     ke dua pembimbing secara serentak dari dua request berbeda.
//  2. IncrementSupervised (UPDATE bersyarat atomik) — mencegah banyak
//     request berebut slot terakhir pembimbing.
//  3. InsertSupervision — UNIQUE (student_id) sebagai jaring pengaman
//     terakhir; bila gagal, rollback membatalkan langkah 2 juga.
func (s *SupervisorService) Assign(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	supervisorID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	var req model.AssignRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	var assigned model.Supervisor
	err := s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		// Seluruh query WAJIB memakai repo (terikat ke tx), bukan s.supervisors
		// (terikat ke pool dan di luar transaction).
		repo := s.supervisors.WithTx(tx)

		// Langkah 1 — kunci mahasiswa
		if err := repo.LockStudent(ctx, req.StudentID); err != nil {
			return err
		}

		// Langkah 2 — ambil slot secara atomik
		var err error
		assigned, err = repo.IncrementSupervised(ctx, supervisorID)
		if err != nil {
			return err
		}

		// Langkah 3 — catat relasi; bila gagal, rollback membatalkan langkah 2
		return repo.InsertSupervision(ctx, req.StudentID, supervisorID)
	})
	if err != nil {
		return translateSupervisorError(err)
	}

	return helper.Created(c, "mahasiswa berhasil ditetapkan ke pembimbing", assigned,
		"/api/v1/supervisors/"+strconv.Itoa(supervisorID))
}

// ---------- POST /supervisions/switch ----------
//
// Mengganti pembimbing mahasiswa secara atomik.
//
// Bila pembimbing baru penuh, mahasiswa TETAP dengan pembimbing lama.
// Bebas deadlock karena kedua pembimbing selalu dikunci dalam urutan id menaik.
func (s *SupervisorService) Switch(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.SwitchRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	var joined model.Supervisor
	err := s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		repo := s.supervisors.WithTx(tx)

		// Langkah 1 — kunci mahasiswa
		if err := repo.LockStudent(ctx, req.StudentID); err != nil {
			return err
		}

		// Langkah 2 — cari & kunci pembimbing lama di dalam transaction
		var oldSupervisorID int
		err := tx.QueryRow(ctx,
			`SELECT supervisor_id FROM supervisions
			 WHERE student_id = $1 FOR UPDATE`,
			req.StudentID).Scan(&oldSupervisorID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return repository.ErrNoSupervisor
			}
			return err
		}

		if oldSupervisorID == req.NewSupervisorID {
			return helper.BadRequest("pembimbing baru sama dengan pembimbing lama")
		}

		// Langkah 3 — kunci KEDUA pembimbing dalam urutan id menaik
		// Ini yang mencegah deadlock: A pindah dari sup1→sup2 dan
		// B pindah dari sup2→sup1 pada saat yang sama tidak akan saling
		// menunggu karena urutan pengunciannya seragam.
		if err := repo.LockSupervisors(ctx, oldSupervisorID, req.NewSupervisorID); err != nil {
			return err
		}

		// Langkah 4 — lepas dari pembimbing lama
		if err := repo.DeleteSupervision(ctx, req.StudentID); err != nil {
			return err
		}
		if _, err := repo.DecrementSupervised(ctx, oldSupervisorID); err != nil {
			return err
		}

		// Langkah 5 — daftarkan ke pembimbing baru
		var err2 error
		joined, err2 = repo.IncrementSupervised(ctx, req.NewSupervisorID)
		if err2 != nil {
			return err2
		}
		return repo.InsertSupervision(ctx, req.StudentID, req.NewSupervisorID)
	})
	if err != nil {
		return translateSupervisorError(err)
	}

	return helper.Success(c, fiber.StatusOK, "berhasil mengganti pembimbing", joined)
}

// ---------- PATCH /supervisors/:id ----------
//
// Mengubah kuota pembimbing dengan optimistic locking.
//
// Tidak perlu WithTx karena hanya satu pernyataan UPDATE — satu
// pernyataan di PostgreSQL sudah atomik dengan sendirinya (A.3).
func (s *SupervisorService) UpdateQuota(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	var req model.UpdateSupervisorQuotaRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	updated, err := s.supervisors.UpdateQuota(ctx, id, req.Quota, req.Version)
	if err != nil {
		return translateSupervisorError(err)
	}
	return helper.Success(c, fiber.StatusOK, "kuota pembimbing berhasil diubah", updated)
}

// translateSupervisorError memetakan error repository ke response HTTP.
//
// Seluruh kegagalan aturan bisnis memakai 409 dengan kode berbeda.
// Yang tidak dikenali menjadi 500 — termasuk pelanggaran CHECK constraint
// yang hanya terpicu bila kode aplikasi kita sendiri yang salah.
func translateSupervisorError(err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("pembimbing atau mahasiswa tidak ditemukan")
	case errors.Is(err, repository.ErrSupervisorFull):
		return helper.NewError(fiber.StatusConflict, "SUPERVISOR_FULL",
			"kuota pembimbing sudah penuh")
	case errors.Is(err, repository.ErrAlreadyHasSupervisor):
		return helper.NewError(fiber.StatusConflict, "ALREADY_HAS_SUPERVISOR",
			"mahasiswa sudah memiliki pembimbing")
	case errors.Is(err, repository.ErrNoSupervisor):
		return helper.NewError(fiber.StatusConflict, "NO_SUPERVISOR",
			"mahasiswa belum memiliki pembimbing")
	case errors.Is(err, repository.ErrSupervisorVersion):
		return helper.NewError(fiber.StatusConflict, "VERSION_CONFLICT",
			"data pembimbing sudah diubah pihak lain, muat ulang lalu coba lagi")
	case errors.Is(err, repository.ErrQuotaBelowSupervised):
		return helper.NewError(fiber.StatusConflict, "QUOTA_BELOW_SUPERVISED",
			"kuota tidak boleh lebih kecil dari jumlah mahasiswa bimbingan saat ini")
	default:
		return helper.Internal(err)
	}
}

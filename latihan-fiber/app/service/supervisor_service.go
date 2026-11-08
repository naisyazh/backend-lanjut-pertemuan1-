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

// List godoc
//
// @Summary Menampilkan daftar pembimbing tugas akhir
// @Tags supervisors
// @Produce json
// @Success 200 {object} model.WebResponse{data=[]model.Supervisor}
// @Failure 401 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /supervisors [get]
func (s *SupervisorService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	list, err := s.supervisors.FindAll(ctx)
	if err != nil {
		return helper.Internal(err)
	}
	return helper.Success(c, fiber.StatusOK, "daftar pembimbing berhasil diambil", list)
}

// Get godoc
//
// @Summary Menampilkan detail pembimbing
// @Tags supervisors
// @Produce json
// @Param id path int true "Supervisor ID"
// @Success 200 {object} model.WebResponse{data=model.Supervisor}
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /supervisors/{id} [get]
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

// Assign godoc
//
// @Summary Menetapkan pembimbing untuk mahasiswa (admin)
// @Tags supervisors
// @Accept json
// @Produce json
// @Param id path int true "Supervisor ID"
// @Param body body model.AssignRequest true "Student ID"
// @Success 201 {object} model.WebResponse{data=model.Supervisor}
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 403 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 409 {object} model.ErrorResponse
// @Failure 422 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /supervisors/{id}/assign [post]
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

// Switch godoc
//
// @Summary Mengganti pembimbing tugas akhir (admin)
// @Tags supervisions
// @Accept json
// @Produce json
// @Param body body model.SwitchRequest true "Student ID dan Supervisor Baru"
// @Success 200 {object} model.WebResponse{data=model.Supervisor}
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 403 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 409 {object} model.ErrorResponse
// @Failure 422 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /supervisions/switch [post]
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

// UpdateQuota godoc
//
// @Summary Mengubah kuota pembimbing dengan optimistic locking (admin)
// @Tags supervisors
// @Accept json
// @Produce json
// @Param id path int true "Supervisor ID"
// @Param body body model.UpdateSupervisorQuotaRequest true "Kuota baru dan version"
// @Success 200 {object} model.WebResponse{data=model.Supervisor}
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 403 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 409 {object} model.ErrorResponse
// @Failure 422 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /supervisors/{id} [patch]
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

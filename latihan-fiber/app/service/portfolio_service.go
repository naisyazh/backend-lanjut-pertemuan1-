package service

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
	"latihan-fiber/helper"
)

// PortfolioService memakai DUA database:
// - PostgreSQL: memverifikasi mahasiswa benar-benar ada
// - MongoDB: menyimpan data portofolio
type PortfolioService struct {
	portfolios repository.PortfolioRepository
	students   repository.StudentRepository
}

func NewPortfolioService(
	portfolios repository.PortfolioRepository,
	students repository.StudentRepository,
) *PortfolioService {
	return &PortfolioService{portfolios: portfolios, students: students}
}

// ---------- GET /students/:id/portfolio ----------

func (s *PortfolioService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id mahasiswa harus berupa angka positif")
	}

	// Tolak bila mahasiswanya tidak ada di PostgreSQL
	if _, err := s.students.FindByID(ctx, studentID); err != nil {
		return translatePortfolioError(err)
	}

	portfolio, err := s.portfolios.FindByStudent(ctx, studentID)
	if err != nil {
		return translatePortfolioError(err)
	}
	return helper.Success(c, fiber.StatusOK, "portofolio berhasil diambil", portfolio)
}

// ---------- PUT /students/:id/portfolio ----------
// Membuat portofolio baru atau mengganti seluruhnya (upsert).

func (s *PortfolioService) Upsert(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id mahasiswa harus berupa angka positif")
	}

	var req model.UpsertPortfolioRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	// Tolak bila mahasiswanya tidak ada di PostgreSQL
	if _, err := s.students.FindByID(ctx, studentID); err != nil {
		return translatePortfolioError(err)
	}

	portfolio, err := s.portfolios.Upsert(ctx, studentID, req)
	if err != nil {
		return translatePortfolioError(err)
	}
	return helper.Success(c, fiber.StatusOK, "portofolio berhasil disimpan", portfolio)
}

// ---------- POST /students/:id/portfolio/projects ----------
// Menambah satu proyek dengan $push, dibatasi 20 secara atomik.

func (s *PortfolioService) AddProject(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id mahasiswa harus berupa angka positif")
	}

	var req model.AddProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	// Tolak bila mahasiswanya tidak ada di PostgreSQL
	if _, err := s.students.FindByID(ctx, studentID); err != nil {
		return translatePortfolioError(err)
	}

	portfolio, err := s.portfolios.AddProject(ctx, studentID, model.Project{
		Title:        req.Title,
		Technologies: req.Technologies,
		Link:         req.Link,
		Year:         req.Year,
	})
	if err != nil {
		return translatePortfolioError(err)
	}
	return helper.Success(c, fiber.StatusCreated, "proyek berhasil ditambahkan", portfolio)
}

// ---------- DELETE /students/:id/portfolio/projects/:projectId ----------

func (s *PortfolioService) DeleteProject(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id mahasiswa harus berupa angka positif")
	}

	projectID := c.Params("projectId")
	if projectID == "" {
		return helper.BadRequest("id proyek tidak boleh kosong")
	}

	portfolio, err := s.portfolios.DeleteProject(ctx, studentID, projectID)
	if err != nil {
		return translatePortfolioError(err)
	}
	return helper.Success(c, fiber.StatusOK, "proyek berhasil dihapus", portfolio)
}

// ---------- GET /portfolios/stats/technologies ----------
// Aggregation pipeline: 10 teknologi paling banyak dipakai.

func (s *PortfolioService) TopTechnologies(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	result, err := s.portfolios.TopTechnologies(ctx, 10)
	if err != nil {
		return helper.Internal(err)
	}
	return helper.Success(c, fiber.StatusOK, "statistik teknologi berhasil diambil", result)
}

func translatePortfolioError(err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("mahasiswa atau portofolio tidak ditemukan")
	case errors.Is(err, repository.ErrProjectLimit):
		return helper.NewError(fiber.StatusConflict, "PROJECT_LIMIT",
			"jumlah proyek sudah mencapai batas maksimal 20")
	case errors.Is(err, repository.ErrProjectNotFound):
		return helper.NotFound("proyek tidak ditemukan")
	case errors.Is(err, repository.ErrInvalidID):
		return helper.BadRequest("format id proyek tidak valid")
	default:
		return helper.Internal(err)
	}
}

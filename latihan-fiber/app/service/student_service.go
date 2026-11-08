package service

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
	"latihan-fiber/helper"
)

type StudentService struct {
	repo         repository.StudentRepository
	authzChecker *helper.AuthzChecker
}

func NewStudentService(repo repository.StudentRepository, authzChecker *helper.AuthzChecker) *StudentService {
	return &StudentService{
		repo:         repo,
		authzChecker: authzChecker,
	}
}

// List godoc
//
// @Summary Menampilkan daftar student
// @Tags students
// @Produce json
// @Produce text/csv
// @Param limit query int false "Jumlah data per halaman" default(10)
// @Param cursor query string false "Cursor untuk pagination"
// @Param search query string false "Cari NIM atau nama"
// @Success 200 {object} model.WebResponse{data=[]model.Student}
// @Failure 401 {object} model.ErrorResponse
// @Failure 403 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /students [get]
func (s *StudentService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	authUser, ok := c.Locals(helper.LocalsAuthUser).(*model.AuthUser)
	if !ok {
		return helper.Unauthorized("user tidak terautentikasi")
	}

	if authUser.Role == "user" {
		return helper.Forbidden("tidak memiliki akses untuk melihat daftar student")
	}

	// Format dipilih SEBELUM query dijalankan.
	format, err := helper.Negotiate(c, helper.FormatJSON, helper.FormatCSV)
	if err != nil {
		return err
	}

	q, err := helper.ParseCursorQuery(c)
	if err != nil {
		return err
	}

	rows, err := s.repo.FindAfterCursor(ctx, q)
	if err != nil {
		return helper.Internal(err)
	}

	hasMore := len(rows) > q.Limit
	if hasMore {
		rows = rows[:q.Limit]
	}

	if format == helper.FormatCSV {
		return helper.WriteStudentsCSV(c, rows)
	}

	meta := &model.CursorMeta{Limit: q.Limit, HasMore: hasMore}
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		meta.NextCursor = helper.EncodeCursor(last.CreatedAt, last.ID)
	}

	return helper.SuccessCursor(c, "daftar student berhasil diambil", rows, meta)
}

// Get godoc
//
// @Summary Menampilkan detail student berdasarkan ID
// @Tags students
// @Produce json
// @Param id path int true "Student ID"
// @Success 200 {object} model.WebResponse{data=model.Student}
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 403 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /students/{id} [get]
func (s *StudentService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	authUser, ok := c.Locals(helper.LocalsAuthUser).(*model.AuthUser)
	if !ok {
		return helper.Unauthorized("user tidak terautentikasi")
	}

	canAccess, err := s.authzChecker.CanAccessStudent(ctx, authUser.UserID, authUser.Role, id, "student:read:any")
	if err != nil {
		return helper.Internal(err)
	}

	if !canAccess {
		return helper.Forbidden("tidak memiliki akses untuk melihat data student ini")
	}

	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateStudentError(err)
	}

	return helper.Success(c, fiber.StatusOK, "student ditemukan", student)
}

// Create godoc
//
// @Summary Menambahkan student baru
// @Tags students
// @Accept json
// @Produce json
// @Param body body model.CreateStudentRequest true "Data student baru"
// @Success 201 {object} model.WebResponse{data=model.Student}
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 403 {object} model.ErrorResponse
// @Failure 409 {object} model.ErrorResponse "NIM sudah terdaftar"
// @Failure 422 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /students [post]
func (s *StudentService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	authUser, ok := c.Locals(helper.LocalsAuthUser).(*model.AuthUser)
	if !ok {
		return helper.Unauthorized("user tidak terautentikasi")
	}

	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)
	req.Grade = strings.TrimSpace(req.Grade)

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	baru, err := s.repo.Create(ctx, model.Student{
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: true,
		OwnerID:  authUser.UserID,
	})
	if err != nil {
		return translateStudentError(err)
	}

	return helper.Created(c, "student berhasil dibuat", baru,
		"/api/v1/students/"+strconv.Itoa(baru.ID))
}

// Replace godoc
//
// @Summary Mengganti seluruh data student (PUT)
// @Tags students
// @Accept json
// @Produce json
// @Param id path int true "Student ID"
// @Param body body model.ReplaceStudentRequest true "Data lengkap student"
// @Success 200 {object} model.WebResponse{data=model.Student}
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 403 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 422 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /students/{id} [put]
func (s *StudentService) Replace(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	authUser, ok := c.Locals(helper.LocalsAuthUser).(*model.AuthUser)
	if !ok {
		return helper.Unauthorized("user tidak terautentikasi")
	}

	canAccess, err := s.authzChecker.CanAccessStudent(ctx, authUser.UserID, authUser.Role, id, "student:update:any")
	if err != nil {
		return helper.Internal(err)
	}

	if !canAccess {
		return helper.Forbidden("tidak memiliki akses untuk mengubah data student ini")
	}

	var req model.ReplaceStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateStudentError(err)
	}

	hasil, err := s.repo.Update(ctx, model.Student{
		ID:       id,
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: req.IsActive,
		OwnerID:  existing.OwnerID,
	})
	if err != nil {
		return translateStudentError(err)
	}

	return helper.Success(c, fiber.StatusOK, "student berhasil diganti seluruhnya", hasil)
}

// Patch godoc
//
// @Summary Memperbarui sebagian data student (PATCH)
// @Tags students
// @Accept json
// @Produce json
// @Param id path int true "Student ID"
// @Param body body model.PatchStudentRequest true "Field yang ingin diubah"
// @Success 200 {object} model.WebResponse{data=model.Student}
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 403 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 422 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /students/{id} [patch]
func (s *StudentService) Patch(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	authUser, ok := c.Locals(helper.LocalsAuthUser).(*model.AuthUser)
	if !ok {
		return helper.Unauthorized("user tidak terautentikasi")
	}

	canAccess, err := s.authzChecker.CanAccessStudent(ctx, authUser.UserID, authUser.Role, id, "student:update:any")
	if err != nil {
		return helper.Internal(err)
	}

	if !canAccess {
		return helper.Forbidden("tidak memiliki akses untuk mengubah data student ini")
	}

	var req model.PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if IsEmptyPatchStudent(req) {
		return helper.BadRequest("tidak ada field yang diubah")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	saatIni, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateStudentError(err)
	}

	updated := ApplyPatchStudent(saatIni, req)

	hasil, err := s.repo.Update(ctx, updated)
	if err != nil {
		return translateStudentError(err)
	}

	return helper.Success(c, fiber.StatusOK, "student berhasil diperbarui sebagian", hasil)
}

// Delete godoc
//
// @Summary Menghapus student berdasarkan ID
// @Tags students
// @Param id path int true "Student ID"
// @Success 204 "Berhasil dihapus; tanpa body"
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 403 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /students/{id} [delete]
func (s *StudentService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	authUser, ok := c.Locals(helper.LocalsAuthUser).(*model.AuthUser)
	if !ok {
		return helper.Unauthorized("user tidak terautentikasi")
	}

	if authUser.Role != "admin" {
		return helper.Forbidden("hanya admin yang bisa menghapus data student")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return translateStudentError(err)
	}

	return helper.NoContent(c)
}

// translateStudentError mengubah error repository menjadi AppError.
func translateStudentError(err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("student tidak ditemukan")
	case errors.Is(err, repository.ErrDuplicate):
		return helper.Conflict("NIM sudah dipakai")
	default:
		return helper.Internal(err)
	}
}

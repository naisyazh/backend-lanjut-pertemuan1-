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

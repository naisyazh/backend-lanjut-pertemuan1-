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
	repo        repository.StudentRepository
	authzChecker *helper.AuthzChecker
}

func NewStudentService(repo repository.StudentRepository, authzChecker *helper.AuthzChecker) *StudentService {
	return &StudentService{
		repo:        repo,
		authzChecker: authzChecker,
	}
}

func (s *StudentService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	// Ambil user dari context
	authUser, ok := c.Locals(helper.LocalsAuthUser).(*model.AuthUser)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "user tidak terautentikasi")
	}

	// User biasa tidak bisa akses list
	if authUser.Role == "user" {
		return helper.Fail(c, fiber.StatusForbidden, "tidak memiliki akses untuk melihat daftar student")
	}

	q := helper.ParseListQuery(c)
	students, total, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data student")
	}

	totalPages := 0
	if q.Limit > 0 {
		totalPages = (total + q.Limit - 1) / q.Limit
	}

	return helper.SuccessList(c, "daftar student berhasil diambil", students, &model.Meta{
		Page:       q.Page,
		Limit:      q.Limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

func (s *StudentService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	// Ambil user dari context
	authUser, ok := c.Locals(helper.LocalsAuthUser).(*model.AuthUser)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "user tidak terautentikasi")
	}

	// Cek authorization
	canAccess, err := s.authzChecker.CanAccessStudent(ctx, authUser.UserID, authUser.Role, id, "student:read:any")
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal memeriksa akses")
	}

	if !canAccess {
		return helper.Fail(c, fiber.StatusForbidden, "tidak memiliki akses untuk melihat data student ini")
	}

	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateErrorStudent(c, err, "gagal mengambil data student")
	}

	return helper.Success(c, fiber.StatusOK, "student ditemukan", student)
}

func (s *StudentService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	// Ambil user dari context
	authUser, ok := c.Locals(helper.LocalsAuthUser).(*model.AuthUser)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "user tidak terautentikasi")
	}

	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)
	req.Grade = strings.TrimSpace(req.Grade)

	if errs := ValidateCreateStudent(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	// Auto-assign owner_id ke user yang membuat
	ownerID := authUser.UserID
	if authUser.Role == "admin" || authUser.Role == "staff" {
		// Admin dan staff bisa buat atas nama siapa pun, default ke diri sendiri
		ownerID = authUser.UserID
	}

	baru, err := s.repo.Create(ctx, model.Student{
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: true,
		OwnerID:  ownerID,
	})
	if err != nil {
		return translateErrorStudent(c, err, "gagal menyimpan student")
	}

	return helper.Created(c, "student berhasil dibuat", baru,
		"/api/v1/students/"+strconv.Itoa(baru.ID))
}

func (s *StudentService) Replace(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	// Ambil user dari context
	authUser, ok := c.Locals(helper.LocalsAuthUser).(*model.AuthUser)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "user tidak terautentikasi")
	}

	// Cek authorization untuk update
	canAccess, err := s.authzChecker.CanAccessStudent(ctx, authUser.UserID, authUser.Role, id, "student:update:any")
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal memeriksa akses")
	}

	if !canAccess {
		return helper.Fail(c, fiber.StatusForbidden, "tidak memiliki akses untuk mengubah data student ini")
	}

	var req model.ReplaceStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest,
			"body harus berupa JSON yang valid")
	}

	if errs := ValidateReplaceStudent(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	// Ambil data lama untuk preserve owner_id
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateErrorStudent(c, err, "gagal mengambil data student")
	}

	hasil, err := s.repo.Update(ctx, model.Student{
		ID:       id,
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: req.IsActive,
		OwnerID:  existing.OwnerID, // Preserve owner_id
	})
	if err != nil {
		return translateErrorStudent(c, err, "gagal memperbarui student")
	}

	return helper.Success(c, fiber.StatusOK, "student berhasil diganti seluruhnya", hasil)
}

func (s *StudentService) Patch(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	// Ambil user dari context
	authUser, ok := c.Locals(helper.LocalsAuthUser).(*model.AuthUser)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "user tidak terautentikasi")
	}

	// Cek authorization untuk update
	canAccess, err := s.authzChecker.CanAccessStudent(ctx, authUser.UserID, authUser.Role, id, "student:update:any")
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal memeriksa akses")
	}

	if !canAccess {
		return helper.Fail(c, fiber.StatusForbidden, "tidak memiliki akses untuk mengubah data student ini")
	}

	var req model.PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest,
			"body harus berupa JSON yang valid")
	}

	if IsEmptyPatchStudent(req) {
		return helper.Fail(c, fiber.StatusBadRequest, "tidak ada field yang diubah")
	}

	saatIni, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateErrorStudent(c, err, "gagal mengambil data student")
	}

	updated, errs := ApplyPatchStudent(saatIni, req)
	if len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	hasil, err := s.repo.Update(ctx, updated)
	if err != nil {
		return translateErrorStudent(c, err, "gagal memperbarui student")
	}

	return helper.Success(c, fiber.StatusOK, "student berhasil diperbarui sebagian", hasil)
}

func (s *StudentService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	// Ambil user dari context
	authUser, ok := c.Locals(helper.LocalsAuthUser).(*model.AuthUser)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "user tidak terautentikasi")
	}

	// Hanya admin yang bisa delete
	if authUser.Role != "admin" {
		return helper.Fail(c, fiber.StatusForbidden, "hanya admin yang bisa menghapus data student")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return translateErrorStudent(c, err, "gagal menghapus student")
	}

	return helper.NoContent(c)
}

func translateErrorStudent(c *fiber.Ctx, err error, pesanUmum string) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.Fail(c, fiber.StatusNotFound, "student tidak ditemukan")
	case errors.Is(err, repository.ErrDuplicate):
		return helper.Fail(c, fiber.StatusConflict, "NIM sudah dipakai")
	default:
		return helper.Fail(c, fiber.StatusInternalServerError, pesanUmum)
	}
}

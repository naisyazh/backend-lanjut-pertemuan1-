package service

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
	"latihan-fiber/helper"
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// List godoc
//
// @Summary Menampilkan daftar user
// @Tags users
// @Produce json
// @Produce text/csv
// @Param limit query int false "Jumlah data per halaman" default(10)
// @Param cursor query string false "Cursor untuk pagination"
// @Param search query string false "Cari username atau email"
// @Success 200 {object} model.WebResponse{data=[]model.User}
// @Failure 401 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /users [get]
func (s *UserService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	// Format dipilih SEBELUM query dijalankan. Bila client meminta format
	// yang tidak dapat kita hasilkan, tidak ada gunanya membebani database
	// untuk hasil yang akan dibuang.
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

	// Baris tambahan hasil limit+1 dipotong di sini. Ia hanya penanda bahwa
	// masih ada halaman berikutnya, bukan bagian dari halaman ini.
	hasMore := len(rows) > q.Limit
	if hasMore {
		rows = rows[:q.Limit]
	}

	if format == helper.FormatCSV {
		return helper.WriteUsersCSV(c, rows)
	}

	meta := &model.CursorMeta{Limit: q.Limit, HasMore: hasMore}
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		meta.NextCursor = helper.EncodeCursor(last.CreatedAt, last.ID)
	}

	return helper.SuccessCursor(c, "daftar user berhasil diambil", rows, meta)
}

// Get godoc
//
// @Summary Menampilkan detail user berdasarkan ID
// @Tags users
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} model.WebResponse{data=model.User}
// @Failure 400 {object} model.ErrorResponse "ID bukan angka positif"
// @Failure 401 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse "User tidak ditemukan"
// @Security BearerAuth
// @Router /users/{id} [get]
func (s *UserService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateUserError(err)
	}

	return helper.Success(c, fiber.StatusOK, "user ditemukan", user)
}

// Create godoc
//
// @Summary Membuat user baru
// @Tags users
// @Accept json
// @Produce json
// @Param body body model.CreateUserRequest true "Data user baru"
// @Success 201 {object} model.WebResponse{data=model.User}
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 409 {object} model.ErrorResponse "Username sudah dipakai"
// @Failure 422 {object} model.ErrorResponse "Validasi gagal"
// @Security BearerAuth
// @Router /users [post]
func (s *UserService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	newUser, err := s.repo.Create(ctx, model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		IsActive: true,
	})
	if err != nil {
		return translateUserError(err)
	}

	return helper.Created(c, "user berhasil dibuat", newUser,
		"/api/v1/users/"+strconv.Itoa(newUser.ID))
}

// Replace godoc
//
// @Summary Mengganti seluruh data user (PUT)
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param body body model.ReplaceUserRequest true "Data user lengkap"
// @Success 200 {object} model.WebResponse{data=model.User}
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse "User tidak ditemukan"
// @Failure 422 {object} model.ErrorResponse "Validasi gagal"
// @Security BearerAuth
// @Router /users/{id} [put]
func (s *UserService) Replace(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	var req model.ReplaceUserRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	result, err := s.repo.Update(ctx, model.User{
		ID:       id,
		Username: req.Username,
		Email:    req.Email,
		IsActive: req.IsActive,
	})
	if err != nil {
		return translateUserError(err)
	}

	return helper.Success(c, fiber.StatusOK, "user berhasil diganti seluruhnya", result)
}

// Patch godoc
//
// @Summary Memperbarui sebagian data user (PATCH)
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param body body model.PatchUserRequest true "Field yang ingin diperbarui"
// @Success 200 {object} model.WebResponse{data=model.User}
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse "User tidak ditemukan"
// @Failure 422 {object} model.ErrorResponse "Validasi gagal"
// @Security BearerAuth
// @Router /users/{id} [patch]
func (s *UserService) Patch(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	var req model.PatchUserRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if IsEmptyPatch(req) {
		return helper.BadRequest("tidak ada field yang diubah")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	current, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateUserError(err)
	}

	updated := ApplyPatch(current, req)

	result, err := s.repo.Update(ctx, updated)
	if err != nil {
		return translateUserError(err)
	}

	return helper.Success(c, fiber.StatusOK, "user berhasil diperbarui sebagian", result)
}

// Delete godoc
//
// @Summary Menghapus user berdasarkan ID
// @Tags users
// @Param id path int true "User ID"
// @Success 204 "Berhasil dihapus; tanpa body"
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse "User tidak ditemukan"
// @Security BearerAuth
// @Router /users/{id} [delete]
func (s *UserService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return translateUserError(err)
	}

	return helper.NoContent(c)
}

// translateUserError mengubah error repository menjadi AppError.
// Tidak ada fiber.Ctx — fungsi ini hanya menerjemahkan, tidak tahu HTTP.
// Yang tidak dikenali menjadi Internal — fail closed: lebih baik 500 daripada menebak.
func translateUserError(err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("user tidak ditemukan")
	case errors.Is(err, repository.ErrDuplicate):
		return helper.Conflict("username sudah dipakai")
	default:
		return helper.Internal(err)
	}
}

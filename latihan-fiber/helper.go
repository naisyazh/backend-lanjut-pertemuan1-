package main

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"latihan-fiber/app/model"
)

// reqCtx memberi batas waktu untuk setiap operasi basis data.
// Tanpa batas waktu, satu query yang menggantung dapat menahan koneksi
// selamanya dan lama-lama menghabiskan seluruh isi pool.
func reqCtx(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.UserContext(), 5*time.Second)
}

// ok mengembalikan respons sukses dengan data.
func ok(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(model.WebResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

// okList mengembalikan respons sukses dengan data berupa daftar dan metadata paginasi.
func okList(c *fiber.Ctx, message string, data interface{}, meta *model.Meta) error {
	return c.Status(fiber.StatusOK).JSON(model.WebResponse{
		Status:  "success",
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// created mengembalikan respons 201 Created dengan header Location.
func created(c *fiber.Ctx, message string, data interface{}, location string) error {
	c.Set("Location", location)
	return c.Status(fiber.StatusCreated).JSON(model.WebResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

// noContent mengembalikan respons 204 No Content tanpa body.
func noContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// fail mengembalikan respons gagal dengan status dan pesan tertentu.
func fail(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(model.WebResponse{
		Status:  "fail",
		Message: message,
	})
}

// failValidation mengembalikan respons gagal validasi dengan detail kesalahan per field.
func failValidation(c *fiber.Ctx, errors map[string]string) error {
	return c.Status(fiber.StatusBadRequest).JSON(model.WebResponse{
		Status:  "fail",
		Message: "validasi gagal",
		Data:    errors,
	})
}

// parseListQuery mengambil parameter query dari request.
func parseListQuery(c *fiber.Ctx) model.ListQuery {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	sort := c.Query("sort", "id")
	validSorts := map[string]bool{"id": true, "username": true, "email": true, "created_at": true}
	if !validSorts[sort] {
		sort = "id"
	}

	order := c.Query("order", "asc")
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	search := c.Query("search", "")

	var isActive *bool
	if c.Query("is_active") != "" {
		val := c.Query("is_active") == "true"
		isActive = &val
	}

	return model.ListQuery{
		Page:     page,
		Limit:    limit,
		Search:   search,
		Sort:     sort,
		Order:    order,
		IsActive: isActive,
	}
}

// paramID mengambil parameter :id dari path dan memvalidasinya.
func paramID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}

// requireJSON adalah middleware yang memastikan Content-Type adalah application/json.
func requireJSON(c *fiber.Ctx) error {
	if c.Method() == fiber.MethodPost || c.Method() == fiber.MethodPut || c.Method() == fiber.MethodPatch {
		if c.Get("Content-Type") != "application/json" {
			return fail(c, fiber.StatusUnsupportedMediaType, "Content-Type harus application/json")
		}
	}
	return c.Next()
}

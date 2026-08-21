package main

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

var students = []Student{
	{ID: 1, NIM: "434241068", Name: "Naisya", Grade: 85.5, IsActive: true},
	{ID: 2, NIM: "434241124", Name: "Fitrah", Grade: 78.0, IsActive: true},
}

var nextID = 3

func GetStudents(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	search := c.Query("search", "")
	sort := c.Query("sort", "")
	filterActive := c.Query("aktif", "")

	result := make([]Student, 0)
	for _, s := range students {
		if search != "" && !strings.Contains(strings.ToLower(s.Name), strings.ToLower(search)) {
			continue
		}

		if filterActive != "" {
			isActive, _ := strconv.ParseBool(filterActive)
			if s.IsActive != isActive {
				continue
			}
		}

		result = append(result, s)
	}

	if sort != "" {
		
	}

	start := (page - 1) * limit
	end := start + limit

	if start > len(result) {
		start = len(result)
	}
	if end > len(result) {
		end = len(result)
	}

	paginatedResult := result[start:end]

	return SuccessResponse(c, 200, "Berhasil mengambil data students", fiber.Map{
		"students": paginatedResult,
		"total":    len(result),
		"page":     page,
		"limit":    limit,
	})
}

func GetStudentByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return ErrorResponse(c, 400, "ID harus berupa angka")
	}

	student := findStudentByID(id)
	if student == nil {
		return ErrorResponse(c, 404, "Student tidak ditemukan")
	}

	return SuccessResponse(c, 200, "Berhasil mengambil data student", student)
}

func CreateStudent(c *fiber.Ctx) error {
	// Cek Content-Type
	if c.Get("Content-Type") != "application/json" {
		return ErrorResponse(c, 415, "Content-Type harus application/json")
	}

	var req CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, 400, "Format JSON tidak valid")
	}

	if req.NIM == "" || req.Name == "" {
		return ErrorResponse(c, 422, "NIM dan Name wajib diisi")
	}

	if isNIMExists(req.NIM, 0) {
		return ErrorResponse(c, 409, "NIM sudah terdaftar")
	}

	newStudent := Student{
		ID:       nextID,
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: req.IsActive,
	}
	students = append(students, newStudent)
	nextID++

	return SuccessResponse(c, 201, "Student berhasil ditambahkan", newStudent)
}

func UpdateStudent(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return ErrorResponse(c, 400, "ID harus berupa angka")
	}

	student := findStudentByID(id)
	if student == nil {
		return ErrorResponse(c, 404, "Student tidak ditemukan")
	}

	var req UpdateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, 400, "Format JSON tidak valid")
	}

	if req.NIM == "" || req.Name == "" {
		return ErrorResponse(c, 422, "NIM dan Name wajib diisi")
	}

	if isNIMExists(req.NIM, id) {
		return ErrorResponse(c, 409, "NIM sudah terdaftar")
	}

	student.NIM = req.NIM
	student.Name = req.Name
	student.Grade = req.Grade
	student.IsActive = req.IsActive

	return SuccessResponse(c, 200, "Student berhasil diupdate", student)
}

func PatchStudent(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return ErrorResponse(c, 400, "ID harus berupa angka")
	}

	student := findStudentByID(id)
	if student == nil {
		return ErrorResponse(c, 404, "Student tidak ditemukan")
	}

	var req PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, 400, "Format JSON tidak valid")
	}

	if req.NIM != nil {
		if isNIMExists(*req.NIM, id) {
			return ErrorResponse(c, 409, "NIM sudah terdaftar")
		}
		student.NIM = *req.NIM
	}
	if req.Name != nil {
		student.Name = *req.Name
	}
	if req.Grade != nil {
		student.Grade = *req.Grade
	}
	if req.IsActive != nil {
		student.IsActive = *req.IsActive
	}

	return SuccessResponse(c, 200, "Student berhasil diupdate", student)
}

func DeleteStudent(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return ErrorResponse(c, 400, "ID harus berupa angka")
	}

	index := -1
	for i, s := range students {
		if s.ID == id {
			index = i
			break
		}
	}

	if index == -1 {
		return ErrorResponse(c, 404, "Student tidak ditemukan")
	}

	students = append(students[:index], students[index+1:]...)

	return c.SendStatus(204) // 204 No Content
}

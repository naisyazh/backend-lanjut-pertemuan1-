package service

import (
	"github.com/gofiber/fiber/v2"

	"latihan-fiber/app/model"
	"latihan-fiber/helper"
)

// GetNilaiByNIM - Handler untuk ambil nilai berdasarkan NIM
// GET /api/v1/nilai/nim/:nim
func (s *NilaiService) GetNilaiByNIM_Handler(c *fiber.Ctx) error {
	nim := c.Params("nim")
	
	if nim == "" {
		return helper.Fail(c, fiber.StatusBadRequest, "NIM tidak boleh kosong")
	}

	nilaiList, err := s.GetNilaiByNIM(nim)
	if err != nil {
		return helper.Fail(c, fiber.StatusNotFound, err.Error())
	}

	return helper.Success(c, fiber.StatusOK, "Berhasil mengambil data nilai", nilaiList)
}

// GetAllNilai - Handler untuk ambil semua nilai
// GET /api/v1/nilai
func (s *NilaiService) GetAllNilai_Handler(c *fiber.Ctx) error {
	nilaiList, err := s.GetAllNilai()
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, err.Error())
	}

	return helper.Success(c, fiber.StatusOK, "Berhasil mengambil semua data nilai", nilaiList)
}

// CreateNilai - Handler untuk tambah nilai baru
// POST /api/v1/nilai
func (s *NilaiService) CreateNilai_Handler(c *fiber.Ctx) error {
	var req model.CreateNilaiRequest
	
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "Format request tidak valid")
	}

	nilai, err := s.CreateNilai(req)
	if err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, err.Error())
	}

	return helper.Success(c, fiber.StatusCreated, "Berhasil menambahkan nilai", nilai)
}

// GetStatistikNilai - Handler untuk statistik nilai mahasiswa
// GET /api/v1/nilai/statistik/:nim
func (s *NilaiService) GetStatistikNilai_Handler(c *fiber.Ctx) error {
	nim := c.Params("nim")
	
	if nim == "" {
		return helper.Fail(c, fiber.StatusBadRequest, "NIM tidak boleh kosong")
	}

	statistik, err := s.GetStatistikNilai(nim)
	if err != nil {
		return helper.Fail(c, fiber.StatusNotFound, err.Error())
	}

	return helper.Success(c, fiber.StatusOK, "Berhasil mengambil statistik nilai", statistik)
}

// GetNilaiByMatkul - Handler untuk cari nilai berdasarkan mata kuliah
// GET /api/v1/nilai/matkul?nama=praktikum
func (s *NilaiService) GetNilaiByMatkul_Handler(c *fiber.Ctx) error {
	matkul := c.Query("nama")
	
	if matkul == "" {
		return helper.Fail(c, fiber.StatusBadRequest, "Parameter 'nama' mata kuliah tidak boleh kosong")
	}

	// Simple filter from all nilai
	allNilai, err := s.GetAllNilai()
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, err.Error())
	}

	var filtered []model.NilaiWithStudent
	for _, nilai := range allNilai {
		// Simple case-insensitive search
		if contains(nilai.NamaMataKuliah, matkul) {
			filtered = append(filtered, nilai)
		}
	}

	return helper.Success(c, fiber.StatusOK, "Berhasil mencari nilai berdasarkan mata kuliah", filtered)
}

// Helper function for simple string contains (case-insensitive)
func contains(source, target string) bool {
	return len(source) >= len(target) && 
		   (source == target || 
		    strContainsIgnoreCase(source, target))
}

func strContainsIgnoreCase(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	return strContains(s, substr)
}

func toLower(s string) string {
	result := make([]rune, len(s))
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			result[i] = r + 32
		} else {
			result[i] = r
		}
	}
	return string(result)
}

func strContains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
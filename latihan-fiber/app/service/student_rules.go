package service

import (
	"strings"

	"latihan-fiber/app/model"
)

// ValidateCreateStudent validasi untuk POST student
func ValidateCreateStudent(req model.CreateStudentRequest) map[string]string {
	errs := map[string]string{}

	if strings.TrimSpace(req.NIM) == "" {
		errs["nim"] = "wajib diisi"
	}
	if strings.TrimSpace(req.Name) == "" {
		errs["name"] = "wajib diisi"
	}

	return errs
}

// ValidateReplaceStudent validasi untuk PUT student
func ValidateReplaceStudent(req model.ReplaceStudentRequest) map[string]string {
	errs := map[string]string{}

	if strings.TrimSpace(req.NIM) == "" {
		errs["nim"] = "wajib diisi pada PUT"
	}
	if strings.TrimSpace(req.Name) == "" {
		errs["name"] = "wajib diisi pada PUT"
	}

	return errs
}

// ApplyPatchStudent terapkan perubahan PATCH untuk student.
// Pemeriksaan bentuk sudah dikerjakan tag sebelum fungsi ini dipanggil.
func ApplyPatchStudent(
	current model.Student, req model.PatchStudentRequest,
) model.Student {
	if req.NIM != nil {
		current.NIM = strings.TrimSpace(*req.NIM)
	}
	if req.Name != nil {
		current.Name = strings.TrimSpace(*req.Name)
	}
	if req.Grade != nil {
		current.Grade = *req.Grade
	}
	if req.IsActive != nil {
		current.IsActive = *req.IsActive
	}
	return current
}

// IsEmptyPatchStudent cek apakah PATCH student kosong
func IsEmptyPatchStudent(req model.PatchStudentRequest) bool {
	return req.NIM == nil && req.Name == nil && req.Grade == nil && req.IsActive == nil
}

package service

import (
	"strings"

	"latihan-fiber/app/model"
)

// ValidateCreateStudent dan ValidateReplaceStudent dihapus.
// Aturan validasi sudah dipindahkan ke tag validate pada struct request.
// Yang tersisa hanyalah aturan yang tidak dapat dinyatakan sebagai tag.

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

// IsEmptyPatchStudent memeriksa body PATCH yang tidak berisi field apa pun.
// Tidak dapat ditulis sebagai tag karena menyangkut hubungan antar field.
func IsEmptyPatchStudent(req model.PatchStudentRequest) bool {
	return req.NIM == nil && req.Name == nil && req.Grade == nil && req.IsActive == nil
}

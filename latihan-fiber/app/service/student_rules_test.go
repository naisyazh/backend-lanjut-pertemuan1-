package service

import (
	"testing"

	"latihan-fiber/app/model"
)

func TestValidateCreateStudent(t *testing.T) {
	tests := []struct {
		name string
		req  model.CreateStudentRequest
		want int // jumlah error
	}{
		{
			name: "valid",
			req: model.CreateStudentRequest{
				NIM:   "434241068",
				Name:  "Naisya Gina Azzahra",
				Grade: "A",
			},
			want: 0,
		},
		{
			name: "nim kosong",
			req: model.CreateStudentRequest{
				NIM:   "",
				Name:  "Naisya",
				Grade: "A",
			},
			want: 1,
		},
		{
			name: "name kosong",
			req: model.CreateStudentRequest{
				NIM:   "434241068",
				Name:  "",
				Grade: "A",
			},
			want: 1,
		},
		{
			name: "semua kosong",
			req: model.CreateStudentRequest{
				NIM:   "",
				Name:  "",
				Grade: "",
			},
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateCreateStudent(tt.req)
			if len(errs) != tt.want {
				t.Errorf("ValidateCreateStudent() got %d errors, want %d. Errors: %v",
					len(errs), tt.want, errs)
			}
		})
	}
}

func TestApplyPatchStudent(t *testing.T) {
	initial := model.Student{
		ID:       1,
		NIM:      "434241068",
		Name:     "Naisya",
		Grade:    "A",
		IsActive: true,
	}

	// Test: ubah grade saja
	newGrade := "B"
	result, errs := ApplyPatchStudent(initial, model.PatchStudentRequest{
		Grade: &newGrade,
	})

	if len(errs) != 0 {
		t.Fatalf("tidak seharusnya ada error: %v", errs)
	}
	if result.Grade != "B" {
		t.Error("grade seharusnya berubah menjadi B")
	}
	if result.NIM != "434241068" {
		t.Error("field yang tidak dikirim seharusnya tidak berubah")
	}
}

func TestIsEmptyPatchStudent(t *testing.T) {
	// Test: patch kosong
	if !IsEmptyPatchStudent(model.PatchStudentRequest{}) {
		t.Error("seharusnya mendeteksi patch kosong")
	}

	// Test: patch ada isinya
	grade := "A"
	if IsEmptyPatchStudent(model.PatchStudentRequest{Grade: &grade}) {
		t.Error("seharusnya mendeteksi patch tidak kosong")
	}
}

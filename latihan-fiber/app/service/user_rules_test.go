package service

import (
	"testing"

	"latihan-fiber/app/model"
)

// Perhatikan: pengujian ini tidak menyalakan server, tidak menyentuh
// database, dan tidak membuat fiber.Ctx.

func TestCountTotalPages(t *testing.T) {
	cases := []struct{ total, limit, want int }{
		{0, 10, 0},
		{1, 10, 1},
		{10, 10, 1},
		{11, 10, 2},
		{137, 20, 7},
	}

	for _, tc := range cases {
		if got := CountTotalPages(tc.total, tc.limit); got != tc.want {
			t.Errorf("total=%d limit=%d: harap %d, dapat %d",
				tc.total, tc.limit, tc.want, got)
		}
	}
}

func TestApplyPatch(t *testing.T) {
	initial := model.User{ID: 1, Username: "sari", Email: "sari@mail.com", IsActive: true}
	inactive := false
	result, errs := ApplyPatch(initial, model.PatchUserRequest{IsActive: &inactive})

	if len(errs) != 0 {
		t.Fatalf("tidak seharusnya ada error: %v", errs)
	}
	if result.IsActive {
		t.Error("is_active seharusnya berubah menjadi false")
	}
	if result.Username != "sari" {
		t.Error("field yang tidak dikirim seharusnya tidak berubah")
	}
}

func TestValidateCreate(t *testing.T) {
	tests := []struct {
		name string
		req  model.CreateUserRequest
		want int // jumlah error yang diharapkan
	}{
		{
			name: "valid",
			req: model.CreateUserRequest{
				Username: "naisya",
				Email:    "naisya@mail.com",
				Password: "password123",
			},
			want: 0,
		},
		{
			name: "username kosong",
			req: model.CreateUserRequest{
				Username: "",
				Email:    "test@mail.com",
				Password: "password123",
			},
			want: 1,
		},
		{
			name: "email invalid",
			req: model.CreateUserRequest{
				Username: "naisya",
				Email:    "bukan-email",
				Password: "password123",
			},
			want: 1,
		},
		{
			name: "password terlalu pendek",
			req: model.CreateUserRequest{
				Username: "naisya",
				Email:    "naisya@mail.com",
				Password: "123",
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateCreate(tt.req)
			if len(errs) != tt.want {
				t.Errorf("ValidateCreate() got %d errors, want %d. Errors: %v",
					len(errs), tt.want, errs)
			}
		})
	}
}

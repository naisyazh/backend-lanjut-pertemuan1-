package service

import (
	"testing"

	"latihan-fiber/app/model"
)

func TestCheckPasswordStrength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  string
	}{
		{
			name:     "Valid strong password",
			password: "StrongPass123",
			wantErr:  "",
		},
		{
			name:     "Password too short",
			password: "abc123",
			wantErr:  "minimal 8 karakter",
		},
		{
			name:     "Password without letter",
			password: "12345678",
			wantErr:  "harus memuat huruf dan angka",
		},
		{
			name:     "Password without digit",
			password: "abcdefgh",
			wantErr:  "harus memuat huruf dan angka",
		},
		{
			name:     "Common weak password",
			password: "password123",
			wantErr:  "password terlalu umum",
		},
		{
			name:     "Another weak password",
			password: "admin123",
			wantErr:  "password terlalu umum",
		},
		{
			name:     "Valid password with symbols",
			password: "MyPass123!",
			wantErr:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkPasswordStrength(tt.password)
			if got != tt.wantErr {
				t.Errorf("checkPasswordStrength() = %v, want %v", got, tt.wantErr)
			}
		})
	}
}

func TestValidateRegister(t *testing.T) {
	tests := []struct {
		name    string
		req     model.RegisterRequest
		wantErr bool
		errKey  string
	}{
		{
			name: "Valid registration",
			req: model.RegisterRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "validpass123",
			},
			wantErr: false,
		},
		{
			name: "Empty username",
			req: model.RegisterRequest{
				Username: "",
				Email:    "test@example.com",
				Password: "validpass123",
			},
			wantErr: true,
			errKey:  "username",
		},
		{
			name: "Short username",
			req: model.RegisterRequest{
				Username: "ab",
				Email:    "test@example.com",
				Password: "validpass123",
			},
			wantErr: true,
			errKey:  "username",
		},
		{
			name: "Invalid email",
			req: model.RegisterRequest{
				Username: "testuser",
				Email:    "invalid-email",
				Password: "validpass123",
			},
			wantErr: true,
			errKey:  "email",
		},
		{
			name: "Weak password",
			req: model.RegisterRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			wantErr: true,
			errKey:  "password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateRegister(tt.req)
			
			if tt.wantErr {
				if len(errs) == 0 {
					t.Errorf("ValidateRegister() expected errors but got none")
				}
				if tt.errKey != "" {
					if _, exists := errs[tt.errKey]; !exists {
						t.Errorf("ValidateRegister() expected error key %s but not found", tt.errKey)
					}
				}
			} else {
				if len(errs) > 0 {
					t.Errorf("ValidateRegister() expected no errors but got: %v", errs)
				}
			}
		})
	}
}

func TestValidateLogin(t *testing.T) {
	tests := []struct {
		name    string
		req     model.LoginRequest
		wantErr bool
	}{
		{
			name: "Valid login",
			req: model.LoginRequest{
				Username: "testuser",
				Password: "anypassword",
			},
			wantErr: false,
		},
		{
			name: "Empty username",
			req: model.LoginRequest{
				Username: "",
				Password: "anypassword",
			},
			wantErr: true,
		},
		{
			name: "Empty password",
			req: model.LoginRequest{
				Username: "testuser",
				Password: "",
			},
			wantErr: true,
		},
		{
			name: "Both empty",
			req: model.LoginRequest{
				Username: "",
				Password: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateLogin(tt.req)
			
			if tt.wantErr && len(errs) == 0 {
				t.Errorf("ValidateLogin() expected errors but got none")
			}
			if !tt.wantErr && len(errs) > 0 {
				t.Errorf("ValidateLogin() expected no errors but got: %v", errs)
			}
		})
	}
}
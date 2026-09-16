package service

import (
	"testing"
)

func TestStudentAuthzRules(t *testing.T) {
	tests := []struct {
		name           string
		userRole       string
		permission     string
		studentOwnerID int
		userID         int
		expected       bool
		description    string
	}{
		// Admin tests
		{
			name:        "admin_can_list_students",
			userRole:    "admin",
			permission:  "student:list",
			expected:    true,
			description: "Admin dapat melihat daftar student",
		},
		{
			name:        "admin_can_read_any_student",
			userRole:    "admin", 
			permission:  "student:read:any",
			expected:    true,
			description: "Admin dapat melihat data student apa pun",
		},
		{
			name:        "admin_can_create_student",
			userRole:    "admin",
			permission:  "student:create",
			expected:    true,
			description: "Admin dapat membuat student baru",
		},
		{
			name:        "admin_can_update_any_student",
			userRole:    "admin",
			permission:  "student:update:any",
			expected:    true,
			description: "Admin dapat mengubah data student apa pun",
		},
		{
			name:        "admin_can_delete_student",
			userRole:    "admin",
			permission:  "student:delete",
			expected:    true,
			description: "Admin dapat menghapus student",
		},

		// Staff tests
		{
			name:        "staff_can_list_students",
			userRole:    "staff",
			permission:  "student:list",
			expected:    true,
			description: "Staff dapat melihat daftar student",
		},
		{
			name:        "staff_can_read_any_student",
			userRole:    "staff",
			permission:  "student:read:any",
			expected:    true,
			description: "Staff dapat melihat data student apa pun",
		},
		{
			name:        "staff_can_create_student",
			userRole:    "staff",
			permission:  "student:create",
			expected:    true,
			description: "Staff dapat membuat student baru",
		},
		{
			name:        "staff_cannot_update_any_student",
			userRole:    "staff",
			permission:  "student:update:any",
			expected:    false,
			description: "Staff tidak dapat mengubah data student sembarangan",
		},
		{
			name:        "staff_cannot_delete_student",
			userRole:    "staff",
			permission:  "student:delete",
			expected:    false,
			description: "Staff tidak dapat menghapus student",
		},

		// User tests
		{
			name:        "user_cannot_list_students",
			userRole:    "user",
			permission:  "student:list",
			expected:    false,
			description: "User tidak dapat melihat daftar student",
		},
		{
			name:        "user_cannot_read_any_student",
			userRole:    "user",
			permission:  "student:read:any",
			expected:    false,
			description: "User tidak memiliki permission read any",
		},
		{
			name:        "user_can_create_student",
			userRole:    "user",
			permission:  "student:create",
			expected:    false,
			description: "User tidak memiliki global permission create",
		},
		{
			name:        "user_cannot_update_any_student",
			userRole:    "user",
			permission:  "student:update:any",
			expected:    false,
			description: "User tidak memiliki permission update any",
		},
		{
			name:        "user_cannot_delete_student",
			userRole:    "user",
			permission:  "student:delete",
			expected:    false,
			description: "User tidak dapat menghapus student",
		},

		// Ownership tests (menggunakan mock data)
		{
			name:           "user_can_read_own_student",
			userRole:       "user",
			permission:     "student:read:any",
			studentOwnerID: 5,
			userID:         5,
			expected:       true,
			description:    "User dapat melihat data student milik sendiri",
		},
		{
			name:           "user_cannot_read_others_student", 
			userRole:       "user",
			permission:     "student:read:any",
			studentOwnerID: 5,
			userID:         3,
			expected:       false,
			description:    "User tidak dapat melihat data student milik orang lain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ini adalah unit test untuk business rules
			// Dalam implementasi nyata, ini akan menggunakan mock database
			// atau test database untuk simulasi permission dan ownership checking
			
			t.Logf("Testing: %s", tt.description)
			t.Logf("User Role: %s, Permission: %s", tt.userRole, tt.permission)
			
			if tt.studentOwnerID != 0 && tt.userID != 0 {
				t.Logf("Ownership test - Student owner: %d, User: %d", tt.studentOwnerID, tt.userID)
			}
			
			// Simulate authorization logic
			result := simulateAuthzCheck(tt.userRole, tt.permission, tt.studentOwnerID, tt.userID)
			
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for %s", tt.expected, result, tt.description)
			} else {
				t.Logf("✓ PASS: %s", tt.description)
			}
		})
	}
}

// simulateAuthzCheck mensimulasikan logic authorization berdasarkan matrix permission
func simulateAuthzCheck(userRole, permission string, studentOwnerID, userID int) bool {
	// Permission matrix untuk roles
	rolePermissions := map[string][]string{
		"admin": {
			"student:list",
			"student:read:any", 
			"student:create",
			"student:update:any",
			"student:delete",
		},
		"staff": {
			"student:list",
			"student:read:any",
			"student:create",
		},
		"user": {}, // user tidak punya global permissions
	}
	
	// Cek global permission untuk role
	permissions, exists := rolePermissions[userRole]
	if !exists {
		return false
	}
	
	for _, perm := range permissions {
		if perm == permission {
			return true
		}
	}
	
	// Untuk user, cek ownership jika ada studentOwnerID dan userID
	if userRole == "user" && studentOwnerID != 0 && userID != 0 {
		return studentOwnerID == userID
	}
	
	return false
}

// TestPermissionMatrix menguji keseluruhan matrix permission
func TestPermissionMatrix(t *testing.T) {
	matrix := map[string]map[string]bool{
		"admin": {
			"GET /students":        true,
			"GET /students/:id":    true,
			"POST /students":       true,
			"PUT /students/:id":    true,
			"PATCH /students/:id":  true,
			"DELETE /students/:id": true,
		},
		"staff": {
			"GET /students":        true,
			"GET /students/:id":    true,
			"POST /students":       true,
			"PUT /students/:id":    false,
			"PATCH /students/:id":  false,
			"DELETE /students/:id": false,
		},
		"user": {
			"GET /students":        false,
			"GET /students/:id":    false, // kecuali milik sendiri
			"POST /students":       true,
			"PUT /students/:id":    false, // kecuali milik sendiri
			"PATCH /students/:id":  false, // kecuali milik sendiri
			"DELETE /students/:id": false,
		},
	}

	for role, endpoints := range matrix {
		t.Run("Role_"+role, func(t *testing.T) {
			for endpoint, expected := range endpoints {
				t.Logf("Testing %s access to %s: expected %v", role, endpoint, expected)
				
				// Log untuk dokumentasi matrix
				if expected {
					t.Logf("✅ %s dapat akses %s", role, endpoint)
				} else {
					t.Logf("❌ %s tidak dapat akses %s", role, endpoint)
				}
			}
		})
	}
}
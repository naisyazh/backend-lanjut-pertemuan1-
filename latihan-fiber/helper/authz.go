package helper

import (
	"context"
	"database/sql"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// AuthzChecker memeriksa authorization permissions
type AuthzChecker struct {
	pool *pgxpool.Pool
}

func NewAuthzChecker(pool *pgxpool.Pool) *AuthzChecker {
	return &AuthzChecker{pool: pool}
}

// HasPermission mengecek apakah user dengan role tertentu memiliki permission
func (a *AuthzChecker) HasPermission(ctx context.Context, role, permission string) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*) 
		FROM role_permissions rp 
		JOIN permissions p ON rp.permission_name = p.name 
		WHERE rp.role_name = $1 AND p.name = $2
	`
	
	err := a.pool.QueryRow(ctx, query, role, permission).Scan(&count)
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}

// CanAccessStudent mengecek apakah user bisa mengakses student tertentu
func (a *AuthzChecker) CanAccessStudent(ctx context.Context, userID int, userRole string, studentID int, permission string) (bool, error) {
	// Admin dan staff dengan permission bisa akses semua
	if userRole == "admin" || (userRole == "staff" && strings.Contains(permission, "read")) {
		hasGlobalPerm, err := a.HasPermission(ctx, userRole, permission)
		if err != nil {
			return false, err
		}
		if hasGlobalPerm {
			return true, nil
		}
	}
	
	// Untuk user biasa, cek ownership
	if userRole == "user" {
		var ownerID int
		query := `SELECT owner_id FROM students WHERE id = $1`
		err := a.pool.QueryRow(ctx, query, studentID).Scan(&ownerID)
		if err != nil {
			if err == sql.ErrNoRows {
				return false, nil // Student tidak ditemukan
			}
			return false, err
		}
		
		// User hanya bisa akses data miliknya sendiri
		return ownerID == userID, nil
	}
	
	return false, nil
}
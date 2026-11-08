package repository

import (
	"context"
)

type RoleRepository interface {
	LoadPermissions(ctx context.Context) (map[string][]string, error)
}

type rolePostgresRepository struct {
	db DBTX
}

func NewRoleRepository(db DBTX) RoleRepository {
	return &rolePostgresRepository{db: db}
}

func (r *rolePostgresRepository) LoadPermissions(ctx context.Context) (map[string][]string, error) {
	rows, err := r.db.Query(ctx, "SELECT role_name, permission_name FROM role_permissions")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string][]string)
	for rows.Next() {
		var role, perm string
		if err := rows.Scan(&role, &perm); err != nil {
			return nil, err
		}
		result[role] = append(result[role], perm)
	}
	return result, rows.Err()
}

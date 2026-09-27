package helper

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"latihan-fiber/app/model"
)

// RequestContext memberi timeout untuk setiap operasi database.
func RequestContext(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.UserContext(), 5*time.Second)
}

// RequestID mengambil request ID dari context atau generate yang baru.
func RequestID(c *fiber.Ctx) string {
	if id := c.Locals("requestid"); id != nil {
		if str, ok := id.(string); ok {
			return str
		}
	}
	// Generate new UUID jika belum ada
	return uuid.New().String()
}

// ParamID membaca parameter :id dari jalur dan memastikan bentuknya benar.
func ParamID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}

var allowedSort = map[string]bool{
	"id": true, "username": true, "email": true, "created_at": true,
	"nim": true, "name": true, "grade": true,
}

// ParseListQuery membaca query string dan memberi nilai bawaan yang aman.
func ParseListQuery(c *fiber.Ctx) model.ListQuery {
	q := model.ListQuery{
		Page:   c.QueryInt("page", 1),
		Limit:  c.QueryInt("limit", 10),
		Search: strings.TrimSpace(c.Query("search")),
		Sort:   c.Query("sort", "id"),
		Order:  strings.ToLower(c.Query("order", "asc")),
	}

	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 {
		q.Limit = 10
	}
	if q.Limit > 100 {
		q.Limit = 100
	}
	if !allowedSort[q.Sort] {
		q.Sort = "id"
	}
	if q.Order != "desc" {
		q.Order = "asc"
	}

	if raw := c.Query("is_active"); raw != "" {
		if v, err := strconv.ParseBool(raw); err == nil {
			q.IsActive = &v
		}
	}

	return q
}

// LocalsAuthUser adalah key untuk menyimpan user yang sedang login di context.
const LocalsAuthUser = "auth_user"

// CurrentUser mengambil user yang sedang login dari context.
func CurrentUser(c *fiber.Ctx) (*model.User, bool) {
	user := c.Locals(LocalsAuthUser)
	if user == nil {
		return nil, false
	}
	if u, ok := user.(*model.User); ok {
		return u, true
	}
	return nil, false
}

// ParseCursorQuery membaca parameter query untuk cursor pagination.
func ParseCursorQuery(c *fiber.Ctx) (model.CursorQuery, error) {
	q := model.CursorQuery{
		Limit:  c.QueryInt("limit", 10),
		Search: strings.TrimSpace(c.Query("search")),
	}

	if q.Limit < 1 {
		q.Limit = 10
	}
	if q.Limit > 100 {
		q.Limit = 100
	}

	if raw := c.Query("is_active"); raw != "" {
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return model.CursorQuery{}, BadRequest("nilai is_active tidak valid, gunakan true atau false")
		}
		q.IsActive = &v
	}

	if encoded := strings.TrimSpace(c.Query("cursor")); encoded != "" {
		cur, err := DecodeCursor(encoded)
		if err != nil {
			return model.CursorQuery{}, BadRequest("cursor tidak valid")
		}
		q.After = &cur
	}

	return q, nil
}

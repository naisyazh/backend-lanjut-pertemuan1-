package middleware

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"

	"latihan-fiber/app/model"
	"latihan-fiber/helper"
)

func RequireAuth(jwtManager *helper.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token, err := bearerToken(c)
		if err != nil {
			// WWW-Authenticate adalah header baku yang menyertai 401.
			c.Set("WWW-Authenticate", `Bearer realm="api"`)
			return helper.Unauthorized("header Authorization tidak ada atau salah bentuk")
		}

		authUser, err := jwtManager.Parse(token)
		if err != nil {
			c.Set("WWW-Authenticate", `Bearer realm="api"`)
			if errors.Is(err, helper.ErrExpiredToken) {
				return helper.Unauthorized("access token kedaluwarsa")
			}
			return helper.Unauthorized("access token tidak valid")
		}

		c.Locals(helper.LocalsAuthUser, &authUser)
		return c.Next()
	}
}

func RequirePermission(authzChecker *helper.AuthzChecker, permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authUser, ok := c.Locals(helper.LocalsAuthUser).(*model.AuthUser)
		if !ok {
			return helper.Unauthorized("user tidak terautentikasi")
		}

		ctx, cancel := helper.RequestContext(c)
		defer cancel()

		hasPermission, err := authzChecker.HasPermission(ctx, authUser.Role, permission)
		if err != nil {
			return helper.Internal(err)
		}

		if !hasPermission && authUser.Role != "user" {
			return helper.Forbidden("tidak memiliki permission: " + permission)
		}

		return c.Next()
	}
}

func bearerToken(c *fiber.Ctx) (string, error) {
	header := c.Get(fiber.HeaderAuthorization)
	if header == "" {
		return "", errors.New("header kosong")
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("format bukan Bearer")
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", errors.New("token kosong")
	}

	return token, nil
}

func LoginRateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        5,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			c.Set("Retry-After", "60")
			return helper.TooManyRequests("terlalu banyak percobaan login, coba lagi dalam satu menit")
		},
	})
}

package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"latihan-fiber/app/service"
	"latihan-fiber/helper"
	"latihan-fiber/middleware"
)

// Register memetakan URL ke method pada service.
//
// Perhatikan isi file ini: tidak ada logika bisnis, tidak ada query,
// tidak ada validasi. Hanya daftar alamat dan siapa yang melayaninya.
func Register(app *fiber.App, pool *pgxpool.Pool,
	userService *service.UserService, studentService *service.StudentService, nilaiService *service.NilaiService,
	authService *service.AuthService, jwtManager *helper.JWTManager) {

	api := app.Group("/api/v1")

	api.Get("/health", healthCheck(pool))

	// Authentication endpoints - public (tidak butuh auth)
	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/register", authService.Register)
	auth.Post("/login", middleware.LoginRateLimiter(), authService.Login)
	auth.Post("/refresh", authService.Refresh)
	auth.Post("/logout", authService.Logout)
	auth.Get("/me", middleware.RequireAuth(jwtManager), authService.Me)

	users := api.Group("/users", middleware.RequireJSON)
	users.Get("/", userService.List)
	users.Get("/:id", userService.Get)
	users.Post("/", userService.Create)
	users.Put("/:id", userService.Replace)
	users.Patch("/:id", userService.Patch)
	users.Delete("/:id", userService.Delete)

	// Students endpoints - PROTECTED (butuh authentication)
	students := api.Group("/students", middleware.RequireJSON, middleware.RequireAuth(jwtManager))
	students.Get("/", studentService.List)
	students.Get("/:id", studentService.Get)
	students.Post("/", studentService.Create)
	students.Put("/:id", studentService.Replace)
	students.Patch("/:id", studentService.Patch)
	students.Delete("/:id", studentService.Delete)

	// Nilai endpoints - PROTECTED (butuh authentication)
	nilai := api.Group("/nilai", middleware.RequireJSON, middleware.RequireAuth(jwtManager))
	nilai.Get("/", nilaiService.GetAllNilai_Handler)                    // GET /api/v1/nilai
	nilai.Post("/", nilaiService.CreateNilai_Handler)                  // POST /api/v1/nilai
	nilai.Get("/nim/:nim", nilaiService.GetNilaiByNIM_Handler)         // GET /api/v1/nilai/nim/123456
	nilai.Get("/statistik/:nim", nilaiService.GetStatistikNilai_Handler) // GET /api/v1/nilai/statistik/123456
	nilai.Get("/matkul", nilaiService.GetNilaiByMatkul_Handler)        // GET /api/v1/nilai/matkul?nama=praktikum
}

// healthCheck melaporkan kondisi layanan beserta databasenya.
func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.Fail(c, fiber.StatusServiceUnavailable,
				"database tidak dapat dihubungi")
		}

		return helper.Success(c, fiber.StatusOK, "server dan database berjalan", nil)
	}
}

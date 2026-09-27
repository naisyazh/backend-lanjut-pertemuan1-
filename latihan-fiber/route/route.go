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

func Register(app *fiber.App, pool *pgxpool.Pool,
	userService *service.UserService, studentService *service.StudentService, nilaiService *service.NilaiService,
	authService *service.AuthService, jwtManager *helper.JWTManager) {

	// Buat authz checker
	authzChecker := helper.NewAuthzChecker(pool)

	api := app.Group("/api/v1")

	api.Get("/health", healthCheck(pool))
	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/register", authService.Register)
	auth.Post("/login", middleware.LoginRateLimiter(), authService.Login)
	auth.Post("/refresh", authService.Refresh)
	auth.Post("/logout", authService.Logout)
	auth.Get("/me", middleware.RequireAuth(jwtManager), authService.Me)

	users := api.Group("/users", middleware.RequireJSON, middleware.RequireAuth(jwtManager))
	users.Get("/", userService.List)
	users.Get("/:id", userService.Get)
	users.Post("/", userService.Create)
	users.Put("/:id", userService.Replace)
	users.Patch("/:id", userService.Patch)
	users.Delete("/:id", userService.Delete)

	// Students dengan RBAC authorization
	students := api.Group("/students", middleware.RequireJSON, middleware.RequireAuth(jwtManager))
	students.Get("/", middleware.RequirePermission(authzChecker, "student:list"), studentService.List)
	students.Get("/:id", studentService.Get) // Authorization dicek di service layer untuk ownership
	students.Post("/", middleware.RequirePermission(authzChecker, "student:create"), studentService.Create)
	students.Put("/:id", studentService.Replace) // Authorization dicek di service layer untuk ownership
	students.Patch("/:id", studentService.Patch) // Authorization dicek di service layer untuk ownership
	students.Delete("/:id", middleware.RequirePermission(authzChecker, "student:delete"), studentService.Delete)

	nilai := api.Group("/nilai", middleware.RequireJSON, middleware.RequireAuth(jwtManager))
	nilai.Get("/", nilaiService.GetAllNilai_Handler)                    
	nilai.Post("/", nilaiService.CreateNilai_Handler)                  
	nilai.Get("/nim/:nim", nilaiService.GetNilaiByNIM_Handler)         
	nilai.Get("/statistik/:nim", nilaiService.GetStatistikNilai_Handler) 
	nilai.Get("/matkul", nilaiService.GetNilaiByMatkul_Handler)       
}

func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.Internal(err)
		}

		return helper.Success(c, fiber.StatusOK, "server dan database berjalan", nil)
	}
}

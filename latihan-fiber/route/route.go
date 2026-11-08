package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	swagger "github.com/gofiber/swagger"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"latihan-fiber/app/service"
	"latihan-fiber/helper"
	"latihan-fiber/middleware"
)

// Dependencies mengumpulkan seluruh ketergantungan yang dibutuhkan route.
type Dependencies struct {
	Pool                *pgxpool.Pool
	Mongo               *mongo.Client
	JWT                 *helper.JWTManager
	Permissions         helper.PermissionChecker
	UserService         *service.UserService
	StudentService      *service.StudentService
	NilaiService        *service.NilaiService
	AuthService         *service.AuthService
	CourseService       *service.CourseService
	SupervisorService   *service.SupervisorService
	ReviewService       *service.ReviewService
	PortfolioService    *service.PortfolioService
	NotificationService *service.NotificationService
	EnableSwagger       bool
}

func Register(app *fiber.App, deps Dependencies) {
	if deps.EnableSwagger {
		// Swagger UI: http://localhost:3000/swagger/index.html
		app.Get("/swagger/*", swagger.HandlerDefault)
	}

	perms := deps.Permissions

	api := app.Group("/api/v1")

	api.Get("/health", healthCheck(deps.Pool, deps.Mongo))

	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/register", deps.AuthService.Register)
	auth.Post("/login", middleware.LoginRateLimiter(), deps.AuthService.Login)
	auth.Post("/refresh", deps.AuthService.Refresh)
	auth.Post("/logout", deps.AuthService.Logout)
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)

	users := api.Group("/users", middleware.RequireJSON, middleware.RequireAuth(deps.JWT))
	users.Get("/", deps.UserService.List)
	users.Get("/:id", deps.UserService.Get)
	users.Post("/", deps.UserService.Create)
	users.Put("/:id", deps.UserService.Replace)
	users.Patch("/:id", deps.UserService.Patch)
	users.Delete("/:id", deps.UserService.Delete)

	// Students dengan RBAC authorization
	students := api.Group("/students", middleware.RequireJSON, middleware.RequireAuth(deps.JWT))
	students.Get("/", middleware.RequirePermission(perms, "student:list"), deps.StudentService.List)
	students.Get("/:id", deps.StudentService.Get)
	students.Post("/", middleware.RequirePermission(perms, "student:create"), deps.StudentService.Create)
	students.Put("/:id", deps.StudentService.Replace)
	students.Patch("/:id", deps.StudentService.Patch)
	students.Delete("/:id", middleware.RequirePermission(perms, "student:delete"), deps.StudentService.Delete)

	nilai := api.Group("/nilai", middleware.RequireJSON, middleware.RequireAuth(deps.JWT))
	nilai.Get("/", deps.NilaiService.GetAllNilai_Handler)
	nilai.Post("/", deps.NilaiService.CreateNilai_Handler)
	nilai.Get("/nim/:nim", deps.NilaiService.GetNilaiByNIM_Handler)
	nilai.Get("/statistik/:nim", deps.NilaiService.GetStatistikNilai_Handler)
	nilai.Get("/matkul", deps.NilaiService.GetNilaiByMatkul_Handler)

	// --- kelas dan KRS ---
	courses := api.Group("/courses",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT))

	// "/mine" didaftarkan SEBELUM "/:id". Fiber mencocokkan route sesuai
	// urutan pendaftaran; bila dibalik, "mine" akan dianggap sebagai :id.
	courses.Get("/", deps.CourseService.List)
	courses.Get("/mine", deps.CourseService.Mine)
	courses.Get("/:id", deps.CourseService.Get)
	courses.Post("/:id/enroll", deps.CourseService.Enroll)
	courses.Delete("/:id/enroll", deps.CourseService.Drop)
	courses.Patch("/:id", middleware.RequirePermission(perms, "course:manage"), deps.CourseService.UpdateQuota)
	// endpoint ulasan kelas
	courses.Get("/:id/reviews/summary", deps.ReviewService.Summary)
	courses.Get("/:id/reviews", deps.ReviewService.ListByCourse)
	courses.Post("/:id/reviews", deps.ReviewService.Create)

	enrollments := api.Group("/enrollments",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT))
	enrollments.Post("/swap", deps.CourseService.Swap)

	// --- pembimbing tugas akhir ---
	supervisors := api.Group("/supervisors",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT))
	supervisors.Get("/", deps.SupervisorService.List)
	supervisors.Get("/:id", deps.SupervisorService.Get)
	supervisors.Post("/:id/assign",
		middleware.RequirePermission(perms, "supervision:manage"),
		deps.SupervisorService.Assign)
	supervisors.Patch("/:id",
		middleware.RequirePermission(perms, "supervision:manage"),
		deps.SupervisorService.UpdateQuota)

	supervisions := api.Group("/supervisions",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT))
	supervisions.Post("/switch",
		middleware.RequirePermission(perms, "supervision:manage"),
		deps.SupervisorService.Switch)

	// --- ulasan kelas (MongoDB) ---
	reviews := api.Group("/reviews",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT))
	reviews.Get("/:id", deps.ReviewService.Get)
	reviews.Patch("/:id", deps.ReviewService.Update)
	reviews.Delete("/:id", deps.ReviewService.Delete)
	reviews.Post("/:id/replies", deps.ReviewService.Reply)
	reviews.Post("/:id/helpful", deps.ReviewService.MarkHelpful)

	// --- portofolio mahasiswa (MongoDB) ---
	students2 := api.Group("/students",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT))
	students2.Get("/:id/portfolio", deps.PortfolioService.Get)
	students2.Put("/:id/portfolio", deps.PortfolioService.Upsert)
	students2.Post("/:id/portfolio/projects", deps.PortfolioService.AddProject)
	students2.Delete("/:id/portfolio/projects/:projectId", deps.PortfolioService.DeleteProject)

	portfolios := api.Group("/portfolios",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT))
	portfolios.Get("/stats/technologies", deps.PortfolioService.TopTechnologies)

	// --- notifikasi (diisi oleh worker, dibaca lewat API) ---
	notifications := api.Group("/notifications",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT))
	notifications.Get("/mine", deps.NotificationService.Mine)

	admin := api.Group("/admin",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT))
	admin.Get("/queue",
		middleware.RequirePermission(perms, "queue:monitor"),
		deps.NotificationService.Monitor)
}

func healthCheck(pool *pgxpool.Pool, mongoClient *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.ServiceUnavailable("PostgreSQL tidak dapat dihubungi")
		}
		if mongoClient != nil {
			if err := mongoClient.Ping(ctx, nil); err != nil {
				return helper.ServiceUnavailable("MongoDB tidak dapat dihubungi")
			}
		}
		return helper.Success(c, fiber.StatusOK, "server, PostgreSQL, dan MongoDB berjalan", nil)
	}
}

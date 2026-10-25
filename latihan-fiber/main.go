package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"latihan-fiber/app/repository"
	"latihan-fiber/app/service"
	"latihan-fiber/config"
	"latihan-fiber/database"
	"latihan-fiber/helper"
	"latihan-fiber/route"
)

// main hanya berisi urutan perakitan. Tidak ada logika bisnis,
// tidak ada query, dan tidak ada satu pun handler di sini.
func main() {
	// 1. Konfigurasi dan logger
	config.LoadEnv()
	logger := config.NewLogger()

	// 2. PostgreSQL
	pool, err := database.NewPool(context.Background())
	if err != nil {
		logger.Error("gagal terhubung ke database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	// 3. MongoDB
	mongoClient, err := database.NewMongoClient(context.Background())
	if err != nil {
		logger.Error("gagal terhubung ke MongoDB", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = mongoClient.Disconnect(ctx)
	}()
	mongoDB := mongoClient.Database(config.GetEnv("MONGO_DATABASE", "praktikum_backend"))

	// 4. Perakitan repository
	userRepository := repository.NewUserRepository(pool)
	userService := service.NewUserService(userRepository)

	authzChecker := helper.NewAuthzChecker(pool)

	studentRepository := repository.NewStudentRepository(pool)
	studentService := service.NewStudentService(studentRepository, authzChecker)

	nilaiRepository := repository.NewNilaiRepository(pool)
	nilaiService := service.NewNilaiService(nilaiRepository)

	// 5. Auth service dan JWT manager
	jwtManager := helper.NewJWTManager(
		config.GetEnv("JWT_SECRET", ""),
		config.GetEnv("JWT_ISSUER", "praktikum-backend"),
		time.Duration(config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 15))*time.Minute,
	)
	tokenRepository := repository.NewTokenRepository(pool)
	authService := service.NewAuthService(
		userRepository,
		tokenRepository,
		jwtManager,
		7*24*time.Hour,
	)

	// 6. Course service
	courseRepository := repository.NewCourseRepository(pool)
	txManager := repository.NewTxManager(pool)
	courseService := service.NewCourseService(courseRepository, txManager)

	// 7. Supervisor service
	supervisorRepository := repository.NewSupervisorRepository(pool)
	supervisorService := service.NewSupervisorService(supervisorRepository, txManager)

	// 8. Review service (MongoDB)
	reviewRepository := repository.NewReviewRepository(mongoDB)
	if err := reviewRepository.EnsureIndexes(context.Background()); err != nil {
		logger.Error("gagal membuat index MongoDB", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Permission set untuk review:moderate (dibaca sekali saat startup)
	permSet := helper.NewPermissionSet(map[string][]string{
		"admin": {"review:moderate", "course:manage", "supervision:manage"},
		"user":  {},
	})

	reviewService := service.NewReviewService(reviewRepository, courseRepository, permSet)

	// 9. Portfolio service (Tugas Mandiri C)
	portfolioRepository := repository.NewPortfolioRepository(mongoDB)
	if err := portfolioRepository.EnsureIndexes(context.Background()); err != nil {
		logger.Error("gagal membuat index portfolio MongoDB", slog.String("error", err.Error()))
		os.Exit(1)
	}
	portfolioService := service.NewPortfolioService(portfolioRepository, studentRepository)

	// 9. Aplikasi
	app := config.NewApp(logger, route.Dependencies{
		Pool:              pool,
		Mongo:             mongoClient,
		JWT:               jwtManager,
		Permissions:       authzChecker,
		UserService:       userService,
		StudentService:    studentService,
		NilaiService:      nilaiService,
		AuthService:       authService,
		CourseService:     courseService,
		SupervisorService: supervisorService,
		ReviewService:     reviewService,
		PortfolioService:  portfolioService,
	})

	port := config.GetEnv("APP_PORT", "3002")

	go func() {
		if err := app.Listen(":" + port); err != nil {
			logger.Error("server berhenti", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	logger.Info("server berjalan", slog.String("port", port))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("sinyal berhenti diterima, menutup server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		logger.Error("gagal menutup server dengan rapi", slog.String("error", err.Error()))
	}

	logger.Info("server berhenti dengan rapi")
}

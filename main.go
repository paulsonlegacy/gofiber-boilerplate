package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"boilerplate/internal/config"
	"boilerplate/internal/domain/entity"
	"boilerplate/internal/domain/repo"
	"boilerplate/internal/domain/usecase"
	"boilerplate/internal/http/handler"
	"boilerplate/internal/http/middleware"
	"boilerplate/internal/infra/cache"
	"boilerplate/internal/infra/db"
	"boilerplate/internal/infra/logger"
	"boilerplate/internal/infra/mail"
	"boilerplate/internal/infra/queue"
	"boilerplate/internal/infra/storage"
	"boilerplate/internal/infra/telegram"
	"boilerplate/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	BASE_PATH string
	ENV_PATH  string
	Validator = validator.New()
)

// init resolves the project root and .env path before main runs.
func init() {
	var basePath string

	if wd, err := os.Getwd(); err == nil {
		basePath = wd
	}

	if _, err := os.Stat(filepath.Join(basePath, ".env")); os.IsNotExist(err) {
		_, filename, _, ok := runtime.Caller(0)
		if ok {
			basePath = filepath.Dir(filepath.Dir(filename))
		}
	}

	BASE_PATH = basePath
	ENV_PATH = filepath.Join(basePath, ".env")
}

func main() {
	// ── 1. Load config ─────────────────────────────────────────────────────
	cfg := config.InitConfig(ENV_PATH)

	// ── 2. Init logger ─────────────────────────────────────────────────────
	log := logger.InitLogger(cfg.LoggingConfig)
	defer log.Sync()

	log.Info("Starting application",
		zap.String("env", cfg.LoggingConfig.EnvType),
		zap.String("port", cfg.ServerConfig.ServerPort),
	)

	// ── 3. Init database ───────────────────────────────────────────────────
	db := initDB(cfg)

	// ── 4. Init infrastructure ─────────────────────────────────────────────
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobQueue := queue.NewDurableQueue(ctx, db, 200)

	primarySMTP := mail.NewSMTPMailer(cfg.EmailConfig.Primary)
	multiMailer := mail.NewMultiProviderMailer(primarySMTP)

	storageProvider := storage.NewCloudinaryStorage(cfg.StorageConfig.CloudinaryConfig)

	telegramSvc := telegram.NewTelegramService(
		cfg.TelegramConfig.Token,
		cfg.TelegramConfig.ChatID,
	)

	memCache := cache.NewMemoryCache()

	// ── 5. Init service layer ──────────────────────────────────────────────
	svc := service.NewService(
		ctx,
		cfg,
		db,
		log,
		multiMailer,
		jobQueue,
		storageProvider,
		telegramSvc,
		memCache,
	)

	// ── 6. Init repositories ───────────────────────────────────────────────
	userRepo := repo.NewUserRepository(db)
	adminRepo := repo.NewAdminRepository(db)
	walletRepo := repo.NewWalletRepository(db)
	txRepo := repo.NewTransactionRepository(db)

	// ── 7. Init usecases ───────────────────────────────────────────────────
	authUC := usecase.NewAuthUsecase(svc, userRepo, adminRepo)
	userUC := usecase.NewUserUsecase(svc, userRepo, walletRepo, txRepo)
	adminUC := usecase.NewAdminUsecase(svc, adminRepo)
	walletUC := usecase.NewWalletUsecase(svc, walletRepo, txRepo, userRepo)

	// ── 8. Init handlers ───────────────────────────────────────────────────
	authHandler := handler.NewAuthHandler(authUC, userUC, svc, Validator)
	userHandler := handler.NewUserHandler(userUC, svc, Validator)
	adminHandler := handler.NewAdminHandler(adminUC, userUC, svc, Validator)
	walletHandler := handler.NewWalletHandler(walletUC, svc, Validator)
	systemHandler := handler.NewSystemHandler(svc, db)

	// ── 9. Init Fiber app ──────────────────────────────────────────────────
	app := fiber.New(fiber.Config{
		AppName:      cfg.SiteConfig.SiteName,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			log.Error("Unhandled error", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "An unexpected error occurred",
				"error":   fiber.Map{"code": "INTERNAL_SERVER_ERROR", "message": err.Error()},
			})
		},
	})

	// ── 10. Global middleware ──────────────────────────────────────────────
	app.Use(helmet.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: false,
	}))
	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
	}))
	app.Use(middleware.SanitizeInputs())

	// ── 11. Auth middleware ────────────────────────────────────────────────
	authMW := middleware.NewAuthMiddleware(cfg.ServerConfig, cfg.JWTConfig, db)
	app.Use(authMW.Handler())

	// ── 12. Register routes ────────────────────────────────────────────────
	registerRoutes(app, authHandler, userHandler, adminHandler, walletHandler, systemHandler)

	// ── 13. Start background workers & cron ───────────────────────────────
	svc.StartBackgroundWorkers(jobQueue, 5)
	svc.StartCronJobs()

	// ── 14. Start server ───────────────────────────────────────────────────
	addr := fmt.Sprintf("%s:%s", cfg.ServerConfig.ServerAddress, cfg.ServerConfig.ServerPort)
	go func() {
		log.Info("Server listening", zap.String("address", addr))
		if err := app.Listen(addr); err != nil {
			log.Fatal("Server error", zap.Error(err))
		}
	}()

	// ── 15. Graceful shutdown ──────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info("Shutdown signal received — draining connections...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Error("Graceful shutdown failed", zap.Error(err))
	}

	log.Info("Server stopped cleanly")
}

// initDB returns a *gorm.DB — postgres for production (DB_URL set), sqlite for development.
func initDB(cfg *config.Config) *gorm.DB {
	if !cfg.ServerConfig.Debug && cfg.DBConfig.DBURL != "" {
		return db.NewPostgresDB(cfg.DBConfig)
	}
	return db.NewSQLiteDB("database.db")
}

// registerRoutes wires all handlers to their route groups.
// Extend this function as your project grows — keep it the single source of truth for routing.
func registerRoutes(
	app *fiber.App,
	auth *handler.AuthHandler,
	user *handler.UserHandler,
	admin *handler.AdminHandler,
	wallet *handler.WalletHandler,
	system *handler.SystemHandler,
) {
	// ── Health & metrics ───────────────────────────────────────────────────
	app.Get("/health", system.Health)
	app.Get("/metrics", system.Metrics)

	api := app.Group("/api")

	// ── Public auth routes ─────────────────────────────────────────────────
	authGroup := api.Group("/auth")
	authGroup.Post("/register", auth.Register)
	authGroup.Post("/login", auth.Login)
	authGroup.Post("/refresh-token", auth.RefreshToken)
	authGroup.Post("/forgot-password", auth.ForgotPassword)
	authGroup.Post("/reset-password", auth.ResetPassword)

	// ── User routes (authenticated, role: user) ────────────────────────────
	userGroup := api.Group("/user", middleware.ActorTypeRequired("user"))
	userGroup.Get("/profile", user.GetProfile)
	userGroup.Put("/profile", user.UpdateProfile)
	userGroup.Post("/logout", auth.Logout)

	// ── Wallet routes (user) ───────────────────────────────────────────────
	walletGroup := api.Group("/wallet", middleware.ActorTypeRequired("user"))
	walletGroup.Get("/", wallet.GetWallet)
	walletGroup.Get("/transactions", wallet.ListTransactions)

	// ── Admin routes (authenticated, role: admin) ──────────────────────────
	adminGroup := api.Group("/admin", middleware.ActorTypeRequired("admin"))

	adminGroup.Get("/users", admin.ListUsers)
	adminGroup.Get("/users/:id", admin.GetUser)
	adminGroup.Put("/users/:id", admin.UpdateUser)
	adminGroup.Post("/users/:id/suspend", admin.SuspendUser)
	adminGroup.Post("/users/:id/activate", admin.ActivateUser)
	adminGroup.Delete("/users/:id", admin.DeleteUser)

	// ── Superadmin routes ──────────────────────────────────────────────────
	superAdminGroup := api.Group("/superadmin", middleware.SuperAdminOnly())
	superAdminGroup.Get("/admins", admin.ListAdmins)
	superAdminGroup.Post("/admins", admin.CreateAdmin)
	superAdminGroup.Put("/admins/:id", admin.UpdateAdmin)
	superAdminGroup.Delete("/admins/:id", admin.DeleteAdmin)

	_ = entity.User{} // ensure entity package is imported
}

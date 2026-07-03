package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/6sLOGAN78/go-protask/internal/config"
	"github.com/6sLOGAN78/go-protask/internal/database"
	"github.com/6sLOGAN78/go-protask/internal/handler"
	"github.com/6sLOGAN78/go-protask/internal/logger"
	"github.com/6sLOGAN78/go-protask/internal/repository"
	"github.com/6sLOGAN78/go-protask/internal/router"
	"github.com/6sLOGAN78/go-protask/internal/server"
	"github.com/6sLOGAN78/go-protask/internal/service"
)

// During graceful shutdown we allow up to 30 seconds
// for ongoing requests, database operations, etc. to finish.
const DefaultContextTimeout = 30

func main() {

	// ============================================================
	// 1. LOAD APPLICATION CONFIGURATION
	// ============================================================
	//
	// Reads configuration from environment variables, config files,
	// secrets manager, etc.
	//
	// Example:
	// ENV=prod
	// DB_HOST=localhost
	// PORT=8080
	//
	// The entire application depends on this configuration.
	//
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	// ============================================================
	// 2. INITIALIZE OBSERVABILITY / LOGGER SERVICE
	// ============================================================
	//
	// Creates the underlying logging backend.
	loggerService := logger.NewLoggerService(cfg.Observability)

	// IMPORTANT:
	//
	// defer DOES NOT run here.
	//
	// It means:
	// "When main() exits, call Shutdown()"
	//
	// Why?
	//
	// Logging systems often buffer logs/traces in memory.
	//
	// Shutdown() gives them a chance to:
	// - flush remaining logs
	// - send remaining traces
	// - close network connections
	// - release resources
	//
	// Without this, some logs may never reach New Relic.
	//
	defer loggerService.Shutdown()

	// Application logger used throughout the codebase.
	log := logger.NewLoggerWithService(
		cfg.Observability,
		loggerService,
	)

	// ============================================================
	// 3. DATABASE MIGRATIONS
	// ============================================================
	//
	// Migrations update the DATABASE SCHEMA.
	//
	// Example:
	//
	// Version 1:
	// users(id, name)
	//
	// Version 2:
	// users(id, name, email)
	//
	// Migration:
	//
	// ALTER TABLE users ADD COLUMN email TEXT;
	//
	// IMPORTANT:
	// Migrations do NOT usually create a brand-new database.
	//
	// They modify an existing database safely while keeping data.
	//
	// Existing rows:
	//
	// id | name
	// 1  | Ayush
	//
	// become:
	//
	// id | name | email
	// 1  | Ayush | NULL
	//
	// ------------------------------------------------------------
	// WHY SKIP LOCAL?
	// ------------------------------------------------------------
	//
	// If ENV=local:
	//
	//     cfg.Primary.Env == "local"
	//
	// migrations are skipped.
	//
	// Possible reasons:
	// - developer runs migrations manually
	// - local DB managed separately
	// - avoid changing schema every startup
	//
	// ------------------------------------------------------------
	// PRODUCTION FLOW
	// ------------------------------------------------------------
	//
	// Developer:
	//
	// Add migration file
	//      ↓
	// git push
	//      ↓
	// CI/CD pipeline
	//      ↓
	// Deploy new version
	//      ↓
	// Run migrations
	//      ↓
	// Start application
	//
	// Existing production data remains intact.
	//
	if cfg.Primary.Env != "local" {
		if err := database.Migrate(
			context.Background(),
			&log,
			cfg,
		); err != nil {
			log.Fatal().
				Err(err).
				Msg("failed to migrate database")
		}
	}

	// ============================================================
	// 4. CREATE SERVER
	// ============================================================
	//
	// Usually the central dependency container.
	//
	// May hold:
	// - database connection
	// - config
	// - logger
	// - http server
	// - newRel
	//
	srv, err := server.New(
		cfg,
		&log,
		loggerService,
	)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("failed to initialize server")
	}

	// ============================================================
	// 5. CREATE REPOSITORIES
	// ============================================================
	//
	// Repository layer talks directly to database.
	//
	// Examples:
	// - UserRepository
	// - TaskRepository
	// - ProjectRepository
	//
	// SQL queries usually live here.
	//
	repos := repository.NewRepositories(srv)

	// ============================================================
	// 6. CREATE SERVICES
	// ============================================================
	//
	// Business logic layer.
	//
	// Examples:
	// - validation
	// - authorization
	// - business rules
	//
	// Flow:
	//
	// Handler
	//   ↓
	// Service
	//   ↓
	// Repository
	//   ↓
	// Database
	//
	services, serviceErr := service.NewServices(
		srv,
		repos,
	)
	if serviceErr != nil {
		log.Fatal().
			Err(serviceErr).
			Msg("could not create services")
	}

	// ============================================================
	// 7. CREATE HANDLERS
	// ============================================================
	//
	// Handlers receive HTTP requests.
	//
	// Example:
	//
	// POST /tasks
	// GET  /tasks
	//
	// Handler receives request
	//      ↓
	// Calls service
	//      ↓
	// Returns response
	//
	handlers := handler.NewHandlers(
		srv,
		services,
	)

	// ============================================================
	// 8. CREATE ROUTER
	// ============================================================
	//
	// Maps URLs to handlers.
	//
	// Example:
	//
	// POST /tasks -> TaskHandler.Create
	// GET  /tasks -> TaskHandler.List
	//
	r := router.NewRouter(
		srv,
		handlers,
		services,
	)

	// ============================================================
	// 9. SETUP HTTP SERVER
	// ============================================================
	//
	// Creates something similar to:
	//
	// http.Server{
	//     Addr: ":8080",
	//     Handler: r,
	// }
	//
	srv.SetupHTTPServer(r)

	// ============================================================
	// 10. LISTEN FOR CTRL+C / SIGINT
	// ============================================================
	//
	// When user presses Ctrl+C:
	//
	// os.Interrupt
	//
	// ctx.Done() becomes ready.
	//
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
	)

	// ============================================================
	// 11. START SERVER IN BACKGROUND
	// ============================================================
	//
	// srv.Start() blocks forever while serving requests.
	//
	// Therefore it must run in a goroutine.
	//
	// Otherwise:
	//
	// srv.Start()
	//
	// would never allow the code below
	// to wait for shutdown signals.
	//
	go func() {
		if err = srv.Start(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {

			log.Fatal().
				Err(err).
				Msg("failed to start server")
		}
	}()

	// ============================================================
	// 12. WAIT FOR SHUTDOWN SIGNAL
	// ============================================================
	//
	// Application stays here until:
	//
	// Ctrl+C
	// kill -SIGINT
	//
	<-ctx.Done()

	// ============================================================
	// 13. CREATE SHUTDOWN TIMEOUT
	// ============================================================
	//
	// Allow active requests to finish.
	//
	// Example:
	//
	// Request starts
	// Ctrl+C pressed
	// Request completes
	// Server exits
	//
	ctx, cancel := context.WithTimeout(
		context.Background(),
		DefaultContextTimeout*time.Second,
	)

	// ============================================================
	// 14. GRACEFUL SHUTDOWN
	// ============================================================
	//
	// Stops accepting NEW requests.
	//
	// Waits for CURRENT requests to finish.
	//
	// Closes connections safely.
	//
	if err = srv.Shutdown(ctx); err != nil {
		log.Fatal().
			Err(err).
			Msg("server forced to shutdown")
	}

	// Cleanup context resources.
	stop()
	cancel()

	// Final log message.
	log.Info().Msg("server exited properly")
}
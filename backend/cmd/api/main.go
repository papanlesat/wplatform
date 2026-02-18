package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"wplatform/backend/internal/auth"
	"wplatform/backend/internal/backup"
	"wplatform/backend/internal/config"
	"wplatform/backend/internal/docker"
	"wplatform/backend/internal/domain"
	"wplatform/backend/internal/envvar"
	"wplatform/backend/internal/monitor"
	"wplatform/backend/internal/project"
	"wplatform/backend/internal/scheduler"
	"wplatform/backend/internal/ssl"
	"wplatform/backend/internal/validation"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize validator
	validation.Init()

	// Parse JWT expiration
	jwtExpiration, err := time.ParseDuration(cfg.JWTExpiration)
	if err != nil {
		log.Fatalf("Failed to parse JWT expiration: %v", err)
	}

	// Initialize database
	if err := config.InitDB(cfg.DatabaseURL); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer config.CloseDB()

	// Initialize Docker client
	dockerClient, err := docker.NewClient()
	if err != nil {
		log.Fatalf("Failed to initialize Docker client: %v", err)
	}
	defer dockerClient.Close()

	// Initialize Docker managers
	networkManager := docker.NewNetworkManager(dockerClient)
	volumeManager := docker.NewVolumeManager(dockerClient)
	containerManager := docker.NewContainerManager(dockerClient)
	containerManager.SetWordPressImage(cfg.WordPressImage)

	// Initialize domain validator
	domainValidator := domain.NewDNSValidatorWithResolver(cfg.DNSResolver)

	// Initialize auth service and handler
	authRepo := auth.NewUserRepository(config.DB)
	authService := auth.NewAuthService(authRepo, cfg.JWTSecret, jwtExpiration)
	authHandler := auth.NewAuthHandler(authService)

	// Initialize project service and handler
	projectRepo := project.NewProjectRepository(config.DB)
	projectService := project.NewProjectService(
		projectRepo,
		dockerClient,
		networkManager,
		volumeManager,
		containerManager,
		domainValidator,
		cfg.ServerIP,
	)
	projectHandler := project.NewProjectHandler(projectService)

	// Initialize environment variable service and handler
	envVarRepo := envvar.NewEnvironmentVariableRepository(config.DB)
	envVarService := envvar.NewEnvironmentVariableService(envVarRepo, containerManager)
	envVarHandler := envvar.NewEnvironmentVariableHandler(envVarService)

	// Initialize SSL service and handler
	sslRepo := ssl.NewSSLRepository(config.DB)
	sslService := ssl.NewSSLService(sslRepo, containerManager)
	sslHandler := ssl.NewSSLHandler(sslService)

	// Initialize backup service and handler
	backupRepo := backup.NewBackupRepository(config.DB)
	scheduleRepo := backup.NewBackupScheduleRepository(config.DB)
	backupBasePath := cfg.BackupBasePath
	backupService := backup.NewBackupService(backupRepo, scheduleRepo, projectRepo, dockerClient, containerManager, backupBasePath)
	backupHandler := backup.NewBackupHandler(backupService, scheduleRepo, projectRepo)

	if err := os.MkdirAll(backupBasePath, 0755); err != nil {
		log.Fatalf("Failed to create backup directory: %v", err)
	}

	// Initialize backup scheduler
	backupScheduler := scheduler.NewBackupScheduler(backupService, scheduleRepo, projectRepo)
	backupScheduler.Start()
	defer backupScheduler.Stop()

	// Initialize monitor service and handler
	monitorService := monitor.NewMonitorService(dockerClient, containerManager)
	monitorHandler := monitor.NewMonitorHandler(monitorService)

	// Initialize domain handler
	domainHandler := domain.NewDomainHandler(domainValidator)

	// Setup router
	router := mux.NewRouter()

	// Public routes
	authHandler.RegisterRoutes(router)

	// Protected routes (require authentication)
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.Use(auth.JWTMiddleware(cfg.JWTSecret))
	projectHandler.RegisterRoutes(apiRouter)
	envVarHandler.RegisterRoutes(apiRouter)
	sslHandler.RegisterRoutes(apiRouter)
	backupHandler.RegisterRoutes(apiRouter)
	monitorHandler.RegisterRoutes(apiRouter)
	domainHandler.RegisterRoutes(apiRouter)

	// CORS middleware (for development)
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Start server
	srv := &http.Server{
		Handler:      router,
		Addr:         ":" + cfg.ServerPort,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("Starting server on %s", cfg.ServerPort)

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/YubiApp/internal/config"
	"github.com/YubiApp/internal/database"
	"github.com/YubiApp/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Server struct {
	config                *config.Config
	db                    *gorm.DB
	authService           *services.AuthService
	userService           *services.UserService
	roleService           *services.RoleService
	resourceService       *services.ResourceService
	permissionService     *services.PermissionService
	deviceService         *services.DeviceService
	actionService         *services.ActionService
	deviceRegService      *services.DeviceRegistrationService
	sessionService        *services.SessionService
	locationService       *services.LocationService
	userStatusService     *services.UserStatusService
	userActivityService   *services.UserActivityService
	httpServer            *http.Server
}

// New creates a new server instance
func New(cfg *config.Config) *Server {
	// Initialize database
	db, err := initDatabase(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize services
	authService := services.NewAuthService(db, cfg)
	userService := services.NewUserService(db)
	roleService := services.NewRoleService(db)
	resourceService := services.NewResourceService(db)
	permissionService := services.NewPermissionService(db)
	deviceService := services.NewDeviceService(db)
	actionService := services.NewActionService(db)
	deviceRegService := services.NewDeviceRegistrationService(db)
	sessionService := services.NewSessionService(cfg)
	locationService := services.NewLocationService(db)
	userStatusService := services.NewUserStatusService(db)
	userActivityService := services.NewUserActivityService(db)

	// Set Gin mode
	if !cfg.Server.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	// Setup router
	router := setupRouter(authService, userService, roleService, resourceService, permissionService, deviceService, actionService, deviceRegService, sessionService, locationService, userStatusService, userActivityService)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.Timeout,
		WriteTimeout: cfg.Server.Timeout,
		IdleTimeout:  cfg.Server.Timeout * 2,
	}

	return &Server{
		config:                cfg,
		db:                    db,
		authService:           authService,
		userService:           userService,
		roleService:           roleService,
		resourceService:       resourceService,
		permissionService:     permissionService,
		deviceService:         deviceService,
		actionService:         actionService,
		deviceRegService:      deviceRegService,
		sessionService:        sessionService,
		locationService:       locationService,
		userStatusService:     userStatusService,
		userActivityService:   userActivityService,
		httpServer:            httpServer,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	log.Printf("Starting server on %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	// Close session service (Redis connection)
	if s.sessionService != nil {
		if err := s.sessionService.Close(); err != nil {
			log.Printf("Error closing session service: %v", err)
		}
	}
	return s.httpServer.Shutdown(ctx)
}

// initDatabase initializes the database connection
func initDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto migrate database models
	if err := db.AutoMigrate(
		&database.User{},
		&database.Role{},
		&database.Resource{},
		&database.Permission{},
		&database.Action{},
		&database.Device{},
		&database.Session{},
		&database.AuthenticationLog{},
		&database.DeviceRegistration{},
		&database.Location{},
		&database.UserStatus{},
		&database.UserActivityHistory{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
} 
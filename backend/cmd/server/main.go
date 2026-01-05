package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bbp/backend/config"
	"github.com/bbp/backend/pkg/database"
	"github.com/bbp/backend/pkg/jwt"
	"github.com/bbp/backend/internal/handler/http"
	"github.com/bbp/backend/internal/middleware"
	"github.com/bbp/backend/internal/repository/models"
	"github.com/bbp/backend/internal/repository/sqlite"
	"github.com/bbp/backend/internal/usecase/auth"
	"github.com/bbp/backend/internal/usecase/user"
	"github.com/bbp/backend/internal/usecase/veto"
	"github.com/bbp/backend/internal/usecase/map_pool"
	"github.com/bbp/backend/internal/usecase/room"
	"github.com/bbp/backend/internal/handler/websocket"
	ws "github.com/bbp/backend/pkg/websocket"
	"github.com/gin-gonic/gin"
)

func main() {
	// Загружаем конфигурацию
	cfg := config.Load()

	// Инициализируем базу данных
	db, err := database.NewDatabase(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close(db)

	// Выполняем миграции
	if err := database.Migrate(db,
		&models.UserModel{},
		&models.GameModel{},
		&models.MapModel{},
		&models.MapPoolModel{},
		&models.VetoSessionModel{},
		&models.VetoActionModel{},
		&models.RoomModel{},
		&models.RoomParticipantModel{},
	); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Database migrations completed successfully")

	// Настраиваем Gin
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Подключаем middleware
	router.Use(middleware.CORSMiddleware(cfg.CORSOrigin))
	router.Use(middleware.ErrorHandlerMiddleware())
	
	// Rate limiting для всех endpoints
	// В development используем более мягкие лимиты, в production - строгие
	if cfg.Environment == "development" {
		// Для development: 500 запросов в минуту (более мягкие лимиты)
		router.Use(middleware.RateLimitMiddleware(500, time.Minute))
	} else {
		// Для production: 100 запросов в минуту (стандартные лимиты)
		router.Use(middleware.DefaultRateLimitMiddleware())
	}
	
	// Строгий rate limiting для auth endpoints
	// В development используем более мягкие лимиты
	var authRateLimit gin.HandlerFunc
	if cfg.Environment == "development" {
		// Для development: 100 запросов в минуту для auth
		authRateLimit = middleware.RateLimitMiddleware(100, time.Minute)
	} else {
		// Для production: 20 запросов в минуту для auth
		authRateLimit = middleware.StrictRateLimitMiddleware()
	}

	// Базовый health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Server is running",
		})
	})

	// Инициализируем JWT сервис
	jwtService := jwt.NewJWTService(cfg.JWTSecret, cfg.JWTExpiry)

	// Инициализируем репозитории
	userRepo := sqlite.NewUserRepository(db)
	vetoSessionRepo := sqlite.NewVetoSessionRepository(db)
	vetoActionRepo := sqlite.NewVetoActionRepository(db)
	roomRepo := sqlite.NewRoomRepository(db)
	mapRepo := sqlite.NewMapRepository(db)
	mapPoolRepo := sqlite.NewMapPoolRepository(db)
	gameRepo := sqlite.NewGameRepository(db)

	// Инициализируем use cases для авторизации
	registerUseCase := auth.NewRegisterUseCase(userRepo, jwtService)
	loginUseCase := auth.NewLoginUseCase(userRepo, jwtService)
	getCurrentUserUseCase := auth.NewGetCurrentUserUseCase(userRepo)

	// Инициализируем use cases для пользователя
	getProfileUseCase := user.NewGetProfileUseCase(userRepo)
	updateProfileUseCase := user.NewUpdateProfileUseCase(userRepo)
	getSessionsUseCase := user.NewGetSessionsUseCase(vetoSessionRepo)
	getRoomsUseCase := user.NewGetRoomsUseCase(roomRepo)

	// Инициализируем VetoLogicService
	vetoLogicService := veto.NewVetoLogicService()

	// Инициализируем use cases для veto
	createSessionUseCase := veto.NewCreateSessionUseCase(vetoSessionRepo, mapPoolRepo, gameRepo, vetoLogicService)
	getSessionUseCase := veto.NewGetSessionUseCase(vetoSessionRepo)
	getNextActionUseCase := veto.NewGetNextActionUseCase(vetoSessionRepo, mapPoolRepo, vetoLogicService)
	banMapUseCase := veto.NewBanMapUseCase(vetoSessionRepo, vetoActionRepo, mapRepo, mapPoolRepo, vetoLogicService)
	pickMapUseCase := veto.NewPickMapUseCase(vetoSessionRepo, vetoActionRepo, mapRepo, mapPoolRepo, vetoLogicService)
	selectSideUseCase := veto.NewSelectSideUseCase(vetoSessionRepo, vetoActionRepo, mapPoolRepo, vetoLogicService)
	resetSessionUseCase := veto.NewResetSessionUseCase(vetoSessionRepo, vetoActionRepo)
	startSessionUseCase := veto.NewStartSessionUseCase(vetoSessionRepo)

	// Инициализируем use cases для map pools
	getPoolsUseCase := map_pool.NewGetPoolsUseCase(mapPoolRepo, gameRepo)
	getPoolUseCase := map_pool.NewGetPoolUseCase(mapPoolRepo)
	createCustomPoolUseCase := map_pool.NewCreateCustomPoolUseCase(mapPoolRepo, mapRepo, gameRepo)
	deletePoolUseCase := map_pool.NewDeletePoolUseCase(mapPoolRepo)

	// Инициализируем use cases для rooms
	createRoomUseCase := room.NewCreateRoomUseCase(roomRepo, gameRepo, mapPoolRepo)
	getRoomUseCase := room.NewGetRoomUseCase(roomRepo)
	getRoomBySessionUseCase := room.NewGetRoomBySessionUseCase(roomRepo)
	getRoomsListUseCase := room.NewGetRoomsListUseCase(roomRepo)
	joinRoomUseCase := room.NewJoinRoomUseCase(roomRepo)
	leaveRoomUseCase := room.NewLeaveRoomUseCase(roomRepo)
	deleteRoomUseCase := room.NewDeleteRoomUseCase(roomRepo)
	updateRoomUseCase := room.NewUpdateRoomUseCase(roomRepo)

	// Инициализируем handlers
	authHandler := http.NewAuthHandler(registerUseCase, loginUseCase, getCurrentUserUseCase)
	userHandler := http.NewUserHandler(getProfileUseCase, updateProfileUseCase, getSessionsUseCase, getRoomsUseCase)
	// Инициализируем WebSocket manager (нужен для RoomHandler)
	wsManager := ws.NewManager()
	go wsManager.Run()

	vetoHandler := http.NewVetoHandler(createSessionUseCase, getSessionUseCase, getNextActionUseCase, banMapUseCase, pickMapUseCase, selectSideUseCase, resetSessionUseCase, startSessionUseCase, mapPoolRepo, roomRepo, wsManager)
	mapPoolHandler := http.NewMapPoolHandler(getPoolsUseCase, getPoolUseCase, createCustomPoolUseCase, deletePoolUseCase)
	roomHandler := http.NewRoomHandler(createRoomUseCase, getRoomUseCase, getRoomBySessionUseCase, getRoomsListUseCase, joinRoomUseCase, leaveRoomUseCase, deleteRoomUseCase, updateRoomUseCase, wsManager)

	// Инициализируем WebSocket handler
	roomWebSocketHandler := websocket.NewRoomWebSocketHandler(
		wsManager,
		roomRepo,
		vetoSessionRepo,
		vetoActionRepo,
		mapRepo,
		mapPoolRepo,
		jwtService,
		banMapUseCase,
		pickMapUseCase,
		resetSessionUseCase,
		startSessionUseCase,
	)

	// API routes
	api := router.Group("/api")
	{
		// Public routes (не требуют авторизации)
		auth := api.Group("/auth")
		auth.Use(authRateLimit) // Строгий rate limiting для auth
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/me", middleware.AuthMiddleware(jwtService), authHandler.GetCurrentUser)
		}

		// Protected routes (требуют авторизации)
		users := api.Group("/users")
		users.Use(middleware.AuthMiddleware(jwtService))
		{
			users.GET("/profile", userHandler.GetProfile)
			users.PUT("/profile", userHandler.UpdateProfile)
			users.GET("/sessions", userHandler.GetSessions)
			users.GET("/rooms", userHandler.GetRooms)
		}

		// Veto routes (публичные, но могут быть созданы с авторизацией)
		// ВАЖНО: Более специфичные маршруты должны идти раньше общих
		vetoGroup := api.Group("/veto")
		{
			sessions := vetoGroup.Group("/sessions")
			{
				sessions.POST("", vetoHandler.CreateSession)
				// Специфичные маршруты идут первыми
				sessions.GET("/share/:token", vetoHandler.GetSessionByShareToken)
				sessions.GET("/:id/next-action", vetoHandler.GetNextAction)
				sessions.POST("/:id/start", vetoHandler.StartSession)
				sessions.POST("/:id/ban", vetoHandler.BanMap)
				sessions.POST("/:id/pick", vetoHandler.PickMap)
				sessions.POST("/:id/select-side", vetoHandler.SelectSide)
				sessions.POST("/:id/reset", vetoHandler.ResetSession)
				// Общий маршрут GET /:id должен быть последним
				sessions.GET("/:id", vetoHandler.GetSession)
			}
		}

		// Map Pools routes (требуют авторизации)
		mapPools := api.Group("/map-pools")
		mapPools.Use(middleware.AuthMiddleware(jwtService))
		{
			mapPools.GET("/games/:gameId", mapPoolHandler.GetPools)
			mapPools.GET("/:id", mapPoolHandler.GetPool)
			mapPools.POST("", mapPoolHandler.CreateCustomPool)
			mapPools.DELETE("/:id", mapPoolHandler.DeletePool)
		}

		// Rooms routes
		api.GET("/rooms", roomHandler.GetRooms)
		rooms := api.Group("/rooms")
		rooms.Use(middleware.AuthMiddleware(jwtService))
		{
			rooms.POST("", roomHandler.CreateRoom)
			rooms.GET("/by-session/:sessionId", roomHandler.GetRoomBySession)
			rooms.GET("/:id", roomHandler.GetRoom)
			rooms.POST("/:id/join", roomHandler.JoinRoom)
			rooms.POST("/:id/leave", roomHandler.LeaveRoom)
			rooms.PUT("/:id", roomHandler.UpdateRoom)
			rooms.DELETE("/:id", roomHandler.DeleteRoom)
			rooms.GET("/:id/participants", roomHandler.GetParticipants)
		}

		// WebSocket routes (auth handled in handler via query param)
		router.GET("/ws/room/:roomId", roomWebSocketHandler.HandleWebSocket)
	}

	// Запускаем сервер
	serverAddr := ":" + cfg.Port
	log.Printf("Server starting on %s", serverAddr)

	// Graceful shutdown
	go func() {
		if err := router.Run(serverAddr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Ожидаем сигнал для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
}

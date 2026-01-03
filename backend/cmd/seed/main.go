package main

import (
	"log"
	"os"

	"github.com/bbp/backend/config"
	"github.com/bbp/backend/internal/repository/sqlite"
	"github.com/bbp/backend/pkg/database"
	"github.com/bbp/backend/pkg/seed"
	"github.com/bbp/backend/internal/repository/models"
)

func main() {
	log.Println("Starting seed command...")

	// Загружаем конфигурацию
	cfg := config.Load()

	// Инициализируем базу данных
	db, err := database.NewDatabase(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close(db)

	// Выполняем миграции перед seed
	log.Println("Running database migrations...")
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

	// Инициализируем репозитории
	gameRepo := sqlite.NewGameRepository(db)
	mapRepo := sqlite.NewMapRepository(db)
	mapPoolRepo := sqlite.NewMapPoolRepository(db)

	// Выполняем seed
	if err := seed.Seed(gameRepo, mapRepo, mapPoolRepo); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
		os.Exit(1)
	}

	log.Println("Seed completed successfully!")
}

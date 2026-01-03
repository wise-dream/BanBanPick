package database

import (
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewDatabase создает новое подключение к SQLite базе данных
func NewDatabase(dsn string) (*gorm.DB, error) {
	// Создаем директорию для БД, если её нет
	dbDir := filepath.Dir(dsn)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Настраиваем GORM с логированием
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// Подключаемся к SQLite
	db, err := gorm.Open(sqlite.Open(dsn), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Включаем foreign keys для SQLite
	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	return db, nil
}

// Migrate выполняет автоматические миграции для всех моделей
func Migrate(db *gorm.DB, models ...interface{}) error {
	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	
	// Создаем уникальный индекс для room_participants (room_id, user_id)
	// SQLite не поддерживает частичные индексы с WHERE, поэтому создаем обычный
	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_room_participants_unique 
		ON room_participants(room_id, user_id)
	`).Error; err != nil {
		// Игнорируем ошибку если индекс уже существует
		if err.Error() != "UNIQUE constraint failed" {
			return fmt.Errorf("failed to create unique index: %w", err)
		}
	}
	
	return nil
}

// Close закрывает подключение к базе данных
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

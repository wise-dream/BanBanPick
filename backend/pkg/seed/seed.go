package seed

import (
	"fmt"
	"log"

	"github.com/bbp/backend/internal/domain/entities"
	"github.com/bbp/backend/internal/domain/repositories"
)

// Seed заполняет базу данных начальными данными
// Функция идемпотентна - можно вызывать несколько раз без дублирования данных
func Seed(
	gameRepo repositories.GameRepository,
	mapRepo repositories.MapRepository,
	mapPoolRepo repositories.MapPoolRepository,
) error {
	log.Println("Starting database seeding...")

	// 1. Создаем игру Valorant
	valorantGame := &entities.Game{
		ID:       1,
		Name:     "Valorant",
		Slug:     "valorant",
		IsActive: true,
	}

	existingGame, err := gameRepo.GetByID(1)
	if err != nil {
		return fmt.Errorf("failed to check existing game: %w", err)
	}

	if existingGame == nil {
		if err := gameRepo.Create(valorantGame); err != nil {
			return fmt.Errorf("failed to create Valorant game: %w", err)
		}
		log.Println("✓ Created Valorant game")
	} else {
		log.Println("✓ Valorant game already exists")
	}

	// 2. Создаем карты Valorant
	maps := []struct {
		ID            uint
		Name          string
		Slug          string
		ImageURL      string
		IsCompetitive bool
		IsActive      bool
	}{
		{1, "Abyss", "abyss", "/images/abyss.png", true, true},
		{2, "Ascent", "ascent", "/images/ascent.png", false, true},
		{3, "Bind", "bind", "/images/bind.png", true, true},
		{4, "Breeze", "breeze", "/images/breeze.png", false, true},
		{5, "Corrode", "corrode", "/images/corrode.png", true, true},
		{6, "Fracture", "fracture", "/images/fracture.png", false, true},
		{7, "Haven", "haven", "/images/haven.png", true, true},
		{8, "Icebox", "icebox", "/images/icebox.png", false, true},
		{9, "Lotus", "lotus", "/images/lotus.png", false, true},
		{10, "Pearl", "pearl", "/images/pearl.png", true, true},
		{11, "Split", "split", "/images/split.png", true, true},
		{12, "Sunset", "sunset", "/images/sunset.png", true, true},
	}

	createdMaps := 0
	for _, mapData := range maps {
		existingMap, err := mapRepo.GetByID(mapData.ID)
		if err != nil {
			return fmt.Errorf("failed to check existing map %s: %w", mapData.Name, err)
		}

		if existingMap == nil {
			mapEntity := &entities.Map{
				ID:            mapData.ID,
				GameID:        1,
				Name:          mapData.Name,
				Slug:          mapData.Slug,
				ImageURL:      mapData.ImageURL,
				IsActive:      mapData.IsActive,
				IsCompetitive: mapData.IsCompetitive,
			}

			if err := mapRepo.Create(mapEntity); err != nil {
				return fmt.Errorf("failed to create map %s: %w", mapData.Name, err)
			}
			createdMaps++
		}
	}

	if createdMaps > 0 {
		log.Printf("✓ Created %d maps", createdMaps)
	} else {
		log.Println("✓ All maps already exist")
	}

	// 3. Создаем системные пулы карт (All Maps и Competitive Maps)
	// Системные пулы доступны всем пользователям (user_id = NULL)
	allMaps, err := mapRepo.GetByGameID(1)
	if err != nil {
		return fmt.Errorf("failed to get maps for system pools: %w", err)
	}

	// Проверяем и создаем пул "All Maps"
	existingAllMapsPool, err := mapPoolRepo.GetByID(1)
	if err != nil {
		return fmt.Errorf("failed to check existing All Maps pool: %w", err)
	}

	if existingAllMapsPool == nil {
		allMapsPool := &entities.MapPool{
			ID:       1,
			GameID:   1,
			UserID:   nil, // Системный пул - без привязки к пользователю
			Name:     "All Maps",
			Type:     entities.MapPoolTypeAll,
			IsSystem: true,
			Maps:     allMaps,
		}

		if err := mapPoolRepo.Create(allMapsPool); err != nil {
			return fmt.Errorf("failed to create All Maps pool: %w", err)
		}
		log.Println("✓ Created All Maps system pool")
	} else {
		log.Println("✓ All Maps system pool already exists")
	}

	// Проверяем и создаем пул "Competitive Maps"
	existingCompetitivePool, err := mapPoolRepo.GetByID(2)
	if err != nil {
		return fmt.Errorf("failed to check existing Competitive Maps pool: %w", err)
	}

	if existingCompetitivePool == nil {
		// Фильтруем только соревновательные карты
		competitiveMaps := make([]entities.Map, 0)
		for _, m := range allMaps {
			if m.IsCompetitive {
				competitiveMaps = append(competitiveMaps, m)
			}
		}

		competitivePool := &entities.MapPool{
			ID:       2,
			GameID:   1,
			UserID:   nil, // Системный пул - без привязки к пользователю
			Name:     "Competitive Maps",
			Type:     entities.MapPoolTypeCompetitive,
			IsSystem: true,
			Maps:     competitiveMaps,
		}

		if err := mapPoolRepo.Create(competitivePool); err != nil {
			return fmt.Errorf("failed to create Competitive Maps pool: %w", err)
		}
		log.Println("✓ Created Competitive Maps system pool")
	} else {
		log.Println("✓ Competitive Maps system pool already exists")
	}

	log.Println("Database seeding completed successfully!")
	return nil
}

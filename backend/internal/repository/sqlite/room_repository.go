package sqlite

import (
	"errors"
	"github.com/bbp/backend/internal/domain/entities"
	"github.com/bbp/backend/internal/domain/repositories"
	"github.com/bbp/backend/internal/repository/models"
	"gorm.io/gorm"
)

type roomRepository struct {
	db *gorm.DB
}

func NewRoomRepository(db *gorm.DB) repositories.RoomRepository {
	return &roomRepository{db: db}
}

func (r *roomRepository) Create(room *entities.Room) error {
	model := &models.RoomModel{
		OwnerID:         room.OwnerID,
		Name:            room.Name,
		Code:            room.Code,
		Password:        room.Password,
		Type:            string(room.Type),
		Status:          string(room.Status),
		GameID:          room.GameID,
		MapPoolID:       room.MapPoolID,
		VetoSessionID:   room.VetoSessionID,
		MaxParticipants: room.MaxParticipants,
	}

	if err := r.db.Create(model).Error; err != nil {
		return err
	}

	room.ID = model.ID
	room.CreatedAt = model.CreatedAt
	room.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *roomRepository) GetByID(id uint) (*entities.Room, error) {
	var model models.RoomModel
	if err := r.db.First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	room := toRoomEntity(&model)
	
	// Загружаем участников отдельно
	participants, err := r.GetParticipants(id)
	if err != nil {
		return nil, err
	}
	room.Participants = participants

	return room, nil
}

func (r *roomRepository) GetByCode(code string) (*entities.Room, error) {
	var model models.RoomModel
	if err := r.db.Where("code = ?", code).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	room := toRoomEntity(&model)
	
	// Загружаем участников отдельно
	participants, err := r.GetParticipants(model.ID)
	if err != nil {
		return nil, err
	}
	room.Participants = participants

	return room, nil
}

func (r *roomRepository) GetByOwnerID(ownerID uint) ([]entities.Room, error) {
	var modelList []models.RoomModel
	if err := r.db.Where("owner_id = ?", ownerID).Find(&modelList).Error; err != nil {
		return nil, err
	}

	rooms := make([]entities.Room, len(modelList))
	for i, model := range modelList {
		room := toRoomEntity(&model)
		participants, _ := r.GetParticipants(model.ID)
		room.Participants = participants
		rooms[i] = *room
	}

	return rooms, nil
}

func (r *roomRepository) GetPublicRooms(limit, offset int) ([]entities.Room, error) {
	filter := &repositories.RoomFilter{
		Type: func() *string { s := "public"; return &s }(),
	}
	return r.GetRooms(filter, limit, offset)
}

func (r *roomRepository) GetRooms(filter *repositories.RoomFilter, limit, offset int) ([]entities.Room, error) {
	var modelList []models.RoomModel
	query := r.db.Model(&models.RoomModel{})
	
	// Применяем фильтры
	if filter != nil {
		if filter.Type != nil {
			query = query.Where("type = ?", *filter.Type)
		}
		if filter.Status != nil {
			query = query.Where("status = ?", *filter.Status)
		}
	}
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&modelList).Error; err != nil {
		return nil, err
	}

	rooms := make([]entities.Room, len(modelList))
	for i, model := range modelList {
		room := toRoomEntity(&model)
		participants, _ := r.GetParticipants(model.ID)
		room.Participants = participants
		rooms[i] = *room
	}

	return rooms, nil
}

func (r *roomRepository) Update(room *entities.Room) error {
	model := &models.RoomModel{
		ID:              room.ID,
		OwnerID:         room.OwnerID,
		Name:            room.Name,
		Code:            room.Code,
		Type:            string(room.Type),
		Status:          string(room.Status),
		GameID:         room.GameID,
		MapPoolID:      room.MapPoolID,
		VetoSessionID:  room.VetoSessionID,
		MaxParticipants: room.MaxParticipants,
	}

	return r.db.Model(&models.RoomModel{}).Where("id = ?", room.ID).Updates(model).Error
}

func (r *roomRepository) Delete(id uint) error {
	return r.db.Delete(&models.RoomModel{}, id).Error
}

func (r *roomRepository) AddParticipant(participant *entities.RoomParticipant) error {
	model := &models.RoomParticipantModel{
		RoomID:   participant.RoomID,
		UserID:   participant.UserID,
		Role:     string(participant.Role),
		JoinedAt: participant.JoinedAt,
	}

	if err := r.db.Create(model).Error; err != nil {
		return err
	}

	participant.ID = model.ID
	return nil
}

func (r *roomRepository) RemoveParticipant(roomID, userID uint) error {
	return r.db.Where("room_id = ? AND user_id = ?", roomID, userID).Delete(&models.RoomParticipantModel{}).Error
}

func (r *roomRepository) GetParticipants(roomID uint) ([]entities.RoomParticipant, error) {
	type ParticipantWithUser struct {
		models.RoomParticipantModel
		Username *string `gorm:"column:username"`
	}

	var results []ParticipantWithUser
	if err := r.db.Table("room_participants").
		Select("room_participants.*, users.username").
		Joins("LEFT JOIN users ON room_participants.user_id = users.id").
		Where("room_participants.room_id = ?", roomID).
		Find(&results).Error; err != nil {
		return nil, err
	}

	participants := make([]entities.RoomParticipant, len(results))
	for i, result := range results {
		participants[i] = *toRoomParticipantEntityWithUsername(&result.RoomParticipantModel, result.Username)
	}

	return participants, nil
}

func (r *roomRepository) GetParticipant(roomID, userID uint) (*entities.RoomParticipant, error) {
	var model models.RoomParticipantModel
	if err := r.db.Where("room_id = ? AND user_id = ?", roomID, userID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return toRoomParticipantEntity(&model), nil
}

// GetUserRoom получает комнату, в которой участвует пользователь
func (r *roomRepository) GetByVetoSessionID(sessionID uint) (*entities.Room, error) {
	var model models.RoomModel
	if err := r.db.Where("veto_session_id = ?", sessionID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	room := toRoomEntity(&model)
	
	// Загружаем участников отдельно
	participants, err := r.GetParticipants(model.ID)
	if err != nil {
		return nil, err
	}
	room.Participants = participants

	return room, nil
}

func (r *roomRepository) GetUserRoom(userID uint) (*entities.Room, error) {
	var participant models.RoomParticipantModel
	if err := r.db.Where("user_id = ?", userID).First(&participant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// Получаем комнату по ID из participant
	return r.GetByID(participant.RoomID)
}

// Count подсчитывает количество комнат с фильтром
func (r *roomRepository) Count(filter *repositories.RoomFilter) (int64, error) {
	var count int64
	query := r.db.Model(&models.RoomModel{})

	if filter != nil {
		if filter.Type != nil {
			query = query.Where("type = ?", *filter.Type)
		}
		if filter.Status != nil {
			query = query.Where("status = ?", *filter.Status)
		}
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func toRoomEntity(model *models.RoomModel) *entities.Room {
	return &entities.Room{
		ID:              model.ID,
		OwnerID:         model.OwnerID,
		Name:            model.Name,
		Code:            model.Code,
		Password:        model.Password,
		Type:            entities.RoomType(model.Type),
		Status:          entities.RoomStatus(model.Status),
		GameID:          model.GameID,
		MapPoolID:       model.MapPoolID,
		VetoSessionID:   model.VetoSessionID,
		MaxParticipants: model.MaxParticipants,
		CreatedAt:       model.CreatedAt,
		UpdatedAt:       model.UpdatedAt,
		Participants:    []entities.RoomParticipant{}, // Загружаются отдельно через GetParticipants
	}
}

func toRoomParticipantEntity(model *models.RoomParticipantModel) *entities.RoomParticipant {
	return toRoomParticipantEntityWithUsername(model, nil)
}

func toRoomParticipantEntityWithUsername(model *models.RoomParticipantModel, username *string) *entities.RoomParticipant {
	return &entities.RoomParticipant{
		ID:       model.ID,
		RoomID:   model.RoomID,
		UserID:   model.UserID,
		Username: username,
		Role:     entities.ParticipantRole(model.Role),
		JoinedAt: model.JoinedAt,
	}
}

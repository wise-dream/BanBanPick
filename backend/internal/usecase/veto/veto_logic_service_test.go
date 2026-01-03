package veto

import (
	"testing"

	"github.com/bbp/backend/internal/domain/entities"
)

func TestGetCurrentStep(t *testing.T) {
	service := NewVetoLogicService()

	tests := []struct {
		name    string
		actions []entities.VetoAction
		want    int
	}{
		{
			name:    "no actions",
			actions: []entities.VetoAction{},
			want:    1,
		},
		{
			name: "one action",
			actions: []entities.VetoAction{
				{ID: 1},
			},
			want: 2,
		},
		{
			name: "three actions",
			actions: []entities.VetoAction{
				{ID: 1},
				{ID: 2},
				{ID: 3},
			},
			want: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.GetCurrentStep(tt.actions)
			if got != tt.want {
				t.Errorf("GetCurrentStep() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetCurrentTeam(t *testing.T) {
	service := NewVetoLogicService()

	tests := []struct {
		name        string
		sessionType entities.VetoType
		step        int
		want        string
	}{
		{
			name:        "step 1 - team A",
			sessionType: entities.VetoTypeBo1,
			step:        1,
			want:        "A",
		},
		{
			name:        "step 2 - team B",
			sessionType: entities.VetoTypeBo1,
			step:        2,
			want:        "B",
		},
		{
			name:        "step 3 - team A",
			sessionType: entities.VetoTypeBo3,
			step:        3,
			want:        "A",
		},
		{
			name:        "step 4 - team B",
			sessionType: entities.VetoTypeBo3,
			step:        4,
			want:        "B",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.GetCurrentTeam(tt.sessionType, tt.step)
			if got != tt.want {
				t.Errorf("GetCurrentTeam() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetNextActionType_Bo1(t *testing.T) {
	service := NewVetoLogicService()

	session := &entities.VetoSession{
		Type: entities.VetoTypeBo1,
	}

	tests := []struct {
		name             string
		actions          []entities.VetoAction
		availableMapsCount int
		want             NextActionType
	}{
		{
			name:              "step 1",
			actions:           []entities.VetoAction{},
			availableMapsCount: 7,
			want:              NextActionTypeBan,
		},
		{
			name: "step 2",
			actions: []entities.VetoAction{
				{ActionType: entities.VetoActionTypeBan},
			},
			availableMapsCount: 6,
			want:              NextActionTypeBan,
		},
		{
			name: "step 6",
			actions: []entities.VetoAction{
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypeBan},
			},
			availableMapsCount: 2,
			want:              NextActionTypeBan,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.GetNextActionType(session, tt.actions, tt.availableMapsCount)
			if got != tt.want {
				t.Errorf("GetNextActionType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetNextActionType_Bo3(t *testing.T) {
	service := NewVetoLogicService()

	session := &entities.VetoSession{
		Type: entities.VetoTypeBo3,
	}

	tests := []struct {
		name             string
		actions          []entities.VetoAction
		availableMapsCount int
		want             NextActionType
	}{
		{
			name:              "step 1 - ban",
			actions:           []entities.VetoAction{},
			availableMapsCount: 7,
			want:              NextActionTypeBan,
		},
		{
			name: "step 2 - ban",
			actions: []entities.VetoAction{
				{ActionType: entities.VetoActionTypeBan},
			},
			availableMapsCount: 6,
			want:              NextActionTypeBan,
		},
		{
			name: "step 3 - pick",
			actions: []entities.VetoAction{
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypeBan},
			},
			availableMapsCount: 5,
			want:              NextActionTypePick,
		},
		{
			name: "step 6 - pick",
			actions: []entities.VetoAction{
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypePick},
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypeBan},
			},
			availableMapsCount: 2,
			want:              NextActionTypePick,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.GetNextActionType(session, tt.actions, tt.availableMapsCount)
			if got != tt.want {
				t.Errorf("GetNextActionType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetNextActionType_Bo5(t *testing.T) {
	service := NewVetoLogicService()

	session := &entities.VetoSession{
		Type: entities.VetoTypeBo5,
	}

	tests := []struct {
		name             string
		actions          []entities.VetoAction
		availableMapsCount int
		want             NextActionType
	}{
		{
			name:              "step 1 - ban",
			actions:           []entities.VetoAction{},
			availableMapsCount: 7,
			want:              NextActionTypeBan,
		},
		{
			name: "step 4 - both",
			actions: []entities.VetoAction{
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypeBan},
			},
			availableMapsCount: 4,
			want:              NextActionTypeBoth,
		},
		{
			name: "step 5 - pick (after ban)",
			actions: []entities.VetoAction{
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypeBan},
			},
			availableMapsCount: 3,
			want:              NextActionTypePick,
		},
		{
			name: "step 5 - ban (after pick)",
			actions: []entities.VetoAction{
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypePick},
			},
			availableMapsCount: 3,
			want:              NextActionTypeBan,
		},
		{
			name: "step 6 - pick",
			actions: []entities.VetoAction{
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypePick},
			},
			availableMapsCount: 2,
			want:              NextActionTypePick,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.GetNextActionType(session, tt.actions, tt.availableMapsCount)
			if got != tt.want {
				t.Errorf("GetNextActionType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsVetoFinished_Bo1(t *testing.T) {
	service := NewVetoLogicService()

	session := &entities.VetoSession{
		Type: entities.VetoTypeBo1,
	}

	tests := []struct {
		name          string
		actions       []entities.VetoAction
		availableMaps []entities.Map
		want          bool
	}{
		{
			name:          "not finished - 7 maps",
			actions:       []entities.VetoAction{},
			availableMaps: []entities.Map{{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}, {ID: 5}, {ID: 6}, {ID: 7}},
			want:          false,
		},
		{
			name: "not finished - 2 maps",
			actions: []entities.VetoAction{
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypeBan},
			},
			availableMaps: []entities.Map{{ID: 1}, {ID: 2}},
			want:          false,
		},
		{
			name: "finished - 1 map",
			actions: []entities.VetoAction{
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypeBan},
			},
			availableMaps: []entities.Map{{ID: 1}},
			want:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.IsVetoFinished(session, tt.actions, tt.availableMaps)
			if got != tt.want {
				t.Errorf("IsVetoFinished() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsVetoFinished_Bo3(t *testing.T) {
	service := NewVetoLogicService()

	session := &entities.VetoSession{
		Type: entities.VetoTypeBo3,
	}

	tests := []struct {
		name          string
		actions       []entities.VetoAction
		availableMaps []entities.Map
		want          bool
	}{
		{
			name:          "not finished - no picks",
			actions:       []entities.VetoAction{},
			availableMaps: []entities.Map{{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}, {ID: 5}},
			want:          false,
		},
		{
			name: "not finished - 1 pick",
			actions: []entities.VetoAction{
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypePick},
			},
			availableMaps: []entities.Map{{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}},
			want:          false,
		},
		{
			name: "finished - 2 picks, 1 map left",
			actions: []entities.VetoAction{
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypePick},
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypePick},
			},
			availableMaps: []entities.Map{{ID: 1}},
			want:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.IsVetoFinished(session, tt.actions, tt.availableMaps)
			if got != tt.want {
				t.Errorf("IsVetoFinished() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsVetoFinished_Bo5(t *testing.T) {
	service := NewVetoLogicService()

	session := &entities.VetoSession{
		Type: entities.VetoTypeBo5,
	}

	tests := []struct {
		name          string
		actions       []entities.VetoAction
		availableMaps []entities.Map
		want          bool
	}{
		{
			name:          "not finished - no picks",
			actions:       []entities.VetoAction{},
			availableMaps: []entities.Map{{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}, {ID: 5}, {ID: 6}, {ID: 7}},
			want:          false,
		},
		{
			name: "not finished - 2 picks",
			actions: []entities.VetoAction{
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypePick},
				{ActionType: entities.VetoActionTypePick},
			},
			availableMaps: []entities.Map{{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}, {ID: 5}},
			want:          false,
		},
		{
			name: "finished - 4 picks, 1 map left",
			actions: []entities.VetoAction{
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypePick},
				{ActionType: entities.VetoActionTypePick},
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypePick},
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypeBan},
				{ActionType: entities.VetoActionTypePick},
			},
			availableMaps: []entities.Map{{ID: 1}},
			want:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.IsVetoFinished(session, tt.actions, tt.availableMaps)
			if got != tt.want {
				t.Errorf("IsVetoFinished() = %v, want %v", got, tt.want)
			}
		})
	}
}

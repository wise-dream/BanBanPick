package veto

import "errors"

var (
	ErrSessionNotFound        = errors.New("session not found")
	ErrSessionFinished        = errors.New("session is already finished")
	ErrSessionAlreadyStarted  = errors.New("session is already started")
	ErrInvalidAction          = errors.New("invalid action")
	ErrMapNotFound            = errors.New("map not found")
	ErrMapAlreadyBanned       = errors.New("map is already banned")
	ErrMapAlreadyPicked       = errors.New("map is already picked")
	ErrNotYourTurn            = errors.New("not your turn")
	ErrInvalidTeam            = errors.New("invalid team")
	ErrMapPoolNotFound        = errors.New("map pool not found")
	ErrInvalidMapPool         = errors.New("invalid map pool")
	ErrGameNotFound           = errors.New("game not found")
	ErrInvalidSessionType     = errors.New("invalid session type")
)
package room

import "errors"

var (
	ErrRoomNotFound      = errors.New("room not found")
	ErrGameNotFound      = errors.New("game not found")
	ErrMapPoolNotFound   = errors.New("map pool not found")
	ErrInvalidRoom       = errors.New("invalid room")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrRoomFull          = errors.New("room is full")
	ErrAlreadyInRoom     = errors.New("user is already in a room")
	ErrInvalidCode       = errors.New("invalid room code")
	ErrCannotJoinPrivate = errors.New("cannot join private room without code")
)

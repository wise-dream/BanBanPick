package map_pool

import "errors"

var (
	ErrMapPoolNotFound    = errors.New("map pool not found")
	ErrGameNotFound       = errors.New("game not found")
	ErrMapNotFound        = errors.New("map not found")
	ErrInvalidMapPool     = errors.New("invalid map pool")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrCannotDeleteSystem = errors.New("cannot delete system map pool")
	ErrPoolHasNoMaps      = errors.New("map pool must have at least one map")
)

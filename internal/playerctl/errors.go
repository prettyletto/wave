package playerctl

import "errors"

var (
	ErrPlayerctlNotFound = errors.New("playerctl not found")
	ErrNoPlayersFound    = errors.New("no players found")
	ErrNoActivePlayer    = errors.New("no active player")
	ErrInvalidResponse   = errors.New("invalid playerctl response")
)

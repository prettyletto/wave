package playerctl

import (
	"fmt"
	"strings"
)

type PlayStatus string

const (
	StatusPlaying PlayStatus = "Playing"
	StatusPaused  PlayStatus = "Paused"
	StatusStop    PlayStatus = "Stopped"
)

func ParsePlayStatus(s string) (PlayStatus, error) {
	switch strings.TrimSpace(s) {
	case string(StatusPlaying):
		return StatusPlaying, nil
	case string(StatusPaused):
		return StatusPaused, nil
	case string(StatusStop):
		return StatusStop, nil
	default:
		return "", fmt.Errorf("invalid playback status: %q", s)
	}
}

type LoopStatus string

const (
	LoopNone     LoopStatus = "None"
	LoopTrack    LoopStatus = "Track"
	LoopPlaylist LoopStatus = "Playlist"
)

func ParseLoopStatus(s string) (LoopStatus, error) {
	switch strings.TrimSpace(s) {
	case string(LoopNone):
		return LoopNone, nil
	case string(LoopTrack):
		return LoopTrack, nil
	case string(LoopPlaylist):
		return LoopPlaylist, nil
	default:
		return "", fmt.Errorf("ivalid loop status: %q", s)
	}
}

type ShuffleStatus string

const (
	ShuffleOff    ShuffleStatus = "Off"
	ShuffleOn     ShuffleStatus = "On"
	ShuffleToggle ShuffleStatus = "Toggle"
)

func ParseShuffleStatus(s string) (ShuffleStatus, error) {
	switch strings.TrimSpace(s) {
	case string(ShuffleOff):
		return ShuffleOff, nil
	case string(ShuffleOn):
		return ShuffleOn, nil
	case string(ShuffleToggle):
		return ShuffleToggle, nil
	default:
		return "", fmt.Errorf("ivalid loop status: %q", s)
	}
}

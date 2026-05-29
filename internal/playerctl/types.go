package playerctl

import (
	"fmt"
	"strconv"
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
		return "", fmt.Errorf("invalid loop status: %q", s)
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
		return "", fmt.Errorf("invalid shuffle status: %q", s)
	}
}

type TrackInfo struct {
	Player   string
	Status   PlayStatus
	Title    string
	Artist   string
	Album    string
	LengthUS int64
	Position int64
}

func (t TrackInfo) String() string {
	return fmt.Sprintf(
		"Player: %s\nStatus: %s\nTitle: %s\nArtist: %s\nAlbum: %s\nLength: %d\nPosition: %d", t.Player,
		t.Status,
		t.Title,
		t.Artist,
		t.Album,
		t.LengthUS,
		t.Position,
	)
}

func ParseTrackInfo(s string) (TrackInfo, error) {
	parts := strings.Split(strings.TrimSpace(s), "\t")

	if len(parts) != 6 {
		return TrackInfo{}, fmt.Errorf("invalid metadata field count: got %d", len(parts))
	}

	st, err := ParsePlayStatus(parts[1])
	if err != nil {
		return TrackInfo{}, fmt.Errorf("invalid status: %w", err)
	}

	length := int64(0)
	if strings.TrimSpace(parts[5]) != "" {
		length, err = strconv.ParseInt(strings.TrimSpace(parts[5]), 10, 64)
		if err != nil {
			return TrackInfo{}, fmt.Errorf("invalid mpris:length %q: %w", parts[5], err)
		}
	}

	return TrackInfo{
		Player:   strings.TrimSpace(parts[0]),
		Status:   st,
		Title:    strings.TrimSpace(parts[2]),
		Artist:   strings.TrimSpace(parts[3]),
		Album:    strings.TrimSpace(parts[4]),
		LengthUS: length,
	}, nil

}

package tui

import (
	"time"

	"github.com/prettyletto/wave/internal/playerctl"
	"github.com/prettyletto/wave/internal/wpctl"
)

type playersMsg struct {
	players []string
	err     error
}

type nowMsg struct {
	info playerctl.TrackInfo
	err  error
}

type streamsMsg struct {
	streams []wpctl.Stream
	err     error
}

type streamMsg struct {
	stream wpctl.Stream
	err    error
}

type toggleMsg struct {
	err error
}

type previousMsg struct {
	err error
}

type nextMsg struct {
	err error
}

type seekMsg struct {
	err error
}

type shuffleMsg struct {
	shuffle playerctl.ShuffleStatus
	err     error
}

type loopMsg struct {
	loop playerctl.LoopStatus
	err  error
}

type shuffleToggleMsg struct{ err error }
type loopSetMsg struct{ err error }

type (
	volumeMsg struct{ err error }
	muteMsg   struct{ err error }
)

type tickMsg time.Time

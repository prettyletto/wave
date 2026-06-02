package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/prettyletto/wave/internal/playerctl"
	"github.com/prettyletto/wave/internal/state"
	"github.com/prettyletto/wave/internal/wpctl"
)

type listSection int

const (
	sectionPlayers listSection = iota
	sectionStreams
)

type Player interface {
	Players(context.Context) ([]string, error)

	Now(context.Context) (playerctl.TrackInfo, error)
	NowForPlayer(context.Context, string) (playerctl.TrackInfo, error)

	ToggleForPlayer(context.Context, string) error
	LoopForPlayer(context.Context, string) (playerctl.LoopStatus, error)
	SetLoopForPlayer(context.Context, string, string) error
	ShuffleForPlayer(context.Context, string) (playerctl.ShuffleStatus, error)
	ToggleShuffleForPlayer(context.Context, string) error
	SeekForPlayer(context.Context, string, int) error
	PreviousForPlayer(context.Context, string) error
	NextForPlayer(context.Context, string) error
}

type Audio interface {
	Streams(context.Context) ([]wpctl.Stream, error)
	StreamByID(context.Context, string) (wpctl.Stream, error)
	ChangeVolume(context.Context, string, float64, int) error
	ToggleMute(context.Context, string) error
}

type model struct {
	player Player
	audio  Audio

	now                playerctl.TrackInfo
	loop               playerctl.LoopStatus
	shuffle            playerctl.ShuffleStatus
	loopUnsupported    bool
	shuffleUnsupported bool
	err                error

	players        []string
	selectedPlayer string

	streams        []wpctl.Stream
	selectedStream string
	currentStream  wpctl.Stream

	listOpen     bool
	listSection  listSection
	playerCursor int
	streamCursor int
	streamLocked bool
}

func NewModel(p Player, a Audio) model {
	selectedPlayer, _ := state.LoadSelectedPlayer()

	return model{
		player:         p,
		audio:          a,
		selectedPlayer: selectedPlayer,
	}
}

func (m model) withSelectedPlayer(cmd func(string) tea.Cmd) (tea.Model, tea.Cmd) {
	if m.selectedPlayer == "" {
		return m, nil
	}
	return m, cmd(m.selectedPlayer)
}

func (m model) withSelectedStream(cmd func(string) tea.Cmd) (tea.Model, tea.Cmd) {
	if m.selectedStream == "" {
		return m, nil
	}
	return m, cmd(m.selectedStream)
}

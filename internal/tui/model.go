package tui

import (
	"context"
	"fmt"
	"slices"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/prettyletto/wave/internal/playerctl"
	"github.com/prettyletto/wave/internal/share"
	"github.com/prettyletto/wave/internal/state"
)

type Player interface {
	Players(context.Context) ([]string, error)

	Now(context.Context) (playerctl.TrackInfo, error)
	NowForPlayer(context.Context, string) (playerctl.TrackInfo, error)

	ToggleForPlayer(context.Context, string) error
	SeekForPlayer(context.Context, string, int) error
	PreviousForPlayer(context.Context, string) error
	NextForPlayer(context.Context, string) error
}

type playersMsg struct {
	players []string
	err     error
}

type nowMsg struct {
	info playerctl.TrackInfo
	err  error
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

type tickMsg time.Time

type model struct {
	player Player
	now    playerctl.TrackInfo
	err    error

	players        []string
	selectedPlayer string
}

func NewModel(p Player) model {
	selected, _ := state.LoadSelectedPlayer()

	return model{player: p, selectedPlayer: selected}
}

func fetchPlayersCmd(p Player) tea.Cmd {
	return func() tea.Msg {
		players, err := p.Players(context.Background())
		return playersMsg{players: players, err: err}
	}
}

func fetchNowCmd(p Player, selectedPlayer string) tea.Cmd {
	return func() tea.Msg {
		if selectedPlayer != "" {
			info, err := p.NowForPlayer(context.Background(), selectedPlayer)
			return nowMsg{info: info, err: err}
		}
		info, err := p.Now(context.Background())
		return nowMsg{info: info, err: err}
	}
}

func toggleCmd(p Player, selectedPlayer string) tea.Cmd {
	return func() tea.Msg {
		err := p.ToggleForPlayer(context.Background(), selectedPlayer)
		return toggleMsg{err: err}
	}
}

func seekCmd(p Player, selectedPlayer string, seconds int) tea.Cmd {
	return func() tea.Msg {
		err := p.SeekForPlayer(context.Background(), selectedPlayer, seconds)
		return seekMsg{err: err}
	}
}

func prevCmd(p Player, selectedPlayer string) tea.Cmd {
	return func() tea.Msg {
		err := p.PreviousForPlayer(context.Background(), selectedPlayer)
		return previousMsg{err: err}
	}
}

func nextCmd(p Player, selectedPlayer string) tea.Cmd {
	return func() tea.Msg {
		err := p.NextForPlayer(context.Background(), selectedPlayer)
		return nextMsg{err: err}
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second/2, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		fetchPlayersCmd(m.player),
		fetchNowCmd(m.player, m.selectedPlayer),
		tickCmd())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case playersMsg:
		m.players = msg.players
		m.err = msg.err
		if msg.err != nil {
			return m, nil
		}

		if m.selectedPlayer != "" && containsPlayer(m.players, m.selectedPlayer) {
			return m, nil
		}

		if m.now.Player != "" && containsPlayer(m.players, m.now.Player) {
			m.selectedPlayer = m.now.Player
			_ = state.SaveSelectedPlayer(m.selectedPlayer)
			return m, fetchNowCmd(m.player, m.selectedPlayer)
		}

		if len(m.players) > 0 {
			m.selectedPlayer = m.players[0]
			_ = state.SaveSelectedPlayer(m.selectedPlayer)
			return m, fetchNowCmd(m.player, m.selectedPlayer)
		}

	case nowMsg:
		if msg.err == nil && m.selectedPlayer == "" && msg.info.Player != "" {
			m.selectedPlayer = msg.info.Player
			_ = state.SaveSelectedPlayer(m.selectedPlayer)
		}
		m.now = msg.info
		m.err = msg.err
		return m, nil
	case toggleMsg:
		m.err = msg.err
		if msg.err != nil {
			return m, nil
		}
		return m, fetchNowCmd(m.player, m.selectedPlayer)
	case seekMsg:
		m.err = msg.err
		if msg.err != nil {
			return m, nil
		}
		return m, fetchNowCmd(m.player, m.selectedPlayer)
	case tickMsg:
		return m, tea.Batch(
			fetchPlayersCmd(m.player),
			fetchNowCmd(m.player, m.selectedPlayer),
			tickCmd())
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			m.selectedPlayer = nextPlayer(m.players, m.selectedPlayer, 1)
			_ = state.SaveSelectedPlayer(m.selectedPlayer)
			return m, fetchNowCmd(m.player, m.selectedPlayer)
		case "shift+tab":
			m.selectedPlayer = nextPlayer(m.players, m.selectedPlayer, -1)
			_ = state.SaveSelectedPlayer(m.selectedPlayer)
			return m, fetchNowCmd(m.player, m.selectedPlayer)
		case " ":
			return m, toggleCmd(m.player, m.selectedPlayer)
		case "ctrl+left":
			return m, prevCmd(m.player, m.selectedPlayer)
		case "ctrl+right":
			return m, nextCmd(m.player, m.selectedPlayer)
		case "left":
			return m, seekCmd(m.player, m.selectedPlayer, -5)
		case "right":
			return m, seekCmd(m.player, m.selectedPlayer, 5)
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("wave\n\nerror: %v\n\n[q] quit\n", m.err)
	}
	return fmt.Sprintf(
		"wave\n\nPlayer: %s\nStatus: %s\nTitle: %s\nArtist: %s\nPosition: %d\nLength: %d\nProgress: %s \n\n[␣]toggle  [q] quit\n",
		m.now.Player, m.now.Status, m.now.Title, m.now.Artist, m.now.Position, m.now.LengthUS, share.ProgressLabel(m.now.Position, m.now.LengthUS),
	)
}

func containsPlayer(players []string, selected string) bool {
	return slices.Contains(players, selected)
}

func selectedPlayerIndex(players []string, selected string) int {
	return slices.Index(players, selected)
}

func nextPlayer(players []string, selected string, delta int) string {
	if len(players) == 0 {
		return ""
	}

	i := selectedPlayerIndex(players, selected)
	if i == -1 {
		return players[0]
	}
	i = ((i + delta + len(players)) % len(players))
	return players[i]
}

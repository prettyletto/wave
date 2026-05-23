package tui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/prettyletto/wave/internal/playerctl"
	"github.com/prettyletto/wave/internal/share"
)

type Player interface {
	Now(context.Context) (playerctl.TrackInfo, error)
	Toggle(context.Context) error 

}

type nowMsg struct {
	info playerctl.TrackInfo
	err  error
}

type toggleMsg struct {
	err  error
}

type tickMsg time.Time

type model struct {
	player Player
	now    playerctl.TrackInfo
	err    error
}

func NewModel(p Player) model {
	return model{player: p}
}

func fetchNowCmd(p Player) tea.Cmd {
	return func() tea.Msg {
		info, err := p.Now(context.Background())
		return nowMsg{info: info, err: err}
	}
}

func toggleCmd(p Player) tea.Cmd {
	return func () tea.Msg {
		err:= p.Toggle(context.Background())
		return toggleMsg{err:err}
	}

}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Init() tea.Cmd {
	return tea.Batch(fetchNowCmd(m.player), tickCmd())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case nowMsg:
		m.now = msg.info
		m.err = msg.err
		return m, nil
	case tickMsg:
		return m, tea.Batch(fetchNowCmd(m.player), tickCmd())
	case tea.KeyMsg:
		switch msg.String() {
		case " ":
		return m, toggleCmd(m.player)
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
		m.now.Player, m.now.Status, m.now.Title, m.now.Artist, m.now.Position, m.now.LengthUS,share.ProgressLabel(m.now.Position, m.now.LengthUS),
	)
}

package tui

import (
	"errors"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/prettyletto/wave/internal/playerctl"
	"github.com/prettyletto/wave/internal/state"
	"github.com/prettyletto/wave/internal/wpctl"
)

func (m model) Init() tea.Cmd {
	return tea.Batch(
		fetchPlayersCmd(m.player),
		fetchNowCmd(m.player, m.selectedPlayer),
		fetchStreamsCmd(m.audio),
		fetchStreamCmd(m.audio, m.selectedStream),
		getLoopCmd(m.player, m.selectedPlayer),
		getShuffleCmd(m.player, m.selectedPlayer),
		tickCmd(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case playersMsg:
		return m.handlePlayersMsg(msg)
	case nowMsg:
		return m.handleNowMsg(msg)
	case loopMsg:
		return m.handleLoopMsg(msg)
	case loopSetMsg:
		return m.handleLoopSetMsg(msg)
	case shuffleMsg:
		return m.handleShuffleMsg(msg)
	case shuffleToggleMsg:
		return m.handleShuffleToggleMsg(msg)
	case streamsMsg:
		return m.handleStreamsMsg(msg)
	case streamMsg:
		return m.handleStreamMsg(msg)
	case toggleMsg:
		return m.handleToggleMsg(msg)
	case seekMsg:
		return m.handleSeekMsg(msg)
	case volumeMsg:
		return m.handleVolumeMsg(msg)
	case muteMsg:
		return m.handleMuteMsg(msg)
	case tickMsg:
		return m.handleTickMsg()
	case tea.KeyMsg:
		if m.listOpen {
			return m.handleListKey(msg)
		}
		return m.handleGlobalKey(msg)
	}

	return m, nil
}

func (m model) handlePlayersMsg(msg playersMsg) (tea.Model, tea.Cmd) {
	m.players = msg.players
	m.err = msg.err
	if msg.err != nil {
		return m, nil
	}

	if len(m.players) == 0 {
		m.selectedPlayer = ""
		m.now = playerctl.TrackInfo{}
		m.loop = ""
		m.shuffle = ""
		m.loopUnsupported = false
		m.shuffleUnsupported = false
		m.currentStream = wpctl.Stream{}
		m.selectedStream = ""
		m.err = nil
		return m, nil
	}

	if m.selectedPlayer != "" && containsPlayer(m.players, m.selectedPlayer) {
		return m, nil
	}

	if m.now.Player != "" && containsPlayer(m.players, m.now.Player) {
		m.selectedPlayer = m.now.Player
		_ = state.SaveSelectedPlayer(m.selectedPlayer)
		return m, tea.Batch(fetchNowCmd(m.player, m.selectedPlayer), m.syncSelectedStreamCmd())
	}

	m.selectedPlayer = m.players[0]
	_ = state.SaveSelectedPlayer(m.selectedPlayer)
	return m, tea.Batch(fetchNowCmd(m.player, m.selectedPlayer), m.syncSelectedStreamCmd())
}

func (m model) handleNowMsg(msg nowMsg) (tea.Model, tea.Cmd) {
	if errors.Is(msg.err, playerctl.ErrNoActivePlayer) || errors.Is(msg.err, playerctl.ErrNoPlayersFound) {
		m.now = playerctl.TrackInfo{}
		m.err = nil
		return m, nil
	}

	if msg.err == nil && m.selectedPlayer == "" && msg.info.Player != "" {
		m.selectedPlayer = msg.info.Player
		_ = state.SaveSelectedPlayer(m.selectedPlayer)
		return m, tea.Batch(
			fetchNowCmd(m.player, m.selectedPlayer),
			getLoopCmd(m.player, m.selectedPlayer),
			getShuffleCmd(m.player, m.selectedPlayer),
			m.syncSelectedStreamCmd(),
		)
	}

	m.now = msg.info
	m.err = msg.err
	return m, nil
}

func (m model) handleLoopMsg(msg loopMsg) (tea.Model, tea.Cmd) {
	if errors.Is(msg.err, playerctl.ErrUnsupported) {
		m.loop = ""
		m.loopUnsupported = true
		m.err = nil
		return m, nil
	}
	if msg.err != nil {
		m.err = msg.err
		return m, nil
	}
	m.loop = msg.loop
	m.loopUnsupported = false
	m.err = nil
	return m, nil
}

func (m model) handleLoopSetMsg(msg loopSetMsg) (tea.Model, tea.Cmd) {
	if errors.Is(msg.err, playerctl.ErrUnsupported) {
		m.loop = ""
		m.loopUnsupported = true
		m.err = nil
		return m, nil
	}
	m.err = msg.err
	if msg.err != nil {
		return m, nil
	}
	return m, getLoopCmd(m.player, m.selectedPlayer)
}

func (m model) handleShuffleMsg(msg shuffleMsg) (tea.Model, tea.Cmd) {
	if errors.Is(msg.err, playerctl.ErrUnsupported) {
		m.shuffle = ""
		m.shuffleUnsupported = true
		m.err = nil
		return m, nil
	}
	if msg.err != nil {
		m.err = msg.err
		return m, nil
	}
	m.shuffle = msg.shuffle
	m.shuffleUnsupported = false
	m.err = nil
	return m, nil
}

func (m model) handleShuffleToggleMsg(msg shuffleToggleMsg) (tea.Model, tea.Cmd) {
	if errors.Is(msg.err, playerctl.ErrUnsupported) {
		m.shuffle = ""
		m.shuffleUnsupported = true
		m.err = nil
		return m, nil
	}
	m.err = msg.err
	if msg.err != nil {
		return m, nil
	}
	return m, getShuffleCmd(m.player, m.selectedPlayer)
}

func (m model) handleStreamsMsg(msg streamsMsg) (tea.Model, tea.Cmd) {
	m.streams = msg.streams
	m.err = msg.err
	if msg.err != nil {
		return m, nil
	}
	return m, m.syncSelectedStreamCmd()
}

func (m model) handleStreamMsg(msg streamMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		return m, nil
	}
	if msg.stream.ID != "" {
		if stream, ok := findStreamByID(m.streams, msg.stream.ID); ok && stream.Name != "" {
			msg.stream.Name = stream.Name
		}
		m.currentStream = msg.stream
	}
	return m, nil
}

func (m model) handleToggleMsg(msg toggleMsg) (tea.Model, tea.Cmd) {
	m.err = msg.err
	if msg.err != nil {
		return m, nil
	}
	return m, fetchNowCmd(m.player, m.selectedPlayer)
}

func (m model) handleSeekMsg(msg seekMsg) (tea.Model, tea.Cmd) {
	m.err = msg.err
	if msg.err != nil {
		return m, nil
	}
	return m, fetchNowCmd(m.player, m.selectedPlayer)
}

func (m model) handleVolumeMsg(msg volumeMsg) (tea.Model, tea.Cmd) {
	m.err = msg.err
	if msg.err != nil {
		return m, nil
	}
	return m, fetchStreamCmd(m.audio, m.selectedStream)
}

func (m model) handleMuteMsg(msg muteMsg) (tea.Model, tea.Cmd) {
	m.err = msg.err
	if msg.err != nil {
		return m, nil
	}
	return m, fetchStreamCmd(m.audio, m.selectedStream)
}

func (m model) handleTickMsg() (tea.Model, tea.Cmd) {
	return m, tea.Batch(
		fetchPlayersCmd(m.player),
		fetchNowCmd(m.player, m.selectedPlayer),
		fetchStreamsCmd(m.audio),
		fetchStreamCmd(m.audio, m.selectedStream),
		tickCmd(),
	)
}

func (m model) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+l", "esc":
		m.closeList()
		return m, nil
	case "left", "h":
		m.listSection = sectionPlayers
		return m, nil
	case "right", "l":
		m.listSection = sectionStreams
		return m, nil
	case "up", "k":
		if m.listSection == sectionPlayers {
			m.playerCursor = moveCursor(m.playerCursor, len(m.players), -1)
		} else {
			m.streamCursor = moveCursor(m.streamCursor, len(m.streams), -1)
		}
		return m, nil
	case "down", "j":
		if m.listSection == sectionPlayers {
			m.playerCursor = moveCursor(m.playerCursor, len(m.players), 1)
		} else {
			m.streamCursor = moveCursor(m.streamCursor, len(m.streams), 1)
		}
		return m, nil
	case "enter":
		return m.handleListSelection()
	}

	return m, nil
}

func (m model) handleListSelection() (tea.Model, tea.Cmd) {
	if m.listSection == sectionPlayers {
		if len(m.players) == 0 || m.playerCursor >= len(m.players) {
			return m, nil
		}

		next := m.players[m.playerCursor]
		if next == "" {
			return m, nil
		}

		m.selectedPlayer = next
		m.streamLocked = false
		_ = state.SaveSelectedPlayer(m.selectedPlayer)
		m.closeList()

		return m, tea.Batch(
			fetchNowCmd(m.player, m.selectedPlayer),
			getLoopCmd(m.player, m.selectedPlayer),
			getShuffleCmd(m.player, m.selectedPlayer),
			m.syncSelectedStreamCmd(),
		)
	}

	if len(m.streams) == 0 || m.streamCursor >= len(m.streams) {
		return m, nil
	}

	m.selectedStream = m.streams[m.streamCursor].ID
	m.streamLocked = true
	m.closeList()
	return m, fetchStreamCmd(m.audio, m.selectedStream)
}

func (m model) handleGlobalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+l":
		m.openList()
		return m, nil
	case "tab":
		return m.handlePlayerCycle(1)
	case "shift+tab":
		return m.handlePlayerCycle(-1)
	case " ":
		return m.withSelectedPlayer(func(player string) tea.Cmd {
			return toggleCmd(m.player, player)
		})
	case "ctrl+left":
		return m.withSelectedPlayer(func(player string) tea.Cmd {
			return prevCmd(m.player, player)
		})
	case "ctrl+right":
		return m.withSelectedPlayer(func(player string) tea.Cmd {
			return nextCmd(m.player, player)
		})
	case "left":
		return m.withSelectedPlayer(func(player string) tea.Cmd {
			return seekCmd(m.player, player, -5)
		})
	case "right":
		return m.withSelectedPlayer(func(player string) tea.Cmd {
			return seekCmd(m.player, player, 5)
		})
	case "-":
		return m.withSelectedStream(func(streamID string) tea.Cmd {
			return volumeCmd(m.audio, streamID, m.currentStream.Volume, -5)
		})
	case "=":
		return m.withSelectedStream(func(streamID string) tea.Cmd {
			return volumeCmd(m.audio, streamID, m.currentStream.Volume, 5)
		})
	case "m":
		return m.withSelectedStream(func(streamID string) tea.Cmd {
			return muteCmd(m.audio, streamID)
		})
	case "s":
		return m.withSelectedPlayer(func(player string) tea.Cmd {
			return toggleShuffleCmd(m.player, player)
		})
	case "r":
		return m.withSelectedPlayer(func(player string) tea.Cmd {
			return setLoopCmd(m.player, player, nextLoopStatus(m.loop))
		})
	case "q", "ctrl+c":
		return m, tea.Quit
	}

	return m, nil
}

func (m model) handlePlayerCycle(delta int) (tea.Model, tea.Cmd) {
	next := nextPlayer(m.players, m.selectedPlayer, delta)
	if next == "" || next == m.selectedPlayer {
		return m, nil
	}

	m.selectedPlayer = next
	m.streamLocked = false
	_ = state.SaveSelectedPlayer(m.selectedPlayer)
	return m, tea.Batch(
		fetchNowCmd(m.player, m.selectedPlayer),
		getLoopCmd(m.player, m.selectedPlayer),
		getShuffleCmd(m.player, m.selectedPlayer),
		m.syncSelectedStreamCmd(),
	)
}

package tui

import (
	"fmt"
	"strings"

	"github.com/prettyletto/wave/internal/wpctl"
)

func (m *model) openList() {
	m.listOpen = true
	m.listSection = sectionPlayers

	m.playerCursor = selectedPlayerIndex(m.players, m.selectedPlayer)
	if m.playerCursor < 0 {
		m.playerCursor = 0
	}
	m.streamCursor = selectedStreamIndex(m.streams, m.selectedStream)
	if m.streamCursor < 0 {
		m.streamCursor = 0
	}
}

func (m *model) closeList() {
	m.listOpen = false
}

func moveCursor(current, length, delta int) int {
	if length == 0 {
		return 0
	}

	current += delta
	if current < 0 {
		current = length - 1
	}
	if current >= length {
		current = 0
	}

	return current
}

func renderPlayerList(players []string, selected string, cursor int, focused bool) string {
	if len(players) == 0 {
		return "  -"
	}

	var b strings.Builder
	for i, player := range players {
		prefix := " "
		if focused && i == cursor {
			prefix = "> "
		}

		label := player
		if player == selected {
			label += " [active]"
		}

		b.WriteString(prefix)
		b.WriteString(label)
		b.WriteByte('\n')
	}

	return strings.TrimRight(b.String(), "\n")
}

func renderStreamList(streams []wpctl.Stream, selected string, cursor int, focused bool) string {
	if len(streams) == 0 {
		return "  -"
	}

	var b strings.Builder
	for i, stream := range streams {
		prefix := " "
		if focused && i == cursor {
			prefix = "> "
		}

		label := stream.Name
		if label == "" {
			label = stream.AppName
		}
		if label == "" {
			label = stream.Binary
		}
		if label == "" {
			label = stream.ID
		}

		if stream.ID == selected {
			label += " [active]"
		}

		b.WriteString(prefix)
		b.WriteString(label)
		b.WriteByte('\n')
	}

	return strings.TrimRight(b.String(), "\n")
}

func (m model) listView() string {
	if !m.listOpen {
		return ""
	}

	playerTitle := "Players"
	streamTitle := "Streams"

	if m.listSection == sectionPlayers {
		playerTitle = "> Players"
	}
	if m.listSection == sectionStreams {
		streamTitle = "> Streams"
	}

	return fmt.Sprintf(
		"\nSelection\n\n%s\n%s\n\n%s\n%s\n\n[up/down] move  [left/right] focus  [enter] select  [esc] close\n",
		playerTitle,
		renderPlayerList(m.players, m.selectedPlayer, m.playerCursor, m.listSection == sectionPlayers),
		streamTitle,
		renderStreamList(m.streams, m.selectedStream, m.streamCursor, m.listSection == sectionStreams),
	)
}

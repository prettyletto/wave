package tui

import (
	"slices"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/prettyletto/wave/internal/playerctl"
	"github.com/prettyletto/wave/internal/wpctl"
)

func containsPlayer(players []string, selected string) bool {
	return slices.Contains(players, selected)
}

func containsStream(streams []wpctl.Stream, selected string) bool {
	for _, stream := range streams {
		if stream.ID == selected {
			return true
		}
	}
	return false
}

func findStreamByID(streams []wpctl.Stream, id string) (wpctl.Stream, bool) {
	for _, stream := range streams {
		if stream.ID == id {
			return stream, true
		}
	}
	return wpctl.Stream{}, false
}

func matchStreamID(player string, streams []wpctl.Stream) string {
	playerName := normalizePlayerName(player)
	if playerName == "" {
		return ""
	}

	bestID := ""
	bestScore := 0
	tied := false

	for _, stream := range streams {
		score := streamMatchScore(playerName, stream)
		if score == 0 {
			continue
		}

		if score > bestScore {
			bestID = stream.ID
			bestScore = score
			tied = false
			continue
		}

		if score == bestScore {
			tied = true
		}
	}

	if tied {
		return ""
	}

	return bestID
}

func selectedPlayerIndex(players []string, selected string) int {
	return slices.Index(players, selected)
}

func selectedStreamIndex(streams []wpctl.Stream, selected string) int {
	for i, stream := range streams {
		if stream.ID == selected {
			return i
		}
	}
	return -1
}

func normalizePlayerName(player string) string {
	player = strings.TrimSpace(strings.ToLower(player))
	if idx := strings.Index(player, ".instance"); idx != -1 {
		player = player[:idx]
	}
	if idx := strings.IndexRune(player, '.'); idx != -1 {
		player = player[:idx]
	}
	return normalizeAppName(player)
}

func normalizeAppName(value string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(value)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func streamMatchScore(player string, stream wpctl.Stream) int {
	if name := normalizeAppName(stream.Name); name != "" {
		switch {
		case name == player:
			return 300
		case strings.Contains(name, player), strings.Contains(player, name):
			return 180
		default:
			return 0
		}
	}

	fields := []struct {
		value string
		exact int
		fuzzy int
	}{
		{value: normalizeAppName(stream.AppName), exact: 220, fuzzy: 130},
		{value: normalizeAppName(stream.Binary), exact: 180, fuzzy: 100},
	}

	best := 0
	for _, field := range fields {
		if field.value == "" {
			continue
		}

		switch {
		case field.value == player:
			if field.exact > best {
				best = field.exact
			}
		case strings.Contains(field.value, player), strings.Contains(player, field.value):
			if field.fuzzy > best {
				best = field.fuzzy
			}
		}
	}

	return best
}

func nextPlayer(players []string, selected string, delta int) string {
	if len(players) == 0 {
		return ""
	}

	i := selectedPlayerIndex(players, selected)
	if i == -1 {
		return players[0]
	}

	i = (i + delta + len(players)) % len(players)
	return players[i]
}

func nextLoopStatus(current playerctl.LoopStatus) playerctl.LoopStatus {
	switch current {
	case playerctl.LoopNone:
		return playerctl.LoopTrack
	case playerctl.LoopTrack:
		return playerctl.LoopPlaylist
	case playerctl.LoopPlaylist:
		return playerctl.LoopNone
	default:
		return playerctl.LoopNone
	}
}

func (m *model) syncSelectedStreamCmd() tea.Cmd {
	if m.streamLocked && m.selectedStream != "" && containsStream(m.streams, m.selectedStream) {
		return fetchStreamCmd(m.audio, m.selectedStream)
	}

	next := matchStreamID(m.selectedPlayer, m.streams)
	if next == "" {
		m.selectedStream = ""
		m.currentStream = wpctl.Stream{}
		m.streamLocked = false
		return nil
	}

	if next == m.selectedStream && containsStream(m.streams, next) {
		return fetchStreamCmd(m.audio, next)
	}

	m.selectedStream = next
	m.streamLocked = false
	return fetchStreamCmd(m.audio, m.selectedStream)
}

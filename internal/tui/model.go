package tui

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/prettyletto/wave/internal/playerctl"
	"github.com/prettyletto/wave/internal/share"
	"github.com/prettyletto/wave/internal/state"
	"github.com/prettyletto/wave/internal/wpctl"
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

type Audio interface {
	Streams(context.Context) ([]wpctl.Stream, error)
	StreamByID(context.Context, string) (wpctl.Stream, error)
	ChangeVolume(context.Context, string, int) error
	ToggleMute(context.Context, string) error
}

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
type volumeMsg struct{ err error }
type muteMsg struct{ err error }

type tickMsg time.Time

type model struct {
	player Player
	audio  Audio

	now playerctl.TrackInfo
	err error

	players        []string
	selectedPlayer string

	streams        []wpctl.Stream
	selectedStream string
	currentStream  wpctl.Stream
}

func NewModel(p Player, a Audio) model {
	selectedPlayer, _ := state.LoadSelectedPlayer()

	return model{
		player:         p,
		audio:          a,
		selectedPlayer: selectedPlayer,
	}
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

func fetchStreamsCmd(a Audio) tea.Cmd {
	return func() tea.Msg {
		streams, err := a.Streams(context.Background())
		return streamsMsg{streams: streams, err: err}
	}
}

func fetchStreamCmd(a Audio, selectedStream string) tea.Cmd {
	return func() tea.Msg {
		if selectedStream == "" {
			return streamMsg{}
		}
		stream, err := a.StreamByID(context.Background(), selectedStream)
		return streamMsg{stream: stream, err: err}
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
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func volumeCmd(a Audio, selectedStream string, delta int) tea.Cmd {
	return func() tea.Msg {
		return volumeMsg{err: a.ChangeVolume(context.Background(), selectedStream, delta)}
	}
}

func muteCmd(a Audio, selectedStream string) tea.Cmd {
	return func() tea.Msg {
		return muteMsg{err: a.ToggleMute(context.Background(), selectedStream)}
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		fetchPlayersCmd(m.player),
		fetchNowCmd(m.player, m.selectedPlayer),
		fetchStreamsCmd(m.audio),
		fetchStreamCmd(m.audio, m.selectedStream),
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
			return m, tea.Batch(fetchNowCmd(m.player, m.selectedPlayer), m.syncSelectedStreamCmd())
		}

		if len(m.players) > 0 {
			m.selectedPlayer = m.players[0]
			_ = state.SaveSelectedPlayer(m.selectedPlayer)
			return m, tea.Batch(fetchNowCmd(m.player, m.selectedPlayer), m.syncSelectedStreamCmd())
		}

	case nowMsg:
		if msg.err == nil && m.selectedPlayer == "" && msg.info.Player != "" {
			m.selectedPlayer = msg.info.Player
			_ = state.SaveSelectedPlayer(m.selectedPlayer)
			return m, tea.Batch(fetchNowCmd(m.player, m.selectedPlayer), m.syncSelectedStreamCmd())
		}
		m.now = msg.info
		m.err = msg.err
		return m, nil
	case streamsMsg:
		m.streams = msg.streams
		m.err = msg.err
		if msg.err != nil {
			return m, nil
		}
		return m, m.syncSelectedStreamCmd()
	case streamMsg:
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
	case volumeMsg:
		m.err = msg.err
		if msg.err != nil {
			return m, nil
		}
		return m, fetchStreamCmd(m.audio, m.selectedStream)

	case muteMsg:
		m.err = msg.err
		if msg.err != nil {
			return m, nil
		}
		return m, fetchStreamCmd(m.audio, m.selectedStream)
	case tickMsg:
		return m, tea.Batch(
			fetchPlayersCmd(m.player),
			fetchNowCmd(m.player, m.selectedPlayer),
			fetchStreamsCmd(m.audio),
			fetchStreamCmd(m.audio, m.selectedStream),
			tickCmd())
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			next := nextPlayer(m.players, m.selectedPlayer, 1)
			if next == "" {
				return m, nil
			}
			m.selectedPlayer = next
			_ = state.SaveSelectedPlayer(m.selectedPlayer)
			return m, tea.Batch(fetchNowCmd(m.player, m.selectedPlayer), m.syncSelectedStreamCmd())
		case "shift+tab":
			next := nextPlayer(m.players, m.selectedPlayer, -1)
			if next == "" {
				return m, nil
			}
			m.selectedPlayer = next
			_ = state.SaveSelectedPlayer(m.selectedPlayer)
			return m, tea.Batch(fetchNowCmd(m.player, m.selectedPlayer), m.syncSelectedStreamCmd())
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
				return volumeCmd(m.audio, streamID, -5)
			})

		case "=":
			return m.withSelectedStream(func(streamID string) tea.Cmd {
				return volumeCmd(m.audio, streamID, 5)
			})

		case "m":
			return m.withSelectedStream(func(streamID string) tea.Cmd {
				return muteCmd(m.audio, streamID)
			})
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

	streamName := "-"
	volume := "-"
	muted := "no"

	if m.currentStream.ID != "" {
		streamName = m.currentStream.Name
		if streamName == "" {
			streamName = m.currentStream.AppName
		}
		volume = fmt.Sprintf("%.0f%%", m.currentStream.Volume*100)
		if m.currentStream.Muted {
			muted = "yes"
		}
	}

	return fmt.Sprintf(
		"wave\n\nPlayer: %s\nStatus: %s\nTitle: %s\nArtist: %s\nPosition: %d\nLength: %d\nProgress: %s\n\nAudio Stream: %s\nVolume: %s\nMuted: %s\n\n[tab] player  [-/=] volume  [m] mute  [space] toggle  [q] quit\n",
		m.now.Player,
		m.now.Status,
		m.now.Title,
		m.now.Artist,
		m.now.Position,
		m.now.LengthUS,
		share.ProgressLabel(m.now.Position, m.now.LengthUS),
		streamName,
		volume,
		muted,
	)
}

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

func nextStream(streams []wpctl.Stream, selected string, delta int) string {
	if len(streams) == 0 {
		return ""
	}

	i := selectedStreamIndex(streams, selected)
	if i == -1 {
		return streams[0].ID
	}

	i = (i + delta + len(streams)) % len(streams)
	return streams[i].ID
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

func (m *model) syncSelectedStreamCmd() tea.Cmd {
	next := matchStreamID(m.selectedPlayer, m.streams)
	if next == "" {
		m.selectedStream = ""
		m.currentStream = wpctl.Stream{}
		return nil
	}

	if next == m.selectedStream && containsStream(m.streams, next) {
		return fetchStreamCmd(m.audio, next)
	}

	m.selectedStream = next
	return fetchStreamCmd(m.audio, m.selectedStream)
}

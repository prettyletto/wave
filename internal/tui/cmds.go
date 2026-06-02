package tui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/prettyletto/wave/internal/playerctl"
)

func fetchPlayersCmd(p Player) tea.Cmd {
	return func() tea.Msg {
		players, err := p.Players(context.Background())
		if err != nil {
			return playersMsg{players: nil, err: err}
		}

		active := make([]string, 0, len(players))
		for _, player := range players {
			if _, err := p.NowForPlayer(context.Background(), player); err == nil {
				active = append(active, player)
			}
		}

		return playersMsg{players: active, err: nil}
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

func getShuffleCmd(p Player, selectedPlayer string) tea.Cmd {
	return func() tea.Msg {
		sf, err := p.ShuffleForPlayer(context.Background(), selectedPlayer)
		return shuffleMsg{shuffle: sf, err: err}
	}
}

func getLoopCmd(p Player, selectedPlayer string) tea.Cmd {
	return func() tea.Msg {
		loop, err := p.LoopForPlayer(context.Background(), selectedPlayer)
		return loopMsg{loop: loop, err: err}
	}
}

func toggleShuffleCmd(p Player, selectedPlayer string) tea.Cmd {
	return func() tea.Msg {
		err := p.ToggleShuffleForPlayer(context.Background(), selectedPlayer)
		return shuffleToggleMsg{err: err}
	}
}

func setLoopCmd(p Player, selectedPlayer string, loop playerctl.LoopStatus) tea.Cmd {
	return func() tea.Msg {
		err := p.SetLoopForPlayer(context.Background(), selectedPlayer, string(loop))
		return loopSetMsg{err: err}
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

func volumeCmd(a Audio, selectedStream string, currentVolume float64, delta int) tea.Cmd {
	return func() tea.Msg {
		return volumeMsg{err: a.ChangeVolume(context.Background(), selectedStream, currentVolume, delta)}
	}
}

func muteCmd(a Audio, selectedStream string) tea.Cmd {
	return func() tea.Msg {
		return muteMsg{err: a.ToggleMute(context.Background(), selectedStream)}
	}
}

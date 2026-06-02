package tui

import (
	"fmt"

	"github.com/prettyletto/wave/internal/share"
)

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("wave\n\nerror: %v\n\n[q] quit\n", m.err)
	}

	if len(m.players) == 0 && m.now.Player == "" {
		return "wave\n\nno player available right now\n\n[q] quit\n"
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

	shuffle := "-"
	switch {
	case m.shuffleUnsupported:
		shuffle = "unavailable"
	case m.shuffle != "":
		shuffle = string(m.shuffle)
	}

	loop := "-"
	switch {
	case m.loopUnsupported:
		loop = "unavailable"
	case m.loop != "":
		loop = string(m.loop)
	}

	help := "[tab] player  [ctrl+l] list  [r] repeat  [s] shuffle  [-/=] volume  [m] mute  [space] toggle  [q] quit"
	if m.listOpen {
		help = "[up/down] move  [left/right] focus  [enter] select  [esc] close  [q] quit"
	}

	base := fmt.Sprintf(
		"wave\n\nPlayer: %s\nArtWork: %s\nStatus: %s\nTitle: %s\nArtist: %s\nPosition: %d\nLength: %d\nProgress: %s\nRepeat: %s\nShuffle: %s\n\nAudio Stream: %s\nVolume: %s\nMuted: %s\n\n%s\n",
		m.now.Player,
		m.now.ArtUrl,
		m.now.Status,
		m.now.Title,
		m.now.Artist,
		m.now.Position,
		m.now.LengthUS,
		share.ProgressLabel(m.now.Position, m.now.LengthUS),
		loop,
		shuffle,
		streamName,
		volume,
		muted,
		help,
	)

	if m.listOpen {
		return base + m.listView()
	}

	return base
}

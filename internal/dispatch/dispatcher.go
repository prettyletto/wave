package dispatch

import (
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/prettyletto/wave/internal/playerctl"
	"github.com/prettyletto/wave/internal/tui"
	"github.com/prettyletto/wave/internal/wpctl"
)

type Player interface {
	Play(context.Context) error
	Pause(context.Context) error
	Toggle(context.Context) error
	ToggleForPlayer(context.Context, string) error
	Stop(context.Context) error
	Seek(context.Context, int) error
	SeekForPlayer(context.Context, string, int) error
	Next(context.Context) error
	Previous(context.Context) error
	NextForPlayer(context.Context, string) error
	PreviousForPlayer(context.Context, string) error

	Players(context.Context) ([]string, error)
	Status(context.Context) (playerctl.PlayStatus, error)
	MetaDataKey(context.Context, string) (string, error)
	Now(context.Context) (playerctl.TrackInfo, error)
	NowForPlayer(context.Context, string) (playerctl.TrackInfo, error)
	Volume(context.Context) (float64, error)
	SetVolume(context.Context, float64) error
	Loop(context.Context) (playerctl.LoopStatus, error)
	SetLoop(context.Context, string) error
	LoopForPlayer(context.Context, string) (playerctl.LoopStatus, error)
	SetLoopForPlayer(context.Context, string, string) error
	Shuffle(context.Context) (playerctl.ShuffleStatus, error)
	SetShuffle(context.Context, string) error
	ShuffleForPlayer(context.Context, string) (playerctl.ShuffleStatus, error)
	ToggleShuffleForPlayer(context.Context, string) error
}

type Audio interface {
	Streams(context.Context) ([]wpctl.Stream, error)
	StreamByID(context.Context, string) (wpctl.Stream, error)
	ChangeVolume(context.Context, string, int) error
	ToggleMute(context.Context, string) error
}

type Dispatcher struct {
	player Player
	audio  Audio
	out    io.Writer
}

func New(player Player, audio Audio, out io.Writer) *Dispatcher {
	return &Dispatcher{
		player: player,
		audio:  audio,
		out:    out,
	}
}

func (d *Dispatcher) Dispatch(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return d.tui()
	}

	switch args[0] {
	case "streams":
		return d.streams(ctx)
	case "help", "-h", "--help":
		return d.help()
	case "play":
		return d.player.Play(ctx)
	case "pause":
		return d.player.Pause(ctx)
	case "toggle":
		return d.player.Toggle(ctx)
	case "stop":
		return d.player.Stop(ctx)
	case "next":
		return d.player.Next(ctx)
	case "prev", "previous":
		return d.player.Previous(ctx)
	case "status":
		status, err := d.player.Status(ctx)
		if err != nil {
			return err
		}
		fmt.Fprintln(d.out, status)
		return nil
	case "volume":
		return d.volume(ctx, args[1:])
	case "loop":
		return d.loop(ctx, args[1:])
	case "shuffle":
		return d.shuffle(ctx, args[1:])
	case "players":
		return d.players(ctx)
	case "metadata":
		return d.metadata(ctx, args[1])
	case "now":
		return d.now(ctx)
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func (d *Dispatcher) now(ctx context.Context) error {
	ti, err := d.player.Now(ctx)
	if err != nil {
		return err
	}

	fmt.Fprintln(d.out, ti)
	return nil
}

func (d *Dispatcher) metadata(ctx context.Context, key string) error {
	pls, err := d.player.MetaDataKey(ctx, key)
	if err != nil {
		return err
	}

	fmt.Fprintln(d.out, pls)
	return nil
}

func (d *Dispatcher) volume(ctx context.Context, args []string) error {
	if len(args) == 0 {
		volume, err := d.player.Volume(ctx)
		if err != nil {
			return err
		}
		fmt.Fprintf(d.out, "%.2f\n", volume)
		return nil
	}
	volume, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return fmt.Errorf("invalid volume %q", args[0])
	}

	return d.player.SetVolume(ctx, volume)
}

func (d *Dispatcher) streams(ctx context.Context) error {
	pls, err := d.audio.Streams(ctx)
	if err != nil {
		return err
	}

	fmt.Fprintln(d.out, pls)
	return nil
}

func (d *Dispatcher) players(ctx context.Context) error {
	pls, err := d.player.Players(ctx)
	if err != nil {
		return err
	}

	fmt.Fprintln(d.out, pls)
	return nil
}

func (d *Dispatcher) loop(ctx context.Context, args []string) error {
	if len(args) == 0 {
		loop, err := d.player.Loop(ctx)
		if err != nil {
			return err
		}

		fmt.Fprintln(d.out, loop)
		return nil
	}

	return d.player.SetLoop(ctx, args[0])
}

func (d *Dispatcher) shuffle(ctx context.Context, args []string) error {
	if len(args) == 0 {
		shuffle, err := d.player.Shuffle(ctx)
		if err != nil {
			return err
		}

		fmt.Fprintln(d.out, shuffle)
		return nil
	}

	return d.player.SetShuffle(ctx, args[0])
}

func (d *Dispatcher) help() error {
	fmt.Fprintln(d.out, "usage: wave <command>")
	fmt.Fprintln(d.out)
	fmt.Fprintln(d.out, "commands:")
	fmt.Fprintln(d.out, "  status")
	fmt.Fprintln(d.out, "  metadata <key>")
	fmt.Fprintln(d.out, "  now")
	fmt.Fprintln(d.out, "  players")
	fmt.Fprintln(d.out, "  play")
	fmt.Fprintln(d.out, "  pause")
	fmt.Fprintln(d.out, "  toggle")
	fmt.Fprintln(d.out, "  stop")
	fmt.Fprintln(d.out, "  next")
	fmt.Fprintln(d.out, "  prev")
	fmt.Fprintln(d.out, "  volume [value]")
	fmt.Fprintln(d.out, "  loop [None|Track|Playlist]")
	fmt.Fprintln(d.out, "  shuffle [On|Off|Toggle]")
	return nil
}

func (d *Dispatcher) tui() error {
	return tui.Run(d.player, d.audio)
}

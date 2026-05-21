package dispatch

import (
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/prettyletto/wave/internal/playerctl"
)

type Player interface {
	Play(context.Context) error
	Pause(context.Context) error
	Toggle(context.Context) error
	Stop(context.Context) error
	Next(context.Context) error
	Previous(context.Context) error

	Players(context.Context) ([]string, error)
	Status(context.Context) (playerctl.PlayStatus, error)
	Volume(context.Context) (float64, error)
	SetVolume(context.Context, float64) error
	Loop(context.Context) (playerctl.LoopStatus, error)
	SetLoop(context.Context, string) error
	Shuffle(context.Context) (playerctl.ShuffleStatus, error)
	SetShuffle(context.Context, string) error
}

type Dispatcher struct {
	player Player
	out    io.Writer
}

func New(player Player, out io.Writer) *Dispatcher {
	return &Dispatcher{
		player: player,
		out:    out,
	}
}

func (d *Dispatcher) Dispatch(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return d.help()
	}

	switch args[0] {
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
	default:
		return fmt.Errorf("unkown command %q", args[0])
	}
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

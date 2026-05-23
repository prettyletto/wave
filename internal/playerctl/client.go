package playerctl

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type Runner interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}

type Client struct {
	runner Runner
	player string
}

type Option func(*Client)

func WithPlayer(name string) Option {
	return func(c *Client) { c.player = strings.TrimSpace(name) }
}

func New(r Runner, opts ...Option) *Client {
	if r == nil {
		r = ExecRunner{}
	}

	c := &Client{runner: r}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) args(base ...string) []string {
	if c.player == "" {
		return base
	}

	return append([]string{"--player", c.player}, base...)
}

func (c *Client) run(ctx context.Context, args ...string) (string, error) {
	out, err := c.runner.Run(ctx, "playerctl", c.args(args...)...)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

func (c *Client) runForPlayer(ctx context.Context, player string, args ...string) (string, error) {
	player = strings.TrimSpace(player)
	if player == "" {
		return c.run(ctx, args...)
	}

	fullArgs := append([]string{"--player", player}, args...)

	out, err := c.runner.Run(ctx, "playerctl", fullArgs...)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

func (c *Client) ToggleForPlayer(ctx context.Context, player string) error {
	_, err := c.runForPlayer(ctx, player, "play-pause")
	return err
}

func (c *Client) SeekForPlayer(ctx context.Context, player string, seconds int) error {
	sign := "+"

	if seconds < 0 {
		sign = "-"
		seconds = -seconds
	}

	_, err := c.runForPlayer(ctx, player, "position", fmt.Sprintf("%d%s", seconds, sign))
	return err
}

func (c *Client) Players(ctx context.Context) ([]string, error) {
	out, err := c.run(ctx, "-l")
	if err != nil {
		return []string{}, err
	}

	if out == "" {
		return []string{}, nil
	}

	players := strings.Split(out, "\n")
	filtered := players[:0]
	for _, player := range players {
		player = strings.TrimSpace(player)
		if player != "" {
			filtered = append(filtered, player)
		}
	}

	return filtered, nil
}

func (c *Client) Play(ctx context.Context) error {
	_, err := c.run(ctx, "play")
	return err
}

func (c *Client) Pause(ctx context.Context) error {
	_, err := c.run(ctx, "pause")
	return err
}

func (c *Client) Toggle(ctx context.Context) error {
	_, err := c.run(ctx, "play-pause")
	return err
}

func (c *Client) Position(ctx context.Context) error {
	_, err := c.run(ctx, "position")
	return err
}

func (c *Client) Stop(ctx context.Context) error {
	_, err := c.run(ctx, "stop")
	return err
}

func (c *Client) Next(ctx context.Context) error {
	_, err := c.run(ctx, "next")
	return err
}

func (c *Client) Previous(ctx context.Context) error {
	_, err := c.run(ctx, "previous")
	return err
}

func (c *Client) SetPosition(ctx context.Context, seconds int) error {
	_, err := c.run(ctx, "position", strconv.Itoa(seconds))

	return err
}

func (c *Client) Seek(ctx context.Context, seconds int) error {
	sign := "+"

	if seconds < 0 {
		sign = "-"
		seconds = -seconds
	}

	_, err := c.run(ctx, "position", fmt.Sprintf("%d%s", seconds, sign))
	return err
}

func (c *Client) SetVolume(ctx context.Context, volume float64) error {
	_, err := c.run(ctx, "volume", fmt.Sprintf("%.2f", volume))
	return err
}

func (c *Client) Volume(ctx context.Context) (float64, error) {
	out, err := c.run(ctx, "volume")
	if err != nil {
		return 0, err
	}

	return strconv.ParseFloat(out, 64)
}

func (c *Client) NowForPlayer(ctx context.Context, player string) (TrackInfo, error) {
	const format = "{{playerName}}\t{{status}}\t{{xesam:title}}\t{{xesam:artist}}\t{{xesam:album}}\t{{mpris:length}}"

	out, err := c.runForPlayer(ctx, player, "metadata", "--format", format)
	if err != nil {
		return TrackInfo{}, err
	}

	info, err := ParseTrackInfo(out)
	if err != nil {
		return TrackInfo{}, err
	}

	posOut, err := c.runForPlayer(ctx, info.Player, "position")
	if err == nil {
		if sec, convErr := strconv.ParseFloat(strings.TrimSpace(posOut), 64); convErr == nil {
			info.Position = int64(sec * 1_000_000)
		}
	}

	return info, nil
}

func (c *Client) Now(ctx context.Context) (TrackInfo, error) {
	return c.NowForPlayer(ctx, "")
}

func (c *Client) MetaDataKey(ctx context.Context, key string) (string, error) {
	out, err := c.run(ctx, "metadata", key)
	return out, err
}

func (c *Client) Status(ctx context.Context) (PlayStatus, error) {
	out, err := c.run(ctx, "status")
	if err != nil {
		return "", err
	}

	return ParsePlayStatus(out)
}

func (c *Client) Loop(ctx context.Context) (LoopStatus, error) {
	out, err := c.run(ctx, "loop")
	if err != nil {
		return "", err
	}

	return ParseLoopStatus(out)
}

func (c *Client) SetLoop(ctx context.Context, s string) error {
	l, err := ParseLoopStatus(s)
	if err != nil {
		return err
	}

	_, err = c.run(ctx, "loop", string(l))
	return err
}

func (c *Client) Shuffle(ctx context.Context) (ShuffleStatus, error) {
	out, err := c.run(ctx, "shuffle")
	if err != nil {
		return "", err
	}
	return ParseShuffleStatus(out)
}

func (c *Client) SetShuffle(ctx context.Context, s string) error {
	sf, err := ParseShuffleStatus(s)
	if err != nil {
		return err
	}

	_, err = c.run(ctx, "shuffle", string(sf))
	return err
}

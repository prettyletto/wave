package wpctl

import (
	"context"
	"os/exec"
	"strings"
)

type Runner interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}

type Stream struct {
	ID      string
	Name    string
	AppName string
	Binary  string
	Volume  float64
	Muted   bool
}

type Client struct {
	runner Runner
}

func New(r Runner) *Client {
	if r == nil {
		r = ExecRunner{}
	}

	return &Client{runner: r}
}

func (c *Client) run(ctx context.Context, args ...string) (string, error) {
	out, err := c.runner.Run(ctx, "wpctl", args...)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

func (c *Client) Streams(ctx context.Context) ([]Stream, error)

func (c *Client) SetVolume(ctx context.Context, streamID string, volume float64) error

func (c *Client) ChangeVolume(ctx context.Context, streamID string, deltaPercent int) error

func (c *Client) ToggleMute(ctx context.Context, streamID string) error

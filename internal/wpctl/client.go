package wpctl

import (
	"context"
	"encoding/json"
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

func (c *Client) run(ctx context.Context, name string, args ...string) (string, error) {
	out, err := c.runner.Run(ctx, name, args...)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

func (c *Client) Streams(ctx context.Context) ([]Stream, error) {
	streams, err := c.streamsFromDump(ctx)
	if err != nil {
		return nil, err
	}

	for i := range streams {
		volume, muted, err := c.streamVolume(ctx, streams[i].ID)
		if err != nil {
			continue
		}
		streams[i].Volume = volume
		streams[i].Muted = muted
	}

	return streams, nil
}

func (c *Client) StreamByID(ctx context.Context, id string) (Stream, error) {
	streams, err := c.streamsFromDump(ctx)
	if err != nil {
		return Stream{}, err
	}

	for _, stream := range streams {
		if stream.ID != id {
			continue
		}

		volume, muted, err := c.streamVolume(ctx, id)
		if err != nil {
			return Stream{}, err
		}

		stream.Volume = volume
		stream.Muted = muted
		return stream, nil
	}

	return Stream{}, fmt.Errorf("stream %s not found", id)
}

func (c *Client) SetVolume(ctx context.Context, streamID string, volume float64) error {
	_, err := c.run(ctx, "wpctl", "set-volume", streamID, fmt.Sprintf("%.2f", volume))
	return err
}

func (c *Client) ChangeVolume(ctx context.Context, streamID string, deltaPercent int) error {
	sign := "+"
	if deltaPercent < 0 {
		sign = "-"
		deltaPercent = -deltaPercent
	}

	_, err := c.run(ctx, "wpctl", "set-volume", streamID, fmt.Sprintf("%d%%%s", deltaPercent, sign))
	return err
}

func (c *Client) ToggleMute(ctx context.Context, streamID string) error {
	_, err := c.run(ctx, "wpctl", "set-mute", streamID, "toggle")
	return err
}

func (c *Client) streamVolume(ctx context.Context, id string) (float64, bool, error) {
	out, err := c.run(ctx, "wpctl", "get-volume", id)
	if err != nil {
		return 0, false, fmt.Errorf("get stream volume %s: %w", id, err)
	}

	return parseVolume(out)
}

func (c *Client) streamsFromDump(ctx context.Context) ([]Stream, error) {
	out, err := c.run(ctx, "pw-dump")
	if err != nil {
		return nil, err
	}

	var objects []pwObject
	if err := json.Unmarshal([]byte(out), &objects); err != nil {
		return nil, fmt.Errorf("parse pw-dump: %w", err)
	}

	clients := make(map[int]pwProps)
	for _, obj := range objects {
		if obj.Type != "PipeWire:Interface:Client" || obj.Info == nil {
			continue
		}
		clients[obj.ID] = obj.Info.Props
	}

	streams := make([]Stream, 0)
	for _, obj := range objects {
		if obj.Type != "PipeWire:Interface:Node" || obj.Info == nil {
			continue
		}

		props := obj.Info.Props
		if props.String("media.class") != "Stream/Output/Audio" {
			continue
		}

		clientProps := clients[props.Int("client.id")]
		stream := Stream{
			ID:      strconv.Itoa(obj.ID),
			Name:    preferredName(props, clientProps),
			AppName: preferredAppName(props, clientProps),
			Binary:  preferredBinary(props, clientProps),
		}

		if stream.Name == "" && stream.AppName == "" && stream.Binary == "" {
			continue
		}

		streams = append(streams, stream)
	}

	return streams, nil
}

func preferredName(nodeProps, clientProps pwProps) string {
	for _, candidate := range []string{
		nodeProps.String("application.name"),
		clientProps.String("application.name"),
		nodeProps.String("node.name"),
		nodeProps.String("node.description"),
		nodeProps.String("media.name"),
		clientProps.String("media.name"),
	} {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" || isGenericName(candidate) {
			continue
		}
		return candidate
	}

	return ""
}

func preferredAppName(nodeProps, clientProps pwProps) string {
	for _, candidate := range []string{
		nodeProps.String("application.name"),
		clientProps.String("application.name"),
		nodeProps.String("node.name"),
		nodeProps.String("media.name"),
		clientProps.String("media.name"),
	} {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" || isGenericName(candidate) {
			continue
		}
		return candidate
	}

	return ""
}

func preferredBinary(nodeProps, clientProps pwProps) string {
	for _, candidate := range []string{
		nodeProps.String("application.process.binary"),
		clientProps.String("application.process.binary"),
	} {
		candidate = strings.TrimSpace(candidate)
		if candidate != "" {
			return candidate
		}
	}

	return ""
}

func isGenericName(value string) bool {
	value = strings.TrimSpace(strings.ToLower(value))
	return value == "" || value == "playback" || strings.HasPrefix(value, "output_")
}

func parseVolume(out string) (float64, bool, error) {
	fields := strings.Fields(out)
	if len(fields) < 2 {
		return 0, false, fmt.Errorf("unexpected volume output %q", out)
	}

	volume, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return 0, false, fmt.Errorf("invalid volume value %q: %w", fields[1], err)
	}

	return volume, strings.Contains(out, "[MUTED]"), nil
}

type pwObject struct {
	ID   int           `json:"id"`
	Type string        `json:"type"`
	Info *pwObjectInfo `json:"info"`
}

type pwObjectInfo struct {
	Props pwProps `json:"props"`
}

type pwProps map[string]any

func (p pwProps) String(key string) string {
	if p == nil {
		return ""
	}

	value, ok := p[key]
	if !ok {
		return ""
	}

	s, _ := value.(string)
	return s
}

func (p pwProps) Int(key string) int {
	if p == nil {
		return 0
	}

	value, ok := p[key]
	if !ok {
		return 0
	}

	switch v := value.(type) {
	case float64:
		return int(v)
	case int:
		return v
	default:
		return 0
	}
}

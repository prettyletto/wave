package wpctl

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
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

func (c *Client) run(ctx context.Context, args ...string) (string, error) {
	out, err := c.runner.Run(ctx, "wpctl", args...)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

func (c *Client) Streams(ctx context.Context) ([]Stream, error) {
	out, err := c.run(ctx, "status")
	if err != nil {
		return nil, err
	}

	ids := parseStatusStreamIDs(out)
	streams := make([]Stream, 0, len(ids))

	for _, id := range ids {
		stream, err := c.streamByID(ctx, id)
		if err != nil {
			continue
		}
		streams = append(streams, stream)
	}

	return streams, nil
}

func (c *Client) streamByID(ctx context.Context, id string) (Stream, error) {
	inspectOut, err := c.run(ctx, "inspect", id)
	if err != nil {
		return Stream{}, fmt.Errorf("inspect stream %s: %w", id, err)
	}

	volumeOut, err := c.run(ctx, "get-volume", id)
	if err != nil {
		return Stream{}, fmt.Errorf("get stream volume %s: %w", id, err)
	}

	stream := parseInspect(id, inspectOut)

	volume, muted, err := parseVolume(volumeOut)
	if err != nil {
		return Stream{}, fmt.Errorf("parse stream volume %s: %w", id, err)
	}

	stream.Volume = volume
	stream.Muted = muted

	return stream, nil
}

func (c *Client) SetVolume(ctx context.Context, streamID string, volume float64) error {
	_, err := c.run(ctx, "set-volume", streamID, fmt.Sprintf("%.2f", volume))
	return err
}

func (c *Client) ChangeVolume(ctx context.Context, streamID string, deltaPercent int) error {
	sign := "+"

	if deltaPercent < 0 {
		sign = "-"
		deltaPercent = -deltaPercent
	}

	_, err := c.run(ctx, "set-volume", streamID, fmt.Sprintf("%d%%%s", deltaPercent, sign))
	return err
}

func (c *Client) ToggleMute(ctx context.Context, streamID string) error {
	_, err := c.run(ctx, "set-mute", streamID, "toggle")
	return err
}

func parseStatusStreamIDs(out string) []string {
	lines := strings.Split(out, "\n")
	ids := make([]string, 0)

	inAudio := false
	inStreams := false
	streamIndent := -1

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if trimmed == "Audio" {
			inAudio = true
			inStreams = false
			streamIndent = -1
			continue
		}

		if trimmed == "Video" {
			break
		}

		if !inAudio {
			continue
		}

		if strings.Contains(trimmed, "Streams:") {
			inStreams = true
			streamIndent = -1
			continue
		}

		if !inStreams {
			continue
		}

		if isAudioSubsectionHeader(trimmed) {
			break
		}

		id, ok := parseLeadingID(trimmed)
		if !ok {
			continue
		}

		indent := leadingWhitespaceWidth(line)

		if streamIndent == -1 {
			streamIndent = indent
			ids = append(ids, id)
			continue
		}

		if indent == streamIndent {
			ids = append(ids, id)
		}
	}

	return ids
}

func parseInspect(id, out string) Stream {
	stream := Stream{
		ID: id,
	}

	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(line, "node.description = "):
			stream.Name = unquoteValue(line)
		case strings.HasPrefix(line, "node.nick = "):
			if stream.Name == "" {
				stream.Name = unquoteValue(line)
			}
		case strings.HasPrefix(line, "media.name = "):
			if stream.Name == "" {
				stream.Name = unquoteValue(line)
			}
		case strings.HasPrefix(line, "application.name = "):
			stream.AppName = unquoteValue(line)
		case strings.HasPrefix(line, "application.process.binary = "):
			stream.Binary = unquoteValue(line)
		}
	}

	if stream.Name == "" {
		stream.Name = stream.AppName
	}
	if stream.Name == "" {
		stream.Name = stream.Binary
	}

	return stream
}

func isAudioSubsectionHeader(trimmed string) bool {
	return strings.Contains(trimmed, "Devices:") ||
		strings.Contains(trimmed, "Sinks:") ||
		strings.Contains(trimmed, "Sources:") ||
		strings.Contains(trimmed, "Filters:")
}

func leadingWhitespaceWidth(s string) int {
	count := 0
	for _, r := range s {
		if r == ' ' || r == '\t' {
			count++
			continue
		}
		break
	}
	return count
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

	muted := strings.Contains(out, "[MUTED]")

	return volume, muted, nil
}

var leadingIDPattern = regexp.MustCompile(`(?:^|[^\d])(\d+)\.`)

func parseLeadingID(line string) (string, bool) {
	matches := leadingIDPattern.FindStringSubmatch(line)
	if len(matches) != 2 {
		return "", false
	}

	return matches[1], true
}

func unquoteValue(line string) string {
	idx := strings.Index(line, "=")
	if idx == -1 {
		return ""
	}

	value := strings.TrimSpace(line[idx+1:])
	value = strings.Trim(value, "\"")
	return value
}

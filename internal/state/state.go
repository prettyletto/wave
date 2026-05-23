package state

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const selectedPlayerFile = "selected-player"

func dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(base, "wave"), nil
}

func LoadSelectedPlayer() (string, error) {
	dir, err := dir()
	if err != nil {
		return "", err
	}

	b, err := os.ReadFile(filepath.Join(dir, selectedPlayerFile))
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(b)), nil
}

func SaveSelectedPlayer(player string) error {
	player = strings.TrimSpace(player)
	if player == "" {
		return nil
	}

	dir, err := dir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, selectedPlayerFile), []byte(player+"\n"), 0o644)
}

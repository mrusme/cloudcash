package lib

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

func Secret(ctx context.Context, value string, command string) (string, error) {
	if value != "" || command == "" {
		return value, nil
	}

	out, err := shellCommand(ctx, command).Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && len(bytes.TrimSpace(exitErr.Stderr)) > 0 {
			return "", fmt.Errorf("%q: %w: %s", command, err, bytes.TrimSpace(exitErr.Stderr))
		}
		return "", fmt.Errorf("%q: %w", command, err)
	}

	line, _, _ := strings.Cut(string(out), "\n")
	secret := strings.TrimSpace(line)
	if secret == "" {
		return "", fmt.Errorf("%q: no output", command)
	}

	return secret, nil
}

package shell

import (
	"context"
	"errors"
	"os/exec"
	"strings"
)

// Runner executes shell commands.
type Runner interface {
	Run(ctx context.Context, command string) (string, error)
}

// Shell runs commands via /bin/sh -c.
type Shell struct{}

func (Shell) Run(ctx context.Context, command string) (string, error) {
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", command)
	raw, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(raw))
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return output, ctx.Err()
	}
	return output, err
}

// Quote wraps a value for safe use in shell commands.
func Quote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

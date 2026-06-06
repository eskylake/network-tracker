package check

import (
	"context"
	"time"

	"github.com/eskylake/network-tracker/internal/shell"
)

func CommandChecker(name, category, command string, ok func(string) bool, classify func(string, error) (Status, string)) Checker {
	return commandChecker{
		name:     name,
		category: category,
		command:  command,
		ok:       ok,
		classify: classify,
		runner:   shell.Shell{},
	}
}

type commandChecker struct {
	name     string
	category string
	command  string
	ok       func(string) bool
	classify func(string, error) (Status, string)
	runner   shell.Runner
}

func (c commandChecker) Name() string     { return c.name }
func (c commandChecker) Category() string { return c.category }

func (c commandChecker) Run(ctx context.Context) Result {
	start := time.Now()
	output, err := c.runner.Run(ctx, c.command)
	if c.classify != nil {
		status, summary := c.classify(output, err)
		return finished(c.name, c.category, status, summary, output, start, err)
	}
	if err != nil {
		return finished(c.name, c.category, StatusWarning, "command failed", outputOrError(output, err), start, err)
	}
	if c.ok == nil || c.ok(output) {
		return finished(c.name, c.category, StatusOK, firstLine(output, "ok"), output, start, nil)
	}
	return finished(c.name, c.category, StatusWarning, firstLine(output, "unexpected output"), output, start, nil)
}

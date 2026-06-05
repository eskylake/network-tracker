package checks

import (
	"context"
	"errors"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"
)

type Status string

const (
	StatusOK      Status = "ok"
	StatusWarning Status = "warning"
	StatusError   Status = "error"
	StatusUnknown Status = "unknown"
)

type Result struct {
	Name       string
	Category   string
	Status     Status
	Summary    string
	Details    string
	StartedAt  time.Time
	FinishedAt time.Time
	Duration   time.Duration
	Err        error
}

type Checker interface {
	Name() string
	Category() string
	Run(context.Context) Result
}

type CheckFunc struct {
	CheckName     string
	CheckCategory string
	Fn            func(context.Context) Result
}

func (c CheckFunc) Name() string     { return c.CheckName }
func (c CheckFunc) Category() string { return c.CheckCategory }
func (c CheckFunc) Run(ctx context.Context) Result {
	start := time.Now()
	result := c.Fn(ctx)
	if result.Name == "" {
		result.Name = c.CheckName
	}
	if result.Category == "" {
		result.Category = c.CheckCategory
	}
	if result.StartedAt.IsZero() {
		result.StartedAt = start
	}
	if result.FinishedAt.IsZero() {
		result.FinishedAt = time.Now()
	}
	result.Duration = result.FinishedAt.Sub(result.StartedAt)
	return result
}

type Runner struct {
	Timeout time.Duration
	Checks  []Checker
}

func (r Runner) Run(ctx context.Context) []Result {
	results := make([]Result, 0, len(r.Checks))
	out := make(chan Result, len(r.Checks))
	var wg sync.WaitGroup

	for _, checker := range r.Checks {
		checker := checker
		wg.Add(1)
		go func() {
			defer wg.Done()
			checkCtx, cancel := context.WithTimeout(ctx, r.Timeout)
			defer cancel()
			out <- checker.Run(checkCtx)
		}()
	}

	wg.Wait()
	close(out)

	for result := range out {
		results = append(results, result)
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Category == results[j].Category {
			return results[i].Name < results[j].Name
		}
		return results[i].Category < results[j].Category
	})
	return results
}

func NewResult(name, category string, status Status, summary, details string, err error) Result {
	now := time.Now()
	return Result{
		Name:       name,
		Category:   category,
		Status:     status,
		Summary:    summary,
		Details:    strings.TrimSpace(details),
		StartedAt:  now,
		FinishedAt: now,
		Err:        err,
	}
}

type CommandRunner interface {
	Run(ctx context.Context, command string) (string, error)
}

type ShellCommandRunner struct{}

func (ShellCommandRunner) Run(ctx context.Context, command string) (string, error) {
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", command)
	raw, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(raw))
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return output, ctx.Err()
	}
	return output, err
}

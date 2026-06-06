package check

import (
	"context"
	"strings"
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

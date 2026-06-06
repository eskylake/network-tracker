package check

import (
	"errors"
	"strings"
	"time"
)

func finished(name, category string, status Status, summary, details string, start time.Time, err error) Result {
	return Result{
		Name:       name,
		Category:   category,
		Status:     status,
		Summary:    summary,
		Details:    strings.TrimSpace(details),
		StartedAt:  start,
		FinishedAt: time.Now(),
		Duration:   time.Since(start),
		Err:        err,
	}
}

func firstLine(output, fallback string) string {
	output = strings.TrimSpace(output)
	if output == "" {
		return fallback
	}
	return strings.Split(output, "\n")[0]
}

func outputOrError(output string, err error) string {
	output = strings.TrimSpace(output)
	if output != "" {
		return output
	}
	if err != nil {
		return err.Error()
	}
	return ""
}

func stringErrors(values []string) []error {
	errs := make([]error, 0, len(values))
	for _, value := range values {
		if value != "" {
			errs = append(errs, errors.New(value))
		}
	}
	return errs
}

func unique(values []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, value := range values {
		if !seen[value] {
			seen[value] = true
			out = append(out, value)
		}
	}
	return out
}

func nonEmpty(values ...string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

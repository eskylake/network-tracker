package check

import (
	"context"
	"sort"
	"sync"
)

func MergeResults(current, updates []Result) []Result {
	merged := append([]Result{}, current...)
	for _, update := range updates {
		replaced := false
		for i, result := range merged {
			if result.Name == update.Name {
				merged[i] = update
				replaced = true
				break
			}
		}
		if !replaced {
			merged = append(merged, update)
		}
	}
	sortResults(merged)
	return merged
}

func sortResults(results []Result) {
	sort.Slice(results, func(i, j int) bool {
		if results[i].Category == results[j].Category {
			return results[i].Name < results[j].Name
		}
		return results[i].Category < results[j].Category
	})
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
	sortResults(results)
	return results
}

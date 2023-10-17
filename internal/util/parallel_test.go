package util_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/easterthebunny/automation-cli/internal/util"
)

func TestRunWithContext(t *testing.T) {
	jobFunc := func(ctx context.Context, output int) int {
		timer := time.NewTimer(50 * time.Millisecond)

		select {
		case <-timer.C:
			return output
		case <-ctx.Done():
			timer.Stop()
			return 0
		}
	}

	p := util.NewParallel[int](10)
	jobs := []util.Job[int]{}
	expected := 0

	for i := 1; i <= 20; i++ {
		expected += i

		out := i

		jobs = append(jobs, func(ctx context.Context) int {
			return jobFunc(ctx, out)
		})
	}

	results := p.RunWithContext(context.Background(), jobs)

	assert.Len(t, results, 20)

	var total int
	for _, result := range results {
		total += result
	}

	assert.Equal(t, expected, total)
}

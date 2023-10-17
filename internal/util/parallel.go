package util

import (
	"context"
	"sync"
)

type Job[T any] func(context.Context) T

type Parallel[T any] struct {
	workers chan *worker[T]
	limit   int
	active  int
	mu      sync.Mutex
	wg      sync.WaitGroup
	results []T
}

func NewParallel[T any](parallelization int) *Parallel[T] {
	return &Parallel[T]{
		workers: make(chan *worker[T], parallelization),
		limit:   parallelization,
		results: make([]T, 0),
	}
}

func (p *Parallel[T]) RunWithContext(ctx context.Context, jobs []Job[T]) []T {
	for i := range jobs {
		if ctx.Err() != nil {
			break
		}

		var wkr *worker[T]

		if p.active < p.limit {
			wkr = &worker[T]{
				queue: p.workers,
			}
			p.active++
		} else {
			select {
			case wkr = <-p.workers:
			case <-ctx.Done():
				break
			}
		}

		p.wg.Add(1)

		go p.doJob(ctx, jobs[i], wkr)
	}

	p.wg.Wait()

	return p.results
}

func (p *Parallel[T]) aggregate(result T) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.results = append(p.results, result)
}

func (p *Parallel[T]) doJob(ctx context.Context, job Job[T], wkr *worker[T]) {
	wkr.do(ctx, job, p.aggregate)
	p.wg.Done()
}

type worker[T any] struct {
	queue chan *worker[T]
}

func (w *worker[T]) do(ctx context.Context, wrk Job[T], f func(T)) {
	var result T

	if ctx.Err() != nil {
		return
	} else {
		result = wrk(ctx)
	}

	f(result)

	select {
	case w.queue <- w:
	default:
	}
}

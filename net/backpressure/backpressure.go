// Package backpressure provides a generic, high‑performance pipeline with configurable
// parallelism and built‑in back‑pressure via buffered channels.
package backpressure

import (
	"context"
	"sync"
)

// Stage is a single processing step. It receives an item and returns an error.
// If the error is non‑nil, the pipeline aborts early.
type Stage[T any] func(ctx context.Context, item T) error

// Option is a functional option for Pipeline.
type Option[T any] func(*Pipeline[T])

// WithWorkers sets the number of concurrent workers for each stage.
// Default is 1.
func WithWorkers[T any](workers int) Option[T] {
	return func(p *Pipeline[T]) {
		if workers < 1 {
			workers = 1
		}
		p.workers = workers
	}
}

// WithBufferSize sets the capacity of the channel between stages.
// Default is 0 (unbuffered).
func WithBufferSize[T any](size int) Option[T] {
	return func(p *Pipeline[T]) {
		if size < 0 {
			size = 0
		}
		p.bufSize = size
	}
}

// Pipeline connects a sequence of Stages with buffered channels and manages their execution.
type Pipeline[T any] struct {
	stages  []Stage[T]
	workers int
	bufSize int
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewPipeline creates a new pipeline with the given stages and options.
func NewPipeline[T any](stages []Stage[T], opts ...Option[T]) *Pipeline[T] {
	ctx, cancel := context.WithCancel(context.Background())
	p := &Pipeline[T]{
		stages:  stages,
		workers: 1,
		ctx:     ctx,
		cancel:  cancel,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Run starts the pipeline. It returns two channels:
//   - input:  the caller sends items into this channel
//   - output: the pipeline emits processed items on this channel
//
// The pipeline stops when input is closed or a stage returns an error.
func (p *Pipeline[T]) Run() (input chan<- T, output <-chan T) {
	in := make(chan T, p.bufSize)
	out := make(chan T, p.bufSize)

	current := in
	for _, stage := range p.stages {
		next := make(chan T, p.bufSize)
		p.startStage(current, next, stage)
		current = next
	}
	// last stage writes to output channel
	p.startStage(current, out, nil)

	// close the final output when all workers are done
	go func() {
		p.wg.Wait()
		close(out)
	}()

	return in, out
}

// startStage launches workers that read from `in`, apply `stage` (if non‑nil),
// and write to `out`. If stage returns an error, the pipeline is cancelled.
func (p *Pipeline[T]) startStage(in <-chan T, out chan<- T, stage Stage[T]) {
	p.wg.Add(p.workers)
	for i := 0; i < p.workers; i++ {
		go func() {
			defer p.wg.Done()
			for {
				select {
				case <-p.ctx.Done():
					return
				case item, ok := <-in:
					if !ok {
						return
					}
					if stage != nil {
						if err := stage(p.ctx, item); err != nil {
							p.cancel()
							return
						}
					}
					// forward item
					select {
					case out <- item:
					case <-p.ctx.Done():
						return
					}
				}
			}
		}()
	}
}

// Wait blocks until the pipeline finishes.
func (p *Pipeline[T]) Wait() {
	p.wg.Wait()
}

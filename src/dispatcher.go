package main

import (
	"sync"
)

// A pool of workers channels that are registered with the dispatcher.
type Dispatcher struct {
	WorkerPool    chan chan *Job
	Workers       []*Worker
	MaxWorkers    int
	WorkerOptions *WorkerOptions
}

// Creates a new dispatcher to handle new job requests
func NewDispatcher(maxWorkers int, options *WorkerOptions) *Dispatcher {
	pool := make(chan chan *Job, maxWorkers)
	return &Dispatcher{
		WorkerPool:    pool,
		WorkerOptions: options,
		MaxWorkers:    maxWorkers}
}

// Returns the buffer size for a single worker
func (d *Dispatcher) GetBufferSize(n int) int {
	if !d.WorkerOptions.SpreadBuffer {
		return d.WorkerOptions.BufferSize
	}
	slizeSize := int(d.WorkerOptions.BufferSize / (2 * (d.MaxWorkers - 1)))
	return int(float32(d.WorkerOptions.BufferSize)*0.75) + (n * slizeSize)
}

// Creates and starts the workers
func (d *Dispatcher) Start() {
	for i := 0; i < d.MaxWorkers; i++ {
		options := &WorkerOptions{
			BufferSize:   d.GetBufferSize(i),
			RetryAttempt: d.WorkerOptions.RetryAttempt}

		// Create a new worker
		worker := NewWorker(i, options, d.WorkerPool)
		worker.Start()

		// Add the worker into the list
		d.Workers = append(d.Workers, worker)
	}
}

// Creates and starts the workers and listen for new job requests
func (d *Dispatcher) Run() {
	d.Start()
	go d.dispatch()
}

// Stops all the workers
func (d *Dispatcher) Stop() {
	var wg sync.WaitGroup
	for i := range d.Workers {
		wg.Add(1)
		d.Workers[i].Stop(&wg)
	}
	wg.Wait()
}

// Listening for new job requests
func (d *Dispatcher) dispatch() {
	for {
		select {
		case job := <-jobQueue:
			select {
			case jobChannel := <-d.WorkerPool:
				jobChannel <- job
			}
		}
	}
}

package jobs

import (
	"errors"

	"github.com/gocraft/work"
)

type (
	// WorkerPoolContext struct
	WorkerPoolContext struct{}

	// Worker struct
	Worker struct {
		pool *work.WorkerPool
	}
)

// NewWorker is a factory of the Worker
func NewWorker(pool *work.WorkerPool) *Worker {
	return &Worker{pool: pool}
}

// RegisterJobs in worker pool
// Docs: github.com/gocraft/work
func (c *Worker) RegisterJobs() error {
	if c.pool == nil {
		return errors.New("worker pool is not running")
	}

	// Set up worker pool handlers
	c.pool.JobWithOptions(QueueSendEmail, work.JobOptions{Priority: 1000, MaxFails: 5}, c.SendEmail)

	return nil
}

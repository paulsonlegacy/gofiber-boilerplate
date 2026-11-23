package queue

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// DurableQueue is the default queue implementation using a buffered channel.
// For production workloads with persistence requirements, consider replacing the
// channel with a Redis stream or a DB-backed queue using the same Queue interface.
type DurableQueue struct {
	jobs chan Job
	db   *gorm.DB // reserved for future durable job persistence
}

// NewDurableQueue creates a new queue with a channel of the given buffer size.
// The context controls the lifecycle — when it's cancelled, the channel is closed.
func NewDurableQueue(ctx context.Context, db *gorm.DB, size int) *DurableQueue {
	fmt.Println("[Queue] Initializing job queue...")

	q := &DurableQueue{
		jobs: make(chan Job, size),
		db:   db,
	}

	go func() {
		<-ctx.Done()
		fmt.Println("[Queue] Context cancelled — closing job queue")
		close(q.jobs)
	}()

	return q
}

// Push enqueues a job. Returns an error if the channel buffer is full (drops the job).
// For critical jobs, consider retrying on error or using a persistent queue instead.
func (q *DurableQueue) Push(job Job) error {
	select {
	case q.jobs <- job:
		return nil
	default:
		fmt.Printf("[Queue] FULL — dropped job type=%s\n", job.Type)
		return fmt.Errorf("queue is full: unable to enqueue job of type %s", job.Type)
	}
}

// GetQueue returns the underlying channel for workers to range over.
func (q *DurableQueue) GetQueue() chan Job {
	return q.jobs
}

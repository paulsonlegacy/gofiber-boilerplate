// Package queue provides a durable in-process job queue backed by channels.
// Jobs are pushed by usecases/services and consumed by workers running in the background.
// The queue interface allows swapping in Redis, RabbitMQ, or SQS in the future
// without changing any calling code — just implement the Queue interface.
package queue

// Queue is the contract all queue implementations must satisfy.
type Queue interface {
	Push(job Job) error
	GetQueue() chan Job
}

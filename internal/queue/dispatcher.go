package queue

import (
	"context"
	"sync"
)

type PaymentJob struct {
	CorrelationID string
	Amount        float64
	Attempts      int
}

type Worker interface {
	ProcessPayment(job *PaymentJob)
}

type Dispatcher struct {
	workerPool chan struct{}
	jobQueue   chan *PaymentJob
	worker     Worker
	globalWg   sync.WaitGroup
}

func NewDispatcher(worker Worker, maxWorkers int, maxQueueSize int) *Dispatcher {
	return &Dispatcher{
		workerPool: make(chan struct{}, maxWorkers),
		jobQueue:   make(chan *PaymentJob, maxQueueSize),
		worker:     worker,
	}
}

func (d *Dispatcher) Start(ctx context.Context) {
	d.globalWg.Add(1)

	var wg sync.WaitGroup

	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			d.globalWg.Done()
			return
		case job := <-d.jobQueue:
			wg.Add(1)
			d.workerPool <- struct{}{}

			go func(job *PaymentJob) {
				defer wg.Done()
				defer func() { <-d.workerPool }()
				d.worker.ProcessPayment(job)
			}(job)
		}
	}
}

func (d *Dispatcher) Wait() {
	d.globalWg.Wait()
}

func (d *Dispatcher) Enqueue(job *PaymentJob) {
	d.jobQueue <- job
}

func (d *Dispatcher) Clear() {
	for len(d.jobQueue) > 0 {
		<-d.jobQueue
	}

	for len(d.workerPool) > 0 {
		<-d.workerPool
	}
}

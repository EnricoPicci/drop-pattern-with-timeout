package workerpool

import (
	"fmt"
	"time"
)

type Worker[T Request] struct {
	id int
}

func NewWorker[T Request](id int) *Worker[T] {
	w := Worker[T]{id: id}
	return &w
}

func (w *Worker[T]) start(pool *WorkerPool[T]) {
	fmt.Printf("Worker %v started\n", w.id)

	var startIdleTime = time.Now()

	for req := range pool.inChan {
		// add the time spent idle - the startIdleTime value is reset at the end of the processing logic
		pool.addIdleTime(startIdleTime)

		if pool.isPoolToHalt() {
			// the pool is halted for the duration specified
			pool.halt()
		}
		// calculate how long the request has been waiting before being picked up by one worker of the pool
		waitDuration := time.Since(req.GetCreated())
		req.SetWaitDuration(waitDuration)

		// execute the request
		w.execReq(req, time.Duration(pool.procTime)*pool.timeUnit)

		pool.addRequestWaitTime(req)

		startIdleTime = time.Now()
	}
	pool.wgPool.Done()
	fmt.Printf("Worker %v shutting down\n", w.id)
}

// execute a request
func (w *Worker[T]) execReq(req Request, procTime time.Duration) {
	// sleep time that simulates the work done while processing a request
	time.Sleep(procTime)
	fmt.Printf("===>>>> Request executed with parameter %v - wait time %v\n", req.GetParam(), req.GetWaitDuration())
}

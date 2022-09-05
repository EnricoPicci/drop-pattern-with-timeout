package workerpool

import (
	"fmt"
	"time"

	"github.com/EnricoPicci/drop-pattern-with-timeout/src/request"
)

type Worker struct {
	id int
}

func NewWorker(id int) *Worker {
	w := Worker{id: id}
	return &w
}

func (w *Worker) start(pool *WorkerPool) {
	fmt.Printf("Worker %v started\n", w.id)

	var startIdleTime = time.Now()

	for req := range pool.inChan {
		// add the time spent idle - the startIdleTime value is reset at the end of the processing logic
		pool.addIdleTime(startIdleTime)

		pool.waitIfHalted()

		// calculate how long the request has been waiting before being picked up by one worker of the pool
		waitDuration := time.Since(req.Created)
		req.WaitDuration = waitDuration

		// execute the request
		w.execReq(req, time.Duration(pool.procTime)*pool.TimeUnit)

		pool.addRequest(req)

		startIdleTime = time.Now()
	}
	pool.wgPool.Done()
	fmt.Printf("Worker %v shutting down\n", w.id)
}

// execute a request
func (w *Worker) execReq(req request.Request, procTime time.Duration) {
	// sleep time that simulates the work done while processing a request
	time.Sleep(procTime)
	fmt.Printf("===>>>> Request executed with parameter %v - wait time %v\n", req.Param, req.WaitDuration)
}

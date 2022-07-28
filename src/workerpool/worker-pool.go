package workerpool

import (
	"sync"
	"time"

	"github.com/EnricoPicci/drop-pattern-with-timeout/src/request"
)

// WorkerPool type
type WorkerPool struct {
	// set the size of the worker pool
	poolSize int
	// time it takes to process a request
	procTime int
	// time after which the pool is halted
	haltPoolTime int
	// duration of the halt of the pool
	haltPoolDuration int

	// channel over which the pool receives the requests to process
	inChan chan request.Request
	// wait group used to control the closing of the pool
	wgPool sync.WaitGroup
	// requests processed
	requests []request.Request

	// a flag that signals if the halt period has already passed
	muHaltPeriodPassedCounter sync.Mutex
	haltPeriodPassedCounter   int
	// the time when the pool is started
	startPoolTime time.Time
	// measure the time spent by workers idle, i.e. ready to process a request but with no request coming in
	muWorkersIdleTime sync.Mutex
	workersIdleTime   time.Duration
	// protect the update of request related data
	muReq sync.Mutex
	// measure the time spent by requests waiting to be taken in by a worker
	cumulativeReqWaitTime time.Duration

	timeUnit time.Duration
}

func NewWorkerPool(
	inChan chan request.Request,
	poolSize int,
	reqInterval int,
	procTime int,
	numReq int,
	haltPoolTime int,
	haltPoolDuration int,
	timeUnit time.Duration,
) *WorkerPool {
	wp := WorkerPool{
		inChan:           inChan,
		poolSize:         poolSize,
		procTime:         procTime,
		haltPoolTime:     haltPoolTime,
		haltPoolDuration: haltPoolDuration,
		timeUnit:         timeUnit,

		requests: make([]request.Request, 0, numReq),
	}
	return &wp
}

// start the pool
func (wp *WorkerPool) Start() {
	wp.wgPool.Add(wp.poolSize)

	wp.startPoolTime = time.Now()
	// start the workers
	i := 0
	for i < wp.poolSize {
		w := NewWorker(i)
		go w.start(wp)
		i++
	}
}

// stop the pool
func (wp *WorkerPool) Stop() {
	close(wp.inChan)

	// This Wait makes sure that we return from this function before all requests in the channel have been completely processed
	wp.wgPool.Wait()
}

// add the time a request has waited before the pool has taken it in to start its processing
func (wp *WorkerPool) addRequest(req request.Request) {
	wp.muReq.Lock()
	wp.requests = append(wp.requests, req)
	// update the cumulative wait time
	wp.cumulativeReqWaitTime = wp.cumulativeReqWaitTime + req.WaitDuration
	wp.muReq.Unlock()
}

// add the time spent idle
func (wp *WorkerPool) addIdleTime(start time.Time) {
	wp.muWorkersIdleTime.Lock()
	wp.workersIdleTime = wp.workersIdleTime + time.Since(start)
	wp.muWorkersIdleTime.Unlock()
}

// returns the average of the time each worker has been idle waiting for requests to come in to be processed
func (wp *WorkerPool) AvgWorkerIdleTime() time.Duration {
	return time.Duration(int(wp.workersIdleTime) / wp.poolSize)
}

// returns the average time a request has been waiting from the moment it has been created and the moment a worker has taken it in to start its processing
func (wp *WorkerPool) AvgRequestWaitTime(numReq int) time.Duration {
	return time.Duration(int(wp.cumulativeReqWaitTime) / numReq)
}

// returns the requests processed
func (wp *WorkerPool) GetRequests() []request.Request {
	return wp.requests
}

// returns true if the pool has to be halted, i.e. if the time when the halt has to occurr has passed and the duration of the halt has not been passed
func (wp *WorkerPool) isPoolToHalt() bool {
	// after haltPoolTime the pool is halted
	haltPoolAfter := time.Duration(wp.haltPoolTime) * wp.timeUnit
	haltPool := time.Since(wp.startPoolTime) > haltPoolAfter
	wp.muHaltPeriodPassedCounter.Lock()
	haltPeriodNotPassed := wp.haltPeriodPassedCounter < wp.poolSize
	wp.muHaltPeriodPassedCounter.Unlock()
	return haltPool && haltPeriodNotPassed
}

// the pool is halted for the duration specified
func (wp *WorkerPool) halt() {
	haltDuration := time.Duration(wp.haltPoolDuration) * wp.timeUnit
	time.Sleep(haltDuration)
	wp.muHaltPeriodPassedCounter.Lock()
	wp.haltPeriodPassedCounter++
	wp.muHaltPeriodPassedCounter.Unlock()
}

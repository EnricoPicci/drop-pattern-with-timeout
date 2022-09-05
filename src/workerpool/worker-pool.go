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

	// protect the update of request related data
	muReq sync.Mutex
	// requests processed
	requests []request.Request

	// a flag that signals if thethe server is halted
	halted   bool
	muHalted sync.Mutex
	// the channels stored in restoredChans will be closed when the pool is restored to signal that the server is back to normal operations
	restoredChans []chan struct{}

	// the time when the pool is started
	startPoolTime time.Time
	// measure the time spent by workers idle, i.e. ready to process a request but with no request coming in
	muWorkersIdleTime sync.Mutex
	workersIdleTime   time.Duration
	// measure the time spent by requests waiting to be taken in by a worker
	cumulativeReqWaitTime time.Duration

	TimeUnit time.Duration
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
		TimeUnit:         timeUnit,

		requests: make([]request.Request, 0, numReq),
	}

	// manages the halt and restore of the server based on the values of haltPoolTime and haltPoolDuration properties
	go wp.halt()
	go wp.restore()

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

// add a request to the collection of requests processed by the pool and
// update the cumulative time that measure how long requests have waited before entering the pool to start processing
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

// sets the halted flag to true when the server has to be halted
func (wp *WorkerPool) halt() {
	// after haltPoolTime the pool is halted
	haltPoolAfter := time.Duration(wp.haltPoolTime) * wp.TimeUnit
	time.Sleep(haltPoolAfter)
	wp.muHalted.Lock()
	wp.halted = true
	wp.muHalted.Unlock()
}

// if the server is halted it waits until it is restored to normal operations
func (wp *WorkerPool) waitIfHalted() {
	var isHalted bool
	var restored chan struct{}
	// if the pool is halted, we want to add a chan to the "restoredChans" slice in an isolated way, i.e. we want to avoid the risk
	// that in another goroutine the pool is restored to normal work before we have added the chan to the "restoredChans" slice
	wp.muHalted.Lock()
	isHalted = wp.halted
	if isHalted {
		restored = make(chan struct{})
		wp.restoredChans = append(wp.restoredChans, restored)
	}
	// if the pool is halted we wait until the restored channel signals - this is done outside the Lock on the wp.muHalted otherwise the first goroutine
	// running a worker that enters the wp.muHalted Lock block would bring to a stall since the signalling to the restored channel is also protected
	// by the same semaphore
	wp.muHalted.Unlock()
	if isHalted {
		// restored is closed when the server is restored to signal that operations are back to normal
		<-restored
	}
}

// reset the halted flag to false when the server has to be come back to life
func (wp *WorkerPool) restore() {
	// when the halt period has elapsed we restore the server and signal all which are interested that the server is back
	// by closing the channel they have passed in
	haltPoolAfter := time.Duration(wp.haltPoolTime) * wp.TimeUnit
	haltDuration := time.Duration(wp.haltPoolDuration) * wp.TimeUnit
	time.Sleep(haltPoolAfter + haltDuration)
	// the write on wp.muHalted is protected by a semaphore to prevent concurrent writing
	wp.muHalted.Lock()
	wp.halted = false
	for i := range wp.restoredChans {
		close(wp.restoredChans[i])
	}
	wp.muHalted.Unlock()
}

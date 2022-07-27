package main

import (
	"flag"
	"fmt"
	"sync"
	"time"
)

// In this example we implement drop pattern with timeout

type Request struct {
	Param        int
	created      time.Time
	waitDuration time.Duration
}

// set the size of the worker pool
var poolSize int

// time unit utilized to calculate durations
var timeUnit = time.Millisecond

// interval between requests
var reqInterval int

// time it takes to process a request
var procTime int

// time after which the pool is halted
var haltPoolTime int

// duration of the halt of the pool
var haltPoolDuration int

// a flag that signals if the halt period has already passed
var haltPeriodPassedCounter int
var muHaltPeriodPassedCounter sync.Mutex

// the time when the pool is started
var startPoolTime time.Time

// measure the time spent by workers idle, i.e. ready to process a request but with no request coming in
var muWorkersIdleTime sync.Mutex
var workersIdleTime time.Duration

// measure the time spent by requests waiting to be taken in by a worker
var muReqWaitTime sync.Mutex
var cumulativeReqWaitTime time.Duration

func main() {
	_poolSize := flag.Int("poolSize", 10, "number of workers in the worker pool")
	_reqInterval := flag.Int("reqInterval", 100, "interval between requests in miliseconds")
	_procTime := flag.Int("procTime", 1000, "time it takes to process a request")
	_numReq := flag.Int("numReq", 100, "number of requests coming in to be processed")
	_haltPoolTime := flag.Int("haltPoolTime", 2000, "at which time (after start) the pool is halted and stops processing in milliseconds")
	_haltPoolDuration := flag.Int("haltPoolDuration", 1000, "for how long the pool is halted in milliseconds")
	flag.Parse()

	avgIdleTime, avgWaitTime := workerPoolWithoutDropPattern(*_poolSize, *_reqInterval, *_procTime, *_numReq, *_haltPoolTime, *_haltPoolDuration)

	fmt.Printf("Average idle time for a worker: %v\n", avgIdleTime)
	fmt.Printf("Average wait time for a request: %v\n", avgWaitTime)
}

func workerPoolWithoutDropPattern(_poolSize int,
	_reqInterval int,
	_procTime int,
	numReq int,
	_haltPoolTime int,
	_haltPoolDuration int) (idleTime time.Duration, waitTime time.Duration) {

	poolSize = _poolSize
	reqInterval = _reqInterval
	procTime = _procTime
	haltPoolTime = _haltPoolTime
	haltPoolDuration = _haltPoolDuration

	haltPeriodPassedCounter = 0
	workersIdleTime = 0
	cumulativeReqWaitTime = 0

	fmt.Println("Start processing requests")
	fmt.Print("\n")

	flag.VisitAll(func(f *flag.Flag) {
		fmt.Printf("%s: %s  (%s)\n", f.Name, f.Value, f.Usage)
	})
	fmt.Print("\n")

	// the channel that provides requests to the pool has a buffer equal to the number of requests
	// this makes sure that the requests can come in at the same rythm even if the pool is halted
	inPoolCh := make(chan Request, numReq)

	var wgPool sync.WaitGroup
	wgPool.Add(poolSize)

	// start the worker pool
	startPool(inPoolCh, &wgPool, poolSize)

	// we simulate a stream of incoming requests
	for i := 0; i < numReq; i++ {
		// interval between each incoming request
		var intervalBetweenRequests = time.Duration(reqInterval) * timeUnit
		time.Sleep(time.Duration(intervalBetweenRequests))
		req := Request{i, time.Now(), 0}

		inPoolCh <- req

		fmt.Printf("Sent %v\n", req.Param)
	}

	// when there are no more requests that can enter the pool we close the in pool channel
	close(inPoolCh)

	// This Wait makes sure that we do not shut down the program before all requests in the channel have been completely processed
	// Without this wait we run the risk of terminating main, and therfore shutting down the whole program, while some requests
	// are still being processed by some of the workers
	wgPool.Wait()

	idleTime = time.Duration(int(workersIdleTime) / poolSize)
	waitTime = time.Duration(int(cumulativeReqWaitTime) / numReq)
	return
}

func startPool(inCh <-chan Request, wg *sync.WaitGroup, poolSize int) {
	startPoolTime = time.Now()
	// start the workers
	i := 0
	for i < poolSize {
		go doWork(inCh, wg, i)
		i++
	}
}

func doWork(inCh <-chan Request, wg *sync.WaitGroup, i int) {
	fmt.Printf("Worker %v started\n", i)

	var startIdleTime = time.Now()

	for req := range inCh {
		// add the time spent idle - the startIdleTime value is reset at the end of the processing logic
		muWorkersIdleTime.Lock()
		workersIdleTime = workersIdleTime + time.Since(startIdleTime)
		muWorkersIdleTime.Unlock()

		// after haltPoolTime the pool is halted
		haltPoolAfter := time.Duration(haltPoolTime) * timeUnit
		haltPool := time.Since(startPoolTime) > haltPoolAfter
		muHaltPeriodPassedCounter.Lock()
		haltPeriodNotPassed := haltPeriodPassedCounter < poolSize
		muHaltPeriodPassedCounter.Unlock()
		if haltPool && haltPeriodNotPassed {
			if haltPoolDuration > 1 {
				fmt.Println(">>>>>")
			}
			// the pool is halted for the duration specified
			haltDuration := time.Duration(haltPoolDuration) * timeUnit
			time.Sleep(haltDuration)
			muHaltPeriodPassedCounter.Lock()
			haltPeriodPassedCounter++
			muHaltPeriodPassedCounter.Unlock()
		}

		// calculate how long the request has been waiting before being picked up by one worker of the pool
		waitDuration := time.Since(req.created)
		req.waitDuration = waitDuration

		// sleep time that simulates the work done while processing a request
		var processingTime = time.Duration(procTime) * timeUnit
		time.Sleep(time.Duration(processingTime))
		fmt.Printf("===>>>> Request executed with parameter %v - wait time %v\n", req.Param, req.waitDuration)

		// update the cumulative wait time
		muReqWaitTime.Lock()
		cumulativeReqWaitTime = cumulativeReqWaitTime + req.waitDuration
		muReqWaitTime.Unlock()
		startIdleTime = time.Now()
	}
	wg.Done()
	fmt.Printf("Worker %v shutting down\n", i)
}

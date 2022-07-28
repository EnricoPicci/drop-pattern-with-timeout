package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/EnricoPicci/drop-pattern-with-timeout/src/request"
	"github.com/EnricoPicci/drop-pattern-with-timeout/src/workerpool"
)

// In this example we implement drop pattern with timeout

// time unit utilized to calculate durations
var timeUnit = time.Millisecond

func main() {
	_poolSize := flag.Int("poolSize", 10, "number of workers in the worker pool")
	_reqInterval := flag.Int("reqInterval", 100, "interval between requests in miliseconds")
	_procTime := flag.Int("procTime", 1000, "time it takes to process a request")
	_numReq := flag.Int("numReq", 100, "number of requests coming in to be processed")
	_haltPoolTime := flag.Int("haltPoolTime", 2000, "at which time (after start) the pool is halted and stops processing in milliseconds")
	_haltPoolDuration := flag.Int("haltPoolDuration", 1000, "for how long the pool is halted in milliseconds")
	flag.Parse()

	flag.VisitAll(func(f *flag.Flag) {
		fmt.Printf("%s: %s  (%s)\n", f.Name, f.Value, f.Usage)
	})
	fmt.Print("\n")

	avgIdleTime, avgWaitTime := workerPoolWithoutDropPattern(*_poolSize, *_reqInterval, *_procTime, *_numReq, *_haltPoolTime, *_haltPoolDuration)

	fmt.Printf("Average idle time for a worker: %v\n", avgIdleTime)
	fmt.Printf("Average wait time for a request: %v\n", avgWaitTime)
}

func workerPoolWithoutDropPattern(
	poolSize int,
	reqInterval int,
	procTime int,
	numReq int,
	haltPoolTime int,
	haltPoolDuration int) (idleTime time.Duration, waitTime time.Duration) {

	fmt.Println("Start processing requests")
	fmt.Print("\n")

	// the channel that provides requests to the pool has a buffer equal to the number of requests
	// this makes sure that the requests can come in at the same rythm even if the pool is halted
	inPoolCh := make(chan *request.Request, numReq)

	pool := workerpool.NewWorkerPool(inPoolCh, poolSize, reqInterval, procTime, numReq, haltPoolTime, haltPoolDuration, timeUnit)

	// start the worker pool
	pool.Start()

	// we simulate a stream of incoming requests
	for i := 0; i < numReq; i++ {
		// interval between each incoming request
		var intervalBetweenRequests = time.Duration(reqInterval) * timeUnit
		time.Sleep(time.Duration(intervalBetweenRequests))
		req := request.Request{Param: i, Created: time.Now(), WaitDuration: 0}

		inPoolCh <- &req

		fmt.Printf("Sent %v\n", req.Param)
	}

	// when there are no more requests that can enter the pool we can stop the pool
	pool.Stop()

	idleTime = pool.AvgWorkerIdleTime()
	waitTime = pool.AvgRequestWaitTime(numReq)
	return
}

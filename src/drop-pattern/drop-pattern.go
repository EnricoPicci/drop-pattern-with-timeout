package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/EnricoPicci/drop-pattern-with-timeout/src/request"
	"github.com/EnricoPicci/drop-pattern-with-timeout/src/waitingroom"
	"github.com/EnricoPicci/drop-pattern-with-timeout/src/workerpool"
)

// In this example we implement drop pattern with timeout

// time unit utilized to calculate durations
var timeUnit = time.Millisecond

func main() {
	poolSize := flag.Int("poolSize", 10, "number of workers in the worker pool")
	reqInterval := flag.Int("reqInterval", 100, "interval between requests in miliseconds")
	procTime := flag.Int("procTime", 1000, "time it takes to process a request")
	numReq := flag.Int("numReq", 100, "number of requests coming in to be processed")
	haltPoolTime := flag.Int("haltPoolTime", 2000, "at which time (after start) the pool is halted and stops processing in milliseconds")
	haltPoolDuration := flag.Int("haltPoolDuration", 1000, "for how long the pool is halted in milliseconds")
	timeout := flag.Int("timeout", 500, "the timeout after which an incoming request not yet taken by the worker pool is dropped")
	flag.Parse()

	flag.VisitAll(func(f *flag.Flag) {
		fmt.Printf("%s: %s  (%s)\n", f.Name, f.Value, f.Usage)
	})
	fmt.Print("\n")

	avgIdleTime, avgWaitTime, requestsSentToPool, requestsDropped := workerPoolWithDropPattern(
		*poolSize, *reqInterval, *procTime, *numReq, *haltPoolTime, *haltPoolDuration, *timeout,
	)

	fmt.Printf("Average idle time for a worker: %v\n", avgIdleTime)
	fmt.Printf("Average wait time for a request: %v\n", avgWaitTime)
	fmt.Printf("Number of requests sent to pool: %v\n", len(requestsSentToPool))
	fmt.Printf("Number of requests dropped: %v\n", len(requestsDropped))
}

func workerPoolWithDropPattern(
	poolSize int,
	reqInterval int,
	procTime int,
	numReq int,
	haltPoolTime int,
	haltPoolDuration int,
	timeout int) (idleTime time.Duration, waitTime time.Duration, requestsProcessed []request.Request, requestsDropped []request.Request) {

	fmt.Println("Start processing requests")
	fmt.Print("\n")

	// the channel that provides requests to the pool is unbuffered - this is mandatory for the drop pattern to work
	inPoolCh := make(chan request.Request)
	pool := workerpool.NewWorkerPool(inPoolCh, poolSize, reqInterval, procTime, numReq, haltPoolTime, haltPoolDuration, timeUnit)

	// the channel that allows request to enter the waiting room
	waitingRoomCh := make(chan request.Request)
	waitingRoom := waitingroom.New(waitingRoomCh, inPoolCh, timeout, timeUnit)

	// start the worker pool
	pool.Start()
	// start the waiting room
	waitingRoom.Open()

	// we simulate a stream of incoming requests
	for i := 0; i < numReq; i++ {
		// interval between each incoming request
		var intervalBetweenRequests = time.Duration(reqInterval) * timeUnit
		time.Sleep(time.Duration(intervalBetweenRequests))
		req := request.Request{Param: i, Created: time.Now()}

		// the request is sent to the waiting room
		waitingRoom.LetIn(req)
	}

	// close the waiting room since there are no more requests that can arrive
	waitingRoom.Close()
	// when there are no more requests that can enter the pool we can stop the pool
	pool.Stop()

	idleTime = pool.AvgWorkerIdleTime()
	waitTime = pool.AvgRequestWaitTime(numReq)
	requestsProcessed = pool.GetRequests()
	requestsDropped = waitingRoom.ReqDropped
	return
}

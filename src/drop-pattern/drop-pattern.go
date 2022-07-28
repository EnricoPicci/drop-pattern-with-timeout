package main

import (
	"context"
	"flag"
	"fmt"
	"sync"
	"time"

	"github.com/EnricoPicci/drop-pattern-with-timeout/src/request"
	"github.com/EnricoPicci/drop-pattern-with-timeout/src/workerpool"
)

// In this example we implement drop pattern with timeout

// time unit utilized to calculate durations
var timeUnit = time.Millisecond

var muReqDropped sync.Mutex
var requestDropped = 0

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

	avgIdleTime, avgWaitTime := workerPoolWithDropPattern(*poolSize, *reqInterval, *procTime, *numReq, *haltPoolTime, *haltPoolDuration, *timeout)

	fmt.Printf("Average idle time for a worker: %v\n", avgIdleTime)
	fmt.Printf("Average wait time for a request: %v\n", avgWaitTime)
}

func workerPoolWithDropPattern(
	poolSize int,
	reqInterval int,
	procTime int,
	numReq int,
	haltPoolTime int,
	haltPoolDuration int,
	timeout int) (idleTime time.Duration, waitTime time.Duration) {

	fmt.Println("Start processing requests")
	fmt.Print("\n")

	// the channel that provides requests to the pool is unbuffered - this is mandatory for the drop pattern to work
	inPoolCh := make(chan request.Request)

	pool := workerpool.NewWorkerPool(inPoolCh, poolSize, reqInterval, procTime, numReq, haltPoolTime, haltPoolDuration, timeUnit)

	// start the worker pool
	pool.Start()

	ctx := context.Background()

	var wgReq sync.WaitGroup

	// we simulate a stream of incoming requests
	for i := 0; i < numReq; i++ {
		// interval between each incoming request
		var intervalBetweenRequests = time.Duration(reqInterval) * timeUnit
		time.Sleep(time.Duration(intervalBetweenRequests))
		req := request.Request{Param: i, Created: time.Now(), WaitDuration: 0}

		wgReq.Add(1)

		// within this goroutine we implement the drop with timeout pattern
		go processOrDrop(ctx, req, inPoolCh, &wgReq, timeout)

		fmt.Printf("Sent %v\n", req.Param)
	}

	// wait for all the gouroutines launched for each request to either be able to pass the request to the pool or to drop it
	wgReq.Wait()
	// when there are no more requests that can enter the pool we can stop the pool
	pool.Stop()

	idleTime = pool.AvgWorkerIdleTime()
	waitTime = pool.AvgRequestWaitTime(numReq)
	return
}

// this function implements the drop with timeout pattern
func processOrDrop(ctx context.Context, req request.Request, inPoolCh chan<- request.Request, wgReq *sync.WaitGroup, timeout int) {
	defer wgReq.Done()

	// the timeout context
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*timeUnit)
	defer cancel()

	// wait until the request enters the input channel of the worker pool or the context times out
	select {
	case inPoolCh <- req:
		// the request is sent to the input channel of the pool
	case <-ctx.Done():
		// the context times out and the request is dropped
		fmt.Printf("Request %v dropped\n", req.Param)
		muReqDropped.Lock()
		requestDropped++
		muReqDropped.Unlock()
	}
}

package main

import (
	"runtime"
	"testing"
	"time"
)

// If the system is balanced,the pool is able to process the requests without imposing to them too much delay and, at the same time,
// without keeping the workers too idle
func TestWorkerPoolWithoutDropPattern_balanced_system(t *testing.T) {
	poolSize := 10
	reqInterval := 100
	procTime := 1000
	numReq := 100
	haltPoolTime := 0
	haltPoolDuration := 0

	// set procs to 1 to limit variance in the tests
	runtime.GOMAXPROCS(1)

	idleTime, waitTime := workerPoolWithoutDropPattern(poolSize, reqInterval, procTime, numReq, haltPoolTime, haltPoolDuration)

	// test that the idle time is not too high - this is a test based on the results obtained on my machine
	if idleTime > 1*time.Second {
		t.Errorf("The idle time seems too high: %v", idleTime)
	}
	t.Logf("Idle time: %v", idleTime)

	// test that the idle time is not too high - this is a test based on the results obtained on my machine
	if waitTime > 1*time.Millisecond {
		t.Errorf("The wait time seems too high: %v", waitTime)
	}
	t.Logf("Wait time: %v", waitTime)
}

// If the system has too many workers for the number of requests coming in, then the workers are idle a lot of time
func TestWorkerPoolWithoutDropPattern_many_workers(t *testing.T) {
	// 10 workers are just enough to process this volume of requests - 100 are more than required
	poolSize := 100

	reqInterval := 100
	procTime := 1000
	numReq := 100
	haltPoolTime := 0
	haltPoolDuration := 0

	// set procs to 1 to limit variance in the tests
	runtime.GOMAXPROCS(1)

	idleTime, waitTime := workerPoolWithoutDropPattern(poolSize, reqInterval, procTime, numReq, haltPoolTime, haltPoolDuration)

	// test that the idle time is not too high - this is a test based on the results obtained on my machine
	if idleTime < 5*time.Second {
		t.Errorf("The idle time seems too low : %v", idleTime)
	}
	t.Logf("Idle time: %v", idleTime)

	// test that the idle time is not too high - this is a test based on the results obtained on my machine
	if waitTime > 1*time.Millisecond {
		t.Errorf("The wait time seems too high: %v", waitTime)
	}
	t.Logf("Wait time: %v", waitTime)
}

// If the system is balanced so that it can process the flow of requests without introducing wait time for the requests and without
// keeping workers idle, but then an event occurs that introduces a delay (a temporary block), then the delay is propageted over time
func TestWorkerPoolWithoutDropPattern_temporary_block_occurs(t *testing.T) {
	// use a pool size of 1 for this test so that the requests have to be processed one by one and so we can see, as effect of the stop,
	// a delay very close to the duration of the stop for all requests - if we had more than one worker, some requests would have to wait less than
	// the others since they arrive in after the stop occurred and therefore they have a lower delay.
	// For instance, if we have 2 workers and each request come in every 1 sec, as soon as the stop occurs the first request will have to wait
	// for the intire duration of the stop while the following one will arrive 1 sec later and therefore will have to wait 1 sec less
	poolSize := 1
	reqInterval := 100
	procTime := 100
	numReq := 100

	haltPoolTime := 1000
	haltPoolDuration := 2000

	// set procs to 1 to limit variance in the tests
	runtime.GOMAXPROCS(1)

	idleTime, waitTime := workerPoolWithoutDropPattern(poolSize, reqInterval, procTime, numReq, haltPoolTime, haltPoolDuration)

	// test that the idle time is not too high - this is a test based on the results obtained on my machine
	if idleTime > 1*time.Second {
		t.Errorf("The idle time seems too low : %v", idleTime)
	}
	t.Logf("Idle time: %v", idleTime)

	// test that the wait time is not too low - this is a test based on the results obtained on my machine
	if waitTime < 1500*time.Millisecond {
		t.Errorf("The wait time seems too low: %v", waitTime)
	}
	t.Logf("Wait time: %v", waitTime)
}

package main

import (
	"runtime"
	"testing"
	"time"
)

// If the system is balanced,the pool is able to process the requests without imposing to them too much delay and, at the same time,
// without keeping the workers too idle
func TestWorkerPoolWithDropPattern_balanced_system(t *testing.T) {
	poolSize := 10
	reqInterval := 100
	procTime := 1000
	numReq := 100
	haltPoolTime := 0
	haltPoolDuration := 0
	timeout := 500

	// set procs to 1 to limit variance in the tests
	runtime.GOMAXPROCS(1)

	idleTime, waitTime := workerPoolWithDropPattern(poolSize, reqInterval, procTime, numReq, haltPoolTime, haltPoolDuration, timeout)

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
func TestWorkerPoolWithDropPattern_many_workers(t *testing.T) {
	// 10 workers are just enough to process this volume of requests - 100 are more than required
	poolSize := 100

	reqInterval := 100
	procTime := 1000
	numReq := 100
	haltPoolTime := 0
	haltPoolDuration := 0
	timeout := 500

	// set procs to 1 to limit variance in the tests
	runtime.GOMAXPROCS(1)

	idleTime, waitTime := workerPoolWithDropPattern(poolSize, reqInterval, procTime, numReq, haltPoolTime, haltPoolDuration, timeout)

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
// since though a timeout is applied, the wait time should be in the range of the timeout
func TestWorkerPoolWithDropPattern_temporary_block_occurs(t *testing.T) {
	poolSize := 10
	reqInterval := 100
	procTime := 1000
	numReq := 100

	haltPoolTime := 1000
	haltPoolDuration := 2000
	timeout := 500

	// set procs to 1 to limit variance in the tests
	runtime.GOMAXPROCS(1)

	idleTime, waitTime := workerPoolWithDropPattern(poolSize, reqInterval, procTime, numReq, haltPoolTime, haltPoolDuration, timeout)

	// test that the idle time is not too high - this is a test based on the results obtained on my machine
	if idleTime > 1*time.Second {
		t.Errorf("The idle time seems too low : %v", idleTime)
	}
	t.Logf("Idle time: %v", idleTime)

	// test that the wait time is not too high since there is a timeout
	// consider just the last requests so that we ignore the first ones, not affected by the delay, and the ones that had a much longer wait time
	// given that they were affected by the long delay
	// this is a test based on the results obtained on my machine
	if waitTime > time.Duration(timeout*3)*time.Millisecond {
		t.Errorf("The wait time seems too high: %v", waitTime)
	}
	t.Logf("Wait time: %v", waitTime)
}

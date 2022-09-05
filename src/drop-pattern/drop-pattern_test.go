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

	idleTime, waitTime, requestsProcessed, requestsDropped := workerPoolWithDropPattern(poolSize, reqInterval, procTime, numReq, haltPoolTime, haltPoolDuration, timeout)

	// test that the idle time is not too high - this is a test based on the results obtained on my machine
	if idleTime > 1*time.Second {
		t.Errorf("The idle time seems too high: %v", idleTime)
	}
	t.Logf("Idle time: %v", idleTime)

	// test that the idle time is not too high - this is a test based on the results obtained on my machine
	if waitTime > 1*time.Millisecond {
		t.Errorf("The wait time seems too high: %v", waitTime)
	}

	// test that there are no requests which have been dropped
	if len(requestsDropped) != 0 {
		t.Error("No requests should have been dropped")
	}

	// test that all requests have been processed
	numReqProcessed := len(requestsProcessed)
	if numReqProcessed != numReq {
		t.Errorf("Some requests are missing. Requests processed: %v - Requests expected: %v",
			numReqProcessed, numReq)
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

	idleTime, waitTime, requestsProcessed, requestsDropped := workerPoolWithDropPattern(poolSize, reqInterval, procTime, numReq, haltPoolTime, haltPoolDuration, timeout)

	// test that the idle time is not too high - this is a test based on the results obtained on my machine
	if idleTime < 5*time.Second {
		t.Errorf("The idle time seems too low : %v", idleTime)
	}
	t.Logf("Idle time: %v", idleTime)

	// test that the idle time is not too high - this is a test based on the results obtained on my machine
	if waitTime > 1*time.Millisecond {
		t.Errorf("The wait time seems too high: %v", waitTime)
	}

	// test that there are no requests which have been dropped
	if len(requestsDropped) != 0 {
		t.Error("No requests should have been dropped")
	}

	// test that all requests have been processed
	numReqProcessed := len(requestsProcessed)
	if numReqProcessed != numReq {
		t.Errorf("Some requests are missing. Requests processed: %v - Requests expected: %v",
			numReqProcessed, numReq)
	}

	t.Logf("Wait time: %v", waitTime)
}

// The system is balanced. It has a single worker and the processing time is equal to the interval between 2 requests.
// This means that it can process the flow of requests without introducing wait time for the requests and without
// keeping workers idle, but then an event occurs that introduces a delay (a temporary block), then the delay is propageted over time
// since though a timeout is applied, the wait time should be in the range of the timeout
func TestWorkerPoolWithDropPattern_temporary_block_occurs_1_worker(t *testing.T) {
	poolSize := 1
	reqInterval := 100
	procTime := 100
	numReq := 100

	haltPoolTime := 0
	haltPoolDuration := 2000
	timeout := 500

	// set procs to 1 to limit variance in the tests
	runtime.GOMAXPROCS(1)

	idleTime, waitTime, requestsProcessed, requestsDropped := workerPoolWithDropPattern(poolSize, reqInterval, procTime, numReq, haltPoolTime, haltPoolDuration, timeout)

	// test that the idle time is not too high - this is a test based on the results obtained on my machine
	if idleTime > 1*time.Second {
		t.Errorf("The idle time seems too low : %v", idleTime)
	}
	t.Logf("Idle time: %v", idleTime)

	// test that the wait time of the last requests sent to the pool is close to the value timeout - reqInterval
	// the reason the expected delay "timeout - procTime" instead of simply "timeout" is the following:
	// - after the server start operating again after the stop, the first requests to be processed will be those which have been waiting for a time
	//   close to the value of timeout
	// - due to the delay introduced by channel to channel communications and so on, it is likely to occur that one request gets dropped, after the
	//   restore of the server, since it is waiting for a bit more than the timeout
	// - when this happens, the following request to be processed is the one which came just after the one dropped (remember that in this test there is
	//   only 1 worker) and this request has therefore been waiting for a period equal to "timeout - reqInterval"
	// - from this request onwards, all the other requests processed will carry a delay close to "timeout - reqInterval"
	//
	// consider just the last requests so that we ignore the first ones, not affected by the delay, and the ones that had a much longer wait time
	// given that they had already entered the worker pool when the halt occurred
	// this is a test based on the results obtained on my machine
	timeoutDuration := time.Duration(timeout) * timeUnit
	lowerWaitTimeThreshold := (timeout - reqInterval) * 9 / 10
	lowerWaitTimeThresholdDuration := time.Duration(lowerWaitTimeThreshold) * timeUnit
	lastRequestsProcessed := requestsProcessed[len(requestsProcessed)-10:]
	for _, req := range lastRequestsProcessed {
		if req.WaitDuration > timeoutDuration {
			t.Errorf("The last requests processed should have waited less then the timeout %v - the request %v has waited %v",
				timeoutDuration, req.Param, req.WaitDuration)
		}
		if req.WaitDuration < lowerWaitTimeThresholdDuration {
			t.Errorf("The wait time for the last requests processed should close to the value of timeout %v - the request %v has waited %v",
				lowerWaitTimeThresholdDuration, req.Param, req.WaitDuration)
		}
	}
	if waitTime > time.Duration(timeout*3)*timeUnit {
		t.Errorf("The wait time seems too high: %v", waitTime)
	}

	// test that there are some requests which have been dropped
	if len(requestsDropped) == 0 {
		t.Error("Some requests should have been dropped")
	}

	// test that all requests have been either sent to the pool or dropped
	numReqProcessed := len(requestsProcessed)
	numReqDropped := len(requestsDropped)
	if numReqProcessed+numReqDropped != numReq {
		t.Errorf("Some requests are missing. Requests processed: %v - Requests dropped: %v - Requests expected: %v",
			numReqProcessed, numReqDropped, numReq)
	}

	t.Logf("Wait time: %v", waitTime)
}

// If the system is balanced so that it can process the flow of requests without introducing wait time for the requests and without
// keeping workers idle.
// After "haltPoolTime" a temporary block of the pool occurs and all processing is stopped for a duration equal "haltPoolDuration".
// Since a timeout (lower than "haltPoolDuration") has been defined, some requests will be dropped.
// When the pool resumes normal operations, there will be a queue of requests waiting to be processed. The length of the queue would be
// approximatively equal to "timeout"/"reqInterval". When the pool resumes normal operations all the workers will become available
// to process new requests approximatively at the same time, and therefore a number of requests equal to "poolSize" will start being processed
// more or less at the same time. These requests will have accumulated a "delay" which ranges from a value between "timeout" and "timeout"-"reqInterval"
// (for the oldest request) to a value between "reqInterval" and 0 (for the last request that was received before resuming the pool).
// When the workers start processing the requests again, they will take "procTime" to process each request. So, in this "procTime" interval,
// other requests will come in and will form a queue. When, approximatively at the same time, the workers complete the processing of their requests
// and are ready to pick up new requests, the queue will have a length of about "procTime"/"reqInterval".
func TestWorkerPoolWithDropPattern_temporary_block_occurs_10_workers(t *testing.T) {
	poolSize := 10
	reqInterval := 100
	procTime := 1000
	numReq := 100

	haltPoolTime := 1000
	haltPoolDuration := 2000
	timeout := 1000

	// set procs to 1 to limit variance in the tests
	runtime.GOMAXPROCS(1)

	pool, waitingRoom := newPoolAndWaitingRoom(poolSize, reqInterval, procTime, numReq, haltPoolTime, haltPoolDuration, timeout)

	// take the number of requests waiting at regular interval
	queueLenghts := make([]int, numReq)
	ticker := time.NewTicker(time.Duration(reqInterval) * pool.TimeUnit)
	done := make(chan struct{})
	count := 0
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				queueLenghts[count] = waitingRoom.QueueLength
				count++
				if count == numReq {
					close(done)
				}
			}
		}
	}()

	idleTime, waitTime, requestsProcessed, requestsDropped := _workerPoolWithDropPattern(pool, waitingRoom, numReq, reqInterval)

	// wait for the queue lenghts to be all collected before starting the assertions
	<-done

	lastQueueLengths := queueLenghts[len(queueLenghts)-10:]
	expectedLastQueueLengths := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	for i, v := range lastQueueLengths {
		if v != expectedLastQueueLengths[i] {
			t.Errorf("The number of items in the wating room queue at the end of the test is %v and not %v as expected",
				lastQueueLengths, expectedLastQueueLengths)
		}
	}

	// test that the idle time is not too high - this is a test based on the results obtained on my machine
	if idleTime > 1*time.Second {
		t.Errorf("The idle time seems too low : %v", idleTime)
	}
	t.Logf("Idle time: %v", idleTime)

	// test that the wait time is not too high - this is a test based on the results obtained on my machine
	if waitTime > time.Duration(timeout*3)*timeUnit {
		t.Errorf("The wait time seems too high: %v", waitTime)
	}

	// test that there are some requests which have been dropped
	if len(requestsDropped) == 0 {
		t.Error("Some requests should have been dropped")
	}

	// test that all requests have been either sent to the pool or dropped
	numReqProcessed := len(requestsProcessed)
	numReqDropped := len(requestsDropped)
	if numReqProcessed+numReqDropped != numReq {
		t.Errorf("Some requests are missing. Requests processed: %v - Requests dropped: %v - Requests expected: %v",
			numReqProcessed, numReqDropped, numReq)
	}

	t.Logf("Wait time: %v", waitTime)
}

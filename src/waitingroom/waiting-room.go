package waitingroom

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/EnricoPicci/drop-pattern-with-timeout/src/request"
)

type WaitingRoom struct {
	inChan   chan request.Request
	outChan  chan<- request.Request
	timeout  int
	timeUnit time.Duration

	muReqSentToPool sync.Mutex
	ReqSentToPool   []request.Request

	muReqDropped sync.Mutex
	ReqDropped   []request.Request

	WgReq sync.WaitGroup

	// holds the number of requests which are waiting in the waiting room
	QueueLength int
}

func New(inChan chan request.Request, outChan chan request.Request, timeout int, timeUnit time.Duration) *WaitingRoom {
	wr := WaitingRoom{
		inChan:     inChan,
		outChan:    outChan,
		timeout:    timeout,
		timeUnit:   timeUnit,
		ReqDropped: make([]request.Request, 0),
	}
	return &wr
}

func (wr *WaitingRoom) Open() {
	go func() {
		ctx := context.Background()

		for req := range wr.inChan {
			// within this goroutine we implement the drop with timeout pattern
			go wr.sendOrDrop(ctx, req)
			fmt.Printf("Request %v in the waiting room\n", req.Param)
		}
	}()
}

func (wr *WaitingRoom) Close() {
	close(wr.inChan)
	wr.WgReq.Wait()
}

func (wr *WaitingRoom) LetIn(req request.Request) {
	wr.WgReq.Add(1)
	wr.inChan <- req
}

// this function implements the drop with timeout pattern
func (wr *WaitingRoom) sendOrDrop(ctx context.Context, req request.Request) {
	// defer wr.wgReq.Done()
	defer wr.WgReq.Done()

	// the timeout context
	ctx, cancel := context.WithTimeout(ctx, wr.getTimeout())
	defer cancel()

	wr.QueueLength++
	// wait until the request enters the input channel of the worker pool or the context times out
	select {
	case wr.outChan <- req:
		// the request is sent to the output channel
		wr.sentToPool(req)
	case <-ctx.Done():
		// the context times out and the request is dropped
		wr.drop(req)
	}
	wr.QueueLength--
}

func (wr *WaitingRoom) sentToPool(req request.Request) {
	fmt.Printf("Request %v sent to pool\n", req.Param)
	wr.muReqSentToPool.Lock()
	wr.ReqSentToPool = append(wr.ReqSentToPool, req)
	wr.muReqSentToPool.Unlock()
}

func (wr *WaitingRoom) drop(req request.Request) {
	fmt.Printf("Request %v dropped\n", req.Param)
	wr.muReqDropped.Lock()
	wr.ReqDropped = append(wr.ReqDropped, req)
	wr.muReqDropped.Unlock()
}

func (wr *WaitingRoom) getTimeout() time.Duration {
	return time.Duration(wr.timeout) * wr.timeUnit
}

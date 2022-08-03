package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func main() {
	ch_1 := make(chan int)
	ch_2 := make(chan int)

	var wg_1 sync.WaitGroup
	var wg_2 sync.WaitGroup

	go work(ch_2, &wg_2)

	go func() {
		for v := range ch_1 {
			time.Sleep(100000000)
			ctx := context.Background()

			wg_1.Add(1)

			go send(ctx, v, &wg_1, ch_2)

			fmt.Printf("Sent %v\n", v)
		}
	}()

	for i := 0; i < 10; i++ {
		ch_1 <- i
	}

	close(ch_1)
	wg_1.Wait()
	close(ch_2)
	wg_2.Wait()
}

func send(ctx context.Context, v int, wg_1 *sync.WaitGroup, ch_2 chan int) {
	defer wg_1.Done()

	select {
	case ch_2 <- v * 2:
		// the request is sent to the output channel
	case <-ctx.Done():
		panic("should never enter here")
	}
}

func work(ch_2 chan int, wg_2 *sync.WaitGroup) {
	defer wg_2.Done()

	for vv := range ch_2 {
		// time.Sleep(100)
		fmt.Println(vv)
	}
}

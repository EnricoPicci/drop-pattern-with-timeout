# The drop pattern with a worker pool

In this example we implement the drop pattern using a worker pool.

THe worker pool, by default, is balanced in terms of its throughput and the frequency of incoming requests. In other words the number of workers is able to process the requests coming in without too much delay for the requests and without too much idle time of the workers. The workers are not idle for long and the requests do not wait too long to enter in the pool to be processed.

This is obtained having a pool size of 10 (10 workers), a processing time of 1000ms and an interval between 2 requests of 100ms. This means that in 1 sec (the time a worker takes to process a request) there are 10 requests coming in and 10 workers able to process them. This is a situation of equilibrium.

If an event occurs, i.e. at a certain moment for whatever reason the workers stop processing the requests for a certain duration, than this delay is going to hit the subsequent requests.

The timeout provided ensures that if a request waits too long (i.e. more than the timeout) before entering the pool to be processed, than it is dropped

These parameters have default values which can be overridden by command line params

- poolSize: number of workers in the worker pool
- reqInterval: interval between requests in miliseconds
- procTime: the it takes for a worker to process a request
- numReq: number of requests coming in to be processed
- haltPoolTime: at which time (after start) the pool is halted and stops processing in milliseconds
- haltPoolDuration: for how long the pool is halted in milliseconds
- timeout: the timeout after which an incoming request not yet taken by the worker pool is dropped

## build

From the root project folder run the command
`go build -o ./bin/drop-pattern ./src/drop-pattern`

## run

Playing with the parameters passed by command line allows to see how, in different conditions, the system behaves differently.

### a balanced system

A pool with 10 workers, able to process each request in 1 sec, with a flow of 10 requests per second coming in is a balanced system, i.e. a system where we do not expect to see requests wait too much to be taken in by one worker and where we do not expect to see the workers stay idle for long.

The average wait time (i.e. the time a request spend before being taken in by one worker) and the average idle time for workers (i.e.the time a worker spends idle just waiting for the next request to come in) are printed on the console at the end of the processing.

In this case no request is dropped

From the root project folder run the command
`./bin/drop-pattern -poolSize 10 -reqInterval 100 -procTime 1000 -numReq 100 -haltPoolDuration 0 -timeout 500`

### the worker pool stops working for a certain period

If a sudden event occurs which halts all workers at a certain point in time for a certain duration, then this introduces a delay in the processing of the requests entering after this event and this delay is propagated over time.

With the following command we introduce an halt of 2 secs after the 1 sec from the start of the processing. We can see that the requests in the pool when the block occurs are affected by a wait time of about 2 secs, i.e. the duration of the halt.

Some requests trying to enter the pool during the block are dropped because their timeout expires.

After the block has been removed and the worker pool is brough back to normal operation, the requests not dropped will be afftected by a wait time which is just a bit smaller than the timeout (i.e. the delay introduced by the block is limited to the timeout)

From the root project folder run the command
`./bin/drop-pattern -poolSize 10 -reqInterval 100 -procTime 1000 -numReq 100 -haltPoolDuration 2000 - haltPoolTime 1000 -timeout 500`

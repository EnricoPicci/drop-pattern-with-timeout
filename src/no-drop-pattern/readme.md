# Worker pool without drop pattern

In this example of a worker pool where there is no drop pattern implemented.

THe worker pool, by default, is balanced in terms of its throughput and the frequency of incoming requests. In other words the number of workers is able to process the requests coming in without too much delay for the requests and without too much idle time of the workers. The workers are not idle for long and the requests do not wait too long to enter in the pool to be processed.

This is obtained having a pool size of 10 (10 workers), a processing time of 1000ms and an interval between 2 requests of 100ms. This means that in 1 sec (the time a worker takes to process a request) there are 10 requests coming in and 10 workers able to process them. This is a situation of equilibrium.

If an event occurs, i.e. at a certain moment for whatever reason the workers stop processing the requests for a certain duration, than this delay is going to hit all subsequent requests. In other words all subsequent requests will show a delay in their processing.

These parameters have default values which can be overridden by command line params

- poolSize: number of workers in the worker pool
- reqInterval: interval between requests in miliseconds
- procTime: the it takes for a worker to process a request
- numReq: number of requests coming in to be processed
- haltPoolTime: at which time (after start) the pool is halted and stops processing in milliseconds
- haltPoolDuration: for how long the pool is halted in milliseconds

## build

From the root project folder run the command
`go build -o ./bin/no-drop-pattern ./src/no-drop-pattern`

## run

Playing with the parameters passed by command line allows to see how, in different conditions, the system behaves differently.

### a balanced system

A pool with 10 workers, able to process each request in 1 sec, with a flow of 10 requests per second coming in is a balanced system, i.e. a system where we do not expect to see requests wait too much to be taken in by one worker and where we do not expect to see the workers stay idle for long.

The average wait time (i.e. the time a request spend before being taken in by one worker) and the average idle time for workers (i.e.the time a worker spends idle just waiting for the next request to come in) are printed on the console at the end of the processing.

From the root project folder run the command
`./bin/no-drop-pattern -poolSize 10 -reqInterval 100 -procTime 1000 -numReq 100 -haltPoolDuration 0`

### a delay hit all requests after a certain moment

If a sudden event occurs which halts all workers at a certain point in time for a certain duration, then this introduces a delay in the processing of the requests entering after this event and this delay is propagated over time.

With the following command we introduce an halt of 2 secs after the 1 sec from the start of the processing. We can see that all requests processed after the halt periodo has ended are affected by a wait time of about 2 secs, i.e. the duration of the halt.

This is similar to what happens in an high traffic highway. When there are no blockers, all the cars flow more or less at the same speed given that the traffic is high and nobody can go too fast (there is not much space between one car and the next and all lanes are busy). As soon a block occurs, e.g. for a car crash, a queue starts and cars are blocked. When the obstacle is removed and traffic can get back to normal, if the traffic remains heavy, for many hours cars will have to stop for more or less the same amount of time that the previous cars spent blocked.

From the root project folder run the command
`./bin/no-drop-pattern -poolSize 10 -reqInterval 100 -procTime 1000 -numReq 100 -haltPoolDuration 2000 - haltPoolTime 1000`

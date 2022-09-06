# The drop pattern with timeout

In this repo there is an example of implementation in Go of the "drop pattern with timeout", described in some details in [this article](https://medium.com/@enrico-piccinin/drop-pattern-with-timeout-in-go-29c41b0488b7).

The actual implmentation can be found in the [src/drop-pattern](./src/drop-pattern/) folder and is described in this [readme.md file](./src/drop-pattern/readme).

The folder [src/no-drop-pattern](./src/no-drop-pattern/) implements the same example without using any pattern to control backpressure. It can be used to compare the results.

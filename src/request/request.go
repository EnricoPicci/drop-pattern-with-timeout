package request

import "time"

type Request struct {
	Param        int
	Created      time.Time
	WaitDuration time.Duration
}

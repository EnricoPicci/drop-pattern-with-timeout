package request

import "time"

type Request struct {
	Param        int
	Created      time.Time
	WaitDuration time.Duration
}

func (r *Request) GetParam() any {
	return r.Param
}
func (r *Request) GetCreated() time.Time {
	return r.Created
}
func (r *Request) GetWaitDuration() time.Duration {
	return r.WaitDuration
}
func (r *Request) SetWaitDuration(d time.Duration) {
	r.WaitDuration = d
}

package signals

import "time"

type SampleState int32

const (
	StateOK SampleState = iota
	StateBAD
	StateUNKNOWN
)

type Sample struct {
	At         time.Time
	State      SampleState
	Definitive bool
	Detail     string
}

package clock

import "time"

// Real implements the Clock port returning the actual wall-clock time.
// Use in production wiring; inject a fake/stub in tests.
type Real struct{}

func (Real) Now() time.Time { return time.Now() }

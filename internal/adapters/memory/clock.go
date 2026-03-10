package memory

import "time"

// RealClock returns the actual wall-clock time. Used in production wiring.
type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now() }

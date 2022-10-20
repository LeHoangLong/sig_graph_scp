package utility

import "time"

type ClockI interface {
	Now() time.Time
}

package utility

import "time"

type ClockWall struct {
}

func NewClockWall() *ClockWall {
	return &ClockWall{}
}

func (c *ClockWall) Now() time.Time {
	return time.Now()
}

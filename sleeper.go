package screw

import (
	"math/rand"
	"time"
)

type sleeper struct {
	intervals []int
}

func newSleeper() *sleeper {
	return &sleeper{
		intervals: []int{
			100,
			200,
			400,
			800,
			1600,
			3200,
		},
	}
}

func (s *sleeper) Sleep(err error) bool {
	if len(s.intervals) == 0 {
		debugf("done sleeping, despite error %+v", err)
		return false
	}

	interval := s.intervals[0]
	s.intervals = s.intervals[1:]

	sleepTime := time.Duration(interval+rand.Intn(50)) * time.Millisecond
	debugf("sleeping %v because of error %+v", sleepTime, err)
	time.Sleep(sleepTime)

	return true
}

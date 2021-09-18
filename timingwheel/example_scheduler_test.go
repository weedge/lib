package timingwheel

import (
	"fmt"
	"time"
)

type EveryScheduler struct {
	Interval time.Duration
}

func (s *EveryScheduler) Next(prev time.Time) time.Time {
	return prev.Add(s.Interval)
}

func Example_scheduleTimer() {
	tw := NewTimingWheel(time.Millisecond, 20)
	tw.Start()
	defer tw.Stop()

	exitC := make(chan time.Time)
	t := tw.ScheduleFunc(&EveryScheduler{3*time.Second}, func() {
		fmt.Println("The timer fires")
		exitC <- time.Now().UTC()
	})

	<-exitC
	<-exitC

	// We need to stop the timer since it will be restarted again and again.
	for !t.Stop() {
	}

	// Output:
	// The timer fires
	// The timer fires
}

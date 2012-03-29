package ircclient

// Implement Undernet ircu's throttling algorithm here

import (
	"time"
)

type throttleIrcu struct {
	lastsent  time.Time
	tscounter time.Time
}

func newthrottleIrcu() *throttleIrcu {
	// Could be handy later
	return &throttleIrcu{lastsent: time.Now(), tscounter: time.Now()}
}

func (tm *throttleIrcu) WaitSend(line string) {
	tm.lastsent = time.Now()
	if tm.lastsent.Sub(tm.tscounter) > 0 {
		tm.tscounter = time.Now()
	}
	tm.tscounter = tm.tscounter.Add(time.Duration(2+len(line)/120) * time.Second)
	t := tm.tscounter.Sub(time.Now())
	if t-(10*time.Second) > 0 {
		time.Sleep(t - (10 * time.Second))
	}
	return
}

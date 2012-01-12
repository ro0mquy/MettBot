package ircclient

// Implement Undernet ircu's throttling algorithm here

import (
	"time"
)

type throttleIrcu struct {
	lastsent  int64
	tscounter int64
}

func newthrottleIrcu() *throttleIrcu {
	// Could be handy later
	return &throttleIrcu{lastsent: time.Seconds(), tscounter: time.Seconds()}
}

func (tm *throttleIrcu) WaitSend(line string) {
	tm.lastsent = time.Seconds()
	if tm.lastsent > tm.tscounter {
		tm.tscounter = time.Seconds()
	}
	tm.tscounter += int64((2 + len(line)/120))
	t := tm.tscounter - time.Seconds()
	if t > 10 {
		time.Sleep((t - 10) * 1e9)
	}
	return
}

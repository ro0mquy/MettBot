package ircconn

// Implement Undernet ircu's throttling algorithm here

import (
	"time"
	"fmt"
)

type ThrottleMgr struct {
	lastsent int64
	tscounter int64
}

func NewThrottleMgr() *ThrottleMgr {
	// Could be handy later
	return &ThrottleMgr{lastsent: time.Seconds(), tscounter: time.Seconds()}
}

func (tm *ThrottleMgr) WaitSend(line string) {
	tm.lastsent = time.Seconds()
	if(tm.lastsent > tm.tscounter) {
		tm.tscounter = time.Seconds()
	}
	tm.tscounter += int64((2 + len(line) / 120))
	t := tm.tscounter - time.Seconds()
	fmt.Println(t)
	if t > 10 {
		time.Sleep((t - 10) * 1e9)
	}
	return
}

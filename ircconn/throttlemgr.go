package ircconn

import (
)

type ThrottleMgr struct {
}

func NewThrottleMgr() *ThrottleMgr {
	// Could be handy later
	return &ThrottleMgr{}
}

func (tm *ThrottleMgr) WaitSend(line string) {
	// TODO: More sophisticated logic
	return
}

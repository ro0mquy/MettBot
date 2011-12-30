package ircclient

// Handles basic IRC protocol messages (like PING)

import (
	"log"
	"time"
)

type BasicProtocol struct {
	ic       *IRCClient
	timer    *time.Timer
	lastping int64
	done     chan bool
}

func (bp *BasicProtocol) Register(cl *IRCClient) {
	bp.ic = cl
	bp.done = make(chan bool)
	// Send a PING message every few minutes to detect locked-up
	// server connection
	go func() {
		for {
			select {
			case _ = <-bp.done:
				return
			default:
			}
			time.Sleep(1e9) // TODO
			if bp.lastping != 0 {
				continue
			}
			bp.lastping = time.Seconds()
			bp.ic.conn.Output <- "PING :client"
			bp.timer = time.NewTimer(5e9) // TODO
			go func() {
				select {
				case _ = <-bp.timer.C:
					bp.ic.Disconnect("(Client) timer expired")
				}
			}()
		}
	}()
}
func (bp *BasicProtocol) String() string {
	return "basic"
}
func (bp *BasicProtocol) ProcessLine(msg *IRCMessage) {
	switch msg.Command {
	case "PING":
		if len(msg.Args) != 1 {
			log.Printf("WARNING: Invalid PING received")
		}
		bp.ic.conn.Output <- "PONG :" + msg.Args[0]
	case "PONG":
		bp.lastping = 0
		bp.timer.Stop()
	}
}
func (bp *BasicProtocol) Unregister() {
	bp.done <- true
}
func (bp *BasicProtocol) Info() {
}
func (bp *BasicProtocol) ProcessCommand(cmd *IRCCommand) {
}

package ircclient

// Handles basic IRC protocol messages (like PING)

import (
	"log"
)

type BasicProtocol struct {
	ic *IRCClient
}

func (bp *BasicProtocol) Register(cl *IRCClient) {
	bp.ic = cl
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
	}
}
func (bp *BasicProtocol) Unregister() {
}

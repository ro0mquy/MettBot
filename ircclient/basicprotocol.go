package ircclient

// Handles basic IRC protocol messages (like PING)

import (
	"log"
)

type basicProtocol struct {
	ic       *IRCClient
}

func (bp *basicProtocol) Register(cl *IRCClient) {
	bp.ic = cl
}

func (bp *basicProtocol) String() string {
	return "basic"
}

func (bp *basicProtocol) Usage(cmd string) string {
	// stub, no commands here
	return ""
}

func (bp *basicProtocol) ProcessLine(msg *IRCMessage) {
	switch msg.Command {
	case "PING":
		if len(msg.Args) != 1 {
			log.Printf("WARNING: Invalid PING received")
		}
		bp.ic.conn.Output <- "PONG :" + msg.Args[0]
	}
}
func (bp *basicProtocol) Unregister() {
}

func (bp *basicProtocol) Info() string {
	return "basic irc protocol (e.g. PING), implemented as plugin."
}

func (bp *basicProtocol) ProcessCommand(cmd *IRCCommand) {
}

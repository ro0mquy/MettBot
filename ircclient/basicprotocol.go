package ircclient

// Handles basic IRC protocol messages (like PING)

import (
	"log"
	"time"
)

type basicProtocol struct {
	ic          *IRCClient
	chanTimeout chan *IRCMessage
}

func (bp *basicProtocol) Register(cl *IRCClient) {
	bp.ic = cl
	bp.chanTimeout = make(chan *IRCMessage)

	// start pinging routine
	go func() {
		for {
			time.Sleep(2 * time.Minute)
			bp.ic.SendLine("PING :" + bp.ic.GetStringOption("Server", "nick"))

			select {
			case <-bp.chanTimeout:
			case <-time.After(20 * time.Second):
				log.Println("Ping timeout")
				bp.ic.Disconnect("Ping timeout")
			}

		}
	}()
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
		bp.ic.SendLine("PONG :" + msg.Args[0])
	case "PONG":
		bp.chanTimeout <- msg
	}
}
func (bp *basicProtocol) Unregister() {
}

func (bp *basicProtocol) Info() string {
	return "basic irc protocol (e.g. PING), implemented as plugin."
}

func (bp *basicProtocol) ProcessCommand(cmd *IRCCommand) {
}

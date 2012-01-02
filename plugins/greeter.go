package plugins

import (
	"ircclient"
	"fmt"
	"log"
)

type Greeter struct {
	ic *ircclient.IRCClient
}


func (g *Greeter) Register(ic *ircclient.IRCClient) {
	g.ic= ic
}

func (g *Greeter) String() string {
	return "greeter"
}

func (g *Greeter) Info() string {
	return "greets everyone joining a channel"
}

func (g *Greeter) ProcessLine(msg *ircclient.IRCMessage) {
	if msg.Command != "JOIN" { // TODO: numeric command
		return
	}
	ret := fmt.Sprintf("PRIVMSG %s hi %s", msg.Target, msg.Source)
	log.Println("greeter ret: " + ret)
	g.ic.SendLine(ret)
}

func (g *Greeter) ProcessCommand(cmd *ircclient.IRCCommand) {
	return
}

func (g *Greeter) Unregister() {
	// nothing to do here
}


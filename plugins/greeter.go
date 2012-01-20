package plugins

import (
	"ircclient"
	"fmt"
	"log"
	"strings"
)

type Greeter struct {
	ic *ircclient.IRCClient
}

func (g *Greeter) Register(ic *ircclient.IRCClient) {
	g.ic = ic
}

func (g *Greeter) String() string {
	return "greeter"
}

func (g *Greeter) Info() string {
	return "greets everyone joining a channel"
}

func (g *Greeter) ProcessLine(msg *ircclient.IRCMessage) {
	if msg.Command != "JOIN" {
		return
	}
	just_nick := strings.SplitN(msg.Source, "!", 2)[0]
	c, exists := g.ic.GetPlugin("config")
	if !exists {
		log.Println("plugin \"config\" doesn't exist")
		return
	}
	conf, _ := c.(*ircclient.ConfigPlugin)
	conf.Lock() // muss ich bei read-only garned, oder?
	our_nick, _ := conf.Conf.String("Server", "nick")
	conf.Unlock()

	if just_nick == our_nick { // don't welcome ourselves
		log.Println("o: " + our_nick + ", j: " + just_nick)
		return
	}
	ret := fmt.Sprintf("PRIVMSG %s :hi %s", msg.Target, just_nick)
	//log.Println("greeter ret: " + ret)
	g.ic.SendLine(ret)
}

func (g *Greeter) ProcessCommand(cmd *ircclient.IRCCommand) {
	return
}

func (g *Greeter) Unregister() {
	// nothing to do here
}

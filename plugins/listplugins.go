package plugins

import (
	"ircclient"
	"strings"
	"log"
	"fmt"
)

type ListPlugins struct {
	ic *ircclient.IRCClient
}

func (lp *ListPlugins) Register(ic *ircclient.IRCClient) {
	lp.ic = ic
}

func (lp ListPlugins) String() string {
	return "listplugins"
}

func (lp ListPlugins) Info() string {
	return "Lists all currently registered Plugins"
}

func (lp ListPlugins) ProcessLine(msg *ircclient.IRCMessage) {
	return
}

/**
 * the array-foo makes it easy to leave out the last ", "
 * because strings.Join() does that for us
 **/
func (lp ListPlugins) ProcessCommand(cmd *ircclient.IRCCommand) {
	if cmd.Command != "listplugins" {
		return
	}

	a := make([]string, 0)
	for plug := range lp.ic.IterPlugins() {
		a = append(a, plug.String())
	}

	// channel vs. query
	var ret_targ string
	if cmd.Target[0] == '#' {
		ret_targ = cmd.Target
	} else {
		ret_targ = cmd.Source
	}
	ret := fmt.Sprintf("PRIVMSG %s :%s", ret_targ, strings.Join(a, ", "))
	log.Println("listplugins output: " + ret)
	lp.ic.SendLine(ret)
}

func (lp ListPlugins) Unregister() {
	// nothing to do here
}

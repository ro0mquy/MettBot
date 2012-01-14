package plugins

import (
	"ircclient"
	"strings"
)

type ListPlugins struct {
	ic *ircclient.IRCClient
}

func (lp *ListPlugins) Register(ic *ircclient.IRCClient) {
	lp.ic = ic
	ic.RegisterCommandHandler("listplugins", 0, 0, lp)
}

func (lp *ListPlugins) String() string {
	return "listplugins"
}

func (lp *ListPlugins) Info() string {
	return "Lists all currently registered Plugins"
}

func (lp *ListPlugins) ProcessLine(msg *ircclient.IRCMessage) {
	return
}

/**
 * the array-foo makes it easy to leave out the last ", "
 * because strings.Join() does that for us
 **/
func (lp *ListPlugins) ProcessCommand(cmd *ircclient.IRCCommand) {
	a := make([]string, 0)
	for plug := range lp.ic.IterPlugins() {
		a = append(a, plug.String())
	}

	lp.ic.Reply(cmd, strings.Join(a, ", "))
}

func (lp *ListPlugins) Unregister() {
	// nothing to do here
}

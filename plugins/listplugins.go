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
	ic.RegisterCommandHandler("listcommands", 0, 0, lp)
}

func (lp *ListPlugins) String() string {
	return "listplugins"
}

func (lp *ListPlugins) Info() string {
	return "Lists all currently registered plugins and commands"
}

func (lp *ListPlugins) ProcessLine(msg *ircclient.IRCMessage) {
	return
}

func (lp *ListPlugins) Usage(cmd string) string {
	switch cmd {
	case "listplugins":
		return "listplugins: lists all loaded plugins"
	case "help":
		fallthrough
	case "listcommands":
		return cmd + ": list all available commands"
	}
	return ""
}

/**
 * the array-foo makes it easy to leave out the last ", "
 * because strings.Join() does that for us
 **/
func (lp *ListPlugins) ProcessCommand(cmd *ircclient.IRCCommand) {
	switch cmd.Command {
	case "listplugins":
		a := make([]string, 0)
		for plug := range lp.ic.IterPlugins() {
			a = append(a, plug.String())
		}

		lp.ic.Reply(cmd, strings.Join(a, ", "))
	case "help":
		if len(cmd.Args) == 0 {
			fallthrough // alias for listcommands
		} else {
			return lp.ic.GetUsage(cmd.Args[0])
		}
	case "listcommands":
		c := lp.ic.IterHandlers()
		commands := ""
		for e := range c {
			if commands != "" {
				commands += ", "
			}
			commands += e.Command
		}
		lp.ic.Reply(cmd, commands)
	}
}

func (lp *ListPlugins) Unregister() {
	// nothing to do here
}

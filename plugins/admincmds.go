package plugins

import (
	"strings"
	"ircclient"
)

type AdminPlugin struct {
	ic *ircclient.IRCClient
	authplugin *AuthPlugin
}

func (q *AdminPlugin) Register(cl *ircclient.IRCClient) {
	q.ic = cl
	authplugin, _ := q.ic.GetPlugin("auth")
	q.authplugin = authplugin.(*AuthPlugin)
}

func (q *AdminPlugin) String() string {
	return "admin"
}

func (q *AdminPlugin) Info() string {
	return "provides commands for bot-admins"
}

func (q *AdminPlugin) ProcessLine(msg *ircclient.IRCMessage) {
	if msg.Command != "JOIN" {
		return
	}
	if q.authplugin.GetAccessLevel(msg.Source) >= 200 {
		q.ic.SendLine("MODE " + msg.Target + " +o " + strings.SplitN(msg.Source, "!", 2)[0])
		return
	}
}

func (q *AdminPlugin) ProcessCommand(cmd *ircclient.IRCCommand) {
	if q.authplugin.GetAccessLevel(cmd.Source) < 400 {
		// Don't send any error messages
		return
	}
	if cmd.Command == "inviteme" && len(cmd.Args) == 1 {
		q.ic.SendLine("INVITE " + strings.SplitN(cmd.Source, "!", 2)[0] + " " + cmd.Args[0])
		return
	}
}

func (q *AdminPlugin) Unregister() {
	return
}

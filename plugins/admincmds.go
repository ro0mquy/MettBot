package plugins

import (
	"strings"
	"ircclient"
)

type AdminPlugin struct {
	ic *ircclient.IRCClient
}

func (q *AdminPlugin) Register(cl *ircclient.IRCClient) {
	q.ic = cl
	q.ic.RegisterCommandHandler("inviteme", 1, 400, q)
}

func (q *AdminPlugin) String() string {
	return "admin"
}

func (q *AdminPlugin) Info() string {
	return "provides commands for bot-admins"
}

func (q *AdminPlugin) Usage(cmd string) string {
	switch cmd {
	case "inviteme":
		return "inviteme <chanelname>"
	}
	return ""
}

func (q *AdminPlugin) ProcessLine(msg *ircclient.IRCMessage) {
	if msg.Command != "JOIN" {
		return
	}
	if q.ic.GetAccessLevel(msg.Source) >= 200 {
		q.ic.SendLine("MODE " + msg.Target + " +o " + strings.SplitN(msg.Source, "!", 2)[0])
		return
	}
}

func (q *AdminPlugin) ProcessCommand(cmd *ircclient.IRCCommand) {
	q.ic.SendLine("INVITE " + strings.SplitN(cmd.Source, "!", 2)[0] + " " + cmd.Args[0])
}

func (q *AdminPlugin) Unregister() {
	return
}

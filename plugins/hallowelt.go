package plugins

import (
	"ircclient"
)

type HalloWeltPlugin struct {
	ic *ircclient.IRCClient
}

func (q *HalloWeltPlugin) Register(cl *ircclient.IRCClient) {
	q.ic = cl
}

func (q *HalloWeltPlugin) String() string {
	return "hallowelt"
}

func (q *HalloWeltPlugin) Info() string {
	return "DomJudge live ticker"
}

func (q *HalloWeltPlugin) Usage(cmd string) string {
	return "This plugin provides no commands"
}

func (q *HalloWeltPlugin) ProcessLine(msg *ircclient.IRCMessage) {
}

func (q *HalloWeltPlugin) ProcessCommand(cmd *ircclient.IRCCommand) {
}

func (q *HalloWeltPlugin) Unregister() {
}

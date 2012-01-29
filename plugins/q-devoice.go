package plugins

import (
	"strings"
	"ircclient"
	"fmt"
)

type QDevoicePlugin struct {
	ic *ircclient.IRCClient
}

func (q *QDevoicePlugin) Register(cl *ircclient.IRCClient) {
	q.ic = cl
}

func (q *QDevoicePlugin) String() string {
	return "q-devoice"
}

func (q *QDevoicePlugin) Info() string {
	return "automatically de-voices people who got voice by Norad after saying \"!q\""
}

func (q *QDevoicePlugin) Usage(cmd string) string {
	// stub for interface satisfaction
	return ""
}

func (q *QDevoicePlugin) ProcessLine(msg *ircclient.IRCMessage) {
	if len(msg.Args) >= 1 && strings.SplitN(msg.Args[0], " ", 2)[0] == "!q" && strings.Index(msg.Source, "siccegge") == 0 {
		line := fmt.Sprintf("MODE %s +v siccegge", msg.Target, "siccegge")
		q.ic.SendLine(line)
	}
	if strings.Index(msg.Source, "cl-faui2k9") == 0 && msg.Command == "MODE" &&
		msg.Args[0] == "+v" && len(msg.Args) > 1 {
		line := fmt.Sprintf("MODE %s -v :%s", msg.Target, msg.Args[1])
		q.ic.SendLine(line)
	}
}

func (q *QDevoicePlugin) ProcessCommand(cmd *ircclient.IRCCommand) {
	return
}

func (q *QDevoicePlugin) Unregister() {
	return
}

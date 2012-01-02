package plugins


import (
	"ircclient"
	"fmt"
	"log"
)


type QDevoicePlugin struct {
	ic *ircclient.IRCClient
}

func (q *QDevoicePlugin) Register(cl *ircclient.IRCClient) {
	q.ic= cl
}

func (q *QDevoicePlugin) String() string {
	return "q-devoice"
}

func (q *QDevoicePlugin) Info() string {
	return "automatically de-voices people who got voice by Norad after saying \"!q\""
}

func (q *QDevoicePlugin) ProcessLine(msg *ircclient.IRCMessage) {
	return
}

func (q *QDevoicePlugin) ProcessCommand(cmd *ircclient.IRCCommand) {
	if cmd.Source == "cl-faui2k9" && cmd.Command == "MODE" &&
	   cmd.Args[0] == "+v" && len(cmd.Args) > 1 {
		   line := fmt.Sprintf("MODE %s -v :%s", cmd.Target, cmd.Args[1])
		   log.Println("devoice: " + line)
		   q.ic.SendLine(line)
	   }
}

func (q *QDevoicePlugin) Unregister() {
	return
}


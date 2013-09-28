package plugins

import (
	"../ircclient"
	"../answers"
	"strings"
)

type DongPlugin struct {
	ic *ircclient.IRCClient
}

func (q *DongPlugin) String() string {
	return "dong"
}

func (q *DongPlugin) Info() string {
	return `sends back a dong sound for every \a`
}

func (q *DongPlugin) Usage(cmd string) string {
	// just for interface saturation
	return ""
}

func (q *DongPlugin) Register(cl *ircclient.IRCClient) {
	q.ic = cl
}

func (q *DongPlugin) Unregister() {
	return
}

func (q *DongPlugin) ProcessLine(msg *ircclient.IRCMessage) {
	if msg.Command != "PRIVMSG" {
		//only process messages from chatrooms
		return
	}

	count := strings.Count(msg.Args[0], `\a`)
	count -= strings.Count(msg.Args[0], `\\a`)
	if count < 1 {
		return
	}

	str := answers.RandStr("dong")
	message := ""

	for i := 0; i < count; i++ {
		message += str + " "
	}

	// don't send messages longer than 400 chars, so we won't hit the 512 chars limit
	if len(message) > 400 {
		message = message[:400]
	}

	var target string
	if msg.Target != q.ic.GetStringOption("Server", "nick") {
		target = msg.Target
	} else {
		target = strings.SplitN(msg.Source, "!", 2)[0]
	}
	q.ic.SendLine("NOTICE " + target + " :" + message)
}

func (q *DongPlugin) ProcessCommand(cmd *ircclient.IRCCommand) {
	return
}

package plugins

import (
	"ircclient"
	"log"
)

const (
	quit_min_auth_level = 300
	default_quit_msg    = "Bye."
)

type QuitHandler struct {
	ic *ircclient.IRCClient
}

func (q *QuitHandler) Register(ic *ircclient.IRCClient) {
	q.ic = ic

	if q.ic.GetStringOption("Quit", "quitmsg") == "" {
		log.Println("added default quitmsg value of \"" + default_quit_msg + "\" to config file")
		q.ic.SetStringOption("Quit", "quitmsg", default_quit_msg)
	}

	q.ic.RegisterCommandHandler("quit", 0, quit_min_auth_level, q)
}

func (q *QuitHandler) String() string {
	return "quit"
}

func (q *QuitHandler) Info() string {
	return "handles the quit command"
}

func (q *QuitHandler) Usage(cmd string) string {
	switch cmd {
	case "quit":
		return "quit: quits this bot"
	}
}

func (q *QuitHandler) ProcessLine(msg *ircclient.IRCMessage) {
	// empty
}

func (q *QuitHandler) ProcessCommand(cmd *ircclient.IRCCommand) {
	q.ic.Disconnect(q.ic.GetStringOption("Quit", "quitmsg"))
}

func (q *QuitHandler) Unregister() {
	// empty
}

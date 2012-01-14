package plugins

import (
	"ircclient"
	"log"
	"fmt"
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

	if q.ic.GetStringOption("Quit", "quit_minlevel") == "" {
		q.ic.SetIntOption("Quit", "quit_minlevel", quit_min_auth_level)
		log.Println("added default quit_minlevel value of \"" + fmt.Sprintf("%d", quit_min_auth_level) + "\" to config file")
		// no return either, sorry ;)
	}

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

func (q *QuitHandler) ProcessLine(msg *ircclient.IRCMessage) {
	// empty
}

func (q *QuitHandler) ProcessCommand(cmd *ircclient.IRCCommand) {
	q.ic.Disconnect(q.ic.GetStringOption("Quit", "quitmsg"))
}

func (q *QuitHandler) Unregister() {
	// empty
}

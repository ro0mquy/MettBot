// quit command is not that much config-related,
// and i wanted access control for it anyway, so
// while i'm at it, i decided to put it into an extra file

package plugins

import (
	"ircclient"
	//"log"
)

const (
	quit_min_auth_level = "300"
	default_quit_msg    = "Bye."
)

type QuitHandler struct {
	ic *ircclient.IRCClient
}

func (q *QuitHandler) Register(ic *ircclient.IRCClient) {
	q.ic = ic
}

func (q *QuitHandler) String() string {
	return "quit"
}

func (q *QuitHandler) Info() string {
	return "handles the quit command with authentication"
}

func (q *QuitHandler) ProcessLine(msg *ircclient.IRCMessage) {
	// empty
}

func (q *QuitHandler) ProcessCommand(cmd *ircclient.IRCCommand) {
	if cmd.Command != "quit" {
		return
	}
	lvl := q.ic.GetAccessLevel(cmd.Source)

	_ = lvl
	// TODO!
	/*
		if ! q.config.Conf.HasSection("Quit") {
			log.Println("no \"Quit\" section.. adding one for your convenience")
			q.config.Conf.AddSection("Quit")
			// no return here so the next if does its job as well and we have a
			// working default config after just one failed attempt
		}
		if ! q.config.Conf.HasOption("Quit", "quit_minlevel") {
			q.config.Conf.AddOption("Quit", "quit_minlevel", quit_min_auth_level)
			log.Println("added default quit_minlevel value of \"" + quit_min_auth_level + "\" to config file")
			// no return here either, sorry ;)
		}

		lvl_needed, err := q.config.Conf.Int("Quit", "quit_minlevel")
		if err != nil {
			q.ic.Reply(cmd, err.String())
			q.config.Unlock()
			return
		}

		if lvl_needed > lvl {
			q.ic.Reply(cmd, "not authorized to quit this bot")
			q.config.Unlock()
			return
		}

		if ! q.config.Conf.HasOption("Quit", "quitmsg") {
			log.Println("added default quitmsg value of \"" + default_quit_msg + "\" to config file")
			q.config.Conf.AddOption("Quit", "quitmsg", default_quit_msg)
			q.config.Unlock()
			q.ic.Disconnect(default_quit_msg)
		} else {
			quitmsg, err := q.config.Conf.String("Quit", "quitmsg")
			if err != nil {
				q.ic.Reply(cmd, err.String())
			}
			q.config.Unlock()
			q.ic.Disconnect(quitmsg)
		}
	*/
}

func (q *QuitHandler) Unregister() {
	// empty
}

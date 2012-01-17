package plugins

import (
	"strings"
	"ircclient"
	"rand"
	"log"
)

type DecidePlugin struct {
	ic *ircclient.IRCClient
	requests chan *ircclient.IRCMessage
	current *ircclient.IRCMessage
	done bool
}

func (d *DecidePlugin) Register(cl *ircclient.IRCClient) {
	d.ic = cl
	d.requests = make(chan *ircclient.IRCMessage, 64)
}

func (d *DecidePlugin) String() string {
	return "decide"
}

func (d *DecidePlugin) Info() string {
	return "always gives a different answer than cl-faui2k9 does"
}

func (d *DecidePlugin) ProcessLine(msg *ircclient.IRCMessage) {
	if len(msg.Args) == 0 {
		return
	}
	if msg.Args[0] == "!decide" {
		d.requests <- msg
		return
	}
	if strings.Index(msg.Source, "cl-faui2k9") == 0 && msg.Command == "PRIVMSG" {
		if d.done {
			select {
			case d.current = <-d.requests:
				d.done = false
			default:
			}
		}
		if !d.done {
			if strings.Split(d.current.Source, "!")[0] == strings.Split(msg.Args[0], ":")[0] {
				if len(msg.Args) == 1 {
					log.Println("cl-faui2k11 gibt leere Antwort")
					d.done = true
					return
				}
				if len(d.current.Args) <= 2 {
					switch msg.Args[1] {
					case "Yes":
						d.ic.SendLine(d.current.Args[0] + " No")
					case "No":
						d.ic.SendLine(d.current.Args[0] + " Yes")
					}
				} else {
					var i int
					for i = 1; i < len(d.current.Args); i++ {
						if d.current.Args[i] == msg.Args[1] {
							break
						}
					}
					r := rand.Intn(len(d.current.Args) - 2) + 1
					if r >= i {
						r++
					}
					d.ic.SendLine(d.current.Args[0] + d.current.Args[r])
				}
				d.done = true
			}
		}
	}
}

func (d *DecidePlugin) ProcessCommand(cmd *ircclient.IRCCommand) {
	return
}

func (d *DecidePlugin) Unregister() {
	return
}

func (d *DecidePlugin) Usage(cmd string) string {
	return ""
}

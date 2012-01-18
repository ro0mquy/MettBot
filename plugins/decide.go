package plugins

import (
	"strings"
	"ircclient"
	"rand"
	"log"
	"fmt" ///////////////////////i
)

type DecidePlugin struct {
	ic *ircclient.IRCClient
	requests chan *ircclient.IRCMessage
	current *ircclient.IRCMessage
	done bool
}

func (d *DecidePlugin) Register(cl *ircclient.IRCClient) {
	d.done = true
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
	if strings.Index(msg.Args[len(msg.Args) - 1], "!decide") == 0 {
		fmt.Println("push")
		d.requests <- msg
		return
	}
	cmd := ircclient.ParseCommand(msg)
	if strings.Index(cmd.Source, "cl-faui2k9") == 0 && msg.Command == "PRIVMSG" {
		if d.done {
			select {
			case d.current = <-d.requests:
				d.done = false
				fmt.Println("pop")
			default:
				fmt.Println("default")
			}
		}
		if !d.done {
			current := ircclient.ParseCommand(d.current)
			fmt.Println("not done")

			if strings.Split(d.current.Source, "!")[0] == strings.Split(cmd.Command, ":")[0] {
				if len(cmd.Args) == 0 {
					log.Println("cl-faui2k11 gibt leere Antwort")
					d.done = true
					return
				}
				if len(current.Args) <= 1 {
					switch cmd.Args[0] {
					case "Yes":
						d.ic.Reply(cmd, strings.Split(d.current.Source, "!")[0] + ": No")
					case "No":
						d.ic.Reply(cmd, strings.Split(d.current.Source, "!")[0] + ": Yes")
					default:
						fmt.Println("uiae")
					}
				} else {
					var i int
					for i = 0; i < len(current.Args); i++ {
						if current.Args[i] == cmd.Args[0] {
							break
						}
					}
					r := rand.Intn(len(current.Args) - 1)
					if r >= i {
						r++
					}
					d.ic.Reply(cmd, strings.Split(d.current.Source, "!")[0] + ": " + current.Args[r])
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

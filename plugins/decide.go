package plugins

import (
	"strings"
	"ircclient"
	"rand"
	"log"
	"time"
)

type DecidePlugin struct {
	ic *ircclient.IRCClient
	requests chan *ircclient.IRCMessage
	boolchans chan chan bool
	current *ircclient.IRCMessage
	done bool
}

func (d *DecidePlugin) Register(cl *ircclient.IRCClient) {
	d.done = true
	d.ic = cl
	d.requests = make(chan *ircclient.IRCMessage, 64)
	d.boolchans = make(chan chan bool, 64)
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
		d.requests <- msg
		go func() {
			newChan := make(chan bool)
			d.boolchans <- newChan
			t := time.NewTimer(1e10) //time.AfterFunc(1e10, func () { } )
			select {
			case _ = <-newChan:
			case _ = <-t.C:
				_ = <-d.boolchans
				cmd := ircclient.ParseCommand(msg)
				if len(cmd.Args) <= 1 {
					if rand.Intn(2) == 0 {
						d.ic.Reply(cmd, strings.Split(cmd.Source, "!")[0] + ": Yes")
					} else {
						d.ic.Reply(cmd, strings.Split(cmd.Source, "!")[0] + ": No")
					}
				} else {
					d.ic.Reply(cmd, strings.Split(cmd.Source, "!")[0] + ": " + cmd.Args[rand.Intn(len(cmd.Args))])
				}
			}
		} ()
		return
	}
	cmd := ircclient.ParseCommand(msg)
	if strings.Index(cmd.Source, "cl-fau2k9") == 0 && msg.Command == "PRIVMSG" {
		if d.done {
			select {
			case d.current = <-d.requests:
				d.done = false
			default:
			}
		}
		if !d.done {
			current := ircclient.ParseCommand(d.current)
			if strings.Split(d.current.Source, "!")[0] == strings.Split(cmd.Command, ":")[0] {
				(<-d.boolchans) <- true
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

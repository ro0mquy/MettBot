package ircclient

import (
	"os"
	"log"
)

type IRCClient struct {
	conn *IRCConn
	// TODO better config format
	conf map[string]string
	plugins map[string]Plugin
}

func NewIRCClient(hostport, nick, rname, ident string) *IRCClient {
	c := &IRCClient{nil, make(map[string]string), make(map[string]Plugin)}
	c.conf["nick"] = nick
	c.conf["hostport"] = hostport
	c.conf["rname"] = rname
	c.conf["ident"] = ident
	return c
}

func (ic *IRCClient) RegisterPlugin(p Plugin) os.Error {
	if _, ok := ic.plugins[p.String()]; ok == true {
		return os.NewError("Plugin already exists")
	}
	p.Register(ic)
	ic.plugins[p.String()] = p
	return nil
}

func (ic *IRCClient) Connect() os.Error {
	ic.conn = NewIRCConn()
	e := ic.conn.Connect(ic.conf["hostport"])
	if e != nil {
		log.Println("Can't connect " + e.String())
		return e
	}
	ic.conn.Output <- "NICK " + ic.conf["nick"] + "\n"
	ic.conn.Output <- "USER " + ic.conf["ident"] + " * Q :" + ic.conf["rname"] + "\n"
	nick := ic.conf["nick"]
	for {
		s := ParseServerLine(<-ic.conn.Input)
		// TODO: Forward processed commands to plugins
		switch s.Command {
		case "433":
			nick = nick + "_"
			ic.conn.Output <- "NICK " + nick + "\n"
		case "PING":
			ic.conn.Output <- "PONG :" + s.Args[0]
		case "NOTICE":
			// ignore
		case "001":
			return nil
		default:
			log.Fatal("Unexpected server message")
		}
	}
	return nil // Never happens
}

func (ic *IRCClient) Disconnect(quitmsg string) {
	ic.conn.Output <- "QUIT :" + quitmsg
	ic.conn.Flush()
	ic.conn.Quit()	// Shouldn't be needed, as server closes connection
}

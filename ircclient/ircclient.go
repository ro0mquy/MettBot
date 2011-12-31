package ircclient

import (
	"os"
	"log"
)

type IRCClient struct {
	conn *IRCConn
	// TODO better config format
	conf    map[string]string
	plugins map[string]Plugin
}

func NewIRCClient(hostport, nick, rname, ident string) *IRCClient {
	c := &IRCClient{nil, make(map[string]string), make(map[string]Plugin)}
	c.conf["nick"] = nick
	c.conf["hostport"] = hostport
	c.conf["rname"] = rname
	c.conf["ident"] = ident
	c.RegisterPlugin(&BasicProtocol{})
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

func (ic IRCClient) GetPlugins() map[string]Plugin {
	return ic.plugins
}

func (ic *IRCClient) Connect() os.Error {
	ic.conn = NewIRCConn()
	e := ic.conn.Connect(ic.conf["hostport"])
	if e != nil {
		log.Println("Can't connect " + e.String())
		return e
	}
	ic.conn.Output <- "NICK " + ic.conf["nick"]
	ic.conn.Output <- "USER " + ic.conf["ident"] + " * Q :" + ic.conf["rname"]
	nick := ic.conf["nick"]
	for {
		line, ok := <-ic.conn.Input
		if !ok {
			return <-ic.conn.Err
		}
		s := ParseServerLine(line)
		if s == nil {
			// Ignore empty lines
			continue
		}
		for _, p := range ic.plugins {
			p.ProcessLine(s)
		}
		switch s.Command {
		case "433":
			// Nickname already in use
			nick = nick + "_"
			ic.conn.Output <- "NICK " + nick
		case "001":
			// Successfully registered
			return nil
		}
	}
	return nil
}

func (ic *IRCClient) InputLoop() os.Error {
	for {
		in, ok := <-ic.conn.Input
		if !ok {
			return <-ic.conn.Err
		}

		s:= ParseServerLine(in)
		if s == nil {
			continue
		}

		for _, p := range ic.plugins {
			go p.ProcessLine(s)
		}
	}
	panic("This never happens")
}

func (ic *IRCClient) Disconnect(quitmsg string) {
	ic.conn.Output <- "QUIT :" + quitmsg
	ic.conn.Flush()
	ic.conn.Quit()
}

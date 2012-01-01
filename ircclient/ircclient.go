package ircclient

import (
	"os"
)

type IRCClient struct {
	conn *IRCConn
	// TODO better config format
	conf    map[string]string
	plugins map[string]Plugin
}

func NewIRCClient(hostport, nick, rname, ident string, trigger byte) *IRCClient {
	c := &IRCClient{nil, make(map[string]string), make(map[string]Plugin)}
	c.conf["nick"] = nick
	c.conf["hostport"] = hostport
	c.conf["rname"] = rname
	c.conf["ident"] = ident
	c.conf["trigger"] = string(trigger) // XXX
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

		// Invoke plugin line handlers.
		// At this point, it makes no sense to
		// process "commands". If a plugin needs
		// interaction in this state, it should be
		// low-level.
		s := ParseServerLine(line)
		if s == nil {
			continue
		}
		for _, p := range ic.plugins {
			go p.ProcessLine(s)
		}

		switch s.Command {
		case "433":
			// Nickname already in use
			nick = nick + "_"
			ic.conf["nick"] = nick
			ic.conn.Output <- "NICK " + nick
		case "001":
			// Successfully registered
			return nil
		}
	}
	return nil
}

func (ic *IRCClient) dispatchHandlers(in string) {
	var c *IRCCommand = nil

	s := ParseServerLine(in)
	if s == nil {
		return
	}
	if (s.Command == "PRIVMSG" || s.Command == "NOTICE") && (s.Target == ic.conf["nick"] || s.Args[0][0] == ic.conf["trigger"][0]) {
		c = ParseCommand(s)
		// Strip trigger, if necessary
		if c != nil && s.Target != ic.conf["nick"] && len(c.Command) != 0 {
			c.Command = c.Command[1:len(c.Command)]
		}
	}

	for _, p := range ic.plugins {
		go p.ProcessLine(s)
		if c != nil {
			go p.ProcessCommand(c)
		}
	}
}

func (ic *IRCClient) InputLoop() os.Error {
	for {
		in, ok := <-ic.conn.Input
		if !ok {
			return <-ic.conn.Err
		}
		ic.dispatchHandlers(in)
	}
	panic("This never happens")
}

func (ic *IRCClient) Disconnect(quitmsg string) {
	ic.conn.Output <- "QUIT :" + quitmsg
	ic.conn.Flush()
	ic.conn.Quit()
}

func (ic *IRCClient) GetConfOpt(option string) string {
	return ic.conf[option]
}

func (ic *IRCClient) SetConfOpt(option string) {
	ic.conf[option] = option
}

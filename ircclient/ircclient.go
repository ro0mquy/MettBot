package ircclient

import (
	"os"
	"log"
)

type IRCClient struct {
	conn *IRCConn
	conf map[string]interface{}
	plugins map[string]Plugin
}

func NewIRCClient(hostport, nick, rname, ident string) *IRCClient {
	c := &IRCClient{nil, make(map[string]interface{}), make(map[string]Plugin)}
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

func (ic *IRCClient) Connect() {
	ic.conn = NewIRCConn()
	hp, ok := ic.conf["hostport"].(string)
	if !ok {
		log.Fatal("Assertion failed")
	}
	ic.conn.Connect(hp)
}

func (ic *IRCClient) Disconnect(quitmsg string) {
	// TODO
}

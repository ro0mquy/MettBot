// Package ircclient provides the main interface for library users
// It manages a single connection to the server and the associated
// configuration and plugins.
package ircclient

import (
	"os"
	"strings"
	"fmt"
)

type IRCClient struct {
	conn       *IRCConn
	conf       map[string]string
	plugins    *pluginStack
	handlers map[string]handler
	disconnect chan bool
}

type handler struct {
	handler Plugin
	minparams int
	minaccess int
}

// Returns a new IRCClient connection with the given configuration options.
// It will not connect to the given server until Connect() has been called,
// so you can register plugins before connecting
func NewIRCClient(hostport, nick, rname, ident string, trigger string) *IRCClient {
	c := &IRCClient{nil, make(map[string]string), newPluginStack(), make(map[string]handler), make(chan bool)}
	c.conf["nick"] = nick
	c.conf["hostport"] = hostport
	c.conf["rname"] = rname
	c.conf["ident"] = ident
	c.conf["trigger"] = trigger
	c.RegisterPlugin(&basicProtocol{})
	return c
}

// Registers a new plugin. Plugins can be registered at any time, even before
// the actual connection attempt. The plugin's Unregister() function will already
// be called when the connection is lost.
func (ic *IRCClient) RegisterPlugin(p Plugin) os.Error {
	if _, ok := ic.plugins.GetPlugin(p.String()); ok == true {
		return os.NewError("Plugin already exists")
	}
	p.Register(ic)
	ic.plugins.Push(p)
	return nil
}

// Registers a command handler. Plugin callbacks will only be called if
// the command matches. Note that only a single plugin per command may
// be registered. This function is not synchronized, e.g., it shall only
// be called during registration (as Plugin.Register()-calls are currently
// sequential).
func (ic *IRCClient) RegisterCommandHandler(command string, minparams int, minaccess int, plugin Plugin) os.Error {
	if plug, err := ic.handlers[command]; err {
		return os.NewError("Handler is already registered by plugin: " + plug.handler.String())
	}
	ic.handlers[command] = handler{plugin, minparams, minaccess}
	return nil
}

// Connects to the server specified on object creation. If the chosen nickname is
// already in use, it will automatically be suffixed with an single underscore until
// an unused nickname is found. This function blocks until the connection attempt
// has been finished.
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
		for p := range ic.plugins.Iter() {
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
	if (s.Command == "PRIVMSG" || s.Command == "NOTICE") && (s.Target == ic.conf["nick"] || strings.Index(s.Args[0], ic.conf["trigger"]) == 0) {
		c = ParseCommand(s)
		// Strip trigger, if necessary
		if c != nil && s.Target != ic.conf["nick"] && len(c.Command) != 0 {
			c.Command = c.Command[len(ic.conf["trigger"]):len(c.Command)]
		}
	}

	// Call line handlers
	for p := range ic.plugins.Iter() {
		go p.ProcessLine(s)
	}

	// Call command handler
	if c == nil {
		return
	}
	if handler, err := ic.handlers[c.Command]; err == true {
		// TODO: Authorization level check
		if len(c.Args) < handler.minparams {
			ic.Reply(c, "This command requires at least " + fmt.Sprintf("%d", handler.minparams) + " parameters")
			return
		}
		go handler.handler.ProcessCommand(c)
	}
}

// Starts the actual command processing. This function will block until the connection
// has either been lost or Disconnect() has been called (by a plugin or by the library
// user).
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

// Disconnects from the server with the given quit message. All plugins wil be unregistered
// and pending messages in queue (e.g. because of floodprotection) will be flushed. This will
// also make InputLoop() return.
func (ic *IRCClient) Disconnect(quitmsg string) {
	ic.shutdown()
	ic.conn.Output <- "QUIT :" + quitmsg
	ic.conn.Quit()
}

// Gets one of the configuration options supplied to the NewIRCClient() method. Valid config
// options usually include:
//  - nick
//  - hostport (colon-seperated host and port to connect to)
//  - rname (the real name)
//  - ident
//  - trigger
func (ic *IRCClient) GetConfOpt(option string) string {
	return ic.conf[option]
}

// Sets a configuration option (see also GetConfOpt())
func (ic *IRCClient) SetConfOpt(option string) {
	ic.conf[option] = option
}

// Dumps a raw line to the server socket. This is usually called by plugins, but may also
// be used by the library user.
func (ic *IRCClient) SendLine(line string) {
	ic.conn.Output <- line
}

func (ic *IRCClient) shutdown() {
	for ic.plugins.Size() != 0 {
		p := ic.plugins.Pop()
		p.Unregister()
	}
}

// Gets the current nickname. Note: This is equivalent to a call to GetConfOpt("nick") and
// might be removed in the future. Better use GetConfOpt() for this purpose
func (ic *IRCClient) GetNick() string {
	return ic.conf["nick"]
}

// Returns a channel on which all plugins will be sent. Use it to iterate over all registered
// plugins.
func (ic *IRCClient) IterPlugins() <-chan Plugin {
	return ic.plugins.Iter()
}

// Get the pointer to a specific plugin that has been registered using RegisterPlugin()
// Name is the name the plugin identifies itself with when String() is called on it.
func (ic *IRCClient) GetPlugin(name string) (Plugin, bool) {
	return ic.plugins.GetPlugin(name)
}

// Sends a reply to a parsed message from a user. This is mostly intended for plugins
// and will automatically distinguish between channel and query messages. Note: Notice
// replies will currently be sent to the client using PRIVMSG, this may change in the
// future.
func (ic *IRCClient) Reply(cmd *IRCCommand, message string) {
	var target string
	if cmd.Target != ic.GetNick() {
		target = cmd.Target
	} else {
		target = strings.SplitN(cmd.Source, "!", 2)[0]
	}
	ic.SendLine("PRIVMSG " + target + " :" + message)
}

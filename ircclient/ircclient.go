// Package ircclient provides the main interface for library users
// It manages a single connection to the server and the associated
// configuration and plugins.
package ircclient

import (
	"os"
	"log"
	"strings"
	"fmt"
)

type IRCClient struct {
	conn       *IRCConn
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
func NewIRCClient(configfile string) *IRCClient {
	c := &IRCClient{nil, newPluginStack(), make(map[string]handler), make(chan bool)}
	c.RegisterPlugin(&basicProtocol{})
	c.RegisterPlugin(NewConfigPlugin(configfile))
	c.RegisterPlugin(new(authPlugin))
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

// Gets one of the configuration options stored in the config object. Valid config
// options for section "Server" usually include:
//  - nick
//  - hostport (colon-seperated host and port to connect to)
//  - realname (the real name)
//  - ident
//  - trigger
func (ic *IRCClient) GetStringOption(section, option string) string {
	c, _ := ic.GetPlugin("conf")
	if c == nil {
		log.Fatal("wtf?")
	}
	cf, _ := c.(*ConfigPlugin)
	cf.Lock()
	retval, _ := cf.Conf.String(section, option)
	cf.Unlock()
	return retval
}
func (ic *IRCClient) SetStringOption(section, option, value string) {
	c, _ := ic.GetPlugin("conf")
	cf, _ := c.(*ConfigPlugin)
	cf.Lock()
	// TODO
	cf.Unlock()
}
func (ic *IRCClient) DelOption(section, option string) {
	// TODO
}
func (ic *IRCClient) GetOptions(section string) []string {
	return nil
}
// TODO: SetIntOption...

func (ic *IRCClient) GetAccessLevel(host string) int {
	// TODO
	return 0
}

func (ic *IRCClient) SetAccessLevel(host string, level int) {
	// TODO
}

func (ic *IRCClient) DelAccessLevel(host string) {
	// TODO
}

// Connects to the server specified on object creation. If the chosen nickname is
// already in use, it will automatically be suffixed with an single underscore until
// an unused nickname is found. This function blocks until the connection attempt
// has been finished.
func (ic *IRCClient) Connect() os.Error {
	ic.conn = NewIRCConn()
	e := ic.conn.Connect(ic.GetStringOption("Server", "host"))
	if e != nil {
		return e
	}
	ic.conn.Output <- "NICK " + ic.GetStringOption("Server", "nick")
	ic.conn.Output <- "USER " + ic.GetStringOption("Server", "ident") + " * Q :" + ic.GetStringOption("Server", "realname")
	nick := ic.GetStringOption("Server", "nick")
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
			ic.SetStringOption("Server", "nick", nick)
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
	if (s.Command == "PRIVMSG" || s.Command == "NOTICE") && (s.Target == ic.GetStringOption("Server", "nick") || strings.Index(s.Args[0], ic.GetStringOption("Server", "trigger")) == 0) {
		c = ParseCommand(s)
		// Strip trigger, if necessary
		if c != nil && s.Target != ic.GetStringOption("Server", "nick") && len(c.Command) != 0 {
			c.Command = c.Command[len(ic.GetStringOption("Server", "trigger")):len(c.Command)]
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
	if cmd.Target != ic.GetStringOption("Server", "nick") {
		target = cmd.Target
	} else {
		target = strings.SplitN(cmd.Source, "!", 2)[0]
	}
	ic.SendLine("PRIVMSG " + target + " :" + message)
}

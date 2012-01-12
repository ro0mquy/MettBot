package plugins

// This plugin manages a common config-file pointer
// and the locks on it.

import (
	"sync"
	"ircclient"
	"github.com/kless/goconfig/config"
	"log"
)

type ConfigPlugin struct {
	ic   *ircclient.IRCClient
	filename string
	Conf *config.Config
	// Operations to the Config structure should be atomic
	lock *sync.Mutex
}

func NewConfigPlugin() *ConfigPlugin {
	c, ok := config.ReadDefault("go-faui2k11.cfg")
	if ok != nil {
		c = config.NewDefault()
		c.AddSection("Server")
		c.AddOption("Server", "host", "dpaulus.dyndns.org:6667")
		c.AddOption("Server", "nick", "testbot")
		c.AddOption("Server", "ident", "ident")
		c.AddOption("Server", "realname", "TestBot Client")
		c.AddOption("Server", "trigger", ".")
		c.AddSection("Auth")
		c.WriteFile("go-faui2k11.cfg", 0644, "go-faui2k11 default config file")
		log.Println("Note: A new default configuration file has been generated in go-faui2k11.cfg. Please edit it to suit your needs and restart go-faui2k11 then")
		return nil
	}
	return &ConfigPlugin{nil, "go-faui2k11.cfg", c, new(sync.Mutex)}
}
func (cp *ConfigPlugin) Register(cl *ircclient.IRCClient) {
	cp.ic = cl
}
func (cp *ConfigPlugin) String() string {
	return "config"
}
func (cp *ConfigPlugin) ProcessLine(msg *ircclient.IRCMessage) {
	// Empty
}
func (cp *ConfigPlugin) Unregister() {
	cp.lock.Lock()
	cp.Conf.WriteFile("go-faui2k11.cfg", 0644, "go-faui2k11 config")
	cp.lock.Unlock()
}
func (cp *ConfigPlugin) Info() string {
	return "run-time configuration manager plugin"
}
func (cp *ConfigPlugin) ProcessCommand(cmd *ircclient.IRCCommand) {
	switch cmd.Command {
	case "version":
		cp.ic.Reply(cmd, "This is go-faui2k11, version 0.01a")
	case "writeconf":
		authplugin, ok := cp.ic.GetPlugin("auth")
		if !ok {
			cp.ic.Reply(cmd, "Sorry, no authentication plugin loaded")
			return
		}
		auth := authplugin.(*AuthPlugin)
		if auth.GetAccessLevel(cmd.Source) < 400 {
			cp.ic.Reply(cmd, "You are not authorized to do that")
			return
		}
		cp.lock.Lock()
		cp.Conf.WriteFile("go-faui2k11.cfg", 0644, "go-faui2k11 config")
		cp.Conf, _ = config.ReadDefault(cp.filename)
		cp.lock.Unlock()
		cp.ic.Reply(cmd, "Successfully flushed cached config entries")
	}
}

func (cp *ConfigPlugin) Lock() {
	cp.lock.Lock()
}
func (cp *ConfigPlugin) Unlock() {
	cp.lock.Unlock()
}

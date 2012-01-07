package plugins

// This plugin manages a common config-file pointer
// and the locks on it.

import (
	"strings"
	"sync"
	"ircclient"
	"github.com/kless/goconfig/config"
)

type ConfigPlugin struct {
	ic       *ircclient.IRCClient
	Conf     *config.Config
	// Operations to the Config structure should be atomic
	lock     *sync.Mutex
}

func NewConfigPlugin(conf *config.Config) *ConfigPlugin {
	return &ConfigPlugin{nil, conf, new(sync.Mutex)}
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
	// Proof of concept implementation
	var target string
	if cmd.Target != cp.ic.GetNick() {
		target = cmd.Target
	} else {
		target = strings.SplitN(cmd.Source, "!", 2)[0]
	}
	switch cmd.Command {
	case "version":
		cp.ic.SendLine("PRIVMSG " + target + " :This is go-faui2k11, version 0.01a")
	}
}

func (cp *ConfigPlugin) Lock() {
	cp.lock.Lock()
}
func (cp *ConfigPlugin) Unlock() {
	cp.lock.Unlock()
}

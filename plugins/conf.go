package plugins

// This plugin manages a common config-file pointer
// and the locks on it.

import (
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
	cp.Conf.WriteFile("go-faui2k11", 0644, "go-faui2k11 config")
	cp.lock.Unlock()
}
func (cp *ConfigPlugin) Info() string {
	return "run-time configuration manager plugin"
}
func (cp *ConfigPlugin) ProcessCommand(cmd *ircclient.IRCCommand) {
	// TODO
}

func (cp *ConfigPlugin) Lock() {
	cp.lock.Lock()
}
func (cp *ConfigPlugin) Unlock() {
	cp.lock.Unlock()
}

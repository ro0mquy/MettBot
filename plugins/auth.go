package plugins

import (
	"strconv"
	"ircclient"
	"fmt"
	"regexp"
)

type AuthPlugin struct {
	ic       *ircclient.IRCClient
	confplugin *ConfigPlugin
}

func NewAuthPlugin() *AuthPlugin {
	return &AuthPlugin{nil, nil}
}
func (a *AuthPlugin) Register(cl *ircclient.IRCClient) {
	a.ic = cl
	plugin, ok := a.ic.GetPlugin("config")
	if !ok {
		panic("AuthPlugin: Register: Unable to get configuration manager plugin")
	}
	a.confplugin, _ = plugin.(*ConfigPlugin)
	if !a.confplugin.Conf.HasSection("Auth") {
		panic("No \"Auth\" section in config file present")
	}
	options, _ := a.confplugin.Conf.Options("Auth")
	for _, mask := range options {
		if _, err := regexp.Compile(mask); err != nil {
			panic(err.String())
		}
	}
}
func (a *AuthPlugin) String() string {
	return "auth"
}
func (a *AuthPlugin) ProcessLine(msg *ircclient.IRCMessage) {
	// Empty
}
func (a *AuthPlugin) Unregister() {
	// Empty
}
func (a *AuthPlugin) Info() string {
	return "Access control manager"
}
func (a *AuthPlugin) ProcessCommand(cmd *ircclient.IRCCommand) {
	switch cmd.Command {
	case "myaccess": fallthrough
	case "mya":
		level := a.GetAccessLevel(cmd.Source)
		slevel := fmt.Sprintf("%d", level)
		a.ic.Reply(cmd, "Your access level is: " + slevel)
	case "addaccess":
		if len(cmd.Args) != 2 {
			a.ic.Reply(cmd, "addaccess takes two arguments: mask and access level")
			return
		}
		level := a.GetAccessLevel(cmd.Source)
		newlevel, err := strconv.Atoi(cmd.Args[1])
		if err != nil {
			a.ic.Reply(cmd, "Error: " + err.String())
		}
		if level < newlevel || level < 400 {
			a.ic.Reply(cmd, "You are not authorized to do this")
		}
		if _, err := regexp.Compile(cmd.Args[0]); err != nil {
			a.ic.Reply(cmd, "Error: Unable to compile regexp: " + err.String())
			return
		}
		a.confplugin.Lock()
		a.confplugin.Conf.AddOption("Auth", cmd.Args[0], fmt.Sprintf("%d", newlevel))
		a.confplugin.Unlock()
		a.ic.Reply(cmd, "Permissions granted")
	case "delaccess":
		if len(cmd.Args) != 1 {
			a.ic.Reply(cmd, "delaccess takes mask to delete as an argument")
			return
		}
		success := a.confplugin.Conf.RemoveOption("Auth", cmd.Args[0])
		if success {
			a.ic.Reply(cmd, "Successfully removed mask")
		} else {
			a.ic.Reply(cmd, "Mask not found")
		}
	}
}
func (a *AuthPlugin) GetAccessLevel(host string) int {
	a.confplugin.Lock()
	options, _ := a.confplugin.Conf.Options("Auth")
	maxaccess := 0
	for _, mask := range options {
		if match, _ := regexp.MatchString(mask, host); match == true {
			newaccess, _ := a.confplugin.Conf.Int("Auth", mask)
			if newaccess > maxaccess {
				maxaccess = newaccess
			}
		}
	}
	defer a.confplugin.Unlock()
	return maxaccess
}

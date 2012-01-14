package ircclient

import (
	"strconv"
	"fmt"
	"regexp"
)

type authPlugin struct {
	ic         *IRCClient
	confplugin *ConfigPlugin
}

func NewauthPlugin() *authPlugin {
	return &authPlugin{nil, nil}
}
func (a *authPlugin) Register(cl *IRCClient) {
	a.ic = cl
	/* TODO: Use new interface in IRCClient
	options, _ := a.confplugin.Conf.Options("Auth")
	for _, mask := range options {
		if _, err := regexp.Compile(mask); err != nil {
			panic(err.String())
		}
	}
	*/
	a.ic.RegisterCommandHandler("mya", 0, 0, a)
}
func (a *authPlugin) String() string {
	return "auth"
}
func (a *authPlugin) ProcessLine(msg *IRCMessage) {
	// Empty
}
func (a *authPlugin) Unregister() {
	// Empty
}
func (a *authPlugin) Info() string {
	return "Access control manager"
}
func (a *authPlugin) ProcessCommand(cmd *IRCCommand) {
	switch cmd.Command {
	case "myaccess":
		fallthrough
	case "mya":
		level := a.GetAccessLevel(cmd.Source)
		slevel := fmt.Sprintf("%d", level)
		a.ic.Reply(cmd, "Your access level is: "+slevel)
	case "addaccess":
		if len(cmd.Args) != 2 {
			a.ic.Reply(cmd, "addaccess takes two arguments: mask and access level")
			return
		}
		level := a.GetAccessLevel(cmd.Source)
		newlevel, err := strconv.Atoi(cmd.Args[1])
		if err != nil {
			a.ic.Reply(cmd, "Error: "+err.String())
		}
		if level < newlevel || level < 400 {
			a.ic.Reply(cmd, "You are not authorized to do this")
			return
		}
		if _, err := regexp.Compile(cmd.Args[0]); err != nil {
			a.ic.Reply(cmd, "Error: Unable to compile regexp: "+err.String())
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
		level := a.GetAccessLevel(cmd.Source)
		a.confplugin.Lock()
		defer a.confplugin.Unlock()
		dlevel, success := a.confplugin.Conf.Int("Auth", cmd.Args[0])
		if success == nil {
			if dlevel >= level || level != 500 {
				a.ic.Reply(cmd, "Can't remove mask: Has higher privileges than you")
				return
			}
			a.confplugin.Conf.RemoveOption("Auth", cmd.Args[0])
			a.ic.Reply(cmd, "Successfully removed mask")
		} else {
			a.ic.Reply(cmd, "Mask not found")
		}
	}
}
func (a *authPlugin) GetAccessLevel(host string) int {
/*
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
*/
	return 500
}

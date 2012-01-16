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
	options := a.ic.GetOptions("Auth")
	for _, mask := range options {
		if _, err := regexp.Compile(mask); err != nil {
			panic(err.String())
		}
	}
	a.ic.RegisterCommandHandler("mya", 0, 0, a)
	a.ic.RegisterCommandHandler("myaccess", 0, 0, a)
	a.ic.RegisterCommandHandler("addaccess", 0, 400, a)
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
		if level == 500 {
			a.ic.Reply(cmd, "Your access level is over 9000")
		} else {
			a.ic.Reply(cmd, "Your access level is: "+slevel)
		}
	case "addaccess":
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
		//a.ic.SetIntOption("Auth", cmd.Args[0], newlevel)
		a.SetAccessLevel(cmd.Args[0], newlevel)
		a.ic.Reply(cmd, "Permissions granted")
	case "delaccess":
		if len(cmd.Args) != 1 {
			a.ic.Reply(cmd, "delaccess takes mask to delete as an argument")
			return
		}
		level := a.GetAccessLevel(cmd.Source)
		dlevel, ok := a.ic.GetIntOption("Auth", cmd.Args[0])
		if ok != nil {
			if dlevel >= level || level != 500 {
				a.ic.Reply(cmd, "Can't remove mask: Has higher privileges than you")
				return
			}
			//a.ic.RemoveOption("Auth", cmd.Args[0])
			a.DelAccessLevel(cmd.Args[0])
			a.ic.Reply(cmd, "Successfully removed mask")
		} else {
			a.ic.Reply(cmd, "Mask not found")
		}
	}
}

func (a *authPlugin) SetAccessLevel(host string, level int) {
	a.ic.SetIntOption("Auth", host, level)
}

func (a *authPlugin) DelAccessLevel(mask string) {
	a.ic.RemoveOption("Auth", mask)
}

func (a *authPlugin) GetAccessLevel(host string) int {
	options := a.ic.GetOptions("Auth")
	maxaccess := 0
	for _, mask := range options {
		if match, _ := regexp.MatchString(mask, host); match == true {
			newaccess, _ := a.ic.GetIntOption("Auth", mask)
			if newaccess > maxaccess {
				maxaccess = newaccess
			}
		}
	}
	return maxaccess
}

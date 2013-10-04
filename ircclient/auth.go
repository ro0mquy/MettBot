package ircclient

import (
	"fmt"
	"regexp"
	"strconv"
)

type authPlugin struct {
	ic         *IRCClient
}

func (a *authPlugin) Register(cl *IRCClient) {
	a.ic = cl
	options := a.ic.GetOptions("Auth")
	for _, mask := range options {
		if _, err := regexp.Compile(mask); err != nil {
			panic(err)
		}
	}
	a.ic.RegisterCommandHandler("mya", 0, 0, a)
	a.ic.RegisterCommandHandler("myaccess", 0, 0, a)
	a.ic.RegisterCommandHandler("addaccess", 2, 400, a)
	a.ic.RegisterCommandHandler("delaccess", 1, 400, a)
}

func (a *authPlugin) String() string {
	return "auth"
}

func (a *authPlugin) Usage(cmd string) string {
	switch cmd {
	case "myaccess", "mya":
		return cmd + ": tells you what access-level (i.e. permissions) you have"
	case "addaccess":
		return "addaccess <hostmask> <level>: adds access-level <level> for hostmask <hostmask>"
	case "delaccess":
		return "delaccess <hostmask>: removes access-level for hostmask <hostmask>"
	}
	// shouldn't be a problem, this usage isn't called unless we're registered for it
	return ""
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
	case "myaccess", "mya":
		level := a.GetAccessLevel(cmd.Source)
		slevel := fmt.Sprintf("%d", level)
		if level == 500 {
			a.ic.Reply(cmd, "Your access level is over 9000")
		} else {
			a.ic.Reply(cmd, "Your access level is: "+slevel)
		}

	case "addaccess":
		if _, err := regexp.Compile(cmd.Args[0]); err != nil {
			a.ic.Reply(cmd, "Error: Unable to compile regexp: "+err.Error())
			return
		}

		userLevel := a.GetAccessLevel(cmd.Source)
		targetLevel, _ := a.ic.GetIntOption("Auth", cmd.Args[0])
		newLevel, err := strconv.Atoi(cmd.Args[1])
		if err != nil {
			a.ic.Reply(cmd, "Error: "+err.Error())
		}

		if userLevel < newLevel || userLevel <= targetLevel {
			a.ic.Reply(cmd, "You are not authorized to do this")
			return
		}
		a.SetAccessLevel(cmd.Args[0], newLevel)
		a.ic.Reply(cmd, "Permissions granted")

	case "delaccess":
		level := a.GetAccessLevel(cmd.Source)
		dlevel, err := a.ic.GetIntOption("Auth", cmd.Args[0])
		if err != nil {
			a.ic.Reply(cmd, "Mask not found")
			return
		}

		if dlevel >= level {
			a.ic.Reply(cmd, "Can't remove mask: Has higher privileges than you")
			return
		}
		a.DelAccessLevel(cmd.Args[0])
		a.ic.Reply(cmd, "Successfully removed mask")
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

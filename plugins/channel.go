package plugins

import (
	"ircclient"
)

type ChannelsPlugin struct {
	ic *ircclient.IRCClient
}

func (q *ChannelsPlugin) Register(cl *ircclient.IRCClient) {
	q.ic = cl
}

func (q *ChannelsPlugin) String() string {
	return "channel"
}

func (q *ChannelsPlugin) Info() string {
	return "Manages channel auto-join and possibly options"
}

func (q *ChannelsPlugin) ProcessLine(msg *ircclient.IRCMessage) {
	if msg.Command != "001" {
		return
	}
	/* When registering, join channels */
	options := q.ic.GetOptions("Channels")
	for _, key := range options {
		q.ic.SendLine("JOIN #" + key)
	}
}

func (q *ChannelsPlugin) ProcessCommand(cmd *ircclient.IRCCommand) {
	// TODO: Delchannel
	if cmd.Command != "addchannel" && cmd.Command != "join" && cmd.Command != "part" {
		return
	}
	if q.ic.GetAccessLevel(cmd.Source) < 200 {
		q.ic.Reply(cmd, "You are not authorized to do that")
		return
	}
	if len(cmd.Args) < 1 {
		q.ic.Reply(cmd, "Too few parameters. Please specify a channel name.")
		return
	}
	if cmd.Command == "join" {
		q.ic.SendLine("JOIN #" + cmd.Args[0])
		return
	}
	if cmd.Command == "part" {
		q.ic.SendLine("PART #" + cmd.Args[0])
		return
	}
	// TODO: Quick'n'dirty. Check whether channel already exists and strip #, if
	// existent.
	q.ic.SetStringOption("Channels", cmd.Args[0], "42")
	q.ic.SendLine("JOIN #" + cmd.Args[0])
}

func (q *ChannelsPlugin) Unregister() {
	return
}

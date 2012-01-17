package plugins

import (
	"ircclient"
)

type ChannelsPlugin struct {
	ic *ircclient.IRCClient
}

func (q *ChannelsPlugin) Register(cl *ircclient.IRCClient) {
	q.ic = cl
	cl.RegisterCommandHandler("join", 1, 200, q)
	cl.RegisterCommandHandler("part", 1, 200, q)
	cl.RegisterCommandHandler("addchannel", 1, 400, q)
}

func (q *ChannelsPlugin) String() string {
	return "channel"
}

func (q *ChannelsPlugin) Info() string {
	return "Manages channel auto-join and possibly options"
}

func (cp *ChannelsPlugin) Usage(cmd string) string {
	switch cmd {
	case "join":
		return "join <channel_without_#>, makes the bot join #<channel>"
	case "part":
		return "part <channel_without_#>, parts the bot from #<channel>"
	case "addchannel":
		return "addchannel <channel_without_#>, adds #<channel> to the bot's autojoin list"
	}
	return ""
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
	switch cmd.Command {
	case "join":
		q.ic.SendLine("JOIN #" + cmd.Args[0])
	case "part":
		q.ic.SendLine("PART #" + cmd.Args[0])
	case "addchannel":
		// TODO: Quick'n'dirty. Check whether channel already exists and strip #, if
		// existent.
		q.ic.SetStringOption("Channels", cmd.Args[0], "42")
		q.ic.SendLine("JOIN #" + cmd.Args[0])
	}
}

func (q *ChannelsPlugin) Unregister() {
	return
}

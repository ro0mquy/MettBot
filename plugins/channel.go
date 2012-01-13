package plugins

import (
	"ircclient"
)

type ChannelsPlugin struct {
	ic         *ircclient.IRCClient
	authplugin *AuthPlugin
	confplugin *ConfigPlugin
}

func (q *ChannelsPlugin) Register(cl *ircclient.IRCClient) {
	q.ic = cl
	plugin, ok := q.ic.GetPlugin("config")
	authplugin, ok2  := q.ic.GetPlugin("auth")
	if !ok || !ok2 {
		panic("ChannelsPlugin: Register: Unable to get necessary plugins")
	}
	q.confplugin = plugin.(*ConfigPlugin)
	q.authplugin = authplugin.(*AuthPlugin)
	if !q.confplugin.Conf.HasSection("Channels") {
		q.confplugin.Conf.AddSection("Channels")
	}
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
	q.confplugin.Lock()
	defer q.confplugin.Unlock()
	options, _ := q.confplugin.Conf.Options("Channels")
	for _, key := range options {
		q.ic.SendLine("JOIN #" + key)
	}
}

func (q *ChannelsPlugin) ProcessCommand(cmd *ircclient.IRCCommand) {
	// TODO: Delchannel
	if cmd.Command != "addchannel" {
		return
	}
	if q.authplugin.GetAccessLevel(cmd.Source) < 200 {
		q.ic.Reply(cmd, "You are not authorized to do that")
		return
	}
	if len(cmd.Args) < 1 {
		q.ic.Reply(cmd, "Too few parameters. Please specify a channel name.")
		return
	}
	q.confplugin.Lock()
	defer q.confplugin.Unlock()
	// TODO: Quick'n'dirty. Check whether channel already exists and strip #, if
	// existent.
	q.confplugin.Conf.AddOption("Channels", cmd.Args[0], "")
	q.ic.SendLine("JOIN #" + cmd.Args[0])
}

func (q *ChannelsPlugin) Unregister() {
	return
}

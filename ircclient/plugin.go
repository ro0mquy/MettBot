package ircclient

type Plugin interface {
	Register(cl *IRCClient)
	String() string
	Info()
	ProcessLine(msg *IRCMessage)
	ProcessCommand(cmd *IRCCommand)
	Unregister()
}

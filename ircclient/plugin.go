package ircclient

type Plugin interface {
	Register(cl *IRCClient)
	String() string
	Info() string
	ProcessLine(msg *IRCMessage)
	ProcessCommand(cmd *IRCCommand)
	Unregister()
}

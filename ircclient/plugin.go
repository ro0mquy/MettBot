package ircclient

type Plugin interface {
	Register(cl *IRCClient)
	String() string
	//ProcessLine(msg IRCMessage)
	//Unregister()
}

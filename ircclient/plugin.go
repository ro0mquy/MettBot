package ircclient

// The interface to be implemented by all plugins
type Plugin interface {
	// This function is called by IRCClient when registering the plugin.
	// It takes a pointer to the parent IRCClient object that is needed for
	// almost any kind of functionality, so you should definitively save it ;-)
	Register(cl *IRCClient)
	// Returns the name of the plugin, as a short string (e.g.: auth, lecture, ...)
	String() string
	// Returns the description of the plugin in human-readable form, about one line long.
	Info() string
	// Returns usage for the Command cmd
	Usage(cmd string) string
	// This method is called by the parent IRCClient when a new line from server
	// arrives, regardless of the other state of the connection. This means, if
	// the plugin is registered soon enough, this handler method is also called
	// during registration phase, when authentication hasn't been performed.
	ProcessLine(msg *IRCMessage)
	// This method is called when an command directed to the bot has been received
	// and parsed. It is NOT called during registration phase, specifically, the
	// initial NOTICEs are _NOT_ passed to this method. Replies can be easily sent
	// using the Reply() function of the parent IRCClient.
	ProcessCommand(cmd *IRCCommand)
	// Automatically called when the connection is lost. Should perform cleanup
	// work and not expect that the plugin is used again.
	Unregister()
}

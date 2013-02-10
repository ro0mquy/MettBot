package main

import (
	"./mettbot"
	"fmt"
	irc "github.com/fluffle/goirc/client"
)

func main() {
	// create new IRC connection
	mett := mettbot.NewMettbot(*mettbot.Nick, *mettbot.Nick, *mettbot.Longnick)

	// Set up handlers
	mett.AddHandler("connected", func(conn *irc.Conn, line *irc.Line) { mett.HandlerConnected() })
	mett.AddHandler("disconnected", func(conn *irc.Conn, line *irc.Line) { mett.HandlerDisconnected() })
	mett.AddHandler("join", func(conn *irc.Conn, line *irc.Line) { mett.HandlerJoin(line) })
	mett.AddHandler("privmsg", func(conn *irc.Conn, line *irc.Line) { mett.HandlerPrivmsg(line) })
	mett.AddHandler("topic", func(conn *irc.Conn, line *irc.Line) { mett.HandlerTopic(line) })

	// set up a goroutine to read commands from stdin
	go mett.ReadStdin()

	// set up a goroutine to do parsey things with the stuff from stdin
	go mett.ParseStdin()

	// Set up a go routine to write quotes to file
	go mett.WriteQuote(*mettbot.Quotes, mett.QuotesPrnt, mett.QuotesLinesPrnt)
	go mett.WriteQuote(*mettbot.Metts, mett.MettsPrnt, mett.MettsLinesPrnt)

	// Go routine to post regulary mett content
	go mett.CheckMett()

	for !mett.ReallyQuit {
		// connect to server
		if err := mett.Connect(*mettbot.Host); err != nil {
			fmt.Printf("Connection error: %s\n", err)
			return
		}

		// wait on quit channel
		<-mett.Quitted
	}
}

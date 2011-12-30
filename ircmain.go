package main

import (
	"ircclient"
)

var server_lines= []string{
	":fu-berlin.de 020 * :Please wait while we process your connection.",
	":fu-berlin.de 001 osntauohe :Welcome to the Internet Relay Network osntauohe!~osntauohe@176.99.114.122",
	":fu-berlin.de 002 osntauohe :Your host is fu-berlin.de, running version 2.11.2p2",
	":fu-berlin.de 003 osntauohe :This server was created Wed Dec 8 2010 at 17:45:14 CET",
	":fu-berlin.de 004 osntauohe fu-berlin.de 2.11.2p2 aoOirw abeiIklmnoOpqrRstv",
	":fu-berlin.de 005 osntauohe RFC2812 PREFIX=(ov)@+ CHANTYPES=#&!+ MODES=3 CHANLIMIT=#&!+:21 NICKLEN=15 TOPICLEN=255 KICKLEN=255 MAXLIST=beIR:64 CHANNELLEN=50 IDCHAN=!:5 CHANMODES=beIR,k,l,imnpstaqr :are supported by this server",
	":fu-berlin.de 005 osntauohe PENALTY FNC EXCEPTS=e INVEX=I CASEMAPPING=ascii NETWORK=IRCnet :are supported by this server",
	":fu-berlin.de 042 osntauohe 276BAY2UY :your unique ID",
	":fu-berlin.de 251 osntauohe :There are 61364 users and 7 services on 30 servers",
	":fu-berlin.de 252 osntauohe 109 :operators online",
	":fu-berlin.de 254 osntauohe 34263 :channels formed",
	":fu-berlin.de 255 osntauohe :I have 1547 users, 1 services and 1 servers",
	":fu-berlin.de 265 osntauohe 1547 1864 :Current local users 1547, max 1864",
	":fu-berlin.de 266 osntauohe 61364 77287 :Current global users 61364, max 77287",
	":fu-berlin.de 375 osntauohe :- fu-berlin.de Message of the Day - ",
	":fu-berlin.de 372 osntauohe :- 8/12/2010 17:33",
	":fu-berlin.de 372 osntauohe :- ",
	":fu-berlin.de 372 osntauohe :- Willkommen auf dem IRCnet-Server der Freien Universitaet Berlin, ZEDAT",
	":fu-berlin.de 372 osntauohe :- ",
	":fu-berlin.de 372 osntauohe :-    Verbindliche Benutzungsregeln und weitere Informationen gibt es ",
	":fu-berlin.de 372 osntauohe :-    unter http://irc.fu-berlin.de",
	":fu-berlin.de 372 osntauohe :- ",
	":fu-berlin.de 372 osntauohe :-                                                                       ",
	":fu-berlin.de 372 osntauohe :- ",
	":fu-berlin.de 372 osntauohe :- Viel Spass wuenschen",
	":fu-berlin.de 372 osntauohe :- ",
	":fu-berlin.de 372 osntauohe :- Oliver 'ob' Brandmueller, Timo 'fuechsle' Fuchs, Tanja 'tawi' Wittke",
	":fu-berlin.de 372 osntauohe :- ",
	":fu-berlin.de 376 osntauohe :End of MOTD command.",
}

func main() {
	for _, line := range server_lines {
		ircclient.ParseServerLine(line)
	}
	s := ircclient.NewIRCConn()
	s.Connect("localhost:6667")
	s.Output <- "Hello, world\n"
	s.Output <- "Asdf!\n"
	s.Quit()
	//	go output_channel(serverData)
	//	for {
	//		fmt.Scanln(st)
	//		clientInput <- st
	//	}
}

func output_channel(c <-chan string) {
	//	for {
	//		s := <- c
	//		fmt.Println("-> " + s)
	//	}
}

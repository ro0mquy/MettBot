package main

import (
	"bufio"
	"flag"
	"fmt"
	irc "github.com/fluffle/goirc/client"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var host *string = flag.String("host", "irc.ps0ke.de:2342", "IRC server")
var channel *string = flag.String("channel", "#metttest", "IRC channel")
var nick *string = flag.String("nick", "rohmett", "IRC nick")
var longnick *string = flag.String("longnick", "Le MettBot", "IRC fullname")
var timeformat *string = flag.String("timeformat", "2006-01-02T15:04", "Time format string (standard date: 2006-01-02T15:04:05")
var quotes *string = flag.String("quotes", "mett_quotes.txt", "Quote database file")

func init() {
	flag.Parse()
}

type Mettbot struct {
	*irc.Conn
	Quitted    chan bool
	Prnt       chan string
	Input      chan string
	ReallyQuit bool
}

func NewMettbot(nick string, args ...string) *Mettbot {
	bot := &Mettbot{
		irc.SimpleClient(nick, args...),
		make(chan bool),
		make(chan string),
		make(chan string, 4),
		false,
	}
	bot.EnableStateTracking()
	return bot
}

func (bot *Mettbot) hConnected()    { bot.Join(*channel) }
func (bot *Mettbot) hDisconnected() { bot.Quitted <- true }

func (bot *Mettbot) hPrivmsg(line *irc.Line) {
	actChannel := line.Args[0]
	msg := line.Args[1]

	if msg[0] == '!' {
		cmd := msg
		args := ""
		idx := strings.Index(msg, " ")
		if idx != -1 {
			cmd = msg[:idx]
			args = msg[idx+1:]
		}

		switch {
		case cmd == "!help":
			bot.Help(actChannel)
		case args == "":
			bot.Syntax(actChannel)
		case cmd == "!quote":
			bot.cQuote(actChannel, args, line.Time)
		case cmd == "!print":
			bot.cPrint(actChannel, args)
		default:
			bot.Syntax(actChannel)
		}
	}
}

func (bot *Mettbot) Syntax(channel string) {
	bot.Notice(channel, "Wrong Syntax. Try !help")
}

func (bot *Mettbot) Help(channel string) {
	bot.Notice(channel, "Mett")
}

func (bot *Mettbot) cQuote(channel string, msg string, t time.Time) {
	s := fmt.Sprintln(t.Format(*timeformat), msg)
	fmt.Print(s)
	bot.Prnt <- s
	bot.Notice(channel, "Added Quote to Database")
}

func (bot *Mettbot) cPrint(channel string, msg string) {
	num, err := strconv.Atoi(msg)
	if err != nil {
		bot.Syntax(channel)
		return
	}

	fi, err := os.Open(*quotes)
	if err != nil {
		log.Println(err)
		bot.Notice(channel, "Failed to open quote database")
		return
	}
	defer fi.Close()

	reader := bufio.NewReader(fi)
	quote := ""
	for ; num >= 0; num-- {
		quote, err = reader.ReadString('\n')
		if err == io.EOF {
			bot.Notice(channel, "Quote not found")
			return
		}
		if err != nil {
			log.Println(err)
			bot.Notice(channel, "Failed to read from quote database")
			return
		}
	}
	bot.Notice(channel, quote)
}

func (bot *Mettbot) readStdin() {
	con := bufio.NewReader(os.Stdin)
	for {
		s, err := con.ReadString('\n')
		if err != nil {
			// wha?, maybe ctrl-D...
			close(bot.Input)
			break
		}
		// no point in sending empty lines down the channel
		if len(s) > 2 {
			bot.Input <- s[0 : len(s)-1]
		}
	}
}

func (bot *Mettbot) parseStdin() {
	for cmd := range bot.Input {
		if cmd[0] == ':' {
			switch idx := strings.Index(cmd, " "); {
			case cmd[1] == 'd':
				fmt.Printf(bot.String())
			case cmd[1] == 'f':
				if len(cmd) > 2 && cmd[2] == 'e' {
					// enable flooding
					bot.Flood = true
				} else if len(cmd) > 2 && cmd[2] == 'd' {
					// disable flooding
					bot.Flood = false
				}
				for i := 0; i < 20; i++ {
					bot.Privmsg(*channel, "salami!1!")
				}
			case idx == -1:
				continue
			case cmd[1] == 'q':
				bot.ReallyQuit = true
				bot.Quit(cmd[idx+1 : len(cmd)])
			case cmd[1] == 'j':
				bot.Join(cmd[idx+1 : len(cmd)])
			case cmd[1] == 'p':
				bot.Part(cmd[idx+1 : len(cmd)])
			case cmd[1] == 'm':
				bot.Privmsg(*channel, cmd[idx+1:len(cmd)])
			case cmd[1] == 'a':
				bot.Action(*channel, cmd[idx+1:len(cmd)])
			}
		} else {
			bot.Raw(cmd)
		}
	}
}

func (bot *Mettbot) writeQuote() {
	for messages := range bot.Prnt {
		fo, err := os.OpenFile(*quotes, syscall.O_WRONLY+syscall.O_CREAT+syscall.O_APPEND, 0644)
		if err != nil {
			log.Println(err)
			bot.Notice(*channel, "Couldn't open quote database")
			continue
		}
		defer fo.Close()

		_, err = fo.WriteString(messages)
		if err != nil {
			log.Println(err)
			bot.Notice(*channel, "Couldn't write to quote database")
		}
		fo.Close()
	}
}

func main() {
	// create new IRC connection
	mett := NewMettbot(*nick, *nick, *longnick)

	// Set up handlers
	mett.AddHandler("connected", func(conn *irc.Conn, line *irc.Line) { mett.hConnected() })
	mett.AddHandler("disconnected", func(conn *irc.Conn, line *irc.Line) { mett.hDisconnected() })
	mett.AddHandler("privmsg", func(conn *irc.Conn, line *irc.Line) { mett.hPrivmsg(line) })

	// set up a goroutine to read commands from stdin
	go mett.readStdin()

	// set up a goroutine to do parsey things with the stuff from stdin
	go mett.parseStdin()

	// Set up a go routine to write quotes to file
	go mett.writeQuote()

	for !mett.ReallyQuit {
		// connect to server
		if err := mett.Connect(*host); err != nil {
			fmt.Printf("Connection error: %s\n", err)
			return
		}

		// wait on quit channel
		<-mett.Quitted
	}
}

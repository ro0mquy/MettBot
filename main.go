package main

import (
	"bufio"
	"flag"
	"fmt"
	irc "github.com/fluffle/goirc/client"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
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
	LinesPrnt  chan int
	Input      chan string
	ReallyQuit bool
	Topics     map[string]string
}

func NewMettbot(nick string, args ...string) *Mettbot {
	bot := &Mettbot{
		irc.SimpleClient(nick, args...), // *irc.Conn
		make(chan bool),                 // Quitted
		make(chan string),               // Prnt
		make(chan int),                  // LinesPrnt
		make(chan string, 4),            // Input
		false,                           // ReallyQuit
		make(map[string]string), // Topics
	}
	bot.EnableStateTracking()
	return bot
}

func (bot *Mettbot) hConnected()    { bot.Join(*channel) }
func (bot *Mettbot) hDisconnected() { bot.Quitted <- true }

func (bot *Mettbot) hJoin(line *irc.Line) {
	time.Sleep(1000 * time.Millisecond)
	actChannel := line.Args[0]
	bot.Topics[actChannel] = bot.ST.GetChannel(actChannel).Topic
}

func (bot *Mettbot) hTopic(line *irc.Line) {
	actChannel := line.Args[0]
	newTopic := line.Args[1]
	oldTopic := bot.Topics[actChannel]
	bot.Topics[actChannel] = newTopic
	//bot.Notice(actChannel, "Old topic: " + oldTopic)
	//bot.Notice(actChannel, "New topic: "+newTopic)
	bot.Notice(actChannel, bot.diffTopic(oldTopic, newTopic))
}

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
			bot.Help(actChannel, args, line.Nick)
		case cmd == "!colors":
			for i := 0; i < 16; i++ {
				bot.Notice(*channel, fmt.Sprintf("\x03%v %v", i, i))
			}
		case args == "":
			bot.Syntax(actChannel)
		case cmd == "!quote":
			bot.cQuote(actChannel, args, line.Time)
		case cmd == "!print":
			bot.cPrint(actChannel, args, line.Nick)
		default:
			bot.Syntax(actChannel)
		}
	}
}

func (bot *Mettbot) Syntax(channel string) {
	bot.Notice(channel, "Wrong Syntax. Try !help")
}

func (bot *Mettbot) Help(channel, args, nick string) {
	//bot.Notice(channel, "Mett")
	if args == "seriöslich" {
		bot.Privmsg(nick, "MettBot")
		bot.Privmsg(nick, "")
		bot.Privmsg(nick, "!quote <$nick> $quote -- add $quote to quote database")
		bot.Privmsg(nick, "!print $integer       -- print out quote number $integer")
		bot.Privmsg(nick, "!help seriöslich      -- show this help")
	} else {
		bot.Syntax(channel)
	}
}

func (bot *Mettbot) cQuote(channel string, msg string, t time.Time) {
	s := fmt.Sprintln(t.Format(*timeformat), msg)
	fmt.Print(s)
	bot.Prnt <- s
	bot.Notice(channel, fmt.Sprint("Added Quote #", <-bot.LinesPrnt, " to Database"))
}

func (bot *Mettbot) cPrint(channel, msg, nick string) {
	num, err := strconv.Atoi(msg)
	if err != nil {
		bot.Syntax(channel)
		return
	}
	if num < 0 {
		bot.Action(channel, "slaps "+nick)
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
		fo, err := os.OpenFile(*quotes, syscall.O_RDWR+syscall.O_CREAT, 0644)
		if err != nil {
			log.Println(err)
			bot.Notice(*channel, "Couldn't open quote database")
			continue
		}
		defer fo.Close()

		foReader := bufio.NewReader(fo)
		lines := 0 //last quote has no newline

		for {
			_, err = foReader.ReadString('\n')
			if err != nil {
				break
			}
			lines++
		}
		bot.LinesPrnt <- lines

		_, err = fo.WriteString(messages)
		if err != nil {
			log.Println(err)
			bot.Notice(*channel, "Couldn't write to quote database")
		}
		fo.Close()
	}
}

func (bot *Mettbot) diffTopic(oldTopic, newTopic string) string {
	oldFile, err := ioutil.TempFile("", ".mettbotWdiffOld")
	if err != nil {
		log.Println(err)
		return ""
	}

	n, err := oldFile.WriteString(oldTopic)
	if n != len(oldTopic) || err != nil {
		log.Println(err)
		return ""
	}
	oldFile.Close()

	newFile, err := ioutil.TempFile("", ".mettbotWdiffNew")
	if err != nil {
		log.Println(err)
		return ""
	}

	n, err = newFile.WriteString(newTopic)
	if n != len(newTopic) || err != nil {
		log.Println(err)
		return ""
	}

	newFile.Close()
	defer func() {
		os.Remove(oldFile.Name())
		os.Remove(newFile.Name())
	}()

	db := "❣" // DeletionBegin
	de := "❢" // DeletionEnd
	ib := "¶" // InsertionBegin
	ie := "⁋" // InsertionEnd

	for {
		rdb := rune(rand.Intn(255-32-3) + 32)
		rde := rdb + 1
		rib := rde + 1
		rie := rib + 1

		contains := strings.ContainsRune(oldTopic, rdb)
		contains = contains || strings.ContainsRune(oldTopic, rde)
		contains = contains || strings.ContainsRune(oldTopic, rib)
		contains = contains || strings.ContainsRune(oldTopic, rie)

		contains = contains || strings.ContainsRune(newTopic, rdb)
		contains = contains || strings.ContainsRune(newTopic, rde)
		contains = contains || strings.ContainsRune(newTopic, rib)
		contains = contains || strings.ContainsRune(newTopic, rie)

		if contains == false {
			db = fmt.Sprintf("%c", rdb)
			de = fmt.Sprintf("%c", rde)
			ib = fmt.Sprintf("%c", rib)
			ie = fmt.Sprintf("%c", rie)
			break
		}
	}

	coloring := map[string]string{ // http://oreilly.com/pub/h/1953
		db: "\x035\x1F\x02",
		de: "\x0F\x0315",
		ib: "\x033\x02",
		ie: "\x0F\x0315",
	}

	cmd := exec.Command("wdiff", "-w"+db, "-x"+de, "-y"+ib, "-z"+ie, oldFile.Name(), newFile.Name())
	out, _ := cmd.Output()
	outStr := string(out)

	for n, v := range coloring {
		outStr = strings.Replace(outStr, n, v, -1)
		fmt.Printf("Replace %#v with %#v\n", n, v)
	}

	return outStr
}

func main() {
	// create new IRC connection
	mett := NewMettbot(*nick, *nick, *longnick)

	// Set up handlers
	mett.AddHandler("connected", func(conn *irc.Conn, line *irc.Line) { mett.hConnected() })
	mett.AddHandler("disconnected", func(conn *irc.Conn, line *irc.Line) { mett.hDisconnected() })
	mett.AddHandler("join", func(conn *irc.Conn, line *irc.Line) { mett.hJoin(line) })
	mett.AddHandler("privmsg", func(conn *irc.Conn, line *irc.Line) { mett.hPrivmsg(line) })
	mett.AddHandler("topic", func(conn *irc.Conn, line *irc.Line) { mett.hTopic(line) })

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

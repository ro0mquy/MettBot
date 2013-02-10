package mettbot

import (
	a "./answers"
	"bufio"
	"fmt"
	irc "github.com/fluffle/goirc/client"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func (bot *Mettbot) Command(actChannel, msg string, line *irc.Line) {
	cmd := msg
	args := ""
	idx := strings.Index(msg, " ")
	if idx != -1 {
		cmd = msg[:idx]
		args = msg[idx+1:]
	}

	switch {
	case cmd == "!help":
		bot.cHelp(actChannel, args, line.Nick)
	case cmd == "!mett":
		bot.Mett()
		if idx == -1 {
			bot.PostMett(*Channel)
		} else {
			bot.cMett(actChannel, args)
		}
	case cmd == "!colors":
		for i := 0; i < 16; i++ {
			bot.Notice(*Channel, fmt.Sprintf("\x03%v %v", i, i))
		}
	case args == "":
		bot.Syntax(actChannel)
	case cmd == "!add":
		bot.cAdd(actChannel, args, line.Time)
	case cmd == "!quote":
		bot.cQuote(actChannel, args, line.Nick)
	case cmd == "!search":
		bot.cSearch(actChannel, args)
	default:
		bot.Syntax(actChannel)
	}
}

func (bot *Mettbot) cHelp(channel, args, nick string) {
	if args == "seriöslich" {
		bot.Privmsg(nick, "MettBot")
		bot.Privmsg(nick, "")
		bot.Privmsg(nick, "!add <$nick> $quote   -- add a new quote to the database, timestamp is added automagically")
		bot.Privmsg(nick, "!quote $integer       -- print a quote from the database")
		bot.Privmsg(nick, "!search $string       -- searches the quote database (put '/' around your string for regex)")
		bot.Privmsg(nick, "!mett                 -- post a random entry from the mett database")
		bot.Privmsg(nick, "!mett $mettcontent    -- add new mettcontent to the mett database")
		bot.Privmsg(nick, "!help seriöslich      -- show this help text")
	} else {
		bot.Syntax(channel)
	}
}

func (bot *Mettbot) cAdd(channel string, msg string, t time.Time) {
	s := fmt.Sprintln(t.Format(*Timeformat), msg)
	log.Print("Quote: " + s)
	bot.QuotesPrnt <- s
	bot.Notice(channel, fmt.Sprintf(a.RandStr(a.AddedQuote), <-bot.QuotesLinesPrnt))
}

func (bot *Mettbot) cMett(channel string, s string) {
	log.Print("Mett: " + s)
	bot.MettsPrnt <- s + "\n"
	bot.Notice(channel, fmt.Sprintf(a.RandStr(a.AddedMett), <-bot.MettsLinesPrnt))
}

func (bot *Mettbot) cQuote(channel, msg, nick string) {
	num, err := strconv.Atoi(msg)
	if err != nil {
		bot.Syntax(channel)
		return
	}
	if num < 0 {
		bot.Action(channel, fmt.Sprintf(a.RandStr(a.OffendNick), nick))
		return
	}

	fi, err := os.Open(*Quotes)
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
			bot.Notice(channel, a.RandStr(a.QuoteNotFound))
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

func (bot *Mettbot) cSearch(channel, msg string) {
	fi, err := os.Open(*Quotes)
	if err != nil {
		log.Println(err)
		bot.Notice(channel, "Failed to open quote database")
		return
	}
	defer fi.Close()

	reader := bufio.NewReader(fi)
	for n := 0; ; n++ {
		quote, err := reader.ReadString('\n')
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Println(err)
			bot.Notice(channel, "Failed to read from quote database")
			return
		}

		ok := false
		if msg[0] == '/' && msg[len(msg)-1] == '/' {
			ok, err = regexp.MatchString(strings.ToLower(msg[1:len(msg)-1]), strings.ToLower(quote))
		} else {
			ok = strings.Contains(strings.ToLower(quote), strings.ToLower(msg))
		}

		if ok && err == nil {
			bot.Notice(channel, strconv.Itoa(n)+" "+quote)
		}
	}
}

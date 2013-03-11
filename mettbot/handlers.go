package mettbot

import (
	a "./answers"
	"fmt"
	irc "github.com/fluffle/goirc/client"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

func (bot *Mettbot) HandlerConnected() {
	bot.Join(*Channel)
	bot.MumblePing.StartMumblePing()
}

func (bot *Mettbot) HandlerDisconnected() {
	bot.MumblePing.StopMumblePing()
	bot.Quitted <- true
}

func (bot *Mettbot) HandlerJoin(line *irc.Line) {
	time.Sleep(1000 * time.Millisecond)
	actChannel := line.Args[0]
	bot.Topics[actChannel] = bot.ST.GetChannel(actChannel).Topic
}

func (bot *Mettbot) HandlerTopic(line *irc.Line) {
	actChannel := line.Args[0]
	newTopic := line.Args[1]
	oldTopic := bot.Topics[actChannel]
	bot.Topics[actChannel] = newTopic
	bot.Notice(actChannel, bot.diffTopic(oldTopic, newTopic))
}

func (bot *Mettbot) HandlerPrivmsg(line *irc.Line) {
	actChannel := line.Args[0]
	if actChannel == *Nick {
		actChannel = line.Nick
	}
	msg := line.Args[1]

	if strings.HasPrefix(msg, "!") {
		if rand.Float64() < *Probability {
			bot.Notice(actChannel, a.RandStr(a.IgnoreCmd))
		} else {
			bot.Command(actChannel, msg, line)
		}
	}

	if strings.HasPrefix(msg, *Nick+":") || strings.Contains(msg, "mettbot") || rand.Float64() < *Randomanswer {
		filteredMsg := msg
		if strings.HasPrefix(msg, *Nick+":") {
			if len(msg) > len(*Nick)+2 {
				filteredMsg = msg[len(*Nick)+2:]
			}
		}
		bot.Mentioned(actChannel, filteredMsg)
	}

	if matchedTwitter, _ := regexp.MatchString(*Twitterregex, msg); matchedTwitter {
		bot.GetTweet(actChannel, msg)
	}

	if line.Nick == "firebird" {
		if rand.Float64() < *Firebird {
			go bot.firebird(actChannel)
		}
	}

	if strings.Contains(msg, "\\a") {
		bot.DongDong(actChannel, msg)
	}

	if strings.Contains(strings.ToLower(msg), "mett") {
		bot.Mett()
	} else {
		bot.MsgSinceMett++
		if bot.MsgSinceMett >= *Offmessages {
			bot.Mett()
			bot.Notice(*Channel, fmt.Sprintf(a.RandStr(a.MettTime), bot.GetMett(*Channel)))
		} else if bot.MsgSinceMett == int(float32(*Offmessages)*0.95) {
			bot.Notice(*Channel, a.RandStr(a.Warning))
		}
	}
}

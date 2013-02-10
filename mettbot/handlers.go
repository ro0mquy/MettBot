package mettbot

import (
	a "./answers"
	irc "github.com/fluffle/goirc/client"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

func (bot *Mettbot) HandlerConnected()    { bot.Join(*Channel) }
func (bot *Mettbot) HandlerDisconnected() { bot.Quitted <- true }

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
	matchedTwitter, _ := regexp.MatchString(*Twitterregex, msg)

	switch {
	case strings.Index(msg, "!") == 0:
		if rand.Intn(int(1 / *Probability)) == 0 {
			bot.Notice(actChannel, a.RandStr(a.IgnoreCmd))
		} else {
			bot.Command(actChannel, msg, line)
		}
	case strings.Index(msg, *Nick+":") == 0:
		bot.Mett()
		bot.Mentioned(actChannel)
	case matchedTwitter:
		bot.GetTweet(actChannel, msg)
	case strings.Contains(msg, "mett") || strings.Contains(msg, "Mett") || strings.Contains(msg, "METT"):
		bot.Mett()
	default:
		bot.MsgSinceMett++
		if bot.MsgSinceMett >= *Offmessages {
			bot.Mett()
			bot.PostMett(*Channel)
		} else if bot.MsgSinceMett == int(float32(*Offmessages)*0.95) {
			bot.Notice(*Channel, a.RandStr(a.Warning))
		}
	}
}

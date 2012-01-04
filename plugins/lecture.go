package plugins

import (
	"ircclient"
	"time"
	"container/list"
	"json"
	"sync"
	"fmt"
)

type LecturePlugin struct {
	ic            *ircclient.IRCClient
	confplugin    *ConfigPlugin
	authplugin    *AuthPlugin
	notifications *list.List
	done          chan bool
	update        chan bool
	lock          sync.Mutex
}

type notification struct {
	when  int64
	entry configEntry
}

type configEntry struct {
	Name     string // AuD
	Time     string // Mon 13:15
	Channel  string // #faui2k11
	LongName string // Algorithmen und Datenstrukturen
	Academic string // Brinda
	Venue    string // H11
}

// Fills the list of notifications with all lectures for 
// the current day.
func (l *LecturePlugin) fillNotificationList() {
	l.notifications = list.New()

	l.confplugin.Lock()
	l.lock.Lock()
	defer l.lock.Unlock()
	defer l.confplugin.Unlock()
	options, _ := l.confplugin.Conf.Options("Lectures")
	curtime := time.LocalTime()
	for _, key := range options {
		value, _ := l.confplugin.Conf.String("Lectures", key)
		var lecture configEntry
		if err := json.Unmarshal([]byte(value), &lecture); err != nil {
			// panics should only happen during initialization, during runtime,
			// all config entries are checked before insertion.
			panic("LecturePlugin: Invalid JSON for key " + key + " : " + err.String())
		}

		timertime, err := time.Parse("Mon 15:04", lecture.Time)
		if err != nil {
			panic("LecturePlugin: Unable to parse time \"" + lecture.Time + "\" in config: " + err.String())
		}

		// Only consider notifications for the current day
		// TODO: Notifications for next day should also be considered
		// if necessary (because we send notifications ~10minutes
		// before the lecture starts)
		if timertime.Weekday != curtime.Weekday {
			continue
		}

		save1, save2, save3 := timertime.Hour, timertime.Minute, timertime.Second
		*timertime = *curtime
		timertime.Hour, timertime.Minute, timertime.Second = save1, save2, save3
		// TODO: Make this configurable
		timertime = time.SecondsToLocalTime(timertime.Seconds() - (60))
		if timertime.Seconds() >= curtime.Seconds() {
			fmt.Println("Registered lecture")
			l.notifications.PushFront(notification{timertime.Seconds(), lecture})
		}
	}
}

// Gets seconds until beginning of next day
func (l *LecturePlugin) untilNextDay() int64 {
	t := time.LocalTime()
	t.Hour, t.Minute, t.Second = 0, 0, 1
	// First second of new day
	sec := t.Seconds() + (24 * 60 * 60)
	return sec - time.Seconds()
}

func (l *LecturePlugin) sendNotifications() {
	for {
		var nextNotification int64 = time.Seconds() + (24 * 3600)
		// List contains the element we sent notifications for,
		// so we can remove them in the second loop.
		li := list.New()
		l.lock.Lock()
		for e := l.notifications.Front(); e != nil; e = e.Next() {
			notify := e.Value.(notification)
			entry := notify.entry
			if notify.when <= time.Seconds() {
				l.ic.SendLine("PRIVMSG " + entry.Channel + " :inb4 (" + entry.Time + "): \"" + entry.LongName + "\" (" + entry.Name + ") bei " + entry.Academic + ", Ort: " + entry.Venue)
				li.PushFront(e)
			} else if nextNotification > notify.when {
				nextNotification = notify.when
			}
		}
		for e := li.Front(); e != nil; e = e.Next() {
			li.Remove(e.Value.(*list.Element))
		}
		l.lock.Unlock()
		select {
		case <-l.done:
			return
		case <-time.After(l.untilNextDay() * 1e9):
			l.fillNotificationList()
		case <-time.After((nextNotification - time.Seconds()) * 1e9):
		case <-l.update:
			// Send notifications and refresh timer
		}
	}
}

func (l *LecturePlugin) Register(cl *ircclient.IRCClient) {
	l.ic = cl
	plugin, ok := l.ic.GetPlugin("config")
	if !ok {
		panic("LecturePlugin: Register: Unable to get configuration manager plugin")
	}
	l.confplugin, _ = plugin.(*ConfigPlugin)
	authplugin, ok := l.ic.GetPlugin("auth")
	if !ok {
		panic("LecturePlugin: Register: Unable to get authorization plugin")
	}
	l.authplugin, _ = authplugin.(*AuthPlugin)
	if !l.confplugin.Conf.HasSection("Lectures") {
		//l.confplugin.Conf.AddSection("Lectures")
		//test := &configEntry{"AuD", "Mon 13:15", "#go-faui2k11", "Algorithmen und Datenstrukturen", "Brinda", "H11"}
		//js, _ := json.Marshal(test)
		l.confplugin.Conf.AddOption("Lectures", test.Name, string(js))
	}
	l.done = make(chan bool)
	l.update = make(chan bool)
	l.fillNotificationList()
	go l.sendNotifications()
}

func (l *LecturePlugin) String() string {
	return "lecture"
}

func (l *LecturePlugin) Info() string {
	return "lecture notifications"
}

func (l *LecturePlugin) ProcessLine(msg *ircclient.IRCMessage) {
}

func (l *LecturePlugin) ProcessCommand(cmd *ircclient.IRCCommand) {
	if cmd.Command != "reglecture" && cmd.Command != "dellecture" {
		return
	}
	if l.authplugin.GetAccessLevel(cmd.Source) < 300 {
		l.ic.Reply(cmd, "You are not authorized to do that")
		return
	}
	switch cmd.Command {
	case "reglecture":
		if len(cmd.Args) != 6 {
			l.ic.Reply(cmd, "reglesson takes exactly 6 arguments:")
			l.ic.Reply(cmd, "Syntax: reglesson NAME TIME CHANNEL LONGNAME ACADEMIC VENUE")
			l.ic.Reply(cmd, "Example: reglesson AuD \"Mon 13:15\" #faui2k11 \"Algorithmen und Datenstrukturen\" Brinda H11")
			return
		}
		lecture := configEntry{cmd.Args[0], cmd.Args[1], cmd.Args[2], cmd.Args[3], cmd.Args[4], cmd.Args[5]}
		jlecture, _ := json.Marshal(lecture)
		l.confplugin.Lock()
		l.confplugin.Conf.AddOption("Lectures", fmt.Sprintf("%d", time.Seconds()), string(jlecture))
		l.confplugin.Unlock()
		l.ic.Reply(cmd, "Lecture added.")
		l.fillNotificationList()
		l.update <- true

	case "dellecture":
		// TODO
	}
}

func (l *LecturePlugin) Unregister() {
	l.done <- true
}

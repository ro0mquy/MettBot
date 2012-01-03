package plugins

import (
	"ircclient"
	"time"
	"container/list"
	"json"
	"log"
	"sync"
)

type LecturePlugin struct {
	ic            *ircclient.IRCClient
	confplugin    *ConfigPlugin
	notifications *list.List
	done chan bool
	update chan bool
	lock sync.Mutex
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
		if timertime.Weekday != curtime.Weekday {
			continue
		}
		save1, save2, save3 := timertime.Hour, timertime.Minute, timertime.Second
		*timertime = *curtime
		timertime.Hour, timertime.Minute, timertime.Second = save1, save2, save3
		// Notify 15 minutes before lecture
		// XXX - magic constant
		if timertime.Seconds()-(10*60) >= curtime.Seconds() {
			l.notifications.PushFront(notification{timertime.Seconds() - (10 * 60), lecture})
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
		l.lock.Lock()
		for e := l.notifications.Front(); e != nil; e = e.Next() {
			notify := e.Value.(notification)
			entry := notify.entry
			if notify.when <= time.Seconds() {
				l.ic.SendLine("PRIVMSG " + entry.Channel + " :inb4 (" + entry.Time + "): \"" + entry.LongName + "\" (" + entry.Name + ") bei " + entry.Academic + ", Ort: " + entry.Venue)
			} else if nextNotification > notify.when {
				nextNotification = notify.when
			}
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
	if !l.confplugin.Conf.HasSection("Lectures") {
		l.confplugin.Conf.AddSection("Lectures")
		test := &configEntry{"AuD", "Mon 13:15", "#go-faui2k11", "Algorithmen und Datenstrukturen", "Brinda", "H11"}
		js, _ := json.Marshal(test)
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
	// TODO: Lecture registration
}

func (l *LecturePlugin) Unregister() {
	l.done <- true
	// TODO: Write config
}

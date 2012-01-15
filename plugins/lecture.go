package plugins

import (
	"os"
	"ircclient"
	"time"
	"container/list"
	"json"
	"sync"
	"fmt"
	"log"
)

const notifyBefore = 600 // TODO: config

type LecturePlugin struct {
	ic            *ircclient.IRCClient
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

// Gets the _next_ time "date" matches, in seconds.
func nextAt(date string) (int64, os.Error) {
	timertime, err := time.Parse("Mon 15:04", date)
	if err != nil {
		return 0, err
	}
	curtime := time.LocalTime()

	// XXX - This is still quite ugly. Any ideas
	// on how to improve it?
	weekday := timertime.Weekday
	save1, save2, save3 := timertime.Hour, timertime.Minute, timertime.Second
	*timertime = *curtime
	timertime.Hour, timertime.Minute, timertime.Second = save1, save2, save3

	for timertime.Weekday != weekday || timertime.Seconds()-curtime.Seconds() <= notifyBefore {
		timertime = time.SecondsToLocalTime(timertime.Seconds() + (24 * 60 * 60))
	}
	return timertime.Seconds(), nil
}

// Fills the list of notifications with all lectures and
// the timestamp they take place
// Returns: Time for next lecture, -1 if no lecture is registered
func (l *LecturePlugin) fillNotificationList() int64 {
	var retval int64 = -1

	l.notifications = list.New()

	l.lock.Lock()
	defer l.lock.Unlock()

	options := l.ic.GetOptions("Lectures")
	for _, key := range options {
		value := l.ic.GetStringOption("Lectures", key)
		var lecture configEntry
		if err := json.Unmarshal([]byte(value), &lecture); err != nil {
			// panics should only happen during initialization, during runtime,
			// all config entries are checked before insertion.
			panic("LecturePlugin: Invalid JSON for key " + key + " : " + err.String())
		}

		time, err := nextAt(lecture.Time)
		if err != nil {
			log.Printf("Unable to parse time value for lecture %s: %s\n", lecture.Name, err.String())
			continue
		}
		notifyTime := time - notifyBefore
		l.notifications.PushFront(notification{notifyTime, lecture})

		if notifyTime < retval || retval == -1 {
			retval = notifyTime
		}
	}

	return retval
}

func (l *LecturePlugin) sendNotifications() {
	for {
		// TODO: container/heap and selective re-adding or so.
		// However, this should work for now...
		nextNotification := l.fillNotificationList()
		var timerChan <-chan int64
		// If nextNotification is less than zero, just wait indefinitely on this chan
		if nextNotification < 0 {
			timerChan = make(chan int64)
		} else {
			timerChan = time.After((nextNotification - time.Seconds()) * 1e9)
		}
		select {
		case <-l.done:
			return
		case <-timerChan:
		case <-l.update:
			// Send notifications and refresh timer
		}

		l.lock.Lock()
		for e := l.notifications.Front(); e != nil; e = e.Next() {
			notify := e.Value.(notification)
			entry := notify.entry
			if notify.when <= time.Seconds() {
				l.ic.SendLine("PRIVMSG " + entry.Channel + " :inb4 (" + entry.Time + "): \"" + entry.LongName + "\" (" + entry.Name + ") bei " + entry.Academic + ", Ort: " + entry.Venue)
			}
		}
		l.lock.Unlock()
	}
}

func (l *LecturePlugin) Register(cl *ircclient.IRCClient) {
	l.ic = cl
	l.done = make(chan bool)
	l.update = make(chan bool)
	go l.sendNotifications()
	cl.RegisterCommandHandler("reglecture", 0, 300, l)
	// TODO: dellecture
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
	switch cmd.Command {
	case "reglecture":
		if len(cmd.Args) != 6 {
			l.ic.Reply(cmd, "reglecture takes exactly 6 arguments:")
			l.ic.Reply(cmd, "Syntax: reglecture NAME TIME CHANNEL LONGNAME ACADEMIC VENUE")
			l.ic.Reply(cmd, "Example: reglecture AuD \"Mon 13:15\" #faui2k11 \"Algorithmen und Datenstrukturen\" Brinda H11")
			return
		}
		lecture := configEntry{cmd.Args[0], cmd.Args[1], cmd.Args[2], cmd.Args[3], cmd.Args[4], cmd.Args[5]}
		_, err := time.Parse("Mon 15:04", lecture.Time)
		if err != nil {
			l.ic.Reply(cmd, "Invalid date specified: "+err.String())
			return
		}
		jlecture, _ := json.Marshal(lecture)
		l.ic.SetStringOption("Lectures", fmt.Sprintf("%d", time.Seconds()), string(jlecture))
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

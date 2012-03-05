package plugins

// Important: Register this module _AFTER_ successful connection!

import (
	"log"
	"fmt"
	"ircclient"
	"http"
	"xml"
	"time"
)

type root struct {
	Events Events
}

type Events struct {
	Event []Event
}

type Event struct {
	Id         int `xml:"attr"`
	Submission *Submission
	Judging    *Judging
}

type Submission struct {
	Id       int `xml:"attr"`
	Team     string
	Problem  string
	Language string
}

type Judging struct {
	Id       int    `xml:"attr"`
	Submitid int    `xml:"attr"`
	Result   string `xml:"chardata"`
}

type HalloWeltPlugin struct {
	ic   *ircclient.IRCClient
	done chan bool
	solved map[string](map[string]bool)
	subid map[int]int
}

func (q *HalloWeltPlugin) Register(cl *ircclient.IRCClient) {
	q.ic = cl
	var client http.Client
	q.done = make(chan bool)
	q.subid = make(map[int]int)
	q.solved = make(map[string](map[string]bool))
	p, err := q.ic.GetIntOption("HalloWelt", "polling")
	polling := int64(p)
	channel := q.ic.GetStringOption("HalloWelt", "channel")
	url := q.ic.GetStringOption("HalloWelt", "url")
	if err != nil || channel == "" || url == "" {
		log.Println("WARNING: No complete HalloWelt configuration found. Setting defaults. Please edit your config and reload this plugin")
		q.ic.SetStringOption("HalloWelt", "channel", "hallowelt")
		q.ic.SetStringOption("HalloWelt", "url", "https://EDITTHIS")
		q.ic.SetIntOption("HalloWelt", "polling", 60)
		go func() {
			<-q.done
			q.done <- true
		}()
		return
	}
	go func() {
		last := -1
		for {
			t := time.After(polling * 1e9)
			select {
			case <-t:
			case <-q.done:
				q.done <- true
				return
			}
			response, err := client.Get(url)
			if response.StatusCode != 200 || err != nil {
				log.Println("ERROR: (HalloWelt): Unable to get current event list from DomJudge")
				time.Sleep(120 * 1e9)
				continue
			}
			var res root
			// Parse XML
			xml.Unmarshal(response.Body, &res)
			response.Body.Close()
			if err != nil || last == len(res.Events.Event) {
				continue
			}
			if last == -1 {
				last = len(res.Events.Event)
				for i := 0; i < last; i = i + 1 {
					if res.Events.Event[i].Submission != nil {
						q.subid[res.Events.Event[i].Submission.Id] = i
					}
					if res.Events.Event[i].Judging != nil {
						id, b := q.subid[res.Events.Event[i].Judging.Submitid]
						if b == false {
							continue
						}
						if q.solved[res.Events.Event[id].Submission.Team] == nil {
							q.solved[res.Events.Event[id].Submission.Team] = make(map[string]bool)
						}
						q.solved[res.Events.Event[id].Submission.Team][res.Events.Event[id].Submission.Problem] = true
					}
				}
				continue
			}
			// Report new submissions
			for i := last; i < len(res.Events.Event); i = i + 1 {
				jd := res.Events.Event[i].Submission
				if jd != nil {
					q.subid[jd.Id] = i
					continue
				}
				ev := res.Events.Event[i].Judging
				if ev == nil || ev.Result != "correct" {
					continue
				}
				tries, team, problem := 0, "", ""
				for i := len(res.Events.Event) - 1; i >= 0; i = i - 1 {
					if res.Events.Event[i].Submission == nil {
						continue
					}
					if res.Events.Event[i].Submission.Id == ev.Submitid {
						tries = 1
						team = res.Events.Event[i].Submission.Team
						problem = res.Events.Event[i].Submission.Problem
						continue
					}
					if tries != 0 && res.Events.Event[i].Submission.Problem == problem && res.Events.Event[i].Submission.Team == team {
						tries = tries + 1
					}
				}
				if tries == 0 || team == "DOMjudge" || q.solved[team][problem] == true {
					// Ignore invalid input
					continue
				}
				if q.solved[team] == nil {
					q.solved[team] = make(map[string]bool)
				}
				q.solved[team][problem] = true
				q.ic.SendLine("PRIVMSG #" + channel + " :" + team + " solved " + problem + " (after " + fmt.Sprintf("%d", tries-1) + " failed attempts)")
			}
			last = len(res.Events.Event)
		}
	}()
}

func (q *HalloWeltPlugin) String() string {
	return "hallowelt"
}

func (q *HalloWeltPlugin) Info() string {
	return "DomJudge live ticker"
}

func (q *HalloWeltPlugin) Usage(cmd string) string {
	return "This plugin provides no commands"
}

func (q *HalloWeltPlugin) ProcessLine(msg *ircclient.IRCMessage) {
}

func (q *HalloWeltPlugin) ProcessCommand(cmd *ircclient.IRCCommand) {
}

func (q *HalloWeltPlugin) Unregister() {
	q.done <- true
	<-q.done
}

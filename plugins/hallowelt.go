package plugins

import (
	"log"
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
	ic *ircclient.IRCClient
	done chan bool
}

func (q *HalloWeltPlugin) Register(cl *ircclient.IRCClient) {
	q.ic = cl
	var client http.Client
	q.done = make(chan bool)
//	go func() {
		for {
			// TODO: Configurable
			t := time.After(1 * 1e9)
			select {
			case <- t:
			case <- q.done:
			q.done <- true
			return;
			}
			//response, _ := client.Get("https://bot:hallowelt@icpc.informatik.uni-erlangen.de/domjudge/plugin/event.php")
			// Um den DOMJudge nicht uebermaessig in der Entwicklungsphase zu pollen 
			// TODO: Konfigurierbar
			response, err := client.Get("http://d-paulus.de/tmp.xml")
			if response.StatusCode != 200 || err != nil {
				log.Println("ERROR: (HalloWelt): Unable to get current event list from DomJudge")
				time.Sleep(120 * 1e9)
				continue
			}
			var res root
			// Parse XML
			xml.Unmarshal(response.Body, &res)
			response.Body.Close()
			last, err := q.ic.GetIntOption("HalloWelt", "last")
			if err != nil {
				q.ic.SetIntOption("HalloWelt", "last", len(res.Events.Event))
			}
			if last == len(res.Events.Event) {
				continue
			}
			for i := last; i < len(res.Events.Event); i = i + 1 {
				ev := res.Events.Event[i].Judging
				if ev == nil {
					continue
				}
				if ev.Result != "correct" {
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
				if tries == 0 {
					// Ignore invalid input
					continue
				}
				log.Printf("%s solved %s (after %d failed attempts)", team, problem, tries - 1)
			}
			log.Fatal("Bye")
		}
//	}()
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

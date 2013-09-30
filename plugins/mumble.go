package plugins

import (
	"../ircclient"
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"regexp"
	"time"
)

const (
	audiomett_topic_regexp      = `((?:^|\|)[^|]*)audiomett:[^|]*?(\s*(?:$|\|))`
	audiomett_topic_replacement = "${1}audiomett: %d %s$2"
	ping_intervall              = 30 * time.Second
)

type MumblePlugin struct {
	ic                   *ircclient.IRCClient
	quit                 chan bool
	audiomettTopicRegexp *regexp.Regexp
	topic                string
	lastUsers            int32
	running              bool
}

func (q *MumblePlugin) String() string {
	return "mumble"
}

func (q *MumblePlugin) Info() string {
	return "Displays the number of users connected to the mumble server in the channel topic"
}

func (q *MumblePlugin) Usage(cmd string) string {
	// just for interface saturation
	return ""
}

func (q *MumblePlugin) Register(cl *ircclient.IRCClient) {
	q.ic = cl
	q.quit = make(chan bool)
	q.audiomettTopicRegexp = regexp.MustCompile(audiomett_topic_regexp)
	q.topic = ""
	q.lastUsers = 0
	q.running = false

	// check for config
	if q.ic.GetStringOption("Mumble", "server") == "" || q.ic.GetStringOption("Mumble", "channel") == "" {
		log.Fatal("Need server and channel for Mumble ping")
	}

	q.Start()
}

func (q *MumblePlugin) Unregister() {
	q.Stop()
}

func (q *MumblePlugin) ProcessLine(msg *ircclient.IRCMessage) {
	// log topic
	if msg.Command == "332" && msg.Args[0][1:] == q.ic.GetStringOption("Mumble", "channel") { // announce of topic during joining of channel
		q.topic = msg.Args[1]
	} else if msg.Command == "TOPIC" && msg.Target[1:] == q.ic.GetStringOption("Mumble", "channel") {
		q.topic = msg.Args[0]
	}
}

func (q *MumblePlugin) ProcessCommand(cmd *ircclient.IRCCommand) {
	return
}

func (q *MumblePlugin) run() {
	for {
		select {
		case <-time.After(ping_intervall):
			_, users, _, _, err := q.ping(q.ic.GetStringOption("Mumble", "server"))
			if err != nil && !err.(net.Error).Timeout() {
				log.Println(err)
				continue
			}

			if users == q.lastUsers {
				continue
			}

			var u string
			if users == 1 {
				u = "user"
			} else {
				u = "users"
			}

			replacement := fmt.Sprintf(audiomett_topic_replacement, users, u)
			newTopic := q.audiomettTopicRegexp.ReplaceAllString(q.topic, replacement)
			if newTopic == q.topic {
				continue
			}

			q.lastUsers = users

			// set topic in topicdiff plugin so, it won't get diffed
			topicdiff := q.ic.GetPlugin("topicdiff").(*TopicDiffPlugin)
			if topicdiff != nil {
				topicdiff.SetTopic(q.ic.GetStringOption("Mumble", "channel"), newTopic)
			}

			q.ic.SendLine("TOPIC #" + q.ic.GetStringOption("Mumble", "channel") + " :" + newTopic)
		case <-q.quit:
			return
		}
	}
}

func (q *MumblePlugin) Start() {
	if !q.running {
		q.running = true
		go q.run()
	}
}

func (q *MumblePlugin) Stop() {
	if q.running {
		q.quit <- true
		q.running = false
	}
}

func (q *MumblePlugin) ping(server string) (version string, users_connected, users_maximum, allowed_bandwidth int32, err error) {
	// Get UDPConn
	addr, err := net.ResolveUDPAddr("udp", server)
	if err != nil {
		return
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return
	}
	conn.SetDeadline(time.Now().Add(5 * time.Second))

	// Send request
	request := make([]byte, 12)
	for i := 0; i < 4; i++ {
		request[i] = 0
	}
	bufrand := bufio.NewReader(rand.Reader)
	for i := 4; i < 4+8; i++ {
		request[i], err = bufrand.ReadByte()
		if err != nil {
			return
		}
	}

	n, err := conn.Write(request)
	if err != nil {
		return
	} else if n != len(request) {
		// TODO check if this might actually happen
		err = fmt.Errorf("UDP request got sent incompletely")
		return
	}

	// Read response
	response := make([]byte, 24)
	n, err = conn.Read(response)
	if err != nil {
		return
	} else if n != len(response) {
		err = fmt.Errorf("Unexpected UDP response length")
		return
	}

	// Verify response
	for i := 4; i < 4+8; i++ {
		if response[i] != request[i] {
			err = fmt.Errorf("Response ident does not match request ident")
			return
		}
	}

	// Parse response
	version = fmt.Sprintf("%d.%d.%d", response[1], response[2], response[3])
	err = binary.Read(bytes.NewReader(response[12:16]), binary.BigEndian, &users_connected)
	if err != nil {
		return
	}
	err = binary.Read(bytes.NewReader(response[16:20]), binary.BigEndian, &users_maximum)
	if err != nil {
		return
	}
	err = binary.Read(bytes.NewReader(response[20:24]), binary.BigEndian, &allowed_bandwidth)
	if err != nil {
		return
	}

	return
}

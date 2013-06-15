package mettbot

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"regexp"
	"time"
)

type _mumbleping struct {
	bot     *Mettbot
	server  string
	last    int32
	running bool
	quit    chan bool
}

func (m *_mumbleping) InitMumblePing(mett *Mettbot) {
	m.bot = mett
	m.server = *MumbleServer
	m.running = false
	m.quit = make(chan bool)
}

func (m *_mumbleping) run() {
	audiomettTopicRegexp := regexp.MustCompile(*MumbleTopicregex)
	for {
		select {
		case <-time.After(30 * time.Second):
			_, users, _, _, err := m.doMumblePing()
			if err != nil {
				log.Println(err)
				continue
			}
			if users == m.last {
				continue
			}
			m.last = users
			oldTopic := m.bot.ST.GetChannel(*Channel).Topic
			if oldTopic == "" {
				continue
			}
			var u string
			if users == 1 {
				u = "user"
			} else {
				u = "users"
			}
			repl := fmt.Sprintf("${1}audiomett: %d %s$2", users, u)
			newTopic := audiomettTopicRegexp.ReplaceAllString(oldTopic, repl)
			if newTopic == oldTopic {
				continue
			}
			m.bot.Topic(*Channel, newTopic)
		case <-m.quit:
			return
		}
	}
}

func (m *_mumbleping) StartMumblePing() {
	if !m.running {
		m.running = true
		go m.run()
	}
}

func (m *_mumbleping) StopMumblePing() {
	if m.running {
		m.quit <- true
		m.running = false
	}
}

func (m *_mumbleping) doMumblePing() (version string, users_connected, users_maximum, allowed_bandwidth int32, err error) {
log.Println("Pinging Mumble server...")
	// Get UDPConn
log.Println("Doing address lookup...")
	addr, err := net.ResolveUDPAddr("udp", m.server)
	if err != nil {
		return
	}
log.Println("Connecting to Server...")
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return
	}

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
log.Println("Sending request...")
	n, err := conn.Write(request)
	if err != nil {
		return
	} else if n != len(request) {
		// TODO check if this might actually happen
		err = errors.New("UDP request got sent incompletely")
		return
	}

	// Read response
log.Println("Reading respons...")
	response := make([]byte, 24)
	n, err = conn.Read(response)
	if err != nil {
		return
	} else if n != len(response) {
		err = errors.New("Unexpected UDP response length")
	}

	// Verify response
log.Println("Verifing response...")
	for i := 4; i < 4+8; i++ {
		if response[i] != request[i] {
			err = errors.New("Response ident does not match request ident")
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

log.Println("Done pinging.")
	return
}

package plugins

import (
	"../ircclient"
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"hash"
	"log"
	"net"
	"strings"
)

type EvaluationPlugin struct {
	ic   *ircclient.IRCClient
	c    net.Conn
	hmac hash.Hash
	done chan bool
}

func (q *EvaluationPlugin) Register(cl *ircclient.IRCClient) {
	q.ic = cl
	q.done = make(chan bool, 1)
	var hmacKey []byte

	// Read key for HMAC from config
	if value := q.ic.GetStringOption("Eval", "key"); value == "" {
		log.Println("WARNING: No HMAC key for evaluation plugin specified")
		hmacKey = make([]byte, 20)
		if n, err := rand.Read(hmacKey); n != 20 || err != nil {
			log.Printf("ERROR: Unable to generate one. Exiting. (%s)", err.Error())
			return
		}
		q.ic.SetStringOption("Eval", "key", fmt.Sprintf("%x", hmacKey))
		log.Println("WARNING: Auto-generated one.")
	} else {
		// Hm. This is hackish. Proposals? :-)
		hmacKey = hmacKey[0:0]
		for len(value) > 1 {
			var tmp int
			fmt.Sscanf(value[0:2], "%x", &tmp)
			hmacKey = append(hmacKey, uint8(tmp))
			value = value[2:]
		}
	}
	// Network config
	var hostport string
	if hostport = q.ic.GetStringOption("Eval", "hostport"); hostport == "" {
		log.Println("WARNING: Added default listener for EvaluationPlugin")
		q.ic.SetStringOption("Eval", "hostport", "0.0.0.0:5486")
		hostport = "0.0.0.0:4386"
	}
	// Channel
	var channel string
	if channel = q.ic.GetStringOption("Eval", "channel"); channel == "" {
		log.Println("WARNING: No channel set in evaluation plugin options. Please set one.")
		return
	}

	addr, err := net.ResolveUDPAddr("udp", hostport)
	if err != nil {
		log.Println("ERROR: Internal error in EvaluationPlugin")
		return
	}
	q.c, err = net.ListenUDP("udp", addr)
	if err != nil {
		log.Println("ERROR: Unable to open listener for evaluation plugin")
		return
	}
	q.hmac = hmac.New(sha1.New, hmacKey)
	go func() {
		buf := make([]byte, 512)
		for {
			n, err := q.c.Read(buf)
			select {
			case <-q.done:
				return
			default:
			}
			if err != nil {
				log.Println("ERROR: Unable to receive on evaluation notification")
				return
			}
			if n < q.hmac.Size()+1 {
				log.Println("WARNING: Invalid evaluation packet received")
				continue
			}
			hash := buf[0:20]
			q.hmac.Reset()
			q.hmac.Write(buf[20:n])
			if bytes.Compare(hash, q.hmac.Sum(nil)) != 0 {
				log.Println("WARNING: Wrong HMAC on evaluation packet")
				continue
			}
			payload := string(buf[20:n])
			parts := strings.SplitN(payload, "\t", 3)
			if len(parts) < 3 {
				continue
			}
			q.ic.SendLine("PRIVMSG #" + channel + " :New lecture evaluation has been added in Semester " + parts[1] + ": " + parts[0] + ", " + parts[2])
		}
	}()
}

func (q *EvaluationPlugin) String() string {
	return "evaluation"
}

func (q *EvaluationPlugin) Info() string {
	return "notifies when new evaluations are available"
}

func (q *EvaluationPlugin) Usage(cmd string) string {
	return "This plugin provides no commands"
}

func (q *EvaluationPlugin) ProcessLine(msg *ircclient.IRCMessage) {
}

func (q *EvaluationPlugin) ProcessCommand(cmd *ircclient.IRCCommand) {
	return
}

func (q *EvaluationPlugin) Unregister() {
	q.done <- true
	if q.c != nil {
		q.c.Close()
	}
	return
}

package plugins

import (
	"../ircclient"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strings"
)

type TopicDiffPlugin struct {
	ic     *ircclient.IRCClient
	topics map[string]string
}

func (q *TopicDiffPlugin) String() string {
	return "topicdiff"
}

func (q *TopicDiffPlugin) Info() string {
	return `diffs new and old topic wordwise`
}

func (q *TopicDiffPlugin) Usage(cmd string) string {
	// just for interface saturation
	return ""
}

func (q *TopicDiffPlugin) Register(cl *ircclient.IRCClient) {
	q.ic = cl
	q.topics = make(map[string]string)
}

func (q *TopicDiffPlugin) Unregister() {
	return
}

func (q *TopicDiffPlugin) ProcessLine(msg *ircclient.IRCMessage) {
	if msg.Command == "332" { // announce of topic during joining of channel
		q.topics[msg.Args[0]] = msg.Args[1]
	} else if msg.Command == "TOPIC" {
		oldTopic := q.topics[msg.Target]
		newTopic := msg.Args[0]
		q.topics[msg.Target] = newTopic

		message, err := q.diff(oldTopic, newTopic)
		if err != nil {
			log.Println(err)
			return
		}

		q.ic.ReplyMsg(msg, message)
	}
}

func (q *TopicDiffPlugin) ProcessCommand(cmd *ircclient.IRCCommand) {
	return
}

func (q *TopicDiffPlugin) diff(oldTopic, newTopic string) (outStr string, err error) {
	oldFile, err := ioutil.TempFile("", ".mettbotWdiffOld")
	if err != nil {
		return
	}

	n, err := oldFile.WriteString(oldTopic)
	if n != len(oldTopic) || err != nil {
		return
	}
	oldFile.Close()

	newFile, err := ioutil.TempFile("", ".mettbotWdiffNew")
	if err != nil {
		return
	}

	n, err = newFile.WriteString(newTopic)
	if n != len(newTopic) || err != nil {
		return
	}
	newFile.Close()

	defer func() {
		os.Remove(oldFile.Name())
		os.Remove(newFile.Name())
	}()

	var db string // DeletionBegin
	var de string // DeletionEnd
	var ib string // InsertionBegin
	var ie string // InsertionEnd

	for i := 0; true; i++ {
		// generate new limiters
		rdb := rune(rand.Intn(127-32) + 32)
		rde := rune(rand.Intn(127-32) + 32)
		rib := rune(rand.Intn(127-32) + 32)
		rie := rune(rand.Intn(127-32) + 32)

		// check if some limiters are equal
		contains := rdb == rde || rdb == rib || rdb == rie
		contains = contains || rde == rib || rde == rie || rib == rie

		// check if strings contain limiters
		contains = contains || strings.ContainsRune(oldTopic, rdb)
		contains = contains || strings.ContainsRune(oldTopic, rde)
		contains = contains || strings.ContainsRune(oldTopic, rib)
		contains = contains || strings.ContainsRune(oldTopic, rie)

		contains = contains || strings.ContainsRune(newTopic, rdb)
		contains = contains || strings.ContainsRune(newTopic, rde)
		contains = contains || strings.ContainsRune(newTopic, rib)
		contains = contains || strings.ContainsRune(newTopic, rie)

		if contains == false {
			db = fmt.Sprintf("%c", rdb)
			de = fmt.Sprintf("%c", rde)
			ib = fmt.Sprintf("%c", rib)
			ie = fmt.Sprintf("%c", rie)
			break
		}

		// don't generate endless loop if no machting limiters exist
		if i > 1000000 {
			err = fmt.Errorf("Found no machting limiters for diffing")
			return
		}
	}

	cmd := exec.Command("wdiff", "-w"+db, "-x"+de, "-y"+ib, "-z"+ie, oldFile.Name(), newFile.Name())
	out, _ := cmd.Output()
	outStr = string(out)

	coloring := map[string]string{ // http://oreilly.com/pub/h/1953
		db: "\x035\x1F\x02",
		de: "\x0F\x0315",
		ib: "\x033\x02",
		ie: "\x0F\x0315",
	}

	for n, v := range coloring {
		outStr = strings.Replace(outStr, n, v, -1)
	}

	return outStr, nil
}

package plugins

import (
	"../ircclient"
	"log"
	"os/exec"
	"strings"
)

type CorrectionPlugin struct {
	ic       *ircclient.IRCClient
	lastMsgs map[string]string
}

func (q *CorrectionPlugin) String() string {
	return "correction"
}

func (q *CorrectionPlugin) Info() string {
	return `applies a 's/foo/bar/'-style correction line to the last message of a user`
}

func (q *CorrectionPlugin) Usage(cmd string) string {
	// just for interface saturation
	return ""
}

func (q *CorrectionPlugin) Register(cl *ircclient.IRCClient) {
	q.ic = cl
	q.lastMsgs = make(map[string]string)
}

func (q *CorrectionPlugin) Unregister() {
	return
}

func (q *CorrectionPlugin) ProcessLine(msg *ircclient.IRCMessage) {
	if msg.Command != "PRIVMSG" || len(msg.Args) == 0 {
		return
	}

	if strings.HasPrefix(msg.Args[0], "s/") {
		correction, err := q.correct(q.lastMsgs[msg.Source], msg.Args[0])
		if err != nil {
			_, ok := err.(*exec.ExitError)
			if !ok {
				log.Println(err)
				return
			}
			q.ic.ReplyMsg(msg, correction)
			return
		}
		q.ic.ReplyMsg(msg, strings.SplitN(msg.Source, "!", 2)[0]+" meant: "+correction)
	} else {
		q.lastMsgs[msg.Source] = msg.Args[0]
	}
}

func (q *CorrectionPlugin) ProcessCommand(cmd *ircclient.IRCCommand) {
	return
}

func (q *CorrectionPlugin) correct(message, replacement string) (string, error) {
	sed := exec.Command("sed", replacement)
	sed.Stdin = strings.NewReader(message)
	correction, err := sed.CombinedOutput()
	return string(correction), err
}

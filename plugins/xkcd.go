package plugins

import (
	"ircclient"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"json"
	"rand"
)

type comic struct {
	Num float64
	Title string
	Transcript string
	Alt string
}

func (c *comic) readJSON(path string) (err os.Error) {
	var (
		file *os.File
		raw []byte
	)
	if file, err = os.Open(path); err != nil {
		return
	}
	if raw, err = ioutil.ReadAll(file); err != nil {
		return
	}
	err = json.Unmarshal(raw, c)
	file.Close()
	return
}

func (c *comic) titleContains(text string) bool {
	return strings.Contains(c.Title, text)
}

func (c *comic) contains(text string) bool {
	return strings.Contains(c.Title, text) || strings.Contains(c.Transcript, text)
}

// randomComic returns a comic number between 1 and x.maxComic (inclusive) except 404.
func (x *XKCDPlugin) randomComic() int {
	r := rand.Intn(x.maxComic - 1)
	if r < 403 {
		return r + 1
	}
	return r + 2
}

// matchingComic returns the number a comic that contains all of the strings is args.
// If no comic is found, it returns -1 or 404.
func (x *XKCDPlugin) matchingComic(args []string) int {
	return -1
}

type XKCDPlugin struct {
	ic *ircclient.IRCClient
	maxComic int
}

func (x *XKCDPlugin) Register(cl *ircclient.IRCClient) {
	x.ic = cl
	x.ic.RegisterCommandHandler("xkcd", 0, 0, x)
	x.maxComic = 1000
}

func (x *XKCDPlugin) String() string {
	return "xkcd"
}

func (x *XKCDPlugin) Info() string {
	return "search xkcd"
}

func (x *XKCDPlugin) ProcessLine(msg *ircclient.IRCMessage) {
}

func (x *XKCDPlugin) ProcessCommand(cmd *ircclient.IRCCommand) {
	if cmd.Command == "xkcd" {
		var number int
		if len(cmd.Args) == 0 {
			number = x.randomComic()
		} else {
			number = x.matchingComic(cmd.Args)
		}
		if number == -1 {
			x.ic.Reply(cmd, "Sorry, didn’t find a matching comic.")
		} else {
			x.ic.Reply(cmd, fmt.Sprintf("http://xkcd.org/%v/", number))
		}
	}
}

func (x *XKCDPlugin) Unregister() {
}

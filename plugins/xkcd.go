package plugins

import (
	"ircclient"
	"fmt"
	"log"
	"io/ioutil"
	"os"
	"strings"
	"http"
	"json"
	"rand"
	"time"
)

type comic struct {
	Num int
	Title string
}

func getCurrentComic() int {
	response, _ := http.Get("http://xkcd.com/info.0.json")
	if response.StatusCode != 200 {
		return 0
	}
	content, _ := ioutil.ReadAll(response.Body)
	var c comic
	if err := json.Unmarshal(content, &c); err != nil {
		return 0
	}
	return c.Num
}

func (c *comic) readJSON(number int) (err os.Error) {
	raw, err := ioutil.ReadFile(fmt.Sprintf("comics/%v.json", number))
	if err != nil {
		if downloadJSON(number) {
			err = c.readJSON(number)
		}
		return
	}
	err = json.Unmarshal(raw, c)
	return
}

func downloadJSON(number int) bool {
	url := fmt.Sprintf("http://xkcd.com/%v/info.0.json", number)
	response, _ := http.Get(url)
	if response.StatusCode != 200 {
		return false
	}
	content, _ := ioutil.ReadAll(response.Body)
	ioutil.WriteFile(fmt.Sprintf("comics/%v.json", number), content, 0644)
	return true
}

func (c *comic) titleContains(text string) bool {
	return strings.Contains(strings.ToLower(c.Title), strings.ToLower(text))
}

// randomComic returns a comic number between 1 and x.maxComic (inclusive) except 404.
func (x *XKCDPlugin) randomComic() int {
	r := rand.Intn(x.maxComic - 1)
	if r < 403 {
		return r + 1
	}
	return r + 2
}

const prob404 float32 = 0.5

// matchingComic returns the number of a comic that contains all of the strings in args.
// If no comic is found, it returns -1 or 404.
func (x *XKCDPlugin) matchingComic(args []string) int {
	currentTime := time.LocalTime()
	if currentTime.Day != x.lastUpdate.Day || currentTime.Month != x.lastUpdate.Month || currentTime.Year != x.lastUpdate.Year {
		x.updateComics()
		x.lastUpdate = currentTime
	}
	numbers := make([]int, 0, 10)
	for _, c := range x.comics {
		contains := true
		for _, a := range args {
			contains = contains && c.titleContains(a)
		}
		if contains {
			numbers = append(numbers, c.Num)
		}
	}
	if len(numbers) == 0 {
		if rand.Float32() < prob404 {
			return 404
		}
		return -1
	}
	return numbers[rand.Intn(len(numbers))]
}

type XKCDPlugin struct {
	ic *ircclient.IRCClient
	maxComic int
	comics []comic
	lastUpdate *time.Time
}

func (x *XKCDPlugin) Register(cl *ircclient.IRCClient) {
	var err os.Error
	x.ic = cl
	x.ic.RegisterCommandHandler("xkcd", 0, 0, x)
	x.maxComic = getCurrentComic()
	if x.maxComic == 0 {
		if x.maxComic, err = x.ic.GetIntOption("XKCD", "maxComic"); err != nil {
			x.maxComic = 0
		}
	}
	x.ic.SetIntOption("XKCD", "maxComic", x.maxComic)
	x.comics = make([]comic, 0, x.maxComic)
	if err = os.MkdirAll("comics", 0755); err != nil {
		log.Fatalln(err)
	}
	for i := 1; i <= x.maxComic; i++ {
		var c comic
		if err := c.readJSON(i); err == nil {
			x.comics = append(x.comics, c)
		}
	}
	x.lastUpdate = time.LocalTime()
}

func (x *XKCDPlugin) updateComics() {
	newMax := getCurrentComic()
	var err os.Error
	if newMax == 0 {
		if newMax, err = x.ic.GetIntOption("XKCD", "maxComic"); err != nil {
			newMax = 0
		}
	}
	if newMax <= x.maxComic {
		return
	}
	x.ic.SetIntOption("XKCD", "maxComic", newMax)
	for i := x.maxComic; i <= newMax; i++ {
		var c comic
		if err = c.readJSON(i); err == nil {
			x.comics = append(x.comics, c)
		}
	}
}


func (x *XKCDPlugin) String() string {
	return "xkcd"
}

func (x *XKCDPlugin) Info() string {
	return "search xkcd"
}

func (x *XKCDPlugin) Usage(cmd string) string {
	switch cmd {
	case "xkcd":
		return "xkcd <search term>: returns the url for the xkcd comic containing <search term>"
	}
	return ""
}

func (x *XKCDPlugin) ProcessLine(msg *ircclient.IRCMessage) {
}

func (x *XKCDPlugin) ProcessCommand(cmd *ircclient.IRCCommand) {
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

func (x *XKCDPlugin) Unregister() {
}

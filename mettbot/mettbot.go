package mettbot

import (
	a "./answers"
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	irc "github.com/fluffle/goirc/client"
	"html"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var Host *string = flag.String("host", "irc.ps0ke.de:2342", "IRC server")
var Channel *string = flag.String("channel", "#metttest", "IRC channel")
var Nick *string = flag.String("nick", "rohmett", "IRC nick")
var Longnick *string = flag.String("longnick", "Le MettBot", "IRC fullname")
var Timeformat *string = flag.String("timeformat", "2006-01-02T15:04", "Time format string (standard date: 2006-01-02T15:04:05)")
var Quotes *string = flag.String("quotes", "mett_quotes.txt", "Quote database file")
var Metts *string = flag.String("metts", "mett_metts.txt", "Metts database file")
var Offtime *int = flag.Int("offtime", 4, "Number of hours of offtopic content befor posting mett content")
var Offmessages *int = flag.Int("offmessages", 300, "Number of messages of offtopic content befor posting mett content")
var Probability *float64 = flag.Float64("probability", 0.1, "Probability that the bot ignores a command")
var MumbleServer *string = flag.String("mumbleserver", "avidrain.de:64738", "Mumble server")
var MumbleTopicregex *string = flag.String("mumbletopicregex", "((?:^|\\|)[^|]*)audiomett:[^|]*?(\\s*(?:$|\\|))", "The regex to match Mumble topic snippet")
var Twitterregex *string = flag.String("twitterregex", "\\S*twitter\\.com\\/\\S+\\/status(es)?\\/(\\d+)\\S*", "The regex to match Twitter URLs")
var Firebird *float64 = flag.Float64("firebird", 0.001, "Probability firebird gets a question")
var Randomanswer *float64 = flag.Float64("randomanswer", 0.05, "The probability that the bot randomly answers to a users message")

func init() {
	flag.Parse()
	rand.Seed(time.Now().Unix())
}

type Mettbot struct {
	*irc.Conn
	MumblePing      *_mumbleping
	Quitted         chan bool
	QuotesPrnt      chan string
	MettsPrnt       chan string
	QuotesLinesPrnt chan int
	MettsLinesPrnt  chan int
	Input           chan string
	IsMett          chan bool
	ReallyQuit      bool
	Topics          map[string]string
	MsgSinceMett    int
}

func NewMettbot(nick string, args ...string) *Mettbot {
	bot := &Mettbot{
		irc.Conn:        irc.SimpleClient(nick, args...),
		MumblePing:      new(_mumbleping),
		Quitted:         make(chan bool),
		QuotesPrnt:      make(chan string),
		MettsPrnt:       make(chan string),
		QuotesLinesPrnt: make(chan int),
		MettsLinesPrnt:  make(chan int),
		Input:           make(chan string, 4),
		IsMett:          make(chan bool),
		ReallyQuit:      false,
		Topics:          make(map[string]string),
		MsgSinceMett:    0,
	}
	bot.EnableStateTracking()
	bot.MumblePing.InitMumblePing(bot)
	bot.Flood = true

	return bot
}

func (bot *Mettbot) Syntax(channel string) {
	bot.Notice(channel, a.RandStr(a.Syntax))
}

func (bot *Mettbot) Mett() {
	bot.MsgSinceMett = 0
	bot.IsMett <- true
}

func (bot *Mettbot) CheckMett() {
	for {
		select {
		case <-bot.IsMett:
		case <-time.After(time.Duration(*Offtime) * time.Hour):
			hour := time.Now().Hour()
			if hour < 1 || hour >= 8 {
				bot.Notice(*Channel, fmt.Sprintf(a.RandStr(a.MettTime), bot.GetMett(*Channel)))
			}
		}
	}
}

func (bot *Mettbot) GetMett(channel string) string {
	fi, err := os.Open(*Metts)
	if err != nil {
		log.Println(err)
		bot.Notice(channel, "Failed to open mett database")
		return "Error"
	}
	defer fi.Close()

	reader := bufio.NewReader(fi)
	lines := 0
	for {
		_, err = reader.ReadString('\n')
		if err != nil {
			break
		}
		lines++
	}

	num := rand.Intn(lines)
	mett := ""

	_, err = fi.Seek(0, 0)
	if err != nil {
		log.Println(err)
	}
	for ; num >= 0; num-- {
		mett, err = reader.ReadString('\n')
		if err == io.EOF {
			log.Println("PostMett: reached EOF")
			return "Error"
		}
		if err != nil {
			log.Println(err)
			bot.Notice(channel, "Failed to read from mett database")
			return "Error"
		}
	}
	return mett
}

func (bot *Mettbot) diffTopic(oldTopic, newTopic string) string {
	oldFile, err := ioutil.TempFile("", ".mettbotWdiffOld")
	if err != nil {
		log.Println(err)
		return ""
	}

	n, err := oldFile.WriteString(oldTopic)
	if n != len(oldTopic) || err != nil {
		log.Println(err)
		return ""
	}
	oldFile.Close()

	newFile, err := ioutil.TempFile("", ".mettbotWdiffNew")
	if err != nil {
		log.Println(err)
		return ""
	}

	n, err = newFile.WriteString(newTopic)
	if n != len(newTopic) || err != nil {
		log.Println(err)
		return ""
	}

	newFile.Close()
	defer func() {
		os.Remove(oldFile.Name())
		os.Remove(newFile.Name())
	}()

	db := "❣" // DeletionBegin
	de := "❢" // DeletionEnd
	ib := "¶" // InsertionBegin
	ie := "⁋" // InsertionEnd

	for {
		rdb := rune(rand.Intn(255-32) + 32)
		rde := rune(rand.Intn(255-32) + 32)
		rib := rune(rand.Intn(255-32) + 32)
		rie := rune(rand.Intn(255-32) + 32)

		contains := strings.ContainsRune(oldTopic, rdb)
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
	}

	coloring := map[string]string{ // http://oreilly.com/pub/h/1953
		db: "\x035\x1F\x02",
		de: "\x0F\x0315",
		ib: "\x033\x02",
		ie: "\x0F\x0315",
	}

	cmd := exec.Command("wdiff", "-w"+db, "-x"+de, "-y"+ib, "-z"+ie, oldFile.Name(), newFile.Name())
	out, _ := cmd.Output()
	outStr := string(out)

	for n, v := range coloring {
		outStr = strings.Replace(outStr, n, v, -1)
	}

	return outStr
}

func (bot *Mettbot) GetTweet(channel, url string) {
	regex, err := regexp.Compile(*Twitterregex)
	if err != nil {
		log.Println(err)
		return
	}
	sub := regex.FindStringSubmatch(url)
	tweetUrl := fmt.Sprintf("https://api.twitter.com/1/statuses/show.json?id=%v&trim_user=false&include_entities=true", sub[2])

	resp, err := http.Get(tweetUrl)
	if err != nil {
		log.Println(err)
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	type linkurl struct {
		Url          string
		Expanded_url string
	}
	type entity struct {
		Urls []linkurl
	}
	type usr struct {
		Screen_name string
	}
	type tweet struct {
		Text     string
		User     usr
		Entities entity
	}

	var twt tweet
	err = json.Unmarshal(body, &twt)
	if err != nil {
		log.Println(err)
		return
	}

	tweetText := twt.Text
	for _, u := range twt.Entities.Urls {
		tweetText = strings.Replace(tweetText, u.Url, u.Expanded_url, -1)
	}
	tweetText = html.UnescapeString(tweetText)

	tweetText = strings.Map(func(r rune) rune {
		if r < ' ' {
			return ' '
		}
		return r
	}, tweetText)

	bot.Notice(channel, "@"+twt.User.Screen_name+": "+tweetText)
}

func (bot *Mettbot) firebird(channel string) {
	time.Sleep(time.Duration(rand.Intn(3)+3) * time.Second)
	bot.Notice(channel, "\x0313"+a.RandStr(a.Firebird))
}

func (bot *Mettbot) DongDong(channel, msg string) {
	count := strings.Count(msg, "\\a")
	str := a.RandStr(a.Dong)
	notice := ""

	for i := 0; i < count; i++ {
		notice += str + " "
	}

	if len(notice) > 400 {
		notice = notice[:400]
	}
	bot.Notice(channel, notice)
}

func (bot *Mettbot) WriteQuote(filename string, prnt <-chan string, linesPrnt chan<- int) {
	for message := range prnt {
		fo, err := os.OpenFile(filename, syscall.O_RDWR+syscall.O_CREAT, 0644)
		if err != nil {
			log.Println(err)
			bot.Notice(*Channel, "Couldn't open quote database")
			continue
		}
		defer fo.Close()

		foReader := bufio.NewReader(fo)
		lines := 0 //last quote has no newline

		for {
			_, err = foReader.ReadString('\n')
			if err != nil {
				break
			}
			lines++
		}
		linesPrnt <- lines

		_, err = fo.WriteString(message)
		if err != nil {
			log.Println(err)
			bot.Notice(*Channel, "Couldn't write to quote database")
		}
		fo.Close()
	}
}

func (bot *Mettbot) ReadStdin() {
	con := bufio.NewReader(os.Stdin)
	for {
		s, err := con.ReadString('\n')
		if err != nil {
			// wha?, maybe ctrl-D...
			close(bot.Input)
			break
		}
		// no point in sending empty lines down the channel
		if len(s) > 2 {
			bot.Input <- s[0 : len(s)-1]
		}
	}
}

func (bot *Mettbot) ParseStdin() {
	for cmd := range bot.Input {
		if cmd[0] == ':' {
			idx := strings.Index(cmd, " ")
			msg := cmd[idx+1:]

			switch {
			case cmd[1] == 'd':
				fmt.Printf(bot.String())
			case cmd[1] == 'f':
				if len(cmd) > 2 && cmd[2] == 'e' {
					// enable flooding
					bot.Flood = true
				} else if len(cmd) > 2 && cmd[2] == 'd' {
					// disable flooding
					bot.Flood = false
				}
			case idx == -1:
				continue
			case cmd[1] == 'q':
				bot.ReallyQuit = true
				bot.Quit(msg)
			case cmd[1] == 'j':
				bot.Join(msg)
			case cmd[1] == 'p':
				bot.Part(msg)
			case cmd[1] == 'm':
				bot.Privmsg(*Channel, msg)
			case cmd[1] == 'a':
				bot.Action(*Channel, msg)
			case cmd[1] == 'n':
				bot.Notice(*Channel, msg)
			case cmd[1] == 's':
				midx := strings.Index(msg, " ")
				if midx == -1 {
					fmt.Println("Wrong Syntax")
					continue
				}
				val := msg[midx+1:]
				switch msg[:midx] {
				case "channel":
					bot.Join(val)
					*Channel = val
				case "nick":
					bot.Nick(val)
					*Nick = val
				case "quotes":
					*Quotes = val
				case "metts":
					*Metts = val
				case "offtime":
					num, err := strconv.Atoi(val)
					if err != nil {
						fmt.Println("No Number")
						continue
					}
					*Offtime = num
				case "offmessages":
					num, err := strconv.Atoi(val)
					if err != nil {
						fmt.Println("No Number")
						continue
					}
					*Offmessages = num
				case "probability":
					num, err := strconv.ParseFloat(val, 64)
					if err != nil {
						fmt.Println("No Flaot")
						continue
					}
					*Probability = num
				case "firebird":
					num, err := strconv.ParseFloat(val, 64)
					if err != nil {
						fmt.Println("No Flaot")
						continue
					}
					*Firebird = num
				case "randomanswer":
					num, err := strconv.ParseFloat(val, 64)
					if err != nil {
						fmt.Println("No Flaot")
						continue
					}
					*Randomanswer = num
				default:
					fmt.Println("Unknown variable")
				}
			}
		} else {
			bot.Raw(cmd)
		}
	}
}

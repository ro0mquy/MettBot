package plugins

import (
	"../ircclient"
	"bufio"
	"fmt"
	"github.com/willf/bloom"
	"log"
	"math"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	expected_urls        = 20000
	false_positives_rate = 0.001
	url_regex            = `(https?|ftp)://[-A-Za-z0-9+&@#/%?=~_|!:,.;]*[-A-Za-z0-9+&@#/%=~_|]`
)

type AltPlugin struct {
	sync.RWMutex
	ic    *ircclient.IRCClient
	bf    *bloom.BloomFilter
	regex *regexp.Regexp
}

func (q *AltPlugin) String() string {
	return "altbot"
}

func (q *AltPlugin) Info() string {
	return "altbot that indicates if a link was already posted"
}

func (q *AltPlugin) Usage(cmd string) string {
	// plugin has no commands
	return ""
}

func (q *AltPlugin) Register(ic *ircclient.IRCClient) {
	q.ic = ic
	q.bf = bloom.NewWithEstimates(expected_urls, false_positives_rate)
	q.regex = regexp.MustCompile(url_regex)
	q.fillFilter()
}

func (q *AltPlugin) Unregister() {
	// nothing to do here
}

func (q *AltPlugin) ProcessLine(msg *ircclient.IRCMessage) {
	if msg.Command != "PRIVMSG" {
		return
	}

	urls := q.regex.FindAllString(msg.Args[0], -1)
	for _, url := range urls {
		numsPosted, firstPosted := q.testAndAdd(url)
		if numsPosted == 0 || firstPosted.IsZero() {
			continue
		}

		// number of months between now and the first post
		// number should be between 2 and 12
		months := time.Since(firstPosted).Hours() / 24 / 30
		numberOfAs := int(math.Min(math.Max(2, months), 12))
		numberOfBangs := int(numsPosted)

		aaalt := strings.Repeat("a", numberOfAs) + "lt" + strings.Repeat("!", numberOfBangs)
		q.ic.ReplyMsg(msg, aaalt)
	}
}

func (q *AltPlugin) ProcessCommand(cmd *ircclient.IRCCommand) {
	// interface saturation
	return
}

// return a channel on which all lines of file <file> are send
func (q *AltPlugin) lines(file string) <-chan string {
	strChan := make(chan string)
	go func(sc chan<- string) {
		// close chan
		defer close(sc)

		q.RLock()
		defer q.RUnlock()

		f, err := os.Open(file)
		if err != nil {
			log.Println(err)
			return
		}
		defer f.Close()

		// new scanner that splits at line endings
		fScanner := bufio.NewScanner(f)
		for fScanner.Scan() {
			sc <- fScanner.Text()
		}

		err = fScanner.Err()
		if err != nil {
			log.Println(err)
			return
		}
	}(strChan)
	return strChan
}

// writes the line <line> to the file <file>
// returns an error if any occurs
func (q *AltPlugin) writeLine(file, line string) (err error) {
	q.Lock()
	defer q.Unlock()

	f, err := os.OpenFile(file, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	_, err = f.WriteString(line + "\n")
	if err != nil {
		return
	}

	return nil
}

// adds the time <t> plus the string <url> to the url-file
func (q *AltPlugin) addToFile(url string, t time.Time) {
	line := fmt.Sprintf("%v %v", t.Format(q.ic.GetStringOption("Altbot", "timeformat")), url)
	err := q.writeLine(q.ic.GetStringOption("Altbot", "file"), line)
	if err != nil {
		log.Println(err)
	}
}

// fills the bloom filter with the entries from the url-file
func (q *AltPlugin) fillFilter() {
	linesChan := q.lines(q.ic.GetStringOption("Altbot", "file"))
	for line := range linesChan {
		// url is second substring
		subs := strings.SplitN(line, " ", 2)
		if len(subs) < 2 {
			continue
		}
		q.bf.Add([]byte(subs[1]))
	}
}

// tests for the <url>, adds it to the file
// returns the number of times it was posted before and when the first post happend if it was already posted
func (q *AltPlugin) testAndAdd(url string) (numsPosted uint, firstPosted time.Time) {
	contained := q.bf.TestAndAdd([]byte(url))
	if contained {
		numsPosted, firstPosted = q.getNumAndDate(url)
	}
	q.addToFile(url, time.Now())
	return
}

// returns the number of times <url> occurs and when the first occurence happend
// if <url> is not in the file, returns 0, nil
func (q *AltPlugin) getNumAndDate(url string) (numsPosted uint, firstPosted time.Time) {
	linesChan := q.lines(q.ic.GetStringOption("Altbot", "file"))
	for line := range linesChan {
		// url is second substring and date is first
		subs := strings.SplitN(line, " ", 2)
		if len(subs) < 2 {
			continue
		}

		if subs[1] != url {
			continue
		}
		numsPosted++

		date, err := time.ParseInLocation(q.ic.GetStringOption("Altbot", "timeformat"), subs[0], time.Local)
		if err != nil {
			log.Println(err)
			continue
		}

		if date.Before(firstPosted) || firstPosted.IsZero() {
			firstPosted = date
		}
	}
	return
}

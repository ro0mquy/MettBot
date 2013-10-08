package plugins

import (
	"../answers"
	"../ircclient"
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	default_time_format = "2006-01-02T15:04"
)

type QuoteDBPlugin struct {
	sync.RWMutex
	ic *ircclient.IRCClient
}

func (q *QuoteDBPlugin) String() string {
	return "quotedb"
}

func (q *QuoteDBPlugin) Info() string {
	return "collects and displays quotes from users"
}

func (q *QuoteDBPlugin) Usage(cmd string) string {
	switch cmd {
	case "quote":
		return "quote <arg>: if arg is an integer, return quote number <arg>, else if <arg> is empty return a random quote, else return a random quote from user <arg>"
	case "search":
		return "search <pattern>: search for <pattern> in quote database, interpret <pattern> as regex if it starts and ends with an '/'"
	case "add":
		return "add <quote>: adds the string <quote> appended to the current time to the database"
	}
	return ""
}

func (q *QuoteDBPlugin) Register(cl *ircclient.IRCClient) {
	q.ic = cl

	if q.ic.GetStringOption("QuoteDB", "timeformat") == "" {
		log.Println("added default timeformat value of \"" + default_time_format + "\" to config file")
		q.ic.SetStringOption("QuoteDB", "timeformat", default_time_format)
	}

	q.ic.RegisterCommandHandler("quote", 0, 0, q)
	q.ic.RegisterCommandHandler("search", 1, 0, q)
	q.ic.RegisterCommandHandler("add", 1, 0, q)
}

func (q *QuoteDBPlugin) Unregister() {
	return
}

func (q *QuoteDBPlugin) ProcessLine(msg *ircclient.IRCMessage) {
	return
}

func (q *QuoteDBPlugin) ProcessCommand(cmd *ircclient.IRCCommand) {
	switch cmd.Command {
	case "quote":
		var out string
		if len(cmd.Args) == 0 {
			// no argument, get random quote
			out = q.getRandomQuote("")
			if out == "" {
				out = "No quotes in database"
			}
		} else {
			num64, err := strconv.ParseUint(cmd.Args[0], 10, 0)
			if err != nil {
				// argument is no integer, so it must be a string specifing a user
				out = q.getRandomQuote(cmd.Args[0])
				if out == "" {
					out = "No quotes from user " + cmd.Args[0]
				}
			} else {
				// argument is integer
				out = q.getQuote(uint(num64))
				if out == "" {
					out = "Quote not found"
				}
			}
		}
		q.ic.Reply(cmd, out)
	case "search":
		results := q.searchQuotes(strings.Join(cmd.Args, " "))
		if len(results) == 0 {
			q.ic.Reply(cmd, "Didn't find any matching quotes")
		}

		for _, quote := range results {
			q.ic.Reply(cmd, quote)
		}
	case "add":
		num := q.writeQuote(strings.Join(cmd.Args, " "), time.Now())
		out := fmt.Sprintf(answers.RandStr("addedQuote"), num)
		q.ic.Reply(cmd, out)
	}
}

// return a channel on which all lines of file <file> are send
func (q *QuoteDBPlugin) lines(file string) <-chan string {
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

// get the line <lineno> from file <file>
// if the line dosn't exist or a error occurs an empty string is returned
func (q *QuoteDBPlugin) getLine(file string, lineno uint) string {
	strChan := q.lines(file)
	var i uint = 0
	var out string
	for line := range strChan {
		if i == lineno {
			out = line
		}
		i++
	}
	return out
}

// returns the quote number <num>
// if the quote doesn't exist, returns an empty string
func (q *QuoteDBPlugin) getQuote(num uint) string {
	return q.getLine(q.ic.GetStringOption("QuoteDB", "file"), num)
}

// returns a random quote from user <user>
// if <user> is an empty string, returns a random quote from every user
// if there are no quotes from this user, returns an empty string
func (q *QuoteDBPlugin) getRandomQuote(user string) string {
	strChan := q.lines(q.ic.GetStringOption("QuoteDB", "file"))
	var i uint = 0
	quotes := make([]uint, 0)

	for line := range strChan {
		subs := strings.SplitN(line, " ", 3)
		if len(subs) >= 2 {
			// second substring is <user>
			if strings.Contains(subs[1], user) {
				quotes = append(quotes, i)
			}
		}
		i++
	}

	if len(quotes) == 0 {
		// no quotes found
		return ""
	}

	randQuote := quotes[rand.Intn(len(quotes))]
	out := q.getQuote(randQuote)
	return out
}

// searches for quotes and returns them as a string slice
// if <pattern> begins and ends with an '/', interprets it as a regular expression
func (q *QuoteDBPlugin) searchQuotes(pattern string) (results []string) {
	results = make([]string, 0)
	var regex *regexp.Regexp
	if len(pattern) > 1 && strings.HasPrefix(pattern, "/") && strings.HasSuffix(pattern, "/") {
		var err error
		regex, err = regexp.Compile(pattern[1 : len(pattern)-1]) // trim '/'s
		if err != nil {
			results = append(results, "Couldn't search: "+err.Error())
			return
		}
	}

	strChan := q.lines(q.ic.GetStringOption("QuoteDB", "file"))
	var i int = 0
	for line := range strChan {
		// if specified, match the regular expression or just check if <line> contains <pattern>
		if (regex != nil && regex.MatchString(line)) || strings.Contains(line, pattern) {
			results = append(results, strconv.Itoa(i)+" "+line)
		}
		i++
	}
	return results
}

// writes the line <line> to the file <file>
// returns an error if any occurs
func (q *QuoteDBPlugin) writeLine(file, line string) (err error) {
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

// adds a line to the database consisting of the time <t> and the string <quote>
// returns which number the new quote is
func (q *QuoteDBPlugin) writeQuote(quote string, t time.Time) uint {
	line := fmt.Sprintf("%v %v", t.Format(q.ic.GetStringOption("QuoteDB", "timeformat")), quote)
	err := q.writeLine(q.ic.GetStringOption("QuoteDB", "file"), line)
	if err != nil {
		log.Println(err)
	}

	strChan := q.lines(q.ic.GetStringOption("QuoteDB", "file"))
	var i uint = 0
	for _ = range strChan {
		i++
	}
	return i - 1 // quotes numbering starts at 0
}

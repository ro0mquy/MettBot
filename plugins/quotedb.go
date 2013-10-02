package plugins

import (
	"strings"
	"strconv"
	"math/rand"
	"../ircclient"
	"sync"
	"os"
	"bufio"
	"log"
)

type QuoteDBPlugin struct {
	sync.Mutex
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
	}
	return ""
}

func (q *QuoteDBPlugin) Register(cl *ircclient.IRCClient) {
	q.ic = cl
	q.ic.RegisterCommandHandler("quote", 0, 0, q)
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
	}
}

// return a channel on which all lines of file <file> are send
func (q *QuoteDBPlugin) lines(file string) (<-chan string) {
	strChan := make(chan string)
	go func(sc chan<- string) {
		// close chan
		defer close(sc)

		q.Lock()
		defer q.Unlock()

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
	} (strChan)
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
// if there are no quotes from this user, return an empty string
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
	out := q.getLine(q.ic.GetStringOption("QuoteDB", "file"), randQuote)
	return out
}

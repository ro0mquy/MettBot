package plugins

import (
	"../answers"
	"../ircclient"
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync"
)


type MettDBPlugin struct {
	sync.RWMutex
	ic *ircclient.IRCClient
}

func (q *MettDBPlugin) String() string {
	return "mettdb"
}

func (q *MettDBPlugin) Info() string {
	return "collects and displays mett related stuff"
}

func (q *MettDBPlugin) Usage(cmd string) string {
	switch cmd {
	case "mett":
		return "mett <content>: if <content> is specified, add to the database, if invoked without arguments, display random mett content"
	}
	return ""
}

func (q *MettDBPlugin) Register(cl *ircclient.IRCClient) {
	q.ic = cl
	q.ic.RegisterCommandHandler("mett", 0, 0, q)
}

func (q *MettDBPlugin) Unregister() {
	return
}

func (q *MettDBPlugin) ProcessLine(msg *ircclient.IRCMessage) {
	return
}

func (q *MettDBPlugin) ProcessCommand(cmd *ircclient.IRCCommand) {
	switch cmd.Command {
	case "mett":
		var out string
		if len(cmd.Args) == 0 {
			// no argument, get random mett
			out = q.getRandomMett()
			if out == "" {
				out = "No metts in database"
			}
		} else {
			// add line of mett
			num := q.writeMett(strings.Join(cmd.Args, " "))
			out = fmt.Sprintf(answers.RandStr("addedMett"), num)
		}
		q.ic.Reply(cmd, out)
	}
}

// return a channel on which all lines of file <file> are send
func (q *MettDBPlugin) lines(file string) <-chan string {
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
func (q *MettDBPlugin) getLine(file string, lineno uint) string {
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

// returns the mett number <num>
// if the quote doesn't exist, returns an empty string
func (q *MettDBPlugin) getMett(num uint) string {
	return q.getLine(q.ic.GetStringOption("MettDB", "file"), num)
}

// returns a random mett
// if there are no metts, returns an empty string
func (q *MettDBPlugin) getRandomMett() string {
	strChan := q.lines(q.ic.GetStringOption("MettDB", "file"))
	var i int = 0
	for _ = range strChan {
		i++
	}

	return q.getMett(uint(rand.Intn(i)))
}

// writes the line <line> to the file <file>
// returns an error if any occurs
func (q *MettDBPlugin) writeLine(file, line string) (err error) {
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

// adds the line <line> to the database
// returns which number the new mett is
func (q *MettDBPlugin) writeMett(line string) uint {
	err := q.writeLine(q.ic.GetStringOption("MettDB", "file"), line)
	if err != nil {
		log.Println(err)
	}

	strChan := q.lines(q.ic.GetStringOption("MettDB", "file"))
	var i uint = 0
	for _ = range strChan {
		i++
	}
	return i - 1 // quotes numbering starts at 0
}

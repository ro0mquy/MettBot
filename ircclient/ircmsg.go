package ircclient

import (
	"strings"
	"log"
)

type IRCMessage struct {
	Source   string
	Target   string
	Command  string
	Args     []string
	Complete string
}

type IRCCommand struct {
	Source  string
	Command string
	Args    []string
}

func ParseCommand(msg *IRCMessage, trigger byte) *IRCCommand {
	if msg.Command != "PRIVMSG" || len(msg.Args) == 0 || msg.Args[0][0] != trigger {
		return nil
	}
	toParse := msg.Args[0]
	ret := &IRCCommand{"", "", make([]string, 0)}
	for i, last, matchP := 1, 0, false; i < len(toParse); i++ {
		// Match "longer strings in parenthesis"
		if toParse[i] == '"' && (toParse[i-1] == ' ' || toParse[i-1] == '\t' || matchP) {
			if !matchP {
				last = i
				matchP = true
			} else if toParse[i-1] == '\\' {
				// escaped parenthesis
			} else {
				ret.Args = append(ret.Args, toParse[last+1:i])
				matchP = false
				// Skip whitespace
				i++
				for i < len(toParse) && (toParse[i] == ' ' || toParse[i] == '\t') {
					i++
				}
				last = i
			}
			// Match texts seperated by whitespace
		} else if !matchP && (toParse[i] == ' ' || toParse[i] == '\t') {
			log.Println(toParse[last:i])
			ret.Args = append(ret.Args, toParse[last:i])
			for i < len(toParse) && (toParse[i] == ' ' || toParse[i] == '\t') {
				i++
			}
			last = i
			i--
		}
	}
	ret.Command = ret.Args[0][1 : len(ret.Args[0])] // Strip off trigger
	ret.Source = msg.Source
	ret.Args = ret.Args[1 : len(ret.Args)]
	return ret
}

func ParseServerLine(line string) *IRCMessage {
	im := &IRCMessage{"", "", "", make([]string, 0), line}

	if len(line) == 0 || strings.Trim(line, " \t\n\r") == "" {
		log.Println("ParseIrcLine: empty line")
		return im
	}

	// source and target
	if line[0] == ':' {
		parts := strings.SplitN(line[1:], " ", 4) // 4: src cmd target rest
		im.Source = parts[0]
		im.Command = parts[1]
		im.Target = parts[2]
		// cut them off
		if len(parts) > 3 {
			line = parts[3]
		} else {
			line = ""
		}
	} else {
		parts := strings.SplitN(line, " ", 2) // cmd, rest
		im.Command = parts[0]
		if len(parts) > 1 {
			line = parts[1]
		} else {
			line = ""
		}
	}

	args := strings.SplitN(line, ":", 2)
	for _, a := range strings.Split(args[0], " ") {
		if a != "" {
			im.Args = append(im.Args, a)
		}
	}
	if len(args) > 1 {
		im.Args = append(im.Args, args[1])
	}

	//log.Printf("im: %#v\n", im)
	return im
}

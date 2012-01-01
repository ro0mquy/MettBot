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
	Target  string
	Args    []string
}

func ParseCommand(msg *IRCMessage) *IRCCommand {
	var lastByte byte = ' '

	if len(msg.Args) == 0 {
		return nil
	}

	toParse := msg.Args[0]
	ret := &IRCCommand{msg.Source, "", msg.Target, make([]string, 0)}
	for i, last, matchP := 0, 0, false; i < len(toParse); i++ {
		//log.Printf("Now at: %c\n", toParse[i])

		if toParse[i] == '"' && (lastByte == ' ' || lastByte == '\t' || matchP) {
			// Match "longer strings in quotes"
			if !matchP {
				// Opening quotes found
				last = i + 1
				matchP = true
			} else if lastByte == '\\' {
				// escaped quotes. Ignore it
			} else {
				// Closing quotes found
				ret.Args = append(ret.Args, toParse[last:i])
				matchP = false
				// New token starts after quote
				last = i + 1
				lastByte = ' '
				continue
			}
		} else if !matchP && (toParse[i] == ' ' || toParse[i] == '\t') {
			// Match texts seperated by whitespace
			// This also skips over consecutive whitespace
			if lastByte != ' ' && lastByte != '\t' {
				// Whitespace after word, add word
				ret.Args = append(ret.Args, toParse[last:i])
			}
			last = i + 1
		} else if i == len(toParse)-1 && toParse[i] != ' ' && toParse[i] != '\t' {
			// Last token in string. Add it.
			ret.Args = append(ret.Args, toParse[last:i+1])
			break
		}
		lastByte = toParse[i]
	}
	//log.Printf("%#v\n", ret.Args)
	if len(ret.Args) > 0 {
		ret.Command = ret.Args[0]
		ret.Args = ret.Args[1:len(ret.Args)]
	}
	if len(ret.Args) > 1 {
		ret.Args = ret.Args[1:len(ret.Args)]
	}
	return ret
}

func ParseServerLine(line string) *IRCMessage {
	im := &IRCMessage{"", "", "", make([]string, 0), line}

	if len(line) == 0 || strings.Trim(line, " \t\n\r") == "" {
		return nil
	}

	// source and target
	if line[0] == ':' {
		parts := strings.SplitN(line[1:], " ", 4) // 4: src cmd target rest
		line = "" // if there is something left of the line, it will
				  // be added in case 4:
		switch len(parts) {
		case 4:
			line = parts[3]
			fallthrough
		case 3:
			im.Target = parts[2]
			fallthrough
		case 2:
			im.Command = parts[1]
			fallthrough
		case 1:
			im.Source = parts[0]
		default:
			log.Printf("len(parts)=%d, ignoring", len(parts))
			return nil
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

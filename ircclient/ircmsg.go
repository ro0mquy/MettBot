package ircclient

import (
	"strings"
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
	} else if len(ret.Args) > 1 {
		ret.Args = ret.Args[1:len(ret.Args)]
	}
	return ret
}

func ParseServerLine(line string) *IRCMessage {
	im := &IRCMessage{"", "", "", make([]string, 0), line}

	if len(line) == 0 || strings.Trim(line, " \t\n\r") == "" {
		return nil
	}

	// Omit : at beginning of line
	if line[0] == ':' {
		line = line[1:]
	}
	// Split
	var parts []string = make([]string, 0)
	for {
		if line[0] == ':' {
			line = line[1:]	// Strip the :
			parts = append(parts, line)
			break
		}
		split := strings.SplitN(line, " ", 2)
		parts = append(parts, split[0])
		if len(split) == 2 && len(split[1]) > 0 {
			line = split[1]
		} else {
			break
		}
	}

	if len(parts) <= 2 {
		im.Command = parts[0]
		if len(parts) == 2 {
			im.Args = []string{parts[1]}
		}
	} else {
		im.Source = parts[0]
		im.Command = parts[1]
		im.Target = parts[2]
		if len(parts) >= 4 {
			parts = parts[3:]
			im.Args = parts
		}
	}

	//log.Printf("im: %#v\n", im)
	return im
}

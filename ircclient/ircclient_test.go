package ircclient

import (
	"testing"
)

var server_lines = []string{
	":fu-berlin.de 020 * :Please wait while we process your connection.",
	":fu-berlin.de 001 osntauohe :Welcome to the Internet Relay Network osntauohe!~osntauohe@176.99.114.122",
	":fu-berlin.de 042 osntauohe 276BAY2UY :your unique ID",
	":fu-berlin.de 375 osntauohe :- fu-berlin.de Message of the Day - ",
	":fu-berlin.de 372 osntauohe :- Willkommen auf dem IRCnet-Server der Freien Universitaet Berlin, ZEDAT",
	":fu-berlin.de 376 osntauohe :End of MOTD command.",
}

var parsed_structs = []IRCMessage{
	{"fu-berlin.de", "*", "020", []string{"Please wait while we process your connection."}, server_lines[0]},
	{"fu-berlin.de", "osntauohe", "001", []string{"Welcome to the Internet Relay Network osntauohe!~osntauohe@176.99.114.122"}, server_lines[1]},
	{"fu-berlin.de", "osntauohe", "042", []string{"276BAY2UY", "your unique ID"}, server_lines[2]},
	{"fu-berlin.de", "osntauohe", "375", []string{"- fu-berlin.de Message of the Day - "}, server_lines[3]},
	{"fu-berlin.de", "osntauohe", "372", []string{"- Willkommen auf dem IRCnet-Server der Freien Universitaet Berlin, ZEDAT"}, server_lines[4]},
	{"fu-berlin.de", "osntauohe", "376", []string{"End of MOTD command."}, server_lines[5]},
}

func ircMessage_deep_equals(m1, m2 *IRCMessage) bool {
	return m1.Source == m2.Source &&
		m1.Target == m2.Target &&
		m1.Command == m2.Command &&
		string_array_deep_equals(m1.Args, m2.Args) &&
		m1.Complete == m2.Complete
}

func string_array_deep_equals(a, b []string) bool {
	for i, _ := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestParseServerLine(t *testing.T) {
	for i, line := range server_lines {
		test_parsed := ParseServerLine(line)
		if !ircMessage_deep_equals(test_parsed, &parsed_structs[i]) {
			t.Logf("input \"%s\" fails, wrong result is \"%#v\", should be \"%#v\"", line, test_parsed, parsed_structs[i])
			t.Fail()
		}
	}
}

//
//func main() {
//	fmt.Println("== ircmsg::ParseServerLine() ==")
//	for _, line := range server_lines {
//		fmt.Printf("%s => %#v\n", line, ircclient.ParseServerLine(line))
//	}
//
//	fmt.Println("== irccmd::ParseCommand() ==")
//	args := []string{"!das hier ist ein  \"Test! \\\"für\" das  "}
//	msg := &ircclient.IRCMessage{"", "", "PRIVMSG", args, ""}
//	ret := ircclient.ParseCommand(msg, '!')
//	fmt.Printf("%#v", ret.Args)
//}

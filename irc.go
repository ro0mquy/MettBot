package main

import (
    "bufio";
    "exec";
    "fmt";
    "net";
    "os";
    "regexp";
    "strconv";
    "strings";
    "syscall";
    "time";
)

type EventChecker func(evt *IRCEvent) bool
type IRCListener func(evt *IRCEvent)

type IRCConn struct {
    sock *net.TCPConn;
    read, write chan string; //bidirectional
    Write chan<- string;
    checkers map[chan *IRCEvent]EventChecker;
    listeners map[IRCListener]EventChecker;
}

type IRCEvent struct {
    form int;
    message, target, secondTarget, source, fullsource, raw string;
    replycode int;
    time int64;
}

// ----------------------------------------- Maintaining an IRCConn
// This works on freenode. No guarantees elsewhere.
func NewConnection() *IRCConn {
    read := make(chan string, 1000);
    write := make(chan string, 1000);
    checkers := make(map[chan *IRCEvent]EventChecker);
    listeners := make(map[IRCListener]EventChecker);
    return &IRCConn{nil, read, write, write, checkers, listeners};
}
func DialIRC(server, nick string) (*IRCConn, os.Error) {
    con := NewConnection();
    err := con.Connect(server, nick);
    if err != nil {
        return nil, err;
    }
    return con, nil
}
func (con *IRCConn) Connect(server, nick string) os.Error {
    //no multiserver support: if a connection already exists, just complain
    if con.sock != nil {
        return os.NewError("Connection already established");
    }
    addr, err := net.ResolveTCPAddr(server);
    if err != nil { return err }
    sock, err := net.DialTCP("tcp", nil, addr);
    if err != nil { return err }
    rer := bufio.NewReader(sock);
    wer := bufio.NewWriter(sock);    
    write := make(chan string, 1000); //distinct from con.Write

    //check to make sure a session has been established. TODO PREFIX/CHANMODE/etc. parsing
    replychan := con.OneTimeChan(func(evt *IRCEvent) bool {
        return evt.form == REPLY && (evt.replycode == 4 || evt.replycode == 433); //RPL_MYINFO or ERR_NICKNAMEINUSE
    });

    go func() {
        for {
            str, err := rer.ReadString(byte('\n'));
            if err != nil { fmt.Printf("read: %s\n", err); break; }
            con.read <- str[0:len(str)-2];
        }
    }();
    go func() {
        for {
            str := <-write;
            if closed(write) { break; }
            fmt.Printf("<- %q\n", str);
            err := wer.WriteString(str + "\r\n");
            if err != nil { fmt.Printf("write: %s\n", err); break; }
            wer.Flush();
        }
    }();
    go con.dispatcher();

    write <- "NICK " + nick;
    write <- "USER bot * * :...";

    //this means the user can't attach listeners until appropriate reply code is received
    replyevt := <-replychan;
    if (replyevt.replycode == 433) {
        con.Close(); //a bit harsh
        return os.NewError("nick already in use");
    }
    for {
        if str, ok := <-con.write; ok {
            write <- str
        } else {
            break
        }
    }
    con.write = write;
    con.Write = write;
    con.sock = sock;
    return nil;
}

func (con *IRCConn) dispatcher() {
    read := con.read;
    for {
        str := <-read;
        if closed(read) {
            break;
        }
        evt := ParseMessage(str);
        //do some ugly checking
        if evt.form == PING {
            con.Write <- "PONG :" + evt.message;
        } else if evt.form == CTCP_VERSION {
            con.Write <- "NOTICE " + evt.source + " :\x01VERSION go-bot\x01"
        }
        for ch, check := range con.checkers {
            if closed(ch) {
                con.checkers[ch] = nil, false; //concurrent modification allowed
            } else if check == nil || check(evt) {
                if ch <- evt {}
            }
        }
        for listen, check := range con.listeners {
            if check == nil || check(evt) {
                go listen(evt);
            }
        }
    } //exiting!
    for ch, _ := range con.checkers {
        con.UnregisterChan(ch);
    }
    for listen, _ := range con.listeners {
        con.RemoveListener(listen);
    }
}

//an example utility function
func (con *IRCConn) Identify(password string) bool {
    con.Write <- "NICKSERV IDENTIFY " + password;
    evt := <-con.OneTimeoutChan(func(evt *IRCEvent) bool {
        return evt.form == PRIV_NOTICE && evt.source == "NickServ" && //these strings for freenode
               (strings.HasPrefix(evt.message, "You are already logged in") ||
                strings.HasPrefix(evt.message, "You are now identified") ||
                strings.HasPrefix(evt.message, "Invalid password"))
    }, 10*1000); //10 seconds in ms
    if evt == nil {
        return false;
    }
    return strings.HasPrefix(evt.message, "Invalid password");
}

//supposed to cleanly close everything down. however, this causes an RTS error
func (con *IRCConn) Close() {
    close(con.read);
    close(con.Write);
    con.sock.Close();
}

// ----------------------------------------- generate EventCheckers
//The event's form must have at least all the bits that forms has.
func NewFormChecker(forms int) EventChecker {
    return func(evt *IRCEvent) bool {
        return (evt.form & forms) == forms;
    }
}
func NewRegexChecker(reg regexp.Regexp) EventChecker {
    return func(evt *IRCEvent) bool {
        return len(evt.message) != 0 && reg.MatchString(evt.message);
    }
}
// ----------------------------------------- add IRCListeners
// Gor every applicable event, a goroutine is created that passes it along.
// They may run concurrently, and order is not guaranteed
func (con *IRCConn) AddListener(listen IRCListener, check EventChecker) {
    con.listeners[listen] = check;
}
func (con *IRCConn) AddEverythingListener(listen IRCListener) {
    con.listeners[listen] = nil;
}
func (con *IRCConn) RemoveListener(listen IRCListener) {
    con.listeners[listen] = nil, false;
}

// ----------------------------------------- get IRCEvent channels
// these use Go's chans to get IRCEvents in a coordinated way
func (con *IRCConn) NewEverythingChan() chan *IRCEvent {
    return con.NewBoundChan(nil, 500);
}
func (con *IRCConn) NewBoundChan(check EventChecker, capacity int) chan *IRCEvent {
    ch := make(chan *IRCEvent, capacity);
    con.checkers[ch] = check;
    return ch;
}
func (con *IRCConn) NewChan(check EventChecker) chan *IRCEvent {
    return con.NewBoundChan(check, 100);
}
func (con *IRCConn) UnregisterChan(ch chan *IRCEvent) {
    con.checkers[ch] = nil, false;
    close(ch);
}
func (con *IRCConn) OneTimeChan(check EventChecker) chan *IRCEvent {
    ch := con.NewBoundChan(check, 1);
    ch2 := make(chan *IRCEvent, 1);
    go func() {
        ch2 <- (<-ch);
        con.UnregisterChan(ch);
        close(ch2);
    }();
    return ch2;
}
func (con *IRCConn) OneTimeoutChan(check EventChecker, ms int64) chan *IRCEvent {
    ch := con.NewBoundChan(check, 1);
    ch2 := make(chan *IRCEvent, 1);
    go func() {
        ch2 <- (<-ch);
        con.UnregisterChan(ch);
        close(ch2);
    }();
    go func() {
        time.Sleep(1000000*ms);
        con.UnregisterChan(ch);
    }();
    return ch2;
}

// ----------------------------------------- parse IRCEvent

//parse words until we reach a message boundary.
//"hello world" -> (["hello", "world"], "")
//"to you :the message" -> (["to", "you"], "the message")
func wordsMessage(str string) (buf []string, result string) {
    buf = make([]string, 5);
    var start, end, pos int;
    result = "";
    leng := len(str);
    for {
        for start < leng && str[start] == ' ' { start++; }
        if start == leng { break; }
        if str[start] == ':' { result = str[start+1:len(str)]; break; }
        for end = start+1; end < leng && str[end] != ' '; { end++; }
        if pos >= len(buf) {
            buf2 := make([]string, len(buf)*2);
            for i, x := range buf { buf2[i] = x; }
            buf = buf2;
        }
        buf[pos] = str[start:end];
        pos++;
        start = end;
    }
    buf = buf[0:pos];
    return;
}
//this will be replaced with CHANTYPES parsing
func isChannel(str string) bool {
    return len(str) != 0 && strings.Index("#", str[0:1]) != -1;
}

func ParseMessage(line string) *IRCEvent {
    evt := new(IRCEvent);
    evt.time = time.Seconds();
    evt.raw = line; //raw
    evt.form = UNKNOWN;
    if (len(line) == 0) { return evt; }
    tokens, msg := wordsMessage(line[1:len(line)]);
    if len(tokens) >= 2 && line[0] == ':' {
        evt.message = "";
        prefix := tokens[0];
        evt.fullsource = prefix;
        if exclam := strings.Index(prefix, "!"); exclam != -1 {
            prefix = prefix[0:exclam];
        }
        evt.source = prefix;

        command := tokens[1];
        code, err := strconv.Atoi(command);
        var third, fourth string;
        if (len(tokens) > 2) { third = tokens[2] }
        if (len(tokens) > 3) { fourth = tokens[3] }

        if (err == nil) {
            evt.form = REPLY;
            evt.replycode = code;
            //FIXME: info can occur before :, and this doesn't include it
            evt.message = msg;
            evt.target = third;
        } else { switch command {
            case "INVITE":
                evt.form = INVITE;
                evt.target = third;
                evt.secondTarget = fourth;
            case "KICK":
                evt.form = KICK;
                evt.message = msg;
                evt.target = fourth;
                evt.secondTarget = third;
            case "NICK":
                evt.form = NICK;
                evt.target = third;
            case "QUIT":
                evt.form = QUIT;
                evt.message = msg;
            case "JOIN":
                evt.form = JOIN;
                evt.target = msg;
            case "PART":
                evt.form = PART;
                evt.message = msg;
                evt.target = third;
            case "MODE":
                evt.form = MODE;
                //FIXME: sample format ":host MODE #channel [msg]"
                evt.message = "???";
                evt.target = third;
            case "TOPIC":
                evt.form = TOPIC;
                evt.message = msg;
                evt.target = third;
            case "PRIVMSG":
                evt.target = third;
                leng := len(msg);
                 if leng > 2 && msg[0] == 1 && msg[leng - 1] == 1 {
                    msg = msg[1:leng-1];
                    action := "ACTION ";
                    if strings.HasPrefix(msg, action) {
                        msg = ":" + msg; //FIXME: msg[len(action):len(msg)]; <- this is causing SIGSEGV. why?
                        if isChannel(third) { evt.form = CHAN_ACTION } else { evt.form = PRIV_ACTION }
                    } else if strings.HasPrefix(msg, "DCC ") {
                        evt.form = DCC;
                    } else { switch msg {
                        case "PING": evt.form = CTCP_PING;
                        case "TIME": evt.form = CTCP_TIME;
                        case "VERSION": evt.form = CTCP_VERSION;
                        default: evt.form = CTCP_OTHER;
                    }}

                } else {
                    if isChannel(third) { evt.form = CHAN_MESSAGE } else { evt.form = PRIV_MESSAGE }
                }
                evt.message = msg;
            case "NOTICE":
                if isChannel(third) { evt.form = CHAN_NOTICE } else { evt.form = PRIV_NOTICE }
                evt.target = third;
                evt.message = msg;
        }}
    } else if strings.HasPrefix(line, "PING :") {
        evt.form = PING;
        evt.message = msg;
    }
    return evt;

}

const (
    NICK = 0;
    QUIT = 1 << iota;
    JOIN;
    PART;
    MODE;
    TOPIC;
    INVITE;
    KICK;
    REPLY;
    DCC;
    PING;
//ideally, have these as combinations of forms with CHANNEL/PRIVATE bits
    MESSAGE;
    NOTICE;
    ACTION;
    CHANNEL;
    PRIVATE;
    CTCP_PING;
    CTCP_TIME;
    CTCP_VERSION;
    CTCP_OTHER;
    UNKNOWN;
    CHAN_MESSAGE = MESSAGE | CHANNEL;
    PRIV_MESSAGE = MESSAGE | PRIVATE;
    CHAN_NOTICE = NOTICE | CHANNEL;
    PRIV_NOTICE = NOTICE | PRIVATE;
    CHAN_ACTION = ACTION | CHANNEL;
    PRIV_ACTION = ACTION | PRIVATE;
)


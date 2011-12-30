package ircclient

import (
	"net"
	"os"
	"log"
	"bufio"
)

type IRCConn struct {
	conn *net.TCPConn
	bio  *bufio.ReadWriter
	tmgr *ThrottleIrcu
	done chan bool

	Output chan string
	Input  chan string
}

func NewIRCConn() *IRCConn {
	return &IRCConn{done: make(chan bool), Output: make(chan string, 50), Input: make(chan string, 50), tmgr: new(ThrottleIrcu)}
}

func (ic *IRCConn) Connect(hostport string) os.Error {
	if len(hostport) == 0 {
		return os.NewError("empty server addr, not connecting")
	}
	if ic.conn != nil {
		log.Printf("warning: already connected to " + ic.conn.RemoteAddr().String())
	}
	c, err := net.Dial("tcp", hostport)
	if err != nil {
		return err
	}
	ic.conn, _ = c.(*net.TCPConn)
	ic.bio = bufio.NewReadWriter(bufio.NewReader(ic.conn), bufio.NewWriter(ic.conn))

	go func() {
		for {
			// TODO: err
			s, err := ic.bio.ReadString('\n')
			if err != nil {
				select {
				case d := <-ic.done:
					ic.done <- d
					return
				default:
					log.Printf("Can't read from input channel: " + err.String())
					return
				}
			}
			ic.Input <- s
			log.Print("<< " + s)
		}
	}()
	go func() {
		for {
			select {
			case s := <-ic.Output:
				ic.tmgr.WaitSend(s)
				log.Print(">> " + s)
				if _, err = ic.bio.WriteString(s); err != nil {
					log.Printf("Can't write to output channel: " + err.String())
					return
				}
				ic.bio.Flush()
			case d := <-ic.done:
				ic.done <- d
				if d {
					return
				}
			}
		}
	}()

	return nil
}

func (ic *IRCConn) Flush() {
	// TODO implement
	// Should block until all data to server has been sent
	// and bufio flushed
}

func (ic *IRCConn) Quit() {
	ic.conn.Close()
	ic.done <- true
}

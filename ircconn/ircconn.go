package ircconn

import (
	"net"
	"os"
	"log"
	"bufio"
)


type IRCConn struct {
	conn *net.TCPConn
	bio *bufio.ReadWriter
	done chan bool

	Output chan string
	Input chan string
}

func NewIRCConn() *IRCConn {
	return &IRCConn{done: make(chan bool), Output: make(chan string), Input: make(chan string)}
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
				log.Printf("Can't read from input channel: " + err.String())
				return
			}
			ic.Input <- s
		}
	}()
	go func() {
		for {
			s := <-ic.Output
			if _, err = ic.bio.WriteString(s); err != nil {
				log.Printf("Can't write to output channel: " + err.String())
				return
			}
			ic.bio.Flush()
		}
	}()

	return nil
}

func (ic IRCConn) Quit() {
	ic.done <- true
	ic.Close()
}

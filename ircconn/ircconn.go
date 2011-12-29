package ircconn

import (
	"net"
	"os"
	"log"
	"bytes"
	"multiplex"
)


type IRCConn struct {
	conn *net.TCPConn
	output_mux *multiplex.StringMuxManager
	output chan string
	input_demux *multiplex.StringDemuxManager
	input chan string
	done chan bool
}

func NewIRCConn() *IRCConn {
	return &IRCConn{done: make(chan bool)}
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

	client_to_server := make(chan string, 3)
	ic.input_demux= multiplex.NewStringDemuxManager(client_to_server)
	ic.input= client_to_server

	from_server := make(chan string, 3)
	ic.output_mux= multiplex.NewStringMuxManager(from_server)
	ic.output= from_server

	ic.readServer()
	ic.writeServer()

	return nil
}

// gibt dir nen channel, auf dem alle dinge rauskommen, die wir als verbindung kriegen
func (ic *IRCConn) RegisterListener() <-chan string {
	if ic == nil {
		log.Println("no connection, ic == nil")
		return nil
	}
	c := make(chan string, 1)
	ic.output_mux.Register(c)
	return c
}

// gibt dir nen channel, der in die serververbindung geforwarded wird
func (ic IRCConn) RegisterWriter() chan<- string {
	c := make(chan string, 1)
	ic.input_demux.Register(c)
	return c
}

// gibt dir nen channel, der eine kombination aus den beiden obigen ist
func (ic IRCConn) RegisterListenerWriter() chan string {
	c := make(chan string, 1)
	ic.input_demux.Register(c)
	ic.output_mux.Register(c)
	return c
}

func (ic *IRCConn) writeServer() {
	if ic.conn == nil {
		log.Println("not connected")
		return
	}
	go func() {
		for {
			s := <-ic.input
			ic.conn.Write([]byte(s))
		}
	}()
}


// liest aus conn, schreibt nach output
// das hier is quasi nur die "lese-von-server"-seite
// der verbindung
func (ic IRCConn) readServer() {
	if ic.conn == nil {
		log.Println("not connected")
		return
	}
	go func() {
		fd, err := ic.conn.File()
		if err != nil {
			log.Println(err.String())
			return
		}
		defer fd.Close()
		buf := bytes.NewBuffer(make([]byte, 100))
		_, err = buf.ReadFrom(fd)
		if err != nil {
			log.Println(err.String())
			return
		}
		ic.output <- buf.String()
	}()
}

func (ic IRCConn) Quit() {
	ic.done <- true
}


package ircclient

import (
	"bufio"
	"errors"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"
)

type ircConn struct {
	conn    *net.TCPConn
	bio     *bufio.ReadWriter
	tmgr    *throttleIrcu
	done    chan bool
	flushed chan bool

	Err    chan error
	Output chan string
	Input  chan string
}

func NewircConn() *ircConn {
	return &ircConn{done: make(chan bool, 1), flushed: make(chan bool), Output: make(chan string, 50), Input: make(chan string, 50), tmgr: new(throttleIrcu), Err: make(chan error, 5)}
}

func (ic *ircConn) Connect(hostport string) error {
	if len(os.Args) > 1 { // we're coming from kexec
		fd, err := strconv.Atoi(os.Args[1])
		if err != nil {
			log.Fatal("unable to parse argv[1]" + err.Error())
		}
		file := os.NewFile(uintptr(fd), "conn")
		conn, err := net.FileConn(file)
		if err != nil {
			log.Println("Connection fd is: " + strconv.Itoa(fd))
			log.Fatal("unable to recover conn: " + err.Error())
		}
		ic.conn, _ = conn.(*net.TCPConn)
	} else {
		if len(hostport) == 0 {
			return errors.New("empty server addr, not connecting")
		}
		if ic.conn != nil {
			log.Printf("warning: already connected")
		}
		c, err := net.Dial("tcp", hostport)
		if err != nil {
			return err
		}
		ic.conn, _ = c.(*net.TCPConn)
	}
	// from here on, we're on same behaviour again

	ic.bio = bufio.NewReadWriter(bufio.NewReader(ic.conn), bufio.NewWriter(ic.conn))

	go func() {
		// This goroutine is responsible for doing blocking reads on the input socket
		// and forwarding them to the application
		for {
			s, err := ic.bio.ReadString('\n')
			if err != nil {
				select {
				case d := <-ic.done:
					ic.done <- d
					return
				default:
					ic.Err <- errors.New("ircmessage: receive: " + err.Error())
					ic.Quit()
					return
				}
			}
			s = strings.Trim(s, "\r\n")
			ic.Input <- s
			//log.Println("<< " + s)
		}
	}()
	go func() {
		// This goroutine is responsible for sending the output waiting in channel
		// ic.Output to the server
		for {
			select {
			case s := <-ic.Output:
				s = s + "\r\n"
				ic.tmgr.WaitSend(s)
				//log.Print(">> " + s)
				if _, err := ic.bio.WriteString(s); err != nil {
					ic.Err <- errors.New("ircmessage: send: " + err.Error())
					log.Println("Send failed: " + err.Error())
					ic.Quit()
					return
				}
				ic.bio.Flush()
			case d := <-ic.done:
				// Connection is going to close, flush all data
				ic.done <- d
				for {
					select {
					case s := <-ic.Output:
						s = s + "\r\n"
						ic.tmgr.WaitSend(s)
						log.Print(">> " + s)
						// Do no more error handling here
						if _, err := ic.bio.WriteString(s); err != nil {
							ic.flushed <- true
							return
						}
						ic.bio.Flush()
					default:
						ic.flushed <- true
						// No more data to send
						return
					}
				}
			}
		}
	}()

	return nil
}

func (ic *ircConn) Quit() {
	ic.done <- true

	// Wait until all sends have completed
	select {
	case _ = <-ic.flushed:
	}

	close(ic.Input)
	ic.conn.Close()
	ic.Err <- errors.New("Connection closed by user")
}

// returns the socket of the ircConn or -1 if an error occurs
//
// the file descriptor returned by Conn.File() is a duplicate, with flag CloseOnExec set
// we have to unset the flag manually to successfully exec
func (ic *ircConn) GetSocket() int {
	// get a duplicate of the file descriptor
	file, err := ic.conn.File()
	if err != nil {
		log.Println("Unable to get socket fd:", err.Error())
		return -1
	}
	fd := int(file.Fd())

	// retrieve the current flags set on the file descriptor
	r0, _, e1 := syscall.Syscall(syscall.SYS_FCNTL, uintptr(fd), syscall.F_GETFD, 0)
	if e1 != 0 {
		log.Println("Unable to get file descriptor flags:", e1.Error())
		return -1
	}
	flags := int(r0)
	// unset FD_CLOEXEC (CloseOnExec)
	flags &= ^syscall.FD_CLOEXEC

	// set the old flags without CloseOnExec
	_, _, e1 = syscall.Syscall(syscall.SYS_FCNTL, uintptr(fd), syscall.F_SETFD, uintptr(flags))
	if e1 != 0 {
		log.Println("Unable to set file descriptor flags:", e1.Error())
		return -1
	}
	return fd
}

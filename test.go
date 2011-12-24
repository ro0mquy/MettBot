package main

import (
	"fmt"
	"bufio"
	"net"
	"os"
)

func main() {
	var sock *net.TCPConn
	if s, ret := net.Dial("tcp", "localhost:6667"); ret != nil {
		fmt.Println("Error: " + ret.String())
		return
	} else {
		sock, _ = s.(*net.TCPConn)
	}

	io := bufio.NewReadWriter(bufio.NewReader(sock),
		bufio.NewWriter(sock))
	for {
		var err os.Error
		var in string
		if in, err = io.ReadString('\n'); err != nil {
			fmt.Println("Error! " + err.String())
			return
		}
		if _, err = io.WriteString(in); err != nil {
			fmt.Println("Error! " + err.String())
			return
		}
		io.Flush()
		fmt.Print(in)
	}
}

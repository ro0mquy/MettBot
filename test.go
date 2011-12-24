package main

import (
	"fmt"
	"bufio"
	"net"
)

func main() {
	var sock net.Conn
	if s, ret := net.Dial("tcp", "localhost:6667"); ret != nil {
		fmt.Println("Error")
		return
	} else {
		sock = s
	}
	io := bufio.NewReadWriter(bufio.NewReader(sock),
		bufio.NewWriter(sock))
	io.WriteString("Hallo, Welt")
}

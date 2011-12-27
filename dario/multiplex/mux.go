package multiplex

import (
	"log"
)


type StringMuxManager struct {
	clients []chan string
	input chan string
	done_chan chan bool
}

func NewStringMuxManager(input chan string) *StringMuxManager {
	if input == nil {
		return nil
	}
	cl := make([]chan string, 0)
	done := make(chan bool)
	return &StringMuxManager{cl, input, done}
}

func (sm *StringMuxManager) Register(client chan string) {
	if(client == nil) {
		return
	}
	log.Println(sm)
	sm.clients= append(sm.clients, client)
}

func (sm *StringMuxManager) Work() {
	sm.done_chan= make(chan bool)
	go func(done chan bool) {
		select {
		case <- done:
			return
		case value, ok := <- sm.input:
			if ok {
				for _, c := range sm.clients {
					go func(cc chan string, vv string) {
						cc <- vv
					}(c, value)
				}
			}
		}
	}(sm.done_chan)
}

func (sm StringMuxManager) Send(data string) {
	// just throw data into input channel, Work() will take care of it
	sm.input <- data
}

func(sm StringMuxManager) Quit() {
	sm.done_chan <- true
}


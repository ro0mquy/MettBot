package multiplex

type StringDemuxManager struct {
	inputs []chan string
	output chan string
	done_chan chan bool
}

func NewStringDemuxManager(output chan string) *StringDemuxManager {
	if output == nil {
		return nil
	}
	done := make(chan bool)
	ip := make([]chan string, 0)
	return &StringDemuxManager{ip, output, done}
}

func (sm *StringDemuxManager) Register(source chan string) {
	if source == nil {
		return
	}
	sm.inputs= append(sm.inputs, source)
}

func (sm *StringDemuxManager) Work() {
	go func() {
		for { // endless loop
			for _, c := range sm.inputs {
				select {
				case value, ok := <- c:
					if ok {
						sm.output <- value
					}
				case <- sm.done_chan:
					return
				default:
					// do i have to sleep here to prevent active-wait?
				}
			}
		}
	}()
}

func (sm StringDemuxManager) Send(data string) {
	sm.output <- data
}

func (sm StringDemuxManager) Quit() {
	sm.done_chan <- true
}

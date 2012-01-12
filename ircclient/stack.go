package ircclient

import (
	"container/list"
)

type pluginStack struct {
	list  *list.List
}

func newPluginStack() *pluginStack {
	return &pluginStack{list.New()}
}

func (p *pluginStack) Push(plugin Plugin) {
	if plugin == nil {
		return
	}
	p.list.PushFront(plugin)
}

func (p *pluginStack) Pop() Plugin {
	obj := p.list.Remove(p.list.Front())
	objc, _ := obj.(Plugin)
	return objc
}

func (p *pluginStack) Size() int {
	return p.list.Len()
}

func (p *pluginStack) Iter() <-chan Plugin {
	ch := make(chan Plugin, p.list.Len())
	go func() {
		// Iterate over plugins in the order they were pushed
		for e := p.list.Back(); e != nil; e = e.Prev() {
			v, _ := e.Value.(Plugin)
			ch <- v
		}
		close(ch)
	}()
	return ch
}

func (p *pluginStack) GetPlugin(plugin string) (Plugin, bool) {
	if p.list.Len() == 0 {
		return nil, false
	}
	for e := p.list.Back(); e != nil; e = e.Prev() {
		v, _ := e.Value.(Plugin)
		if v.String() == plugin {
			return v, true
		}
	}
	return nil, false
}

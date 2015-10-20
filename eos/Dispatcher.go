package eos

import (
	"fmt"
	"github.com/gotterdemarung/go-log/log"
	"sync"
)

var Log = log.Context.WithTags("eos", "dispatch")

type node struct {
	f    func(Message)
	next *node
}

type Dispatcher struct {
	m         sync.Mutex
	first     *node
	count     int
	StatCount func(int)
}

func (d *Dispatcher) Register(f func(Message)) {
	node := &node{f, nil}

	d.m.Lock()
	defer d.m.Unlock()
	d.count++
	node.next = d.first
	d.first = node
	go d.StatCount(d.count)
	Log.Infoc("New listener added to dispatcher, total is :total", map[string]interface{}{"total": d.count})

	// Dumping all
	x := d.first
	fmt.Println("REG")
	for x != nil {
		fmt.Printf("  %+v %+v\n", x.f, f)

		x = x.next
	}
}

func (d *Dispatcher) Unregister(f func(Message)) {
	d.m.Lock()
	defer d.m.Unlock()

	// Dumping all
	x := d.first
	fmt.Println("UNREG")
	for x != nil {
		fmt.Printf("  %+v %+v\n", x.f, f)

		x = x.next
	}

	// First position check
	if d.first == nil {
		fmt.Println("NIL !!!!")
		// Nothing to delete
		return
	} else if &(d.first.f) == &f {
		fmt.Println("FIRST !!!!")
		// Removing first node
		d.count--
		d.first = d.first.next
	} else {
		fmt.Println("ELSE !!!!")
		// Middle check
		x := d.first
		for x != nil {
			if x.next != nil && &x.next.f == &f {
				x.next = x.next.next
			}

			x = x.next
		}
	}
	go d.StatCount(d.count)
	Log.Infoc("Listener removed from dispatcher, total is :total", map[string]interface{}{"total": d.count})
}

func (d *Dispatcher) Send(message Message) {
	x := d.first

	for x != nil {
		go x.f(message)
		x = x.next
	}
}

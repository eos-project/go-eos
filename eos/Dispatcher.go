package eos

import (
	"sync"
)

type node struct {
	f 		func(Message)
	next	*node
}

type Dispatcher struct {
	m 			sync.Mutex
	first 		*node
	count 		int
	StatCount 	func(int)
}

func (d *Dispatcher) Register(f func(Message)) {
	node := &node{f, nil}

	d.m.Lock()
	defer d.m.Unlock()
	d.count++
	node.next = d.first
	d.first = node
	go d.StatCount(d.count)
}

func (d *Dispatcher) Unregister(f func(Message)) {
	d.m.Lock()
	defer d.m.Unlock()

	// First position check
	if d.first == nil {
		// Nothing to delete
		return
	} else if &d.first.f == &f {
		// Removing first node
		d.count--
		d.first = d.first.next
	} else {
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
}

func (d *Dispatcher) Send(message Message) {
	x := d.first

	for x != nil {
		go x.f(message)
		x = x.next
	}
}

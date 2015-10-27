package server

import (
	"fmt"
	"github.com/eos-project/go-eos/model"
	"github.com/gotterdemarung/go-log/log"
	"reflect"
	"sync"
)

var l = log.Context.WithTags("eos", "dispatch")

// Dispatcher linked list node
type node struct {
	f    Listener
	next *node
}

// Dispatcher
type Dispatcher struct {
	m         sync.Mutex
	first     *node
	count     int
	StatCount func(int)
}

// Registers new listener
func (d *Dispatcher) Register(f Listener) {
	// Unregister before usage
	d.Unregister(f)

	// Register new listener
	node := &node{f, nil}

	d.m.Lock()
	defer d.m.Unlock()
	d.count++
	node.next = d.first
	d.first = node
	go d.StatCount(d.count)

	l.Context["total"] = d.count
	l.Info("New listener added to dispatcher, total is :total")
}

// Unregisters existing listener (if found)
func (d *Dispatcher) Unregister(f Listener) {
	d.m.Lock()
	defer d.m.Unlock()

	// First position check
	if d.first == nil {
		// Nothing to delete
		return
	} else if funcEq(d.first.f, f) {
		// Removing first node
		d.count--
		d.first = d.first.next
	} else {
		// Middle & last check
		x := d.first
		for x != nil {
			if funcEq(x.f, f) {
				d.count--
				x.f = nil
				break
			}
			if x.next != nil && funcEq(x.next.f, f) {
				d.count--
				x.next = x.next.next
				break
			}

			x = x.next
		}
	}
	go d.StatCount(d.count)

	l.Context["total"] = d.count
	l.Info("Listener removed from dispatcher, total is :total")
}

// Sends message
func (d *Dispatcher) Send(message model.Message) {
	x := d.first

	for x != nil {
		go x.f(message)
		x = x.next
	}
}

// Utility method to dump contents
func (d *Dispatcher) dump() {
	fmt.Println("Dumping")
	x := d.first
	if x == nil || x.f == nil {
		fmt.Println("Empty")
	} else {
		i := 0
		for x != nil {
			if x.f != nil {
				fmt.Printf("%d - %v\n", i, reflect.ValueOf(x.f).Pointer())
			} else {
				fmt.Printf("%d - nil\n", i)
			}
			i++
			x = x.next
		}
	}
}

// Compares for two functions pointer equality
func funcEq(a, b Listener) bool {
	ra := reflect.ValueOf(a)
	rb := reflect.ValueOf(b)

	return ra.Pointer() == rb.Pointer()
}

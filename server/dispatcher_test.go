package server

import (
	"sync"
	"testing"
"github.com/eos-project/go-eos/model"
)

func assertCount(expected, actual int, t *testing.T) {
	if expected == actual {
		t.Logf("[ok] %d == %d", expected, actual)
	} else {
		t.Fatalf("%d != %d", expected, actual)
	}
}

func dis() *Dispatcher {
	return &Dispatcher{StatCount: func(int) {}}
}

func TestDelivery(t *testing.T) {
	d := dis()
	count := 0
	wg := sync.WaitGroup{}
	wg.Add(6)
	f1 := func(model.Message) {
		wg.Done()
		count++
	}

	d.Register(f1)
	for i := 0; i < 6; i++ {
		d.Send(model.Message{})
	}
	wg.Wait()
	assertCount(6, count, t)
}

func TestZeroUnreg(t *testing.T) {
	d := dis()

	assertCount(0, d.count, t)

	f1 := func(model.Message) {}

	// Zero unregistration
	d.Unregister(f1)
	assertCount(0, d.count, t)
}

func TestSimpleRegistration(t *testing.T) {
	d := dis()
	f1 := func(model.Message) {}
	d.Register(f1)
	assertCount(1, d.count, t)
	d.Unregister(f1)
	assertCount(0, d.count, t)
}

func TestDoubleRegistration(t *testing.T) {
	d := dis()
	f1 := func(model.Message) {}
	d.Register(f1)
	d.Register(f1)
	assertCount(1, d.count, t)
	d.Unregister(f1)
	assertCount(0, d.count, t)
}

func TestUnregisterLast(t *testing.T) {
	d := dis()
	f1 := func(model.Message) {}
	f2 := func(model.Message) {}
	f3 := func(model.Message) {}

	d.Register(f1)
	d.Register(f2)
	d.Register(f3)
	assertCount(3, d.count, t)
	d.Unregister(f1)
	assertCount(2, d.count, t)
}

func TestUnregisterMiddle(t *testing.T) {
	d := dis()
	f1 := func(model.Message) {}
	f2 := func(model.Message) {}
	f3 := func(model.Message) {}

	d.Register(f1)
	d.Register(f2)
	d.Register(f3)
	assertCount(3, d.count, t)
	d.Unregister(f2)
	assertCount(2, d.count, t)
}

func TestUnregisterFirst(t *testing.T) {
	d := dis()
	f1 := func(model.Message) {}
	f2 := func(model.Message) {}
	f3 := func(model.Message) {}

	d.Register(f1)
	d.Register(f2)
	d.Register(f3)
	assertCount(3, d.count, t)
	d.Unregister(f3)
	assertCount(2, d.count, t)
}

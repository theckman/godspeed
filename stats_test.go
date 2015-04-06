// Copyright 2014-2015 PagerDuty, Inc, et al. All rights reserved.
// Use of this source code is governed by the BSD 3-Clause
// license that can be found in the LICENSE file.

package godspeed_test

import (
	"bytes"
	"math/rand"
	"testing"

	"github.com/PagerDuty/godspeed"
	"github.com/PagerDuty/godspeed/gspdtest"
)

var chars = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

func randString(n uint) string {
	b := make([]rune, n)

	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}

	return string(b)
}

func TestSend(t *testing.T) {
	const port uint16 = 8127
	var g *godspeed.Godspeed

	l, ctrl, out := gspdtest.BuildListener(port)

	defer l.Close()
	defer close(ctrl)

	go gspdtest.Listener(l, ctrl, out)

	// build a new Godspeed instance with autoTruncation set to true
	g, err := gspdtest.BuildGodspeed(port, true)

	if err != nil {
		t.Error(err.Error())
		return
	}

	//
	// Test whether auto-truncation works
	//

	// add a bunch of tags the pad the body with a lot of content
	for i := 0; i < 2098; i++ {
		g.AddTag(randString(3))
	}

	err = g.Send("test.metric", "c", 42, 1, nil)

	if err != nil {
		t.Error(err.Error())
		return
	}

	a, ok := <-out

	if !ok {
		t.Error(closedChan)
		return
	}

	if len(a) != godspeed.MaxBytes {
		t.Errorf("datagram is an unexpected size (%d) should be %d bytes (MaxBytes)", len(a), godspeed.MaxBytes)
	}

	g.Conn.Close()

	// build a new Godspeed instance with autoTruncation set to false
	g, err = gspdtest.BuildGodspeed(port, false)

	if err != nil {
		t.Error(err.Error())
		return
	}

	defer g.Conn.Close()

	//
	// test whether sending a plain metric works
	//
	err = g.Send("testing.metric", "ms", 256.512, 1, nil)

	if err != nil {
		t.Errorf(err.Error())
		return
	}

	a, ok = <-out

	if !ok {
		t.Error(closedChan)
		return
	}

	b := []byte("testing.metric:256.512|ms")

	if !bytes.Equal(a, b) {
		t.Error(gspdtest.NoGo(a, b))
	}

	//
	// test whether sending a large metric works
	//
	err = g.Send("testing.metric", "g", 5536650702696, 1, nil)

	if err != nil {
		t.Errorf(err.Error())
		return
	}

	a, ok = <-out

	if !ok {
		t.Error(closedChan)
		return
	}

	b = []byte("testing.metric:5536650702696|g")

	if !bytes.Equal(a, b) {
		t.Error(gspdtest.NoGo(a, b))
	}

	//
	// test whether sending a metric with a sample rate works
	//
	err = g.Send("testing.metric", "ms", 256.512, 0.99, nil)

	if err != nil {
		t.Error(err.Error())
		return
	}

	a, ok = <-out

	if !ok {
		t.Error(closedChan)
		return
	}

	b = []byte("testing.metric:256.512|ms|@0.99")

	if !bytes.Equal(a, b) {
		t.Error(gspdtest.NoGo(a, b))
	}

	//
	// test whether metrics are properly sent with the namespace
	//
	g.SetNamespace("godspeed")

	err = g.Send("testing.metric", "ms", 512.1024, 1, nil)

	if err != nil {
		t.Error(err.Error())
		return
	}

	a, ok = <-out

	if !ok {
		t.Error(closedChan)
		return
	}

	b = []byte("godspeed.testing.metric:512.1024|ms")

	if !bytes.Equal(a, b) {
		t.Error(gspdtest.NoGo(a, b))
	}

	//
	// test that adding a tag to the instance sends it with the metric
	//
	g.AddTag("test")

	err = g.Send("testing.metric", "ms", 512.1024, 1, nil)

	if err != nil {
		t.Error(err.Error())
		return
	}

	a, ok = <-out

	if !ok {
		t.Error(closedChan)
		return
	}

	b = []byte("godspeed.testing.metric:512.1024|ms|#test")

	if !bytes.Equal(a, b) {
		t.Error(gspdtest.NoGo(a, b))
	}

	//
	// test whether adding a second tag causes both to get sent with the stat
	//
	g.AddTag("test1")

	err = g.Send("testing.metric", "ms", 512.1024, 1, nil)

	if err != nil {
		t.Error(err.Error())
		return
	}

	a, ok = <-out

	if !ok {
		t.Error(closedChan)
		return
	}

	b = []byte("godspeed.testing.metric:512.1024|ms|#test,test1")

	if !bytes.Equal(a, b) {
		t.Error(gspdtest.NoGo(a, b))
	}

	//
	// test whether adding multiple tags sends all tags with the metric
	//
	g.AddTags([]string{"test2", "test3"})

	err = g.Send("testing.metric", "ms", 512.1024, 1, nil)

	if err != nil {
		t.Error(err.Error())
		return
	}

	a, ok = <-out

	if !ok {
		t.Error(closedChan)
		return
	}

	b = []byte("godspeed.testing.metric:512.1024|ms|#test,test1,test2,test3")

	if !bytes.Equal(a, b) {
		t.Error(gspdtest.NoGo(a, b))
	}

	//
	// test that adding metrics to the stat sends all instance tags with provided tags appended
	//
	err = g.Send("testing.metric", "ms", 512.1024, 1, []string{"test4", "test5", "test3"})

	if err != nil {
		t.Error(err.Error())
		return
	}

	a, ok = <-out

	if !ok {
		t.Error(closedChan)
		return
	}

	b = []byte("godspeed.testing.metric:512.1024|ms|#test,test1,test2,test3,test4,test5")

	if !bytes.Equal(a, b) {
		t.Error(gspdtest.NoGo(a, b))
	}

	//
	// test that adding tags to a single metric doesn't persist on future stats
	//
	err = g.Send("testing.metric", "ms", 512.1024, 1, nil)

	if err != nil {
		t.Error(err.Error())
		return
	}

	a, ok = <-out

	if !ok {
		t.Error(closedChan)
		return
	}

	b = []byte("godspeed.testing.metric:512.1024|ms|#test,test1,test2,test3")

	if !bytes.Equal(a, b) {
		t.Error(gspdtest.NoGo(a, b))
	}

	//
	// test that a failure is returned when autoTruncate is false, and the body is larger than MAX_BYTES
	//
	for i := 0; i < 2100; i++ {
		g.AddTag(randString(3))
	}

	err = g.Send("test.metric", "c", 42, 1, nil)

	if err == nil {
		// pull the message out of the return channel and discard
		_ = <-out
		t.Error("expected error; autoTruncate should be diabled/message should have been too long")
	}
}

func TestCount(t *testing.T) {
	const port uint16 = 8128
	var g *godspeed.Godspeed

	l, ctrl, out := gspdtest.BuildListener(port)

	defer l.Close()
	defer close(ctrl)

	go gspdtest.Listener(l, ctrl, out)

	g, err := gspdtest.BuildGodspeed(port, false)

	if err != nil {
		t.Error(err.Error())
		return
	}

	defer g.Conn.Close()

	err = g.Count("test.count", 1, nil)

	if err != nil {
		t.Error(err.Error())
		return
	}

	a, ok := <-out

	if !ok {
		t.Error(closedChan)
		return
	}

	b := []byte("test.count:1|c")

	if !bytes.Equal(a, b) {
		t.Error(gspdtest.NoGo(a, b))
	}
}

func TestIncr(t *testing.T) {
	const port uint16 = 8129
	var g *godspeed.Godspeed

	l, ctrl, out := gspdtest.BuildListener(port)

	defer l.Close()
	defer close(ctrl)

	go gspdtest.Listener(l, ctrl, out)

	g, err := gspdtest.BuildGodspeed(port, false)

	if err != nil {
		t.Error(err.Error())
		return
	}

	defer g.Conn.Close()

	err = g.Incr("test.incr", nil)

	if err != nil {
		t.Error(err.Error())
		return
	}

	a, ok := <-out

	if !ok {
		t.Error(closedChan)
		return
	}

	b := []byte("test.incr:1|c")

	if !bytes.Equal(a, b) {
		t.Error(gspdtest.NoGo(a, b))
	}
}

func TestDecr(t *testing.T) {
	const port uint16 = 8130
	var g *godspeed.Godspeed

	l, ctrl, out := gspdtest.BuildListener(port)

	defer l.Close()
	defer close(ctrl)

	go gspdtest.Listener(l, ctrl, out)

	g, err := gspdtest.BuildGodspeed(port, false)

	if err != nil {
		t.Error(err.Error())
		return
	}

	defer g.Conn.Close()

	err = g.Decr("test.decr", nil)

	if err != nil {
		t.Error(err.Error())
		return
	}

	a, ok := <-out

	if !ok {
		t.Error(closedChan)
		return
	}

	b := []byte("test.decr:-1|c")

	if !bytes.Equal(a, b) {
		t.Error(gspdtest.NoGo(a, b))
	}
}

func TestGauge(t *testing.T) {
	const port uint16 = 8131
	var g *godspeed.Godspeed

	l, ctrl, out := gspdtest.BuildListener(port)

	defer l.Close()
	defer close(ctrl)

	go gspdtest.Listener(l, ctrl, out)

	g, err := gspdtest.BuildGodspeed(port, false)

	if err != nil {
		t.Error(err.Error())
		return
	}

	defer g.Conn.Close()

	err = g.Gauge("test.gauge", 42, nil)

	if err != nil {
		t.Error(err.Error())
		return
	}

	a, ok := <-out

	if !ok {
		t.Error(closedChan)
		return
	}

	b := []byte("test.gauge:42|g")

	if !bytes.Equal(a, b) {
		t.Error(gspdtest.NoGo(a, b))
	}

}

func TestHistogram(t *testing.T) {
	const port uint16 = 8130
	var g *godspeed.Godspeed

	l, ctrl, out := gspdtest.BuildListener(port)

	defer l.Close()
	defer close(ctrl)

	go gspdtest.Listener(l, ctrl, out)

	g, err := gspdtest.BuildGodspeed(port, false)

	if err != nil {
		t.Error(err.Error())
		return
	}

	defer g.Conn.Close()

	err = g.Histogram("test.hist", 84, nil)

	if err != nil {
		t.Error(err.Error())
		return
	}

	a, ok := <-out

	if !ok {
		t.Error(closedChan)
		return
	}

	b := []byte("test.hist:84|h")

	if !bytes.Equal(a, b) {
		t.Error(gspdtest.NoGo(a, b))
	}

}

func TestTiming(t *testing.T) {
	const port uint16 = 8130
	var g *godspeed.Godspeed

	l, ctrl, out := gspdtest.BuildListener(port)

	defer l.Close()
	defer close(ctrl)

	go gspdtest.Listener(l, ctrl, out)

	g, err := gspdtest.BuildGodspeed(port, false)

	if err != nil {
		t.Error(err.Error())
		return
	}

	defer g.Conn.Close()

	err = g.Timing("test.timing", 2054, nil)

	if err != nil {
		t.Error(err.Error())
		return
	}

	a, ok := <-out

	if !ok {
		t.Error(closedChan)
		return
	}

	b := []byte("test.timing:2054|ms")

	if !bytes.Equal(a, b) {
		t.Error(gspdtest.NoGo(a, b))
	}

}

func TestSet(t *testing.T) {
	const port uint16 = 8130
	var g *godspeed.Godspeed

	l, ctrl, out := gspdtest.BuildListener(port)

	defer l.Close()
	defer close(ctrl)

	go gspdtest.Listener(l, ctrl, out)

	g, err := gspdtest.BuildGodspeed(port, false)

	if err != nil {
		t.Error(err.Error())
		return
	}

	defer g.Conn.Close()

	err = g.Set("test.set", 10, nil)

	if err != nil {
		t.Error(err.Error())
		return
	}

	a, ok := <-out

	if !ok {
		t.Error(closedChan)
		return
	}

	b := []byte("test.set:10|s")

	if !bytes.Equal(a, b) {
		t.Error(gspdtest.NoGo(a, b))
	}

}

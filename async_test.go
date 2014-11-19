// Copyright 2014 PagerDuty, Inc, et al. All rights reserved.
// Use of this source code is governed by the BSD 3-Clause
// license that can be found in the LICENSE file.

package godspeed

import (
	"bytes"
	"fmt"
	"net"
	"testing"
	"time"
)

var extraTestTags = []string{"test8", "test9"}

func buildAsyncGodspeed(port uint16, autoTruncate bool) (g *AsyncGodspeed, err error) {
	g, err = NewAsync("127.0.0.1", port, autoTruncate)
	return
}

func testAsyncBasicFunctionality(t *testing.T, g *AsyncGodspeed, l *net.UDPConn, ctrl chan int, out chan []byte) {
	g.AddTag("test0")
	g.SetNamespace("godspeed")

	g.W.Add(1)
	go g.Send("test.metric", "c", 1, 1, []string{"test1", "test2"}, g.W)

	a, ok := <-out

	if !ok {
		// error and return as there is no reason to run further tests
		// they will most likely fail as this channel has been closed early
		t.Error(closedChan)
		return
	}

	b := []byte("godspeed.test.metric:1|c|#test0,test1,test2")

	if !bytes.Equal(a, b) {
		t.Error(noGo(a, b))
	}
}

func (a *AsyncGodspeed) testWarmUp() {
	a.SetNamespace("godspeed")
	a.AddTags([]string{"test0", "test1"})
}

func TestNewAsync(t *testing.T) {
	// define port for listener
	const port uint16 = 8126
	var g *AsyncGodspeed

	// build the listener and return the following:
	// listener, control channel (close to stop), and the out (return) channel
	l, ctrl, out := buildListener(port)

	// defer cleaning up stuff
	defer l.Close()
	defer close(ctrl)

	// send the listener out to a goroutine
	go listener(l, ctrl, out)

	// build Godspeed
	g, err := buildAsyncGodspeed(port, false)

	if err != nil {
		// we failed to get a Godspeed client
		t.Error(err.Error())
		return
	}

	// defer closing the client for Godspeed
	defer g.Godspeed.Conn.Close()

	// test defined basic functionality
	testAsyncBasicFunctionality(t, g, l, ctrl, out)
}

func TestNewDefaultAsync(t *testing.T) {
	const port uint16 = 8125
	var g *AsyncGodspeed

	l, ctrl, out := buildListener(port)

	defer l.Close()
	defer close(ctrl)

	go listener(l, ctrl, out)

	g, err := NewDefaultAsync()

	if err != nil {
		t.Errorf("unexpected error when building new Godspeed client: %v", err)
		return
	}

	defer g.Godspeed.Conn.Close()

	testAsyncBasicFunctionality(t, g, l, ctrl, out)
}

// TestAsyncAddTags tests g.AddTag() and g.AddTags()
func TestAsyncAddTags(t *testing.T) {
	g, err := buildAsyncGodspeed(DEFAULT_PORT, false)

	if err != nil {
		t.Error(err.Error())
		return
	}

	defer g.Godspeed.Conn.Close()

	g.AddTag("testing0")

	tags := g.Godspeed.Tags

	if len(tags) != 1 && tags[0] != "testing0" {
		t.Error("failed adding 'testing0' tag to client")
	}

	g.AddTags([]string{"testing1", "testing2"})

	tags = g.Godspeed.Tags

	if len(tags) != 3 && tags[0] != "testing0" && tags[1] != "testing1" && tags[2] != "testing3" {
		t.Error("failed to add 'testing1' and 'testing2' to the tags")
	}
}

func TestAsyncSetNamespace(t *testing.T) {
	g, err := buildAsyncGodspeed(DEFAULT_PORT, false)

	if err != nil {
		t.Error(err.Error())
		return
	}

	defer g.Godspeed.Conn.Close()

	g.SetNamespace("testing1")

	if g.Godspeed.Namespace != "testing1" {
		t.Error("failed to set the namespace")
	}
}

func TestAsyncEvent(t *testing.T) {
	const port uint16 = 8125
	var g *AsyncGodspeed

	l, ctrl, out := buildListener(port)

	defer l.Close()
	defer close(ctrl)

	go listener(l, ctrl, out)

	g, err := buildAsyncGodspeed(port, false)

	if err != nil {
		t.Error(err.Error())
		return
	}

	defer g.Godspeed.Conn.Close()

	g.AddTags([]string{"test0", "test1"})

	unix := time.Now().Unix()

	m := make(map[string]string)
	m["date_happened"] = fmt.Sprintf("%d", unix)
	m["hostname"] = "test01"
	m["aggregation_key"] = "xyz"
	m["priority"] = "low"
	m["source_type_name"] = "cassandra"
	m["alert_type"] = "info"

	g.W.Add(1)
	go g.Event("a", "b", m, []string{"test8", "test9"}, g.W)

	a, ok := <-out

	if !ok {
		t.Error(closedChan)
		return
	}

	b := []byte(fmt.Sprintf("_e{1,1}:a|b|d:%d|h:test01|k:xyz|p:low|s:cassandra|t:info|#test0,test1,test8,test9", unix))

	if !bytes.Equal(a, b) {
		t.Error(noGo(a, b))
	}
}

func TestAsyncSend(t *testing.T) {
	const port uint16 = 8126
	var g *AsyncGodspeed

	l, ctrl, out := buildListener(port)

	defer l.Close()
	defer close(ctrl)

	go listener(l, ctrl, out)

	g, err := buildAsyncGodspeed(port, false)

	if err != nil {
		t.Error(err.Error())
		return
	}

	defer g.Godspeed.Conn.Close()

	g.testWarmUp()

	g.W.Add(1)
	go g.Send("test.stat", "g", 42, 0.99, extraTestTags, g.W)

	a, ok := <-out

	if !ok {
		t.Error(closedChan)
		return
	}

	b := []byte("godspeed.test.stat:42|g|@0.990000|#test0,test1,test8,test9")

	if !bytes.Equal(a, b) {
		t.Error(noGo(a, b))
	}
}

func TestAsyncCount(t *testing.T) {
	const port uint16 = 8127
	var g *AsyncGodspeed

	l, ctrl, out := buildListener(port)

	defer l.Close()
	defer close(ctrl)

	go listener(l, ctrl, out)

	g, err := buildAsyncGodspeed(port, false)

	if err != nil {
		return
	}

	defer g.Godspeed.Conn.Close()

	g.testWarmUp()

	g.W.Add(1)
	go g.Count("test.count", 1, extraTestTags, g.W)

	a, ok := <-out

	if !ok {
		t.Error(closedChan)
		return
	}

	b := []byte("godspeed.test.count:1|c|#test0,test1,test8,test9")

	if !bytes.Equal(a, b) {
		t.Error(noGo(a, b))
	}
}

func TestAsyncIncr(t *testing.T) {
	const port uint16 = 8128
	var g *AsyncGodspeed

	l, ctrl, out := buildListener(port)

	defer l.Close()
	defer close(ctrl)

	go listener(l, ctrl, out)

	g, err := buildAsyncGodspeed(port, false)

	if err != nil {
		t.Error(err.Error())
		return
	}

	defer g.Godspeed.Conn.Close()

	g.testWarmUp()

	g.W.Add(1)
	go g.Incr("test.incr", extraTestTags, g.W)

	a, ok := <-out

	if !ok {
		t.Error(closedChan)
		return
	}

	b := []byte("godspeed.test.incr:1|c|#test0,test1,test8,test9")

	if !bytes.Equal(a, b) {
		t.Error(noGo(a, b))
	}
}

func TestAsyncDecr(t *testing.T) {
	const port uint16 = 8129
	var g *AsyncGodspeed

	l, ctrl, out := buildListener(port)

	defer l.Close()
	defer close(ctrl)

	go listener(l, ctrl, out)

	g, err := buildAsyncGodspeed(port, false)

	if err != nil {
		t.Error(err.Error())
		return
	}

	defer g.Godspeed.Conn.Close()

	g.testWarmUp()

	g.W.Add(1)
	go g.Decr("test.decr", extraTestTags, g.W)

	a, ok := <-out

	if !ok {
		t.Error(closedChan)
		return
	}

	b := []byte("godspeed.test.decr:-1|c|#test0,test1,test8,test9")

	if !bytes.Equal(a, b) {
		t.Error(noGo(a, b))
	}
}

func TestAsyncGauge(t *testing.T) {
	const port uint16 = 8130
	var g *AsyncGodspeed

	l, ctrl, out := buildListener(port)

	defer l.Close()
	defer close(ctrl)

	go listener(l, ctrl, out)

	g, err := buildAsyncGodspeed(port, false)

	if err != nil {
		t.Error(err.Error())
		return
	}

	defer g.Godspeed.Conn.Close()

	g.testWarmUp()

	g.W.Add(1)
	go g.Gauge("test.gauge", 42, extraTestTags, g.W)

	a, ok := <-out

	if !ok {
		t.Error(closedChan)
		return
	}

	b := []byte("godspeed.test.gauge:42|g|#test0,test1,test8,test9")

	if !bytes.Equal(a, b) {
		t.Error(noGo(a, b))
	}
}

func TestAsyncHistogram(t *testing.T) {
	const port uint16 = 8131
	var g *AsyncGodspeed

	l, ctrl, out := buildListener(port)

	defer l.Close()
	defer close(ctrl)

	go listener(l, ctrl, out)

	g, err := buildAsyncGodspeed(port, false)

	if err != nil {
		t.Error(err.Error())
		return
	}

	defer g.Godspeed.Conn.Close()

	g.testWarmUp()

	g.W.Add(1)
	go g.Histogram("test.hist", 2, extraTestTags, g.W)

	a, ok := <-out

	if !ok {
		t.Error(closedChan)
		return
	}

	b := []byte("godspeed.test.hist:2|h|#test0,test1,test8,test9")

	if !bytes.Equal(a, b) {
		t.Error(noGo(a, b))
	}
}

func TestAsyncTiming(t *testing.T) {
	const port uint16 = 8132
	var g *AsyncGodspeed

	l, ctrl, out := buildListener(port)

	defer l.Close()
	defer close(ctrl)

	go listener(l, ctrl, out)

	g, err := buildAsyncGodspeed(port, false)

	if err != nil {
		t.Error(err.Error())
		return
	}

	defer g.Godspeed.Conn.Close()

	g.testWarmUp()

	g.W.Add(1)
	go g.Timing("test.timing", 3, extraTestTags, g.W)

	a, ok := <-out

	if !ok {
		t.Error(closedChan)
		return
	}

	b := []byte("godspeed.test.timing:3|ms|#test0,test1,test8,test9")

	if !bytes.Equal(a, b) {
		t.Error(noGo(a, b))
	}
}

func TestAsyncSet(t *testing.T) {
	const port uint16 = 8133
	var g *AsyncGodspeed

	l, ctrl, out := buildListener(port)

	defer l.Close()
	defer close(ctrl)

	go listener(l, ctrl, out)

	g, err := buildAsyncGodspeed(port, false)

	if err != nil {
		t.Error(err.Error())
		return
	}

	defer g.Godspeed.Conn.Close()

	g.testWarmUp()

	g.W.Add(1)
	go g.Set("test.set", 4, extraTestTags, g.W)

	a, ok := <-out

	if !ok {
		t.Error(closedChan)
		return
	}

	b := []byte("godspeed.test.set:4|s|#test0,test1,test8,test9")

	if !bytes.Equal(a, b) {
		t.Error(noGo(a, b))
	}
}

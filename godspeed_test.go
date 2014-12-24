// Copyright 2014 PagerDuty, Inc, et al. All rights reserved.
// Use of this source code is governed by the BSD 3-Clause
// license that can be found in the LICENSE file.

package godspeed

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"testing"
)

const closedChan = "return channel (out) closed prematurely"

func listener(l *net.UDPConn, ctrl chan int, c chan []byte) {
	for {
	NextListen:
		select {
		case _, ctrl_ok := <-ctrl:
			if !ctrl_ok {
				close(c)
				return
			}
		default:
			buffer := make([]byte, 8193)

			_, err := l.Read(buffer)

			if err != nil {
				goto NextListen
			}

			c <- bytes.Trim(buffer, "\x00")
		}
	}
}

func buildListener(port uint16) (*net.UDPConn, chan int, chan []byte) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", port))

	if err != nil {
		panic(fmt.Sprintf("getting address for test listener failed, bailing out. Here's everything I know: %v", err))
	}

	l, err := net.ListenUDP("udp", addr)

	if err != nil {
		panic(fmt.Sprintf("unable to listen for traffic: %v", err))
	}

	return l, make(chan int), make(chan []byte)
}

func buildGodspeed(port uint16, autoTruncate bool) (g *Godspeed, err error) {
	g, err = New("127.0.0.1", port, autoTruncate)
	return
}

func noGo(a, b []byte) string {
	return fmt.Sprintf("%v (%v) is not equal to %v (%v)", string(a), a, string(b), b)
}

func testBasicFunctionality(t *testing.T, g *Godspeed, l *net.UDPConn, ctrl chan int, out chan []byte) {
	err := g.Send("test.metric", "c", 1, 1, nil)

	if err != nil {
		// if the send failed there's no reason to continue the test
		t.Error(err.Error())
		return
	}

	a, ok := <-out

	if !ok {
		// error and return as there is no reason to run further tests
		// they will most likely fail as this channel has been closed early
		t.Error(closedChan)
		return
	}

	b := []byte("test.metric:1|c")

	if !bytes.Equal(a, b) {
		t.Error(noGo(a, b))
	}

	if len(g.Tags) != 0 {
		t.Errorf("there are more than zero tags set: %v", strings.Join(g.Tags, ", "))
	}

	if len(g.Namespace) != 0 {
		t.Errorf("a namespace has already been set: %v", g.Namespace)
	}
}

func TestNew(t *testing.T) {
	// define port for listener
	const port uint16 = 8126
	var g *Godspeed

	// build the listener and return the following:
	// listener, control channel (close to stop), and the out (return) channel
	l, ctrl, out := buildListener(port)

	// defer cleaning up stuff
	defer l.Close()
	defer close(ctrl)

	// send the listener out to a goroutine
	go listener(l, ctrl, out)

	// build Godspeed
	g, err := buildGodspeed(port, false)

	if err != nil {
		// we failed to get a Godspeed client
		t.Error(err.Error())
		return
	}

	// defer closing the client for Godspeed
	defer g.Conn.Close()

	// test defined basic functionality
	testBasicFunctionality(t, g, l, ctrl, out)
}

func TestNewDefault(t *testing.T) {
	const port uint16 = 8125
	var g *Godspeed

	l, ctrl, out := buildListener(port)

	defer l.Close()
	defer close(ctrl)

	go listener(l, ctrl, out)

	g, err := NewDefault()

	if err != nil {
		t.Errorf("unexpected error when building new Godspeed client: %v", err)
		return
	}

	defer g.Conn.Close()

	testBasicFunctionality(t, g, l, ctrl, out)
}

func TestAddTag(t *testing.T) {
	var g *Godspeed

	g, err := buildGodspeed(DEFAULT_PORT, false)

	if err != nil {
		t.Error(err.Error())
		return
	}

	defer g.Conn.Close()

	if len(g.Tags) != 0 {
		t.Errorf("tags have already been set on the Godspeed instance: %v", strings.Join(g.Tags, ", "))
		return
	}

	g.AddTag("test")

	if len(g.Tags) != 1 || g.Tags[0] != "test" {
		t.Errorf("something went wrong adding 'test' tag; current tags: %v", strings.Join(g.Tags, ", "))
	}

	g.AddTag("test2")
	g.AddTag("test") // verify we de-dupe

	if len(g.Tags) != 2 || g.Tags[0] != "test" || g.Tags[1] != "test2" {
		t.Errorf("something went wrong adding 'test2' tag; current tags: %v", strings.Join(g.Tags, ", "))
	}
}

func TestAddTags(t *testing.T) {
	var g *Godspeed

	g, err := buildGodspeed(DEFAULT_PORT, false)

	if err != nil {
		t.Error(err.Error())
		return
	}

	defer g.Conn.Close()

	// make sure no tags are set
	if len(g.Tags) != 0 {
		t.Errorf("tags have already been set on the Godspeed instance: %v", strings.Join(g.Tags, ", "))
		return
	}

	tags := []string{"test1", "test2", "test1"}

	//
	// test that adding the first set of tags worked
	//
	g.AddTags(tags)

	//verify two tags
	if len(g.Tags) != 2 {
		t.Error("expected there to be two tags")
	}

	// match the two tags
	for i, v := range g.Tags {
		if v != tags[i] {
			t.Errorf("expected %v to equal %v", v, tags[i])
		}
	}

	// add some more tags
	tags2 := []string{"test3", "test4", "test5", "test4"}
	tags = append(tags, tags2...)

	//
	// test that appending the second set of tags works without overwriting the first
	//
	g.AddTags(tags2)

	// expected tags
	control := []string{"test1", "test2", "test3", "test4", "test5"}

	// match length
	if len(g.Tags) != 5 {
		t.Error("expected there to be 5 tags")
	}

	// match content
	for i, v := range g.Tags {
		if v != control[i] {
			t.Errorf("expected %v to equal %v", v, tags[i])
		}
	}
}

func TestSetNamespace(t *testing.T) {
	var g *Godspeed

	g, err := buildGodspeed(DEFAULT_PORT, false)

	if err != nil {
		t.Error(err.Error())
		return
	}

	defer g.Conn.Close()

	if len(g.Namespace) != 0 {
		t.Errorf("namespace has already been set on the Godspeed instance: %v", g.Namespace)
	}

	g.SetNamespace("heckman")

	if g.Namespace != "heckman" {
		t.Errorf("failure while trying to set Namespace to 'heckman', is: %v", g.Namespace)
	}
}

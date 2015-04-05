// Copyright 2014 PagerDuty, Inc, et al. All rights reserved.
// Use of this source code is governed by the BSD 3-Clause
// license that can be found in the LICENSE file.

package godspeed_test

import (
	"bytes"
	"net"
	"strings"
	"testing"

	"github.com/PagerDuty/godspeed"
	"github.com/PagerDuty/godspeed/gspdtest"
)

const closedChan = "return channel (out) closed prematurely"

func testBasicFunctionality(t *testing.T, g *godspeed.Godspeed, l *net.UDPConn, ctrl chan int, out chan []byte) {
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
		t.Error(gspdtest.NoGo(a, b))
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
	var g *godspeed.Godspeed

	// build the listener and return the following:
	// listener, control channel (close to stop), and the out (return) channel
	l, ctrl, out := gspdtest.BuildListener(port)

	// defer cleaning up stuff
	defer l.Close()
	defer close(ctrl)

	// send the listener out to a goroutine
	go gspdtest.Listener(l, ctrl, out)

	// build Godspeed
	g, err := gspdtest.BuildGodspeed(port, false)

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
	var g *godspeed.Godspeed

	l, ctrl, out := gspdtest.BuildListener(port)

	defer l.Close()
	defer close(ctrl)

	go gspdtest.Listener(l, ctrl, out)

	g, err := godspeed.NewDefault()

	if err != nil {
		t.Errorf("unexpected error when building new Godspeed client: %v", err)
		return
	}

	defer g.Conn.Close()

	testBasicFunctionality(t, g, l, ctrl, out)
}

func TestAddTag(t *testing.T) {
	var g *godspeed.Godspeed

	g, err := gspdtest.BuildGodspeed(godspeed.DefaultPort, false)

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
	var g *godspeed.Godspeed

	g, err := gspdtest.BuildGodspeed(godspeed.DefaultPort, false)

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
	var g *godspeed.Godspeed

	g, err := gspdtest.BuildGodspeed(godspeed.DefaultPort, false)

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

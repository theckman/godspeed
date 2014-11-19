// Copyright 2014 PagerDuty, Inc, et al. All rights reserved.
// Use of this source code is governed by the BSD 3-Clause
// license that can be found in the LICENSE file.

package godspeed

import (
	"bytes"
	"fmt"
	"testing"
	"time"
)

func TestEvent(t *testing.T) {
	const port uint16 = 8125
	var g *Godspeed

	l, ctrl, out := buildListener(port)

	defer l.Close()
	defer close(ctrl)

	go listener(l, ctrl, out)

	g, err := buildGodspeed(port, false)

	if err != nil {
		t.Error(err.Error())
		return
	}

	defer g.Conn.Close()

	//
	// test that adding tags to both Godspeed and the Event send all tags
	//
	//
	// test whether length validation works
	//
	body := make([]byte, 8192)

	for i := range body {
		body[i] = 'a'
	}

	err = g.Event("some event", string(body), nil, nil)

	if err == nil {
		t.Error("expected error not seen; this should have caused a message length error")
		return
	}

	err = g.Event("some\nother event", "some body", nil, nil)

	if err != nil {
		t.Error(err.Error())
		return
	}

	a, ok := <-out

	if !ok {
		t.Error(closedChan)
		return
	}

	b := []byte("_e{18,9}:some\\\\nother event|some body")

	if !bytes.Equal(a, b) {
		t.Error(noGo(a, b))
	}

	//
	// test whether title validation works
	//
	err = g.Event("", "s", nil, nil)

	// if the title's len is < 1, we should get an error
	if err.Error() != "title must have at least one character" {
		t.Error("expected Event validation to fail due to invalid title")
	}

	//
	// test whether body validation works
	//
	err = g.Event("s", "", nil, nil)

	// if the body's len is < 1, we should get an error
	if err.Error() != "body must have at least one character" {
		t.Error("expectefd Event() validation to fail due to invalid body")
	}

	//
	// test that 'date_happened' value gets passed
	//
	unix := time.Now().Unix()

	m := make(map[string]string)
	m["date_happened"] = fmt.Sprintf("%d", unix)

	err = g.Event("a", "b", m, nil)

	if err != nil {
		t.Error(err.Error())
		return
	}

	a, ok = <-out

	if !ok {
		t.Error(closedChan)
		return
	}

	b = []byte(fmt.Sprintf("_e{1,1}:a|b|d:%d", unix))

	if !bytes.Equal(a, b) {
		t.Error(noGo(a, b))
	}

	//
	// test that 'hostname' value gets passed
	//
	m = make(map[string]string)
	m["hostname"] = "tes|t01"

	err = g.Event("b", "c", m, nil)

	if err != nil {
		t.Error(err.Error())
		return
	}

	a, ok = <-out

	if !ok {
		t.Error(closedChan)
		return
	}

	b = []byte("_e{1,1}:b|c|h:test01")

	if !bytes.Equal(a, b) {
		t.Error(noGo(a, b))
	}

	//
	// test that 'aggregation_key' value gets passed
	//
	m = make(map[string]string)
	m["aggregation_key"] = "xyz"

	err = g.Event("c", "d", m, nil)

	if err != nil {
		t.Error(err.Error())
		return
	}

	a, ok = <-out

	if !ok {
		t.Error(closedChan)
		return
	}

	b = []byte("_e{1,1}:c|d|k:xyz")

	if !bytes.Equal(a, b) {
		t.Error(noGo(a, b))
	}

	//
	// test that 'priority' value gets passed
	//
	m = make(map[string]string)
	m["priority"] = "low"

	err = g.Event("d", "e", m, nil)

	if err != nil {
		t.Error(err.Error())
		return
	}

	a, ok = <-out

	if !ok {
		t.Error(closedChan)
		return
	}

	b = []byte("_e{1,1}:d|e|p:low")

	if !bytes.Equal(a, b) {
		t.Error(noGo(a, b))
	}

	//
	// test that 'source_type_name' value gets passed
	//
	m = make(map[string]string)
	m["source_type_name"] = "cassandra"

	err = g.Event("e", "f", m, nil)

	if err != nil {
		t.Error(err.Error())
		return
	}

	a, ok = <-out

	if !ok {
		t.Error(closedChan)
		return
	}

	b = []byte("_e{1,1}:e|f|s:cassandra")

	if !bytes.Equal(a, b) {
		t.Error(noGo(a, b))
	}

	//
	// test that 'alert_type' value gets passed
	//
	m = make(map[string]string)
	m["alert_type"] = "info"

	err = g.Event("f", "g", m, nil)

	if err != nil {
		t.Error(err.Error())
		return
	}

	a, ok = <-out

	if !ok {
		t.Error(closedChan)
		return
	}

	b = []byte("_e{1,1}:f|g|t:info")

	if !bytes.Equal(a, b) {
		t.Error(noGo(a, b))
	}

	//
	// test that adding all values makes sure that all get passed
	//
	m = make(map[string]string)
	m["date_happened"] = fmt.Sprintf("%d", unix)
	m["hostname"] = "test01"
	m["aggregation_key"] = "xyz"
	m["priority"] = "low"
	m["source_type_name"] = "cassandra"
	m["alert_type"] = "info"

	err = g.Event("g", "h", m, nil)

	if err != nil {
		t.Error(err.Error())
		return
	}

	a, ok = <-out

	if !ok {
		t.Error(closedChan)
		return
	}

	b = []byte(fmt.Sprintf("_e{1,1}:g|h|d:%d|h:test01|k:xyz|p:low|s:cassandra|t:info", unix))

	if !bytes.Equal(a, b) {
		t.Error(noGo(a, b))
	}

	m = nil

	//
	// test that adding tags only to the event works
	//
	err = g.Event("h", "i", nil, []string{"test8", "test9"})

	if err != nil {
		t.Error(err.Error())
		return
	}

	a, ok = <-out

	if !ok {
		t.Error(closedChan)
		return
	}

	b = []byte("_e{1,1}:h|i|#test8,test9")

	if !bytes.Equal(a, b) {
		t.Error(noGo(a, b))
	}

	//
	// test that adding the tags to the Godspeed instance sends them with the event
	//
	tgs := g.AddTags([]string{"test0", "test1"})

	if len(tgs) != 2 {
		t.Errorf("expected there to be 2 tags, there are %d", len(tgs))
		return
	}

	err = g.Event("i", "j", nil, nil)

	if err != nil {
		t.Error(err.Error())
		return
	}

	a, ok = <-out

	if !ok {
		t.Error(closedChan)
		return
	}

	b = []byte("_e{1,1}:i|j|#test0,test1")

	if !bytes.Equal(a, b) {
		t.Error(noGo(a, b))
	}

	//
	// test that adding tags to both Godspeed and the Event send all tags
	//
	err = g.Event("j", "k", nil, []string{"test8", "test9"})

	if err != nil {
		t.Error(err.Error())
		return
	}

	a, ok = <-out

	if !ok {
		t.Error(closedChan)
		return
	}

	b = []byte("_e{1,1}:j|k|#test0,test1,test8,test9")

	if !bytes.Equal(a, b) {
		t.Error(noGo(a, b))
	}
}

// Copyright 2014-2015 PagerDuty, Inc, et al. All rights reserved.
// Use of this source code is governed by the BSD 3-Clause
// license that can be found in the LICENSE file.

package gspdtest

import (
	"bytes"
	"fmt"
	"net"

	"github.com/PagerDuty/godspeed"
)

func Listener(l *net.UDPConn, ctrl chan int, c chan []byte) {
	for {
		select {
		case _, ok := <-ctrl:
			if !ok {
				close(c)
				return
			}
		default:
			buffer := make([]byte, 8193)

			_, err := l.Read(buffer)

			if err != nil {
				continue
			}

			c <- bytes.Trim(buffer, "\x00")
		}
	}
}

func BuildListener(port uint16) (*net.UDPConn, chan int, chan []byte) {
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

func BuildGodspeed(port uint16, autoTruncate bool) (g *godspeed.Godspeed, err error) {
	g, err = godspeed.New("127.0.0.1", port, autoTruncate)
	return
}

func BuildAsyncGodspeed(port uint16, autoTruncate bool) (g *godspeed.AsyncGodspeed, err error) {
	g, err = godspeed.NewAsync("127.0.0.1", port, autoTruncate)
	return
}

func NoGo(a, b []byte) string {
	return fmt.Sprintf("%v (%v) is not equal to %v (%v)", string(a), a, string(b), b)
}

// Copyright 2014 PagerDuty, Inc, et al. All rights reserved.
// Use of this source code is governed by the BSD 3-Clause
// license that can be found in the LICENSE file.

package godspeed_test

import "github.com/PagerDuty/godspeed"

func ExampleGodspeed_Send() {
	g, err := godspeed.NewDefault()

	if err != nil {
		// handle error
	}

	tags := []string{"example"}

	err = g.Send("example.stat", "g", 1, 1, tags)

	if err != nil {
		// handle error
	}
}

func ExampleGodspeed_Count() {
	g, err := godspeed.NewDefault()

	if err != nil {
		// handle error
	}

	err = g.Count("example.count", 1, nil)

	if err != nil {
		// handle error
	}
}

func ExampleGodspeed_Incr() {
	g, _ := godspeed.NewDefault()

	err := g.Incr("example.counter", nil)

	if err != nil {
		// handle error
	}
}

func ExampleGodspeed_Gauge() {
	g, err := godspeed.NewDefault()

	if err != nil {
		// handle error
	}

	err = g.Gauge("example.gauge", 1, nil)

	if err != nil {
		// handle error
	}
}

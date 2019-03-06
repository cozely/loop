// Copyright 2013-2019 Laurent Moussault <laurent.moussault@gmail.com>
// SPDX-License-Identifier: BSD-2-Clause

package loop

import (
	"errors"
	"time"
)

////////////////////////////////////////////////////////////////////////////////

// State represents a state of the loop.
type State interface {
	Enter()
	Leave()
	React()
	Update()
	Render()
}

////////////////////////////////////////////////////////////////////////////////

var (
	state State
	next  State
)

// Running returns true if the loop is running, i.e. when called from inside
// loop.Run.
func Running() bool {
	return state != nil
}

// Goto changes the loop state. The change takes place at next frame.
func Goto(l State) {
	next = l
}

// Stop requests the loop to stop.
func Stop() {
	next = nil
}

////////////////////////////////////////////////////////////////////////////////

var (
	step  = time.Second / 60
	delta time.Duration
	lag   time.Duration
)

// Step returns the time between two consecutive updates. It is a
// fixed value, that only changes when configured with //TODO
//
// See also Delta and Lag.
func Step() time.Duration {
	return step
}

func TimeStep(s time.Duration) Option {
	return func(*private) error {
		step = s
		return nil
	}
}

// Delta returns the time elapsed between the frame to be rendered
// and the previous one.
//
// See also Step and Lag.
func Delta() time.Duration {
	return delta
}

// Lag returns the time elapsed between the last Update and the frame
// being rendered. It should be used during Render to extrapolate (or
// interpolate) the game state.
//
// See also Step and Delta.
func Lag() time.Duration {
	return lag
}

////////////////////////////////////////////////////////////////////////////////

// Run the loop.
func Run(start State) (err error) {
	if state != nil {
		return errors.New("loop.Run: already running")
	}

	// Start
	start.Enter()
	start.React()
	start.Update()
	next = start

	t0 := time.Now()
	t1 := t0
	delta, lag = 0, 0

	// Loop
	for next != nil {
		state = next

		// Apply any pending configuration
		for _, o := range options {
			err := o(&private{})
			if err != nil {
				return err
			}
		}

		// React, and (maybe) Update
		if lag < step {
			state.React()
		}
		for lag >= step {
			lag -= step
			state.React()
			state.Update()
		}

		// Render
		state.Render()

		t0 = t1
		t1 = time.Now()
		delta = t1.Sub(t0)
		stats()
		if delta > 4*step {
			// Prevent "spiral of death" when Render cannot keep up with Update
			delta = 4 * step
		}
		lag += delta
	}

	// Stop
	state.Leave()
	state = nil
	return nil
}

////////////////////////////////////////////////////////////////////////////////

const (
	statsInterval = time.Second / 4
	xrunThreshold = 17 * time.Millisecond
)

var (
	frametime float64
	xruns     int
	interval  struct {
		frames int
		time   time.Duration
		xruns  int
	}
)

// Stats returns the frametime durations of frames; it is updated 4 times per
// second. It also returns the number of overruns (i.e. frame time longer than
// the threshold) during the last measurment interval.
func Stats() (frametime float64, overruns int) {
	return frametime, xruns
}

func stats() {
	interval.frames++
	interval.time += delta
	if delta > xrunThreshold {
		interval.xruns++
	}
	if interval.time >= statsInterval {
		frametime = float64(interval.time) / float64(interval.frames)
		xruns = interval.xruns
		interval.time, interval.frames, interval.xruns = 0, 0, 0
	}
}

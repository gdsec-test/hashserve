// Package backoff provides utilities to provide backoff capabilities to actions.
//
// This may be used in situations where exponential backoff is needed to multiplicatively
// decrease the rate of some process, in order to gradually find an acceptable rate.
package backoff

import (
	"math/rand"
	"time"
)

var (
	defaultJitter     = 0.30
	defaultAttempts   = 15
	defaultMultiplier = 1.5
)

var (
	backoffTab []time.Duration
)

func init() {
	backoffTab = make([]time.Duration, defaultAttempts)
	cur := float64(100 * time.Millisecond)
	for i := 0; i < defaultAttempts; i++ {
		cur *= defaultMultiplier
		backoffTab[i] = time.Duration(cur)
	}
}

// Jitter will call JitterRange with the default jitter.
func Jitter(d time.Duration) time.Duration {
	return JitterRange(d, defaultJitter)
}

// JitterRange will add random jitter within the range [+m, -m) to d.
func JitterRange(d time.Duration, jitter float64) time.Duration {
	jit := 1 + jitter*(rand.Float64()*2-1)
	return time.Duration(jit * float64(d))
}

// Get will apply Jitter to the result of a call to Const.
func Get(attempts int) time.Duration { return Jitter(Const(attempts)) }

// Const will return the amount of time a caller should wait after the given
// number of consecutive failed attempts.
func Const(attempts int) time.Duration {
	if attempts >= len(backoffTab) {
		attempts = len(backoffTab) - 1
	}
	return backoffTab[attempts]
}

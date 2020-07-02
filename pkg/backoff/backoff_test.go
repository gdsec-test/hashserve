package backoff

import (
	"testing"
	"time"
)

func TestBackoff(t *testing.T) {
	t.Run(`Table`, func(t *testing.T) {
		for idx, exp := range backoffTab {
			t.Logf(`backoff test from %v attempts expect %v`, idx, exp)

			got := Get(idx)
			testJitter(t, exp, got)
		}
	})
	t.Run(`Bounds`, func(t *testing.T) {
		l := len(backoffTab) - 1
		for i := l - 2; i < l+10; i++ {
			got := Get(i)
			at := i
			if at > l {
				at = l
			}
			testJitter(t, backoffTab[at], got)
		}
	})
}

func TestConst(t *testing.T) {
	for idx, exp := range backoffTab {
		t.Logf(`backoff test from %v attempts expect %v`, idx, exp)

		got := Const(idx)
		testJitter(t, exp, got)
	}

	l := len(backoffTab) - 1
	for i := l - 2; i < l+10; i++ {
		got := Const(i)
		at := i
		if at > l {
			at = l
		}
		testJitter(t, backoffTab[at], got)
	}
}

func TestJitter(t *testing.T) {
	for d := time.Millisecond * 200; d < time.Minute; d *= 2 {
		for i := 0; i < 1000; i++ {
			t.Logf(`jitter duration %v try #%v`, d, i)
			testJitter(t, d, Jitter(d))
		}
	}
}

func testJitter(t *testing.T, from, got time.Duration) {
	max := time.Duration(float64(from) * defaultJitter)
	if exp := from + max; exp < got {
		t.Errorf(`exp value less than %v; got %v`, exp, got)
	}
	if exp := from - max; exp > got {
		t.Errorf(`exp value greater than %v; got %v`, exp, got)
	}
}

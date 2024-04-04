package history

import (
	"fmt"
	"math"
	"testing"
	"time"
)

func eq(t *testing.T, a, b float64) bool {
	t.Logf("want: %f, got: %f\n", a, b)
	return math.Abs(a-b) < 0.0001
}

func TestHistory(t *testing.T) {
	n := time.Now()

	now = func() time.Time {
		return n
	}

	h := New(10 * time.Minute, "test")
	fmt.Printf("live time; %v\n", h.LiveTime)

	h.Append(1)
	h.Append(2)
	h.Append(3)
	h.Append(4)
	n = n.Add(20 * time.Minute)
	h.Append(5)
	h.Append(6)
	n = n.Add(5 * time.Minute)
	h.Append(7)
	h.Append(8)

	t.Logf("\n%s\n", h)

	if !eq(t, h.Average(10*time.Minute), avg(5, 6, 7, 8)) {
		t.Error("Average returned the wrong value")
	}

	n = n.Add(6 * time.Minute)

	t.Logf("\n%s\n", h)

	if !eq(t, h.Average(10*time.Minute), avg(7, 8)) {
		t.Error("Average returned the wrong value")
	}
}

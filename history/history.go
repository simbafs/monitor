// history is a package that provides a simple history of float64 values. It is used to store a history of values and calculate the average of those values over a given time period. Every operation will automatically remove the records that out of date, so that only the most recent values are kept. The history package is tested in history/history_test.go.
package history

import (
	"fmt"
	"time"
)

var now = time.Now

// History keep records that in live time. It implement String interface
type History struct {
	LiveTime time.Duration
	data     []float64
	time     []time.Time
}

// New creates a new History with the given liveTime.
func New(liveTime time.Duration) *History {
	return &History{
		LiveTime: liveTime,
		data:     []float64{},
		time:     []time.Time{},
	}
}

// after rettuenrs the index of the first element in the history that is after t
func (h *History) after(t time.Time) int {
	for i, v := range h.time {
		if v.After(t) {
			return i
		}
	}
	return 0
}

// update remove the records that are out of date
func (h *History) update() {
	end := now().Add(-h.LiveTime)
	i := h.after(end)

	h.data = h.data[i:]
	h.time = h.time[i:]
}

// Append adds a new data point to the history
func (h *History) Append(data float64) {
	h.data = append(h.data, data)
	h.time = append(h.time, now())
	h.update()
}

// avg is a helper function, which calculates the average of the given data
func avg(data ...float64) float64 {
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}

// Average returns the average of the data in the history over the given duration
func (h *History) Average(duration time.Duration) float64 {
	h.update()
	i := h.after(now().Add(-duration))
	return avg(h.data[i:]...)
}

func (h *History) String() string {
	h.update()
	n := now()
	s := ""
	for i := 0; i < len(h.data); i++ {
		s += fmt.Sprintf("%f %s\n", h.data[i], h.time[i].Sub(n))
	}

	return s
}

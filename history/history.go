// history is a package that provides a simple history of float64 values. It is used to store a history of values and calculate the average of those values over a given time period. Every operation will automatically remove the records that out of date, so that only the most recent values are kept. The history package is tested in history/history_test.go.
package history

import (
	"fmt"
	"time"
)

var now = time.Now

type Record[T any] struct {
	Data T
	Time time.Time
}

// History keep records that in live time. It implement String interface
type History struct {
	LiveTime time.Duration
	records  []Record[float64]
}

// New creates a new History with the given liveTime.
func New(liveTime time.Duration) *History {
	return &History{
		LiveTime: liveTime,
		records:  []Record[float64]{},
	}
}

// Len returns the number of records in the history
func (h *History) Len() int {
	return len(h.records)
}

// Records returns the records in the history
func (h *History) Records() []Record[float64] {
	return h.records
}

// Data returns the data in the history
func (h *History) Datas() []float64 {
	d := make([]float64, h.Len())
	for i, v := range h.records {
		d[i] = v.Data
	}
	return d
}

// Times returns the times of the records in the history
func (h *History) Times() []time.Time {
	t := make([]time.Time, h.Len())
	for i, v := range h.records {
		t[i] = v.Time
	}
	return t
}

// after rettuenrs the index of the first element in the history that is after t
func (h *History) after(t time.Time) int {
	for i, v := range h.records {
		if v.Time.After(t) {
			return i
		}
	}
	return 0
}

// update remove the records that are out of date
func (h *History) update() {
	end := now().Add(-h.LiveTime)
	i := h.after(end)

	h.records = h.records[i:]
}

// Append adds a new data point to the history
func (h *History) Append(data float64) {
	h.records = append(h.records, Record[float64]{
		Data: data,
		Time: now(),
	})
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
	return avg(h.Datas()[i:]...)
}

func (h *History) String() string {
	h.update()
	n := now()
	s := ""
	for _, v := range h.records {
		s += fmt.Sprintf("%f %s\n", v.Data, v.Time.Sub(n))
	}

	return s
}

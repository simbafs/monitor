package history

type History struct {
	Capacity int
	Data     []float64
}

func New(capacity int) *History {
	return &History{
		Capacity: capacity,
		Data:     make([]float64, capacity),
	}
}

func (h *History) Append(n float64) {
	h.Data = append(h.Data, n)
	if len(h.Data) > h.Capacity {
		h.Data = h.Data[1:]
	}
}

func (h *History) Average() float64 {
	if len(h.Data) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range h.Data {
		sum += v
	}
	return sum / float64(len(h.Data))
}

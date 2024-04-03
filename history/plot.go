package history

import (
	"io"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

func history2Points(history *History) plotter.XYs {
	pts := make(plotter.XYs, history.Len())
	for i, r := range history.Records() {
		pts[i].X = float64(r.Time.Unix())
		pts[i].Y = r.Data

	}
	return pts
}

func Plot(histories ...*History) (io.WriterTo, error) {
	// xticks defines how we convert and display time.Time values.
	xticks := plot.TimeTicks{Format: "2006-01-02\n15:04"}

	p := plot.New()
	p.Title.Text = "Resource Usage"
	p.X.Tick.Marker = xticks
	p.Y.Label.Text = "Persentage"
	p.Add(plotter.NewGrid())

	lines := []interface{}{}

	for _, h := range histories {
		data := history2Points(h)
		lines = append(lines, h.Name, data)
	}

	if err := plotutil.AddLinePoints(p, lines...); err != nil {
		return nil, err
	}

	writerTo, err := p.WriterTo(10*vg.Centimeter, 5*vg.Centimeter, "png")
	if err != nil {
		return nil, err
	}

	return writerTo, nil
}

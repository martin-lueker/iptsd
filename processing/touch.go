package processing

import (
	"math"
)

type TouchProcessor struct {
	InvertX bool
	InvertY bool

	MaxTouchPoints int

	inputs   []TouchInput
	contacts []Contact

	last []TouchInput

	distances [][]float64
	indices   [][]int

	freeIndices []bool

	heatmapCache map[int]*Heatmap
}

type TouchInput struct {
	X      int
	Y      int
	Index  int
	IsPalm bool

	contact *Contact
}

func (ti TouchInput) Dist(other TouchInput) float64 {
	dx := float64(ti.X - other.X)
	dy := float64(ti.Y - other.Y)
	return math.Sqrt(dx*dx + dy*dy)
}

func (tp *TouchProcessor) GetHeatmap(width int, height int) *Heatmap {
	if tp.heatmapCache == nil {
		tp.heatmapCache = make(map[int]*Heatmap)
	}

	size := width * height

	hm, ok := tp.heatmapCache[size]
	if !ok {
		hm = &Heatmap{}
		hm.Data = make([]byte, size)
		hm.Visited = make([]bool, size)

		tp.heatmapCache[size] = hm
	}

	hm.Width = width
	hm.Height = height

	return hm
}

func (tp *TouchProcessor) Save() {
	for i := 0; i < tp.MaxTouchPoints; i++ {
		tp.freeIndices[i] = true
	}

	for i := 0; i < tp.MaxTouchPoints; i++ {
		tp.last[i] = tp.inputs[i]

		if tp.inputs[i].Index == -1 {
			continue
		}

		tp.freeIndices[tp.inputs[i].Index] = false
	}
}

func (tp *TouchProcessor) Inputs(hm *Heatmap) []TouchInput {
	if tp.inputs == nil {
		tp.inputs = make([]TouchInput, tp.MaxTouchPoints)
		tp.contacts = make([]Contact, tp.MaxTouchPoints)
	}

	avg := byte(hm.Average())
	for i := 0; i < len(hm.Data); i++ {
		if hm.Data[i] < avg {
			hm.Data[i] = avg - hm.Data[i]
		} else {
			hm.Data[i] = 0
		}
	}

	count := hm.Contacts(tp.contacts)
	GetPalms(tp.contacts, count)

	for i := 0; i < count; i++ {
		x, y := tp.contacts[i].Mean()

		x /= float32(hm.Width - 1)
		y /= float32(hm.Height - 1)

		if tp.InvertX {
			x = 1 - x
		}

		if tp.InvertY {
			y = 1 - y
		}

		tp.inputs[i] = TouchInput{
			X:       int(x * 9600),
			Y:       int(y * 7200),
			Index:   i,
			IsPalm:  tp.contacts[i].isPalm,
			contact: &tp.contacts[i],
		}
	}

	for i := count; i < tp.MaxTouchPoints; i++ {
		tp.inputs[i] = TouchInput{
			X:       0,
			Y:       0,
			Index:   -1,
			contact: &tp.contacts[i],
		}
	}

	if tp.last == nil {
		tp.last = make([]TouchInput, tp.MaxTouchPoints)
		tp.distances = make([][]float64, tp.MaxTouchPoints)
		tp.indices = make([][]int, tp.MaxTouchPoints)
		tp.freeIndices = make([]bool, tp.MaxTouchPoints)

		for i := 0; i < tp.MaxTouchPoints; i++ {
			tp.distances[i] = make([]float64, tp.MaxTouchPoints)
			tp.indices[i] = make([]int, tp.MaxTouchPoints)
		}

		tp.Save()
		return tp.inputs
	}

	tp.TrackFingers(count)

	tp.Save()
	return tp.inputs
}
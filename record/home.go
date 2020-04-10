package record

import ()

type Home struct {
	longitude float64
	latitude  float64
	height    float32
}

func (h *Home) Height() float64 {
	return float64(h.height) / 10
}

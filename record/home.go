package record

import ()

type Home struct {
	Coodinate

	height float32
}

func (h *Home) Height() float64 {
	return float64(h.height) / 10
}

package record

import (
	"math"
)

type Coodinate struct {
	longitude float64
	latitude  float64
}

func (c *Coodinate) Longitude() float64 {
	return c.longitude * 180 / math.Pi
}

func (c *Coodinate) Latitude() float64 {
	return c.latitude * 180 / math.Pi
}

func (c *Coodinate) Valid() bool {
	inRange := math.Abs(c.longitude) <= 180.0 && math.Abs(c.latitude) <= 180.0
	return inRange && c.longitude != 0 && c.latitude != 0
}

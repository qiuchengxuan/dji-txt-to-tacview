package record

import (
	"math"
)

type OSD struct {
	Coodinate

	height  int16
	xSpeed  int16
	ySpeed  int16
	zSpeed  int16
	pitch   int16
	roll    int16
	yaw     int16
	_       [12]byte
	flyTime uint16
}

func (o *OSD) FlyTime() float64 {
	return float64(o.flyTime) / 10
}

func (o *OSD) Height() float64 {
	return float64(o.height) / 10
}

func (o *OSD) Pitch() float64 {
	return float64(o.pitch) / 10
}

func (o *OSD) Roll() float64 {
	return float64(o.roll) / 10
}

func (o *OSD) Yaw() float64 {
	return float64(o.yaw) / 10
}

func (o *OSD) Speed() float64 {
	x, y, z := int(o.xSpeed), int(o.ySpeed), int(o.zSpeed)
	return math.Sqrt(float64(x*x+y*y+z*z)) / 10
}

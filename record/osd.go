package record

import (
	"math"
)

type OSD struct {
	longitude float64
	latitude  float64
	height    int16
	xSpeed    int16
	ySpeed    int16
	zSpeed    int16
	pitch     int16
	roll      int16
	yaw       int16
	_         [12]byte
	flyTime   uint16
}

func (o *OSD) FlyTime() float64 {
	return float64(o.flyTime) / 10
}

func (o *OSD) Longitude() float64 {
	return o.longitude * 180 / math.Pi
}

func (o *OSD) Latitude() float64 {
	return o.latitude * 180 / math.Pi
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

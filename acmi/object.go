package acmi

import (
	"fmt"
	"time"
)

type Objectable interface {
	ID() int
}

type ReferenceObject struct {
	Longitude     int       `acmi:"Longitude"`
	Latitude      int       `acmi:"Latitude"`
	DataSource    string    `acmi:"DataSource"`   // e.g. Mavic Mini
	DataRecorder  string    `acmi:"DataRecorder"` // e.g. DJI Fly
	ReferenceTime time.Time `acmi:"ReferenceTime"`
}

func (o *ReferenceObject) ID() int {
	return 0
}

type ObjectType = string

const (
	ObjectTypeAirRotorcraft ObjectType = "Air+Rotorcraft"
	ObjectTypeHuman         ObjectType = "Ground+Light+Human+Infantry"
)

type Coordinate struct {
	Longitude float64
	Latitude  float64
	Altitude  float64
}

type Attitude struct {
	Roll  float64
	Pitch float64
	Yaw   float64
}

type Transform struct {
	Coordinate
	*Attitude
}

func (t *Transform) String() string {
	coordination := fmt.Sprintf("%.6f|%.6f|%.1f", t.Longitude, t.Latitude, t.Altitude)
	if t.Attitude != nil {
		return coordination + fmt.Sprintf("|%.1f|%.1f|%.1f", t.Roll, t.Pitch, t.Yaw)
	}
	return coordination
}

type Object struct {
	Id        int
	Transform Transform  `acmi:"T"`
	Name      string     `acmi:"Name"`
	Type      ObjectType `acmi:"Type"`
}

func (o *Object) ID() int {
	return o.Id
}

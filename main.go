package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"time"
	"unsafe"

	"github.com/qiuchengxuan/dji-txt-to-tacview/acmi"
	. "github.com/qiuchengxuan/dji-txt-to-tacview/record"
)

type Decoder struct {
	io.ReadSeeker
}

func makeSlice(start uintptr, length int) (data []byte) {
	slice := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	slice.Data = start
	slice.Len = length
	slice.Cap = length
	return
}

func (d *Decoder) Decode(x interface{}, options ...int) error {
	var size int
	if len(options) > 0 {
		size = options[0]
	} else {
		size = int(reflect.Indirect(reflect.ValueOf(x)).Type().Size())
	}
	bytes := makeSlice(reflect.ValueOf(x).Pointer(), size)
	n, err := d.Read(bytes)
	if err != nil {
		return err
	}
	if n != size {
		return io.EOF
	}
	return nil
}

type appOsVersion = uint8

const (
	appOsVersionGO  appOsVersion = 6
	appOsVersionFly appOsVersion = 12
)

type header struct {
	endRecordOffset uint64
	detailsLength   uint16
	appOsVersion    appOsVersion
	encrypted       bool
}

const headerSize int = 100

type detail struct {
	// ignore leading 91 bytes
	timestamp time.Time
	longitude float64
	latitude  float64
	// ignore following 28 bytes

	// ignore leading 137 bytes
	aircraftName string // 32 bytes
	// ignore following 64 bytes
}

func (d *detail) decode(readSeeker io.ReadSeeker) {
	var buffer [32]byte
	readSeeker.Seek(91, io.SeekCurrent)
	readSeeker.Read(buffer[:24])
	timestamp := *(*uint64)(unsafe.Pointer(&buffer[0]))
	d.timestamp = time.Unix(int64(timestamp/1000), int64(timestamp%1000)*1000)
	d.longitude = *(*float64)(unsafe.Pointer(&buffer[8]))
	d.latitude = *(*float64)(unsafe.Pointer(&buffer[16]))
	readSeeker.Seek(28+137, io.SeekCurrent)
	readSeeker.Read(buffer[:])
	d.aircraftName = string(buffer[:bytes.IndexByte(buffer[:], 0)])
	return
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Filename not provided")
	}
	f, err := os.Open(os.Args[len(os.Args)-1])
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	decoder := Decoder{f}
	var header header
	if err := decoder.Decode(&header); err != nil {
		log.Fatal(err)
	}

	if header.appOsVersion < appOsVersionGO {
		log.Fatal("Unsupported os version")
	}

	w := acmi.AcmiWriter{Writer: os.Stdout}
	w.WriteHeader()

	var offset int64
	if header.appOsVersion >= appOsVersionFly { // details follows header
		offset = int64(headerSize)
	} else {
		offset = int64(header.endRecordOffset) + 1
	}
	if _, err := decoder.Seek(offset, io.SeekStart); err != nil {
		log.Fatal(err)
	}
	var detail detail
	detail.decode(decoder)

	var dataRecorder string
	if header.appOsVersion == appOsVersionFly {
		dataRecorder = "DJI Fly"
	} else {
		dataRecorder = "DJI GO"
	}
	w.Dump(&acmi.ReferenceObject{
		Longitude:     int(detail.longitude),
		Latitude:      int(detail.latitude),
		DataSource:    detail.aircraftName,
		DataRecorder:  dataRecorder,
		ReferenceTime: detail.timestamp,
	})

	if header.appOsVersion >= appOsVersionFly {
		offset = int64(headerSize) + int64(header.detailsLength)
	} else {
		offset = int64(headerSize)
	}
	if _, err := decoder.Seek(int64(offset), io.SeekStart); err != nil {
		log.Fatal(err)
	}

	firstOSD, firstHome := true, true

	homeHeight := 0.0
	recordReader := NewRecordReader(decoder)
	for offset <= int64(header.endRecordOffset) {
		record := recordReader.Next()
		offset = recordReader.Offset()
		if record == nil {
			continue
		}

		switch r := record.(type) {
		case *OSD:
			if !r.Coodinate.Valid() {
				continue
			}
			w.Write([]byte(fmt.Sprintf("#%g\n", r.FlyTime())))
			t := acmi.Transform{
				Coordinate: acmi.Coordinate{
					Longitude: r.Longitude(),
					Latitude:  r.Latitude(),
					Altitude:  r.Height() + homeHeight,
				},
				Attitude: &acmi.Attitude{Roll: r.Roll(), Pitch: r.Pitch(), Yaw: r.Yaw()},
			}
			object := acmi.Object{Id: 1, Transform: t}
			if firstOSD {
				object.Name = detail.aircraftName
				object.Type = acmi.ObjectTypeAirRotorcraft
				firstOSD = false
			}
			w.Dump(&object)
		case *Home:
			homeHeight = r.Height()
			if !r.Coodinate.Valid() {
				continue
			}
			t := acmi.Transform{
				Coordinate: acmi.Coordinate{
					Longitude: r.Longitude(),
					Latitude:  r.Latitude(),
					Altitude:  r.Height(),
				},
			}
			object := acmi.Object{Id: 2, Transform: t}
			if firstHome {
				object.Name = "home"
				object.Type = acmi.ObjectTypeHuman
				firstHome = false
			}
			w.Dump(&object)
		}
	}
}

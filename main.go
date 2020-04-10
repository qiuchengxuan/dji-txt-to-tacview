package main

import (
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"strconv"
	"unsafe"

	. "github.com/qiuchengxuan/dji-txt-to-tacview/record"
	. "github.com/qiuchengxuan/dji-txt-to-tacview/unscramble"
)

const (
	fieldTime int = iota
	fieldLongitude
	fieldLatitude
	fieldAltitude
	fieldRoll
	fieldPitch
	fieldYaw
	fieldIAS
)

var fieldName []string = []string{
	fieldTime:      "Time",
	fieldLongitude: "Longitude",
	fieldLatitude:  "Latitude",
	fieldAltitude:  "Altitude",
	fieldRoll:      "Roll (deg)",
	fieldPitch:     "Pitch (deg)",
	fieldYaw:       "Yaw (deg)",
	fieldIAS:       "IAS",
}

const (
	startOfImage = 0xFFD8
	endOfImage   = 0xFFD9
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
	appOsVersionFly appOsVersion = 12
)

type header struct {
	endRecordOffset uint64
	detailsLength   uint16
	appOsVersion    appOsVersion
	encrypted       bool
}

const headerSize int = 100

func handleJPEG(readSeeker io.ReadSeeker) (int64, error) {
	size := 0
	var bytes [2]byte
	readSeeker.Read(bytes[:])
	size += 2
	if binary.BigEndian.Uint16(bytes[:]) == startOfImage {
		for {
			readSeeker.Read(bytes[:])
			size += 2
			if binary.BigEndian.Uint16(bytes[:]) != endOfImage {
				continue
			}
			readSeeker.Read(bytes[:])
			size += 2
			if binary.BigEndian.Uint16(bytes[:]) != startOfImage {
				break
			}
		}
	}
	return readSeeker.Seek(-2, io.SeekCurrent)
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

	skip := headerSize - int(unsafe.Sizeof(header))
	if header.appOsVersion == appOsVersionFly {
		skip += int(header.detailsLength)
	}
	if _, err := decoder.Seek(int64(skip), io.SeekCurrent); err != nil {
		log.Fatal(err)
	}

	w := csv.NewWriter(os.Stdout)
	w.Write(fieldName)

	homeHeight := float64(0)
	offset := 0
	var buffer []byte
	records := [fieldIAS + 1]string{}
	for offset <= int(header.endRecordOffset) {
		var record Record
		if err := decoder.Decode(&record, 2); err != nil {
			if err == io.EOF {
				break
			} else {
				log.Fatal(err)
			}
		}

		if record.Type == RecordTypeJPEG || (record.Type == 0xFF && record.Length == 0xD8) {
			if record.Type == RecordTypeJPEG {
				decoder.Seek(2, io.SeekCurrent)
			} else { // old log JPEG formats
				decoder.Seek(-2, io.SeekCurrent)
			}
			off, _ := handleJPEG(decoder)
			offset = int(off)
			continue
		}

		readSize := int(record.Length) + 1 // including end-of-record
		if len(buffer) < readSize {
			buffer = make([]byte, readSize)
		}
		decoder.Read(buffer[:readSize])
		recordBytes := buffer[:record.Length]
		Unscramble(recordBytes, record.Type)

		for buffer[record.Length] != 0xFF { // locating end-of-record
			decoder.Read(buffer[record.Length:readSize])
		}
		off, _ := decoder.Seek(0, io.SeekCurrent)
		offset = int(off)

		if record.Type != RecordTypeOSD && record.Type != RecordTypeHome {
			continue
		}

		if record.Type == RecordTypeHome {
			homeHeight = (*Home)(unsafe.Pointer(&recordBytes[0])).Height()
			continue
		}

		osd := (*OSD)(unsafe.Pointer(&recordBytes[0]))
		if osd.Longitude() == 0 || osd.Latitude() == 0 {
			continue
		}

		records[fieldTime] = strconv.FormatFloat(osd.FlyTime(), 'f', -1, 64)
		records[fieldLongitude] = fmt.Sprintf("%f", osd.Longitude())
		records[fieldLatitude] = fmt.Sprintf("%f", osd.Latitude())
		records[fieldAltitude] = strconv.FormatFloat(osd.Height()+homeHeight, 'f', -1, 64)
		records[fieldRoll] = strconv.FormatFloat(osd.Roll(), 'f', -1, 64)
		records[fieldPitch] = strconv.FormatFloat(osd.Pitch(), 'f', -1, 64)
		records[fieldYaw] = strconv.FormatFloat(osd.Yaw(), 'f', -1, 64)
		records[fieldIAS] = strconv.FormatFloat(osd.Speed(), 'f', -1, 64)
		w.Write(records[:])
	}
	w.Flush()
}

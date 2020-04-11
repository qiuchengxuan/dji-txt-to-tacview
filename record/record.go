package record

import (
	"encoding/binary"
	"io"
	"unsafe"

	. "github.com/qiuchengxuan/dji-txt-to-tacview/unscramble"
)

type RecordType = uint8

const (
	RecordTypeOSD  = 1
	RecordTypeHome = 2
	RecordTypeJPEG = 57
)

type Record struct {
	Type   RecordType
	Length uint8
}

type RecordReader struct {
	readSeeker io.ReadSeeker
	buffer     []byte
}

const (
	startOfImage = 0xFFD8
	endOfImage   = 0xFFD9
)

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
		return readSeeker.Seek(-2, io.SeekCurrent)
	}
	return readSeeker.Seek(0, io.SeekCurrent)
}

func (d *RecordReader) Next() interface{} {
	d.readSeeker.Read(d.buffer[:2])
	record := *(*Record)(unsafe.Pointer(&d.buffer[0]))

	if record.Type == RecordTypeJPEG || binary.BigEndian.Uint16(d.buffer) == startOfImage {
		if record.Type == RecordTypeJPEG {
			d.readSeeker.Seek(2, io.SeekCurrent) // skip following 2 byte of 0
		} else { // old log JPEG formats
			d.readSeeker.Seek(-2, io.SeekCurrent) // rewind before SOI
		}
		handleJPEG(d.readSeeker)
		return nil
	}

	readSize := int(record.Length) + 1 // including end-of-record
	if len(d.buffer) < readSize {
		d.buffer = make([]byte, readSize)
	}
	buffer := d.buffer
	d.readSeeker.Read(buffer[:readSize])
	recordBytes := buffer[:record.Length]
	Unscramble(recordBytes, record.Type)

	for buffer[record.Length] != 0xFF { // locating end-of-record
		d.readSeeker.Read(buffer[record.Length:readSize])
	}
	switch record.Type {
	case RecordTypeOSD:
		return (*OSD)(unsafe.Pointer(&recordBytes[0]))
	case RecordTypeHome:
		return (*Home)(unsafe.Pointer(&recordBytes[0]))
	}
	return nil
}

func (d *RecordReader) Offset() int64 {
	offset, _ := d.readSeeker.Seek(0, io.SeekCurrent)
	return offset
}

func NewRecordReader(readSeeker io.ReadSeeker) RecordReader {
	return RecordReader{readSeeker, make([]byte, 2)}
}

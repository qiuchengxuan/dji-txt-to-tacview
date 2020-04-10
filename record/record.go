package record

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

package acmi

import (
	"bytes"
	"testing"
	"time"

	"github.com/sergi/go-diff/diffmatchpatch"
)

const expected = `
FileType=text/acmi/tacview
FileVersion=2.1
0,Longitude=116,Latitude=40,DataSource=DJI Mavic Mini,DataRecorder=DJI Fly,ReferenceTime=1970-01-01T08:00:00Z
#0.1
1,T=0|0|0|0|0|0,Name=DJI Mavic Mini,Type=Air+Minor+Rotorcraft
1,T=0|0|0|0|0|0
`

func TestAcmi(t *testing.T) {
	dmp := diffmatchpatch.New()
	var buffer bytes.Buffer
	w := AcmiWriter{&buffer}
	w.WriteHeader()
	w.Dump(&ReferenceObject{116, 40, "DJI Mavic Mini", "DJI Fly", time.Unix(0, 0)})
	w.Write([]byte("#0.1\n"))
	w.Dump(&Object{1, Transform{}, "DJI Mavic Mini", ObjectTypeAirRotorcraft})
	w.Dump(&Object{Id: 1, Transform: Transform{}})
	actual := "\n" + buffer.String()
	if actual != expected {
		t.Fatal(dmp.DiffPrettyText(dmp.DiffMain(expected, actual, false)))
	}
}

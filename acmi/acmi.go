package acmi

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"
)

const header = `
FileType=text/acmi/tacview
FileVersion=2.1
`

type AcmiWriter struct{ io.Writer }

func (w AcmiWriter) WriteHeader() error {
	_, err := w.Writer.Write([]byte(strings.TrimLeft(header, "\n")))
	return err
}

func fieldToString(tag string, field reflect.Value) string {
	switch v := field.Interface().(type) {
	case float64:
		return fmt.Sprintf("%s=%g,", tag, v)
	case string:
		if v == "" {
			return v
		}
		return fmt.Sprintf("%s=%s,", tag, v)
	case time.Time:
		return fmt.Sprintf("%s=%s,", tag, v.Format("2006-01-02T15:04:05Z"))
	default:
		if field.Kind() == reflect.Struct {
			if v, ok := field.Addr().Interface().(fmt.Stringer); ok {
				return fmt.Sprintf("%s=%s,", tag, v.String())
			}
		}
		return fmt.Sprintf("%s=%v,", tag, v)
	}
}

func (w AcmiWriter) Dump(object Objectable) error {
	value := reflect.Indirect(reflect.ValueOf(object))
	valueType := value.Type()
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%x,", object.ID()))
	for i := 0; i < value.NumField(); i++ {
		tag := valueType.Field(i).Tag.Get("acmi")
		if tag == "" {
			continue
		}
		buf.WriteString(fieldToString(tag, value.Field(i)))
	}
	buf.Truncate(buf.Len() - 1)
	buf.WriteRune('\n')
	_, err := w.Writer.Write(buf.Bytes())
	return err
}

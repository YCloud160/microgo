package xlog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Entry struct {
	Key         string
	Value       interface{}
	IsSkip      bool
	IsQuotation bool
}

func (f *Entry) String() string {
	return f.Key + "=" + f.ValueString()
}

func (f *Entry) ValueString() string {
	if f.Value == nil {
		return ""
	}
	switch v := f.Value.(type) {
	case nil:
		return "nil"
	case error:
		if v == nil {
			return "nil"
		}
		return v.Error()
	case *error:
		if v == nil {
			return "nil"
		}
		return (*v).Error()
	case string:
		return v
	case *string:
		if v == nil {
			return ""
		}
		return *v
	case []byte:
		return string(v)
	case *[]byte:
		if v == nil {
			return ""
		}
		return string(*v)
	case fmt.Stringer:
		if v == nil {
			return ""
		}
		return v.String()
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case *uint8:
		if v == nil {
			return "0"
		}
		return strconv.FormatUint(uint64(*v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case *uint16:
		if v == nil {
			return "0"
		}
		return strconv.FormatUint(uint64(*v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case *uint32:
		if v == nil {
			return "0"
		}
		return strconv.FormatUint(uint64(*v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case *uint64:
		if v == nil {
			return "0"
		}
		return strconv.FormatUint(*v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case *uint:
		if v == nil {
			return "0"
		}
		return strconv.FormatUint(uint64(*v), 10)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case *int8:
		if v == nil {
			return "0"
		}
		return strconv.FormatInt(int64(*v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case *int16:
		if v == nil {
			return "0"
		}
		return strconv.FormatInt(int64(*v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case *int32:
		if v == nil {
			return "0"
		}
		return strconv.FormatInt(int64(*v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case *int64:
		if v == nil {
			return "0"
		}
		return strconv.FormatInt(*v, 10)
	case int:
		return strconv.FormatInt(int64(v), 10)
	case *int:
		if v == nil {
			return "0"
		}
		return strconv.FormatInt(int64(*v), 10)
	case bool:
		return strconv.FormatBool(v)
	case *bool:
		if v == nil {
			return "false"
		}
		return strconv.FormatBool(*v)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case *float32:
		if v == nil {
			return "0.0"
		}
		return strconv.FormatFloat(float64(*v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case *float64:
		if v == nil {
			return "0.0"
		}
		return strconv.FormatFloat(*v, 'f', -1, 64)
	}

	if f.Value == nil {
		return "nil"
	}

	bs, err := json.Marshal(f.Value)
	if err != nil {
		return err.Error()
	}
	return fmt.Sprintf("%s", bs)
}

func Field(k string, v interface{}) *Entry {
	f := &Entry{
		Key:         k,
		Value:       v,
		IsSkip:      isSkip(v),
		IsQuotation: isQuotation(v),
	}
	return f
}

func addSkip(k string) *Entry {
	f := &Entry{
		Key:    k,
		Value:  "",
		IsSkip: true,
	}
	return f
}

func isSkip(val interface{}) bool {
	switch v := val.(type) {
	case error:
		if v == nil {
			return true
		}
	case *error:
		if v == nil {
			return true
		}
	}
	return false
}

func isQuotation(val interface{}) bool {
	switch val.(type) {
	case string, *string, []byte, *[]byte, error, *error, fmt.Stringer:
		return true
	case uint8, uint16, uint32, uint64, uint, int8, int16, int32, int64, int, float64, float32, bool:
		return false
	case *uint8, *uint16, *uint32, *uint64, *uint, *int8, *int16, *int32, *int64, *int, *float64, *float32, *bool:
		return false
	}
	valType := reflect.TypeOf(val)
	for valType.Kind() == reflect.Ptr {
		valType = valType.Elem()
	}
	switch valType.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.Struct:
		return false
	}
	return true
}

func EncodeJson(fields ...*Entry) string {
	buf := bytes.NewBuffer(make([]byte, 0, 512))
	buf.WriteByte('{')
	for i, field := range fields {
		if field.IsSkip {
			continue
		}
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteByte('"')
		buf.WriteString(field.Key)
		buf.WriteString("\":")
		if field.IsQuotation {
			buf.WriteByte('"')
			writeString(buf, field.ValueString())
			buf.WriteByte('"')
		} else {
			buf.WriteString(field.ValueString())
		}
	}
	buf.WriteByte('}')
	return buf.String()
}

func writeString(buf *bytes.Buffer, s string) {
	for _, c := range s {
		switch c {
		case '\n':
			buf.WriteString("\\n")
		case '\t':
			buf.WriteString("\\t")
		case '\r':
			buf.WriteString("\\r")
		case '\\':
			buf.WriteString("\\\\")
		default:
			buf.WriteRune(c)
		}
	}
}

func timeField() *Entry {
	return Field("timestamp", time.Now().Format("2006-01-02 15:04:05.000"))
}

func levelField(level Level) *Entry {
	return Field("level", level)
}

func caller() *Entry {
	_, filepath, line, ok := runtime.Caller(3)
	if ok {
		count := 0
		file := ""
		for {
			idx := strings.LastIndex(filepath, string(os.PathSeparator))
			if idx == -1 {
				break
			}
			if count == 0 {
				file = filepath[idx+1:]
			} else {
				file = filepath[idx+1:] + string(os.PathSeparator) + file
			}
			filepath = filepath[:idx]
			count++
			if count == 3 {
				break
			}
		}
		return Field("caller", fmt.Sprintf("%s:%d", file, line))
	}
	return addSkip("caller")
}

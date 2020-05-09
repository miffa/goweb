package newtype

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"
)

const (
	WEB_FORMAT              = "20060102150405"
	WEB_STD_FORMAT          = "2006-01-02 15:04:05"
	WEB_1970_FORMAT         = "2006-01-02T15:04:05Z"
	WEB_1970_FORMAT_DEFAULT = "0000-01-01T00:00:00Z"
	DB_STD_FORMAT           = "2006-01-02 15:04:05"
)

var errNilPtr = errors.New("destination pointer is nil")

func Timestamp() string {
	return time.Now().Format("20060102150405")

}

func Str2Time(str string) (time.Time, error) {
	return time.Parse(WEB_FORMAT, str)

}

var (
	ZeroTime TPaasTime
)

type TPaasTime time.Time

//type TPaasTime struct {
//	time.Time
//}

func TPaasTimeNow() TPaasTime {
	return TPaasTime(time.Now())
}

func (t TPaasTime) Time() time.Time {
	return time.Time(t)
}

func (t TPaasTime) Unix() int64 {
	return time.Time(t).Unix()
}

//func (t TPaasTime) MarshalJSON() ([]byte, error) {
//	if t.Unix() <= 0 {
//		return []byte("\"" + "\""), nil
//		//return []byte("\"" + WEB_1970_FORMAT_DEFAULT + "\""), nil
//
//	}
//	return []byte("\"" + time.Time(t).Format(WEB_1970_FORMAT) + "\""), nil
//}

func (t TPaasTime) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	//buf.WriteByte('"')
	b10 := []byte{}
	myms := t.Time().UnixNano() / 1000000 / 1000
	if myms < 0 {
		myms = 0
	}
	b10 = strconv.AppendInt(b10, myms, 10)
	buf.Write(b10)
	//buf.WriteByte('"')
	return buf.Bytes(), nil
}

func (t *TPaasTime) UnmarshalJSON(data []byte) error {
	timestr := string(data)
	timestr = strings.Trim(timestr, "\"")
	pos := strings.Index(timestr, ".")
	if pos == -1 {
	} else {
		timestr = string(timestr[0:pos]) + "Z"
	}

	tt, err := time.Parse(WEB_1970_FORMAT, timestr)
	if err != nil {
		return err
	}
	ttt := TPaasTime(tt)
	*t = ttt
	return err
}

func (t TPaasTime) Value() (driver.Value, error) {
	return time.Time(t), nil
}

func (t *TPaasTime) Scan(v interface{}) error {
	var mytie time.Time
	convertAssign(&mytie, v)
	ttt := TPaasTime(mytie)
	*t = ttt
	return nil
}

func convertAssign(dest, src interface{}) error {
	// Common cases, without reflect.
	switch s := src.(type) {
	case string:
		switch d := dest.(type) {
		case *string:
			if d == nil {
				return errNilPtr

			}
			*d = s
			return nil
		case *[]byte:
			if d == nil {
				return errNilPtr

			}
			*d = []byte(s)
			return nil
		case *sql.RawBytes:
			if d == nil {
				return errNilPtr

			}
			*d = append((*d)[:0], s...)
			return nil

		}
	case []byte:
		switch d := dest.(type) {
		case *string:
			if d == nil {
				return errNilPtr

			}
			*d = string(s)
			return nil
		case *interface{}:
			if d == nil {
				return errNilPtr

			}
			*d = cloneBytes(s)
			return nil
		case *[]byte:
			if d == nil {
				return errNilPtr

			}
			*d = cloneBytes(s)
			return nil
		case *sql.RawBytes:
			if d == nil {
				return errNilPtr

			}
			*d = s
			return nil

		}
	case time.Time:
		switch d := dest.(type) {
		case *time.Time:
			*d = s
			return nil
		case *string:
			*d = s.Format(time.RFC3339Nano)
			return nil
		case *[]byte:
			if d == nil {
				return errNilPtr

			}
			*d = []byte(s.Format(time.RFC3339Nano))
			return nil
		case *sql.RawBytes:
			if d == nil {
				return errNilPtr

			}
			*d = s.AppendFormat((*d)[:0], time.RFC3339Nano)
			return nil

		}
	case nil:
		switch d := dest.(type) {
		case *interface{}:
			if d == nil {
				return errNilPtr

			}
			*d = nil
			return nil
		case *[]byte:
			if d == nil {
				return errNilPtr

			}
			*d = nil
			return nil
		case *sql.RawBytes:
			if d == nil {
				return errNilPtr

			}
			*d = nil
			return nil

		}

	default:
		return nil
	}
	return nil
}

func cloneBytes(b []byte) []byte {
	if b == nil {
		return nil

	}
	c := make([]byte, len(b))
	copy(c, b)
	return c

}

type NullInt64 struct {
	sql.NullInt64
}

func (v NullInt64) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Int64)

	} else {
		return json.Marshal(0)

	}

}

func (v *NullInt64) UnmarshalJSON(data []byte) error {
	var s *int64
	if err := json.Unmarshal(data, &s); err != nil {
		return err

	}
	if s != nil {
		v.Valid = true
		v.Int64 = *s

	} else {
		v.Valid = false

	}
	return nil

}

type NullString struct {
	sql.NullString
}

func (v NullString) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.String)

	} else {
		return json.Marshal("")

	}

}

func (v *NullString) UnmarshalJSON(data []byte) error {
	var s *string
	if err := json.Unmarshal(data, &s); err != nil {
		return err

	}
	if s != nil {
		v.Valid = true
		v.String = *s

	} else {
		v.Valid = false

	}
	return nil
}

type Model struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt TPaasTime
	UpdatedAt TPaasTime
	DeletedAt *time.Time `sql:"index"`
}

package dbnulltype

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"strconv"
)

var (
	NullInt64_ZERO   NullInt64
	NullFloat64_ZERO NullFloat64
	NullString_ZERO  NullString
)

//
type NullString struct {
	sql.NullString
}

func (v NullString) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.String)
	} else {
		return json.Marshal("NULL")
	}
}

func (v *NullString) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("NULL")) {
		v.String = ""
		v.Valid = false
		return nil
	}

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

type NullInt64 struct {
	sql.NullInt64
}

func (v NullInt64) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(strconv.FormatInt(v.Int64, 10))

	} else {
		return json.Marshal("NULL")

	}

}

func (v *NullInt64) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("NULL")) {
		v.Int64 = 0
		v.Valid = false
		return nil
	}
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

type NullFloat64 struct {
	sql.NullFloat64
}

func (v NullFloat64) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(strconv.FormatFloat(v.Float64, 'f', -1, 32))

	} else {
		return json.Marshal(0)

	}

}

func (v *NullFloat64) UnmarshalJSON(data []byte) error {
	var s *float64
	if err := json.Unmarshal(data, &s); err != nil {
		return err

	}
	if s != nil {
		v.Valid = true
		v.Float64 = *s

	} else {
		v.Valid = false

	}
	return nil

}

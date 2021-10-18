/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package monetdb

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	mdb_CHAR      = "char"    // (L) character string with length L
	mdb_VARCHAR   = "varchar" // (L) string with atmost length L
	mdb_CLOB      = "clob"
	mdb_BLOB      = "blob"
	mdb_DECIMAL   = "decimal"  // (P,S)
	mdb_SMALLINT  = "smallint" // 16 bit integer
	mdb_INT       = "int"      // 32 bit integer
	mdb_BIGINT    = "bigint"   // 64 bit integer
	mdb_HUGEINT   = "hugeint"  // 64 bit integer
	mdb_SERIAL    = "serial"   // special 64 bit integer sequence generator
	mdb_REAL      = "real"     // 32 bit floating point
	mdb_DOUBLE    = "double"   // 64 bit floating point
	mdb_BOOLEAN   = "boolean"
	mdb_DATE      = "date"
	mdb_TIME      = "time"      // (T) time of day
	mdb_TIMESTAMP = "timestamp" // (T) date concatenated with unique time
	mdb_INTERVAL  = "interval"  // (Q) a temporal interval
	mdb_UUID      = "uuid"

	mdb_MONTH_INTERVAL = "month_interval"
	mdb_SEC_INTERVAL   = "sec_interval"
	mdb_WRD            = "wrd"
	mdb_TINYINT        = "tinyint"

	// Not on the website:
	mdb_SHORTINT    = "shortint"
	mdb_MEDIUMINT   = "mediumint"
	mdb_LONGINT     = "longint"
	mdb_FLOAT       = "float"
	mdb_TIMESTAMPTZ = "timestamptz"

	// full names and aliases, spaces are replaced with underscores
	mdb_CHARACTER               = mdb_CHAR
	mdb_CHARACTER_VARYING       = mdb_VARCHAR
	mdb_CHARACHTER_LARGE_OBJECT = mdb_CLOB
	mdb_BINARY_LARGE_OBJECT     = mdb_BLOB
	mdb_NUMERIC                 = mdb_DECIMAL
	mdb_DOUBLE_PRECISION        = mdb_DOUBLE
)

var timeFormats = []string{
	"2006-01-02",
	"2006-01-02 15:04:05",
	"2006-01-02 15:04:05 -0700",
	"2006-01-02 15:04:05 -0700 MST",
	"Mon Jan 2 15:04:05 -0700 MST 2006",
	"15:04:05",
}

type toGoConverter func(string) (driver.Value, error)
type toMonetConverter func(driver.Value) (string, error)

func stripNoQuote(v string) (driver.Value, error) {
	return unquote(strings.TrimSpace(v[0 : len(v)]))
}

func strip(v string) (driver.Value, error) {
	return unquote(strings.TrimSpace(v[1 : len(v)-1]))
}

// from strconv.contains
// contains reports whether the string contains the byte c.
func contains(s string, c byte) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return true
		}
	}
	return false
}

// adapted from strconv.Unquote
func unquote(s string) (string, error) {
	// Is it trivial?  Avoid allocation.
	if !contains(s, '\\') {
		return s, nil
	}

	var runeTmp [utf8.UTFMax]byte
	buf := make([]byte, 0, 3*len(s)/2) // Try to avoid more allocations.
	for len(s) > 0 {
		c, multibyte, ss, err := strconv.UnquoteChar(s, '\'')
		if err != nil {
			fmt.Printf("E: %v\n -> %s\n", err, s)
			return "", err
		}
		s = ss
		if c < utf8.RuneSelf || !multibyte {
			buf = append(buf, byte(c))
		} else {
			n := utf8.EncodeRune(runeTmp[:], c)
			buf = append(buf, runeTmp[:n]...)
		}
	}
	return string(buf), nil
}

func toByteArray(v string) (driver.Value, error) {
	return []byte(v[1 : len(v)-1]), nil
}

func toDouble(v string) (driver.Value, error) {
	return strconv.ParseFloat(v, 64)
}

func toFloat(v string) (driver.Value, error) {
	var r float32
	i, err := strconv.ParseFloat(v, 32)
	if err == nil {
		r = float32(i)
	}
	return r, err
}

func toInt8(v string) (driver.Value, error) {
	var r int8
	i, err := strconv.ParseInt(v, 10, 8)
	if err == nil {
		r = int8(i)
	}
	return r, err
}

func toInt16(v string) (driver.Value, error) {
	var r int16
	i, err := strconv.ParseInt(v, 10, 16)
	if err == nil {
		r = int16(i)
	}
	return r, err
}

func toInt32(v string) (driver.Value, error) {
	var r int32
	i, err := strconv.ParseInt(v, 10, 32)
	if err == nil {
		r = int32(i)
	}
	return r, err
}

func toInt64(v string) (driver.Value, error) {
	var r int64
	i, err := strconv.ParseInt(v, 10, 64)
	if err == nil {
		r = int64(i)
	}
	return r, err
}

func parseTime(v string) (t time.Time, err error) {
	for _, f := range timeFormats {
		t, err = time.Parse(f, v)
		if err == nil {
			return
		}
	}
	return
}

func toBool(v string) (driver.Value, error) {
	return strconv.ParseBool(v)
}

func toDate(v string) (driver.Value, error) {
	t, err := parseTime(v)
	if err != nil {
		return nil, err
	}
	year, month, day := t.Date()
	return Date{year, month, day}, nil
}

func toTime(v string) (driver.Value, error) {
	t, err := parseTime(v)
	if err != nil {
		return nil, err
	}
	hour, min, sec := t.Clock()
	return Time{hour, min, sec}, nil
}
func toTimestamp(v string) (driver.Value, error) {
	return parseTime(v)
}
func toTimestampTz(v string) (driver.Value, error) {
	return parseTime(v)
}

var toGoMappers = map[string]toGoConverter{
	mdb_CHAR:           strip,
	mdb_VARCHAR:        strip,
	mdb_CLOB:           strip,
	mdb_BLOB:           toByteArray,
	mdb_DECIMAL:        toDouble,
	mdb_SMALLINT:       toInt16,
	mdb_INT:            toInt32,
	mdb_WRD:            toInt32,
	mdb_BIGINT:         toInt64,
	mdb_HUGEINT:        toInt64,
	mdb_SERIAL:         toInt64,
	mdb_REAL:           toFloat,
	mdb_DOUBLE:         toDouble,
	mdb_BOOLEAN:        toBool,
	mdb_DATE:           toDate,
	mdb_TIME:           toTime,
	mdb_TIMESTAMP:      toTimestamp,
	mdb_TIMESTAMPTZ:    toTimestampTz,
	mdb_INTERVAL:       strip,
	mdb_MONTH_INTERVAL: strip,
	mdb_SEC_INTERVAL:   strip,
	mdb_TINYINT:        toInt8,
	mdb_SHORTINT:       toInt16,
	mdb_MEDIUMINT:      toInt32,
	mdb_LONGINT:        toInt64,
	mdb_FLOAT:          toFloat,
	mdb_UUID:           stripNoQuote,
}

func toString(v driver.Value) (string, error) {
	return fmt.Sprintf("%v", v), nil
}

func toQuotedString(v driver.Value) (string, error) {
	s := fmt.Sprintf("%v", v)
	s = strings.Replace(s, "\\", "\\\\", -1)
	s = strings.Replace(s, "'", "\\'", -1)
	return fmt.Sprintf("'%v'", s), nil
}

func toNull(v driver.Value) (string, error) {
	return "NULL", nil
}

func toByteString(v driver.Value) (string, error) {
	switch val := v.(type) {
	case []uint8:
		return toQuotedString(string(val))
	default:
		return "", fmt.Errorf("Unsupported type")
	}
}

func toDateTimeString(v driver.Value) (string, error) {
	switch val := v.(type) {
	case Time:
		return toQuotedString(fmt.Sprintf("%02d:%02d:%02d", val.Hour, val.Min, val.Sec))
	case Date:
		return toQuotedString(fmt.Sprintf("%04d-%02d-%02d", val.Year, val.Month, val.Day))
	default:
		return "", fmt.Errorf("Unsupported type")
	}
}

var toMonetMappers = map[string]toMonetConverter{
	"int":          toString,
	"int8":         toString,
	"int16":        toString,
	"int32":        toString,
	"int64":        toString,
	"float":        toString,
	"float32":      toString,
	"float64":      toString,
	"bool":         toString,
	"string":       toQuotedString,
	"nil":          toNull,
	"[]uint8":      toByteString,
	"time.Time":    toQuotedString,
	"monetdb.Time": toDateTimeString,
	"monetdb.Date": toDateTimeString,
}

func convertToGo(value, dataType string) (driver.Value, error) {
	if mapper, ok := toGoMappers[dataType]; ok {
		value := strings.TrimSpace(value)
		return mapper(value)
	}
	return nil, fmt.Errorf("Type not supported: %s", dataType)
}

func convertToMonet(value driver.Value) (string, error) {
	t := reflect.TypeOf(value)
	n := "nil"
	if t != nil {
		n = t.String()
	}

	if mapper, ok := toMonetMappers[n]; ok {
		return mapper(value)
	}
	return "", fmt.Errorf("Type not supported: %v", t)
}

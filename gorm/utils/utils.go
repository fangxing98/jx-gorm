package utils

import (
	"database/sql/driver"
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"unicode"
)

var gormSourceDir string

func init() {
	_, file, _, _ := runtime.Caller(0)
	// compatible solution to get gorm source directory with various operating systems
	gormSourceDir = sourceDir(file)
}

func sourceDir(file string) string {
	dir := filepath.Dir(file)
	dir = filepath.Dir(dir)

	s := filepath.Dir(dir)
	if filepath.Base(s) != "gorm.io" {
		s = dir
	}
	return filepath.ToSlash(s) + "/"
}

// FileWithLineNum return the file name and line number of the current file
func FileWithLineNum() string {
	pcs := [13]uintptr{}
	// the third caller usually from gorm internal
	len := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:len])
	for i := 0; i < len; i++ {
		// second return value is "more", not "ok"
		frame, _ := frames.Next()
		if (!strings.HasPrefix(frame.File, gormSourceDir) ||
			strings.HasSuffix(frame.File, "_test.go")) && !strings.HasSuffix(frame.File, ".gen.go") {
			return string(strconv.AppendInt(append([]byte(frame.File), ':'), int64(frame.Line), 10))
		}
	}

	return ""
}

func IsValidDBNameChar(c rune) bool {
	return !unicode.IsLetter(c) && !unicode.IsNumber(c) && c != '.' && c != '*' && c != '_' && c != '$' && c != '@'
}

// CheckTruth check string true or not
func CheckTruth(vals ...string) bool {
	for _, val := range vals {
		if val != "" && !strings.EqualFold(val, "false") {
			return true
		}
	}
	return false
}

func ToStringKey(values ...interface{}) string {
	results := make([]string, len(values))

	for idx, value := range values {
		if valuer, ok := value.(driver.Valuer); ok {
			value, _ = valuer.Value()
		}

		switch v := value.(type) {
		case string:
			results[idx] = v
		case []byte:
			results[idx] = string(v)
		case uint:
			results[idx] = strconv.FormatUint(uint64(v), 10)
		default:
			results[idx] = "nil"
			vv := reflect.ValueOf(v)
			if vv.IsValid() && !vv.IsZero() {
				results[idx] = fmt.Sprint(reflect.Indirect(vv).Interface())
			}
		}
	}

	return strings.Join(results, "_")
}

func Contains[T int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | bool | float32 | float64 | string](elems []T, elem T) bool {
	for _, e := range elems {
		if elem == e {
			return true
		}
	}
	return false
}

func AssertEqual(x, y interface{}) bool {
	if reflect.DeepEqual(x, y) {
		return true
	}
	if x == nil || y == nil {
		return false
	}

	xval := reflect.ValueOf(x)
	yval := reflect.ValueOf(y)
	if xval.Kind() == reflect.Ptr && xval.IsNil() ||
		yval.Kind() == reflect.Ptr && yval.IsNil() {
		return false
	}

	if valuer, ok := x.(driver.Valuer); ok {
		x, _ = valuer.Value()
	}
	if valuer, ok := y.(driver.Valuer); ok {
		y, _ = valuer.Value()
	}
	return reflect.DeepEqual(x, y)
}

func ToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.FormatInt(int64(v), 10)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	}
	return ""
}

const nestedRelationSplit = "__"

// NestedRelationName nested relationships like `Manager__Company`
func NestedRelationName(prefix, name string) string {
	return prefix + nestedRelationSplit + name
}

// SplitNestedRelationName Split nested relationships to `[]string{"Manager","Company"}`
func SplitNestedRelationName(name string) []string {
	return strings.Split(name, nestedRelationSplit)
}

// JoinNestedRelationNames nested relationships like `Manager__Company`
func JoinNestedRelationNames(relationNames []string) string {
	return strings.Join(relationNames, nestedRelationSplit)
}

// RTrimSlice Right trims the given slice by given length
func RTrimSlice[T any](v []T, trimLen int) []T {
	if trimLen >= len(v) { // trimLen greater than slice len means fully sliced
		return v[:0]
	}
	if trimLen < 0 { // negative trimLen is ignored
		return v[:]
	}
	return v[:len(v)-trimLen]
}

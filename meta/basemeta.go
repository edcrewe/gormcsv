package meta

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

var timeType = reflect.TypeOf(time.Time{})

var timeLayouts = []string{
	time.RFC3339Nano,
	"2006-01-02 15:04:05",
	"2006-01-02",
	"02/01/2006",
}

func convert(value string, target reflect.Type) (reflect.Value, error) {
	if target == timeType {
		parsed, err := parseTime(value)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(parsed), nil
	}
	if target.Kind() == reflect.Pointer {
		if value == "" {
			return reflect.Zero(target), nil
		}
		converted, err := convert(value, target.Elem())
		if err != nil {
			return reflect.Value{}, err
		}
		pointer := reflect.New(target.Elem())
		pointer.Elem().Set(converted)
		return pointer, nil
	}
	if value == "" && target.Kind() != reflect.String {
		return reflect.Zero(target), nil
	}

	converted := reflect.New(target).Elem()
	var err error
	switch target.Kind() {
	case reflect.String:
		converted.SetString(value)
	case reflect.Bool:
		var parsed bool
		parsed, err = strconv.ParseBool(value)
		converted.SetBool(parsed)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var parsed int64
		parsed, err = strconv.ParseInt(value, 10, target.Bits())
		converted.SetInt(parsed)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var parsed uint64
		parsed, err = strconv.ParseUint(value, 10, target.Bits())
		converted.SetUint(parsed)
	case reflect.Float32, reflect.Float64:
		var parsed float64
		parsed, err = strconv.ParseFloat(value, target.Bits())
		converted.SetFloat(parsed)
	default:
		return reflect.Value{}, fmt.Errorf("unsupported Go type %s", target)
	}
	if err != nil {
		return reflect.Value{}, fmt.Errorf("convert %q to %s: %w", value, target, err)
	}
	return converted, nil
}

func parseTime(value string) (time.Time, error) {
	for _, layout := range timeLayouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("convert %q to time.Time: unsupported date format", value)
}

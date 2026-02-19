package engine

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
)

type Parseable[T any] interface {
	Parse(string) (T, error)
}

func ParseValue[T any](s string) (T, error) {
	var zero T

	if p, ok := any(zero).(Parseable[T]); ok {
		return p.Parse(s)
	}

	return parseStringToType[T](s)
}

func parseJSONToType[T any](s string) (T, error) {
	var zero T

	err := json.Unmarshal([]byte(s), &zero)
	log.Println("parseJSONToType err", err)
	if err != nil {
		return zero, err
	}
	return zero, nil
}

func parseStringToType[T any](s string) (T, error) {
	var zero T

	switch any(zero).(type) {

	case string:
		return any(s).(T), nil
	case int:
		v, err := strconv.Atoi(s)
		return any(v).(T), err
	case *int:
		v, err := strconv.Atoi(s)
		if err != nil {
			return zero, err
		}
		return any(&v).(T), nil
	case int8:
		v, err := strconv.ParseInt(s, 10, 8)
		return any(int8(v)).(T), err
	case int16:
		v, err := strconv.ParseInt(s, 10, 16)
		return any(int16(v)).(T), err
	case int32:
		v, err := strconv.ParseInt(s, 10, 32)
		return any(int32(v)).(T), err
	case int64:
		v, err := strconv.ParseInt(s, 10, 64)
		return any(v).(T), err
	case uint:
		v, err := strconv.ParseUint(s, 10, 64)
		return any(uint(v)).(T), err
	case uint8:
		v, err := strconv.ParseUint(s, 10, 8)
		return any(uint8(v)).(T), err
	case uint16:
		v, err := strconv.ParseUint(s, 10, 16)
		return any(uint16(v)).(T), err
	case uint32:
		v, err := strconv.ParseUint(s, 10, 32)
		return any(uint32(v)).(T), err
	case uint64:
		v, err := strconv.ParseUint(s, 10, 64)
		return any(v).(T), err
	case float32:
		v, err := strconv.ParseFloat(s, 32)
		return any(float32(v)).(T), err
	case float64:
		v, err := strconv.ParseFloat(s, 64)
		return any(v).(T), err
	case bool:
		v, err := strconv.ParseBool(s)
		return any(v).(T), err
	default:
		if v, err := parseJSONToType[T](s); err == nil {
			return any(v).(T), nil
		}
		return zero, fmt.Errorf("unsupported type")
	}
}

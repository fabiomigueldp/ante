package storage

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"reflect"
	"sort"
	"time"
)

const canonicalEncodingVersion byte = 1

func EncodeCanonical(value any) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteString("ANTE")
	buf.WriteByte(canonicalEncodingVersion)
	if err := encodeCanonicalValue(buf, reflect.ValueOf(value)); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func CanonicalSHA256(value any) ([]byte, error) {
	encoded, err := EncodeCanonical(value)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(encoded)
	return sum[:], nil
}

func encodeCanonicalValue(buf *bytes.Buffer, value reflect.Value) error {
	if !value.IsValid() {
		buf.WriteByte('0')
		return nil
	}

	if value.CanInterface() {
		if ts, ok := value.Interface().(time.Time); ok {
			buf.WriteByte('T')
			return binary.Write(buf, binary.BigEndian, ts.UTC().Round(0).UnixNano())
		}
	}

	switch value.Kind() {
	case reflect.Pointer:
		if value.IsNil() {
			buf.WriteByte('P')
			buf.WriteByte(0)
			return nil
		}
		buf.WriteByte('P')
		buf.WriteByte(1)
		return encodeCanonicalValue(buf, value.Elem())
	case reflect.Interface:
		if value.IsNil() {
			buf.WriteByte('I')
			buf.WriteByte(0)
			return nil
		}
		buf.WriteByte('I')
		buf.WriteByte(1)
		elem := value.Elem()
		typeName := elem.Type().PkgPath() + "/" + elem.Type().String()
		writeCanonicalBytes(buf, []byte(typeName))
		return encodeCanonicalValue(buf, elem)
	case reflect.Bool:
		buf.WriteByte('B')
		if value.Bool() {
			buf.WriteByte(1)
		} else {
			buf.WriteByte(0)
		}
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		buf.WriteByte('i')
		return binary.Write(buf, binary.BigEndian, value.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		buf.WriteByte('u')
		return binary.Write(buf, binary.BigEndian, value.Uint())
	case reflect.String:
		buf.WriteByte('s')
		writeCanonicalBytes(buf, []byte(value.String()))
		return nil
	case reflect.Slice:
		if value.Type().Elem().Kind() == reflect.Uint8 {
			buf.WriteByte('x')
			if value.IsNil() {
				writeCanonicalLength(buf, 0)
				return nil
			}
			writeCanonicalBytes(buf, value.Bytes())
			return nil
		}
		buf.WriteByte('l')
		writeCanonicalLength(buf, value.Len())
		for i := 0; i < value.Len(); i++ {
			if err := encodeCanonicalValue(buf, value.Index(i)); err != nil {
				return err
			}
		}
		return nil
	case reflect.Array:
		buf.WriteByte('a')
		writeCanonicalLength(buf, value.Len())
		for i := 0; i < value.Len(); i++ {
			if err := encodeCanonicalValue(buf, value.Index(i)); err != nil {
				return err
			}
		}
		return nil
	case reflect.Struct:
		buf.WriteByte('r')
		typ := value.Type()
		fields := make([]reflect.StructField, 0, typ.NumField())
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			if field.PkgPath != "" {
				continue
			}
			fields = append(fields, field)
		}
		writeCanonicalLength(buf, len(fields))
		for _, field := range fields {
			writeCanonicalBytes(buf, []byte(field.Name))
			if err := encodeCanonicalValue(buf, value.FieldByIndex(field.Index)); err != nil {
				return err
			}
		}
		return nil
	case reflect.Map:
		buf.WriteByte('m')
		if value.IsNil() {
			writeCanonicalLength(buf, 0)
			return nil
		}
		keys := value.MapKeys()
		type pair struct {
			keyBytes []byte
			val      reflect.Value
		}
		pairs := make([]pair, 0, len(keys))
		for _, key := range keys {
			encodedKey, err := EncodeCanonical(key.Interface())
			if err != nil {
				return err
			}
			pairs = append(pairs, pair{keyBytes: encodedKey, val: value.MapIndex(key)})
		}
		sort.Slice(pairs, func(i, j int) bool {
			return bytes.Compare(pairs[i].keyBytes, pairs[j].keyBytes) < 0
		})
		writeCanonicalLength(buf, len(pairs))
		for _, pair := range pairs {
			writeCanonicalBytes(buf, pair.keyBytes)
			if err := encodeCanonicalValue(buf, pair.val); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("unsupported canonical type: %s", value.Kind())
	}
}

func writeCanonicalLength(buf *bytes.Buffer, n int) {
	_ = binary.Write(buf, binary.BigEndian, uint64(n))
}

func writeCanonicalBytes(buf *bytes.Buffer, raw []byte) {
	writeCanonicalLength(buf, len(raw))
	_, _ = buf.Write(raw)
}

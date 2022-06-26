package toml

import (
	"reflect"
	"strconv"
	"strings"
)

type serializer struct {
	builder strings.Builder
	depth   int
}

func Serialize(data interface{}) (result string, err error) {
	value := reflect.ValueOf(data)
	// if !value.CanAddr() {
	// 	err = fmt.Errorf("value %#v is not addressable", data)
	// 	return
	// }
	v := value.Elem()
	s := serializer{
		builder: strings.Builder{},
	}
	serializeData(&s, "", "", v)
	result = s.builder.String()
	return
}

func serializeData(s *serializer, path, name string, v reflect.Value) {
	k := v.Kind()
	switch k {
	case reflect.Map:
		serializeMap(s, path, name, v)
	case reflect.Array:
	case reflect.Struct:
		serializeStruct(s, path, name, v)
	default:
		if isNativeType(k) {
			serializeNativeValue(s, name, v)
		} else {
			return
		}
	}
}

func serializeStruct(s *serializer, path, name string, v reflect.Value) {
	t := v.Type()

	currentPath := path + name
	if s.depth > 0 {
		s.builder.WriteRune('[')
		s.builder.WriteString(currentPath)
		s.builder.WriteString("]\n")
	}
	s.depth += 1

	for i := 0; i < v.NumField(); i += 1 {
		fieldVal := v.Field(i)
		fieldInfo := t.Field(i)
		serializeData(s, currentPath, fieldInfo.Name, fieldVal)
	}
}

func serializeMap(s *serializer, path, name string, v reflect.Value) {
	currentPath := path + name
	if s.depth > 0 {
		s.builder.WriteString("[")
		s.builder.WriteString(currentPath)
		s.builder.WriteString("]\n")
	}
	s.depth += 1

	i := 0
	iter := v.MapRange()
	for iter.Next() {
		key := iter.Key()
		keyKind := key.Kind()
		value := iter.Value()

		var keyName string
		if keyKind != reflect.String {
			keyName = key.Type().Name() + strconv.Itoa(i)
		} else {
			keyName = key.String()
		}
		serializeData(s, currentPath, keyName, value)
		i += 1
	}
}

// func serializeArray()

func serializeNativeValue(s *serializer, name string, v reflect.Value) {
	s.builder.WriteString(name)
	s.builder.WriteString(" = ")

	t := v.Type()
	switch t.Kind() {
	case reflect.Bool:
		bl := v.Bool()
		s.builder.WriteString(strconv.FormatBool(bl))

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		integer := v.Int()
		s.builder.WriteString(strconv.Itoa(int(integer)))

	case reflect.String:
		str := v.String()
		s.builder.WriteRune('"')
		s.builder.WriteString(str)
		s.builder.WriteRune('"')

	case reflect.Float32, reflect.Float64:
		f := v.Float()
		s.builder.WriteString(strconv.FormatFloat(f, 'f', 4, 32))
	}
	s.builder.WriteRune('\n')
}

func isNativeType(k reflect.Kind) bool {
	return k == reflect.Bool || k == reflect.Int || k == reflect.Int8 || k == reflect.Int16 ||
		k == reflect.Int32 || k == reflect.Int64 || k == reflect.String || k == reflect.Float32 || k == reflect.Float64
}

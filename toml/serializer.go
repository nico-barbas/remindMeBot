package toml

import (
	"reflect"
	"strconv"
	"strings"
)

func Serialize(data interface{}) (result string, err error) {
	value := reflect.ValueOf(data)
	// if !value.CanAddr() {
	// 	err = fmt.Errorf("value %#v is not addressable", data)
	// 	return
	// }
	v := value.Elem()
	b := strings.Builder{}
	serializeData(&b, "", "root", v)
	result = b.String()
	return
}

func serializeData(b *strings.Builder, path, name string, v reflect.Value) {
	k := v.Kind()
	switch k {
	case reflect.Map:
		serializeMap(b, path, name, v)
	case reflect.Array:
	case reflect.Struct:
		serializeStruct(b, path, name, v)
	default:
		if isNativeType(k) {
			serializeNativeValue(b, name, v)
		} else {
			return
		}
	}
}

func serializeStruct(b *strings.Builder, path, name string, v reflect.Value) {
	t := v.Type()

	currentPath := path + name
	b.WriteRune('[')
	b.WriteString(currentPath)
	b.WriteString("]\n")
	for i := 0; i < v.NumField(); i += 1 {
		fieldVal := v.Field(i)
		fieldInfo := t.Field(i)
		serializeData(b, currentPath, fieldInfo.Name, fieldVal)
	}
}

func serializeMap(b *strings.Builder, path, name string, v reflect.Value) {
	currentPath := path + name
	b.WriteString("[")
	b.WriteString(currentPath)
	b.WriteString("]\n")

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
		serializeData(b, currentPath, keyName, value)
		i += 1
	}
}

// func serializeArray()

func serializeNativeValue(b *strings.Builder, name string, v reflect.Value) {
	b.WriteString(name)
	b.WriteString(" = ")

	t := v.Type()
	switch t.Kind() {
	case reflect.Bool:
		bl := v.Bool()
		b.WriteString(strconv.FormatBool(bl))

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		integer := v.Int()
		b.WriteString(strconv.Itoa(int(integer)))

	case reflect.String:
		str := v.String()
		b.WriteRune('"')
		b.WriteString(str)
		b.WriteRune('"')

	case reflect.Float32, reflect.Float64:
		f := v.Float()
		b.WriteString(strconv.FormatFloat(f, 'f', 4, 32))
	}
	b.WriteRune('\n')
}

func isNativeType(k reflect.Kind) bool {
	return k == reflect.Bool || k == reflect.Int || k == reflect.Int8 || k == reflect.Int16 ||
		k == reflect.Int32 || k == reflect.Int64 || k == reflect.String || k == reflect.Float32 || k == reflect.Float64
}

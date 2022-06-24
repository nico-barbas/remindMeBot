package toml

import (
	"reflect"
	"strconv"
	"strings"
)

func Serialize(data interface{}) string {
	b := strings.Builder{}
	serializeStruct(&b, data)
	return b.String()
}

func serializeStruct(b *strings.Builder, data interface{}) {
	value := reflect.ValueOf(data).Elem()
	t := reflect.TypeOf(data)

	b.WriteRune('[')
	b.WriteString(t.Name())
	b.WriteString("]\n")
	for i := 0; i < value.NumField(); i += 1 {
		fieldVal := value.Field(i)
		fieldInfo := t.Field(i)
		if isNativeType(fieldInfo.Type.Kind()) {
			serializeNativeValue(b, fieldInfo.Name, fieldVal)
		}
	}
}

func serializeNativeValue(b *strings.Builder, name string, data reflect.Value) {
	b.WriteString(name)
	b.WriteString(" = ")

	t := data.Type()
	switch t.Kind() {
	case reflect.Bool:
		bl := data.Bool()
		b.WriteString(strconv.FormatBool(bl))

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		integer := data.Int()
		b.WriteString(strconv.Itoa(int(integer)))

	case reflect.String:
		str := data.String()
		b.WriteRune('"')
		b.WriteString(str)
		b.WriteRune('"')

	case reflect.Float32, reflect.Float64:
		f := data.Float()
		b.WriteString(strconv.FormatFloat(f, 'f', 4, 32))
	}
	b.WriteRune('\n')
}

func isNativeType(k reflect.Kind) bool {
	return k == reflect.Bool || k == reflect.Int || k == reflect.Int8 || k == reflect.Int16 ||
		k == reflect.Int32 || k == reflect.Int64 || k == reflect.String || k == reflect.Float32 || k == reflect.Float64
}

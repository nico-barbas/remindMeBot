package toml

import (
	"fmt"
	"reflect"
)

func Deserialize(input string, outputFormat interface{}) (err error) {
	var root Table
	root, err = Parse(input)
	if err != nil {
		return
	}

	v := reflect.ValueOf(outputFormat).Elem()
	if !v.IsValid() {
		err = fmt.Errorf("invalid value %#v", v)
	}
	err = deserializeValue(root, v)
	return
}

func deserializeValue(tomlValue Value, v reflect.Value) (err error) {
	switch t := tomlValue.(type) {
	case Number:
		deserializeNumber(t, v)

	case Boolean:
		deserializeBoolean(t, v)

	case String:
		deserializeString(t, v)

	case *Array:
	case Table:
		err = deserializeTable(t, v)
	}

	return nil
}

func deserializeTable(t Table, v reflect.Value) (err error) {
	k := v.Kind()
	if !(k == reflect.Map || k == reflect.Struct) {
		err = fmt.Errorf("value %v is not a map or a struct", v)
		return
	}

	switch k {
	case reflect.Map:
		keyType := v.Type().Key()
		if keyType.Kind() != reflect.String {
			err = fmt.Errorf("map %#v's key are not of type string", v)
		}
		elemType := v.Type().Elem()
		for key, value := range t {
			keyValue := reflect.ValueOf(key)
			elemValue := reflect.New(elemType)

			err = deserializeValue(value, elemValue)
			if err != nil {
				return
			}

			v.SetMapIndex(keyValue, elemValue.Elem())
		}

	case reflect.Struct:
		for key, value := range t {
			fieldValue := v.FieldByName(key)
			if fieldValue.IsValid() {
				err = deserializeValue(value, fieldValue)
				if err != nil {
					return
				}
			}
		}
	}

	return
}

func deserializeNumber(n Number, v reflect.Value) (err error) {
	if !isNumberValue(v.Kind()) {
		err = fmt.Errorf("value %v is not a number", v)
		return
	}

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v.SetInt(int64(n))

	case reflect.Float32, reflect.Float64:
		v.SetFloat(float64(n))
	}

	return
}

func deserializeBoolean(b Boolean, v reflect.Value) (err error) {
	if v.Kind() != reflect.Bool {
		err = fmt.Errorf("value %v is not a bool", v)
	}
	v.SetBool(bool(b))
	return
}

func deserializeString(s String, v reflect.Value) (err error) {
	if v.Kind() != reflect.String {
		err = fmt.Errorf("value %v is not a string", v)
	}
	v.SetString(string(s))
	return
}

func isNumberValue(k reflect.Kind) bool {
	return k == reflect.Int || k == reflect.Int8 || k == reflect.Int16 ||
		k == reflect.Int32 || k == reflect.Int64 || k == reflect.Float32 || k == reflect.Float64
}

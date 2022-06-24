package toml

import (
	"fmt"
	"testing"
)

func TestRootKeyValue(t *testing.T) {
	input := `
	hello = 10
	world = "!"
	d = true
	`
	keyVal := []struct {
		key string
		val Value
	}{
		{"hello", Number(10)},
		{"world", String("!")},
		{"d", Boolean(true)},
	}

	result, err := Parse(input)
	if err != nil {
		t.Error(err)
	}
	for _, kv := range keyVal {
		if _, exist := result[kv.key]; !exist {
			t.Errorf("Key %s was not found", kv.key)
		} else {
			val := result[kv.key]
			if kv.val != val {
				t.Errorf("Key %s does not have the correct value. Got %v, expected %v", kv.key, val, kv.val)
			}
		}
	}
}

func TestDottedKeys(t *testing.T) {
	input := `
	hello.world = 10
	`
	result, err := Parse(input)
	if err != nil {
		t.Error(err)
	}
	if len(result) != 1 {
		t.Errorf("Failed toml parsing")
	}
	if ta, exist := result["hello"]; !exist {
		t.Errorf(`Table "hello" was not found`)
	} else {
		table, ok := ta.(Table)
		if !ok {
			t.Errorf(`Key "hello" is not a Table; %#v`, ta)
		}
		if _, exist := table["world"]; !exist {
			t.Errorf(`Key "world" was not found`)
		}
	}
}

func TestNestedTable(t *testing.T) {
	input := `
	[hello]
	world = 10
	[hello.dear]
	world = false
	[hi]
	mom = true
	`

	result, err := Parse(input)
	if err != nil {
		t.Error(err)
	}
	if len(result) != 2 {
		t.Errorf("Failed toml parsing")
	}
	if result["hello"].(Table)["world"] != Number(10) {
		t.Errorf("Key hello.world doesn't have the right value, expected 10 got %#v", result["hello"].(Table)["world"])
	}
	if result["hello"].(Table)["dear"].(Table)["world"] != Boolean(false) {
		t.Errorf("Key hello.world doesn't have the right value, expected false got %#v", result["hello"].(Table)["dear"].(Table)["world"])
	}
	if result["hi"].(Table)["mom"] != Boolean(true) {
		t.Errorf("Key hi.mom doesn't have the right value, expected true got %#v", result["hi"].(Table)["mom"])
	}
}

func TestArrayKeyVal(t *testing.T) {
	input := `
	array = [1, 2, 3]
	`

	result, err := Parse(input)
	if err != nil {
		t.Error(err)
	}
	if len(result) != 1 {
		t.Errorf("Failed toml parsing")
	}
	if array, ok := result["array"].(*Array); !ok {
		t.Errorf(`Key "array" is not an Array; %#v`, result["array"])
	} else {
		if array.length() != 3 {
			t.Errorf(`Array "array" is not of length 3, got %d`, array.length())
		}
	}
}

func TestInlineTable(t *testing.T) {
	input := `
	inline = { hello = 10, world = true }
	`

	result, err := Parse(input)
	if err != nil {
		t.Error(err)
	}
	if len(result) != 1 {
		t.Errorf("Failed toml parsing")
	}
	if len(result["inline"].(Table)) != 2 {
		t.Errorf(`Table "inline" is not of length 2, got %d`, len(result["inline"].(Table)))
	}
	if result["inline"].(Table)["hello"] != Number(10) {
		t.Errorf("Key inline.hello doesn't have the right value, expected 10 got %#v", result["inline"].(Table)["hello"])
	}
	if result["inline"].(Table)["world"] != Boolean(true) {
		t.Errorf("Key inline.world doesn't have the right value, expected true got %#v", result["inline"].(Table)["world"])
	}
}

func TestArrayOfTables(t *testing.T) {
	input := `
	[[array]]
	foo = 10
	bar = false
	[[array]]
	foo = 8
	bar = true
	`

	result, err := Parse(input)
	if err != nil {
		t.Error(err)
	}
	if len(result) != 1 {
		t.Errorf("Failed toml parsing")
	}
	if array, ok := result["array"].(*Array); !ok {
		t.Errorf(`Key "array" is not an Array; %#v`, result["array"])
	} else {
		if array.length() != 2 {
			t.Errorf("Array doesn't have a length of 2, got %d", array.length())
		} else {
			fmt.Printf("%#v\n", array.get(0))
			fmt.Printf("%#v\n", array.get(1))
		}
	}
}

// func TestNestedArrayOfTables(t *testing.T) {
// 	input := `
// 	[[array.hello]]
// 	foo = 10
// 	bar = false
// 	[[array.hello]]
// 	foo = 8
// 	bar = true
// 	`

// 	result, err := Parse(input)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	if len(result) != 1 {
// 		t.Errorf("Failed toml parsing")
// 	}
// 	fmt.Println(result)
// }

func TestSerializeSimpleStruct(t *testing.T) {
	input := struct {
		X int
		Y int
	}{
		X: 10,
		Y: 10,
	}

	expect := "[root]\nX = 10\nY = 10\n"

	result, err := Serialize(&input)

	if err != nil {
		t.Error(err)
	}

	if result != expect {
		t.Errorf("Expect %s, got %s", expect, result)
	}
}

func TestSerializeMap(t *testing.T) {
	input := map[string]int{
		"x": 22,
		"y": 76,
	}

	expect := "[root]\nx = 22\ny = 76\n"

	result, err := Serialize(&input)

	if err != nil {
		t.Error(err)
	}

	if result != expect {
		t.Errorf("Expect %s, got %s", expect, result)
	}
}

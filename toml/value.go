package toml

import "fmt"

const (
	arrayDefaultCap = 16
)

type Value interface {
	toString() string
}

type (
	Number  float64
	Boolean bool
	String  string
	Array   struct {
		data []Value
	}
	Table map[string]Value
)

func (n Number) toString() string { return "Number" }

func (b Boolean) toString() string { return "Boolean" }

func (s String) toString() string { return "String" }

func (a *Array) toString() string { return "Array" }

func makeArray(cap int) *Array {
	return &Array{
		data: make([]Value, 0, cap),
	}
}

func (a *Array) appendValue(v Value) {
	a.data = append(a.data, v)
}

func (a *Array) get(index int) Value {
	return a.data[index]
}

func (a *Array) length() int {
	return len(a.data)
}

func (t Table) toString() string { return "Table" }

func (t Table) insertKeyValue(k key, v Value) error {
	child := t.getChildTable(k.accessors[:k.count-1])
	name := k.name()
	if _, exist := child[name]; !exist {
		child[name] = v
	} else {
		return fmt.Errorf("already key with name %s", name)
	}
	return nil
}

func (t Table) getValue(k key) Value {
	currentTable := t
	for _, token := range k.decl() {
		if value, exist := currentTable[token.lexeme]; exist {
			if table, ok := value.(Table); ok {
				currentTable = table
			}
		} else {
			return nil
		}
	}
	if val, exist := currentTable[k.accessors[k.count-1].lexeme]; exist {
		return val
	} else {
		return nil
	}
}

// No fail state here
//
// If the given accessor doesn't exsit, it will be inserted and returned
func (t Table) getChildTable(accessors []token) Table {
	currentTable := t
	for _, token := range accessors {
		if value, exist := currentTable[token.lexeme]; exist {
			if table, ok := value.(Table); ok {
				currentTable = table
			}
		} else {
			table := make(Table)
			currentTable[token.lexeme] = table
			currentTable = table
		}
	}
	return currentTable
}

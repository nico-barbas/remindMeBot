package toml

import (
	"fmt"
	"strconv"
)

type (
	parser struct {
		lexer    lexer
		previous token
		current  token

		root         Table
		currentTable Table
	}

	// The length is hardcorded for now.
	// But I highly doubt I'll ever need more than
	// 10 accessors
	key struct {
		accessors [10]token
		count     int
	}
)

func Parse(source string) (Table, error) {
	p := parser{}
	p.lexer.initLexer(source)
	p.root = make(Table)
	p.currentTable = p.root

	var err error
pass:
	for {
		t := p.consume()
		switch t.kind {
		case tokenNewline:
			continue
		case tokenOpenBracket:
			if p.consume().kind == tokenOpenBracket {
				err = p.parserArrayOfTables()
				if err != nil {
					break pass
				}
			} else {
				err = p.parseTableDeclaration()
				if err != nil {
					break pass
				}
			}
		case tokenIdentifier:
			err = p.parseKeyValueDeclaration()
			if err != nil {
				break pass
			}
		case tokenEOF:
			break pass
		default:
			continue
		}
	}

	return p.root, nil
}

func (p *parser) consume() token {
	p.previous = p.current
	p.current = p.lexer.scanToken()
	return p.current
}

func (p *parser) parseKeyValueDeclaration() error {
	key, err := p.parseKey(tokenEqual, true)
	if err != nil {
		return err
	}

	valueToken := p.consume()
	value, err := p.parseValue(valueToken)
	if err != nil {
		return err
	}

	err = p.currentTable.insertKeyValue(key, value)
	if err != nil {
		return err
	}

	return nil
}

func (p *parser) parseTableDeclaration() error {
	p.currentTable = p.root
	key, err := p.parseKey(tokenCloseBracket, true)
	if err != nil {
		return err
	}
	newTable := p.currentTable.getChildTable(key.decl())
	p.currentTable = newTable
	return nil
}

func (p *parser) parserArrayOfTables() error {
	p.currentTable = p.root

	key, err := p.parseKey(tokenCloseBracket, false)
	if err != nil {
		return err
	}
	if bracket := p.consume(); bracket.kind != tokenCloseBracket {
		return fmt.Errorf("syntax Error: Expected ']' at line %d", bracket.line)
	}

	var array *Array
	a := p.currentTable.getValue(key)
	if a == nil {
		array = makeArray(arrayDefaultCap)
		p.currentTable.insertKeyValue(key, array)
	} else {
		var ok bool
		array, ok = a.(*Array)
		if !ok {
			return fmt.Errorf("given key does not refer to an array, got %#v", a)
		}
	}

	table := make(Table)
	array.appendValue(table)
	p.currentTable = table
	return nil
}

func (p *parser) parseKey(termToken tokenKind, appendCurrent bool) (key, error) {
	var err error
	k := key{}
	if appendCurrent {
		k.accessors[0] = p.current
		k.count += 1
	}
key:
	for {
		t := p.consume()
		switch t.kind {
		case tokenIdentifier:
			k.accessors[k.count] = t
			k.count += 1
		case tokenDot:
			if p.previous.kind == tokenIdentifier {
				continue
			} else {
				err = fmt.Errorf("invalid syntax at line %d; got %d after '.'", t.line, t.kind)
				break key
			}
		case termToken:
			break key
		default:
			err = fmt.Errorf("invalid syntax at line %d; got %d after %d", t.line, t.kind, p.previous.kind)
		}
	}
	return k, err
}

func (p *parser) parseValue(valueToken token) (v Value, err error) {
	switch valueToken.kind {
	case tokenNumber:
		val, e := strconv.ParseFloat(valueToken.lexeme, 64)
		if e != nil {
			err = e
			break
		}
		v = Number(val)
	case tokenFalse:
		v = Boolean(false)
	case tokenTrue:
		v = Boolean(true)
	case tokenString:
		v = String(valueToken.lexeme[1 : len(valueToken.lexeme)-1])
	case tokenOpenBracket:
		v, err = p.parseArrayValue()
	case tokenOpenBrace:
		v, err = p.parseInlineTableValue()
	}

	return v, err
}

func (p *parser) parseArrayValue() (Value, error) {
	arrayCount := 0
	start := p.lexer.current
lookahead:
	for {
		t := p.consume()
		switch t.kind {
		case tokenCloseBracket:
			break lookahead
		case tokenColon:
			if p.previous.isValueKind() {
				continue
			} else {
				err := fmt.Errorf("invalid syntax at line %d; got %d after '.'", t.line, t.kind)
				return nil, err
			}
		default:
			if t.isValueKind() {
				arrayCount += 1
			}
		}
	}
	p.lexer.current = start
	array := makeArray(arrayCount)
arrayLoop:
	for {
		t := p.consume()
		switch t.kind {
		case tokenCloseBracket:
			break arrayLoop
		case tokenColon:
			continue
		default:
			val, err := p.parseValue(t)
			if err != nil {
				return nil, err
			}
			array.appendValue(val)
		}
	}
	return array, nil
}

func (p *parser) parseInlineTableValue() (Value, error) {
	table := make(Table)
tableLoop:
	for {
		t := p.consume()
		switch t.kind {
		case tokenCloseBrace:
			break tableLoop
		case tokenColon:
			if p.previous.isValueKind() {
				continue
			} else {
				err := fmt.Errorf("invalid syntax at line %d; got %d after '.'", t.line, t.kind)
				return nil, err
			}
		case tokenIdentifier:
			key, err := p.parseKey(tokenEqual, true)
			if err != nil {
				return nil, err
			}

			value, err := p.parseValue(p.consume())
			if err != nil {
				return nil, err
			}
			switch key.count {
			case 1:
				err := table.insertKeyValue(key, value)
				if err != nil {
					return nil, err
				}
			default:
				child := table.getChildTable(key.decl())
				child.insertKeyValue(key, value)
			}
		}
	}
	return table, nil
}

func (k key) decl() []token {
	return k.accessors[:k.count]
}

func (k key) name() string {
	return k.accessors[k.count-1].lexeme
}

package main

import (
	"fmt"
)

type (
	parser struct {
		lexer struct {
			input   []byte
			current int
		}
		current  token
		previous token
	}
)

func parseCommand(input string) (result command, err parserError) {
	parser := parser{}
	parser.lexer.input = []byte(input)
	for {
		var t token
		t, err = parser.consume()
		if !err.isOK() || t.kind == tokenEOF {
			break
		}

		if err = parser.expect(tokenBang); !err.isOK() {
			break
		}
		if err = parser.expectNext(tokenIdentifier); !err.isOK() {
			break
		}

		if cmdKind, exist := commandKeywords[parser.current.text]; exist {
			switch cmdKind {
			case commandBriefMe:
				result, err = parser.parseBriefMeCmd()

			case commandRemindMe:
				result, err = parser.parseRemindMeCmd()

			case commandStaffMe:
				// result, err = parser.parseStaffMeCmd()

			case commandShowTodo:
				// result, err = parser.parseShowTodoCmd()

			case commandShowReminders:
				// result, err = parser.parseShowRemindersCmd()

			}
		} else {
			err = parserError{
				kind:    errorUnknownCommand,
				token:   parser.current,
				details: fmt.Sprintf("!%s", parser.current.text),
			}
			return
		}
	}
	return
}

func (self *parser) peekNextToken() (result token, err parserError) {
	start := self.lexer.current
	result, err = self.scanToken()
	self.lexer.current = start
	return
}

func (self *parser) expectNext(expected tokenKind) (err parserError) {
	var t token
	t, err = self.consume()
	if t.kind != expected {
		err = parserError{
			kind:  errorInvalidSyntax,
			token: self.current,
			details: fmt.Sprintf(
				"Expected %s got %s",
				tokenKindString[expected],
				tokenKindString[self.current.kind],
			),
		}
	}
	return
}

func (self *parser) expect(expected tokenKind) (err parserError) {
	if self.current.kind != expected {
		err = parserError{
			kind:  errorInvalidSyntax,
			token: self.current,
			details: fmt.Sprintf(
				"Expected %s got %s",
				tokenKindString[expected],
				tokenKindString[self.current.kind],
			),
		}
	}
	return
}

func (self *parser) consume() (result token, err parserError) {
	self.previous = self.current
	self.current, err = self.scanToken()
	result = self.current
	return
}

func (self *parser) parseBriefMeCmd() (result *briefMeCommand, err parserError) {
	result = &briefMeCommand{
		kind:     commandBriefMe,
		token:    self.previous,
		cmdToken: self.current,
	}
	return
}

func (self *parser) parseRemindMeCmd() (result *remindMeCommand, err parserError) {
	result = &remindMeCommand{
		kind:     commandRemindMe,
		token:    self.previous,
		cmdToken: self.current,
	}

	var next token
	var start token
	var end token
	if next, err = self.consume(); !err.isOK() {
		return
	}
	if !(next.kind == tokenIdentifier || next.kind == tokenEmote) {
		err = parserError{
			kind:  errorInvalidSyntax,
			token: next,
			details: fmt.Sprintf(
				"Expected either %s or %s got %s",
				tokenKindString[tokenIdentifier],
				tokenKindString[tokenEmote],
				tokenKindString[next.kind],
			),
		}
	}

	start = next
parseIdentifier:
	for {
		if next, err = self.consume(); !err.isOK() {
			return
		}
		switch {
		case next.kind == tokenSeparator:
			end = self.previous
			break parseIdentifier
		case next.kind == tokenIdentifier || next.kind == tokenEmote || next.kind == tokenNumber:
			continue

		default:
			err = parserError{
				kind:  errorInvalidSyntax,
				token: next,
				details: fmt.Sprintf(
					"Expected one of %s, %s or %s, got %s",
					tokenKindString[tokenIdentifier],
					tokenKindString[tokenEmote],
					tokenKindString[tokenNumber],
					tokenKindString[next.kind],
				),
			}
			return
		}
	}
	result.identifier = string(self.lexer.input[start.start:end.end])
	result.sepToken = self.current

	result.date, err = self.parseDate()
	return
}

func (self *parser) parseDate() (result date, err parserError) {
	// Needs to handle:
	// hh:min
	// dd-mm
	// dd-mm-yy
	// dd-mm hh:min
	// dd-mm-yy hh:min

	var t token
	if err = self.expectNext(tokenNumber); !err.isOK() {
		return
	}
	if t, err = self.peekNextToken(); !err.isOK() {
		return
	}

	ddmmyy := [3]token{}
	hhmm := [2]token{}
	switch t.kind {
	case tokenDash:
		ddmmyy, err = self.parseDDMMYY()
		if !err.isOK() {
			return
		}
		if t, err = self.peekNextToken(); !err.isOK() {
			return
		}
		if t.kind == tokenNumber {
			self.consume()
			hhmm, err = self.parseHHMM()
			if !err.isOK() {
				return
			}
		}

	case tokenColon:
		hhmm, err = self.parseHHMM()
		if !err.isOK() {
			return
		}
	}

	result = makeDate(ddmmyy, hhmm)
	return
}

func (self *parser) parseDDMMYY() (ddmmyy [3]token, err parserError) {
	ddmmyy[0] = self.current

	if err = self.expectNext(tokenDash); !err.isOK() {
		return
	}
	if err = self.expectNext(tokenNumber); !err.isOK() {
		return
	}
	ddmmyy[1] = self.current

	var t token
	if t, err = self.peekNextToken(); !err.isOK() {
		return
	}
	if t.kind == tokenDash {
		self.consume()
		if err = self.expectNext(tokenNumber); !err.isOK() {
			return
		}
		ddmmyy[2] = self.current
	}

	return
}

func (self *parser) parseHHMM() (hhmm [2]token, err parserError) {
	hhmm[0] = self.current
	if err = self.expectNext(tokenColon); !err.isOK() {
		return
	}
	if err = self.expectNext(tokenNumber); !err.isOK() {
		return
	}
	hhmm[1] = self.current
	return
}

type (
	tokenKind byte

	token struct {
		text  string
		start int
		end   int
		kind  tokenKind
	}
)

const (
	tokenInvalid tokenKind = iota
	tokenEOF
	tokenIdentifier
	tokenNumber
	tokenEmote
	tokenBang
	tokenDash
	tokenDoubleDash
	tokenColon
	tokenSeparator
)

var tokenKindString = map[tokenKind]string{
	tokenInvalid:    "tokenInvalid",
	tokenEOF:        "tokenEOF",
	tokenIdentifier: "tokenIdentifier",
	tokenNumber:     "tokenNumber",
	tokenEmote:      "tokenEmote",
	tokenBang:       "tokenBang",
	tokenDash:       "tokenDash",
	tokenDoubleDash: "tokenDoubleDash",
	tokenColon:      "tokenColon",
	tokenSeparator:  "tokenSeparator",
}

func (t token) String() string {
	return fmt.Sprintf("[%s  %d:%d]", tokenKindString[t.kind], t.start, t.end)
}

func (self *parser) setInput(input string) {
	self.lexer.input = []byte(input)
	self.lexer.current = 0
}

func (self *parser) scanToken() (result token, err parserError) {
	eof := self.skipWhitespaces()
	result.start = self.lexer.current
	if eof {
		result.end = self.lexer.current
		result.kind = tokenEOF
		return
	}

	c := self.advance()
	switch c {
	case '!':
		result.kind = tokenBang

	case ':':
		start := self.lexer.current
		isEmote := false
	lookahead:
		for {
			if self.isEOF() {
				break lookahead
			}
			next := self.peek()
			switch {
			case next == ':':
				isEmote = true
				self.advance()
				break lookahead
			case isLetter(next) || isNumber(next) || next == '_':
				self.advance()
			default:
				break lookahead
			}
		}

		if isEmote {
			result.kind = tokenEmote
		} else {
			self.lexer.current = start
			result.kind = tokenColon
		}

	case '-':
		if !self.isEOF() && self.peek() == '-' {
			self.advance()
			result.kind = tokenDoubleDash
		} else {
			result.kind = tokenDash
		}

	case '|':
		result.kind = tokenSeparator

	default:
		switch {
		case isLetter(c):
			result.kind = tokenIdentifier
		lexIdentifier:
			for {
				if self.isEOF() {
					break lexIdentifier
				}
				next := self.peek()
				if !isLetter(next) {
					break lexIdentifier
				}
				self.advance()
			}
		case isNumber(c):
			result.kind = tokenNumber
		lexNumber:
			for {
				if self.isEOF() {
					break lexNumber
				}
				next := self.peek()
				if !isNumber(next) {
					break lexNumber
				}
				self.advance()
			}

		default:
			result.end = self.lexer.current
			result.kind = tokenInvalid
			result.text = string(self.lexer.input[result.start:result.end])
			err = parserError{
				kind:    errorInvalidToken,
				token:   result,
				details: fmt.Sprintf("%s is not a valid token", result.text),
			}
			return
		}
	}
	result.end = self.lexer.current
	result.text = string(self.lexer.input[result.start:result.end])
	return
}

func (self *parser) isEOF() bool {
	return self.lexer.current >= len(self.lexer.input)
}

func isLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isNumber(c byte) bool {
	return (c >= '0' && c <= '9')
}

func (self *parser) advance() byte {
	self.lexer.current += 1
	return self.lexer.input[self.lexer.current-1]
}

func (self *parser) peek() byte {
	return self.lexer.input[self.lexer.current]
}

func (self *parser) skipWhitespaces() (eof bool) {
	for {
		if eof = self.isEOF(); eof {
			break
		}
		c := self.peek()
		if c == ' ' || c == '\t' {
			self.advance()
		} else {
			break
		}
	}
	return
}

package toml

const (
	tokenInvalid tokenKind = iota
	tokenNewline
	tokenEOF
	tokenDot
	tokenColon
	tokenEqual
	tokenOpenBracket
	tokenCloseBracket
	tokenOpenBrace
	tokenCloseBrace
	tokenPound
	tokenNumber
	tokenTrue
	tokenFalse
	tokenString
	tokenIdentifier
)

var keywords = map[string]tokenKind{
	"false": tokenFalse,
	"true":  tokenTrue,
}

type (
	tokenKind int

	token struct {
		kind   tokenKind
		lexeme string
		line   int
		column int
		start  int
		end    int
	}

	lexer struct {
		source  string
		line    int
		column  int
		current int
	}
)

func (l *lexer) initLexer(input string) {
	*l = lexer{
		source: input,
	}
}

func (l *lexer) scanToken() token {
	l.skipWhitespaces()
	if l.EOF() {
		return token{kind: tokenEOF}
	}
	t := token{
		line:  l.line,
		start: l.current,
	}

	c := l.advance()
	switch c {
	case '#':
		for {
			next := l.advance()
			if next == '\n' {
				l.line += 1
				break
			}
		}
	case '.':
		t.kind = tokenDot
	case ',':
		t.kind = tokenColon
	case '=':
		t.kind = tokenEqual
	case '[':
		t.kind = tokenOpenBracket
	case ']':
		t.kind = tokenCloseBracket
	case '{':
		t.kind = tokenOpenBrace
	case '}':
		t.kind = tokenCloseBrace
	case '\n':
		t.kind = tokenNewline
	case '"':
		t.kind = tokenString
		for {
			next := l.advance()
			if next == '"' {
				break
			}
		}
	default:
		if isLetter(c) {
			l.lexIdentifer()
			if kind, exist := keywords[l.source[t.start:l.current]]; exist {
				t.kind = kind
			} else {
				t.kind = tokenIdentifier
			}
		} else if isNumber(c) {
			l.lexNumber()
			t.kind = tokenNumber
		}
	}
	t.end = l.current
	t.lexeme = l.source[t.start:t.end]
	return t
}

func (l *lexer) EOF() bool {
	return l.current >= len(l.source)
}

func (l *lexer) advance() byte {
	l.current += 1
	return l.source[l.current-1]
}

func (l *lexer) peek() byte {
	return l.source[l.current]
}

func (l *lexer) skipWhitespaces() {
	for !l.EOF() {
		c := l.peek()
		if c == ' ' || c == '\t' {
			l.advance()
		} else {
			break
		}
	}
}

func (l *lexer) lexIdentifer() {
	for {
		if l.EOF() {
			break
		}
		c := l.peek()
		if isLetter(c) {
			l.advance()
		} else {
			break
		}
	}
}

func (l *lexer) lexNumber() {
	for {
		if l.EOF() {
			break
		}
		c := l.peek()
		if isNumber(c) || c == '.' {
			l.advance()
		} else {
			break
		}
	}
}

// func (l *lexer) remaining() int {
// 	return len(l.source) - l.current
// }

func (t token) isValueKind() bool {
	return t.kind == tokenNumber || t.kind == tokenFalse || t.kind == tokenTrue || t.kind == tokenString
}

func isLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

func isNumber(c byte) bool {
	return c >= '0' && c < '9'
}

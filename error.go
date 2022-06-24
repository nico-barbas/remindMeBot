package main

type (
	parserError struct {
		kind    parserErrorKind
		token   token
		details string
	}

	parserErrorKind int
)

const (
	errorOK parserErrorKind = iota
	errorInvalidToken
	errorInvalidSyntax
	errorInvalidDate
	errorUnknownCommand
)

func (err parserError) isOK() bool {
	return err.kind == errorOK
}

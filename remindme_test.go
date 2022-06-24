package main

import (
	"testing"
	"time"
)

func TestLexer(t *testing.T) {
	inputs := []string{
		"12", "remind", ":myemote:", "!", "-", "--", ":", "|",
		"reminder", "task", "today", "tomorrow",
	}

	const tokenCountPerInput = 2
	expects := []tokenKind{
		tokenNumber, tokenIdentifier, tokenEmote,
		tokenBang, tokenDash, tokenDoubleDash, tokenColon, tokenSeparator,
		tokenReminder, tokenTask, tokenToday, tokenTomorrow,
	}

	parser := parser{}
	for i, input := range inputs {
		parser.setInput(input)
		tokens := []token{}
		for {
			token, err := parser.scanToken()
			if !err.isOK() {
				t.Errorf("lexing error: %s", err.details)
			}
			tokens = append(tokens, token)
			if token.kind == tokenEOF {
				break
			}
		}

		// One relevant token and one EOF
		if len(tokens) != tokenCountPerInput {
			t.Errorf(
				"invalid number of tokens in result %d, expected %d got %d",
				i,
				tokenCountPerInput,
				len(tokens),
			)
		}

		result := tokens[0]
		if result.kind != expects[i] {
			t.Errorf(
				"invalid result, expected %s got %s",
				tokenKindString[expects[i]],
				tokenKindString[result.kind],
			)
		}
	}
}

func TestLexerEdgeCases(t *testing.T) {
	inputs := []string{
		":30",
	}

	expects := [][]tokenKind{
		{tokenColon, tokenNumber, tokenEOF},
	}

	parser := parser{}
	for i, input := range inputs {
		parser.setInput(input)
		tokens := []token{}
		for {
			token, err := parser.scanToken()
			if !err.isOK() {
				t.Errorf("lexing error: %s", err.details)
			}
			tokens = append(tokens, token)
			if token.kind == tokenEOF {
				break
			}
		}

		expect := expects[i]
		// One relevant token and one EOF
		if len(tokens) != len(expect) {
			t.Errorf(
				"invalid number of tokens in result %d, expected %d got %d",
				i,
				len(expect),
				len(tokens),
			)
		}

		for i, kind := range expect {
			result := tokens[i]
			if result.kind != kind {
				t.Errorf(
					"invalid result, expected %s got %s",
					tokenKindString[kind],
					tokenKindString[result.kind],
				)
			}
		}
	}
}

func TestParseDate(t *testing.T) {
	inputs := []string{
		"10-07",
		"10-07-22",
		"12:30",
		"10-07 12:30",
		"10-07-22 12:30",
	}

	y, m, d := time.Now().Date()
	expects := []date{
		{day: 10, month: time.July, year: 2022, hour: 0, min: 0},
		{day: 10, month: time.July, year: 2022, hour: 0, min: 0},
		{day: d, month: m, year: y, hour: 12, min: 30},
		{day: 10, month: time.July, year: 2022, hour: 12, min: 30},
		{day: 10, month: time.July, year: 2022, hour: 12, min: 30},
	}

	parser := parser{}
	for i, input := range inputs {
		t.Logf("input %d", i)
		parser.setInput(input)
		d, err := parser.parseDate()
		if !err.isOK() {
			t.Errorf("parsing error: %s", err.details)
		}

		expect := expects[i]
		if d.day != expect.day {
			t.Errorf(
				"invalid day, expected %d got %d",
				expect.day,
				d.day,
			)
		}
		if d.month != expect.month {
			t.Errorf(
				"invalid day, expected %s got %s",
				expect.month.String(),
				d.month.String(),
			)
		}
		if d.year != expect.year {
			t.Errorf(
				"invalid day, expected %d got %d",
				expect.day,
				d.day,
			)
		}

		if d.hour != expect.hour {
			t.Errorf(
				"invalid day, expected %d got %d",
				expect.hour,
				d.hour,
			)
		}
		if d.min != expect.min {
			t.Errorf(
				"invalid day, expected %d got %d",
				expect.min,
				d.min,
			)
		}
	}
}

func TestParseRemindMeCmd(t *testing.T) {
	input := "!remindme pick up the milk | 18:30"

	y, m, d := time.Now().Date()
	expect := remindMeCommand{
		kind:       commandRemindMe,
		identifier: "pick up the milk",
		date:       date{day: d, month: m, year: y, hour: 18, min: 30},
	}

	result, err := parseCommand(input)
	if !err.isOK() {
		t.Errorf("parsing error: %s", err.details)
	}
	switch r := result.(type) {
	case *remindMeCommand:
		if r.kind != expect.kind {
			t.Errorf(
				"invalid kind, expected %d got %d",
				expect.kind,
				r.kind,
			)
		}

		if r.identifier != expect.identifier {
			t.Errorf(
				"invalid identifier, expected %s got %s",
				expect.identifier,
				r.identifier,
			)
		}

		if !r.date.isEqual(expect.date) {
			t.Errorf(
				"invalid date, expected %#v got %#v",
				expect.date,
				r.date,
			)
		}
	}
}

func TestParseStaffMeCmd(t *testing.T) {
	inputs := []string{
		"!staffme finish writing unit tests | 18:30",
		"!staffme finish writing unit tests",
	}

	y, m, d := time.Now().Date()
	expects := []staffMeCommand{
		{
			kind:       commandStaffMe,
			identifier: "finish writing unit tests",
			hasDueDate: true,
			date:       date{day: d, month: m, year: y, hour: 18, min: 30},
		},
		{
			kind:       commandStaffMe,
			identifier: "finish writing unit tests",
			hasDueDate: false,
		},
	}

	for i, input := range inputs {
		t.Logf("input %d", i)
		result, err := parseCommand(input)
		if !err.isOK() {
			t.Errorf("parsing error: %s", err.details)
		}

		expect := expects[i]
		switch r := result.(type) {
		case *staffMeCommand:
			if r.kind != expect.kind {
				t.Errorf(
					"invalid kind, expected %d got %d",
					expect.kind,
					r.kind,
				)
			}

			if r.identifier != expect.identifier {
				t.Errorf(
					"invalid identifier, expected %s got %s",
					expect.identifier,
					r.identifier,
				)
			}

			if r.hasDueDate != r.hasDueDate {
				t.Errorf(
					"invalid due date flag, expected %t got %t",
					expect.hasDueDate,
					r.hasDueDate,
				)
			}
			if r.hasDueDate && !r.date.isEqual(expect.date) {
				t.Errorf(
					"invalid date, expected %#v got %#v",
					expect.date,
					r.date,
				)
			}
		}
	}
}

func TestParseRemoveMeCmd(t *testing.T) {
	inputs := []string{
		"!removeme task | writing unit tests",
		"!removeme reminder | writing unit tests",
	}

	expects := []removeMeCommand{
		{
			kind:       commandRemoveMe,
			list:       token{kind: tokenTask},
			identifier: "writing unit tests",
		},
		{
			kind:       commandRemoveMe,
			list:       token{kind: tokenReminder},
			identifier: "writing unit tests",
		},
	}

	for i, input := range inputs {
		t.Logf("input %d", i)
		result, err := parseCommand(input)
		if !err.isOK() {
			t.Errorf("parsing error: %s", err.details)
		}

		expect := expects[i]
		switch r := result.(type) {
		case *removeMeCommand:
			if r.kind != expect.kind {
				t.Errorf(
					"invalid kind, expected %d got %d",
					expect.kind,
					r.kind,
				)
			}

			if r.list.kind != expect.list.kind {
				t.Errorf(
					"invalid keyword, expected %s got %s",
					tokenKindString[expect.list.kind],
					tokenKindString[r.list.kind],
				)
			}

			if r.identifier != expect.identifier {
				t.Errorf(
					"invalid identifier, expected %s got %s",
					expect.identifier,
					r.identifier,
				)
			}
		}
	}
}

package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

type Scanner struct {
	reader         io.Reader
	bufferedReader *bufio.Reader
	pos            int
}

func NewScanner(reader io.Reader) (scanner *Scanner) {
	scanner = &Scanner{reader: reader}
	if buf, ok := reader.(*bufio.Reader); ok {
		scanner.bufferedReader = buf
	} else {
		scanner.bufferedReader = bufio.NewReader(reader)
	}
	return
}

func (s *Scanner) SkipSpace() {
	reader := s.bufferedReader
	for r, _, err := reader.ReadRune(); err == nil && unicode.IsSpace(r); r, _, err = reader.ReadRune() {
		s.pos++
	}
	reader.UnreadRune()
	return
}

func (s *Scanner) IsDone() bool {
	s.SkipSpace()
	if _, _, err := s.bufferedReader.ReadRune(); err != nil {
		return true
	}
	s.bufferedReader.UnreadRune()
	return false
}

func (s *Scanner) Scan() (token string, pos int, err error) {
	reader := s.bufferedReader

	s.SkipSpace()
	pos = s.pos

	first, _, err := reader.ReadRune()
	if err != nil {
		return
	}
	switch first {
	case '(':
		token = "("
		return
	case ')':
		token = ")"
		return
	case '\'':
		token = "'"
		return
	}

	var tok []rune
	tok = append(tok, first)
	s.pos++
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			break
		}
		if r == '(' || r == ')' || unicode.IsSpace(r) {
			reader.UnreadRune()
			break
		}
		tok = append(tok, r)
		s.pos++
	}
	token = string(tok)
	pos = s.pos
	return
}

type ParseError struct {
	pos    int
	reason string
}

func (e ParseError) Error() string {
	return fmt.Sprintf("parse error: pos %d: %s", e.pos, e.reason)
}

func parseElement(scanner *Scanner, literal bool, inQuotedList bool, topLevel bool) (result Expression, err error) {
	token, pos, err := scanner.Scan()
	if err == io.EOF && !topLevel {
		return nil, ParseError{pos, "expecting \")\""}
	}
	if err != nil {
		return nil, err
	}
	switch token {
	case ")":
		if topLevel {
			return PTBlank, ParseError{pos, "unexpected \")\""}
		}
		return Symbol(")"), nil
	case "(":
		return parseList(scanner, literal || inQuotedList, false)
	case "'":
		if literal {
			return nil, ParseError{pos, "unexpected quote in quoted expression"}
		}
		result, err = parseElement(scanner, true, inQuotedList, topLevel)
		if err != nil {
			err = ParseError{pos, "expected something to quote"}
		}
		return
	default:
		if literal {
			return QuotedSymbol(token), nil
		}
		return Atomize(token), nil
	}
}

func parseList(scanner *Scanner, quoted bool, topLevel bool) (list *SexpPair, err error) {
	dummy := &SexpPair{PTBlank, EmptyList, quoted}
	tail := dummy
	for element, err := parseElement(scanner, false, quoted, topLevel); element != Symbol(")"); element, err = parseElement(scanner, false, quoted, topLevel) {

		if err == io.EOF && topLevel {
			return dummy.next.(*SexpPair), nil
		}
		if err != nil {
			return dummy.next.(*SexpPair), err
		}

		nextPair := &SexpPair{element, EmptyList, quoted}
		tail.next = nextPair
		tail = nextPair
	}
	return dummy.next.(*SexpPair), nil
}

func Parse(scanner *Scanner) (sexps []Expression, err error) {
	sexp, err := parseList(scanner, false, true)
	sexps = ToSlice(sexp)
	return
}

func ParseLine(line string) (sexps []Expression, err error) {
	scanner := NewScanner(strings.NewReader(line))
	return Parse(scanner)
}

// Atomize infers the data type of a raw string and returns the string converted to this type.
// If it fails to safely convert the string, it simply returns it as a string again.
func Atomize(str string) Expression {
	// First, try to Atomize it as an integer
	if i, err := strconv.ParseInt(str, 10, 32); err == nil {
		return PTInt(int(i))
	}

	// That didn't work? Maybe it's a float
	if f, err := strconv.ParseFloat(str, 64); err == nil {
		return PTFloat(f)
	}

	// Is it a boolean? (special kind of string)
	switch str {
	case "#t":
		return PTBool(true)
	case "#f":
		return PTBool(false)
	}

	// Fuck it; it's a symbol
	return Symbol(str)
}

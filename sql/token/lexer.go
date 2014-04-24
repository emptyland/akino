package token

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type Lexer struct {
	input    io.RuneReader
	pos      int
	last     int
	lahError error
	lahRune  rune
	lahN     int
}

func (self *Lexer) InitWithRuneReader(input io.RuneReader) {
	self.input = input
	self.lahRune, self.lahN, self.lahError = self.input.ReadRune()
}

func (self *Lexer) InitWithReader(input io.Reader) {
	self.InitWithRuneReader(bufio.NewReader(input))
}

func (self *Lexer) Init(input string) {
	self.InitWithRuneReader(strings.NewReader(input))
}

func NewLexer(input string) *Lexer {
	var rv Lexer
	rv.Init(input)
	return &rv
}

func (self *Lexer) Next() (int, Token, string) {
	r, err := self.peek()
	if err != nil {
		return self.eof(err)
	}

	switch r {
	case '`':
		return self.readIdentifier(true)

	case '\'', '"':
		return self.readString(r)

	case '/':
		return self.readSlashPrefix()

	case '-':
		return self.readMinusPrefix()

	case '=':
		return self.readEqualPrefix()

	case '<':
		return self.readLessPrefix()

	case '>':
		return self.readGreatPrefix()

	case '+':
		return self.readToken(PLUS)

	case '*':
		return self.readToken(STAR)

	case '(':
		return self.readToken(LPAREN)

	case ')':
		return self.readToken(RPAREN)

	case '.':
		return self.readToken(DOT)

	case ',':
		return self.readToken(COMMA)

	case ';':
		return self.readToken(SEMI)

	default:
		return self.advance(r)
	}
}

func (self *Lexer) Error() error {
	return self.lahError
}

func (self *Lexer) advance(r rune) (int, Token, string) {
	switch {
	case unicode.IsLetter(r):
		return self.readIdOrKeyword()

	case unicode.IsSpace(r):
		return self.eatSpaceRunes()

	case unicode.IsDigit(r):
		return self.readNumber()

	default:
		return self.illegal("Illegal token rune")
	}
}

func (self *Lexer) readIdOrKeyword() (int, Token, string) {
	pos, rv, lit := self.readIdentifier(false)
	if rv == ILLEGAL {
		return pos, rv, lit
	}
	if tok, found := Keyword[strings.ToUpper(lit)]; found {
		return pos, tok, lit
	} else {
		return pos, rv, lit
	}
}

func (self *Lexer) readIdentifier(has_quote bool) (int, Token, string) {
	self.pos = self.last // Keep this token position

	if has_quote {
		if err := self.skip(); err != nil {
			return self.illegal("Bad quoted identifer, no body")
		}
	}

	r, err := self.read()
	if err != nil {
		return self.eof(err)
	}
	if !unicode.IsLetter(r) {
		return self.illegal("Bad identifier, should starts with a letter")
	}
	var sb bytes.Buffer
	if has_quote {
		sb.WriteRune('`')
		for r != '`' {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
				return self.illegal("Illegal identifier character")
			}
			sb.WriteRune(r)
			if r, err = self.read(); err != nil {
				return self.illegal("Unexpected '`' in end of identifer")
			}
		}
		sb.WriteRune('`')
	} else {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			sb.WriteRune(r)
		}
		for {
			if r, err = self.peek(); err != nil || (!unicode.IsLetter(r) && !unicode.IsDigit(r)) {
				break
			}
			sb.WriteRune(r)
			self.skip()
		}
	}
	return self.pos, ID, sb.String()
}

func (self *Lexer) readString(quote rune) (int, Token, string) {
	self.pos = self.last

	self.skip()
	var sb bytes.Buffer
	sb.WriteRune(quote)
	for {
		if r, err := self.peek(); err != nil {
			return self.illegal("Unexpected EOF in string literal")
		} else if isnewline(r) {
			return self.illegal("Unexpected new line in string literal")
		} else {
			r, _ = self.read()
			sb.WriteRune(r)
			if r == quote {
				break
			}
		}
	}
	return self.pos, STRING_LITERAL, sb.String()
}

func (self *Lexer) readSlashPrefix() (int, Token, string) {
	self.pos = self.last

	self.skip() // skip '/'
	r, err := self.peek()
	if r != '*' || err != nil {
		return self.pos, SLASH, "/"
	}
	self.skip() // skip '*'

	var sb bytes.Buffer
	sb.WriteString("/*")
	for {
		if r, err = self.read(); err != nil {
			return self.illegal(`Unexpected "*/" in end of comment`)
		}
		if isnewline(r) {
			return self.illegal(`New line in /* */ block is not allow`)
		}
		sb.WriteRune(r)
		if r == '*' {
			if r, err = self.peek(); err != nil {
				return self.illegal(`Unexpected "*/" in end of comment`)
			}
			if r == '/' {
				self.skip()
				break
			}
		}
	}
	sb.WriteString("*/")
	return self.pos, COMMENT, sb.String()
}

func (self *Lexer) readMinusPrefix() (int, Token, string) {
	self.pos = self.last

	self.skip() // skip '-'
	r, err := self.peek()
	if r != '-' || err != nil {
		return self.pos, MINUS, "-"
	}
	self.skip() // skip '-'
	// skip "--" total

	var sb bytes.Buffer
	sb.WriteString("--")
	for {
		if r, err = self.read(); err != nil {
			return self.illegal(`Unexpected "*/" in end of comment`)
		}
		sb.WriteRune(r)
		if r == '-' {
			if r, err = self.peek(); err != nil {
				return self.illegal(`Unexpected "*/" in end of comment`)
			}
			if r == '-' {
				self.skip()
				break
			}
		}
	}
	sb.WriteString("--")
	return self.pos, COMMENT, sb.String()
}

func (self *Lexer) readEqualPrefix() (int, Token, string) {
	return self.readEqualPostfix(EQ, EQ, "=")
}

func (self *Lexer) readLessPrefix() (int, Token, string) {
	self.pos = self.last

	self.skip() // skip '<'
	r, err := self.peek()
	switch {
	case err != nil:
		return self.illegal(`Bad "<" prefix token`)

	case r == '=':
		self.skip()
		return self.pos, LE, "<="

	case r == '>':
		self.skip()
		return self.pos, NE, "<>"

	default:
		return self.pos, LT, "<"
	}
}

func (self *Lexer) readGreatPrefix() (int, Token, string) {
	return self.readEqualPostfix(GT, GE, ">")
}

func (self *Lexer) readEqualPostfix(unary, binary Token, prefix string) (int, Token, string) {
	self.pos = self.last

	self.skip() // skip prefix
	r, err := self.peek()
	switch {
	case err != nil:
		return self.illegal(`Bad "` + prefix + `" prefix token`)

	case r == '=':
		self.skip()
		return self.pos, binary, prefix + "="

	default:
		return self.pos, unary, prefix
	}
}

func (self *Lexer) readNumber() (int, Token, string) {
	self.pos = self.last // Keep this token position

	var sb bytes.Buffer
	has_dot := false
	for {
		r, _ := self.peek()
		if unicode.IsDigit(r) {
			self.skip()
			sb.WriteRune(r)
		} else if r == '.' {
			if has_dot {
				return self.illegal("Bad floating number literal")
			} else {
				has_dot = true
			}
			self.skip()
			sb.WriteRune(r)
		} else if unicode.IsLetter(r) {
			return self.illegal("Bad floating number literal")
		} else {
			break
		}
	}
	if has_dot {
		return self.pos, FLOAT_LITERAL, sb.String()
	} else {
		return self.pos, INT_LITERAL, sb.String()
	}
}

func (self *Lexer) readToken(rv Token) (int, Token, string) {
	self.pos = self.last

	var sb bytes.Buffer
	r, err := self.read()
	if err != nil {
		return self.illegal("Bad operator token")
	}
	sb.WriteRune(r)
	return self.pos, rv, sb.String()
}

func (self *Lexer) eatSpaceRunes() (int, Token, string) {
	var r rune
	var err error

	for {
		if r, err = self.peek(); err != nil {
			return self.eof(err)
		}
		if !unicode.IsSpace(r) {
			break
		}
		self.skip()
	}
	return self.Next()
}

func (self *Lexer) peek() (rune, error) {
	rv := self.lahRune
	err := self.lahError
	return rv, err
}

func (self *Lexer) skip() error {
	_, err := self.read()
	return err
}

func (self *Lexer) read() (rune, error) {
	r, n, e := self.lahRune, self.lahN, self.lahError
	self.lahRune, self.lahN, self.lahError = self.input.ReadRune()
	self.last += n
	return r, e
}

func (self *Lexer) illegal(msg string) (int, Token, string) {
	return self.illegalByError(fmt.Errorf(msg))
}

func (self *Lexer) illegalByError(err error) (int, Token, string) {
	if err != nil {
		self.lahError = err
	}
	return self.last, ILLEGAL, ""
}

func (self *Lexer) eof(err error) (int, Token, string) {
	if err != nil {
		self.lahError = err
	}
	return self.last, EOF, ""
}

func isnewline(r rune) bool {
	return r == '\r' || r == '\n'
}

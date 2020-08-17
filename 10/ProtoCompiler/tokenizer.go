package main

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"
)

var keywords = [...]string{
	"class",
	"constructor",
	"function",
	"method",
	"field",
	"static",
	"var",
	"int",
	"char",
	"boolean",
	"void",
	"true",
	"false",
	"null",
	"this",
	"let",
	"do",
	"if",
	"else",
	"while",
	"return",
}

const (
	SYMBOLS = "{}()[].,;+-*/&|<>=~ \t"
	SLASH   = '/'
	STAR    = '*'
	SPACE   = ' '
	TAB     = '\t'
)

var re_identifier = regexp.MustCompile(`[A-Za-z_][0-9A-Za-z_]*`)

type Tokenizer interface {
	Advance() bool
	TokenType() int
	Keyword() string
	Symbol() string
	Identifier() string
	IntVal() string
	StringVal() string
}

type JackTokenizer struct {
	scn         *bufio.Scanner
	buf         []string
	token       string
	skipComment bool
}

func NewJackTokenizer(r io.Reader) *JackTokenizer {
	jtkn := new(JackTokenizer)
	jtkn.scn = bufio.NewScanner(r)
	jtkn.buf = make([]string, 0, 1024)
	jtkn.token = ""
	jtkn.skipComment = false
	return jtkn
}

func (jtkn *JackTokenizer) TokenType() int {
	if len(jtkn.token) == 1 && strings.Index(SYMBOLS, jtkn.token) != -1 {
		return SYMBOL
	}
	if len(jtkn.token) >= 2 && jtkn.token[0] == '"' && jtkn.token[len(jtkn.token)-1] == '"' {
		return STRING_CONST
	}
	if _, err := strconv.Atoi(jtkn.token); err == nil {
		return INT_CONST
	}
	for _, keyword := range keywords {
		if jtkn.token == keyword {
			return KEYWORD
		}
	}
	if re_identifier.MatchString(jtkn.token) {
		return IDENTIFIER
	}
	return INVALID_TOKEN
}

func (jtkn *JackTokenizer) Keyword() string {
	return jtkn.token
}

func (jtkn *JackTokenizer) Symbol() string {
	return jtkn.token
}

func (jtkn *JackTokenizer) Identifier() string {
	return jtkn.token
}

func (jtkn *JackTokenizer) IntVal() string {
	return jtkn.token
}

func (jtkn *JackTokenizer) StringVal() string {
	return jtkn.token
}

func (jtkn *JackTokenizer) parseLine(line string) {
	line = strings.TrimSpace(line)
	if len(line) == 0 {
		return
	}

	var bld strings.Builder
	var b rune
	b = 0

	for _, c := range line {
		if jtkn.skipComment {
			if c == SLASH && b == STAR {
				jtkn.skipComment = false
			}
			b = c
			continue
		}

		if strings.IndexRune(SYMBOLS, c) != -1 {
			if c == SLASH && b == SLASH {
				jtkn.buf = jtkn.buf[:len(jtkn.buf)-1]
				return
			}
			if c == STAR && b == SLASH {
				jtkn.buf = jtkn.buf[:len(jtkn.buf)-1]
				jtkn.skipComment = true
				b = 0
				continue
			}

			if bld.Len() > 0 {
				jtkn.buf = append(jtkn.buf, bld.String())
				bld.Reset()
			}

			if c != SPACE && c != TAB {
				jtkn.buf = append(jtkn.buf, string(c))
			}
		} else {
			bld.WriteRune(c)
		}
		b = c
	}

	if bld.Len() > 0 {
		jtkn.buf = append(jtkn.buf, bld.String())
	}
}

func (jtkn *JackTokenizer) Advance() bool {
	jtkn.token = ""

	for len(jtkn.buf) == 0 && jtkn.scn.Scan() {
		line := jtkn.scn.Text()
		jtkn.parseLine(line)
	}

	if len(jtkn.buf) > 0 {
		jtkn.token, jtkn.buf = jtkn.buf[0], jtkn.buf[1:]
	}

	return len(jtkn.token) != 0
}

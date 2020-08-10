package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

const (
	A_COMMAND = iota
	C_COMMAND
	L_COMMAND
	INVALID_COMMAND
)

type Parser interface {
	HasMoreCommands() bool
	Advance() error
	CommandType() (int, error)
	Symbol() (string, error)
	Dest() (string, error)
	Comp() (string, error)
	Jump() (string, error)
}

type CToken struct {
	dest string
	comp string
	jump string
}

type MyParser struct {
	asm_scanner *bufio.Scanner
	command     string
	cache       string
	token       interface{}
}

func NewMyParser(asm_code io.Reader) *MyParser {
	psr := new(MyParser)
	psr.asm_scanner = bufio.NewScanner(asm_code)
	psr.cache = ""

	return psr
}

func (psr *MyParser) Symbol() (string, error) {
	if symbol, ok := psr.token.(string); ok {
		return symbol, nil
	}
	return "", fmt.Errorf("Symbol: Couldn't convert into string")
}

func (psr *MyParser) Dest() (string, error) {
	if ctoken, ok := psr.token.(CToken); ok {
		return ctoken.dest, nil
	}
	return "", fmt.Errorf("Dest: Couldn't convert into CToken")
}

func (psr *MyParser) Comp() (string, error) {
	if ctoken, ok := psr.token.(CToken); ok {
		return ctoken.comp, nil
	}
	return "", fmt.Errorf("Comp: Couldn't convert into CToken")
}

func (psr *MyParser) Jump() (string, error) {
	if ctoken, ok := psr.token.(CToken); ok {
		return ctoken.jump, nil
	}
	return "", fmt.Errorf("Jump: Couldn't convert into CToken")
}

func (psr *MyParser) HasMoreCommands() bool {
	if len(psr.cache) > 0 {
		return true
	}

	command := ""
	for len(command) == 0 && psr.asm_scanner.Scan() {
		line := psr.asm_scanner.Text()
		if comm := strings.Index(line, "//"); comm != -1 {
			line = line[:comm]
		}
		command = strings.TrimSpace(line)
	}

	psr.cache = command
	return len(psr.cache) > 0
}

func (psr *MyParser) Advance() error {
	if len(psr.cache) == 0 {
		return fmt.Errorf("There is no command.")
	}

	psr.command = psr.cache

	cmd_type, err := psr.CommandType()
	if err != nil {
		return err
	}

	if cmd_type == C_COMMAND {
		ctoken, err := parseC(psr.command)
		if err != nil {
			return err
		}
		psr.token = ctoken
	} else if cmd_type == A_COMMAND {
		symbol, err := parseA(psr.command)
		if err != nil {
			return err
		}
		psr.token = symbol
	} else if cmd_type == L_COMMAND {
		symbol, err := parseL(psr.command)
		if err != nil {
			return err
		}
		psr.token = symbol
	}

	psr.cache = ""

	return nil
}

func parseC(cmd string) (result CToken, err error) {
	tokens := [3]string{}
	has_dest, has_jump := false, false

	idx := 0
	for _, c := range cmd {
		switch {
		case c == '=' && idx > 0 || c == ';' && idx > 1:
			return CToken{}, fmt.Errorf("parseC: Invalid command format.")
		case c == '=':
			has_dest = true
			idx++
		case c == ';':
			has_jump = true
			idx++
		default:
			tokens[idx] += string(c)
		}
	}

	idx = 0
	if has_dest {
		result.dest = tokens[idx]
		idx++
	}
	result.comp = tokens[idx]
	idx++
	if has_jump {
		result.jump = tokens[idx]
	}

	return
}

func parseA(cmd string) (string, error) {
	if cmd[0] != '@' {
		return "", fmt.Errorf("parseA: Invalid command format.")
	}
	return cmd[1:], nil
}

func parseL(cmd string) (string, error) {
	closing := strings.Index(cmd, ")")
	if cmd[0] != '(' || closing != len(cmd)-1 {
		return "", fmt.Errorf("parseL: Invalid label format.")
	}
	return cmd[1:closing], nil
}

func (psr *MyParser) CommandType() (int, error) {
	if len(psr.command) == 0 {
		return INVALID_COMMAND, fmt.Errorf("CommandType: No command is read.")
	}
	if psr.command[0] == '@' {
		return A_COMMAND, nil
	}
	if len(psr.command) >= 3 && psr.command[0] == '(' && psr.command[len(psr.command)-1] == ')' {
		return L_COMMAND, nil
	}
	return C_COMMAND, nil
}

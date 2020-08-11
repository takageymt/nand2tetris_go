package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type VMParser interface {
	Advance() bool
	CommandType() (int, error)
	Arg1() (string, error)
	Arg2() (int, error)
}

type MyVMParser struct {
	vm_source    *bufio.Scanner
	command_args []string
}

func NewMyVMParser(r io.Reader) *MyVMParser {
	psr := new(MyVMParser)
	psr.vm_source = bufio.NewScanner(r)
	psr.command_args = nil
	return psr
}

func (psr *MyVMParser) Advance() bool {
	psr.command_args = nil
	for psr.vm_source.Scan() {
		line := psr.vm_source.Text()
		comments := strings.Index(line, "//")
		if comments != -1 {
			line = line[:comments]
		}
		line = strings.TrimSpace(line)
		if len(line) != 0 {
			psr.command_args = strings.Split(line, " ")
			return true
		}
	}
	return false
}

func SliceContains(slice []string, value string) bool {
	for _, elem := range slice {
		if elem == value {
			return true
		}
	}
	return false
}

func (psr *MyVMParser) CommandType() (int, error) {
	if len(psr.command_args) == 0 {
		return INVALID_COMMAND, fmt.Errorf("CommandType: there is no command.")
	}
	if psr.command_args[0] == "push" {
		if len(psr.command_args) != 3 {
			return INVALID_COMMAND, fmt.Errorf("CommandType: 'push' is called with an invalid format.")
		}
		return C_PUSH, nil
	}
	if psr.command_args[0] == "pop" {
		if len(psr.command_args) != 3 {
			fmt.Println(psr.command_args)
			return INVALID_COMMAND, fmt.Errorf("CommandType: 'pop' is called with an invalid format.")
		}
		return C_POP, nil
	}
	if psr.command_args[0] == "label" {
		if len(psr.command_args) != 2 {
			return INVALID_COMMAND, fmt.Errorf("CommandType: 'label' is called with an invalid format.")
		}
		return C_LABEL, nil
	}
	if psr.command_args[0] == "goto" {
		if len(psr.command_args) != 2 {
			return INVALID_COMMAND, fmt.Errorf("CommandType: 'goto' is called with an invalid format.")
		}
		return C_GOTO, nil
	}
	if psr.command_args[0] == "if-goto" {
		if len(psr.command_args) != 2 {
			return INVALID_COMMAND, fmt.Errorf("CommandType: 'if-goto' is called with an invalid format.")
		}
		return C_IF, nil
	}
	if psr.command_args[0] == "function" {
		if len(psr.command_args) != 3 {
			return INVALID_COMMAND, fmt.Errorf("CommandType: 'function' is called with an invalid format.")
		}
		return C_FUNCTION, nil
	}
	if psr.command_args[0] == "return" {
		if len(psr.command_args) != 1 {
			return INVALID_COMMAND, fmt.Errorf("CommandType: 'return' is called with an invalid format.")
		}
		return C_RETURN, nil
	}
	if psr.command_args[0] == "call" {
		if len(psr.command_args) != 3 {
			return INVALID_COMMAND, fmt.Errorf("CommandType: 'call' is called with an invalid format.")
		}
		return C_CALL, nil
	}
	if SliceContains(ARITHMETIC_COMMANDS[:], psr.command_args[0]) {
		if len(psr.command_args) != 1 {
			return INVALID_COMMAND, fmt.Errorf("CommandType: '%s' is called with an invalid format.", psr.command_args[0])
		}
		return C_ARITHMETIC, nil
	}
	return INVALID_COMMAND, fmt.Errorf("CommandType: undefined command is found.")
}

func (psr *MyVMParser) Arg1() (string, error) {
	command_type, err := psr.CommandType()
	if err != nil {
		return "", err
	}
	if command_type == C_ARITHMETIC && len(psr.command_args) == 1 {
		return psr.command_args[0], nil
	}
	if len(psr.command_args) > 1 {
		return psr.command_args[1], nil
	}
	return "", fmt.Errorf("Arg1: Abnormal behaviour is detected!")
}

func (psr *MyVMParser) Arg2() (int, error) {
	if len(psr.command_args) > 2 {
		return strconv.Atoi(psr.command_args[2])
	}
	return -1, fmt.Errorf("Arg1: Abnormal behaviour is detected!")
}

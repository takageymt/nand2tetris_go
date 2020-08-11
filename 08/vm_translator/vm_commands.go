package main

const (
	C_ARITHMETIC = iota
	C_PUSH
	C_POP
	C_LABEL
	C_GOTO
	C_IF
	C_FUNCTION
	C_RETURN
	C_CALL
	INVALID_COMMAND
)

var ARITHMETIC_COMMANDS [9]string = [9]string{
	"add", "sub", "neg", "eq", "gt", "lt", "and", "or", "not",
}

const (
	POINTER_BASE = 3
	TEMP_BASE    = 5
)

package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

type VMCodeWriter interface {
	SetFileName(filename string)
	WriteInit() error
	WriteArithmetic(command string) error
	WritePushPop(command int, segment string, index int) error
	WriteLabel(label string) error
	WriteGoto(label string) error
	WriteIf(label string) error
	WriteCall(funcName string, numArgs int) error
	WriteFunction(funcName string, numLocals int) error
	WriteReturn() error
	Close()
}

type MyVMCodeWriter struct {
	bw        *bufio.Writer
	src_file  string
	jump_id   int
	func_name string
}

func NewMyVMCodeWriter(w io.Writer) *MyVMCodeWriter {
	cwr := new(MyVMCodeWriter)
	cwr.bw = bufio.NewWriter(w)
	cwr.src_file = ""
	cwr.jump_id = 0
	cwr.func_name = ""
	return cwr
}

func (cwr *MyVMCodeWriter) SetFileName(filename string) {
	cwr.src_file = filename
	cwr.jump_id = 0
}

func (cwr *MyVMCodeWriter) writeCode(code string) (err error) {
	_, err = cwr.bw.WriteString(code + "\n")
	return
}

func (cwr *MyVMCodeWriter) writePopWorkStack() error {
	return cwr.writeCode(`@SP // pop
M=M-1
A=M // pop`)
}

func (cwr *MyVMCodeWriter) writePushWorkStack() error {
	return cwr.writeCode(`@SP // push
A=M
M=D
@SP
M=M+1 // push`)
}

func (cwr *MyVMCodeWriter) writeUnary(asm_inst string) (err error) {
	err = cwr.writePopWorkStack()
	if err != nil {
		return
	}
	err = cwr.writeCode(asm_inst)
	if err != nil {
		return
	}
	err = cwr.writePushWorkStack()
	return
}

func (cwr *MyVMCodeWriter) writeBinary(asm_inst string) (err error) {
	err = cwr.writePopWorkStack()
	if err != nil {
		return
	}
	err = cwr.writeCode("D=M")
	if err != nil {
		return
	}
	err = cwr.writePopWorkStack()
	if err != nil {
		return
	}
	err = cwr.writeCode(asm_inst)
	if err != nil {
		return
	}
	err = cwr.writePushWorkStack()
	return
}

func (cwr *MyVMCodeWriter) getNewLabel() string {
	cwr.jump_id++
	return cwr.src_file + "::" + strconv.Itoa(cwr.jump_id)
}

func (cwr *MyVMCodeWriter) writeCompare(asm_inst string) (err error) {
	err = cwr.writePopWorkStack()
	if err != nil {
		return
	}
	err = cwr.writeCode("D=M")
	if err != nil {
		return
	}
	err = cwr.writePopWorkStack()
	if err != nil {
		return
	}
	err = cwr.writeCode("D=M-D")
	if err != nil {
		return
	}
	label := cwr.getNewLabel()
	endif := cwr.getNewLabel()
	err = cwr.writeCode(fmt.Sprintf(`@%s
%s
D=0
@%s
0;JMP
(%s)
D=-1
(%s)`, label, asm_inst, endif, label, endif))
	if err != nil {
		return
	}
	err = cwr.writePushWorkStack()
	return
}

func (cwr *MyVMCodeWriter) WriteInit() error {
	err := cwr.writeCode(`@256
D=A
@SP
M=D`)
	if err != nil {
		return err
	}
	return cwr.WriteCall("Sys.init", 0)
}

func (cwr *MyVMCodeWriter) WriteArithmetic(command string) error {
	switch command {
	case "neg":
		return cwr.writeUnary("D=-M")
	case "not":
		return cwr.writeUnary("D=!M")
	case "add":
		return cwr.writeBinary("D=D+M")
	case "sub":
		return cwr.writeBinary("D=M-D")
	case "and":
		return cwr.writeBinary("D=D&M")
	case "or":
		return cwr.writeBinary("D=D|M")
	case "eq":
		return cwr.writeCompare("D;JEQ")
	case "gt":
		return cwr.writeCompare("D;JGT")
	case "lt":
		return cwr.writeCompare("D;JLT")
	}
	return fmt.Errorf("WriteArithmetic: Undefined command is given.")
}

func (cwr *MyVMCodeWriter) writePushConstant(value int) (err error) {
	err = cwr.writeCode(fmt.Sprintf(`@%d
D=A`, value))
	if err != nil {
		return
	}
	err = cwr.writePushWorkStack()
	return
}

func (cwr *MyVMCodeWriter) writePushThroughRegister(segment string, index int) (err error) {
	var register string
	switch segment {
	case "local":
		register = "LCL"
	case "argument":
		register = "ARG"
	case "this":
		register = "THIS"
	case "that":
		register = "THAT"
	default:
		return fmt.Errorf("writePopThroughRegister: Invalid segment is given.")
	}
	err = cwr.writeCode(fmt.Sprintf(`@%d
D=A
@%s
A=D+M
D=M`, index, register))
	if err != nil {
		return
	}
	err = cwr.writePushWorkStack()
	return
}

func (cwr *MyVMCodeWriter) writePopThroughRegister(segment string, index int) (err error) {
	var register string
	switch segment {
	case "local":
		register = "LCL"
	case "argument":
		register = "ARG"
	case "this":
		register = "THIS"
	case "that":
		register = "THAT"
	default:
		return fmt.Errorf("writePopThroughRegister: Invalid segment is given.")
	}
	err = cwr.writeCode(fmt.Sprintf(`@%d
D=A
@%s
D=D+M
@R13
M=D`, index, register))
	if err != nil {
		return
	}
	err = cwr.writePopWorkStack()
	if err != nil {
		return
	}
	err = cwr.writeCode(`D=M
@R13
A=M
M=D`)
	return
}

func (cwr *MyVMCodeWriter) writePushByBaseAddress(segment string, index int) (err error) {
	var base_addr int
	switch segment {
	case "pointer":
		base_addr = POINTER_BASE
	case "temp":
		base_addr = TEMP_BASE
	default:
		return fmt.Errorf("writePushByBaseAddress: Invalid segment is given.")
	}
	err = cwr.writeCode(fmt.Sprintf(`@%d
D=A
@%d
A=D+A
D=M`, index, base_addr))
	if err != nil {
		return
	}
	err = cwr.writePushWorkStack()
	return
}

func (cwr *MyVMCodeWriter) writePopByBaseAddress(segment string, index int) (err error) {
	var base_addr int
	switch segment {
	case "pointer":
		base_addr = POINTER_BASE
	case "temp":
		base_addr = TEMP_BASE
	default:
		return fmt.Errorf("writePopByBaseAddress: Invalid segment is given.")
	}
	err = cwr.writeCode(fmt.Sprintf(`@%d
D=A
@%d
D=D+A
@R13
M=D`, index, base_addr))
	if err != nil {
		return
	}
	err = cwr.writePopWorkStack()
	if err != nil {
		return
	}
	err = cwr.writeCode(`D=M
@R13
A=M
M=D`)
	return
}

func (cwr *MyVMCodeWriter) writePushStatic(index int) (err error) {
	err = cwr.writeCode(fmt.Sprintf(`@%s.%d
D=M`, cwr.src_file, index))
	if err != nil {
		return
	}
	err = cwr.writePushWorkStack()
	return
}

func (cwr *MyVMCodeWriter) writePopStatic(index int) (err error) {
	err = cwr.writePopWorkStack()
	if err != nil {
		return
	}
	err = cwr.writeCode(fmt.Sprintf(`D=M
@%s.%d
M=D`, cwr.src_file, index))
	return
}

func (cwr *MyVMCodeWriter) WritePushPop(command int, segment string, index int) error {
	if command == C_PUSH {
		switch segment {
		case "constant":
			return cwr.writePushConstant(index)
		case "local", "argument", "this", "that":
			return cwr.writePushThroughRegister(segment, index)
		case "pointer", "temp":
			return cwr.writePushByBaseAddress(segment, index)
		case "static":
			return cwr.writePushStatic(index)
		}
	} else if command == C_POP {
		switch segment {
		case "constant":
			return cwr.writePopWorkStack()
		case "local", "argument", "this", "that":
			return cwr.writePopThroughRegister(segment, index)
		case "pointer", "temp":
			return cwr.writePopByBaseAddress(segment, index)
		case "static":
			return cwr.writePopStatic(index)
		}
	}
	return fmt.Errorf("writePushPop: Abnormal behaviour is detected!")
}

func (cwr *MyVMCodeWriter) WriteLabel(label string) error {
	return cwr.writeCode(fmt.Sprintf("(%s$%s)", cwr.func_name, label))
}

func (cwr *MyVMCodeWriter) WriteGoto(label string) error {
	return cwr.writeCode(fmt.Sprintf(`@%s$%s
0;JMP`, cwr.func_name, label))
}

func (cwr *MyVMCodeWriter) WriteIf(label string) (err error) {
	err = cwr.writePopWorkStack()
	if err != nil {
		return
	}
	err = cwr.writeCode(fmt.Sprintf(`D=M
@%s$%s
D;JNE`, cwr.func_name, label))
	return
}

func (cwr *MyVMCodeWriter) WriteCall(funcName string, numArgs int) (err error) {
	label := cwr.getNewLabel()
	cwr.writeCode("// call")
	to_save := [...]string{label, "LCL", "ARG", "THIS", "THAT"}
	for _, l := range to_save {
		if l == label {
			err = cwr.writeCode(fmt.Sprintf(`@%s
D=A`, l))
		} else {
			err = cwr.writeCode(fmt.Sprintf(`@%s
D=M`, l))
		}
		if err != nil {
			return
		}
		err = cwr.writePushWorkStack()
		if err != nil {
			return
		}
	}
	err = cwr.writeCode(fmt.Sprintf(`@SP
D=M
@LCL
M=D
@%d
D=D-A
@ARG
M=D
@%s
0;JMP // call
(%s)`, numArgs+5, funcName, label))
	return
}

func (cwr *MyVMCodeWriter) WriteReturn() (err error) {
	// 順番に気をつけないと引数がない時に戻り値でリターンアドレスを上書きしてしまう
	err = cwr.writeCode("// return")
	err = cwr.writeCode(`@LCL
D=M
@5
A=D-A
D=M
@R14
M=D
`)
	if err != nil {
		return
	}
	err = cwr.writePopWorkStack()
	if err != nil {
		return
	}
	err = cwr.writeCode(`D=M
@ARG
A=M
M=D
@ARG
D=M+1
@SP
M=D
@LCL
D=M
@R13
AM=D-1
D=M
@THAT
M=D
@R13
AM=M-1
D=M
@THIS
M=D
@R13
AM=M-1
D=M
@ARG
M=D
@R13
AM=M-1
D=M
@LCL
M=D
@R14
A=M
0;JMP // return`)
	return
}

func (cwr *MyVMCodeWriter) WriteFunction(funcName string, numLocals int) (err error) {
	cwr.func_name = funcName
	err = cwr.writeCode(fmt.Sprintf(`(%s)
D=0`, funcName))
	if err != nil {
		return
	}
	for i := 0; i < numLocals; i++ {
		err = cwr.writePushWorkStack()
		if err != nil {
			return
		}
	}
	return
}

func (cwr *MyVMCodeWriter) Close() {
	label := cwr.getNewLabel()
	cwr.writeCode(fmt.Sprintf(`(%s)
@%s
0;JMP`, label, label))
	cwr.bw.Flush()
}

package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

type VMCodeWriter interface {
	SetFileName(filename string)
	WriteArithmetic(command string) error
	WritePushPop(command int, segment string, index int) error
	Close()
}

type MyVMCodeWriter struct {
	bw       *bufio.Writer
	src_file string
	jump_id  int
}

func NewMyVMCodeWriter(w io.Writer) *MyVMCodeWriter {
	cwr := new(MyVMCodeWriter)
	cwr.bw = bufio.NewWriter(w)
	cwr.src_file = ""
	cwr.jump_id = 0
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
	return cwr.writeCode(`@SP
M=M-1
A=M`)
}

func (cwr *MyVMCodeWriter) writePushWorkStack() error {
	return cwr.writeCode(`@SP
A=M
M=D
@SP
M=M+1`)
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
	return cwr.src_file + "$" + strconv.Itoa(cwr.jump_id)
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

func (cwr *MyVMCodeWriter) WriteArithmetic(command string) (err error) {
	if command == "neg" {
		err = cwr.writeUnary("D=-M")
	} else if command == "not" {
		err = cwr.writeUnary("D=!M")
	} else if command == "add" {
		err = cwr.writeBinary("D=D+M")
	} else if command == "sub" {
		err = cwr.writeBinary("D=M-D")
	} else if command == "and" {
		err = cwr.writeBinary("D=D&M")
	} else if command == "or" {
		err = cwr.writeBinary("D=D|M")
	} else if command == "eq" {
		err = cwr.writeCompare("D;JEQ")
	} else if command == "gt" {
		err = cwr.writeCompare("D;JGT")
	} else if command == "lt" {
		err = cwr.writeCompare("D;JLT")
	} else {
		err = fmt.Errorf("WriteArithmetic: Abnormal behaviour is detected!")
	}
	return
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

func (cwr *MyVMCodeWriter) WritePushPop(command int, segment string, index int) (err error) {
	if command == C_PUSH {
		switch segment {
		case "constant":
			err = cwr.writePushConstant(index)
		case "local", "argument", "this", "that":
			err = cwr.writePushThroughRegister(segment, index)
		case "pointer", "temp":
			err = cwr.writePushByBaseAddress(segment, index)
		case "static":
			err = cwr.writePushStatic(index)
		}
	} else if command == C_POP {
		switch segment {
		case "constant":
			err = cwr.writePopWorkStack()
		case "local", "argument", "this", "that":
			err = cwr.writePopThroughRegister(segment, index)
		case "pointer", "temp":
			err = cwr.writePopByBaseAddress(segment, index)
		case "static":
			err = cwr.writePopStatic(index)
		}
	} else {
		err = fmt.Errorf("writePushPop: Abnormal behaviour is detected!")
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

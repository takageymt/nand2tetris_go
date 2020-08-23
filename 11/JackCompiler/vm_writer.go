package main

import (
	"bufio"
	"fmt"
	"io"
)

const (
	SEG_CONST = iota
	SEG_ARG
	SEG_LOCAL
	SEG_STATIC
	SEG_THIS
	SEG_THAT
	SEG_POINTER
	SEG_TEMP
)

func SegmentString(segment int) string {
	switch segment {
	case SEG_CONST:
		return "constant"
	case SEG_ARG:
		return "argument"
	case SEG_LOCAL:
		return "local"
	case SEG_STATIC:
		return "static"
	case SEG_THIS:
		return "this"
	case SEG_THAT:
		return "that"
	case SEG_POINTER:
		return "pointer"
	case SEG_TEMP:
		return "temp"
	}
	return ""
}

const (
	COM_ADD = iota
	COM_SUB
	COM_NEG
	COM_EQ
	COM_GT
	COM_LT
	COM_AND
	COM_OR
	COM_NOT
)

type VMWriter interface {
	SetClass(name string)
	WritePush(segment, index int)
	WritePop(segment, index int)
	WriteArithmetic(command int)
	WriteLabel(label string)
	WriteGoto(label string)
	WriteIf(label string)
	WriteCall(name string, nArgs int)
	WriteFunction(name string, nLocals int)
	WriteReturn()
	NewLabel() string
	Close()
}

type JackVMWriter struct {
	ch     chan Data
	bw     *bufio.Writer
	class  string
	nLabel int
}

func NewJackVMWriter(w io.Writer) *JackVMWriter {
	jw := new(JackVMWriter)
	jw.bw = bufio.NewWriter(w)
	return jw
}

func (jw *JackVMWriter) NewLabel() string {
	jw.nLabel++
	return fmt.Sprintf("%s::%d", jw.class, jw.nLabel)
}

func (jw *JackVMWriter) SetClass(name string) {
	jw.class = name
	jw.nLabel = 0
}

func (jw *JackVMWriter) Write(code string) {
	jw.bw.WriteString(code + "\n")
}

func (jw *JackVMWriter) WritePush(segment, index int) {
	segs := SegmentString(segment)
	jw.Write(fmt.Sprintf("push %s %d", segs, index))
}

func (jw *JackVMWriter) WritePop(segment, index int) {
	segs := SegmentString(segment)
	jw.Write(fmt.Sprintf("pop %s %d", segs, index))
}

func (jw *JackVMWriter) WriteArithmetic(command int) {
	var cmd string
	switch command {
	case COM_ADD:
		cmd = "add"
	case COM_SUB:
		cmd = "sub"
	case COM_NEG:
		cmd = "neg"
	case COM_EQ:
		cmd = "eq"
	case COM_GT:
		cmd = "gt"
	case COM_LT:
		cmd = "lt"
	case COM_AND:
		cmd = "and"
	case COM_OR:
		cmd = "or"
	case COM_NOT:
		cmd = "not"
	}
	jw.Write(cmd)
}

func (jw *JackVMWriter) WriteLabel(label string) {
	jw.Write(fmt.Sprintf("label %s", label))
}

func (jw *JackVMWriter) WriteGoto(label string) {
	jw.Write(fmt.Sprintf("goto %s", label))
}

func (jw *JackVMWriter) WriteIf(label string) {
	jw.Write(fmt.Sprintf("if-goto %s", label))
}

func (jw *JackVMWriter) WriteCall(name string, nArgs int) {
	jw.Write(fmt.Sprintf("call %s %d", name, nArgs))
}

func (jw *JackVMWriter) WriteFunction(name string, nLocals int) {
	jw.Write(fmt.Sprintf("function %s.%s %d", jw.class, name, nLocals))
}

func (jw *JackVMWriter) WriteReturn() {
	jw.Write("return")
}

func (jw *JackVMWriter) Close() {
	jw.bw.Flush()
}

package main

import "fmt"

const (
	STATIC = iota
	FIELD
	ARG
	VAR
	NONE
)

const (
	CONSTRUCTOR = iota
	FUNCTION
	METHOD
)

type SymbolTable interface {
	EnterSubroutine(routineType int)
	ExitSubroutine()
	Define(name, varType string, kind int) int
	VarCount(kind int) int
	KindOf(name string) int
	TypeOf(name string) string
	IndexOf(name string) int
}

type SymbolAttr struct {
	Name  string
	Type  string
	Kind  int
	Index int
}

type SymbolTableNode struct {
	Table  map[string]SymbolAttr
	Counts [NONE]int
	Next   *SymbolTableNode
}

func NewSymbolTableNode(next *SymbolTableNode) *SymbolTableNode {
	node := new(SymbolTableNode)
	node.Table = make(map[string]SymbolAttr)
	node.Next = next
	return node
}

type SymbolTableList struct {
	head *SymbolTableNode
}

func NewSymbolTableList() *SymbolTableList {
	stl := new(SymbolTableList)
	stl.head = NewSymbolTableNode(nil)
	return stl
}

func (stl *SymbolTableList) EnterSubroutine(routineType int) {
	stl.head = NewSymbolTableNode(stl.head)
	if routineType == METHOD {
		stl.head.Counts[ARG]++
	}
}

func (stl *SymbolTableList) ExitSubroutine() {
	if stl.head == nil {
		fmt.Println("Invalid exit")
		return
	}
	stl.head = stl.head.Next
}

func (stl *SymbolTableList) Define(name, varType string, kind int) int {
	index := stl.head.Counts[kind]
	stl.head.Counts[kind]++
	stl.head.Table[name] = SymbolAttr{Name: name, Type: varType, Kind: kind, Index: index}
	return index
}

func (stl *SymbolTableList) VarCount(kind int) int {
	return stl.head.Counts[kind]
}

func (stl *SymbolTableList) FindSymbol(name string) SymbolAttr {
	for p := stl.head; p != nil; p = p.Next {
		if v, ok := p.Table[name]; ok {
			return v
		}
	}
	return SymbolAttr{Name: "", Type: "", Kind: NONE, Index: -1}
}

func (stl *SymbolTableList) KindOf(name string) int {
	item := stl.FindSymbol(name)
	return item.Kind
}

func (stl *SymbolTableList) TypeOf(name string) string {
	item := stl.FindSymbol(name)
	if item.Kind == NONE {
		return ""
	}
	return item.Type
}

func (stl *SymbolTableList) IndexOf(name string) int {
	item := stl.FindSymbol(name)
	if item.Kind == NONE {
		return -1
	}
	return item.Index
}

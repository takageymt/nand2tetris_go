package main

import (
	"fmt"
	"strconv"
)

type SymbolTable interface {
	AddEntry(symbol string, address int) error
	Contains(symbol string) bool
	GetAddress(symbol string) (int, error)
}

type MySymbolTable struct {
	table map[string]int
}

func NewMySymbolTable() *MySymbolTable {
	sym_tbl := new(MySymbolTable)
	sym_tbl.table = map[string]int{
		"SP":     0x0000,
		"LCL":    0x0001,
		"ARG":    0x0002,
		"THIS":   0x0003,
		"THAT":   0x0004,
		"SCREEN": 0x4000,
		"KBD":    0x6000,
	}
	for i := 0; i < 16; i++ {
		sym_tbl.table["R"+strconv.Itoa(i)] = i
	}

	fmt.Printf("%v\n", sym_tbl)

	return sym_tbl
}

func (sym_tbl *MySymbolTable) AddEntry(symbol string, address int) error {
	if _, ok := sym_tbl.table[symbol]; ok {
		return fmt.Errorf("AddEntry: %v: a symbol is already added.", symbol)
	}
	sym_tbl.table[symbol] = address
	return nil
	//return fmt.Errorf("AddEntry: %v: Invalid symbol format.", symbol)
}

func (sym_tbl *MySymbolTable) Contains(symbol string) bool {
	_, ok := sym_tbl.table[symbol]
	return ok
}

func (sym_tbl *MySymbolTable) GetAddress(symbol string) (int, error) {
	if address, ok := sym_tbl.table[symbol]; ok {
		return address, nil
	}
	return -1, fmt.Errorf("getAddress: %v: No such entry.", symbol)
}

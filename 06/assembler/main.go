package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type Assembler interface {
}

type MyAssembler struct {
	Parser
	Code
	SymbolTable
	inst_address int
	var_address  int
}

func NewMyAssembler(psr Parser, cvr Code, sym_tbl SymbolTable) *MyAssembler {
	asmr := new(MyAssembler)
	asmr.Parser = psr
	asmr.Code = cvr
	asmr.SymbolTable = sym_tbl
	asmr.inst_address = 0
	asmr.var_address = 0x0010
	return asmr
}

func (asmr *MyAssembler) ACommand() (string, error) {
	symbol, err := asmr.Parser.Symbol()
	if err != nil {
		return "", err
	}

	if address, err := strconv.Atoi(symbol); err == nil {
		return "0" + fmt.Sprintf("%015b", address), nil
	}

	if !asmr.Contains(symbol) {
		if err := asmr.AddEntry(symbol, asmr.var_address); err != nil {
			return "", err
		}
		asmr.var_address++
	}

	address, err := asmr.GetAddress(symbol)
	if err != nil {
		return "", err
	}
	return "0" + fmt.Sprintf("%015b", address), nil
}

func (asmr *MyAssembler) CCommand() (string, error) {
	raw_comp, err := asmr.Parser.Comp()
	if err != nil {
		return "", err
	}
	bin_comp, err := asmr.Code.Comp(raw_comp)
	if err != nil {
		return "", err
	}

	raw_dest, err := asmr.Parser.Dest()
	if err != nil {
		return "", err
	}
	bin_dest, err := asmr.Code.Dest(raw_dest)
	if err != nil {
		return "", err
	}

	raw_jump, err := asmr.Parser.Jump()
	if err != nil {
		return "", err
	}
	bin_jump, err := asmr.Code.Jump(raw_jump)
	if err != nil {
		return "", err
	}

	return "111" + bin_comp + bin_dest + bin_jump, nil
}

func main() {
	asm_file := os.Args[1]
	ext := strings.LastIndex(asm_file, ".asm")
	if ext == -1 {
		log.Fatal("Not .asm file is given.")
	}

	hack_file := asm_file[:ext] + ".hack"

	cvr := NewMyConverter()
	sym_tbl := NewMySymbolTable()

	// preparation
	func() {
		r, err := os.Open(asm_file)
		if err != nil {
			log.Fatal(err)
		}
		defer r.Close()

		psr := NewMyParser(r)
		asmr := NewMyAssembler(psr, cvr, sym_tbl)

		for asmr.HasMoreCommands() {
			err = asmr.Advance()
			if err != nil {
				log.Fatalf("%v: %v\n", asmr.inst_address, err.Error())
			}

			cmd_type, err := asmr.CommandType()
			if err != nil {
				log.Fatalf("%v: %v\n", asmr.inst_address, err.Error())
			}
			if cmd_type == A_COMMAND || cmd_type == C_COMMAND {
				asmr.inst_address++
			} else if cmd_type == L_COMMAND {
				symbol, err := asmr.Symbol()
				if err != nil {
					log.Fatalf("%v: %v\n", asmr.inst_address, err.Error())
				}
				if !asmr.Contains(symbol) {
					asmr.AddEntry(symbol, asmr.inst_address)
				}
			}
		}
	}()

	r, err := os.Open(asm_file)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	w, err := os.Create(hack_file)
	if err != nil {
		log.Fatal(err)
	}
	defer w.Close()

	bw := bufio.NewWriter(w)

	psr := NewMyParser(r)
	asmr := NewMyAssembler(psr, cvr, sym_tbl)

	for asmr.HasMoreCommands() {
		err = asmr.Advance()
		if err != nil {
			os.Remove(hack_file)
			log.Fatalf("%v: %v\n", asmr.inst_address, err.Error())
		}

		cmd_type, err := asmr.CommandType()
		if err != nil {
			os.Remove(hack_file)
			log.Fatalf("%v: %v\n", asmr.inst_address, err.Error())
		}

		if cmd_type == C_COMMAND {
			asmr.inst_address++
			instruction, err := asmr.CCommand()
			if err != nil {
				os.Remove(hack_file)
				log.Fatalf("%v: %v\n", asmr.inst_address, err.Error())
			}
			bw.WriteString(instruction + "\n")
		} else if cmd_type == A_COMMAND {
			asmr.inst_address++
			instruction, err := asmr.ACommand()
			if err != nil {
				os.Remove(hack_file)
				log.Fatalf("%v: %v\n", asmr.inst_address, err.Error())
			}
			bw.WriteString(instruction + "\n")
		}
	}
	bw.Flush()
}

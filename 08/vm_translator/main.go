package main

import (
	"fmt"
	"os"
	"strings"
)

func writeCommand(psr VMParser, cwr VMCodeWriter, command_type int) error {
	if command_type == C_PUSH || command_type == C_POP || command_type == C_FUNCTION || command_type == C_CALL {
		arg1, err := psr.Arg1()
		if err != nil {
			return err
		}
		arg2, err := psr.Arg2()
		if err != nil {
			return err
		}
		switch command_type {
		case C_PUSH, C_POP:
			return cwr.WritePushPop(command_type, arg1, arg2)
		case C_FUNCTION:
			return cwr.WriteFunction(arg1, arg2)
		case C_CALL:
			return cwr.WriteCall(arg1, arg2)
		}
		return fmt.Errorf("writeCommand: program should not reach this return but done.")
	} else if command_type != C_RETURN {
		arg1, err := psr.Arg1()
		if err != nil {
			return err
		}
		switch command_type {
		case C_ARITHMETIC:
			return cwr.WriteArithmetic(arg1)
		case C_LABEL:
			return cwr.WriteLabel(arg1)
		case C_GOTO:
			return cwr.WriteGoto(arg1)
		case C_IF:
			return cwr.WriteIf(arg1)
		}
		return fmt.Errorf("writeCommand: program should not reach this return but done.")
	}
	return cwr.WriteReturn()
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("main: No .vm file is given.")
		return
	}

	vm_files := make([]string, 0)
	asm_file := ""

	if strings.HasSuffix(os.Args[1], ".vm") {
		vm_file := os.Args[1]
		vm_files = append(vm_files, vm_file)
		asm_file = strings.TrimSuffix(vm_file, ".vm") + ".asm"
	} else {
		dirname := strings.TrimRight(os.Args[1], "/")
		basename := dirname
		lastDelim := strings.LastIndex(basename, "/")
		if lastDelim != -1 {
			basename = basename[lastDelim+1:]
		}
		asm_file = dirname + "/" + basename + ".asm"
		f, err := os.Open(dirname)
		if err != nil {
			fmt.Printf("main: %s: No such directory.\n")
			return
		}
		defer f.Close()
		fis, err := f.Readdir(0)
		if err != nil {
			fmt.Printf("main: %s is not a directory.\n")
			return
		}
		for _, fi := range fis {
			if !fi.IsDir() && strings.HasSuffix(fi.Name(), ".vm") {
				vm_files = append(vm_files, dirname+"/"+fi.Name())
			}
		}
	}

	if len(vm_files) == 0 {
		fmt.Println("main: No .vm file is given.")
		return
	}

	w, err := os.Create(asm_file)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer w.Close()

	cwr := NewMyVMCodeWriter(w)
	defer cwr.Close()

	err = cwr.WriteInit()
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, vm_file := range vm_files {
		r, err := os.Open(vm_file)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer r.Close()

		basename := vm_file
		ext := strings.Index(basename, ".vm")
		dirs := strings.LastIndex(basename, "/")
		if dirs != -1 {
			basename = basename[dirs+1 : ext]
		}
		cwr.SetFileName(basename)
		psr := NewMyVMParser(r)
		for psr.Advance() {
			command_type, err := psr.CommandType()
			if err != nil {
				fmt.Println(err)
				return
			}
			err = writeCommand(psr, cwr, command_type)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

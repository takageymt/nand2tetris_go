package main

import (
	"fmt"
	"os"
	"strings"
)

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
		ext := strings.Index(vm_file, ".vm")
		asm_file = vm_file[:ext] + ".asm"
	} else {
		dirname := strings.TrimRight(os.Args[1], "/")
		asm_file = dirname + ".asm"
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
				vm_files = append(vm_files, fi.Name())
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
			cmd_type, err := psr.CommandType()
			if err != nil {
				fmt.Println(err)
				return
			}
			if cmd_type == C_ARITHMETIC {
				cmd, err := psr.Arg1()
				if err != nil {
					fmt.Println(err)
					return
				}
				err = cwr.WriteArithmetic(cmd)
			} else if cmd_type == C_PUSH || cmd_type == C_POP {
				seg, err := psr.Arg1()
				if err != nil {
					fmt.Println(err)
					return
				}
				idx, err := psr.Arg2()
				if err != nil {
					fmt.Println(err)
					return
				}
				err = cwr.WritePushPop(cmd_type, seg, idx)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
		}
	}
}

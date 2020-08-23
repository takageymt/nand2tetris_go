package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

func TokenizerTest(tkn Tokenizer, w io.Writer) {
	bw := bufio.NewWriter(w)
	bw.WriteString("<tokens>\n")
	for tkn.Advance() {
		token_type := tkn.TokenType()
		var elem string
		switch token_type {
		case KEYWORD:
			elem = KeywordXML(tkn.Keyword())
		case SYMBOL:
			elem = SymbolXML(tkn.Symbol())
		case IDENTIFIER:
			elem = IdentifierXML(tkn.Identifier())
		case INT_CONST:
			elem = IntegerXML(tkn.IntVal())
		case STRING_CONST:
			elem = StringXML(tkn.StringVal())
		}
		bw.WriteString(elem + "\n")
		fmt.Println(elem)
	}
	bw.WriteString("</tokens>\n")
	bw.Flush()
}

func TokenPipe(tkn Tokenizer, ch chan<- Data) {
	for tkn.Advance() {
		token_type := tkn.TokenType()
		switch token_type {
		case KEYWORD:
			ch <- Data{Type: KEYWORD, Token: tkn.Keyword()}
		case SYMBOL:
			ch <- Data{Type: SYMBOL, Token: tkn.Symbol()}
		case IDENTIFIER:
			ch <- Data{Type: IDENTIFIER, Token: tkn.Identifier()}
		case INT_CONST:
			ch <- Data{Type: INT_CONST, Token: tkn.IntVal()}
		case STRING_CONST:
			ch <- Data{Type: STRING_CONST, Token: tkn.StringVal()}
		}
	}
	close(ch)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("main: No .jack file is given.")
		return
	}

	jack_files := make([]string, 0)

	if strings.HasSuffix(os.Args[1], ".jack") {
		jack_file := os.Args[1]
		jack_files = append(jack_files, jack_file)
	} else {
		dirname := strings.TrimRight(os.Args[1], "/")
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
			if !fi.IsDir() && strings.HasSuffix(fi.Name(), ".jack") {
				jack_files = append(jack_files, dirname+"/"+fi.Name())
			}
		}
	}

	if len(jack_files) == 0 {
		fmt.Println("main: No .jack file is given.")
		return
	}

	jce := NewJackCE()

	for _, jack_file := range jack_files {
		r, err := os.Open(jack_file)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer r.Close()

		tkn := NewJackTokenizer(r)
		dst := strings.TrimSuffix(jack_file, ".jack") + ".vm"
		w, err := os.Create(dst)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer w.Close()

		jw := NewJackVMWriter(w)
		st := NewSymbolTableList()
		ch := make(chan Data)

		wg := &sync.WaitGroup{}
		wg.Add(2)
		go func() {
			defer wg.Done()
			TokenPipe(tkn, ch)
		}()
		go func() {
			defer wg.Done()
			jce.Compile(jw, st, ch)
		}()
		wg.Wait()

		jw.Close()
		w.Close()
		r.Close()
	}
}

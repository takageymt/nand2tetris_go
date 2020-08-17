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
	src := os.Args[1]
	r, err := os.Open(src)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer r.Close()

	tkn := NewJackTokenizer(r)

	dst := strings.TrimSuffix(src, ".jack") + "Parse.xml"
	w, err := os.Create(dst)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer w.Close()

	jce := NewJackCE(w)

	//TokenizerTest(tkn, w)

	ch := make(chan Data)

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		TokenPipe(tkn, ch)
	}()
	go func() {
		defer wg.Done()
		jce.Compile(ch)
	}()
	wg.Wait()
}

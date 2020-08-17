package main

import (
	"bufio"
	"fmt"
	"io"
)

type CompilationEngine interface {
	CompileClass()
	CompileClassVarDec()
	CompileSubroutine()
	CompileParameterList()
	CompileVarDec()
	CompileStatements()
	CompileDo()
	CompileLet()
	CompileWhile()
	CompileReturn()
	CompileIf()
	CompileExpression()
	CompileTerm()
	CompileExpressionList()
}

type JackCE struct {
	ch   <-chan Data
	data Data
	bw   *bufio.Writer
}

func NewJackCE(w io.Writer) *JackCE {
	jce := new(JackCE)
	jce.bw = bufio.NewWriter(w)
	return jce
}

func (jce *JackCE) FlushAdvance() (ok bool) {
	var elem string
	switch jce.data.Type {
	case KEYWORD:
		elem = KeywordXML(jce.data.Token)
	case SYMBOL:
		elem = SymbolXML(jce.data.Token)
	case IDENTIFIER:
		elem = IdentifierXML(jce.data.Token)
	case INT_CONST:
		elem = IntegerXML(jce.data.Token)
	case STRING_CONST:
		elem = StringXML(jce.data.Token)
	}
	jce.bw.WriteString(elem + "\n")
	fmt.Println(elem)

	jce.data, ok = <-jce.ch
	return
}

func (jce *JackCE) Compile(ch <-chan Data) {
	jce.ch = ch
	jce.data = <-jce.ch
	if jce.data.Type != KEYWORD || jce.data.Token != "class" {
		return
	}
	jce.CompileClass()
	jce.bw.Flush()
}

func (jce *JackCE) CompileClass() {
	jce.bw.WriteString("<class>\n")
	if jce.data.Type != KEYWORD || jce.data.Token != "class" {
		return
	}
	jce.FlushAdvance() // class
	if jce.data.Type != IDENTIFIER {
		return
	}
	jce.FlushAdvance() // className
	if jce.data.Type != SYMBOL || jce.data.Token != "{" {
		return
	}
	jce.FlushAdvance() // {
	for jce.data.Type == KEYWORD && (jce.data.Token == "static" || jce.data.Token == "field") {
		jce.CompileClassVarDec()
	}
	for jce.data.Type == KEYWORD {
		switch jce.data.Token {
		case "constructor", "function", "method":
			jce.CompileSubroutine()
		default:
			return
		}
	}

	if jce.data.Type != SYMBOL || jce.data.Token != "}" {
		return
	}
	jce.FlushAdvance() // }
	jce.bw.WriteString("</class>\n")
}

func (jce *JackCE) CompileClassVarDec() {
	jce.bw.WriteString("<classVarDec>\n")
	if jce.data.Type != KEYWORD || (jce.data.Token != "static" && jce.data.Token != "field") {
		return
	}
	jce.FlushAdvance() // field|static
	if jce.data.Type != KEYWORD && jce.data.Type != IDENTIFIER {
		return
	}
	if jce.data.Type == KEYWORD && !jce.data.IsPrimitive() {
		return
	}
	jce.FlushAdvance() // type
	if jce.data.Type != IDENTIFIER {
		return
	}
	jce.FlushAdvance() // varName
	for jce.data.Type == SYMBOL && jce.data.Token == "," {
		jce.FlushAdvance() // ,
		if jce.data.Type != IDENTIFIER {
			return
		}
		jce.FlushAdvance() // varName
	}
	if jce.data.Type != SYMBOL || jce.data.Token != ";" {
		return
	}
	jce.FlushAdvance() // ;
	jce.bw.WriteString("</classVarDec>\n")
}

func (jce *JackCE) CompileSubroutine() {
	jce.bw.WriteString("<subroutineDec>\n")
	if jce.data.Type != KEYWORD || (jce.data.Token != "constructor" && jce.data.Token != "function" && jce.data.Token != "method") {
		return
	}
	jce.FlushAdvance() // constructor|function|method
	if jce.data.Type != KEYWORD && jce.data.Type != IDENTIFIER {
		return
	}
	if jce.data.Type == KEYWORD && !jce.data.IsPrimitive() && jce.data.Token != "void" {
		return
	}
	jce.FlushAdvance() // type
	if jce.data.Type != IDENTIFIER {
		return
	}
	jce.FlushAdvance() // subroutineName
	if jce.data.Type != SYMBOL || jce.data.Token != "(" {
		return
	}
	jce.FlushAdvance() // (
	jce.CompileParameterList()
	if jce.data.Type != SYMBOL || jce.data.Token != ")" {
		return
	}
	jce.FlushAdvance() // )

	jce.bw.WriteString("<subroutineBody>\n")
	if jce.data.Type != SYMBOL || jce.data.Token != "{" {
		return
	}
	jce.FlushAdvance() // {
	for jce.data.Type == KEYWORD && jce.data.Token == "var" {
		jce.CompileVarDec()
	}
	jce.CompileStatements()
	if jce.data.Type != SYMBOL || jce.data.Token != "}" {
		return
	}
	jce.FlushAdvance() // }
	jce.bw.WriteString("</subroutineBody>\n")
	jce.bw.WriteString("</subroutineDec>\n")
}

func (jce *JackCE) CompileParameterList() {
	jce.bw.WriteString("<parameterList>\n")
	if jce.data.Type != SYMBOL || jce.data.Token != ")" {
		if !jce.data.IsPrimitive() && jce.data.Type != IDENTIFIER {
			return
		}
		jce.FlushAdvance() // type
		if jce.data.Type != IDENTIFIER {
			return
		}
		jce.FlushAdvance() // varName
		for jce.data.Type != SYMBOL || jce.data.Token != ")" {
			if jce.data.Token != "," {
				return
			}
			jce.FlushAdvance() // ,
			if !jce.data.IsPrimitive() && jce.data.Type != IDENTIFIER {
				return
			}
			jce.FlushAdvance() // type
			if jce.data.Type != IDENTIFIER {
				return
			}
			jce.FlushAdvance() // varName
		}
	}
	jce.bw.WriteString("</parameterList>\n")
}

func (jce *JackCE) CompileVarDec() {
	jce.bw.WriteString("<varDec>\n")
	if jce.data.Type != SYMBOL || jce.data.Token != "var" {
		return
	}
	jce.FlushAdvance() // var
	if jce.data.Type != KEYWORD && jce.data.Type != IDENTIFIER {
		return
	}
	if jce.data.Type == KEYWORD && !jce.data.IsPrimitive() {
		return
	}
	jce.FlushAdvance() // type
	if jce.data.Type != IDENTIFIER {
		return
	}
	jce.FlushAdvance() // varName
	for jce.data.Type == SYMBOL && jce.data.Token == "," {
		jce.FlushAdvance() // ,
		if jce.data.Type != IDENTIFIER {
			return
		}
		jce.FlushAdvance() // varName
	}
	if jce.data.Type != SYMBOL || jce.data.Token != ";" {
		return
	}
	jce.FlushAdvance() // ;
	jce.bw.WriteString("</varDec>\n")
}

func (jce *JackCE) CompileStatements() {
	jce.bw.WriteString("<statements>\n")
	for jce.data.Type == KEYWORD {
		switch jce.data.Token {
		case "let":
			jce.CompileLet()
		case "if":
			jce.CompileIf()
		case "while":
			jce.CompileWhile()
		case "do":
			jce.CompileDo()
		case "return":
			jce.CompileReturn()
		default:
			return
		}
	}
	jce.bw.WriteString("</statements>\n")
}

func (jce *JackCE) CompileDo() {
	jce.bw.WriteString("<doStatement>\n")
	if jce.data.Type != KEYWORD || jce.data.Token != "do" {
		return
	}
	jce.FlushAdvance() // do
	jce.CompileSubroutineCall()
	if jce.data.Type != SYMBOL || jce.data.Token != ";" {
		return
	}
	jce.FlushAdvance() // ;

	jce.bw.WriteString("</doStatement>\n")
}

func (jce *JackCE) CompileSubroutineCall() {
	if jce.data.Type != IDENTIFIER {
		return
	}
	jce.FlushAdvance() // className/varName/subroutineName
	if jce.data.Type != SYMBOL {
		return
	}
	if jce.data.Token == "." {
		jce.FlushAdvance() // .
		jce.CompileSubroutineCall()
		return
	}
	if jce.data.Token != "(" {
		return
	}
	jce.FlushAdvance() // (
	jce.CompileExpressionList()
	if jce.data.Type != SYMBOL || jce.data.Token != ")" {
		return
	}
	jce.FlushAdvance() // )
}

func (jce *JackCE) CompileLet() {
	jce.bw.WriteString("<letStatement>\n")
	if jce.data.Type != KEYWORD || jce.data.Token != "let" {
		return
	}
	jce.FlushAdvance() // let
	if jce.data.Type != IDENTIFIER {
		return
	}
	jce.FlushAdvance() // varName
	if jce.data.Type == SYMBOL && jce.data.Token == "[" {
		jce.FlushAdvance() // [
		jce.CompileExpression()
		if jce.data.Type != SYMBOL || jce.data.Token != "]" {
			return
		}
		jce.FlushAdvance() // ]
	}
	if jce.data.Type != SYMBOL || jce.data.Token != "=" {
		return
	}
	jce.FlushAdvance() // =
	jce.CompileExpression()
	if jce.data.Type != SYMBOL || jce.data.Token != ";" {
		return
	}
	jce.FlushAdvance() // ;
	jce.bw.WriteString("</letStatement>\n")
}

func (jce *JackCE) CompileWhile() {
	jce.bw.WriteString("<whileStatement>\n")
	if jce.data.Type != KEYWORD || jce.data.Token != "while" {
		return
	}
	jce.FlushAdvance() // while
	if jce.data.Type != SYMBOL || jce.data.Token != "(" {
		return
	}
	jce.FlushAdvance() // (
	jce.CompileExpression()
	if jce.data.Type != SYMBOL || jce.data.Token != ")" {
		return
	}
	jce.FlushAdvance() // )
	jce.CompileCodeBlock()
	jce.bw.WriteString("</whileStatement>\n")
}

func (jce *JackCE) CompileReturn() {
	jce.bw.WriteString("<returnStatement>\n")
	if jce.data.Type != KEYWORD || jce.data.Token != "return" {
		return
	}
	jce.FlushAdvance() // return
	if jce.data.Type != SYMBOL || jce.data.Token != ";" {
		jce.CompileExpression()
	}
	if jce.data.Type != SYMBOL || jce.data.Token != ";" {
		return
	}
	jce.FlushAdvance() // ;
	jce.bw.WriteString("</returnStatement>\n")
}

func (jce *JackCE) CompileCodeBlock() {
	fmt.Println(jce.data.Type, jce.data.Token)
	if jce.data.Type != SYMBOL || jce.data.Token != "{" {
		return
	}
	jce.FlushAdvance() // {
	jce.CompileStatements()
	if jce.data.Type != SYMBOL || jce.data.Token != "}" {
		return
	}
	jce.FlushAdvance() // }
}

func (jce *JackCE) CompileIf() {
	jce.bw.WriteString("<ifStatement>\n")
	if jce.data.Type != KEYWORD || jce.data.Token != "if" {
		return
	}
	jce.FlushAdvance() // if
	if jce.data.Type != SYMBOL || jce.data.Token != "(" {
		return
	}
	jce.FlushAdvance() // (
	jce.CompileExpression()
	if jce.data.Type != SYMBOL || jce.data.Token != ")" {
		return
	}
	jce.FlushAdvance() // )
	jce.CompileCodeBlock()
	if jce.data.Type == KEYWORD && jce.data.Token == "else" {
		jce.FlushAdvance() // else
		jce.CompileCodeBlock()
	}
	jce.bw.WriteString("</ifStatement>\n")
}

func (jce *JackCE) CompileExpression() {
	jce.bw.WriteString("<expression>\n")
	jce.CompileTerm()
	jce.bw.WriteString("</expression>\n")
}

func (jce *JackCE) CompileExpressionList() {
	jce.bw.WriteString("<expressionList>\n")
	if jce.data.Type != SYMBOL || jce.data.Token != ")" {
		jce.CompileExpression()
		for jce.data.Type == SYMBOL && jce.data.Token == "," {
			jce.FlushAdvance() // ,
			jce.CompileExpression()
		}
	}
	jce.bw.WriteString("</expressionList>\n")
}

func (jce *JackCE) CompileTerm() {
	jce.bw.WriteString("<term>\n")
	if jce.data.IsConstant() {
		jce.FlushAdvance()
	} else if jce.data.Type == SYMBOL {
		if jce.data.Token == "(" {
			jce.FlushAdvance()
			jce.CompileExpression()
			if jce.data.Type != SYMBOL || jce.data.Token != ")" {
				fmt.Println("ERROR1")
				return
			}
			jce.FlushAdvance()
		} else if jce.data.Token == "-" || jce.data.Token == "~" {
			jce.FlushAdvance()
			jce.CompileTerm()
		} else {
			fmt.Println(jce.data.Token)
			fmt.Println("ERROR2")
			return
		}
	} else if jce.data.Type == IDENTIFIER {
		jce.FlushAdvance()
		if jce.data.Type == SYMBOL {
			if jce.data.Token == "." {
				jce.FlushAdvance()
				jce.CompileSubroutineCall()
			} else if jce.data.Token == "(" {
				jce.FlushAdvance()
				jce.CompileExpressionList()
				if jce.data.Type != SYMBOL || jce.data.Token != ")" {
					fmt.Println("ERROR3")
					return
				}
				jce.FlushAdvance()
			} else if jce.data.Token == "[" {
				jce.FlushAdvance()
				jce.CompileExpression()
				if jce.data.Type != SYMBOL || jce.data.Token != "]" {
					fmt.Println("ERROR4")
					return
				}
			}
		}
	} else {
		fmt.Println("ERROR5")
		return
	}
	jce.bw.WriteString("</term>\n")
}

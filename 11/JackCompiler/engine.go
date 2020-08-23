package main

import (
	"fmt"
	"strconv"
)

type CompilationEngine interface {
	CompileClass()
	CompileClassVarDec()
	CompileSubroutine()
	CompileParameterList() (nArgs int)
	CompileVarDec() (n int)
	CompileStatements()
	CompileDo()
	CompileLet()
	CompileWhile()
	CompileReturn()
	CompileIf()
	CompileExpression()
	CompileTerm()
	CompileExpressionList() (nExp int)
}

type JackCE struct {
	ch   <-chan Data
	data Data
	// bw   *bufio.Writer
	jw      VMWriter
	st      SymbolTable
	class   string
	nFields int
}

func NewJackCE() *JackCE {
	jce := new(JackCE)
	// jce.bw = bufio.NewWriter(w)
	jce.jw = nil
	jce.st = nil
	return jce
}

/*
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
*/
func (jce *JackCE) TokenAdvance() (token string) {
	token = jce.data.Token
	jce.data = <-jce.ch
	return
}

func (jce *JackCE) Compile(jw VMWriter, st SymbolTable, ch <-chan Data) {
	jce.jw = jw
	jce.st = st
	jce.ch = ch
	jce.data = <-jce.ch
	if jce.data.Type != KEYWORD || jce.data.Token != "class" {
		return
	}
	jce.CompileClass()
	//jce.bw.Flush()
}

func (jce *JackCE) CompileClass() {
	//jce.bw.WriteString("<class>\n")
	if jce.data.Type != KEYWORD || jce.data.Token != "class" {
		return
	}
	//jce.FlushAdvance() // class
	jce.TokenAdvance()
	if jce.data.Type != IDENTIFIER {
		return
	}
	//jce.FlushAdvance() // className
	jce.class = jce.TokenAdvance()
	jce.nFields = 0
	jce.jw.SetClass(jce.class)
	if jce.data.Type != SYMBOL || jce.data.Token != "{" {
		return
	}
	//jce.FlushAdvance() // {
	jce.TokenAdvance()
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
	//jce.FlushAdvance() // }
	jce.TokenAdvance()
	//jce.bw.WriteString("</class>\n")
}

func (jce *JackCE) CompileClassVarDec() {
	//jce.bw.WriteString("<classVarDec>\n")
	if jce.data.Type != KEYWORD || (jce.data.Token != "static" && jce.data.Token != "field") {
		return
	}
	//jce.FlushAdvance() // field|static
	var kind int
	if jce.data.Token == "static" {
		kind = STATIC
	} else {
		kind = FIELD
		jce.nFields++
	}
	jce.TokenAdvance()
	if jce.data.Type != KEYWORD && jce.data.Type != IDENTIFIER {
		return
	}
	if jce.data.Type == KEYWORD && !jce.data.IsPrimitive() {
		return
	}
	//jce.FlushAdvance() // type
	atype := jce.TokenAdvance()
	if jce.data.Type != IDENTIFIER {
		return
	}
	//jce.FlushAdvance() // varName
	varName := jce.TokenAdvance()
	jce.st.Define(varName, atype, kind)
	for jce.data.Type == SYMBOL && jce.data.Token == "," {
		//jce.FlushAdvance() // ,
		jce.TokenAdvance()
		if jce.data.Type != IDENTIFIER {
			return
		}
		//jce.FlushAdvance() // varName
		varName = jce.TokenAdvance()
		jce.st.Define(varName, atype, kind)
		if kind == FIELD {
			jce.nFields++
		}
	}
	if jce.data.Type != SYMBOL || jce.data.Token != ";" {
		return
	}
	//jce.FlushAdvance() // ;
	jce.TokenAdvance()
	//jce.bw.WriteString("</classVarDec>\n")
	return
}

func (jce *JackCE) CompileSubroutine() {
	//jce.bw.WriteString("<subroutineDec>\n")
	if jce.data.Type != KEYWORD || (jce.data.Token != "constructor" && jce.data.Token != "function" && jce.data.Token != "method") {
		return
	}
	var routineType int
	switch jce.data.Token {
	case "constructor":
		routineType = CONSTRUCTOR
	case "function":
		routineType = FUNCTION
	case "method":
		routineType = METHOD
	}
	jce.st.EnterSubroutine(routineType)
	//jce.FlushAdvance() // constructor|function|method
	jce.TokenAdvance()
	if jce.data.Type != KEYWORD && jce.data.Type != IDENTIFIER {
		return
	}
	if jce.data.Type == KEYWORD && !jce.data.IsPrimitive() && jce.data.Token != "void" {
		return
	}
	//jce.FlushAdvance() // type
	jce.TokenAdvance()
	if jce.data.Type != IDENTIFIER {
		return
	}
	//jce.FlushAdvance() // subroutineName
	fname := jce.TokenAdvance()
	if jce.data.Type != SYMBOL || jce.data.Token != "(" {
		return
	}
	//jce.FlushAdvance() // (
	jce.TokenAdvance()
	jce.CompileParameterList()
	if jce.data.Type != SYMBOL || jce.data.Token != ")" {
		return
	}
	//jce.FlushAdvance() // )
	jce.TokenAdvance()

	//jce.bw.WriteString("<subroutineBody>\n")
	if jce.data.Type != SYMBOL || jce.data.Token != "{" {
		return
	}
	//jce.FlushAdvance() // {
	jce.TokenAdvance()
	nLocals := 0
	for jce.data.Type == KEYWORD && jce.data.Token == "var" {
		nLocals += jce.CompileVarDec()
	}
	jce.jw.WriteFunction(fname, nLocals)

	if routineType == CONSTRUCTOR {
		jce.jw.WritePush(SEG_CONST, jce.nFields)
		jce.jw.WriteCall("Memory.alloc", 1)
		jce.jw.WritePop(SEG_POINTER, 0)
	} else if routineType == METHOD {
		jce.jw.WritePush(SEG_ARG, 0)
		jce.jw.WritePop(SEG_POINTER, 0)
	}

	jce.CompileStatements()
	if jce.data.Type != SYMBOL || jce.data.Token != "}" {
		return
	}
	//jce.FlushAdvance() // }
	jce.TokenAdvance()
	jce.st.ExitSubroutine()
	//jce.bw.WriteString("</subroutineBody>\n")
	//jce.bw.WriteString("</subroutineDec>\n")
}

func (jce *JackCE) CompileParameterList() (nArgs int) {
	//jce.bw.WriteString("<parameterList>\n")
	nArgs = 0
	if jce.data.Type != SYMBOL || jce.data.Token != ")" {
		if !jce.data.IsPrimitive() && jce.data.Type != IDENTIFIER {
			return
		}
		//jce.FlushAdvance() // type
		atype := jce.TokenAdvance()
		if jce.data.Type != IDENTIFIER {
			return
		}
		//jce.FlushAdvance() // varName
		varName := jce.TokenAdvance()
		jce.st.Define(varName, atype, ARG)
		nArgs++
		for jce.data.Type != SYMBOL || jce.data.Token != ")" {
			if jce.data.Token != "," {
				return
			}
			//jce.FlushAdvance() // ,
			jce.TokenAdvance()
			if !jce.data.IsPrimitive() && jce.data.Type != IDENTIFIER {
				return
			}
			//jce.FlushAdvance() // type
			atype = jce.TokenAdvance()
			if jce.data.Type != IDENTIFIER {
				return
			}
			//jce.FlushAdvance() // varName
			varName = jce.TokenAdvance()
			jce.st.Define(varName, atype, ARG)
			nArgs++
		}
	}
	//jce.bw.WriteString("</parameterList>\n")
	return
}

func (jce *JackCE) CompileVarDec() (n int) {
	//jce.bw.WriteString("<varDec>\n")
	n = 0
	if jce.data.Type != KEYWORD || jce.data.Token != "var" {
		return
	}
	//jce.FlushAdvance() // var
	jce.TokenAdvance()
	if jce.data.Type != KEYWORD && jce.data.Type != IDENTIFIER {
		return
	}
	if jce.data.Type == KEYWORD && !jce.data.IsPrimitive() {
		return
	}
	//jce.FlushAdvance() // type
	atype := jce.TokenAdvance()
	if jce.data.Type != IDENTIFIER {
		return
	}
	//jce.FlushAdvance() // varName
	varName := jce.TokenAdvance()
	jce.st.Define(varName, atype, VAR)
	n++
	for jce.data.Type == SYMBOL && jce.data.Token == "," {
		//jce.FlushAdvance() // ,
		jce.TokenAdvance()
		if jce.data.Type != IDENTIFIER {
			return
		}
		//jce.FlushAdvance() // varName
		varName = jce.TokenAdvance()
		jce.st.Define(varName, atype, VAR)
		n++
	}
	if jce.data.Type != SYMBOL || jce.data.Token != ";" {
		return
	}
	//jce.FlushAdvance() // ;
	jce.TokenAdvance()
	//jce.bw.WriteString("</varDec>\n")
	return
}

func (jce *JackCE) CompileStatements() {
	//jce.bw.WriteString("<statements>\n")
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
	//jce.bw.WriteString("</statements>\n")
}

func (jce *JackCE) CompileDo() {
	//jce.bw.WriteString("<doStatement>\n")
	if jce.data.Type != KEYWORD || jce.data.Token != "do" {
		return
	}
	//jce.FlushAdvance() // do
	jce.TokenAdvance()

	nArgs := 0
	if jce.data.Type != IDENTIFIER {
		return
	}
	//jce.FlushAdvance() // className/varName/subroutineName
	maybeObj := jce.TokenAdvance()
	fname := maybeObj
	if jce.data.Type != SYMBOL {
		return
	}
	if k := jce.st.KindOf(maybeObj); k != NONE {
		switch k {
		case STATIC:
			jce.jw.WritePush(SEG_STATIC, jce.st.IndexOf(maybeObj))
		case FIELD:
			jce.jw.WritePush(SEG_THIS, jce.st.IndexOf(maybeObj))
		case VAR:
			jce.jw.WritePush(SEG_LOCAL, jce.st.IndexOf(maybeObj))
		case ARG:
			jce.jw.WritePush(SEG_ARG, jce.st.IndexOf(maybeObj))
		default:
			return
		}
		nArgs++
		fname = jce.st.TypeOf(maybeObj)
	}
	if jce.data.Token == "." {
		//jce.FlushAdvance() // .
		jce.TokenAdvance()
		if jce.data.Type != IDENTIFIER {
			return
		}
		fname += "." + jce.TokenAdvance()
	} else {
		jce.jw.WritePush(SEG_POINTER, 0)
		nArgs++
		fname = jce.class + "." + fname
	}
	if jce.data.Token != "(" {
		return
	}
	//jce.FlushAdvance() // (
	jce.TokenAdvance()
	nArgs += jce.CompileExpressionList()
	if jce.data.Type != SYMBOL || jce.data.Token != ")" {
		return
	}
	//jce.FlushAdvance() // )
	jce.TokenAdvance()

	jce.jw.WriteCall(fname, nArgs)
	if jce.data.Type != SYMBOL || jce.data.Token != ";" {
		return
	}
	jce.jw.WritePop(SEG_TEMP, 0)
	//jce.FlushAdvance() // ;
	jce.TokenAdvance()
	//jce.bw.WriteString("</doStatement>\n")
}

func (jce *JackCE) CompileLet() {
	//jce.bw.WriteString("<letStatement>\n")
	if jce.data.Type != KEYWORD || jce.data.Token != "let" {
		return
	}
	//jce.FlushAdvance() // let
	jce.TokenAdvance()
	if jce.data.Type != IDENTIFIER {
		return
	}
	//jce.FlushAdvance() // varName
	varName := jce.TokenAdvance()
	kind := jce.st.KindOf(varName)
	var segment int
	switch kind {
	case STATIC:
		segment = SEG_STATIC
	case FIELD:
		segment = SEG_THIS
	case ARG:
		segment = SEG_ARG
	case VAR:
		segment = SEG_LOCAL
	default:
		return
	}
	index := jce.st.IndexOf(varName)

	isArray := false
	if jce.data.Type == SYMBOL && jce.data.Token == "[" {
		isArray = true
		//jce.FlushAdvance() // [
		jce.TokenAdvance()
		jce.jw.WritePush(segment, index)
		jce.CompileExpression()
		if jce.data.Type != SYMBOL || jce.data.Token != "]" {
			return
		}
		jce.jw.WriteArithmetic(COM_ADD)
		//jce.jw.WritePop(SEG_POINTER, 1)
		//jce.FlushAdvance() // ]
		jce.TokenAdvance()
	}
	if jce.data.Type != SYMBOL || jce.data.Token != "=" {
		return
	}
	//jce.FlushAdvance() // =
	jce.TokenAdvance()
	jce.CompileExpression()
	if jce.data.Type != SYMBOL || jce.data.Token != ";" {
		return
	}
	//jce.FlushAdvance() // ;
	jce.TokenAdvance()
	if isArray {
		jce.jw.WritePop(SEG_TEMP, 0)
		jce.jw.WritePop(SEG_POINTER, 1)
		jce.jw.WritePush(SEG_TEMP, 0)
		jce.jw.WritePop(SEG_THAT, 0)
	} else {
		jce.jw.WritePop(segment, index)
	}
	//jce.bw.WriteString("</letStatement>\n")
}

func (jce *JackCE) CompileWhile() {
	//jce.bw.WriteString("<whileStatement>\n")
	if jce.data.Type != KEYWORD || jce.data.Token != "while" {
		return
	}
	//jce.FlushAdvance() // while
	jce.TokenAdvance()
	if jce.data.Type != SYMBOL || jce.data.Token != "(" {
		return
	}
	//jce.FlushAdvance() // (
	jce.TokenAdvance()
	startLabel := jce.jw.NewLabel()
	endLabel := jce.jw.NewLabel()
	jce.jw.WriteLabel(startLabel)
	jce.CompileExpression()
	if jce.data.Type != SYMBOL || jce.data.Token != ")" {
		return
	}
	//jce.FlushAdvance() // )
	jce.TokenAdvance()
	jce.jw.WriteArithmetic(COM_NOT)
	jce.jw.WriteIf(endLabel)
	jce.CompileCodeBlock()
	jce.jw.WriteGoto(startLabel)
	jce.jw.WriteLabel(endLabel)
	//jce.bw.WriteString("</whileStatement>\n")
}

func (jce *JackCE) CompileReturn() {
	//jce.bw.WriteString("<returnStatement>\n")
	if jce.data.Type != KEYWORD || jce.data.Token != "return" {
		return
	}
	//jce.FlushAdvance() // return
	jce.TokenAdvance()
	if jce.data.Type != SYMBOL || jce.data.Token != ";" {
		jce.CompileExpression()
	} else {
		jce.jw.WritePush(SEG_CONST, 0)
	}
	if jce.data.Type != SYMBOL || jce.data.Token != ";" {
		return
	}
	//jce.FlushAdvance() // ;
	jce.TokenAdvance()
	jce.jw.WriteReturn()
	//jce.bw.WriteString("</returnStatement>\n")
}

func (jce *JackCE) CompileCodeBlock() {
	if jce.data.Type != SYMBOL || jce.data.Token != "{" {
		return
	}
	//jce.FlushAdvance() // {
	jce.TokenAdvance()
	jce.CompileStatements()
	if jce.data.Type != SYMBOL || jce.data.Token != "}" {
		return
	}
	//jce.FlushAdvance() // }
	jce.TokenAdvance()
}

func (jce *JackCE) CompileIf() {
	//jce.bw.WriteString("<ifStatement>\n")
	if jce.data.Type != KEYWORD || jce.data.Token != "if" {
		return
	}
	//jce.FlushAdvance() // if
	jce.TokenAdvance()
	if jce.data.Type != SYMBOL || jce.data.Token != "(" {
		return
	}
	//jce.FlushAdvance() // (
	jce.TokenAdvance()
	jce.CompileExpression()
	if jce.data.Type != SYMBOL || jce.data.Token != ")" {
		return
	}
	//jce.FlushAdvance() // )
	jce.TokenAdvance()
	falseLabel := jce.jw.NewLabel()
	jce.jw.WriteArithmetic(COM_NOT)
	jce.jw.WriteIf(falseLabel)
	jce.CompileCodeBlock()
	if jce.data.Type == KEYWORD && jce.data.Token == "else" {
		endLabel := jce.jw.NewLabel()
		jce.jw.WriteGoto(endLabel)
		//jce.FlushAdvance() // else
		jce.TokenAdvance()
		jce.jw.WriteLabel(falseLabel)
		jce.CompileCodeBlock()
		jce.jw.WriteLabel(endLabel)
	} else {
		jce.jw.WriteLabel(falseLabel)
	}
	//jce.bw.WriteString("</ifStatement>\n")
}

func (jce *JackCE) CompileExpression() {
	//jce.bw.WriteString("<expression>\n")
	jce.CompileTerm()
	for jce.data.IsBinOp() {
		//jce.FlushAdvance() // op
		op := jce.TokenAdvance()
		jce.CompileTerm()
		switch op {
		case "+":
			jce.jw.WriteArithmetic(COM_ADD)
		case "-":
			jce.jw.WriteArithmetic(COM_SUB)
		case "*":
			jce.jw.WriteCall("Math.multiply", 2)
		case "/":
			jce.jw.WriteCall("Math.divide", 2)
		case "&":
			jce.jw.WriteArithmetic(COM_AND)
		case "|":
			jce.jw.WriteArithmetic(COM_OR)
		case "<":
			jce.jw.WriteArithmetic(COM_LT)
		case ">":
			jce.jw.WriteArithmetic(COM_GT)
		case "=":
			jce.jw.WriteArithmetic(COM_EQ)
		}
	}
	//jce.bw.WriteString("</expression>\n")
}

func (jce *JackCE) CompileExpressionList() (nExp int) {
	//jce.bw.WriteString("<expressionList>\n")
	nExp = 0
	if jce.data.Type != SYMBOL || jce.data.Token != ")" {
		jce.CompileExpression()
		nExp++
		for jce.data.Type == SYMBOL && jce.data.Token == "," {
			//jce.FlushAdvance() // ,
			jce.TokenAdvance()
			jce.CompileExpression()
			nExp++
		}
	}
	//jce.bw.WriteString("</expressionList>\n")
	return
}

func (jce *JackCE) CompileTerm() {
	//jce.bw.WriteString("<term>\n")
	if jce.data.IsConstant() {
		if jce.data.Type == INT_CONST {
			constVal, _ := strconv.Atoi(jce.data.Token)
			jce.jw.WritePush(SEG_CONST, constVal)
		} else if jce.data.Type == STRING_CONST {
			jce.jw.WritePush(SEG_CONST, len(jce.data.Token))
			jce.jw.WriteCall("String.new", 1)

			for i := 0; i < len(jce.data.Token); i++ {
				jce.jw.WritePush(SEG_CONST, int(jce.data.Token[i]))
				jce.jw.WriteCall("String.appendChar", 2)
			}
		} else if jce.data.Type == KEYWORD {
			switch jce.data.Token {
			case "this":
				jce.jw.WritePush(SEG_POINTER, 0)
			case "true":
				jce.jw.WritePush(SEG_CONST, 0)
				jce.jw.WriteArithmetic(COM_NOT)
			case "false", "null":
				jce.jw.WritePush(SEG_CONST, 0)
			default:
				return
			}
		} else {
			return
		}
		//jce.FlushAdvance()
		jce.TokenAdvance()
	} else if jce.data.Type == SYMBOL {
		if jce.data.Token == "(" {
			//jce.FlushAdvance() // (
			jce.TokenAdvance()
			jce.CompileExpression()
			if jce.data.Type != SYMBOL || jce.data.Token != ")" {
				fmt.Println("ERROR1")
				return
			}
			//jce.FlushAdvance() // )
			jce.TokenAdvance()
		} else if jce.data.Token == "-" {
			//jce.FlushAdvance() -
			jce.TokenAdvance()
			jce.CompileTerm()
			jce.jw.WriteArithmetic(COM_NEG)
		} else if jce.data.Token == "~" {
			//jce.FlushAdvance() // ~
			jce.TokenAdvance()
			jce.CompileTerm()
			jce.jw.WriteArithmetic(COM_NOT)
		} else {
			fmt.Println(jce.data.Token)
			fmt.Println("ERROR2")
			return
		}
	} else if jce.data.Type == IDENTIFIER {
		//jce.FlushAdvance()
		id := jce.TokenAdvance()
		kind := jce.st.KindOf(id)
		var segment, index int
		if kind != NONE {
			switch jce.st.KindOf(id) {
			case STATIC:
				segment = SEG_STATIC
			case FIELD:
				segment = SEG_THIS
			case VAR:
				segment = SEG_LOCAL
			case ARG:
				segment = SEG_ARG
			default:
				return
			}
			index = jce.st.IndexOf(id)
		}

		if jce.data.Type == SYMBOL {
			if jce.data.Token == "." {
				// id is expected to be an object or a class
				//jce.FlushAdvance() // .
				jce.TokenAdvance()
				nArgs := 0
				if kind != NONE {
					// id denotes an object
					jce.jw.WritePush(segment, index)
					nArgs++
					id = jce.st.TypeOf(id)
				} // else id is an class
				if jce.data.Type != IDENTIFIER {
					return
				}
				fname := jce.TokenAdvance()
				if jce.data.Token != "(" {
					return
				}
				jce.TokenAdvance() // (
				nArgs += jce.CompileExpressionList()
				if jce.data.Type != SYMBOL || jce.data.Token != ")" {
					return
				}
				//jce.FlushAdvance() // )
				jce.TokenAdvance()
				jce.jw.WriteCall(id+"."+fname, nArgs)
			} else if jce.data.Token == "(" {
				//jce.FlushAdvance() // (
				jce.TokenAdvance()
				jce.jw.WritePush(SEG_POINTER, 0)
				nArgs := jce.CompileExpressionList()
				if jce.data.Type != SYMBOL || jce.data.Token != ")" {
					fmt.Println("ERROR3")
					return
				}
				//jce.FlushAdvance() // )
				jce.TokenAdvance()
				jce.jw.WriteCall(jce.class+"."+id, nArgs+1)
			} else if jce.data.Token == "[" {
				//jce.FlushAdvance() // [
				jce.TokenAdvance()
				jce.jw.WritePush(segment, index)
				jce.CompileExpression()
				if jce.data.Type != SYMBOL || jce.data.Token != "]" {
					fmt.Println("ERROR4")
					return
				}
				jce.jw.WriteArithmetic(COM_ADD)
				jce.jw.WritePop(SEG_POINTER, 1)
				jce.jw.WritePush(SEG_THAT, 0)
				//jce.FlushAdvance() // ]
				jce.TokenAdvance()
			} else {
				if kind == NONE {
					return
				}
				jce.jw.WritePush(segment, index)
			}
		}
	} else {
		fmt.Println("ERROR5")
		return
	}
	//jce.bw.WriteString("</term>\n")
}

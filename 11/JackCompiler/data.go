package main

const (
	KEYWORD = iota
	SYMBOL
	IDENTIFIER
	INT_CONST
	STRING_CONST
	INVALID_TOKEN
)

type Data struct {
	Type  int
	Token string
}

func (data Data) IsPrimitive() bool {
	return data.Type == KEYWORD && (data.Token == "int" || data.Token == "char" || data.Token == "boolean")
}

func (data Data) IsConstant() bool {
	if data.Type == INT_CONST || data.Type == STRING_CONST {
		return true
	}
	switch data.Token {
	case "true", "false", "this", "null":
		return data.Type == KEYWORD
	}
	return false
}

func (data Data) IsBinOp() bool {
	switch data.Token {
	case "+", "-", "*", "/", "&", "|", "<", ">", "=":
		return data.Type == SYMBOL
	}
	return false
}

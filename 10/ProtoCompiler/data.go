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
	if data.Type == KEYWORD {
		switch data.Token {
		case "true", "false", "this", "null":
			return true
		}
	}
	return false
}

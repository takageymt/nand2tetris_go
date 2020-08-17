package main

func KeywordXML(keyword string) string {
	return "<keyword>" + keyword + "</keyword>"
}

func SymbolXML(symbol string) string {
	switch symbol {
	case "&":
		return "<symbol>&amp;</symbol>"
	case "<":
		return "<symbol>&lt;</symbol>"
	case ">":
		return "<symbol>&gt;</symbol>"
	default:
		return "<symbol>" + symbol + "</symbol>"
	}
}

func IdentifierXML(id string) string {
	return "<identifier>" + id + "</identifier>"
}

func IntegerXML(val string) string {
	return "<integerConstant>" + val + "</integerConstant>"
}

func StringXML(str string) string {
	return "<stringConstant>" + str + "</stringConstant>"
}

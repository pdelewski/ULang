package main

import (
	"ULang/lexer"
	uqllexer "ULang/uql/lexer"
	uqlparser "ULang/uql/parser"
	"fmt"
)

func main() {
	tokens := lexer.GetTokens(uqllexer.StringToToken("t1.field1 > 10 && t1.field2 < 20"))
	expr, _ := uqlparser.ParseExpression(tokens)
	fmt.Println("Parsed Expression Tree:")
	fmt.Println(uqlparser.PrintLogicalExprString(expr))

	uqlparser.Parse(`
 t1 = from table1
 t2 = where t1.field1 > 10 && t1.field2 < 20
 t3 = select t2.field1
`)

}

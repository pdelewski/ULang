package main

import (
	"ULang/lexer"
	"ULang/uql/ast"
	uqllexer "ULang/uql/lexer"
	uqlparser "ULang/uql/parser"
	"fmt"
	"strings"
)

type State struct {
	depth int
}

func preVisit(state any, expr any) any {
	newState := state.(State)
	newState.depth++
	var builder strings.Builder
	indent := strings.Repeat("  ", newState.depth)
	builder.WriteString(fmt.Sprintf("%sValue:", indent))
	builder.WriteString(lexer.DumpTokensString([]lexer.Token{expr.(ast.LogicalExpr).Value}))
	fmt.Print(builder.String())
	return newState
}

func postVisit(state any, expr any) any {
	newState := state.(State)
	newState.depth--
	return newState
}

func main() {
	tokens := lexer.GetTokens(uqllexer.StringToToken("t1.field1 > 10 && t1.field2 < 20;"))
	expr, _ := uqlparser.ParseExpression(tokens)
	fmt.Println("Parsed Expression Tree:")
	state := State{depth: 0}
	ast.WalkLogicalExpr(expr, state, preVisit, postVisit)

	astTree, err := uqlparser.Parse(`
 t1 = from table1;
 t2 = where t1.field1 > 10 && t1.field2 < 20;
 t3 = select t2.field1;
`)
	if err != 0 {
		fmt.Println("Error parsing query")
	}
	for _, statement := range astTree {
		fmt.Println(statement.Type)
		switch statement.Type {
		case ast.StatementTypeFrom:
			fmt.Println("From:")
			lexer.DumpToken(statement.From.ResultTableExpr)
		case ast.StatementTypeWhere:
			fmt.Println("Where:")
			lexer.DumpToken(statement.Where.ResultTableExpr)
		case ast.StatementTypeSelect:
			fmt.Println("Select:")
			lexer.DumpToken(statement.Select.ResultTableExpr)
		}

	}
}

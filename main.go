package main

import (
	"ULang/lexer"
	"ULang/uql/ast"
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

	visitor := ast.Visitor{
		PreVisitFrom: func(state any, expr ast.From) any {
			newState := state.(State)
			return newState
		},
		PostVisitFrom: func(state any, expr ast.From) any {
			newState := state.(State)
			return newState
		},
		PreVisitWhere: func(state any, expr ast.Where) any {
			newState := state.(State)
			return newState
		},
		PostVisitWhere: func(state any, expr ast.Where) any {
			newState := state.(State)
			return newState
		},
		PreVisitSelect: func(state any, expr ast.Select) any {
			newState := state.(State)
			return newState
		},
		PostVisitSelect: func(state any, expr ast.Select) any {
			newState := state.(State)
			return newState
		},
		PreVisitLogicalExpr: func(state any, expr ast.LogicalExpr) any {
			newState := state.(State)
			newState.depth++
			var builder strings.Builder
			indent := strings.Repeat("  ", newState.depth)
			builder.WriteString(fmt.Sprintf("%sValue:", indent))
			builder.WriteString(lexer.DumpTokensString([]lexer.Token{expr.Value}))
			fmt.Print(builder.String())
			return newState
		},
		PostVisitLogicalExpr: func(state any, expr ast.LogicalExpr) any {
			newState := state.(State)
			newState.depth--
			return newState
		},
	}
	_ = visitor
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
			ast.WalkWhere(statement.Where, State{depth: 0}, visitor)
		case ast.StatementTypeSelect:
			fmt.Println("Select:")
			lexer.DumpToken(statement.Select.ResultTableExpr)
		}
	}
}

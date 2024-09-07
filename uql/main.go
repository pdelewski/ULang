package main

import (
	"fmt"
	"strings"
	"uql/ast"
	"uql/lexer"
	"uql/parser"
)

type State struct {
	depth int
}

func main() {
	visitor := ast.Visitor{
		PreVisitFrom: func(state any, expr ast.From) any {
			newState := state.(State)
			return newState
		},
		PostVisitFrom: func(state any, from ast.From) any {
			newState := state.(State)
			fmt.Println("From:")
			lexer.DumpToken(from.ResultTableExpr)
			lexer.DumpToken(from.TableExpr[0])
			return newState
		},
		PreVisitWhere: func(state any, where ast.Where) any {
			newState := state.(State)
			fmt.Println("Where:")
			lexer.DumpToken(where.ResultTableExpr)
			return newState
		},
		PostVisitWhere: func(state any, expr ast.Where) any {
			newState := state.(State)
			return newState
		},
		PreVisitSelect: func(state any, project ast.Select) any {
			newState := state.(State)
			fmt.Println("Select:")
			lexer.DumpToken(project.ResultTableExpr)
			lexer.DumpToken(project.Fields[0])
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

	astTree, err := parser.Parse(`
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
			ast.WalkFrom(statement.From, State{depth: 0}, visitor)
		case ast.StatementTypeWhere:
			ast.WalkWhere(statement.Where, State{depth: 0}, visitor)
		case ast.StatementTypeSelect:
			ast.WalkSelect(statement.Select, State{depth: 0}, visitor)
		}
	}
}

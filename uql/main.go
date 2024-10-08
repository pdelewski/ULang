package main

import (
	"fmt"
	"strings"
	"uql/ast"
	"uql/lexer"
	"uql/parser"
)

type State struct {
	depth int8
}

func main() {
	visitor := ast.Visitor{
		PreVisitFrom: func(state any, expr ast.From) any {
			newState := state.(State)
			newState.depth++
			return newState
		},
		PostVisitFrom: func(state any, from ast.From) any {
			newState := state.(State)
			fmt.Println("From:")
			var builder strings.Builder
			indent := strings.Repeat("  ", int(newState.depth))
			builder.WriteString(fmt.Sprintf("%s%s", indent, lexer.DumpTokenString(from.ResultTableExpr)))
			builder.WriteString(fmt.Sprintf("%s%s", indent, lexer.DumpTokenString(from.TableExpr[0])))
			fmt.Print(builder.String())
			newState.depth--
			return newState
		},
		PreVisitWhere: func(state any, where ast.Where) any {
			newState := state.(State)
			newState.depth++
			fmt.Println("Where:")
			var builder strings.Builder
			indent := strings.Repeat("  ", int(newState.depth))
			builder.WriteString(fmt.Sprintf("%s%s", indent, lexer.DumpTokenString(where.ResultTableExpr)))
			fmt.Print(builder.String())
			newState.depth--
			return newState
		},
		PostVisitWhere: func(state any, expr ast.Where) any {
			newState := state.(State)
			return newState
		},
		PreVisitSelect: func(state any, project ast.Select) any {
			newState := state.(State)
			newState.depth++
			fmt.Println("Select:")
			var builder strings.Builder
			indent := strings.Repeat("  ", int(newState.depth))
			builder.WriteString(fmt.Sprintf("%s%s", indent, lexer.DumpTokenString(project.ResultTableExpr)))
			builder.WriteString(fmt.Sprintf("%s%s", indent, lexer.DumpTokenString(project.Fields[0])))
			fmt.Print(builder.String())
			newState.depth--
			return newState
		},
		PostVisitSelect: func(state any, expr ast.Select) any {
			newState := state.(State)
			newState.depth--
			return newState
		},
		PreVisitLogicalExpr: func(state any, expr ast.LogicalExpr) any {
			newState := state.(State)
			newState.depth++
			var builder strings.Builder
			indent := strings.Repeat("  ", int(newState.depth))
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
	var state any
	state = State{depth: 0}
	for _, statement := range astTree {
		fmt.Println(statement.Type)
		switch statement.Type {
		case ast.StatementTypeFrom:
			state = ast.WalkFrom(statement.From, state, visitor)
		case ast.StatementTypeWhere:
			state = ast.WalkWhere(statement.Where, state, visitor)
		case ast.StatementTypeSelect:
			state = ast.WalkSelect(statement.Select, state, visitor)
		}
	}
}

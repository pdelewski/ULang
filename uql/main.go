package main

import (
	"fmt"
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
			var result string
			var indent string
			for i := 0; i < int(newState.depth); i++ {
				indent += "  "
			}
			result += indent
			result += lexer.DumpTokenString(from.ResultTableExpr)
			result += indent
			result += lexer.DumpTokenString(from.TableExpr[0])
			fmt.Print(result)
			newState.depth--
			return newState
		},
		PreVisitWhere: func(state any, where ast.Where) any {
			newState := state.(State)
			newState.depth++
			fmt.Println("Where:")
			var result string
			var indent string
			for i := 0; i < int(newState.depth); i++ {
				indent += "  "
			}
			result += indent
			result += lexer.DumpTokenString(where.ResultTableExpr)
			result += indent
			result += lexer.DumpTokenString(where.ResultTableExpr)
			fmt.Print(result)
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
			var result string
			var indent string
			for i := 0; i < int(newState.depth); i++ {
				indent += "  "
			}
			result += indent
			result += lexer.DumpTokenString(project.ResultTableExpr)
			result += indent
			result += lexer.DumpTokenString(project.Fields[0])
			fmt.Print(result)
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

			var result string
			var indent string
			for i := 0; i < int(newState.depth); i++ {
				indent += "  "
			}
			result += indent
			result += fmt.Sprintf("%sValue:", indent)
			result += lexer.DumpTokensString([]lexer.Token{expr.Value})
			fmt.Print(result)
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
			state = ast.WalkFrom(statement.FromF, state, visitor)
		case ast.StatementTypeWhere:
			state = ast.WalkWhere(statement.WhereF, state, visitor)
		case ast.StatementTypeSelect:
			state = ast.WalkSelect(statement.SelectF, state, visitor)
		}
	}
}

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
		PreVisitJoin: func(state any, join ast.Join) any {
			newState := state.(State)
			newState.depth++
			fmt.Println("Join:")
			var result string
			var indent string
			for i := 0; i < int(newState.depth); i++ {
				indent += "  "
			}
			result += indent
			result += lexer.DumpTokenString(join.ResultTableExpr)
			result += indent
			result += "Left: "
			result += lexer.TokenToString(join.LeftTable)
			result += "\n"
			result += indent
			result += "Right: "
			result += lexer.TokenToString(join.RightTable)
			result += "\n"
			fmt.Print(result)
			return newState
		},
		PostVisitJoin: func(state any, expr ast.Join) any {
			newState := state.(State)
			newState.depth--
			return newState
		},
		PreVisitOrderBy: func(state any, orderBy ast.OrderBy) any {
			newState := state.(State)
			newState.depth++
			fmt.Println("OrderBy:")
			var result string
			var indent string
			for i := 0; i < int(newState.depth); i++ {
				indent += "  "
			}
			result += indent
			result += lexer.DumpTokenString(orderBy.ResultTableExpr)
			result += indent
			result += "Source: "
			result += lexer.TokenToString(orderBy.SourceTable)
			result += "\n"
			result += indent
			result += "Fields: "
			result += lexer.DumpTokensString(orderBy.Fields)
			fmt.Print(result)
			return newState
		},
		PostVisitOrderBy: func(state any, expr ast.OrderBy) any {
			newState := state.(State)
			newState.depth--
			return newState
		},
		PreVisitLimit: func(state any, limit ast.Limit) any {
			newState := state.(State)
			newState.depth++
			fmt.Println("Limit:")
			var result string
			var indent string
			for i := 0; i < int(newState.depth); i++ {
				indent += "  "
			}
			result += indent
			result += lexer.DumpTokenString(limit.ResultTableExpr)
			result += indent
			result += "Source: "
			result += lexer.TokenToString(limit.SourceTable)
			result += "\n"
			result += indent
			result += "Count: "
			result += lexer.TokenToString(limit.Count)
			result += "\n"
			fmt.Print(result)
			return newState
		},
		PostVisitLimit: func(state any, expr ast.Limit) any {
			newState := state.(State)
			newState.depth--
			return newState
		},
		PreVisitGroupBy: func(state any, groupBy ast.GroupBy) any {
			newState := state.(State)
			newState.depth++
			fmt.Println("GroupBy:")
			var result string
			var indent string
			for i := 0; i < int(newState.depth); i++ {
				indent += "  "
			}
			result += indent
			result += lexer.DumpTokenString(groupBy.ResultTableExpr)
			result += indent
			result += "Source: "
			result += lexer.TokenToString(groupBy.SourceTable)
			result += "\n"
			result += indent
			result += "Fields: "
			result += lexer.DumpTokensString(groupBy.Fields)
			fmt.Print(result)
			return newState
		},
		PostVisitGroupBy: func(state any, expr ast.GroupBy) any {
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
 t2 = from table2;
 t3 = join t1 t2 on t1.id == t2.id;
 t4 = where t3.field1 > 10 && t3.field2 < 20;
 t5 = orderby t4 t4.field1 desc t4.field2 asc;
 t6 = limit t5 100;
 t7 = groupby t4 t4.category count t4.id sum t4.amount;
 t8 = select t6.field1;
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
		case ast.StatementTypeJoin:
			state = ast.WalkJoin(statement.JoinF, state, visitor)
		case ast.StatementTypeOrderBy:
			state = ast.WalkOrderBy(statement.OrderByF, state, visitor)
		case ast.StatementTypeLimit:
			state = ast.WalkLimit(statement.LimitF, state, visitor)
		case ast.StatementTypeGroupBy:
			state = ast.WalkGroupBy(statement.GroupByF, state, visitor)
		}
	}
	lexer.TokenizeTest()
}

package emitter

import (
	"uql/ast"
	"uql/lexer"
)

type UqlEmitterState struct {
	Depth  int8
	Result string
}

func statementTypeToString(t int8) string {
	if t == ast.StatementTypeFrom {
		return "1"
	}
	if t == ast.StatementTypeWhere {
		return "2"
	}
	if t == ast.StatementTypeSelect {
		return "3"
	}
	if t == ast.StatementTypeJoin {
		return "4"
	}
	if t == ast.StatementTypeOrderBy {
		return "5"
	}
	if t == ast.StatementTypeLimit {
		return "6"
	}
	if t == ast.StatementTypeGroupBy {
		return "7"
	}
	return "0"
}

func EmitUql(statements []ast.Statement) string {
	visitor := ast.Visitor{
		PreVisitFrom: func(state any, expr ast.From) any {
			newState := state.(UqlEmitterState)
			newState.Depth++
			return newState
		},
		PostVisitFrom: func(state any, from ast.From) any {
			newState := state.(UqlEmitterState)
			newState.Result += "From:\n"
			var indent string
			for i := 0; i < int(newState.Depth); i++ {
				indent += "  "
			}
			newState.Result += indent
			newState.Result += lexer.DumpTokenString(from.ResultTableExpr)
			newState.Result += indent
			newState.Result += lexer.DumpTokenString(from.TableExpr[0])
			newState.Depth--
			return newState
		},
		PreVisitWhere: func(state any, where ast.Where) any {
			newState := state.(UqlEmitterState)
			newState.Depth++
			newState.Result += "Where:\n"
			var indent string
			for i := 0; i < int(newState.Depth); i++ {
				indent += "  "
			}
			newState.Result += indent
			newState.Result += lexer.DumpTokenString(where.ResultTableExpr)
			newState.Result += indent
			newState.Result += lexer.DumpTokenString(where.ResultTableExpr)
			newState.Depth--
			return newState
		},
		PostVisitWhere: func(state any, expr ast.Where) any {
			newState := state.(UqlEmitterState)
			return newState
		},
		PreVisitSelect: func(state any, project ast.Select) any {
			newState := state.(UqlEmitterState)
			newState.Depth++
			newState.Result += "Select:\n"
			var indent string
			for i := 0; i < int(newState.Depth); i++ {
				indent += "  "
			}
			newState.Result += indent
			newState.Result += lexer.DumpTokenString(project.ResultTableExpr)
			newState.Result += indent
			newState.Result += lexer.DumpTokenString(project.Fields[0])
			newState.Depth--
			return newState
		},
		PostVisitSelect: func(state any, expr ast.Select) any {
			newState := state.(UqlEmitterState)
			newState.Depth--
			return newState
		},
		PreVisitJoin: func(state any, join ast.Join) any {
			newState := state.(UqlEmitterState)
			newState.Depth++
			newState.Result += "Join:\n"
			var indent string
			for i := 0; i < int(newState.Depth); i++ {
				indent += "  "
			}
			newState.Result += indent
			newState.Result += lexer.DumpTokenString(join.ResultTableExpr)
			newState.Result += indent
			newState.Result += "Left: "
			newState.Result += lexer.TokenToString(join.LeftTable)
			newState.Result += "\n"
			newState.Result += indent
			newState.Result += "Right: "
			newState.Result += lexer.TokenToString(join.RightTable)
			newState.Result += "\n"
			return newState
		},
		PostVisitJoin: func(state any, expr ast.Join) any {
			newState := state.(UqlEmitterState)
			newState.Depth--
			return newState
		},
		PreVisitOrderBy: func(state any, orderBy ast.OrderBy) any {
			newState := state.(UqlEmitterState)
			newState.Depth++
			newState.Result += "OrderBy:\n"
			var indent string
			for i := 0; i < int(newState.Depth); i++ {
				indent += "  "
			}
			newState.Result += indent
			newState.Result += lexer.DumpTokenString(orderBy.ResultTableExpr)
			newState.Result += indent
			newState.Result += "Source: "
			newState.Result += lexer.TokenToString(orderBy.SourceTable)
			newState.Result += "\n"
			newState.Result += indent
			newState.Result += "Fields: "
			newState.Result += lexer.DumpTokensString(orderBy.Fields)
			return newState
		},
		PostVisitOrderBy: func(state any, expr ast.OrderBy) any {
			newState := state.(UqlEmitterState)
			newState.Depth--
			return newState
		},
		PreVisitLimit: func(state any, limit ast.Limit) any {
			newState := state.(UqlEmitterState)
			newState.Depth++
			newState.Result += "Limit:\n"
			var indent string
			for i := 0; i < int(newState.Depth); i++ {
				indent += "  "
			}
			newState.Result += indent
			newState.Result += lexer.DumpTokenString(limit.ResultTableExpr)
			newState.Result += indent
			newState.Result += "Source: "
			newState.Result += lexer.TokenToString(limit.SourceTable)
			newState.Result += "\n"
			newState.Result += indent
			newState.Result += "Count: "
			newState.Result += lexer.TokenToString(limit.Count)
			newState.Result += "\n"
			return newState
		},
		PostVisitLimit: func(state any, expr ast.Limit) any {
			newState := state.(UqlEmitterState)
			newState.Depth--
			return newState
		},
		PreVisitGroupBy: func(state any, groupBy ast.GroupBy) any {
			newState := state.(UqlEmitterState)
			newState.Depth++
			newState.Result += "GroupBy:\n"
			var indent string
			for i := 0; i < int(newState.Depth); i++ {
				indent += "  "
			}
			newState.Result += indent
			newState.Result += lexer.DumpTokenString(groupBy.ResultTableExpr)
			newState.Result += indent
			newState.Result += "Source: "
			newState.Result += lexer.TokenToString(groupBy.SourceTable)
			newState.Result += "\n"
			newState.Result += indent
			newState.Result += "Fields: "
			newState.Result += lexer.DumpTokensString(groupBy.Fields)
			return newState
		},
		PostVisitGroupBy: func(state any, expr ast.GroupBy) any {
			newState := state.(UqlEmitterState)
			newState.Depth--
			return newState
		},
		PreVisitLogicalExpr: func(state any, expr ast.LogicalExpr) any {
			newState := state.(UqlEmitterState)
			newState.Depth++
			var indent string
			for i := 0; i < int(newState.Depth); i++ {
				indent += "  "
			}
			newState.Result += indent
			newState.Result += lexer.DumpTokensString([]lexer.Token{expr.Value})
			return newState
		},
		PostVisitLogicalExpr: func(state any, expr ast.LogicalExpr) any {
			newState := state.(UqlEmitterState)
			newState.Depth--
			return newState
		},
	}

	var state any
	state = UqlEmitterState{Depth: 0, Result: ""}

	for i := 0; i < len(statements); i++ {
		statement := statements[i]
		currentState := state.(UqlEmitterState)
		currentState.Result += statementTypeToString(statement.Type)
		currentState.Result += "\n"
		state = currentState

		if statement.Type == ast.StatementTypeFrom {
			state = ast.WalkFrom(statement.FromF, state, visitor)
		}
		if statement.Type == ast.StatementTypeWhere {
			state = ast.WalkWhere(statement.WhereF, state, visitor)
		}
		if statement.Type == ast.StatementTypeSelect {
			state = ast.WalkSelect(statement.SelectF, state, visitor)
		}
		if statement.Type == ast.StatementTypeJoin {
			state = ast.WalkJoin(statement.JoinF, state, visitor)
		}
		if statement.Type == ast.StatementTypeOrderBy {
			state = ast.WalkOrderBy(statement.OrderByF, state, visitor)
		}
		if statement.Type == ast.StatementTypeLimit {
			state = ast.WalkLimit(statement.LimitF, state, visitor)
		}
		if statement.Type == ast.StatementTypeGroupBy {
			state = ast.WalkGroupBy(statement.GroupByF, state, visitor)
		}
	}

	finalState := state.(UqlEmitterState)
	return finalState.Result
}

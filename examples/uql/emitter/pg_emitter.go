package emitter

import (
	"uql/ast"
	"uql/lexer"
)

type EmitterState struct {
	Result string
	First  bool
}

func emitExpression(expr ast.PgExpression) string {
	if expr.Type == ast.PgExprTypeBinaryOp {
		return emitBinaryOp(expr)
	}
	if expr.Type == ast.PgExprTypeUnaryOp {
		return emitUnaryOp(expr)
	}
	if expr.Type == ast.PgExprTypeFunction {
		return emitFunction(expr)
	}
	return lexer.TokenToString(expr.Value)
}

func emitBinaryOp(expr ast.PgExpression) string {
	if len(expr.Expressions) < 2 {
		return lexer.TokenToString(expr.Value)
	}

	var left string
	var right string
	var op string
	left = emitExpression(expr.Expressions[0])
	right = emitExpression(expr.Expressions[1])
	op = lexer.TokenToString(expr.Value)

	if op == "==" {
		op = "="
	}

	var result string
	result += "("
	result += left
	result += " "
	result += op
	result += " "
	result += right
	result += ")"
	return result
}

func emitUnaryOp(expr ast.PgExpression) string {
	if len(expr.Expressions) < 1 {
		return lexer.TokenToString(expr.Value)
	}

	var op string
	var operand string
	op = lexer.TokenToString(expr.Value)
	operand = emitExpression(expr.Expressions[0])

	var result string
	result += op
	result += " "
	result += operand
	return result
}

func emitFunction(expr ast.PgExpression) string {
	var result string
	result += lexer.TokenToString(expr.Value)
	result += "("
	for i := 0; i < len(expr.Expressions); i++ {
		if i > 0 {
			result += ", "
		}
		result += emitExpression(expr.Expressions[i])
	}
	result += ")"
	return result
}

func emitJoinType(joinType int8) string {
	if joinType == ast.PgJoinTypeInner {
		return "INNER"
	}
	if joinType == ast.PgJoinTypeLeft {
		return "LEFT"
	}
	if joinType == ast.PgJoinTypeRight {
		return "RIGHT"
	}
	if joinType == ast.PgJoinTypeFull {
		return "FULL"
	}
	return "INNER"
}

func EmitPostgreSQL(stmt ast.PgSelectStatement) string {
	visitor := ast.PgVisitor{
		PreVisitSelect: func(state any, stmt ast.PgSelectStatement) any {
			s := state.(EmitterState)

			// SELECT clause
			s.Result += "SELECT "
			if stmt.Distinct {
				s.Result += "DISTINCT "
			}

			// Fields
			if len(stmt.Fields) == 0 {
				s.Result += "*"
			} else {
				var first bool
				first = true
				for i := 0; i < len(stmt.Fields); i++ {
					var fieldStr string
					fieldStr = emitExpression(stmt.Fields[i].Expression)
					if len(fieldStr) == 0 || fieldStr == ";" {
						continue
					}
					if !first {
						s.Result += ", "
					}
					first = false
					s.Result += fieldStr
					if len(stmt.Fields[i].Alias.Representation) > 0 {
						s.Result += " AS "
						s.Result += lexer.TokenToString(stmt.Fields[i].Alias)
					}
				}
			}

			// FROM clause
			if len(stmt.From.Table.Representation) > 0 {
				s.Result += " FROM "
				s.Result += lexer.TokenToString(stmt.From.Table)
				if len(stmt.From.Alias.Representation) > 0 {
					s.Result += " AS "
					s.Result += lexer.TokenToString(stmt.From.Alias)
				}
			}

			// JOIN clauses
			for i := 0; i < len(stmt.Joins); i++ {
				join := stmt.Joins[i]
				s.Result += " "
				s.Result += emitJoinType(join.JoinType)
				s.Result += " JOIN "
				s.Result += lexer.TokenToString(join.Table)
				if len(join.Alias.Representation) > 0 {
					s.Result += " AS "
					s.Result += lexer.TokenToString(join.Alias)
				}
				if join.Condition.Type != 0 {
					s.Result += " ON "
					s.Result += emitExpression(join.Condition)
				}
			}

			// WHERE clause
			if stmt.Where.Condition.Type != 0 {
				s.Result += " WHERE "
				s.Result += emitExpression(stmt.Where.Condition)
			}

			// GROUP BY clause
			if len(stmt.GroupBy.Fields) > 0 {
				s.Result += " GROUP BY "
				for i := 0; i < len(stmt.GroupBy.Fields); i++ {
					if i > 0 {
						s.Result += ", "
					}
					s.Result += emitExpression(stmt.GroupBy.Fields[i])
				}
			}

			// HAVING clause
			if stmt.Having.Condition.Type != 0 {
				s.Result += " HAVING "
				s.Result += emitExpression(stmt.Having.Condition)
			}

			// ORDER BY clause
			if len(stmt.OrderBy.Fields) > 0 {
				s.Result += " ORDER BY "
				for i := 0; i < len(stmt.OrderBy.Fields); i++ {
					if i > 0 {
						s.Result += ", "
					}
					s.Result += emitExpression(stmt.OrderBy.Fields[i].Field)
					if stmt.OrderBy.Fields[i].Direction == ast.PgOrderDesc {
						s.Result += " DESC"
					}
				}
			}

			// LIMIT clause
			if len(stmt.Limit.Count.Representation) > 0 {
				s.Result += " LIMIT "
				s.Result += lexer.TokenToString(stmt.Limit.Count)
			}

			// OFFSET clause
			if len(stmt.Offset.Count.Representation) > 0 {
				s.Result += " OFFSET "
				s.Result += lexer.TokenToString(stmt.Offset.Count)
			}

			return s
		},
		PostVisitSelect: func(state any, stmt ast.PgSelectStatement) any {
			return state
		},
		PreVisitExpr: func(state any, expr ast.PgExpression) any {
			return state
		},
		PostVisitExpr: func(state any, expr ast.PgExpression) any {
			return state
		},
	}

	var state any
	state = EmitterState{Result: "", First: true}
	state = ast.WalkPgSelect(stmt, state, visitor)
	finalState := state.(EmitterState)
	return finalState.Result
}

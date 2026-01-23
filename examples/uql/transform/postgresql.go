package transform

import (
	"uql/ast"
	"uql/lexer"
)

type TableBinding struct {
	Name      lexer.Token
	Statement ast.Statement
}

type TransformState struct {
	Bindings []TableBinding
}

func findBinding(state TransformState, name lexer.Token) int {
	nameStr := lexer.TokenToString(name)
	for i := 0; i < len(state.Bindings); i++ {
		if lexer.TokenToString(state.Bindings[i].Name) == nameStr {
			return i
		}
	}
	return -1
}

func TransformToPostgreSQL(uqlAst []ast.Statement) ast.PgSelectStatement {
	var state TransformState
	var result ast.PgSelectStatement

	// First pass: collect all bindings
	for i := 0; i < len(uqlAst); i++ {
		var binding TableBinding
		binding.Statement = uqlAst[i]
		switch uqlAst[i].Type {
		case ast.StatementTypeFrom:
			binding.Name = uqlAst[i].FromF.ResultTableExpr
		case ast.StatementTypeWhere:
			binding.Name = uqlAst[i].WhereF.ResultTableExpr
		case ast.StatementTypeSelect:
			binding.Name = uqlAst[i].SelectF.ResultTableExpr
		case ast.StatementTypeJoin:
			binding.Name = uqlAst[i].JoinF.ResultTableExpr
		case ast.StatementTypeOrderBy:
			binding.Name = uqlAst[i].OrderByF.ResultTableExpr
		case ast.StatementTypeLimit:
			binding.Name = uqlAst[i].LimitF.ResultTableExpr
		case ast.StatementTypeGroupBy:
			binding.Name = uqlAst[i].GroupByF.ResultTableExpr
		}
		state.Bindings = append(state.Bindings, binding)
	}

	// Second pass: build PostgreSQL AST
	for i := 0; i < len(uqlAst); i++ {
		stmt := uqlAst[i]
		switch stmt.Type {
		case ast.StatementTypeFrom:
			result = transformFrom(state, result, stmt.FromF)
		case ast.StatementTypeWhere:
			result = transformWhere(state, result, stmt.WhereF)
		case ast.StatementTypeSelect:
			result = transformSelect(state, result, stmt.SelectF)
		case ast.StatementTypeJoin:
			result = transformJoin(state, result, stmt.JoinF)
		case ast.StatementTypeOrderBy:
			result = transformOrderBy(state, result, stmt.OrderByF)
		case ast.StatementTypeLimit:
			result = transformLimit(state, result, stmt.LimitF)
		case ast.StatementTypeGroupBy:
			result = transformGroupBy(state, result, stmt.GroupByF)
		}
	}

	return result
}

func transformFrom(state TransformState, result ast.PgSelectStatement, from ast.From) ast.PgSelectStatement {
	if len(from.TableExpr) > 0 {
		result.From = ast.PgFromClause{
			Table: from.TableExpr[0],
		}
	}
	return result
}

func transformWhere(state TransformState, result ast.PgSelectStatement, where ast.Where) ast.PgSelectStatement {
	result.Where = ast.PgWhereClause{
		Condition: transformLogicalExpr(where.Expr),
	}
	return result
}

func transformSelect(state TransformState, result ast.PgSelectStatement, sel ast.Select) ast.PgSelectStatement {
	for i := 0; i < len(sel.Fields); i++ {
		field := ast.PgSelectField{
			Expression: ast.PgExpression{
				Type:  int8(ast.PgExprTypeColumn),
				Value: sel.Fields[i],
			},
		}
		result.Fields = append(result.Fields, field)
	}
	return result
}

func transformJoin(state TransformState, result ast.PgSelectStatement, join ast.Join) ast.PgSelectStatement {
	// Find the actual table names from bindings
	leftIdx := findBinding(state, join.LeftTable)
	rightIdx := findBinding(state, join.RightTable)

	var leftTable lexer.Token
	var rightTable lexer.Token

	if leftIdx >= 0 {
		leftStmt := state.Bindings[leftIdx].Statement
		if leftStmt.Type == ast.StatementTypeFrom {
			if len(leftStmt.FromF.TableExpr) > 0 {
				leftTable = leftStmt.FromF.TableExpr[0]
			}
		}
	}

	if rightIdx >= 0 {
		rightStmt := state.Bindings[rightIdx].Statement
		if rightStmt.Type == ast.StatementTypeFrom {
			if len(rightStmt.FromF.TableExpr) > 0 {
				rightTable = rightStmt.FromF.TableExpr[0]
			}
		}
	}

	// Set FROM to left table if not already set
	if len(result.From.Table.Representation) == 0 {
		result.From = ast.PgFromClause{
			Table: leftTable,
			Alias: join.LeftTable,
		}
	}

	// Add JOIN clause
	joinClause := ast.PgJoinClause{
		JoinType:  ast.PgJoinTypeInner,
		Table:     rightTable,
		Alias:     join.RightTable,
		Condition: transformLogicalExpr(join.OnCondition),
	}
	result.Joins = append(result.Joins, joinClause)

	return result
}

func transformOrderBy(state TransformState, result ast.PgSelectStatement, orderBy ast.OrderBy) ast.PgSelectStatement {
	for i := 0; i < len(orderBy.Fields); i++ {
		field := ast.PgOrderByField{
			Field: ast.PgExpression{
				Type:  int8(ast.PgExprTypeColumn),
				Value: orderBy.Fields[i],
			},
			Direction: ast.PgOrderAsc,
		}
		result.OrderBy.Fields = append(result.OrderBy.Fields, field)
	}
	return result
}

func transformLimit(state TransformState, result ast.PgSelectStatement, limit ast.Limit) ast.PgSelectStatement {
	result.Limit = ast.PgLimitClause{
		Count: limit.Count,
	}
	return result
}

func transformGroupBy(state TransformState, result ast.PgSelectStatement, groupBy ast.GroupBy) ast.PgSelectStatement {
	for i := 0; i < len(groupBy.Fields); i++ {
		expr := ast.PgExpression{
			Type:  int8(ast.PgExprTypeColumn),
			Value: groupBy.Fields[i],
		}
		result.GroupBy.Fields = append(result.GroupBy.Fields, expr)
	}
	return result
}

func transformLogicalExpr(expr ast.LogicalExpr) ast.PgExpression {
	result := ast.PgExpression{
		Value: expr.Value,
		Left:  int16(expr.Left),
		Right: int16(expr.Right),
	}

	if expr.Left != 0 || expr.Right != 0 {
		result.Type = int8(ast.PgExprTypeBinaryOp)
		for i := 0; i < len(expr.Expressions); i++ {
			result.Expressions = append(result.Expressions, transformLogicalExpr(expr.Expressions[i]))
		}
	} else {
		result.Type = int8(ast.PgExprTypeValue)
	}

	return result
}

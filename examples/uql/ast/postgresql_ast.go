package ast

import (
	"uql/lexer"
)

const (
	PgJoinTypeInner = 1
	PgJoinTypeLeft  = 2
	PgJoinTypeRight = 3
	PgJoinTypeFull  = 4
)

const (
	PgOrderAsc  = 1
	PgOrderDesc = 2
)

type PgSelectStatement struct {
	Distinct    bool
	Fields      []PgSelectField
	From        PgFromClause
	Joins       []PgJoinClause
	Where       PgWhereClause
	GroupBy     PgGroupByClause
	Having      PgHavingClause
	OrderBy     PgOrderByClause
	Limit       PgLimitClause
	Offset      PgOffsetClause
}

type PgSelectField struct {
	Expression PgExpression
	Alias      lexer.Token
}

type PgFromClause struct {
	Table lexer.Token
	Alias lexer.Token
}

type PgJoinClause struct {
	JoinType  int8
	Table     lexer.Token
	Alias     lexer.Token
	Condition PgExpression
}

type PgWhereClause struct {
	Condition PgExpression
}

type PgGroupByClause struct {
	Fields []PgExpression
}

type PgHavingClause struct {
	Condition PgExpression
}

type PgOrderByField struct {
	Field     PgExpression
	Direction int8
}

type PgOrderByClause struct {
	Fields []PgOrderByField
}

type PgLimitClause struct {
	Count lexer.Token
}

type PgOffsetClause struct {
	Count lexer.Token
}

type PgExpression struct {
	Type        int8
	Value       lexer.Token
	Left        int16
	Right       int16
	Expressions []PgExpression
}

const (
	PgExprTypeValue    = 1
	PgExprTypeColumn   = 2
	PgExprTypeBinaryOp = 3
	PgExprTypeUnaryOp  = 4
	PgExprTypeFunction = 5
)

type PgFunctionCall struct {
	Name      lexer.Token
	Arguments []PgExpression
}

type PgAggregate struct {
	Function lexer.Token
	Field    PgExpression
	Alias    lexer.Token
}

type PgVisitor struct {
	PreVisitSelect  func(state any, stmt PgSelectStatement) any
	PostVisitSelect func(state any, stmt PgSelectStatement) any
	PreVisitExpr    func(state any, expr PgExpression) any
	PostVisitExpr   func(state any, expr PgExpression) any
}

func WalkPgSelect(stmt PgSelectStatement, state any, visitor PgVisitor) any {
	state = visitor.PreVisitSelect(state, stmt)
	state = visitor.PostVisitSelect(state, stmt)
	return state
}

func WalkPgExpr(expr PgExpression, state any, visitor PgVisitor) any {
	state = visitor.PreVisitExpr(state, expr)
	if expr.Left != 0 || expr.Right != 0 {
		state = WalkPgExpr(expr.Expressions[0], state, visitor)
		state = WalkPgExpr(expr.Expressions[1], state, visitor)
	}
	state = visitor.PostVisitExpr(state, expr)
	return state
}

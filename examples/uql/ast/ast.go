package ast

import (
	"uql/lexer"
)

const (
	StatementTypeFrom    = 1
	StatementTypeWhere   = 2
	StatementTypeSelect  = 3
	StatementTypeJoin    = 4
	StatementTypeOrderBy = 5
	StatementTypeLimit   = 6
	StatementTypeGroupBy = 7
)

type Statement struct {
	Type     int8
	FromF    From
	WhereF   Where
	SelectF  Select
	JoinF    Join
	OrderByF OrderBy
	LimitF   Limit
	GroupByF GroupBy
}

type AST []Statement

type From struct {
	TableExpr       []lexer.Token
	ResultTableExpr lexer.Token
}

type LogicalExpr struct {
	Value       lexer.Token
	Left        uint16
	Right       uint16
	Expressions []LogicalExpr
}

type Where struct {
	Expr            LogicalExpr
	ResultTableExpr lexer.Token
}

type Select struct {
	Fields          []lexer.Token
	ResultTableExpr lexer.Token
}

type Join struct {
	LeftTable       lexer.Token
	RightTable      lexer.Token
	OnCondition     LogicalExpr
	ResultTableExpr lexer.Token
}

type OrderBy struct {
	SourceTable     lexer.Token
	Fields          []lexer.Token
	ResultTableExpr lexer.Token
}

type Limit struct {
	SourceTable     lexer.Token
	Count           lexer.Token
	ResultTableExpr lexer.Token
}

type GroupBy struct {
	SourceTable     lexer.Token
	Fields          []lexer.Token
	Aggregates      []Aggregate
	ResultTableExpr lexer.Token
}

type Aggregate struct {
	Function lexer.Token
	Field    lexer.Token
	Alias    lexer.Token
}

type Visitor struct {
	PreVisitFrom         func(state any, expr From) any
	PostVisitFrom        func(state any, expr From) any
	PreVisitWhere        func(state any, expr Where) any
	PostVisitWhere       func(state any, expr Where) any
	PreVisitSelect       func(state any, expr Select) any
	PostVisitSelect      func(state any, expr Select) any
	PreVisitJoin         func(state any, expr Join) any
	PostVisitJoin        func(state any, expr Join) any
	PreVisitOrderBy      func(state any, expr OrderBy) any
	PostVisitOrderBy     func(state any, expr OrderBy) any
	PreVisitLimit        func(state any, expr Limit) any
	PostVisitLimit       func(state any, expr Limit) any
	PreVisitGroupBy      func(state any, expr GroupBy) any
	PostVisitGroupBy     func(state any, expr GroupBy) any
	PreVisitLogicalExpr  func(state any, expr LogicalExpr) any
	PostVisitLogicalExpr func(state any, expr LogicalExpr) any
}

func WalkFrom(expr From,
	state any,
	visitor Visitor,
) any {
	state = visitor.PreVisitFrom(state, expr)
	state = visitor.PostVisitFrom(state, expr)
	return state
}

func WalkWhere(where Where,
	state any,
	visitor Visitor,
) any {
	state = visitor.PreVisitWhere(state, where)
	state = walkLogicalExpr(where.Expr, state, visitor)
	state = visitor.PostVisitWhere(state, where)
	return state
}

func WalkSelect(expr Select,
	state any,
	visitor Visitor,
) any {
	state = visitor.PreVisitSelect(state, expr)
	state = visitor.PostVisitSelect(state, expr)
	return state
}

func walkLogicalExpr(
	expr LogicalExpr,
	state any,
	visitor Visitor,
) any {
	state = visitor.PreVisitLogicalExpr(state, expr)
	if expr.Left != 0 || expr.Right != 0 {
		state = walkLogicalExpr(expr.Expressions[0], state, visitor)
		state = walkLogicalExpr(expr.Expressions[1], state, visitor)
	}
	state = visitor.PostVisitLogicalExpr(state, expr)
	return state
}

func WalkJoin(expr Join,
	state any,
	visitor Visitor,
) any {
	state = visitor.PreVisitJoin(state, expr)
	state = walkLogicalExpr(expr.OnCondition, state, visitor)
	state = visitor.PostVisitJoin(state, expr)
	return state
}

func WalkOrderBy(expr OrderBy,
	state any,
	visitor Visitor,
) any {
	state = visitor.PreVisitOrderBy(state, expr)
	state = visitor.PostVisitOrderBy(state, expr)
	return state
}

func WalkLimit(expr Limit,
	state any,
	visitor Visitor,
) any {
	state = visitor.PreVisitLimit(state, expr)
	state = visitor.PostVisitLimit(state, expr)
	return state
}

func WalkGroupBy(expr GroupBy,
	state any,
	visitor Visitor,
) any {
	state = visitor.PreVisitGroupBy(state, expr)
	state = visitor.PostVisitGroupBy(state, expr)
	return state
}

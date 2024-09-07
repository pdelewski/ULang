package ast

import "ULang/lexer"

const (
	StatementTypeFrom = iota
	StatementTypeWhere
	StatementTypeSelect
)

type Statement struct {
	Type   int8
	From   From
	Where  Where
	Select Select
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

func WalkFrom(expr From,
	state any,
	preVisit func(state any, expr any) any,
	postVisit func(state any, expr any) any,
) {
	preVisit(state, expr)
	postVisit(state, expr)
}

func WalkWhere(expr Where,
	state any,
	preVisit func(state any, expr any) any,
	postVisit func(state any, expr any) any,
) {
	preVisit(state, expr)
	WalkLogicalExpr(expr.Expr, state, preVisit, postVisit)
	postVisit(state, expr)
}

func WalkSelect(expr Select,
	state any,
	preVisit func(state any, expr any) any,
	postVisit func(state any, expr any) any,
) {
	preVisit(state, expr)
	postVisit(state, expr)
}

func WalkLogicalExpr(expr LogicalExpr,
	state any,
	preVisit func(state any, expr any) any,
	postVisit func(state any, expr any) any,
) {
	walkLogicalExpr(expr, 0, state, preVisit, postVisit)
}

func walkLogicalExpr(
	expr LogicalExpr,
	depth int,
	state any,
	preVisit func(state any, expr any) any,
	postVisit func(state any, expr any) any,
) {
	state = preVisit(state, expr)
	if expr.Left != 0 || expr.Right != 0 {
		walkLogicalExpr(expr.Expressions[0], depth+1, state, preVisit, postVisit)
		walkLogicalExpr(expr.Expressions[1], depth+1, state, preVisit, postVisit)
	}
	state = postVisit(state, expr)
}

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

func WalkLogicalExpr(expr LogicalExpr,
	state any,
	preVisit func(state any, expr LogicalExpr) any,
	postVisit func(state any, expr LogicalExpr) any,
) {
	walkLogicalExpr(expr, 0, state, preVisit, postVisit)
}

func walkLogicalExpr(
	expr LogicalExpr,
	depth int,
	state any,
	preVisit func(state any, expr LogicalExpr) any,
	postVisit func(state any, expr LogicalExpr) any,
) {
	state = preVisit(state, expr)
	if expr.Left != 0 || expr.Right != 0 {
		walkLogicalExpr(expr.Expressions[0], depth+1, state, preVisit, postVisit)
		walkLogicalExpr(expr.Expressions[1], depth+1, state, preVisit, postVisit)
	}
	state = postVisit(state, expr)
}

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

type Visitor struct {
	PreVisitFrom         func(state any, expr From) any
	PostVisitFrom        func(state any, expr From) any
	PreVisitWhere        func(state any, expr Where) any
	PostVisitWhere       func(state any, expr Where) any
	PreVisitSelect       func(state any, expr Select) any
	PostVisitSelect      func(state any, expr Select) any
	PreVisitLogicalExpr  func(state any, expr LogicalExpr) any
	PostVisitLogicalExpr func(state any, expr LogicalExpr) any
}

func WalkFrom(expr From,
	state any,
	visitor Visitor,
) {
	state = visitor.PreVisitFrom(state, expr)
	state = visitor.PostVisitFrom(state, expr)
}

func WalkWhere(expr Where,
	state any,
	visitor Visitor,
) {
	state = visitor.PreVisitWhere(state, expr)
	WalkLogicalExpr(expr.Expr, state, visitor)
	state = visitor.PostVisitWhere(state, expr)
}

func WalkSelect(expr Select,
	state any,
	visitor Visitor,
) {
	state = visitor.PreVisitSelect(state, expr)
	state = visitor.PostVisitSelect(state, expr)
}

func WalkLogicalExpr(expr LogicalExpr,
	state any,
	visitor Visitor,
) {
	walkLogicalExpr(expr, state, visitor)
}

func walkLogicalExpr(
	expr LogicalExpr,
	state any,
	visitor Visitor,
) {
	state = visitor.PreVisitLogicalExpr(state, expr)
	if expr.Left != 0 || expr.Right != 0 {
		walkLogicalExpr(expr.Expressions[0], state, visitor)
		walkLogicalExpr(expr.Expressions[1], state, visitor)
	}
	state = visitor.PostVisitLogicalExpr(state, expr)
}

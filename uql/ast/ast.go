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

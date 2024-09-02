package ast

import "ULang/lexer"

type From struct {
	TableExpr []lexer.Token
}

type LogicalExpr struct {
	Op          lexer.Token
	Left        uint16
	Right       uint16
	Expressions []LogicalExpr
}

type Where struct {
	Expr LogicalExpr
}

type Select struct {
	Fields []lexer.Token
}

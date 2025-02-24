package main

import (
	"go/ast"
	"go/token"
	"os"
)

type Emitter interface {
	SetFile(file *os.File)
	PreVisitBasicLit(node *ast.BasicLit, indent int)
	PostVisitBasicLit(node *ast.BasicLit, indent int)
	PreVisitIdent(node *ast.Ident, indent int)
	PostVisitIdent(node *ast.Ident, indent int)
	PreVisitBinaryExpr(node *ast.BinaryExpr, indent int)
	PostVisitBinaryExpr(node *ast.BinaryExpr, indent int)
	PreVisitBinaryExprLeft(node ast.Expr, indent int)
	PostVisitBinaryExprLeft(node ast.Expr, indent int)
	PreVisitBinaryExprRight(node ast.Expr, indent int)
	PostVisitBinaryExprRight(node ast.Expr, indent int)
	PreVisitBinaryExprOperator(op token.Token, indent int)
	PostVisitBinaryExprOperator(op token.Token, indent int)
}

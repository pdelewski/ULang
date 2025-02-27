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
	PreVisitCallExpr(node *ast.CallExpr, indent int)
	PostVisitCallExpr(node *ast.CallExpr, indent int)
	PreVisitCallExprFun(node ast.Expr, indent int)
	PostVisitCallExprFun(node ast.Expr, indent int)
	PreVisitCallExprArgs(node []ast.Expr, indent int)
	PostVisitCallExprArgs(node []ast.Expr, indent int)
	PreVisitCallExprArg(node ast.Expr, index int, indent int)
	PostVisitCallExprArg(node ast.Expr, index int, indent int)
	PreVisitParenExpr(node *ast.ParenExpr, indent int)
	PostVisitParenExpr(node *ast.ParenExpr, indent int)
	PreVisitCompositeLit(node *ast.CompositeLit, indent int)
	PostVisitCompositeLit(node *ast.CompositeLit, indent int)
	PreVisitCompositeLitType(node ast.Expr, indent int)
	PostVisitCompositeLitType(node ast.Expr, indent int)
	PreVisitCompositeLitElts(node []ast.Expr, indent int)
	PostVisitCompositeLitElts(node []ast.Expr, indent int)
	PreVisitCompositeLitElt(node ast.Expr, index int, indent int)
	PostVisitCompositeLitElt(node ast.Expr, index int, indent int)
	PreVisitArrayType(node ast.ArrayType, indent int)
	PostVisitArrayType(node ast.ArrayType, indent int)
	PreVisitSelectorExpr(node *ast.SelectorExpr, indent int)
	PostVisitSelectorExpr(node *ast.SelectorExpr, indent int)
	PreVisitSelectorExprX(node ast.Expr, indent int)
	PostVisitSelectorExprX(node ast.Expr, indent int)
	PreVisitSelectorExprSel(node *ast.Ident, indent int)
	PostVisitSelectorExprSel(node *ast.Ident, indent int)
}

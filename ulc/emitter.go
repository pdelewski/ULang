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
	PreVisitIndexExpr(node *ast.IndexExpr, indent int)
	PostVisitIndexExpr(node *ast.IndexExpr, indent int)
	PreVisitIndexExprX(node *ast.IndexExpr, indent int)
	PostVisitIndexExprX(node *ast.IndexExpr, indent int)
	PreVisitIndexExprIndex(node *ast.IndexExpr, indent int)
	PostVisitIndexExprIndex(node *ast.IndexExpr, indent int)
	PreVisitUnaryExpr(node *ast.UnaryExpr, indent int)
	PostVisitUnaryExpr(node *ast.UnaryExpr, indent int)
	PreVisitSliceExpr(node *ast.SliceExpr, indent int)
	PostVisitSliceExpr(node *ast.SliceExpr, indent int)
	PreVisitSliceExprX(node ast.Expr, indent int)
	PostVisitSliceExprX(node ast.Expr, indent int)
	PreVisitSliceExprXBegin(node ast.Expr, indent int)
	PostVisitSliceExprXBegin(node ast.Expr, indent int)
	PreVisitSliceExprXEnd(node ast.Expr, indent int)
	PostVisitSliceExprXEnd(node ast.Expr, indent int)
	PreVisitSliceExprLow(node ast.Expr, indent int)
	PostVisitSliceExprLow(node ast.Expr, indent int)
	PreVisitSliceExprHigh(node ast.Expr, indent int)
	PostVisitSliceExprHigh(node ast.Expr, indent int)
	PreVisitFuncType(node *ast.FuncType, indent int)
	PostVisitFuncType(node *ast.FuncType, indent int)
	PreVisitFuncTypeResults(node *ast.FieldList, indent int)
	PostVisitFuncTypeResults(node *ast.FieldList, indent int)
	PreVisitFuncTypeResult(node *ast.Field, index int, indent int)
	PostVisitFuncTypeResult(node *ast.Field, index int, indent int)
	PreVisitFuncTypeParams(node *ast.FieldList, indent int)
	PostVisitFuncTypeParams(node *ast.FieldList, indent int)
	PreVisitFuncTypeParam(node *ast.Field, index int, indent int)
	PostVisitFuncTypeParam(node *ast.Field, index int, indent int)
}

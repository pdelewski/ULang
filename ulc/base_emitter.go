package main

import (
	"go/ast"
	"go/token"
	"os"
)

type BaseEmitter struct{}

func (v *BaseEmitter) SetFile(file *os.File)                                  {}
func (v *BaseEmitter) PreVisitBasicLit(node *ast.BasicLit, indent int)        {}
func (v *BaseEmitter) PostVisitBasicLit(node *ast.BasicLit, indent int)       {}
func (v *BaseEmitter) PreVisitIdent(node *ast.Ident, indent int)              {}
func (v *BaseEmitter) PostVisitIdent(node *ast.Ident, indent int)             {}
func (v *BaseEmitter) PreVisitBinaryExpr(node *ast.BinaryExpr, indent int)    {}
func (v *BaseEmitter) PostVisitBinaryExpr(node *ast.BinaryExpr, indent int)   {}
func (v *BaseEmitter) PreVisitBinaryExprLeft(node ast.Expr, indent int)       {}
func (v *BaseEmitter) PostVisitBinaryExprLeft(node ast.Expr, indent int)      {}
func (v *BaseEmitter) PreVisitBinaryExprRight(node ast.Expr, indent int)      {}
func (v *BaseEmitter) PostVisitBinaryExprRight(node ast.Expr, indent int)     {}
func (v *BaseEmitter) PreVisitBinaryExprOperator(op token.Token, indent int)  {}
func (v *BaseEmitter) PostVisitBinaryExprOperator(op token.Token, indent int) {}
func (v *BaseEmitter) PreVisitCallExpr(node *ast.CallExpr, indent int)        {}
func (v *BaseEmitter) PostVisitCallExpr(node *ast.CallExpr, indent int)       {}
func (v *BaseEmitter) PreVisitCallExprFun(node ast.Expr, indent int)          {}
func (v *BaseEmitter) PostVisitCallExprFun(node ast.Expr, indent int)         {}
func (v *BaseEmitter) PreVisitCallExprArgs(node *ast.CallExpr, indent int)    {}
func (v *BaseEmitter) PostVisitCallExprArgs(node *ast.CallExpr, indent int)   {}

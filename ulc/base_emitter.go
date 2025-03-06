package main

import (
	"go/ast"
	"go/token"
	"os"
)

type BaseEmitter struct{}

func (v *BaseEmitter) SetFile(file *os.File)                                                     {}
func (v *BaseEmitter) GetFile() *os.File                                                         { return nil }
func (v *BaseEmitter) PreVisitProgram(indent int)                                                {}
func (v *BaseEmitter) PostVisitProgram(indent int)                                               {}
func (v *BaseEmitter) PreVisitPackage(name string, indent int)                                   {}
func (v *BaseEmitter) PostVisitPackage(name string, indent int)                                  {}
func (v *BaseEmitter) PreVisitBasicLit(node *ast.BasicLit, indent int)                           {}
func (v *BaseEmitter) PostVisitBasicLit(node *ast.BasicLit, indent int)                          {}
func (v *BaseEmitter) PreVisitIdent(node *ast.Ident, indent int)                                 {}
func (v *BaseEmitter) PostVisitIdent(node *ast.Ident, indent int)                                {}
func (v *BaseEmitter) PreVisitBinaryExpr(node *ast.BinaryExpr, indent int)                       {}
func (v *BaseEmitter) PostVisitBinaryExpr(node *ast.BinaryExpr, indent int)                      {}
func (v *BaseEmitter) PreVisitBinaryExprLeft(node ast.Expr, indent int)                          {}
func (v *BaseEmitter) PostVisitBinaryExprLeft(node ast.Expr, indent int)                         {}
func (v *BaseEmitter) PreVisitBinaryExprRight(node ast.Expr, indent int)                         {}
func (v *BaseEmitter) PostVisitBinaryExprRight(node ast.Expr, indent int)                        {}
func (v *BaseEmitter) PreVisitBinaryExprOperator(op token.Token, indent int)                     {}
func (v *BaseEmitter) PostVisitBinaryExprOperator(op token.Token, indent int)                    {}
func (v *BaseEmitter) PreVisitCallExpr(node *ast.CallExpr, indent int)                           {}
func (v *BaseEmitter) PostVisitCallExpr(node *ast.CallExpr, indent int)                          {}
func (v *BaseEmitter) PreVisitCallExprFun(node ast.Expr, indent int)                             {}
func (v *BaseEmitter) PostVisitCallExprFun(node ast.Expr, indent int)                            {}
func (v *BaseEmitter) PreVisitCallExprArgs(node []ast.Expr, indent int)                          {}
func (v *BaseEmitter) PostVisitCallExprArgs(node []ast.Expr, indent int)                         {}
func (v *BaseEmitter) PreVisitCallExprArg(node ast.Expr, index int, indent int)                  {}
func (v *BaseEmitter) PostVisitCallExprArg(node ast.Expr, index int, indent int)                 {}
func (v *BaseEmitter) PreVisitParenExpr(node *ast.ParenExpr, indent int)                         {}
func (v *BaseEmitter) PostVisitParenExpr(node *ast.ParenExpr, indent int)                        {}
func (v *BaseEmitter) PreVisitCompositeLit(node *ast.CompositeLit, indent int)                   {}
func (v *BaseEmitter) PostVisitCompositeLit(node *ast.CompositeLit, indent int)                  {}
func (v *BaseEmitter) PreVisitCompositeLitType(node ast.Expr, indent int)                        {}
func (v *BaseEmitter) PostVisitCompositeLitType(node ast.Expr, indent int)                       {}
func (v *BaseEmitter) PreVisitCompositeLitElts(node []ast.Expr, indent int)                      {}
func (v *BaseEmitter) PostVisitCompositeLitElts(node []ast.Expr, indent int)                     {}
func (v *BaseEmitter) PreVisitCompositeLitElt(node ast.Expr, index int, indent int)              {}
func (v *BaseEmitter) PostVisitCompositeLitElt(node ast.Expr, index int, indent int)             {}
func (v *BaseEmitter) PreVisitArrayType(node ast.ArrayType, indent int)                          {}
func (v *BaseEmitter) PostVisitArrayType(node ast.ArrayType, indent int)                         {}
func (v *BaseEmitter) PreVisitSelectorExpr(node *ast.SelectorExpr, indent int)                   {}
func (v *BaseEmitter) PostVisitSelectorExpr(node *ast.SelectorExpr, indent int)                  {}
func (v *BaseEmitter) PreVisitSelectorExprX(node ast.Expr, indent int)                           {}
func (v *BaseEmitter) PostVisitSelectorExprX(node ast.Expr, indent int)                          {}
func (v *BaseEmitter) PreVisitSelectorExprSel(node *ast.Ident, indent int)                       {}
func (v *BaseEmitter) PostVisitSelectorExprSel(node *ast.Ident, indent int)                      {}
func (v *BaseEmitter) PreVisitIndexExpr(node *ast.IndexExpr, indent int)                         {}
func (v *BaseEmitter) PostVisitIndexExpr(node *ast.IndexExpr, indent int)                        {}
func (v *BaseEmitter) PreVisitIndexExprX(node *ast.IndexExpr, indent int)                        {}
func (v *BaseEmitter) PostVisitIndexExprX(node *ast.IndexExpr, indent int)                       {}
func (v *BaseEmitter) PreVisitIndexExprIndex(node *ast.IndexExpr, indent int)                    {}
func (v *BaseEmitter) PostVisitIndexExprIndex(node *ast.IndexExpr, indent int)                   {}
func (v *BaseEmitter) PreVisitUnaryExpr(node *ast.UnaryExpr, indent int)                         {}
func (v *BaseEmitter) PostVisitUnaryExpr(node *ast.UnaryExpr, indent int)                        {}
func (v *BaseEmitter) PreVisitSliceExpr(node *ast.SliceExpr, indent int)                         {}
func (v *BaseEmitter) PostVisitSliceExpr(node *ast.SliceExpr, indent int)                        {}
func (v *BaseEmitter) PreVisitSliceExprX(node ast.Expr, indent int)                              {}
func (v *BaseEmitter) PostVisitSliceExprX(node ast.Expr, indent int)                             {}
func (v *BaseEmitter) PreVisitSliceExprXBegin(node ast.Expr, indent int)                         {}
func (v *BaseEmitter) PostVisitSliceExprXBegin(node ast.Expr, indent int)                        {}
func (v *BaseEmitter) PreVisitSliceExprXEnd(node ast.Expr, indent int)                           {}
func (v *BaseEmitter) PostVisitSliceExprXEnd(node ast.Expr, indent int)                          {}
func (v *BaseEmitter) PreVisitSliceExprLow(node ast.Expr, indent int)                            {}
func (v *BaseEmitter) PostVisitSliceExprLow(node ast.Expr, indent int)                           {}
func (v *BaseEmitter) PreVisitSliceExprHigh(node ast.Expr, indent int)                           {}
func (v *BaseEmitter) PostVisitSliceExprHigh(node ast.Expr, indent int)                          {}
func (v *BaseEmitter) PreVisitFuncType(node *ast.FuncType, indent int)                           {}
func (v *BaseEmitter) PostVisitFuncType(node *ast.FuncType, indent int)                          {}
func (v *BaseEmitter) PreVisitFuncTypeResults(node *ast.FieldList, indent int)                   {}
func (v *BaseEmitter) PostVisitFuncTypeResults(node *ast.FieldList, indent int)                  {}
func (v *BaseEmitter) PreVisitFuncTypeResult(node *ast.Field, index int, indent int)             {}
func (v *BaseEmitter) PostVisitFuncTypeResult(node *ast.Field, index int, indent int)            {}
func (v *BaseEmitter) PreVisitFuncTypeParams(node *ast.FieldList, indent int)                    {}
func (v *BaseEmitter) PostVisitFuncTypeParams(node *ast.FieldList, indent int)                   {}
func (v *BaseEmitter) PreVisitFuncTypeParam(node *ast.Field, index int, indent int)              {}
func (v *BaseEmitter) PostVisitFuncTypeParam(node *ast.Field, index int, indent int)             {}
func (v *BaseEmitter) PreVisitKeyValueExpr(node *ast.KeyValueExpr, indent int)                   {}
func (v *BaseEmitter) PostVisitKeyValueExpr(node *ast.KeyValueExpr, indent int)                  {}
func (v *BaseEmitter) PreVisitKeyValueExprKey(node ast.Expr, indent int)                         {}
func (v *BaseEmitter) PostVisitKeyValueExprKey(node ast.Expr, indent int)                        {}
func (v *BaseEmitter) PreVisitKeyValueExprValue(node ast.Expr, indent int)                       {}
func (v *BaseEmitter) PostVisitKeyValueExprValue(node ast.Expr, indent int)                      {}
func (v *BaseEmitter) PreVisitFuncLit(node *ast.FuncLit, indent int)                             {}
func (v *BaseEmitter) PostVisitFuncLit(node *ast.FuncLit, indent int)                            {}
func (v *BaseEmitter) PreVisitFuncLitTypeParams(node *ast.FieldList, indent int)                 {}
func (v *BaseEmitter) PostVisitFuncLitTypeParams(node *ast.FieldList, indent int)                {}
func (v *BaseEmitter) PreVisitFuncLitTypeParam(node *ast.Field, index int, indent int)           {}
func (v *BaseEmitter) PostVisitFuncLitTypeParam(node *ast.Field, index int, indent int)          {}
func (v *BaseEmitter) PreVisitFuncLitBody(node *ast.BlockStmt, indent int)                       {}
func (v *BaseEmitter) PostVisitFuncLitBody(node *ast.BlockStmt, indent int)                      {}
func (v *BaseEmitter) PreVisitFuncLitTypeResults(node *ast.FieldList, indent int)                {}
func (v *BaseEmitter) PostVisitFuncLitTypeResults(node *ast.FieldList, indent int)               {}
func (v *BaseEmitter) PreVisitFuncLitTypeResult(node *ast.Field, index int, indent int)          {}
func (v *BaseEmitter) PostVisitFuncLitTypeResult(node *ast.Field, index int, indent int)         {}
func (v *BaseEmitter) PreVisitTypeAssertExpr(node *ast.TypeAssertExpr, indent int)               {}
func (v *BaseEmitter) PostVisitTypeAssertExpr(node *ast.TypeAssertExpr, indent int)              {}
func (v *BaseEmitter) PreVisitTypeAssertExprX(node ast.Expr, indent int)                         {}
func (v *BaseEmitter) PostVisitTypeAssertExprX(node ast.Expr, indent int)                        {}
func (v *BaseEmitter) PreVisitTypeAssertExprType(node ast.Expr, indent int)                      {}
func (v *BaseEmitter) PostVisitTypeAssertExprType(node ast.Expr, indent int)                     {}
func (v *BaseEmitter) PreVisitStarExpr(node *ast.StarExpr, indent int)                           {}
func (v *BaseEmitter) PostVisitStarExpr(node *ast.StarExpr, indent int)                          {}
func (v *BaseEmitter) PreVisitStarExprX(node ast.Expr, indent int)                               {}
func (v *BaseEmitter) PostVisitStarExprX(node ast.Expr, indent int)                              {}
func (v *BaseEmitter) PreVisitInterfaceType(node *ast.InterfaceType, indent int)                 {}
func (v *BaseEmitter) PostVisitInterfaceType(node *ast.InterfaceType, indent int)                {}
func (v *BaseEmitter) PreVisitExprStmt(node *ast.ExprStmt, indent int)                           {}
func (v *BaseEmitter) PostVisitExprStmt(node *ast.ExprStmt, indent int)                          {}
func (v *BaseEmitter) PreVisitExprStmtX(node ast.Expr, indent int)                               {}
func (v *BaseEmitter) PostVisitExprStmtX(node ast.Expr, indent int)                              {}
func (v *BaseEmitter) PreVisitDeclStmt(node *ast.DeclStmt, indent int)                           {}
func (v *BaseEmitter) PostVisitDeclStmt(node *ast.DeclStmt, indent int)                          {}
func (v *BaseEmitter) PreVisitDeclStmtValueSpecType(node *ast.ValueSpec, index int, indent int)  {}
func (v *BaseEmitter) PostVisitDeclStmtValueSpecType(node *ast.ValueSpec, index int, indent int) {}
func (v *BaseEmitter) PreVisitDeclStmtValueSpecNames(node *ast.Ident, index int, indent int)     {}
func (v *BaseEmitter) PostVisitDeclStmtValueSpecNames(node *ast.Ident, index int, indent int)    {}
func (v *BaseEmitter) PreVisitBranchStmt(node *ast.BranchStmt, indent int)                       {}
func (v *BaseEmitter) PostVisitBranchStmt(node *ast.BranchStmt, indent int)                      {}
func (v *BaseEmitter) PreVisitIncDecStmt(node *ast.IncDecStmt, indent int)                       {}
func (v *BaseEmitter) PostVisitIncDecStmt(node *ast.IncDecStmt, indent int)                      {}
func (v *BaseEmitter) PreVisitAssignStmt(node *ast.AssignStmt, indent int)                       {}
func (v *BaseEmitter) PostVisitAssignStmt(node *ast.AssignStmt, indent int)                      {}
func (v *BaseEmitter) PostVisitForStmt(node *ast.ForStmt, indent int)                            {}
func (v *BaseEmitter) PreVisitForStmt(node *ast.ForStmt, indent int)                             {}
func (v *BaseEmitter) PreVisitForStmtInit(node ast.Stmt, indent int)                             {}
func (v *BaseEmitter) PostVisitForStmtInit(node ast.Stmt, indent int)                            {}
func (v *BaseEmitter) PreVisitForStmtCond(node ast.Expr, indent int)                             {}
func (v *BaseEmitter) PostVisitForStmtCond(node ast.Expr, indent int)                            {}
func (v *BaseEmitter) PreVisitForStmtPost(node ast.Stmt, indent int)                             {}
func (v *BaseEmitter) PostVisitForStmtPost(node ast.Stmt, indent int)                            {}
func (v *BaseEmitter) PreVisitAssignStmtLhs(node *ast.AssignStmt, indent int)                    {}
func (v *BaseEmitter) PostVisitAssignStmtLhs(node *ast.AssignStmt, indent int)                   {}
func (v *BaseEmitter) PreVisitAssignStmtRhs(node *ast.AssignStmt, indent int)                    {}
func (v *BaseEmitter) PostVisitAssignStmtRhs(node *ast.AssignStmt, indent int)                   {}
func (v *BaseEmitter) PreVisitAssignStmtLhsExpr(node ast.Expr, index int, indent int)            {}
func (v *BaseEmitter) PostVisitAssignStmtLhsExpr(node ast.Expr, index int, indent int)           {}
func (v *BaseEmitter) PreVisitAssignStmtRhsExpr(node ast.Expr, index int, indent int)            {}
func (v *BaseEmitter) PostVisitAssignStmtRhsExpr(node ast.Expr, index int, indent int)           {}
func (v *BaseEmitter) PreVisitReturnStmt(node *ast.ReturnStmt, indent int)                       {}
func (v *BaseEmitter) PostVisitReturnStmt(node *ast.ReturnStmt, indent int)                      {}
func (v *BaseEmitter) PreVisitReturnStmtResult(node ast.Expr, index int, indent int)             {}
func (v *BaseEmitter) PostVisitReturnStmtResult(node ast.Expr, index int, indent int)            {}
func (v *BaseEmitter) PreVisitIfStmt(node *ast.IfStmt, indent int)                               {}
func (v *BaseEmitter) PostVisitIfStmt(node *ast.IfStmt, indent int)                              {}
func (v *BaseEmitter) PreVisitIfStmtCond(node *ast.IfStmt, indent int)                           {}
func (v *BaseEmitter) PostVisitIfStmtCond(node *ast.IfStmt, indent int)                          {}
func (v *BaseEmitter) PreVisitIfStmtBody(node *ast.IfStmt, indent int)                           {}
func (v *BaseEmitter) PostVisitIfStmtBody(node *ast.IfStmt, indent int)                          {}
func (v *BaseEmitter) PreVisitIfStmtElse(node *ast.IfStmt, indent int)                           {}
func (v *BaseEmitter) PostVisitIfStmtElse(node *ast.IfStmt, indent int)                          {}

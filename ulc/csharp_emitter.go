package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"os"
	"strings"
)

type CSharpEmitter struct {
	file *os.File
	Emitter
	insideForPostCond bool
	assignmentToken   string
}

func (*CSharpEmitter) lowerToBuiltins(selector string) string {
	switch selector {
	case "fmt":
		return ""
	case "Sprintf":
		return "string_format"
	case "Println":
		return "println"
	case "Printf":
		return "printf"
	case "Print":
		return "printf"
	case "len":
		return "std::size"
	}
	return selector
}

func (e *CSharpEmitter) emitToFile(s string) error {
	_, err := e.file.WriteString(s)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return err
	}
	return nil
}

func (e *CSharpEmitter) emitAsString(s string, indent int) string {
	return strings.Repeat(" ", indent) + s
}

func (cppe *CSharpEmitter) SetFile(file *os.File) {
	cppe.file = file
}

func (cppe *CSharpEmitter) GetFile() *os.File {
	return cppe.file
}

func (cppe *CSharpEmitter) PreVisitProgram(indent int) {
	outputFile := "./output.cpp"
	var err error
	cppe.file, err = os.Create(outputFile)
	cppe.SetFile(cppe.file)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	_, err = cppe.file.WriteString("#include <vector>\n" +
		"#include <string>\n" +
		"#include <tuple>\n" +
		"#include <any>\n" +
		"#include <cstdint>\n" +
		"#include <functional>\n" +
		"#include \"../builtins/builtins.h\"\n\n")
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
	cppe.insideForPostCond = false
}

func (cppe *CSharpEmitter) PostVisitProgram(indent int) {
	cppe.file.Close()
}

func (cppe *CSharpEmitter) PreVisitPackage(name string, indent int) {
	if name == "main" {
		return
	}
	str := cppe.emitAsString(fmt.Sprintf("namespace %s\n", name), 0)
	err := cppe.emitToFile(str)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
	err = cppe.emitToFile("{\n\n")
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}

func (cppe *CSharpEmitter) PostVisitPackage(name string, indent int) {
	if name == "main" {
		return
	}
	str := cppe.emitAsString(fmt.Sprintf("} // namespace %s\n\n", name), 0)
	err := cppe.emitToFile(str)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}

func (cppe *CSharpEmitter) PreVisitBasicLit(e *ast.BasicLit, indent int) {
	if e.Kind == token.STRING {
		e.Value = strings.Replace(e.Value, "\"", "", -1)
		if e.Value[0] == '`' {
			e.Value = strings.Replace(e.Value, "`", "", -1)
			cppe.emitToFile(cppe.emitAsString(fmt.Sprintf("R\"(%s)\"", e.Value), 0))
		} else {
			cppe.emitToFile(cppe.emitAsString(fmt.Sprintf("\"%s\"", e.Value), 0))
		}
	} else {
		cppe.emitToFile(cppe.emitAsString(e.Value, 0))
	}
}

func (cppe *CSharpEmitter) PreVisitIdent(e *ast.Ident, indent int) {
	var str string
	name := e.Name
	name = cppe.lowerToBuiltins(name)
	if name == "nil" {
		str = cppe.emitAsString("{}", indent)
		cppe.emitToFile(str)
	} else {
		if n, ok := typesMap[name]; ok {
			str = cppe.emitAsString(n, indent)
			cppe.emitToFile(str)
		} else {
			str = cppe.emitAsString(name, indent)
			cppe.emitToFile(str)
		}
	}
}

func (cppe *CSharpEmitter) PreVisitBinaryExprOperator(op token.Token, indent int) {
	str := cppe.emitAsString(op.String()+" ", 1)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PreVisitCallExprArgs(node []ast.Expr, indent int) {
	str := cppe.emitAsString("(", 0)
	cppe.emitToFile(str)
}
func (cppe *CSharpEmitter) PostVisitCallExprArgs(node []ast.Expr, indent int) {
	str := cppe.emitAsString(")", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PreVisitCallExprArg(node ast.Expr, index int, indent int) {
	if index > 0 {
		str := cppe.emitAsString(", ", 0)
		cppe.emitToFile(str)
	}
}

func (cppe *CSharpEmitter) PreVisitParenExpr(node *ast.ParenExpr, indent int) {
	str := cppe.emitAsString("(", 0)
	cppe.emitToFile(str)
}
func (cppe *CSharpEmitter) PostVisitParenExpr(node *ast.ParenExpr, indent int) {
	str := cppe.emitAsString(")", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PreVisitCompositeLitElts(node []ast.Expr, indent int) {
	str := cppe.emitAsString("{", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PostVisitCompositeLitElts(node []ast.Expr, indent int) {
	str := cppe.emitAsString("}", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PreVisitCompositeLitElt(node ast.Expr, index int, indent int) {
	if index > 0 {
		str := cppe.emitAsString(", ", 0)
		cppe.emitToFile(str)
	}
}

func (cppe *CSharpEmitter) PreVisitArrayType(node ast.ArrayType, indent int) {
	str := cppe.emitAsString("std::vector<", indent)
	cppe.emitToFile(str)
}
func (cppe *CSharpEmitter) PostVisitArrayType(node ast.ArrayType, indent int) {
	str := cppe.emitAsString(">", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PostVisitSelectorExprX(node ast.Expr, indent int) {
	if ident, ok := node.(*ast.Ident); ok {
		if cppe.lowerToBuiltins(ident.Name) == "" {
			return
		}
		scopeOperator := "."
		if _, found := namespaces[ident.Name]; found {
			scopeOperator = "::"
		}
		str := cppe.emitAsString(scopeOperator, 0)
		cppe.emitToFile(str)
	} else {
		str := cppe.emitAsString(".", 0)
		cppe.emitToFile(str)
	}
}

func (cppe *CSharpEmitter) PreVisitIndexExprIndex(node *ast.IndexExpr, indent int) {
	str := cppe.emitAsString("[", 0)
	cppe.emitToFile(str)
}
func (cppe *CSharpEmitter) PostVisitIndexExprIndex(node *ast.IndexExpr, indent int) {
	str := cppe.emitAsString("]", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PreVisitUnaryExpr(node *ast.UnaryExpr, indent int) {
	str := cppe.emitAsString("(", 0)
	str += cppe.emitAsString(node.Op.String(), 0)
	cppe.emitToFile(str)
}
func (cppe *CSharpEmitter) PostVisitUnaryExpr(node *ast.UnaryExpr, indent int) {
	str := cppe.emitAsString(")", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PreVisitSliceExpr(node *ast.SliceExpr, indent int) {
	str := cppe.emitAsString("std::vector<std::remove_reference<decltype(", indent)
	cppe.emitToFile(str)
}
func (cppe *CSharpEmitter) PostVisitSliceExpr(node *ast.SliceExpr, indent int) {
	str := cppe.emitAsString(")", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PostVisitSliceExprX(node ast.Expr, indent int) {
	str := cppe.emitAsString("[0]", 0)
	str += cppe.emitAsString(")>::type>(", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PreVisitSliceExprLow(node ast.Expr, indent int) {
	str := cppe.emitAsString(".begin() ", 0)
	cppe.emitToFile(str)
	if node != nil {
		str := cppe.emitAsString("+ ", 0)
		cppe.emitToFile(str)
	} else {
		log.Println("Low index: <nil>")
	}
}

func (cppe *CSharpEmitter) PreVisitSliceExprXEnd(node ast.Expr, indent int) {
	str := cppe.emitAsString(", ", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PreVisitSliceExprHigh(node ast.Expr, indent int) {
	if node != nil {
		str := cppe.emitAsString(".begin() ", 0)
		cppe.emitToFile(str)
		str = cppe.emitAsString("+ ", 0)
		cppe.emitToFile(str)
	} else {
		str := cppe.emitAsString(".end() ", 0)
		cppe.emitToFile(str)
	}
}

func (cppe *CSharpEmitter) PostVisitSliceExprHigh(node ast.Expr, indent int) {

}

func (cppe *CSharpEmitter) PreVisitFuncType(node *ast.FuncType, indent int) {
	str := cppe.emitAsString("std::function<", indent)
	cppe.emitToFile(str)
}
func (cppe *CSharpEmitter) PostVisitFuncType(node *ast.FuncType, indent int) {
	str := cppe.emitAsString(">", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PreVisitFuncTypeResults(node *ast.FieldList, indent int) {
	if node == nil {
		str := cppe.emitAsString("void", 0)
		cppe.emitToFile(str)
	}
}

func (cppe *CSharpEmitter) PreVisitFuncTypeResult(node *ast.Field, index int, indent int) {
	if index > 0 {
		str := cppe.emitAsString(", ", 0)
		cppe.emitToFile(str)
	}
}

func (cppe *CSharpEmitter) PreVisitKeyValueExpr(node *ast.KeyValueExpr, indent int) {
	str := cppe.emitAsString(".", 0)
	cppe.emitToFile(str)
}
func (cppe *CSharpEmitter) PostVisitKeyValueExpr(node *ast.KeyValueExpr, indent int) {
	str := cppe.emitAsString("\n", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PreVisitFuncTypeParams(node *ast.FieldList, indent int) {
	str := cppe.emitAsString("(", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PostVisitFuncTypeParams(node *ast.FieldList, indent int) {
	str := cppe.emitAsString(")", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PreVisitFuncTypeParam(node *ast.Field, index int, indent int) {
	if index > 0 {
		str := cppe.emitAsString(", ", 0)
		cppe.emitToFile(str)
	}
}

func (cppe *CSharpEmitter) PreVisitKeyValueExprValue(node ast.Expr, indent int) {
	str := cppe.emitAsString("= ", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PreVisitFuncLit(node *ast.FuncLit, indent int) {
	str := cppe.emitAsString("[&](", indent)
	cppe.emitToFile(str)
}
func (cppe *CSharpEmitter) PostVisitFuncLit(node *ast.FuncLit, indent int) {
	str := cppe.emitAsString("}", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PostVisitFuncLitTypeParams(node *ast.FieldList, indent int) {
	str := cppe.emitAsString(")", 0)
	str += cppe.emitAsString("->", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PreVisitFuncLitTypeParam(node *ast.Field, index int, indent int) {
	str := ""
	if index > 0 {
		str += cppe.emitAsString(", ", 0)
	}
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PostVisitFuncLitTypeParam(node *ast.Field, index int, indent int) {
	str := cppe.emitAsString(" ", 0)
	str += cppe.emitAsString(node.Names[0].Name, indent)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PreVisitFuncLitBody(node *ast.BlockStmt, indent int) {
	str := cppe.emitAsString("{\n", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PreVisitFuncLitTypeResults(node *ast.FieldList, indent int) {
	if node == nil {
		str := cppe.emitAsString("void", 0)
		cppe.emitToFile(str)
	}
}

func (cppe *CSharpEmitter) PreVisitFuncLitTypeResult(node *ast.Field, index int, indent int) {
	if index > 0 {
		str := cppe.emitAsString(", ", 0)
		cppe.emitToFile(str)
	}
}

func (cppe *CSharpEmitter) PreVisitTypeAssertExprType(node ast.Expr, indent int) {
	str := cppe.emitAsString("std::any_cast<", indent)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PreVisitTypeAssertExprX(node ast.Expr, indent int) {
	str := cppe.emitAsString(">(std::any(", 0)
	cppe.emitToFile(str)
}
func (cppe *CSharpEmitter) PostVisitTypeAssertExprX(node ast.Expr, indent int) {
	str := cppe.emitAsString("))", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PreVisitStarExpr(node *ast.StarExpr, indent int) {
	str := cppe.emitAsString("*", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PreVisitInterfaceType(node *ast.InterfaceType, indent int) {
	str := cppe.emitAsString("std::any", indent)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PostVisitExprStmtX(node ast.Expr, indent int) {
	str := cppe.emitAsString(";", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PreVisitDeclStmtValueSpecNames(node *ast.Ident, index int, indent int) {
	str := cppe.emitAsString(" ", 0)
	cppe.emitToFile(str)
}
func (cppe *CSharpEmitter) PostVisitDeclStmtValueSpecNames(node *ast.Ident, index int, indent int) {
	cppe.emitToFile(";")
}

func (cppe *CSharpEmitter) PreVisitBranchStmt(node *ast.BranchStmt, indent int) {
	str := cppe.emitAsString(node.Tok.String()+";", indent)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PostVisitIncDecStmt(node *ast.IncDecStmt, indent int) {
	str := cppe.emitAsString(node.Tok.String(), 0)
	if !cppe.insideForPostCond {
		str += cppe.emitAsString(";", 0)
	}
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PreVisitForStmtPost(node ast.Stmt, indent int) {
	if node != nil {
		cppe.insideForPostCond = true
	}
}
func (cppe *CSharpEmitter) PostVisitForStmtPost(node ast.Stmt, indent int) {
	if node != nil {
		cppe.insideForPostCond = false
	}
	str := cppe.emitAsString(")\n", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PreVisitAssignStmt(node *ast.AssignStmt, indent int) {
	str := cppe.emitAsString("", indent)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PostVisitAssignStmt(node *ast.AssignStmt, indent int) {
	str := cppe.emitAsString(";", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PreVisitAssignStmtLhs(node *ast.AssignStmt, indent int) {
	assignmentToken := node.Tok.String()
	if assignmentToken == ":=" && len(node.Lhs) == 1 {
		str := cppe.emitAsString("auto ", indent)
		cppe.emitToFile(str)
	} else if assignmentToken == ":=" && len(node.Lhs) > 1 {
		str := cppe.emitAsString("auto [", indent)
		cppe.emitToFile(str)
	} else if assignmentToken == "=" && len(node.Lhs) > 1 {
		str := cppe.emitAsString("std::tie(", indent)
		cppe.emitToFile(str)
	}
	if assignmentToken != "+=" {
		assignmentToken = "="
	}
	cppe.assignmentToken = assignmentToken
}

func (cppe *CSharpEmitter) PostVisitAssignStmtLhs(node *ast.AssignStmt, indent int) {
	if node.Tok.String() == ":=" && len(node.Lhs) > 1 {
		str := cppe.emitAsString("]", indent)
		cppe.emitToFile(str)
	} else if node.Tok.String() == "=" && len(node.Lhs) > 1 {
		str := cppe.emitAsString(")", indent)
		cppe.emitToFile(str)
	}
}

func (cppe *CSharpEmitter) PreVisitAssignStmtRhs(node *ast.AssignStmt, indent int) {
	str := cppe.emitAsString(cppe.assignmentToken+" ", indent+1)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PreVisitAssignStmtLhsExpr(node ast.Expr, index int, indent int) {
	if index > 0 {
		str := cppe.emitAsString(", ", indent)
		cppe.emitToFile(str)
	}
}

func (cppe *CSharpEmitter) PreVisitReturnStmt(node *ast.ReturnStmt, indent int) {
	str := cppe.emitAsString("return ", indent)
	cppe.emitToFile(str)
	if len(node.Results) > 1 {
		str := cppe.emitAsString("std::make_tuple(", 0)
		cppe.emitToFile(str)
	}
}

func (cppe *CSharpEmitter) PostVisitReturnStmt(node *ast.ReturnStmt, indent int) {
	if len(node.Results) > 1 {
		str := cppe.emitAsString(")", 0)
		cppe.emitToFile(str)
	}
	str := cppe.emitAsString(";", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PreVisitReturnStmtResult(node ast.Expr, index int, indent int) {
	if index > 0 {
		str := cppe.emitAsString(", ", 0)
		cppe.emitToFile(str)
	}
}

func (cppe *CSharpEmitter) PreVisitIfStmtCond(node *ast.IfStmt, indent int) {
	str := cppe.emitAsString("if (", indent)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PostVisitIfStmtCond(node *ast.IfStmt, indent int) {
	str := cppe.emitAsString(")\n", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PreVisitIfStmtElse(node *ast.IfStmt, indent int) {
	str := cppe.emitAsString("else", 1)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PreVisitForStmt(node *ast.ForStmt, indent int) {
	str := cppe.emitAsString("for (", indent)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PostVisitForStmtInit(node ast.Stmt, indent int) {
	if node == nil {
		str := cppe.emitAsString(";", 0)
		cppe.emitToFile(str)
	}
}

func (cppe *CSharpEmitter) PostVisitForStmtCond(node ast.Expr, indent int) {
	str := cppe.emitAsString(";", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PreVisitRangeStmt(node *ast.RangeStmt, indent int) {
	str := cppe.emitAsString("for (auto ", indent)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PostVisitRangeStmtValue(node ast.Expr, indent int) {
	str := cppe.emitAsString(" : ", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PostVisitRangeStmtX(node ast.Expr, indent int) {
	str := cppe.emitAsString(")\n", 0)
	cppe.emitToFile(str)
}
func (cppe *CSharpEmitter) PreVisitSwitchStmt(node *ast.SwitchStmt, indent int) {
	str := cppe.emitAsString("switch (", indent)
	cppe.emitToFile(str)
}
func (cppe *CSharpEmitter) PostVisitSwitchStmt(node *ast.SwitchStmt, indent int) {
	str := cppe.emitAsString("}", indent)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PostVisitSwitchStmtTag(node ast.Expr, indent int) {
	str := cppe.emitAsString(") {\n", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PostVisitCaseClause(node *ast.CaseClause, indent int) {
	cppe.emitToFile("\n")
	str := cppe.emitAsString("break;\n", indent+4)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PostVisitCaseClauseList(node []ast.Expr, indent int) {
	if len(node) == 0 {
		str := cppe.emitAsString("default:\n", indent+2)
		cppe.emitToFile(str)
	}
}

func (cppe *CSharpEmitter) PreVisitCaseClauseListExpr(node ast.Expr, index int, indent int) {
	str := cppe.emitAsString("case ", indent+2)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PostVisitCaseClauseListExpr(node ast.Expr, index int, indent int) {
	str := cppe.emitAsString(":\n", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PreVisitBlockStmt(node *ast.BlockStmt, indent int) {
	str := cppe.emitAsString("{\n", indent)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PostVisitBlockStmt(node *ast.BlockStmt, indent int) {
	str := cppe.emitAsString("}", indent)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PostVisitBlockStmtList(node ast.Stmt, index int, indent int) {
	str := cppe.emitAsString("\n", indent)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PostVisitFuncDecl(node *ast.FuncDecl, indent int) {
	str := cppe.emitAsString("\n\n", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PreVisitFuncDeclBody(node *ast.BlockStmt, indent int) {
	str := cppe.emitAsString("\n", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PreVisitFuncDeclSignatureTypeResults(node *ast.FuncDecl, indent int) {
	if node.Type.Results != nil {
		if len(node.Type.Results.List) > 1 {
			str := cppe.emitAsString("std::tuple<", 0)
			err := cppe.emitToFile(str)
			if err != nil {
				fmt.Println("Error writing to file:", err)
				return
			}
		}
	}
}

func (cppe *CSharpEmitter) PostVisitFuncDeclSignatureTypeParams(node *ast.FuncDecl, indent int) {
	str := cppe.emitAsString(")", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PostVisitFuncDeclSignatureTypeResults(node *ast.FuncDecl, indent int) {
	if node.Type.Results != nil {
		if len(node.Type.Results.List) > 1 {
			str := cppe.emitAsString(">", 0)
			cppe.emitToFile(str)
		}
	} else if node.Name.Name == "main" {
		str := cppe.emitAsString("int", 0)
		cppe.emitToFile(str)
	} else {
		str := cppe.emitAsString("void", 0)
		cppe.emitToFile(str)
	}
	str := cppe.emitAsString("", 1)
	cppe.emitToFile(str)
	str = cppe.emitAsString(node.Name.Name+"(", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PreVisitFuncDeclSignatureTypeResultsList(node *ast.Field, index int, indent int) {
	if index > 0 {
		str := cppe.emitAsString(",", 0)
		cppe.emitToFile(str)
	}
}
func (cppe *CSharpEmitter) PreVisitFuncDeclSignatureTypeParamsList(node *ast.Field, index int, indent int) {
	if index > 0 {
		str := cppe.emitAsString(", ", 0)
		cppe.emitToFile(str)
	}
}

func (cppe *CSharpEmitter) PreVisitFuncDeclSignatureTypeParamsArgName(node *ast.Ident, index int, indent int) {
	cppe.emitToFile(" ")
}

func (cppe *CSharpEmitter) PostVisitFuncDeclSignature(node *ast.FuncDecl, indent int) {
	str := cppe.emitAsString(";\n", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PreVisitGenStructInfo(node GenStructInfo, indent int) {
	str := cppe.emitAsString(fmt.Sprintf("struct %s\n", node.Name), 0)
	err := cppe.emitToFile(str)
	if err != nil {
		fmt.Println("Error writing to file:", err)
	}
	str = cppe.emitAsString("{\n", 0)
	err = cppe.emitToFile(str)
	if err != nil {
		fmt.Println("Error writing to file:", err)
	}
}
func (cppe *CSharpEmitter) PostVisitGenStructInfo(node GenStructInfo, indent int) {
	str := cppe.emitAsString("};\n\n", 0)
	err := cppe.emitToFile(str)
	if err != nil {
		fmt.Println("Error writing to file:", err)
	}
}

func (cppe *CSharpEmitter) PreVisitFuncDeclSignatures(indent int) {
	// Generate forward function declarations
	str := cppe.emitAsString("// Forward declarations\n", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PostVisitFuncDeclSignatures(indent int) {
	str := cppe.emitAsString("\n", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PostVisitGenDeclConst(node *ast.GenDecl, indent int) {
	str := cppe.emitAsString("\n", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PostVisitGenStructFieldType(node ast.Expr, indent int) {
	cppe.emitToFile(" ")
}

func (cppe *CSharpEmitter) PostVisitGenStructFieldName(node *ast.Ident, indent int) {
	cppe.emitToFile(";\n")
}

func (cppe *CSharpEmitter) PreVisitGenDeclConstName(node *ast.Ident, indent int) {
	str := cppe.emitAsString(fmt.Sprintf("constexpr auto %s = ", node.Name), 0)
	cppe.emitToFile(str)
}
func (cppe *CSharpEmitter) PostVisitGenDeclConstName(node *ast.Ident, indent int) {
	str := cppe.emitAsString(";\n", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PreVisitTypeAliasName(node *ast.Ident, indent int) {
	cppe.emitToFile(fmt.Sprintf("using "))
}

func (cppe *CSharpEmitter) PostVisitTypeAliasName(node *ast.Ident, indent int) {
	cppe.emitToFile(" = ")
}

func (cppe *CSharpEmitter) PostVisitTypeAliasType(node ast.Expr, indent int) {
	cppe.emitToFile(";\n\n")
}

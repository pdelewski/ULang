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

func (v *CSharpEmitter) SetFile(file *os.File) {
	v.file = file
}
func (v *CSharpEmitter) GetFile() *os.File {
	return v.file
}

func (v *CSharpEmitter) PreVisitProgram(indent int) {
	outputFile := "./output.cs"
	var err error
	v.file, err = os.Create(outputFile)
	v.SetFile(v.file)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	_, err = v.file.WriteString("using System;\n")
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}

func (v *CSharpEmitter) PostVisitProgram(indent int) {
	v.file.Close()
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

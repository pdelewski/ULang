package main

import (
	"fmt"
	"go/ast"
	"os"
	"strings"
	"unicode"
)

var csTypesMap = map[string]string{
	"int8":   "sbyte",
	"int16":  "short",
	"int32":  "int",
	"int64":  "long",
	"uint8":  "byte",
	"uint16": "ushort",
	"any":    "object",
	"string": "string",
}

type CSharpEmitter struct {
	file *os.File
	Emitter
	insideForPostCond bool
	assignmentToken   string
	forwardDecls      bool
	insideStruct      bool
	numFuncResults    int
	aliases           []string
	isAlias           bool
	currentPackage    string
	stack             []string
	buffer            bool
}

func (v *CSharpEmitter) mergeStackElements(marker string) {
	var merged strings.Builder

	// Process the stack in reverse until we find a marker
	for len(v.stack) > 0 {
		top := v.stack[len(v.stack)-1]
		v.stack = v.stack[:len(v.stack)-1] // Pop element

		// Stop merging when we find a marker
		if strings.HasPrefix(top, marker) {
			v.stack = append(v.stack, merged.String()) // Push merged string
			return
		}

		// Prepend the element to the merged string (reverse order)
		mergedString := top + merged.String() // Prepend instead of append
		merged.Reset()
		merged.WriteString(mergedString)
	}

	panic("unreachable")
}

func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s // Return empty string if input is empty
	}

	// Convert string to rune slice to handle Unicode characters
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0]) // Change the first character to uppercase

	return string(runes) // Convert runes back to string
}

func (*CSharpEmitter) lowerToBuiltins(selector string) string {
	switch selector {
	case "fmt":
		return ""
	case "Sprintf":
		return "string.format"
	case "Println":
		return "Console.WriteLine"
	case "Printf":
		return "Console.Write"
	case "Print":
		return "Console.Write"
	case "len":
		return "Length"
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
	outputFile := "./output.cs"
	var err error
	cppe.file, err = os.Create(outputFile)
	cppe.SetFile(cppe.file)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	_, err = cppe.file.WriteString("using System;\n\n")
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
	cppe.insideForPostCond = false
}

func (cppe *CSharpEmitter) PostVisitProgram(indent int) {
	cppe.file.Close()
}

func (cppe *CSharpEmitter) PreVisitDeclStmtValueSpecNames(node *ast.Ident, index int, indent int) {
	str := cppe.emitAsString(" ", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PostVisitDeclStmtValueSpecNames(node *ast.Ident, index int, indent int) {
	cppe.emitToFile(";")
}

func (cppe *CSharpEmitter) PostVisitGenStructFieldType(node ast.Expr, indent int) {
	cppe.emitToFile(" ")
}

func (cppe *CSharpEmitter) PostVisitGenStructFieldName(node *ast.Ident, indent int) {
	cppe.emitToFile(";\n")
}

func (cppe *CSharpEmitter) PreVisitIdent(e *ast.Ident, indent int) {
	if !cppe.insideStruct {
		return
	}
	var str string
	name := e.Name
	name = cppe.lowerToBuiltins(name)
	if name == "nil" {
		str = cppe.emitAsString("{}", indent)
	} else {
		if n, ok := csTypesMap[name]; ok {
			str = cppe.emitAsString(n, indent)
		} else {
			str = cppe.emitAsString(name, indent)
		}
	}

	if cppe.buffer {
		cppe.stack = append(cppe.stack, str)
	} else {
		cppe.emitToFile(str)
	}

}

func (cppe *CSharpEmitter) PreVisitPackage(name string, indent int) {
	var packageName string
	if name == "main" {
		packageName = "MainClass"
	} else {
		//packageName = capitalizeFirst(name)
		packageName = name
	}

	str := cppe.emitAsString(fmt.Sprintf("namespace %s {\n\n", packageName), indent)
	err := cppe.emitToFile(str)
	for _, alias := range cppe.aliases {
		str := cppe.emitAsString(alias, indent+2)
		cppe.emitToFile(str)
	}
	cppe.currentPackage = packageName
	str = cppe.emitAsString(fmt.Sprintf("public struct %s {\n\n", "Api"), indent+2)
	err = cppe.emitToFile(str)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}

func (cppe *CSharpEmitter) PostVisitPackage(name string, indent int) {
	str := cppe.emitAsString("}\n", indent+2)
	cppe.emitToFile(str)
	err := cppe.emitToFile("}\n")
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}

func (cppe *CSharpEmitter) PreVisitFuncDeclSignatures(indent int) {
	cppe.forwardDecls = true
}

func (cppe *CSharpEmitter) PostVisitFuncDeclSignatures(indent int) {
	cppe.forwardDecls = false
}

func (cppe *CSharpEmitter) PreVisitFuncDeclName(node *ast.Ident, indent int) {
	if cppe.forwardDecls {
		return
	}
	var str string
	if node.Name == "main" {
		str = cppe.emitAsString(fmt.Sprintf("Main"), 0)
	} else {
		str = cppe.emitAsString(fmt.Sprintf("%s", node.Name), 0)
	}
	cppe.emitToFile(str)

}

func (cppe *CSharpEmitter) PreVisitBlockStmt(node *ast.BlockStmt, indent int) {
	str := cppe.emitAsString("{\n", indent+2)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PostVisitBlockStmt(node *ast.BlockStmt, indent int) {
	str := cppe.emitAsString("}", indent+2)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PostVisitBlockStmtList(node ast.Stmt, index int, indent int) {
	str := cppe.emitAsString("\n", indent)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PostVisitFuncDecl(node *ast.FuncDecl, indent int) {
	if cppe.forwardDecls {
		return
	}
	str := cppe.emitAsString("\n\n", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PreVisitGenStructInfo(node GenStructInfo, indent int) {
	str := cppe.emitAsString(fmt.Sprintf("public struct %s\n", node.Name), indent+2)
	str += cppe.emitAsString("{\n", indent+2)
	cppe.emitToFile(str)
	cppe.insideStruct = true
}

func (cppe *CSharpEmitter) PostVisitGenStructInfo(node GenStructInfo, indent int) {
	str := cppe.emitAsString("};\n\n", indent+2)
	cppe.emitToFile(str)
	cppe.insideStruct = false
}

func (cppe *CSharpEmitter) PreVisitArrayType(node ast.ArrayType, indent int) {
	if !cppe.insideStruct {
		return
	}
	cppe.stack = append(cppe.stack, "@@PreVisitArrayType")
	str := cppe.emitAsString("List<", indent)

	cppe.stack = append(cppe.stack, str)
	cppe.buffer = true
}
func (cppe *CSharpEmitter) PostVisitArrayType(node ast.ArrayType, indent int) {
	if !cppe.insideStruct {
		return
	}

	cppe.stack = append(cppe.stack, cppe.emitAsString(">", 0))

	cppe.mergeStackElements("@@PreVisitArrayType")

	if len(cppe.stack) == 1 {
		cppe.emitToFile(cppe.stack[len(cppe.stack)-1])
		cppe.stack = cppe.stack[:len(cppe.stack)-1]
	}

	cppe.buffer = false
}

func (cppe *CSharpEmitter) PreVisitFuncType(node *ast.FuncType, indent int) {
	if !cppe.insideStruct {
		return
	}
	cppe.buffer = true
	cppe.stack = append(cppe.stack, "@@PreVisitFuncType")
	str := cppe.emitAsString("Func<", indent)
	cppe.stack = append(cppe.stack, str)
}
func (cppe *CSharpEmitter) PostVisitFuncType(node *ast.FuncType, indent int) {
	if !cppe.insideStruct {
		return
	}
	cppe.mergeStackElements("@@PreVisitFuncType")

	if len(cppe.stack) == 1 {
		cppe.emitToFile(cppe.stack[len(cppe.stack)-1])
		cppe.stack = cppe.stack[:len(cppe.stack)-1]
	}
	str := cppe.emitAsString(">", 0)
	cppe.emitToFile(str)
	cppe.buffer = false
}

func (cppe *CSharpEmitter) PreVisitFuncTypeParam(node *ast.Field, index int, indent int) {
	if index == 0 && cppe.numFuncResults > 0 {
		str := cppe.emitAsString(", ", 0)
		cppe.stack = append(cppe.stack, str)
	}
	if index > 0 {
		str := cppe.emitAsString(", ", 0)
		cppe.stack = append(cppe.stack, str)
	}
}

func (cppe *CSharpEmitter) PostVisitSelectorExprX(node ast.Expr, indent int) {
	if !cppe.insideStruct {
		return
	}
	var str string
	const scopeOperator = ".Api."
	if ident, ok := node.(*ast.Ident); ok {
		if cppe.lowerToBuiltins(ident.Name) == "" {
			return
		}
	}
	str = cppe.emitAsString(scopeOperator, 0)
	if cppe.buffer {
		cppe.stack = append(cppe.stack, str)
	} else {
		cppe.emitToFile(str)
	}

}

func (cppe *CSharpEmitter) PreVisitFuncTypeResults(node *ast.FieldList, indent int) {
	if node != nil {
		cppe.numFuncResults = len(node.List)
	}
}

func (cppe *CSharpEmitter) PreVisitFuncDeclSignatureTypeParams(node *ast.FuncDecl, indent int) {
	if cppe.forwardDecls {
		return
	}
	cppe.insideStruct = true
	str := cppe.emitAsString("(", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PostVisitFuncDeclSignatureTypeParams(node *ast.FuncDecl, indent int) {
	if cppe.forwardDecls {
		return
	}
	cppe.insideStruct = false
	str := cppe.emitAsString(")", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PreVisitFuncDeclSignatureTypeParamsList(node *ast.Field, index int, indent int) {
	if cppe.forwardDecls {
		return
	}
	if index > 0 {
		str := cppe.emitAsString(", ", 0)
		cppe.emitToFile(str)
	}
}

func (cppe *CSharpEmitter) PreVisitFuncDeclSignatureTypeParamsArgName(node *ast.Ident, index int, indent int) {
	if cppe.forwardDecls {
		return
	}
	cppe.emitToFile(" ")
}

func (cppe *CSharpEmitter) PreVisitFuncDeclSignatureTypeResultsList(node *ast.Field, index int, indent int) {
	if cppe.forwardDecls {
		return
	}
	if index > 0 {
		str := cppe.emitAsString(",", 0)
		cppe.emitToFile(str)
	}
}

func (cppe *CSharpEmitter) PreVisitFuncDeclSignatureTypeResults(node *ast.FuncDecl, indent int) {
	if cppe.forwardDecls {
		return
	}

	cppe.insideStruct = true

	str := cppe.emitAsString("public static ", indent+2)
	cppe.emitToFile(str)
	if node.Type.Results != nil {
		if len(node.Type.Results.List) > 1 {
			str := cppe.emitAsString("Tuple<", 0)
			cppe.emitToFile(str)
		}
	} else {
		str := cppe.emitAsString("void", 0)
		cppe.emitToFile(str)
	}
}

func (cppe *CSharpEmitter) PostVisitFuncDeclSignatureTypeResults(node *ast.FuncDecl, indent int) {
	if cppe.forwardDecls {
		return
	}

	if node.Type.Results != nil {
		if len(node.Type.Results.List) > 1 {
			str := cppe.emitAsString(">", 0)
			cppe.emitToFile(str)
		}
	}

	str := cppe.emitAsString("", 1)
	cppe.emitToFile(str)
	cppe.insideStruct = false
}

func (cppe *CSharpEmitter) PreVisitTypeAliasName(node *ast.Ident, indent int) {
	cppe.stack = append(cppe.stack, "@@PreVisitTypeAliasName")
	cppe.stack = append(cppe.stack, cppe.emitAsString("using ", indent+2))
	cppe.insideStruct = true
	cppe.buffer = true
}

func (cppe *CSharpEmitter) PostVisitTypeAliasName(node *ast.Ident, indent int) {
	cppe.buffer = true
	cppe.stack = append(cppe.stack, " = ")
}

func (cppe *CSharpEmitter) PostVisitTypeAliasType(node ast.Expr, indent int) {
	str := cppe.emitAsString(";\n\n", 0)
	cppe.stack = append(cppe.stack, str)
	cppe.mergeStackElements("@@PreVisitTypeAliasName")
	if len(cppe.stack) == 1 {
		cppe.emitToFile(cppe.stack[len(cppe.stack)-1])
		cppe.stack = cppe.stack[:len(cppe.stack)-1]
	}
	cppe.insideStruct = false
	cppe.buffer = false
}

/*
func (cppe *CSharpEmitter) PreVisitReturnStmt(node *ast.ReturnStmt, indent int) {
	cppe.insideStruct = true
	str := cppe.emitAsString("return ", indent)
	cppe.emitToFile(str)
	if len(node.Results) > 1 {
		str := cppe.emitAsString("Tuple.Create(", 0)
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
	cppe.insideStruct = false
}

func (cppe *CSharpEmitter) PreVisitReturnStmtResult(node ast.Expr, index int, indent int) {
	if index > 0 {
		str := cppe.emitAsString(", ", 0)
		cppe.emitToFile(str)
	}
}
*/

package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/packages"
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
	pkg               *packages.Package
	insideForPostCond bool
	assignmentToken   string
	forwardDecls      bool
	shouldGenerate    bool
	numFuncResults    int
	aliases           []string
	isAlias           bool
	currentPackage    string
	stack             []string
	buffer            bool
	isArray           bool
	arrayType         string
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
		return "SliceBuiltins.Length"
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
	outputFile := "./Program.cs"
	var err error
	cppe.file, err = os.Create(outputFile)
	cppe.SetFile(cppe.file)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	_, err = cppe.file.WriteString("using System;\nusing System.Collections;\nusing System.Collections.Generic;\n\n")
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
	builtin := `public static class SliceBuiltins
{
  public static List<T> Append<T>(this List<T> list, T element)
  {
    var result = new List<T>(list);
    result.Add(element);
    return result;
  }

  public static List<T> Append<T>(this List<T> list, params T[] elements)
  {
    var result = new List<T>(list);
    result.AddRange(elements);
    return result;
  }

  public static List<T> Append<T>(this List<T> list, List<T> elements)
  {
    var result = new List<T>(list);
    result.AddRange(elements);
    return result;
  }

  // Fix: Ensure Length works for collections and not generic T
  public static int Length<T>(ICollection<T> collection)
  {
    return collection.Count;
  }
}
`
	str := cppe.emitAsString(builtin, indent)
	cppe.emitToFile(str)

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
	if cppe.isArray {
		cppe.emitToFile(" = new ")
		cppe.emitToFile(strings.TrimSpace(cppe.arrayType))
		cppe.emitToFile("()")
		cppe.isArray = false
	}
	cppe.emitToFile(";")
}

func (cppe *CSharpEmitter) PreVisitGenStructFieldType(node ast.Expr, indent int) {
	str := cppe.emitAsString("public", indent+2)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PostVisitGenStructFieldType(node ast.Expr, indent int) {
	cppe.emitToFile(" ")
	// clean array marker as we should generate
	// initializer only for expression statements
	// not for struct fields
	cppe.isArray = false
}

func (cppe *CSharpEmitter) PostVisitGenStructFieldName(node *ast.Ident, indent int) {
	cppe.emitToFile(";\n")
}

func (cppe *CSharpEmitter) PreVisitIdent(e *ast.Ident, indent int) {
	obj := cppe.pkg.TypesInfo.Defs[e]
	if obj != nil {
		if v, ok := obj.(*types.Var); ok {
			pos := cppe.pkg.Fset.Position(v.Pos())
			fmt.Printf("Variable %s has type: %s (declared at %s:%d)\n", v.Name(), v.Type().String(), pos.Filename, pos.Line)
		}
	}

	obj = cppe.pkg.TypesInfo.Uses[e]
	if obj != nil {
		if v, ok := obj.(*types.Var); ok {
			usagePos := cppe.pkg.Fset.Position(e.Pos()) // <- position of the usage, not the declaration
			fmt.Printf("Variable %s has type: %s (used at %s:%d)\n", v.Name(), v.Type().String(), usagePos.Filename, usagePos.Line)
		}
	}

	if !cppe.shouldGenerate {
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

func (cppe *CSharpEmitter) PreVisitPackage(pkg *packages.Package, indent int) {
	name := pkg.Name
	cppe.pkg = pkg
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

func (cppe *CSharpEmitter) PostVisitPackage(pkg *packages.Package, indent int) {
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
	cppe.isArray = false
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
	cppe.shouldGenerate = true
}

func (cppe *CSharpEmitter) PostVisitGenStructInfo(node GenStructInfo, indent int) {
	str := cppe.emitAsString("};\n\n", indent+2)
	cppe.emitToFile(str)
	cppe.shouldGenerate = false
}

func (cppe *CSharpEmitter) PreVisitArrayType(node ast.ArrayType, indent int) {
	if !cppe.shouldGenerate {
		return
	}
	cppe.stack = append(cppe.stack, "@@PreVisitArrayType")
	str := cppe.emitAsString("List<", indent)

	cppe.stack = append(cppe.stack, str)
	cppe.buffer = true
}
func (cppe *CSharpEmitter) PostVisitArrayType(node ast.ArrayType, indent int) {
	if !cppe.shouldGenerate {
		return
	}

	cppe.stack = append(cppe.stack, cppe.emitAsString(">", 0))

	cppe.mergeStackElements("@@PreVisitArrayType")
	if len(cppe.stack) == 1 {
		cppe.isArray = true
		cppe.arrayType = cppe.stack[len(cppe.stack)-1]
		cppe.emitToFile(cppe.stack[len(cppe.stack)-1])
		cppe.stack = cppe.stack[:len(cppe.stack)-1]
	}

	cppe.buffer = false
}

func (cppe *CSharpEmitter) PreVisitFuncType(node *ast.FuncType, indent int) {
	if !cppe.shouldGenerate {
		return
	}
	cppe.buffer = true
	cppe.stack = append(cppe.stack, "@@PreVisitFuncType")
	str := cppe.emitAsString("Func<", indent)
	cppe.stack = append(cppe.stack, str)
}
func (cppe *CSharpEmitter) PostVisitFuncType(node *ast.FuncType, indent int) {
	if !cppe.shouldGenerate {
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
	if !cppe.shouldGenerate {
		return
	}
	var str string
	const scopeOperator = "."
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
	cppe.shouldGenerate = true
	str := cppe.emitAsString("(", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PostVisitFuncDeclSignatureTypeParams(node *ast.FuncDecl, indent int) {
	if cppe.forwardDecls {
		return
	}
	cppe.shouldGenerate = false
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

	cppe.shouldGenerate = true

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
	cppe.shouldGenerate = false
}

func (cppe *CSharpEmitter) PreVisitTypeAliasName(node *ast.Ident, indent int) {
	cppe.stack = append(cppe.stack, "@@PreVisitTypeAliasName")
	cppe.stack = append(cppe.stack, cppe.emitAsString("using ", indent+2))
	cppe.shouldGenerate = true
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
	cppe.shouldGenerate = false
	cppe.buffer = false
}

func (cppe *CSharpEmitter) PreVisitReturnStmt(node *ast.ReturnStmt, indent int) {
	cppe.shouldGenerate = true
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
	cppe.shouldGenerate = false
}

func (cppe *CSharpEmitter) PreVisitReturnStmtResult(node ast.Expr, index int, indent int) {
	if index > 0 {
		str := cppe.emitAsString(", ", 0)
		cppe.emitToFile(str)
	}
}

func (v *CSharpEmitter) PreVisitCallExpr(node *ast.CallExpr, indent int) {
	v.shouldGenerate = true
}

func (v *CSharpEmitter) PostVisitCallExpr(node *ast.CallExpr, indent int) {
	v.shouldGenerate = false
}

func (v *CSharpEmitter) PreVisitDeclStmt(node *ast.DeclStmt, indent int) {
	v.shouldGenerate = true
}

func (v *CSharpEmitter) PostVisitDeclStmt(node *ast.DeclStmt, indent int) {
	v.shouldGenerate = false
}

func (cppe *CSharpEmitter) PreVisitAssignStmt(node *ast.AssignStmt, indent int) {
	cppe.shouldGenerate = true
	str := cppe.emitAsString("", indent)
	cppe.emitToFile(str)
}
func (cppe *CSharpEmitter) PostVisitAssignStmt(node *ast.AssignStmt, indent int) {
	str := cppe.emitAsString(";", 0)
	cppe.emitToFile(str)
	cppe.shouldGenerate = false
}

func (cppe *CSharpEmitter) PreVisitAssignStmtRhs(node *ast.AssignStmt, indent int) {
	cppe.shouldGenerate = true
	str := cppe.emitAsString(cppe.assignmentToken+" ", indent+1)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PostVisitAssignStmtRhs(node *ast.AssignStmt, indent int) {
	cppe.shouldGenerate = false
}

func (cppe *CSharpEmitter) PreVisitAssignStmtLhsExpr(node ast.Expr, index int, indent int) {
	if index > 0 {
		str := cppe.emitAsString(", ", indent)
		cppe.emitToFile(str)
	}
}

func (cppe *CSharpEmitter) PreVisitAssignStmtLhs(node *ast.AssignStmt, indent int) {
	cppe.shouldGenerate = true
	assignmentToken := node.Tok.String()
	if assignmentToken == ":=" && len(node.Lhs) == 1 {
		str := cppe.emitAsString("var ", indent)
		cppe.emitToFile(str)
	} else if assignmentToken == ":=" && len(node.Lhs) > 1 {
		str := cppe.emitAsString("var [", indent)
		cppe.emitToFile(str)
	} else if assignmentToken == "=" && len(node.Lhs) > 1 {
		str := cppe.emitAsString("(", indent)
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
	cppe.shouldGenerate = false
}

func (cppe *CSharpEmitter) PreVisitIndexExprIndex(node *ast.IndexExpr, indent int) {
	cppe.shouldGenerate = true
	str := cppe.emitAsString("[", 0)
	cppe.emitToFile(str)

}
func (cppe *CSharpEmitter) PostVisitIndexExprIndex(node *ast.IndexExpr, indent int) {
	cppe.shouldGenerate = false
	str := cppe.emitAsString("]", 0)
	cppe.emitToFile(str)
}

func (v *CSharpEmitter) PreVisitBinaryExpr(node *ast.BinaryExpr, indent int) {
	v.shouldGenerate = true
}
func (v *CSharpEmitter) PostVisitBinaryExpr(node *ast.BinaryExpr, indent int) {
	v.shouldGenerate = false
}

func (cppe *CSharpEmitter) PreVisitBinaryExprOperator(op token.Token, indent int) {
	str := cppe.emitAsString(op.String()+" ", 1)
	cppe.emitToFile(str)
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
func (cppe *CSharpEmitter) PostVisitExprStmtX(node ast.Expr, indent int) {
	str := cppe.emitAsString(";", 0)
	cppe.emitToFile(str)
}

func (v *CSharpEmitter) PreVisitIfStmt(node *ast.IfStmt, indent int) {
	v.shouldGenerate = true
}
func (v *CSharpEmitter) PostVisitIfStmt(node *ast.IfStmt, indent int) {
	v.shouldGenerate = false
}

func (cppe *CSharpEmitter) PreVisitIfStmtCond(node *ast.IfStmt, indent int) {
	str := cppe.emitAsString("if (", indent)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PostVisitIfStmtCond(node *ast.IfStmt, indent int) {
	str := cppe.emitAsString(")\n", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PreVisitForStmt(node *ast.ForStmt, indent int) {
	cppe.insideForPostCond = true
	str := cppe.emitAsString("for (", indent)
	cppe.emitToFile(str)
	cppe.shouldGenerate = true
}

func (cppe *CSharpEmitter) PostVisitForStmtInit(node ast.Stmt, indent int) {
	if node == nil {
		str := cppe.emitAsString(";", 0)
		cppe.emitToFile(str)
	}
}

func (cppe *CSharpEmitter) PostVisitForStmtPost(node ast.Stmt, indent int) {
	if node != nil {
		cppe.insideForPostCond = false
	}
	str := cppe.emitAsString(")\n", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PreVisitIfStmtElse(node *ast.IfStmt, indent int) {
	str := cppe.emitAsString("else", 1)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PostVisitForStmtCond(node ast.Expr, indent int) {
	str := cppe.emitAsString(";", 0)
	cppe.emitToFile(str)
	cppe.shouldGenerate = false
}

func (cppe *CSharpEmitter) PostVisitForStmt(node *ast.ForStmt, indent int) {
	cppe.shouldGenerate = false
}

func (cppe *CSharpEmitter) PreVisitRangeStmt(node *ast.RangeStmt, indent int) {
	cppe.shouldGenerate = true
	str := cppe.emitAsString("foreach (var ", indent)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PostVisitRangeStmtValue(node ast.Expr, indent int) {
	str := cppe.emitAsString(" in ", 0)
	cppe.emitToFile(str)
}

func (cppe *CSharpEmitter) PostVisitRangeStmtX(node ast.Expr, indent int) {
	str := cppe.emitAsString(")\n", 0)
	cppe.emitToFile(str)
	cppe.shouldGenerate = false
}

func (cppe *CSharpEmitter) PreVisitIncDecStmt(node *ast.IncDecStmt, indent int) {
	cppe.shouldGenerate = true
}

func (cppe *CSharpEmitter) PostVisitIncDecStmt(node *ast.IncDecStmt, indent int) {
	str := cppe.emitAsString(node.Tok.String(), 0)
	if !cppe.insideForPostCond {
		str += cppe.emitAsString(";", 0)
	}
	cppe.emitToFile(str)
	cppe.shouldGenerate = false
}

func (v *CSharpEmitter) PreVisitCompositeLitType(node ast.Expr, indent int) {
	str := v.emitAsString("new ", 0)
	v.emitToFile(str)
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

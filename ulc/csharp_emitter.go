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

var destTypes = []string{"sbyte", "short", "int", "long", "byte", "ushort", "object", "string"}

var csTypesMap = map[string]string{
	"int8":   destTypes[0],
	"int16":  destTypes[1],
	"int32":  destTypes[2],
	"int64":  destTypes[3],
	"uint8":  destTypes[4],
	"uint16": destTypes[5],
	"any":    destTypes[6],
	"string": destTypes[7],
}

type PointerAndPosition struct {
	Pointer  string // Pointer to the type
	Position int
}

type CSharpEmitter struct {
	Output string
	file   *os.File
	Emitter
	pkg                   *packages.Package
	insideForPostCond     bool
	assignmentToken       string
	forwardDecls          bool
	shouldGenerate        bool
	numFuncResults        int
	aliases               []string
	isAlias               bool
	currentPackage        string
	stack                 []string
	buffer                bool
	isArray               bool
	arrayType             string
	isTuple               bool
	fileBuffer            string
	PointerAndPositionVec []PointerAndPosition
}

func (e *CSharpEmitter) SearchPointerReverse(target string) *PointerAndPosition {
	for i := len(e.PointerAndPositionVec) - 1; i >= 0; i-- {
		if e.PointerAndPositionVec[i].Pointer == target {
			return &e.PointerAndPositionVec[i]
		}
	}
	return nil // Return nil if the pointer is not found
}

func (e *CSharpEmitter) ExtractSubstring(position int) (string, error) {
	if position < 0 || position >= len(e.fileBuffer) {
		return "", fmt.Errorf("position %d is out of bounds", position)
	}
	return e.fileBuffer[position:], nil
}

func (e *CSharpEmitter) RewriteFileBuffer(position int, oldContent, newContent string) error {
	if position < 0 || position+len(oldContent) > len(e.fileBuffer) {
		return fmt.Errorf("position %d is out of bounds or oldContent does not match", position)
	}
	if e.fileBuffer[position:position+len(oldContent)] != oldContent {
		return fmt.Errorf("oldContent does not match the existing content at position %d", position)
	}
	e.fileBuffer = e.fileBuffer[:position] + newContent + e.fileBuffer[position+len(oldContent):]
	return nil
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
		return "Formatter.Sprintf"
	case "Println":
		return "Console.WriteLine"
	case "Printf":
		return "Formatter.Printf"
	case "Print":
		return "Formatter.Printf"
	case "len":
		return "SliceBuiltins.Length"
	case "append":
		return "SliceBuiltins.Append"
	}
	return selector
}
func (e *CSharpEmitter) emitToFileBuffer(s string, pointer string) error {
	e.PointerAndPositionVec = append(e.PointerAndPositionVec, PointerAndPosition{
		Pointer:  pointer,
		Position: len(e.fileBuffer),
	})
	e.fileBuffer += s
	return nil
}

func (e *CSharpEmitter) emitToFile() error {
	_, err := e.file.WriteString(e.fileBuffer)
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
	cppe.PointerAndPositionVec = make([]PointerAndPosition, 0)
	outputFile := cppe.Output
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
    return collection == null ? 0 : collection.Count;
  }
  public static int Length(string s)
  {
    return s == null ? 0 : s.Length;
  }
}
public class Formatter
{
    public static void Printf(string format, params object[] args)
    {
        // Replace %d → {0}, %s → {1}, etc.
        int argIndex = 0;
        string converted = "";
        for (int i = 0; i < format.Length; i++)
        {
            if (format[i] == '%' && i + 1 < format.Length)
            {
                char next = format[i + 1];
                switch (next)
                {
                    case 'd':
                    case 's':
                    case 'f':
                        converted += "{" + argIndex++ + "}";
                        i++; // Skip format char
                        continue;
                }
            }
            converted += format[i];
        }

        Console.Write(converted, args);
    }
	public static string Sprintf(string format, params object[] args)
    {
        int argIndex = 0;
        string converted = "";

        for (int i = 0; i < format.Length; i++)
        {
            if (format[i] == '%' && i + 1 < format.Length)
            {
                char next = format[i + 1];
                switch (next)
                {
                    case 'd':
                    case 's':
                    case 'f':
                        converted += "{" + argIndex++ + "}";
                        i++; // Skip format character
                        continue;
                    case '%':
                        converted += "%"; // Escaped percent
                        i++;
                        continue;
                }
            }

            converted += format[i];
        }
        return string.Format(converted, args);
    }
}
`
	str := cppe.emitAsString(builtin, indent)
	cppe.emitToFileBuffer(str, "")

	cppe.insideForPostCond = false
}

func (cppe *CSharpEmitter) PostVisitProgram(indent int) {
	cppe.emitToFile()
	cppe.file.Close()
}

func (cppe *CSharpEmitter) PreVisitDeclStmtValueSpecNames(node *ast.Ident, index int, indent int) {
	str := cppe.emitAsString(" ", 0)
	cppe.emitToFileBuffer(str, "")
}

func (cppe *CSharpEmitter) PostVisitDeclStmtValueSpecNames(node *ast.Ident, index int, indent int) {
	var str string
	if cppe.isArray {
		str += " = new "
		str += strings.TrimSpace(cppe.arrayType)
		str += "();"
		cppe.isArray = false
	} else {
		str += " = default;"
	}
	cppe.emitToFileBuffer(str, "")
}

func (cppe *CSharpEmitter) PreVisitGenStructFieldType(node ast.Expr, indent int) {
	str := cppe.emitAsString("public", indent+2)
	cppe.emitToFileBuffer(str, "")
}

func (cppe *CSharpEmitter) PostVisitGenStructFieldType(node ast.Expr, indent int) {
	cppe.emitToFileBuffer(" ", "")
	// clean array marker as we should generate
	// initializer only for expression statements
	// not for struct fields
	cppe.isArray = false
}

func (cppe *CSharpEmitter) PostVisitGenStructFieldName(node *ast.Ident, indent int) {
	cppe.emitToFileBuffer(";\n", "")
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
		str = cppe.emitAsString("default", indent)
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
		cppe.emitToFileBuffer(str, "")
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
	err := cppe.emitToFileBuffer(str, "")

	for _, alias := range cppe.aliases {
		str := cppe.emitAsString(alias, indent+2)
		cppe.emitToFileBuffer(str, "")
	}
	cppe.currentPackage = packageName
	str = cppe.emitAsString(fmt.Sprintf("public class %s {\n\n", "Api"), indent+2)
	err = cppe.emitToFileBuffer(str, "")
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}

func (cppe *CSharpEmitter) PostVisitPackage(pkg *packages.Package, indent int) {
	str := cppe.emitAsString("}\n", indent+2)
	cppe.emitToFileBuffer(str, "")
	err := cppe.emitToFileBuffer("}\n", "")
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

func (cppe *CSharpEmitter) PostVisitFuncDeclSignature(node *ast.FuncDecl, indent int) {
	cppe.isArray = false
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
	cppe.emitToFileBuffer(str, "")

}

func (cppe *CSharpEmitter) PreVisitBlockStmt(node *ast.BlockStmt, indent int) {
	str := cppe.emitAsString("{\n", indent+2)
	cppe.emitToFileBuffer(str, "")
}

func (cppe *CSharpEmitter) PostVisitBlockStmt(node *ast.BlockStmt, indent int) {
	str := cppe.emitAsString("}", indent+2)
	cppe.emitToFileBuffer(str, "")
	cppe.isArray = false
}

func (cppe *CSharpEmitter) PostVisitBlockStmtList(node ast.Stmt, index int, indent int) {
	str := cppe.emitAsString("\n", indent)
	cppe.emitToFileBuffer(str, "")
}

func (cppe *CSharpEmitter) PostVisitFuncDecl(node *ast.FuncDecl, indent int) {
	if cppe.forwardDecls {
		return
	}
	str := cppe.emitAsString("\n\n", 0)
	cppe.emitToFileBuffer(str, "")
}

func (cppe *CSharpEmitter) PreVisitGenStructInfo(node GenStructInfo, indent int) {
	str := cppe.emitAsString(fmt.Sprintf("public class %s\n", node.Name), indent+2)
	str += cppe.emitAsString("{\n", indent+2)
	cppe.emitToFileBuffer(str, "")
	cppe.shouldGenerate = true
}

func (cppe *CSharpEmitter) PostVisitGenStructInfo(node GenStructInfo, indent int) {
	str := cppe.emitAsString("};\n\n", indent+2)
	cppe.emitToFileBuffer(str, "")
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
		cppe.emitToFileBuffer(cppe.stack[len(cppe.stack)-1], "")
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
	var str string
	if node.Results != nil {
		str = cppe.emitAsString("Func<", indent)
	} else {
		str = cppe.emitAsString("Action<", indent)
	}
	cppe.stack = append(cppe.stack, str)
}
func (cppe *CSharpEmitter) PostVisitFuncType(node *ast.FuncType, indent int) {
	if !cppe.shouldGenerate {
		return
	}

	// move return type to the end of the stack
	// return type is traversed first therefore it has to be moved
	// to the end of the stack due to C# syntax
	if len(cppe.stack) > 2 && cppe.numFuncResults > 0 {
		returnType := cppe.stack[2]
		cppe.stack = append(cppe.stack[:2], cppe.stack[3:]...)
		cppe.stack = append(cppe.stack, ",")
		cppe.stack = append(cppe.stack, returnType)
	}
	cppe.stack = append(cppe.stack, cppe.emitAsString(">", 0))

	cppe.mergeStackElements("@@PreVisitFuncType")

	if len(cppe.stack) == 1 {
		cppe.emitToFileBuffer(cppe.stack[len(cppe.stack)-1], "")
		cppe.stack = cppe.stack[:len(cppe.stack)-1]
	}
	cppe.buffer = false
}

func (cppe *CSharpEmitter) PreVisitFuncTypeParam(node *ast.Field, index int, indent int) {
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
	scopeOperator := "."
	if ident, ok := node.(*ast.Ident); ok {
		if cppe.lowerToBuiltins(ident.Name) == "" {
			return
		}
		// if the identifier is a package name, we need to append "Api." to the scope operator
		obj := cppe.pkg.TypesInfo.Uses[ident]
		if obj != nil {
			if _, ok := obj.(*types.PkgName); ok {
				scopeOperator += "Api."
			}
		}
	}

	str = cppe.emitAsString(scopeOperator, 0)
	if cppe.buffer {
		cppe.stack = append(cppe.stack, str)
	} else {
		cppe.emitToFileBuffer(str, "")
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
	cppe.emitToFileBuffer(str, "")
}

func (cppe *CSharpEmitter) PostVisitFuncDeclSignatureTypeParams(node *ast.FuncDecl, indent int) {
	if cppe.forwardDecls {
		return
	}
	cppe.shouldGenerate = false
	str := cppe.emitAsString(")", 0)
	cppe.emitToFileBuffer(str, "")
}

func (cppe *CSharpEmitter) PreVisitFuncDeclSignatureTypeParamsList(node *ast.Field, index int, indent int) {
	if cppe.forwardDecls {
		return
	}
	if index > 0 {
		str := cppe.emitAsString(", ", 0)
		cppe.emitToFileBuffer(str, "")
	}
}

func (cppe *CSharpEmitter) PreVisitFuncDeclSignatureTypeParamsArgName(node *ast.Ident, index int, indent int) {
	if cppe.forwardDecls {
		return
	}
	cppe.emitToFileBuffer(" ", "")
}

func (cppe *CSharpEmitter) PreVisitFuncDeclSignatureTypeResultsList(node *ast.Field, index int, indent int) {
	if cppe.forwardDecls {
		return
	}
	if index > 0 {
		str := cppe.emitAsString(",", 0)
		cppe.emitToFileBuffer(str, "")
	}
}

func (cppe *CSharpEmitter) PreVisitFuncDeclSignatureTypeResults(node *ast.FuncDecl, indent int) {
	if cppe.forwardDecls {
		return
	}

	cppe.shouldGenerate = true

	str := cppe.emitAsString("public static ", indent+2)
	cppe.emitToFileBuffer(str, "")
	if node.Type.Results != nil {
		if len(node.Type.Results.List) > 1 {
			str := cppe.emitAsString("(", 0)
			cppe.emitToFileBuffer(str, "")
		}
	} else {
		str := cppe.emitAsString("void", 0)
		cppe.emitToFileBuffer(str, "")
	}
}

func (cppe *CSharpEmitter) PostVisitFuncDeclSignatureTypeResults(node *ast.FuncDecl, indent int) {
	if cppe.forwardDecls {
		return
	}

	if node.Type.Results != nil {
		if len(node.Type.Results.List) > 1 {
			str := cppe.emitAsString(")", 0)
			cppe.emitToFileBuffer(str, "")
		}
	}

	str := cppe.emitAsString("", 1)
	cppe.emitToFileBuffer(str, "")
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
		cppe.emitToFileBuffer(cppe.stack[len(cppe.stack)-1], "")
		cppe.stack = cppe.stack[:len(cppe.stack)-1]
	}
	cppe.shouldGenerate = false
	cppe.buffer = false
}

func (cppe *CSharpEmitter) PreVisitReturnStmt(node *ast.ReturnStmt, indent int) {
	cppe.shouldGenerate = true
	str := cppe.emitAsString("return ", indent)
	cppe.emitToFileBuffer(str, "")
	if len(node.Results) == 1 {
		tv := cppe.pkg.TypesInfo.Types[node.Results[0]]
		//pos := cppe.pkg.Fset.Position(node.Pos())
		//fmt.Printf("@@Type: %s %s:%d:%d\n", tv.Type, pos.Filename, pos.Line, pos.Column)
		if typeVal, ok := csTypesMap[tv.Type.String()]; ok {
			if !cppe.isTuple && tv.Type.String() != "func()" {
				cppe.emitToFileBuffer("(", "")
				str := cppe.emitAsString(typeVal, 0)
				cppe.emitToFileBuffer(str, "")
				cppe.emitToFileBuffer(")", "")
			}
		}
	}
	if len(node.Results) > 1 {
		str := cppe.emitAsString("(", 0)
		cppe.emitToFileBuffer(str, "")
	}
}

func (cppe *CSharpEmitter) PostVisitReturnStmt(node *ast.ReturnStmt, indent int) {
	if len(node.Results) > 1 {
		str := cppe.emitAsString(")", 0)
		cppe.emitToFileBuffer(str, "")
	}
	str := cppe.emitAsString(";", 0)
	cppe.emitToFileBuffer(str, "")
	cppe.shouldGenerate = false
}

func (cppe *CSharpEmitter) PreVisitReturnStmtResult(node ast.Expr, index int, indent int) {
	if index > 0 {
		str := cppe.emitAsString(", ", 0)
		cppe.emitToFileBuffer(str, "")
	}
}

func (v *CSharpEmitter) PreVisitCallExpr(node *ast.CallExpr, indent int) {
	v.shouldGenerate = true
	v.emitToFileBuffer("", "@PreVisitCallExpr")
}

func (v *CSharpEmitter) PostVisitCallExpr(node *ast.CallExpr, indent int) {
	pointerAndPosition := v.SearchPointerReverse("@PreVisitCallExpr")
	if pointerAndPosition != nil {
		str, _ := v.ExtractSubstring(pointerAndPosition.Position)
		for _, t := range destTypes {
			matchStr := t + "("
			if strings.Contains(str, matchStr) {
				v.RewriteFileBuffer(pointerAndPosition.Position, matchStr, "("+t+")(")
			}
		}
	}
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
	cppe.emitToFileBuffer(str, "")
}
func (cppe *CSharpEmitter) PostVisitAssignStmt(node *ast.AssignStmt, indent int) {
	str := cppe.emitAsString(";", 0)
	cppe.emitToFileBuffer(str, "")
	cppe.shouldGenerate = false
}

func (cppe *CSharpEmitter) PreVisitAssignStmtRhs(node *ast.AssignStmt, indent int) {
	cppe.shouldGenerate = true
	str := cppe.emitAsString(cppe.assignmentToken+" ", indent+1)
	cppe.emitToFileBuffer(str, "")
}

func (cppe *CSharpEmitter) PostVisitAssignStmtRhs(node *ast.AssignStmt, indent int) {
	cppe.shouldGenerate = false
	cppe.isTuple = false
}

func (cppe *CSharpEmitter) PreVisitAssignStmtRhsExpr(node ast.Expr, index int, indent int) {
	cppe.emitToFileBuffer("", "@PreVisitAssignStmtRhsExpr")
}

func (cppe *CSharpEmitter) PostVisitAssignStmtRhsExpr(node ast.Expr, index int, indent int) {
	pointerAndPosition := cppe.SearchPointerReverse("@PreVisitAssignStmtRhsExpr")
	rewritten := false
	if pointerAndPosition != nil {
		str, _ := cppe.ExtractSubstring(pointerAndPosition.Position)
		for _, t := range destTypes {
			matchStr := t + "("
			if strings.Contains(str, matchStr) {
				cppe.RewriteFileBuffer(pointerAndPosition.Position, matchStr, "("+t+")(")
				rewritten = true
			}
		}
	}
	if !rewritten {
		tv := cppe.pkg.TypesInfo.Types[node]
		//pos := cppe.pkg.Fset.Position(node.Pos())
		//fmt.Printf("@@Type: %s %s:%d:%d\n", tv.Type, pos.Filename, pos.Line, pos.Column)
		if typeVal, ok := csTypesMap[tv.Type.String()]; ok {
			if !cppe.isTuple && tv.Type.String() != "func()" {
				cppe.RewriteFileBuffer(pointerAndPosition.Position, "", "("+typeVal+")")
			}
		}
	}
}

func (cppe *CSharpEmitter) PreVisitAssignStmtLhsExpr(node ast.Expr, index int, indent int) {
	if index > 0 {
		str := cppe.emitAsString(", ", indent)
		cppe.emitToFileBuffer(str, "")
	}
}

func (cppe *CSharpEmitter) PreVisitAssignStmtLhs(node *ast.AssignStmt, indent int) {
	cppe.shouldGenerate = true
	assignmentToken := node.Tok.String()
	if assignmentToken == ":=" && len(node.Lhs) == 1 {
		str := cppe.emitAsString("var ", indent)
		cppe.emitToFileBuffer(str, "")
	} else if assignmentToken == ":=" && len(node.Lhs) > 1 {
		str := cppe.emitAsString("var [", indent)
		cppe.emitToFileBuffer(str, "")
	} else if assignmentToken == "=" && len(node.Lhs) > 1 {
		str := cppe.emitAsString("(", indent)
		cppe.emitToFileBuffer(str, "")
		cppe.isTuple = true
	}
	if assignmentToken != "+=" {
		assignmentToken = "="
	}
	cppe.assignmentToken = assignmentToken
}

func (cppe *CSharpEmitter) PostVisitAssignStmtLhs(node *ast.AssignStmt, indent int) {
	if node.Tok.String() == ":=" && len(node.Lhs) > 1 {
		str := cppe.emitAsString("]", indent)
		cppe.emitToFileBuffer(str, "")
	} else if node.Tok.String() == "=" && len(node.Lhs) > 1 {
		str := cppe.emitAsString(")", indent)
		cppe.emitToFileBuffer(str, "")
	}
	cppe.shouldGenerate = false

}

func (cppe *CSharpEmitter) PreVisitIndexExprIndex(node *ast.IndexExpr, indent int) {
	cppe.shouldGenerate = true
	str := cppe.emitAsString("[", 0)
	cppe.emitToFileBuffer(str, "")

}
func (cppe *CSharpEmitter) PostVisitIndexExprIndex(node *ast.IndexExpr, indent int) {
	str := cppe.emitAsString("]", 0)
	cppe.emitToFileBuffer(str, "")
}

func (v *CSharpEmitter) PreVisitBinaryExpr(node *ast.BinaryExpr, indent int) {
	v.shouldGenerate = true
	str := v.emitAsString("(", 1)
	v.emitToFileBuffer(str, "")
}
func (v *CSharpEmitter) PostVisitBinaryExpr(node *ast.BinaryExpr, indent int) {
	str := v.emitAsString(")", 1)
	v.emitToFileBuffer(str, "")
	v.shouldGenerate = false
}

func (cppe *CSharpEmitter) PreVisitBinaryExprOperator(op token.Token, indent int) {
	str := cppe.emitAsString(op.String()+" ", 1)
	cppe.emitToFileBuffer(str, "")
}

func (cppe *CSharpEmitter) PreVisitBasicLit(e *ast.BasicLit, indent int) {
	cppe.stack = append(cppe.stack, "@@PreVisitBasicLit")
	if e.Kind == token.STRING {
		e.Value = strings.Replace(e.Value, "\"", "", -1)
		if e.Value[0] == '`' {
			e.Value = strings.Replace(e.Value, "`", "", -1)
			str := (cppe.emitAsString(fmt.Sprintf("@\"(%s)\"", e.Value), 0))
			cppe.stack = append(cppe.stack, str)
		} else {
			str := (cppe.emitAsString(fmt.Sprintf("@\"%s\"", e.Value), 0))
			cppe.stack = append(cppe.stack, str)
		}
	} else {
		str := (cppe.emitAsString(e.Value, 0))
		cppe.stack = append(cppe.stack, str)
	}
	cppe.buffer = true
}

func (cppe *CSharpEmitter) PostVisitBasicLit(e *ast.BasicLit, indent int) {
	cppe.mergeStackElements("@@PreVisitBasicLit")
	if len(cppe.stack) == 1 {
		cppe.emitToFileBuffer(cppe.stack[len(cppe.stack)-1], "")
		cppe.stack = cppe.stack[:len(cppe.stack)-1]
	}

	cppe.buffer = false
}

func (cppe *CSharpEmitter) PreVisitCallExprArgs(node []ast.Expr, indent int) {
	str := cppe.emitAsString("(", 0)
	cppe.emitToFileBuffer(str, "")
}
func (cppe *CSharpEmitter) PostVisitCallExprArgs(node []ast.Expr, indent int) {
	str := cppe.emitAsString(")", 0)
	cppe.emitToFileBuffer(str, "")
}

func (cppe *CSharpEmitter) PreVisitCallExprArg(node ast.Expr, index int, indent int) {
	if index > 0 {
		str := cppe.emitAsString(", ", 0)
		cppe.emitToFileBuffer(str, "")
	}
}
func (cppe *CSharpEmitter) PostVisitExprStmtX(node ast.Expr, indent int) {
	str := cppe.emitAsString(";", 0)
	cppe.emitToFileBuffer(str, "")
}

func (v *CSharpEmitter) PreVisitIfStmt(node *ast.IfStmt, indent int) {
	v.shouldGenerate = true
}
func (v *CSharpEmitter) PostVisitIfStmt(node *ast.IfStmt, indent int) {
	v.shouldGenerate = false
}

func (cppe *CSharpEmitter) PreVisitIfStmtCond(node *ast.IfStmt, indent int) {
	str := cppe.emitAsString("if (", indent)
	cppe.emitToFileBuffer(str, "")
}

func (cppe *CSharpEmitter) PostVisitIfStmtCond(node *ast.IfStmt, indent int) {
	str := cppe.emitAsString(")\n", 0)
	cppe.emitToFileBuffer(str, "")
}

func (cppe *CSharpEmitter) PreVisitForStmt(node *ast.ForStmt, indent int) {
	cppe.insideForPostCond = true
	str := cppe.emitAsString("for (", indent)
	cppe.emitToFileBuffer(str, "")
	cppe.shouldGenerate = true
}

func (cppe *CSharpEmitter) PostVisitForStmtInit(node ast.Stmt, indent int) {
	if node == nil {
		str := cppe.emitAsString(";", 0)
		cppe.emitToFileBuffer(str, "")
	}
}

func (cppe *CSharpEmitter) PostVisitForStmtPost(node ast.Stmt, indent int) {
	if node != nil {
		cppe.insideForPostCond = false
	}
	str := cppe.emitAsString(")\n", 0)
	cppe.emitToFileBuffer(str, "")
}

func (cppe *CSharpEmitter) PreVisitIfStmtElse(node *ast.IfStmt, indent int) {
	str := cppe.emitAsString("else", 1)
	cppe.emitToFileBuffer(str, "")
}

func (cppe *CSharpEmitter) PostVisitForStmtCond(node ast.Expr, indent int) {
	str := cppe.emitAsString(";", 0)
	cppe.emitToFileBuffer(str, "")
	cppe.shouldGenerate = false
}

func (cppe *CSharpEmitter) PostVisitForStmt(node *ast.ForStmt, indent int) {
	cppe.shouldGenerate = false
	cppe.insideForPostCond = false
}

func (cppe *CSharpEmitter) PreVisitRangeStmt(node *ast.RangeStmt, indent int) {
	cppe.shouldGenerate = true
	str := cppe.emitAsString("foreach (var ", indent)
	cppe.emitToFileBuffer(str, "")
}

func (cppe *CSharpEmitter) PostVisitRangeStmtValue(node ast.Expr, indent int) {
	str := cppe.emitAsString(" in ", 0)
	cppe.emitToFileBuffer(str, "")
}

func (cppe *CSharpEmitter) PostVisitRangeStmtX(node ast.Expr, indent int) {
	str := cppe.emitAsString(")\n", 0)
	cppe.emitToFileBuffer(str, "")
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
	cppe.emitToFileBuffer(str, "")
	cppe.shouldGenerate = false
}

func (v *CSharpEmitter) PreVisitCompositeLitType(node ast.Expr, indent int) {
	str := v.emitAsString("new ", 0)
	v.emitToFileBuffer(str, "")
}

func (cppe *CSharpEmitter) PreVisitCompositeLitElts(node []ast.Expr, indent int) {
	str := cppe.emitAsString("{", 0)
	cppe.emitToFileBuffer(str, "")
}

func (cppe *CSharpEmitter) PostVisitCompositeLitElts(node []ast.Expr, indent int) {
	str := cppe.emitAsString("}", 0)
	cppe.emitToFileBuffer(str, "")
}

func (cppe *CSharpEmitter) PreVisitCompositeLitElt(node ast.Expr, index int, indent int) {
	if index > 0 {
		str := cppe.emitAsString(", ", 0)
		cppe.emitToFileBuffer(str, "")
	}
}
func (cppe *CSharpEmitter) PostVisitSliceExprX(node ast.Expr, indent int) {
	str := cppe.emitAsString("[", 0)
	cppe.emitToFileBuffer(str, "")
	cppe.shouldGenerate = false
}

func (cppe *CSharpEmitter) PostVisitSliceExpr(node *ast.SliceExpr, indent int) {
	str := cppe.emitAsString("]", 0)
	cppe.emitToFileBuffer(str, "")
	cppe.shouldGenerate = true
}

func (cppe *CSharpEmitter) PostVisitSliceExprLow(node ast.Expr, indent int) {
	cppe.emitToFileBuffer("..", "")
}

func (cppe *CSharpEmitter) PreVisitFuncLit(node *ast.FuncLit, indent int) {
	str := cppe.emitAsString("(", indent)
	cppe.emitToFileBuffer(str, "")
}
func (cppe *CSharpEmitter) PostVisitFuncLit(node *ast.FuncLit, indent int) {
	str := cppe.emitAsString("}", 0)
	cppe.emitToFileBuffer(str, "")
}

func (cppe *CSharpEmitter) PostVisitFuncLitTypeParams(node *ast.FieldList, indent int) {
	str := cppe.emitAsString(")", 0)
	str += cppe.emitAsString("=>", 0)
	cppe.emitToFileBuffer(str, "")
}

func (cppe *CSharpEmitter) PreVisitFuncLitTypeParam(node *ast.Field, index int, indent int) {
	str := ""
	if index > 0 {
		str += cppe.emitAsString(", ", 0)
	}
	cppe.emitToFileBuffer(str, "")
}

func (cppe *CSharpEmitter) PostVisitFuncLitTypeParam(node *ast.Field, index int, indent int) {
	str := cppe.emitAsString(" ", 0)
	str += cppe.emitAsString(node.Names[0].Name, indent)
	cppe.emitToFileBuffer(str, "")
}

func (cppe *CSharpEmitter) PreVisitFuncLitBody(node *ast.BlockStmt, indent int) {
	str := cppe.emitAsString("{\n", 0)
	cppe.emitToFileBuffer(str, "")
}

func (cppe *CSharpEmitter) PreVisitFuncLitTypeResults(node *ast.FieldList, indent int) {
	cppe.shouldGenerate = false
}

func (cppe *CSharpEmitter) PreVisitInterfaceType(node *ast.InterfaceType, indent int) {
	str := cppe.emitAsString("object", indent)
	cppe.stack = append(cppe.stack, str)
}

func (cppe *CSharpEmitter) PostVisitInterfaceType(node *ast.InterfaceType, indent int) {
	// emit only if it's not a complex type
	if len(cppe.stack) == 1 {
		cppe.emitToFileBuffer(cppe.stack[len(cppe.stack)-1], "")
		cppe.stack = cppe.stack[:len(cppe.stack)-1]
	}
}

func (cppe *CSharpEmitter) PreVisitKeyValueExprValue(node ast.Expr, indent int) {
	str := cppe.emitAsString("= ", 0)
	cppe.emitToFileBuffer(str, "")
}

func (cppe *CSharpEmitter) PreVisitUnaryExpr(node *ast.UnaryExpr, indent int) {
	str := cppe.emitAsString("(", 0)
	str += cppe.emitAsString(node.Op.String(), 0)
	cppe.emitToFileBuffer(str, "")
}
func (cppe *CSharpEmitter) PostVisitUnaryExpr(node *ast.UnaryExpr, indent int) {
	str := cppe.emitAsString(")", 0)
	cppe.emitToFileBuffer(str, "")
}

func (cppe *CSharpEmitter) PreVisitGenDeclConstName(node *ast.Ident, indent int) {
	// TODO dummy implementation
	// not very well performed
	for constIdent, obj := range cppe.pkg.TypesInfo.Defs {
		if obj == nil {
			continue
		}
		if con, ok := obj.(*types.Const); ok {
			if constIdent.Name != node.Name {
				continue
			}
			constType := con.Type().String()
			constType = strings.TrimPrefix(constType, "untyped ")
			str := cppe.emitAsString(fmt.Sprintf("public const %s %s = ", constType, node.Name), 0)

			cppe.emitToFileBuffer(str, "")
		}
	}
}
func (cppe *CSharpEmitter) PostVisitGenDeclConstName(node *ast.Ident, indent int) {
	str := cppe.emitAsString(";\n", 0)
	cppe.emitToFileBuffer(str, "")
}
func (cppe *CSharpEmitter) PostVisitGenDeclConst(node *ast.GenDecl, indent int) {
	str := cppe.emitAsString("\n", 0)
	cppe.emitToFileBuffer(str, "")
}

func (cppe *CSharpEmitter) PreVisitSwitchStmt(node *ast.SwitchStmt, indent int) {
	cppe.shouldGenerate = true
	str := cppe.emitAsString("switch (", indent)
	cppe.emitToFileBuffer(str, "")
}
func (cppe *CSharpEmitter) PostVisitSwitchStmt(node *ast.SwitchStmt, indent int) {
	str := cppe.emitAsString("}", indent)
	cppe.emitToFileBuffer(str, "")
}

func (cppe *CSharpEmitter) PostVisitSwitchStmtTag(node ast.Expr, indent int) {
	str := cppe.emitAsString(") {\n", 0)
	cppe.emitToFileBuffer(str, "")
}

func (cppe *CSharpEmitter) PostVisitCaseClause(node *ast.CaseClause, indent int) {
	cppe.emitToFileBuffer("\n", "")
	str := cppe.emitAsString("break;\n", indent+4)
	cppe.emitToFileBuffer(str, "")
}

func (cppe *CSharpEmitter) PostVisitCaseClauseList(node []ast.Expr, indent int) {
	if len(node) == 0 {
		str := cppe.emitAsString("default:\n", indent+2)
		cppe.emitToFileBuffer(str, "")
	}
}

func (cppe *CSharpEmitter) PreVisitCaseClauseListExpr(node ast.Expr, index int, indent int) {
	str := cppe.emitAsString("case ", indent+2)
	tv := cppe.pkg.TypesInfo.Types[node]
	if typeVal, ok := csTypesMap[tv.Type.String()]; ok {
		str += "(" + typeVal + ")"
	}
	cppe.emitToFileBuffer(str, "")
	cppe.shouldGenerate = true
}

func (cppe *CSharpEmitter) PostVisitCaseClauseListExpr(node ast.Expr, index int, indent int) {
	str := cppe.emitAsString(":\n", 0)
	cppe.emitToFileBuffer(str, "")
}

func (cppe *CSharpEmitter) PreVisitTypeAssertExprType(node ast.Expr, indent int) {
	str := cppe.emitAsString("(", indent)
	cppe.emitToFileBuffer(str, "")
}

func (cppe *CSharpEmitter) PostVisitTypeAssertExprType(node ast.Expr, indent int) {
	str := cppe.emitAsString(")", indent)
	cppe.emitToFileBuffer(str, "")
}

func (cppe *CSharpEmitter) PreVisitKeyValueExpr(node *ast.KeyValueExpr, indent int) {
	cppe.shouldGenerate = true
}

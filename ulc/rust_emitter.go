package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
	"os"
)

var rustDestTypes = []string{"sbyte", "short", "int", "long", "byte", "ushort", "object", "string"}

var rustTypesMap = map[string]string{
	"int8":   rustDestTypes[0],
	"int16":  rustDestTypes[1],
	"int32":  rustDestTypes[2],
	"int64":  rustDestTypes[3],
	"uint8":  rustDestTypes[4],
	"uint16": rustDestTypes[5],
	"any":    rustDestTypes[6],
	"string": rustDestTypes[7],
}

type RustEmitter struct {
	Output string
	file   *os.File
	Emitter
	pkg                   *packages.Package
	insideForPostCond     bool
	assignmentToken       string
	forwardDecls          bool
	shouldGenerate        bool
	numFuncResults        int
	aliases               map[string]Alias
	currentPackage        string
	stack                 []string
	buffer                bool
	isArray               bool
	arrayType             string
	isTuple               bool
	fileBuffer            string
	PointerAndPositionVec []PointerAndPosition
}

func (*RustEmitter) lowerToBuiltins(selector string) string {
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
func (re *RustEmitter) mergeStackElements(marker string) {
	var merged strings.Builder

	// Process the stack in reverse until we find a marker
	for len(re.stack) > 0 {
		top := re.stack[len(re.stack)-1]
		re.stack = re.stack[:len(re.stack)-1] // Pop element

		// Stop merging when we find a marker
		if strings.HasPrefix(top, marker) {
			re.stack = append(re.stack, merged.String()) // Push merged string
			return
		}

		// Prepend the element to the merged string (reverse order)
		mergedString := top + merged.String() // Prepend instead of append
		merged.Reset()
		merged.WriteString(mergedString)
	}

	panic("unreachable")
}

func (cse *RustEmitter) SearchPointerReverse(target string) *PointerAndPosition {
	for i := len(cse.PointerAndPositionVec) - 1; i >= 0; i-- {
		if cse.PointerAndPositionVec[i].Pointer == target {
			return &cse.PointerAndPositionVec[i]
		}
	}
	return nil // Return nil if the pointer is not found
}

func (cse *RustEmitter) ExtractSubstring(position int) (string, error) {
	if position < 0 || position >= len(cse.fileBuffer) {
		return "", fmt.Errorf("position %d is out of bounds", position)
	}
	return cse.fileBuffer[position:], nil
}

func (cse *RustEmitter) ExtractSubstringBetween(begin int, end int) (string, error) {
	if begin < 0 || end > len(cse.fileBuffer) || begin > end {
		return "", fmt.Errorf("invalid range: begin %d, end %d", begin, end)
	}
	return cse.fileBuffer[begin:end], nil
}

func (cse *RustEmitter) RewriteFileBufferBetween(begin int, end int, content string) error {
	if begin < 0 || end > len(cse.fileBuffer) || begin > end {
		return fmt.Errorf("invalid range: begin %d, end %d", begin, end)
	}
	cse.fileBuffer = cse.fileBuffer[:begin] + content + cse.fileBuffer[end:]
	return nil
}

func (cse *RustEmitter) RewriteFileBuffer(position int, oldContent, newContent string) error {
	if position < 0 || position+len(oldContent) > len(cse.fileBuffer) {
		return fmt.Errorf("position %d is out of bounds or oldContent does not match", position)
	}
	if cse.fileBuffer[position:position+len(oldContent)] != oldContent {
		return fmt.Errorf("oldContent does not match the existing content at position %d", position)
	}
	cse.fileBuffer = cse.fileBuffer[:position] + newContent + cse.fileBuffer[position+len(oldContent):]
	return nil
}

func (re *RustEmitter) emitToFileBuffer(s string, pointer string) error {
	re.PointerAndPositionVec = append(re.PointerAndPositionVec, PointerAndPosition{
		Pointer:  pointer,
		Position: len(re.fileBuffer),
	})
	re.fileBuffer += s
	return nil
}

func (re *RustEmitter) emitToFile() error {
	_, err := re.file.WriteString(re.fileBuffer)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return err
	}
	return nil
}

func (re *RustEmitter) emitAsString(s string, indent int) string {
	return strings.Repeat(" ", indent) + s
}

func (re *RustEmitter) PreVisitProgram(indent int) {
	re.PointerAndPositionVec = make([]PointerAndPosition, 0)
	re.aliases = make(map[string]Alias)
	outputFile := re.Output
	var err error
	re.file, err = os.Create(outputFile)
	re.SetFile(re.file)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	builtin := `use std::fmt;
use std::any::Any;

// Type aliases (Go-style)
type Int8 = i8;
type Int16 = i16;
type Int32 = i32;
type Int64 = i64;
type Uint8 = u8;
type Uint16 = u16;
type Uint32 = u32;
type Uint64 = u64;

// println! equivalent (already in std)
pub fn println<T: fmt::Display>(val: T) {
    println!("{}", val);
}

// printf! - partial simulation
pub fn printf<T: fmt::Display>(val: T) {
    print!("{}", val);
}

// Go-style append (returns a new Vec)
pub fn append<T: Clone>(vec: &Vec<T>, value: T) -> Vec<T> {
    let mut new_vec = vec.clone();
    new_vec.push(value);
    new_vec
}

pub fn append_many<T: Clone>(vec: &Vec<T>, values: &[T]) -> Vec<T> {
    let mut new_vec = vec.clone();
    new_vec.extend_from_slice(values);
    new_vec
}

// Simple string_format using format!
pub fn string_format(fmt_str: &str, args: &[&dyn fmt::Display]) -> String {
    let mut result = String::new();
    let mut split = fmt_str.split("{}");
    for (i, segment) in split.enumerate() {
        result.push_str(segment);
        if i < args.len() {
            result.push_str(&format!("{}", args[i]));
        }
    }
    result
}
`
	str := re.emitAsString(builtin, indent)
	re.emitToFileBuffer(str, "")

	re.insideForPostCond = false
}

func (re *RustEmitter) PostVisitProgram(indent int) {
	re.emitToFile()
	re.file.Close()
}

func (re *RustEmitter) PreVisitFuncDeclSignatures(indent int) {
	re.forwardDecls = true
}

func (re *RustEmitter) PostVisitFuncDeclSignatures(indent int) {
	re.forwardDecls = false
}

func (re *RustEmitter) PreVisitFuncDeclName(node *ast.Ident, indent int) {
	if re.forwardDecls {
		return
	}
	var str string
	str = re.emitAsString(fmt.Sprintf("fn %s", node.Name), 0)
	re.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PreVisitBlockStmt(node *ast.BlockStmt, indent int) {
	str := re.emitAsString("{\n", 1)
	re.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PostVisitBlockStmt(node *ast.BlockStmt, indent int) {
	str := re.emitAsString("}", 1)
	re.emitToFileBuffer(str, "")
	re.isArray = false
}

func (re *RustEmitter) PreVisitFuncDeclSignatureTypeParams(node *ast.FuncDecl, indent int) {
	if re.forwardDecls {
		return
	}
	re.shouldGenerate = true
	str := re.emitAsString("(", 0)
	re.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PostVisitFuncDeclSignatureTypeParams(node *ast.FuncDecl, indent int) {
	if re.forwardDecls {
		return
	}
	re.shouldGenerate = false
	str := re.emitAsString(")", 0)
	re.emitToFileBuffer(str, "")
}

func (cse *RustEmitter) PreVisitIdent(e *ast.Ident, indent int) {
	if !cse.shouldGenerate {
		return
	}

	var str string
	name := e.Name
	name = cse.lowerToBuiltins(name)
	if name == "nil" {
		str = cse.emitAsString("{}", indent)
	} else {
		if n, ok := csTypesMap[name]; ok {
			str = cse.emitAsString(n, indent)
		} else {
			str = cse.emitAsString(name, indent)
		}
	}

	if cse.buffer {
		cse.stack = append(cse.stack, str)
	} else {
		cse.emitToFileBuffer(str, "")
	}

}
func (re *RustEmitter) PreVisitCallExprArgs(node []ast.Expr, indent int) {
	str := re.emitAsString("(", 0)
	re.emitToFileBuffer(str, "")
}
func (re *RustEmitter) PostVisitCallExprArgs(node []ast.Expr, indent int) {
	str := re.emitAsString(")", 0)
	re.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PreVisitBasicLit(e *ast.BasicLit, indent int) {
	re.stack = append(re.stack, "@@PreVisitBasicLit")
	if e.Kind == token.STRING {
		e.Value = strings.Replace(e.Value, "\"", "", -1)
		if e.Value[0] == '`' {
			e.Value = strings.Replace(e.Value, "`", "", -1)
			str := (re.emitAsString(fmt.Sprintf("R\"(%s)\"", e.Value), 0))
			re.stack = append(re.stack, str)
		} else {
			str := (re.emitAsString(fmt.Sprintf("\"%s\"", e.Value), 0))
			re.stack = append(re.stack, str)
		}
	} else {
		str := (re.emitAsString(e.Value, 0))
		re.stack = append(re.stack, str)
	}
	re.buffer = true
}

func (re *RustEmitter) PostVisitBasicLit(e *ast.BasicLit, indent int) {
	re.mergeStackElements("@@PreVisitBasicLit")
	if len(re.stack) == 1 {
		re.emitToFileBuffer(re.stack[len(re.stack)-1], "")
		re.stack = re.stack[:len(re.stack)-1]
	}

	re.buffer = false
}

func (cse *RustEmitter) PreVisitDeclStmtValueSpecType(node *ast.ValueSpec, index int, indent int) {
	cse.emitToFileBuffer("", "@PreVisitDeclStmtValueSpecType")
}

func (cse *RustEmitter) PostVisitDeclStmtValueSpecType(node *ast.ValueSpec, index int, indent int) {
	pointerAndPosition := cse.SearchPointerReverse("@PreVisitDeclStmtValueSpecType")
	if pointerAndPosition != nil {
		for aliasName, alias := range cse.aliases {
			if alias.UnderlyingType == cse.pkg.TypesInfo.Types[node.Type].Type.Underlying().String() {
				cse.RewriteFileBufferBetween(pointerAndPosition.Position, len(cse.fileBuffer), aliasName)
			}
		}
	}
}

func (cse *RustEmitter) PreVisitDeclStmtValueSpecNames(node *ast.Ident, index int, indent int) {
	str := cse.emitAsString(" ", 0)
	cse.emitToFileBuffer(str, "")
}

func (cse *RustEmitter) PostVisitDeclStmtValueSpecNames(node *ast.Ident, index int, indent int) {
	var str string
	if cse.isArray {
		str += " = new "
		str += strings.TrimSpace(cse.arrayType)
		str += "();"
		cse.isArray = false
	} else {
		str += " = default;"
	}
	cse.emitToFileBuffer(str, "")
}

func (cse *RustEmitter) PreVisitGenStructFieldType(node ast.Expr, indent int) {
	str := cse.emitAsString("public", indent+2)
	cse.emitToFileBuffer(str, "")
}

func (cse *RustEmitter) PostVisitGenStructFieldType(node ast.Expr, indent int) {
	cse.emitToFileBuffer(" ", "")
	// clean array marker as we should generate
	// initializer only for expression statements
	// not for struct fields
	cse.isArray = false
}

func (cse *RustEmitter) PostVisitGenStructFieldName(node *ast.Ident, indent int) {
	cse.emitToFileBuffer(";\n", "")
}

func (cse *RustEmitter) PreVisitPackage(pkg *packages.Package, indent int) {
	cse.pkg = pkg
}

func (cse *RustEmitter) PostVisitPackage(pkg *packages.Package, indent int) {
}

func (cse *RustEmitter) PostVisitFuncDeclSignature(node *ast.FuncDecl, indent int) {
	cse.isArray = false
}

func (cse *RustEmitter) PostVisitBlockStmtList(node ast.Stmt, index int, indent int) {
	str := cse.emitAsString("\n", indent)
	cse.emitToFileBuffer(str, "")
}

func (cse *RustEmitter) PostVisitFuncDecl(node *ast.FuncDecl, indent int) {
	if cse.forwardDecls {
		return
	}
	str := cse.emitAsString("\n\n", 0)
	cse.emitToFileBuffer(str, "")
}

func (cse *RustEmitter) PreVisitGenStructInfo(node GenTypeInfo, indent int) {
	str := cse.emitAsString(fmt.Sprintf("public struct %s\n", node.Name), indent+2)
	str += cse.emitAsString("{\n", indent+2)
	cse.emitToFileBuffer(str, "")
	cse.shouldGenerate = true
}

func (cse *RustEmitter) PostVisitGenStructInfo(node GenTypeInfo, indent int) {
	str := cse.emitAsString("};\n\n", indent+2)
	cse.emitToFileBuffer(str, "")
	cse.shouldGenerate = false
}

func (cse *RustEmitter) PreVisitArrayType(node ast.ArrayType, indent int) {
	if !cse.shouldGenerate {
		return
	}
	cse.stack = append(cse.stack, "@@PreVisitArrayType")
	str := cse.emitAsString("List", indent)
	cse.stack = append(cse.stack, str)
	str = cse.emitAsString("<", indent)
	cse.stack = append(cse.stack, str)

	cse.buffer = true
}
func (cse *RustEmitter) PostVisitArrayType(node ast.ArrayType, indent int) {
	if !cse.shouldGenerate {
		return
	}

	cse.stack = append(cse.stack, cse.emitAsString(">", 0))

	cse.mergeStackElements("@@PreVisitArrayType")
	if len(cse.stack) == 1 {
		cse.isArray = true
		cse.arrayType = cse.stack[len(cse.stack)-1]
		cse.emitToFileBuffer(cse.stack[len(cse.stack)-1], "")
		cse.stack = cse.stack[:len(cse.stack)-1]
	}

	cse.buffer = false
}

func (cse *RustEmitter) PreVisitFuncType(node *ast.FuncType, indent int) {
	if !cse.shouldGenerate {
		return
	}
	cse.buffer = true
	cse.stack = append(cse.stack, "@@PreVisitFuncType")
	var str string
	if node.Results != nil {
		str = cse.emitAsString("Func<", indent)
	} else {
		str = cse.emitAsString("Action<", indent)
	}
	cse.stack = append(cse.stack, str)
}
func (cse *RustEmitter) PostVisitFuncType(node *ast.FuncType, indent int) {
	if !cse.shouldGenerate {
		return
	}

	// move return type to the end of the stack
	// return type is traversed first therefore it has to be moved
	// to the end of the stack due to C# syntax
	if len(cse.stack) > 2 && cse.numFuncResults > 0 {
		returnType := cse.stack[2]
		cse.stack = append(cse.stack[:2], cse.stack[3:]...)
		cse.stack = append(cse.stack, ",")
		cse.stack = append(cse.stack, returnType)
	}
	cse.stack = append(cse.stack, cse.emitAsString(">", 0))

	cse.mergeStackElements("@@PreVisitFuncType")

	if len(cse.stack) == 1 {
		cse.emitToFileBuffer(cse.stack[len(cse.stack)-1], "")
		cse.stack = cse.stack[:len(cse.stack)-1]
	}
	cse.buffer = false
}

func (cse *RustEmitter) PreVisitFuncTypeParam(node *ast.Field, index int, indent int) {
	if index > 0 {
		str := cse.emitAsString(", ", 0)
		cse.stack = append(cse.stack, str)
	}
}

func (cse *RustEmitter) PostVisitSelectorExprX(node ast.Expr, indent int) {
	if !cse.shouldGenerate {
		return
	}
	var str string
	scopeOperator := "."
	if ident, ok := node.(*ast.Ident); ok {
		if cse.lowerToBuiltins(ident.Name) == "" {
			return
		}
		// if the identifier is a package name, we need to append "Api." to the scope operator
		obj := cse.pkg.TypesInfo.Uses[ident]
		if obj != nil {
			if _, ok := obj.(*types.PkgName); ok {
				scopeOperator += "Api."
			}
		}
	}

	str = cse.emitAsString(scopeOperator, 0)
	if cse.buffer {
		cse.stack = append(cse.stack, str)
	} else {
		cse.emitToFileBuffer(str, "")
	}

}

func (cse *RustEmitter) PreVisitFuncTypeResults(node *ast.FieldList, indent int) {
	if node != nil {
		cse.numFuncResults = len(node.List)
	}
}

func (cse *RustEmitter) PreVisitFuncDeclSignatureTypeParamsList(node *ast.Field, index int, indent int) {
	if cse.forwardDecls {
		return
	}
	if index > 0 {
		str := cse.emitAsString(", ", 0)
		cse.emitToFileBuffer(str, "")
	}
}

func (cse *RustEmitter) PreVisitFuncDeclSignatureTypeParamsArgName(node *ast.Ident, index int, indent int) {
	if cse.forwardDecls {
		return
	}
	cse.emitToFileBuffer(" ", "")
}

func (cse *RustEmitter) PreVisitFuncDeclSignatureTypeResultsList(node *ast.Field, index int, indent int) {
	if cse.forwardDecls {
		return
	}
	if index > 0 {
		str := cse.emitAsString(",", 0)
		cse.emitToFileBuffer(str, "")
	}
	cse.emitToFileBuffer("", "@PreVisitFuncDeclSignatureTypeResultsList")
}

func (cse *RustEmitter) PostVisitFuncDeclSignatureTypeResultsList(node *ast.Field, index int, indent int) {
	if cse.forwardDecls {
		return
	}
	pointerAndPosition := cse.SearchPointerReverse("@PreVisitFuncDeclSignatureTypeResultsList")
	if pointerAndPosition != nil {
		for aliasName, alias := range cse.aliases {
			if alias.UnderlyingType == cse.pkg.TypesInfo.Types[node.Type].Type.Underlying().String() {
				cse.RewriteFileBufferBetween(pointerAndPosition.Position, len(cse.fileBuffer), aliasName)
			}
		}
	}
}

func (cse *RustEmitter) PreVisitFuncDeclSignatureTypeResults(node *ast.FuncDecl, indent int) {
	if cse.forwardDecls {
		return
	}

	cse.shouldGenerate = true

	if node.Type.Results != nil {
		if len(node.Type.Results.List) > 1 {
			str := cse.emitAsString("(", 0)
			cse.emitToFileBuffer(str, "")
		}
	}
}

func (cse *RustEmitter) PostVisitFuncDeclSignatureTypeResults(node *ast.FuncDecl, indent int) {
	if cse.forwardDecls {
		return
	}

	if node.Type.Results != nil {
		if len(node.Type.Results.List) > 1 {
			str := cse.emitAsString(")", 0)
			cse.emitToFileBuffer(str, "")
		}
	}

	str := cse.emitAsString("", 1)
	cse.emitToFileBuffer(str, "")
	cse.shouldGenerate = false
}

func (cse *RustEmitter) PreVisitTypeAliasName(node *ast.Ident, indent int) {
	cse.stack = append(cse.stack, "@@PreVisitTypeAliasName")
	cse.stack = append(cse.stack, cse.emitAsString("using ", indent+2))
	cse.shouldGenerate = true
	cse.buffer = true
}

func (cse *RustEmitter) PostVisitTypeAliasName(node *ast.Ident, indent int) {
	cse.buffer = true
	cse.stack = append(cse.stack, " = ")
}

func (cse *RustEmitter) PreVisitTypeAliasType(node ast.Expr, indent int) {

}

func (cse *RustEmitter) PostVisitTypeAliasType(node ast.Expr, indent int) {
	str := cse.emitAsString(";\n\n", 0)
	cse.stack = append(cse.stack, str)
	cse.aliases[cse.stack[2]] = Alias{
		PackageName:    cse.pkg.Name + ".Api",
		representation: ConvertToAliasRepr(ParseNestedTypes(cse.stack[4]), []string{"", cse.pkg.Name + ".Api"}),
		UnderlyingType: cse.pkg.TypesInfo.Types[node].Type.String(),
	}
	cse.mergeStackElements("@@PreVisitTypeAliasName")
	if len(cse.stack) == 1 {
		// TODO emit to aliases
		//cse.emitToFileBuffer(cse.stack[len(cse.stack)-1], "")
		cse.stack = cse.stack[:len(cse.stack)-1]
	}
	cse.shouldGenerate = false
	cse.buffer = false
}

func (cse *RustEmitter) PreVisitReturnStmt(node *ast.ReturnStmt, indent int) {
	cse.shouldGenerate = true
	str := cse.emitAsString("return ", indent)
	cse.emitToFileBuffer(str, "")

	if len(node.Results) == 1 {
		tv := cse.pkg.TypesInfo.Types[node.Results[0]]
		//pos := cse.pkg.Fset.Position(node.Pos())
		//fmt.Printf("@@Type: %s %s:%d:%d\n", tv.Type, pos.Filename, pos.Line, pos.Column)
		if typeVal, ok := csTypesMap[tv.Type.String()]; ok {
			if !cse.isTuple && tv.Type.String() != "func()" {
				cse.emitToFileBuffer("(", "")
				str := cse.emitAsString(typeVal, 0)
				cse.emitToFileBuffer(str, "")
				cse.emitToFileBuffer(")", "")
			}
		}
	}
	if len(node.Results) > 1 {
		str := cse.emitAsString("(", 0)
		cse.emitToFileBuffer(str, "")
	}
}

func (cse *RustEmitter) PostVisitReturnStmt(node *ast.ReturnStmt, indent int) {
	if len(node.Results) > 1 {
		str := cse.emitAsString(")", 0)
		cse.emitToFileBuffer(str, "")
	}
	str := cse.emitAsString(";", 0)
	cse.emitToFileBuffer(str, "")
	cse.shouldGenerate = false
}

func (cse *RustEmitter) PreVisitReturnStmtResult(node ast.Expr, index int, indent int) {
	if index > 0 {
		str := cse.emitAsString(", ", 0)
		cse.emitToFileBuffer(str, "")
	}
}

func (cse *RustEmitter) PreVisitCallExpr(node *ast.CallExpr, indent int) {
	cse.shouldGenerate = true
	cse.emitToFileBuffer("", "@PreVisitCallExpr")
}

func (cse *RustEmitter) PostVisitCallExpr(node *ast.CallExpr, indent int) {
	pointerAndPosition := cse.SearchPointerReverse("@PreVisitCallExpr")
	if pointerAndPosition != nil {
		str, _ := cse.ExtractSubstring(pointerAndPosition.Position)
		for _, t := range destTypes {
			matchStr := t + "("
			if strings.Contains(str, matchStr) {
				cse.RewriteFileBuffer(pointerAndPosition.Position, matchStr, "("+t+")(")
			}
		}
	}
	cse.shouldGenerate = false
}

func (cse *RustEmitter) PreVisitDeclStmt(node *ast.DeclStmt, indent int) {
	cse.shouldGenerate = true
}

func (cse *RustEmitter) PostVisitDeclStmt(node *ast.DeclStmt, indent int) {
	cse.shouldGenerate = false
}

func (cse *RustEmitter) PreVisitAssignStmt(node *ast.AssignStmt, indent int) {
	cse.shouldGenerate = true
	str := cse.emitAsString("", indent)
	cse.emitToFileBuffer(str, "")
}
func (cse *RustEmitter) PostVisitAssignStmt(node *ast.AssignStmt, indent int) {
	str := cse.emitAsString(";", 0)
	cse.emitToFileBuffer(str, "")
	cse.shouldGenerate = false
}

func (cse *RustEmitter) PreVisitAssignStmtRhs(node *ast.AssignStmt, indent int) {
	cse.shouldGenerate = true
	str := cse.emitAsString(cse.assignmentToken+" ", indent+1)
	cse.emitToFileBuffer(str, "")
}

func (cse *RustEmitter) PostVisitAssignStmtRhs(node *ast.AssignStmt, indent int) {
	cse.shouldGenerate = false
	cse.isTuple = false
}

func (cse *RustEmitter) PreVisitAssignStmtRhsExpr(node ast.Expr, index int, indent int) {
	cse.emitToFileBuffer("", "@PreVisitAssignStmtRhsExpr")
}

func (cse *RustEmitter) PostVisitAssignStmtRhsExpr(node ast.Expr, index int, indent int) {
	pointerAndPosition := cse.SearchPointerReverse("@PreVisitAssignStmtRhsExpr")
	rewritten := false
	if pointerAndPosition != nil {
		str, _ := cse.ExtractSubstring(pointerAndPosition.Position)
		for _, t := range destTypes {
			matchStr := t + "("
			if strings.Contains(str, matchStr) {
				cse.RewriteFileBuffer(pointerAndPosition.Position, matchStr, "("+t+")(")
				rewritten = true
			}
		}
	}
	if !rewritten {
		tv := cse.pkg.TypesInfo.Types[node]
		//pos := cse.pkg.Fset.Position(node.Pos())
		//fmt.Printf("@@Type: %s %s:%d:%d\n", tv.Type, pos.Filename, pos.Line, pos.Column)
		if typeVal, ok := csTypesMap[tv.Type.String()]; ok {
			if !cse.isTuple && tv.Type.String() != "func()" {
				cse.RewriteFileBuffer(pointerAndPosition.Position, "", "("+typeVal+")")
			}
		}
	}
}

func (cse *RustEmitter) PreVisitAssignStmtLhsExpr(node ast.Expr, index int, indent int) {
	if index > 0 {
		str := cse.emitAsString(", ", indent)
		cse.emitToFileBuffer(str, "")
	}
}

func (cse *RustEmitter) PreVisitAssignStmtLhs(node *ast.AssignStmt, indent int) {
	cse.shouldGenerate = true
	assignmentToken := node.Tok.String()
	if assignmentToken == ":=" && len(node.Lhs) == 1 {
		str := cse.emitAsString("var ", indent)
		cse.emitToFileBuffer(str, "")
	} else if assignmentToken == ":=" && len(node.Lhs) > 1 {
		str := cse.emitAsString("var (", indent)
		cse.emitToFileBuffer(str, "")
	} else if assignmentToken == "=" && len(node.Lhs) > 1 {
		str := cse.emitAsString("(", indent)
		cse.emitToFileBuffer(str, "")
		cse.isTuple = true
	}
	if assignmentToken != "+=" {
		assignmentToken = "="
	}
	cse.assignmentToken = assignmentToken
}

func (cse *RustEmitter) PostVisitAssignStmtLhs(node *ast.AssignStmt, indent int) {
	if node.Tok.String() == ":=" && len(node.Lhs) > 1 {
		str := cse.emitAsString(")", indent)
		cse.emitToFileBuffer(str, "")
	} else if node.Tok.String() == "=" && len(node.Lhs) > 1 {
		str := cse.emitAsString(")", indent)
		cse.emitToFileBuffer(str, "")
	}
	cse.shouldGenerate = false

}

func (cse *RustEmitter) PreVisitIndexExprIndex(node *ast.IndexExpr, indent int) {
	cse.shouldGenerate = true
	str := cse.emitAsString("[", 0)
	cse.emitToFileBuffer(str, "")

}
func (cse *RustEmitter) PostVisitIndexExprIndex(node *ast.IndexExpr, indent int) {
	str := cse.emitAsString("]", 0)
	cse.emitToFileBuffer(str, "")
}

func (cse *RustEmitter) PreVisitBinaryExpr(node *ast.BinaryExpr, indent int) {
	cse.shouldGenerate = true
	str := cse.emitAsString("(", 1)
	cse.emitToFileBuffer(str, "")
}
func (cse *RustEmitter) PostVisitBinaryExpr(node *ast.BinaryExpr, indent int) {
	str := cse.emitAsString(")", 1)
	cse.emitToFileBuffer(str, "")
	cse.shouldGenerate = false
}

func (cse *RustEmitter) PreVisitBinaryExprOperator(op token.Token, indent int) {
	str := cse.emitAsString(op.String()+" ", 1)
	cse.emitToFileBuffer(str, "")
}

func (cse *RustEmitter) PreVisitCallExprArg(node ast.Expr, index int, indent int) {
	if index > 0 {
		str := cse.emitAsString(", ", 0)
		cse.emitToFileBuffer(str, "")
	}
}
func (cse *RustEmitter) PostVisitExprStmtX(node ast.Expr, indent int) {
	str := cse.emitAsString(";", 0)
	cse.emitToFileBuffer(str, "")
}

func (cse *RustEmitter) PreVisitIfStmt(node *ast.IfStmt, indent int) {
	cse.shouldGenerate = true
}
func (cse *RustEmitter) PostVisitIfStmt(node *ast.IfStmt, indent int) {
	cse.shouldGenerate = false
}

func (cse *RustEmitter) PreVisitIfStmtCond(node *ast.IfStmt, indent int) {
	str := cse.emitAsString("if (", 1)
	cse.emitToFileBuffer(str, "")
}

func (cse *RustEmitter) PostVisitIfStmtCond(node *ast.IfStmt, indent int) {
	str := cse.emitAsString(")\n", 0)
	cse.emitToFileBuffer(str, "")
}

func (cse *RustEmitter) PreVisitForStmt(node *ast.ForStmt, indent int) {
	cse.insideForPostCond = true
	str := cse.emitAsString("for (", indent)
	cse.emitToFileBuffer(str, "")
	cse.shouldGenerate = true
}

func (cse *RustEmitter) PostVisitForStmtInit(node ast.Stmt, indent int) {
	if node == nil {
		str := cse.emitAsString(";", 0)
		cse.emitToFileBuffer(str, "")
	}
}

func (cse *RustEmitter) PostVisitForStmtPost(node ast.Stmt, indent int) {
	if node != nil {
		cse.insideForPostCond = false
	}
	str := cse.emitAsString(")\n", 0)
	cse.emitToFileBuffer(str, "")
}

func (cse *RustEmitter) PreVisitIfStmtElse(node *ast.IfStmt, indent int) {
	str := cse.emitAsString("else", 1)
	cse.emitToFileBuffer(str, "")
}

func (cse *RustEmitter) PostVisitForStmtCond(node ast.Expr, indent int) {
	str := cse.emitAsString(";", 0)
	cse.emitToFileBuffer(str, "")
	cse.shouldGenerate = false
}

func (cse *RustEmitter) PostVisitForStmt(node *ast.ForStmt, indent int) {
	cse.shouldGenerate = false
	cse.insideForPostCond = false
}

func (cse *RustEmitter) PreVisitRangeStmt(node *ast.RangeStmt, indent int) {
	cse.shouldGenerate = true
	str := cse.emitAsString("foreach (var ", indent)
	cse.emitToFileBuffer(str, "")
}

func (cse *RustEmitter) PostVisitRangeStmtValue(node ast.Expr, indent int) {
	str := cse.emitAsString(" in ", 0)
	cse.emitToFileBuffer(str, "")
}

func (cse *RustEmitter) PostVisitRangeStmtX(node ast.Expr, indent int) {
	str := cse.emitAsString(")\n", 0)
	cse.emitToFileBuffer(str, "")
	cse.shouldGenerate = false
}

func (cse *RustEmitter) PreVisitIncDecStmt(node *ast.IncDecStmt, indent int) {
	cse.shouldGenerate = true
}

func (cse *RustEmitter) PostVisitIncDecStmt(node *ast.IncDecStmt, indent int) {
	str := cse.emitAsString(node.Tok.String(), 0)
	if !cse.insideForPostCond {
		str += cse.emitAsString(";", 0)
	}
	cse.emitToFileBuffer(str, "")
	cse.shouldGenerate = false
}

func (cse *RustEmitter) PreVisitCompositeLitType(node ast.Expr, indent int) {
	str := cse.emitAsString("new ", 0)
	cse.emitToFileBuffer(str, "")
	cse.emitToFileBuffer("", "@PreVisitCompositeLitType")
}

func (cse *RustEmitter) PostVisitCompositeLitType(node ast.Expr, indent int) {
	pointerAndPosition := cse.SearchPointerReverse("@PreVisitCompositeLitType")
	if pointerAndPosition != nil {
		// TODO not very effective
		// go through all aliases and check if the underlying type matches
		for aliasName, alias := range cse.aliases {
			if alias.UnderlyingType == cse.pkg.TypesInfo.Types[node].Type.Underlying().String() {
				cse.RewriteFileBufferBetween(pointerAndPosition.Position, len(cse.fileBuffer), aliasName)
			}
		}
	}
}

func (cse *RustEmitter) PreVisitCompositeLitElts(node []ast.Expr, indent int) {
	str := cse.emitAsString("{", 0)
	cse.emitToFileBuffer(str, "")
}

func (cse *RustEmitter) PostVisitCompositeLitElts(node []ast.Expr, indent int) {
	str := cse.emitAsString("}", 0)
	cse.emitToFileBuffer(str, "")
}

func (cse *RustEmitter) PreVisitCompositeLitElt(node ast.Expr, index int, indent int) {
	if index > 0 {
		str := cse.emitAsString(", ", 0)
		cse.emitToFileBuffer(str, "")
	}
}

func (cse *RustEmitter) PostVisitSliceExprX(node ast.Expr, indent int) {
	str := cse.emitAsString("[", 0)
	cse.emitToFileBuffer(str, "")
	cse.shouldGenerate = false
}

func (cse *RustEmitter) PostVisitSliceExpr(node *ast.SliceExpr, indent int) {
	str := cse.emitAsString("]", 0)
	cse.emitToFileBuffer(str, "")
	cse.shouldGenerate = true
}

func (cse *RustEmitter) PostVisitSliceExprLow(node ast.Expr, indent int) {
	cse.emitToFileBuffer("..", "")
}

func (cse *RustEmitter) PreVisitFuncLit(node *ast.FuncLit, indent int) {
	str := cse.emitAsString("(", indent)
	cse.emitToFileBuffer(str, "")
}
func (cse *RustEmitter) PostVisitFuncLit(node *ast.FuncLit, indent int) {
	str := cse.emitAsString("}", 0)
	cse.emitToFileBuffer(str, "")
}

func (cse *RustEmitter) PostVisitFuncLitTypeParams(node *ast.FieldList, indent int) {
	str := cse.emitAsString(")", 0)
	str += cse.emitAsString("=>", 0)
	cse.emitToFileBuffer(str, "")
}

func (cse *RustEmitter) PreVisitFuncLitTypeParam(node *ast.Field, index int, indent int) {
	str := ""
	if index > 0 {
		str += cse.emitAsString(", ", 0)
	}
	cse.emitToFileBuffer(str, "")
}

func (cse *RustEmitter) PostVisitFuncLitTypeParam(node *ast.Field, index int, indent int) {
	str := cse.emitAsString(" ", 0)
	str += cse.emitAsString(node.Names[0].Name, indent)
	cse.emitToFileBuffer(str, "")
}

func (cse *RustEmitter) PreVisitFuncLitBody(node *ast.BlockStmt, indent int) {
	str := cse.emitAsString("{\n", 0)
	cse.emitToFileBuffer(str, "")
}

func (cse *RustEmitter) PreVisitFuncLitTypeResults(node *ast.FieldList, indent int) {
	cse.shouldGenerate = false
}

func (cse *RustEmitter) PreVisitInterfaceType(node *ast.InterfaceType, indent int) {
	str := cse.emitAsString("object", indent)
	cse.stack = append(cse.stack, str)
}

func (cse *RustEmitter) PostVisitInterfaceType(node *ast.InterfaceType, indent int) {
	// emit only if it's not a complex type
	if len(cse.stack) == 1 {
		cse.emitToFileBuffer(cse.stack[len(cse.stack)-1], "")
		cse.stack = cse.stack[:len(cse.stack)-1]
	}
}

func (cse *RustEmitter) PreVisitKeyValueExprValue(node ast.Expr, indent int) {
	str := cse.emitAsString("= ", 0)
	cse.emitToFileBuffer(str, "")
}

func (cse *RustEmitter) PreVisitUnaryExpr(node *ast.UnaryExpr, indent int) {
	str := cse.emitAsString("(", 0)
	str += cse.emitAsString(node.Op.String(), 0)
	cse.emitToFileBuffer(str, "")
}
func (cse *RustEmitter) PostVisitUnaryExpr(node *ast.UnaryExpr, indent int) {
	str := cse.emitAsString(")", 0)
	cse.emitToFileBuffer(str, "")
}

func (cse *RustEmitter) PreVisitGenDeclConstName(node *ast.Ident, indent int) {
	// TODO dummy implementation
	// not very well performed
	for constIdent, obj := range cse.pkg.TypesInfo.Defs {
		if obj == nil {
			continue
		}
		if con, ok := obj.(*types.Const); ok {
			if constIdent.Name != node.Name {
				continue
			}
			constType := con.Type().String()
			constType = strings.TrimPrefix(constType, "untyped ")
			if constType == cse.pkg.TypesInfo.Defs[node].Type().String() {
				constType = trimBeforeChar(constType, '.')
			}
			str := cse.emitAsString(fmt.Sprintf("public const %s %s = ", constType, node.Name), 0)

			cse.emitToFileBuffer(str, "")
		}
	}
}
func (cse *RustEmitter) PostVisitGenDeclConstName(node *ast.Ident, indent int) {
	str := cse.emitAsString(";\n", 0)
	cse.emitToFileBuffer(str, "")
}
func (cse *RustEmitter) PostVisitGenDeclConst(node *ast.GenDecl, indent int) {
	str := cse.emitAsString("\n", 0)
	cse.emitToFileBuffer(str, "")
}

func (cse *RustEmitter) PreVisitSwitchStmt(node *ast.SwitchStmt, indent int) {
	cse.shouldGenerate = true
	str := cse.emitAsString("switch (", indent)
	cse.emitToFileBuffer(str, "")
}
func (cse *RustEmitter) PostVisitSwitchStmt(node *ast.SwitchStmt, indent int) {
	str := cse.emitAsString("}", indent)
	cse.emitToFileBuffer(str, "")
}

func (cse *RustEmitter) PostVisitSwitchStmtTag(node ast.Expr, indent int) {
	str := cse.emitAsString(") {\n", 0)
	cse.emitToFileBuffer(str, "")
}

func (cse *RustEmitter) PostVisitCaseClause(node *ast.CaseClause, indent int) {
	cse.emitToFileBuffer("\n", "")
	str := cse.emitAsString("break;\n", indent+4)
	cse.emitToFileBuffer(str, "")
}

func (cse *RustEmitter) PostVisitCaseClauseList(node []ast.Expr, indent int) {
	if len(node) == 0 {
		str := cse.emitAsString("default:\n", indent+2)
		cse.emitToFileBuffer(str, "")
	}
}

func (cse *RustEmitter) PreVisitCaseClauseListExpr(node ast.Expr, index int, indent int) {
	str := cse.emitAsString("case ", indent+2)
	tv := cse.pkg.TypesInfo.Types[node]
	if typeVal, ok := csTypesMap[tv.Type.String()]; ok {
		str += "(" + typeVal + ")"
	}
	cse.emitToFileBuffer(str, "")
	cse.shouldGenerate = true
}

func (cse *RustEmitter) PostVisitCaseClauseListExpr(node ast.Expr, index int, indent int) {
	str := cse.emitAsString(":\n", 0)
	cse.emitToFileBuffer(str, "")
}

func (cse *RustEmitter) PreVisitTypeAssertExprType(node ast.Expr, indent int) {
	str := cse.emitAsString("(", indent)
	cse.emitToFileBuffer(str, "")
}

func (cse *RustEmitter) PostVisitTypeAssertExprType(node ast.Expr, indent int) {
	str := cse.emitAsString(")", indent)
	cse.emitToFileBuffer(str, "")
}

func (cse *RustEmitter) PreVisitKeyValueExpr(node *ast.KeyValueExpr, indent int) {
	cse.shouldGenerate = true
}

func (cse *RustEmitter) PreVisitBranchStmt(node *ast.BranchStmt, indent int) {
	str := cse.emitAsString(node.Tok.String()+";", indent)
	cse.emitToFileBuffer(str, "")
}

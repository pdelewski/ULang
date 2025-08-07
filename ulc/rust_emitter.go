package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"os"

	"golang.org/x/tools/go/packages"
)

var rustDestTypes = []string{"i8", "i16", "i32", "long", "byte", "ushort", "object", "string", "i32"}

var rustTypesMap = map[string]string{
	"int8":   rustDestTypes[0],
	"int16":  rustDestTypes[1],
	"int32":  rustDestTypes[2],
	"int64":  rustDestTypes[3],
	"uint8":  rustDestTypes[4],
	"uint16": rustDestTypes[5],
	"any":    rustDestTypes[6],
	"string": rustDestTypes[7],
	"int":    rustDestTypes[8],
}

type RustEmitter struct {
	Output string
	file   *os.File
	Emitter
	pkg               *packages.Package
	insideForPostCond bool
	assignmentToken   string
	forwardDecls      bool
	shouldGenerate    bool
	numFuncResults    int
	aliases           map[string]Alias
	currentPackage    string
	buffer            bool
	isArray           bool
	arrayType         string
	isTuple           bool
	gir               GoFIR
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
		return "len"
	}
	return selector
}

func (re *RustEmitter) emitAsString(s string, indent int) string {
	return strings.Repeat(" ", indent) + s
}

func (re *RustEmitter) PreVisitProgram(indent int) {
	re.gir.pointerAndPositionVec = make([]PointerAndPosition, 0)
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

pub fn len<T>(slice: &[T]) -> usize {
    slice.len()
}
`
	str := re.emitAsString(builtin, indent)
	re.gir.emitToFileBuffer(str, "")

	re.insideForPostCond = false
}

func (re *RustEmitter) PostVisitProgram(indent int) {
	emitToFile(re.file, re.gir.fileBuffer)
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
	re.gir.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PreVisitBlockStmt(node *ast.BlockStmt, indent int) {
	if re.forwardDecls {
		return
	}
	str := re.emitAsString("{\n", 1)
	re.gir.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PostVisitBlockStmt(node *ast.BlockStmt, indent int) {
	if re.forwardDecls {
		return
	}
	str := re.emitAsString("}", 1)
	re.gir.emitToFileBuffer(str, "")
	re.isArray = false
}

func (re *RustEmitter) PreVisitFuncDeclSignatureTypeParams(node *ast.FuncDecl, indent int) {
	if re.forwardDecls {
		return
	}
	re.shouldGenerate = true
	str := re.emitAsString("(", 0)
	re.gir.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PostVisitFuncDeclSignatureTypeParams(node *ast.FuncDecl, indent int) {
	if re.forwardDecls {
		return
	}
	re.shouldGenerate = false
	str := re.emitAsString(")", 0)
	re.gir.emitToFileBuffer(str, "")

	p1 := SearchPointerReverse("@PreVisitFuncDeclSignatureTypeResults", re.gir.pointerAndPositionVec)
	p2 := SearchPointerReverse("@PostVisitFuncDeclSignatureTypeResults", re.gir.pointerAndPositionVec)
	if p1 != nil && p2 != nil {
		results, err := ExtractSubstringBetween(p1.Position, p2.Position, re.gir.fileBuffer)
		if err != nil {
			fmt.Println("Error extracting results:", err)
			return
		}

		re.gir.fileBuffer, err = RewriteFileBufferBetween(re.gir.fileBuffer, p1.Position, p2.Position, "")
		if err != nil {
			fmt.Println("Error rewriting file buffer:", err)
			return
		}
		if strings.TrimSpace(results) != "" {
			re.gir.fileBuffer += " -> " + results
		}
	}
}

func (re *RustEmitter) PreVisitIdent(e *ast.Ident, indent int) {
	if re.forwardDecls {
		return
	}
	if !re.shouldGenerate {
		return
	}
	re.gir.emitToFileBuffer("", "@PreVisitIdent")

	var str string
	name := e.Name
	name = re.lowerToBuiltins(name)
	if name == "nil" {
		str = re.emitAsString("{}", indent)
	} else {
		if n, ok := rustTypesMap[name]; ok {
			str = re.emitAsString(n, indent)
		} else {
			str = re.emitAsString(name, indent)
		}
	}

	if re.buffer {
		re.gir.stack = append(re.gir.stack, str)
	} else {
		re.gir.emitToFileBuffer(str, "")
	}

}
func (re *RustEmitter) PreVisitCallExprArgs(node []ast.Expr, indent int) {
	if re.forwardDecls {
		return
	}
	str := re.emitAsString("(", 0)
	re.gir.emitToFileBuffer(str, "")
	p1 := SearchPointerReverse("@PreVisitCallExprFun", re.gir.pointerAndPositionVec)
	p2 := SearchPointerReverse("@PostVisitCallExprFun", re.gir.pointerAndPositionVec)
	if p1 != nil && p2 != nil {
		// Extract the substring between the positions of the pointers
		funName, err := ExtractSubstringBetween(p1.Position, p2.Position, re.gir.fileBuffer)
		if err != nil {
			fmt.Println("Error extracting function name:", err)
			return
		}
		if strings.Contains(funName, "len") {
			// add & before the first argument
			str := re.emitAsString("&", 0)
			re.gir.emitToFileBuffer(str, "")
		}
	}
}
func (re *RustEmitter) PostVisitCallExprArgs(node []ast.Expr, indent int) {
	if re.forwardDecls {
		return
	}
	str := re.emitAsString(")", 0)
	re.gir.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PreVisitBasicLit(e *ast.BasicLit, indent int) {
	if re.forwardDecls {
		return
	}
	re.gir.stack = append(re.gir.stack, "@@PreVisitBasicLit")
	var str string
	if e.Kind == token.STRING {
		e.Value = strings.Replace(e.Value, "\"", "", -1)
		if e.Value[0] == '`' {
			e.Value = strings.Replace(e.Value, "`", "", -1)
			str = (re.emitAsString(fmt.Sprintf("R\"(%s)\"", e.Value), 0))
		} else {
			str = (re.emitAsString(fmt.Sprintf("\"%s\"", e.Value), 0))
		}
	} else {
		str = (re.emitAsString(e.Value, 0))
	}
	re.gir.stack = append(re.gir.stack, str)
	re.buffer = true
}

func (re *RustEmitter) PostVisitBasicLit(e *ast.BasicLit, indent int) {
	if re.forwardDecls {
		return
	}
	re.gir.stack = mergeStackElements("@@PreVisitBasicLit", re.gir.stack)
	if len(re.gir.stack) == 1 {
		re.gir.emitToFileBuffer(re.gir.stack[len(re.gir.stack)-1], "")
		re.gir.stack = re.gir.stack[:len(re.gir.stack)-1]
	}

	re.buffer = false
}

func (re *RustEmitter) PreVisitDeclStmtValueSpecType(node *ast.ValueSpec, index int, indent int) {
	if re.forwardDecls {
		return
	}
	re.gir.emitToFileBuffer("", "@PreVisitDeclStmtValueSpecType")
}

func (re *RustEmitter) PostVisitDeclStmtValueSpecType(node *ast.ValueSpec, index int, indent int) {
	if re.forwardDecls {
		return
	}
	pointerAndPosition := SearchPointerReverse("@PreVisitDeclStmtValueSpecType", re.gir.pointerAndPositionVec)
	if pointerAndPosition != nil {
		for aliasName, alias := range re.aliases {
			if alias.UnderlyingType == re.pkg.TypesInfo.Types[node.Type].Type.Underlying().String() {
				re.gir.fileBuffer, _ = RewriteFileBufferBetween(re.gir.fileBuffer, pointerAndPosition.Position, len(re.gir.fileBuffer), aliasName)
			}
		}
	}
	str := re.emitAsString(" ", 0)
	re.gir.emitToFileBuffer(str, "")
	re.gir.emitToFileBuffer("", "@PostVisitDeclStmtValueSpecType")
}

func (re *RustEmitter) PreVisitDeclStmtValueSpecNames(node *ast.Ident, index int, indent int) {
	if re.forwardDecls {
		return
	}
	re.gir.emitToFileBuffer("", "@PreVisitDeclStmtValueSpecNames")
}

func (re *RustEmitter) PostVisitDeclStmtValueSpecNames(node *ast.Ident, index int, indent int) {
	if re.forwardDecls {
		return
	}
	re.gir.emitToFileBuffer("", "@PostVisitDeclStmtValueSpecNames")
	var str string
	if re.isArray {
		str += " = Vec::new();"
		re.isArray = false
	}
	re.gir.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PreVisitGenStructFieldType(node ast.Expr, indent int) {
	if re.forwardDecls {
		return
	}
	str := re.emitAsString("pub ", indent+2)
	re.gir.emitToFileBuffer(str, "")
	re.gir.emitToFileBuffer("", "@PreVisitGenStructFieldType")
}

func (re *RustEmitter) PostVisitGenStructFieldType(node ast.Expr, indent int) {
	if re.forwardDecls {
		return
	}
	re.gir.emitToFileBuffer("", "@PostVisitGenStructFieldType")
	re.gir.emitToFileBuffer(" ", "")
	// clean array marker as we should generate
	// initializer only for expression statements
	// not for struct fields
	re.isArray = false

}

func (re *RustEmitter) PreVisitGenStructFieldName(node *ast.Ident, indent int) {
	if re.forwardDecls {
		return
	}
	re.gir.emitToFileBuffer("", "@PreVisitGenStructFieldName")

}
func (re *RustEmitter) PostVisitGenStructFieldName(node *ast.Ident, indent int) {
	if re.forwardDecls {
		return
	}
	re.gir.emitToFileBuffer("", "@PostVisitGenStructFieldName")
	p1 := SearchPointerReverse("@PreVisitGenStructFieldType", re.gir.pointerAndPositionVec)
	p2 := SearchPointerReverse("@PostVisitGenStructFieldType", re.gir.pointerAndPositionVec)
	p3 := SearchPointerReverse("@PreVisitGenStructFieldName", re.gir.pointerAndPositionVec)
	p4 := SearchPointerReverse("@PostVisitGenStructFieldName", re.gir.pointerAndPositionVec)

	if p1 != nil && p2 != nil && p3 != nil && p4 != nil {
		fieldType, err := ExtractSubstringBetween(p1.Position, p2.Position, re.gir.fileBuffer)
		if err != nil {
			fmt.Println("Error extracting field type:", err)
			return
		}
		fieldName, err := ExtractSubstringBetween(p3.Position, p4.Position, re.gir.fileBuffer)
		if err != nil {
			fmt.Println("Error extracting field name:", err)
			return
		}
		re.gir.fileBuffer, err = RewriteFileBufferBetween(re.gir.fileBuffer, p1.Position, p4.Position, fieldName+":"+fieldType)
		if err != nil {
			fmt.Println("Error rewriting file buffer:", err)
			return
		}
	}

	re.gir.emitToFileBuffer(",\n", "")
}

func (re *RustEmitter) PreVisitPackage(pkg *packages.Package, indent int) {
	if re.forwardDecls {
		return
	}
	re.pkg = pkg
}

func (re *RustEmitter) PostVisitPackage(pkg *packages.Package, indent int) {
	if re.forwardDecls {
		return
	}
}

func (re *RustEmitter) PostVisitFuncDeclSignature(node *ast.FuncDecl, indent int) {
	if re.forwardDecls {
		return
	}
	re.isArray = false
}

func (re *RustEmitter) PostVisitBlockStmtList(node ast.Stmt, index int, indent int) {
	if re.forwardDecls {
		return
	}
	str := re.emitAsString("\n", indent)
	re.gir.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PostVisitFuncDecl(node *ast.FuncDecl, indent int) {
	if re.forwardDecls {
		return
	}
	str := re.emitAsString("\n\n", 0)
	re.gir.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PreVisitGenStructInfo(node GenTypeInfo, indent int) {
	if re.forwardDecls {
		return
	}
	str := re.emitAsString(fmt.Sprintf("pub struct %s\n", node.Name), indent+2)
	str += re.emitAsString("{\n", indent+2)
	re.gir.emitToFileBuffer(str, "")
	re.shouldGenerate = true
}

func (re *RustEmitter) PostVisitGenStructInfo(node GenTypeInfo, indent int) {
	if re.forwardDecls {
		return
	}
	str := re.emitAsString("}\n\n", indent+2)
	re.gir.emitToFileBuffer(str, "")
	re.shouldGenerate = false
}

func (re *RustEmitter) PreVisitArrayType(node ast.ArrayType, indent int) {
	if re.forwardDecls {
		return
	}
	re.gir.stack = append(re.gir.stack, "@@PreVisitArrayType")
	str := re.emitAsString("<", 0)
	re.gir.stack = append(re.gir.stack, str)

	re.buffer = true
}
func (re *RustEmitter) PostVisitArrayType(node ast.ArrayType, indent int) {
	if re.forwardDecls {
		return
	}

	re.gir.stack = append(re.gir.stack, re.emitAsString(">", 0))

	re.gir.stack = mergeStackElements("@@PreVisitArrayType", re.gir.stack)
	if len(re.gir.stack) == 1 {
		re.isArray = true
		re.arrayType = re.gir.stack[len(re.gir.stack)-1]
		re.gir.emitToFileBuffer("Vec", "")
		re.gir.emitToFileBuffer(re.gir.stack[len(re.gir.stack)-1], "")
		re.gir.stack = re.gir.stack[:len(re.gir.stack)-1]
	}

	re.buffer = false
}

func (re *RustEmitter) PreVisitFuncType(node *ast.FuncType, indent int) {
	if re.forwardDecls {
		return
	}
	re.buffer = true
	re.gir.stack = append(re.gir.stack, "@@PreVisitFuncType")
	var str string
	// TODO use Box<dyn Fn> for function types for now
	str = re.emitAsString("Box<dyn Fn(", indent)
	re.gir.stack = append(re.gir.stack, str)
}
func (re *RustEmitter) PostVisitFuncType(node *ast.FuncType, indent int) {
	if re.forwardDecls {
		return
	}

	// move return type to the end of the stack
	// return type is traversed first therefore it has to be moved
	// to the end of the stack due to C# syntax
	if len(re.gir.stack) > 2 && re.numFuncResults > 0 {
		returnType := re.gir.stack[2]
		re.gir.stack = append(re.gir.stack[:2], re.gir.stack[3:]...)
		re.gir.stack = append(re.gir.stack, ",")
		re.gir.stack = append(re.gir.stack, returnType)
	}
	re.gir.stack = append(re.gir.stack, re.emitAsString(")>", 0))

	re.gir.stack = mergeStackElements("@@PreVisitFuncType", re.gir.stack)

	if len(re.gir.stack) == 1 {
		re.gir.emitToFileBuffer(re.gir.stack[len(re.gir.stack)-1], "")
		re.gir.stack = re.gir.stack[:len(re.gir.stack)-1]
	}
	re.buffer = false
}

func (re *RustEmitter) PreVisitFuncTypeParam(node *ast.Field, index int, indent int) {
	if re.forwardDecls {
		return
	}
	if index > 0 {
		str := re.emitAsString(", ", 0)
		re.gir.stack = append(re.gir.stack, str)
	}
}

func (re *RustEmitter) PostVisitSelectorExprX(node ast.Expr, indent int) {
	if re.forwardDecls {
		return
	}
	var str string
	scopeOperator := "."
	if ident, ok := node.(*ast.Ident); ok {
		if re.lowerToBuiltins(ident.Name) == "" {
			return
		}
		// if the identifier is a package name, we need to append "Api." to the scope operator
		obj := re.pkg.TypesInfo.Uses[ident]
		if obj != nil {
			if _, ok := obj.(*types.PkgName); ok {
				scopeOperator += "Api."
			}
		}
	}

	str = re.emitAsString(scopeOperator, 0)
	if re.buffer {
		re.gir.stack = append(re.gir.stack, str)
	} else {
		re.gir.emitToFileBuffer(str, "")
	}

}

func (re *RustEmitter) PreVisitFuncTypeResults(node *ast.FieldList, indent int) {
	if re.forwardDecls {
		return
	}
	if node != nil {
		re.numFuncResults = len(node.List)
	}
}

func (re *RustEmitter) PreVisitFuncDeclSignatureTypeParamsList(node *ast.Field, index int, indent int) {
	if re.forwardDecls {
		return
	}
	if index > 0 {
		str := re.emitAsString(", ", 0)
		re.gir.emitToFileBuffer(str, "")
	}
}

func (re *RustEmitter) PreVisitFuncDeclSignatureTypeParamsArgName(node *ast.Ident, index int, indent int) {
	if re.forwardDecls {
		return
	}
	re.gir.emitToFileBuffer(" ", "")
	re.gir.emitToFileBuffer("", "@PreVisitFuncDeclSignatureTypeParamsArgName")
}

func (re *RustEmitter) PreVisitFuncDeclSignatureTypeResultsList(node *ast.Field, index int, indent int) {
	if re.forwardDecls {
		return
	}
	if index > 0 {
		str := re.emitAsString(",", 0)
		re.gir.emitToFileBuffer(str, "")
	}
	re.gir.emitToFileBuffer("", "@PreVisitFuncDeclSignatureTypeResultsList")
}

func (re *RustEmitter) PostVisitFuncDeclSignatureTypeResultsList(node *ast.Field, index int, indent int) {
	if re.forwardDecls {
		return
	}
	pointerAndPosition := SearchPointerReverse("@PreVisitFuncDeclSignatureTypeResultsList", re.gir.pointerAndPositionVec)
	if pointerAndPosition != nil {
		for aliasName, alias := range re.aliases {
			if alias.UnderlyingType == re.pkg.TypesInfo.Types[node.Type].Type.Underlying().String() {
				re.gir.fileBuffer, _ = RewriteFileBufferBetween(re.gir.fileBuffer, pointerAndPosition.Position, len(re.gir.fileBuffer), aliasName)
			}
		}
	}
}

func (re *RustEmitter) PreVisitFuncDeclSignatureTypeResults(node *ast.FuncDecl, indent int) {
	if re.forwardDecls {
		return
	}
	re.gir.emitToFileBuffer("", "@PreVisitFuncDeclSignatureTypeResults")

	if node.Type.Results != nil {
		if len(node.Type.Results.List) > 1 {
			str := re.emitAsString("(", 0)
			re.gir.emitToFileBuffer(str, "")
		}
	}
}

func (re *RustEmitter) PostVisitFuncDeclSignatureTypeResults(node *ast.FuncDecl, indent int) {
	if re.forwardDecls {
		return
	}

	if node.Type.Results != nil {
		if len(node.Type.Results.List) > 1 {
			str := re.emitAsString(")", 0)
			re.gir.emitToFileBuffer(str, "")
		}
	}

	str := re.emitAsString("", 1)
	re.gir.emitToFileBuffer(str, "")
	re.gir.emitToFileBuffer("", "@PostVisitFuncDeclSignatureTypeResults")
}

func (re *RustEmitter) PreVisitTypeAliasName(node *ast.Ident, indent int) {
	if re.forwardDecls {
		return
	}
	re.gir.stack = append(re.gir.stack, "@@PreVisitTypeAliasName")
	re.gir.stack = append(re.gir.stack, re.emitAsString("using ", indent+2))
	re.shouldGenerate = true
	re.buffer = true
}

func (re *RustEmitter) PostVisitTypeAliasName(node *ast.Ident, indent int) {
	if re.forwardDecls {
		return
	}
	re.buffer = true
	re.gir.stack = append(re.gir.stack, " = ")
}

func (re *RustEmitter) PreVisitTypeAliasType(node ast.Expr, indent int) {
	if re.forwardDecls {
		return
	}
}

func (re *RustEmitter) PostVisitTypeAliasType(node ast.Expr, indent int) {
	if re.forwardDecls {
		return
	}
	str := re.emitAsString(";\n\n", 0)
	re.gir.stack = append(re.gir.stack, str)
	re.aliases[re.gir.stack[2]] = Alias{
		PackageName:    re.pkg.Name + ".Api",
		representation: ConvertToAliasRepr(ParseNestedTypes(re.gir.stack[4]), []string{"", re.pkg.Name + ".Api"}),
		UnderlyingType: re.pkg.TypesInfo.Types[node].Type.String(),
	}
	re.gir.stack = mergeStackElements("@@PreVisitTypeAliasName", re.gir.stack)
	if len(re.gir.stack) == 1 {
		// TODO emit to aliases
		//cse.emitToFileBuffer(cse.stack[len(cse.stack)-1], "")
		re.gir.stack = re.gir.stack[:len(re.gir.stack)-1]
	}
	re.shouldGenerate = false
	re.buffer = false
}

func (re *RustEmitter) PreVisitReturnStmt(node *ast.ReturnStmt, indent int) {
	if re.forwardDecls {
		return
	}
	re.shouldGenerate = true
	str := re.emitAsString("return ", indent)
	re.gir.emitToFileBuffer(str, "")

	if len(node.Results) == 1 {
		tv := re.pkg.TypesInfo.Types[node.Results[0]]
		//pos := cse.pkg.Fset.Position(node.Pos())
		//fmt.Printf("@@Type: %s %s:%d:%d\n", tv.Type, pos.Filename, pos.Line, pos.Column)
		if typeVal, ok := rustTypesMap[tv.Type.String()]; ok {
			if !re.isTuple && tv.Type.String() != "func()" {
				re.gir.emitToFileBuffer("(", "")
				str := re.emitAsString(typeVal, 0)
				re.gir.emitToFileBuffer(str, "")
				re.gir.emitToFileBuffer(")", "")
			}
		}
	}
	if len(node.Results) > 1 {
		str := re.emitAsString("(", 0)
		re.gir.emitToFileBuffer(str, "")
	}
}

func (re *RustEmitter) PostVisitReturnStmt(node *ast.ReturnStmt, indent int) {
	if re.forwardDecls {
		return
	}
	if len(node.Results) > 1 {
		str := re.emitAsString(")", 0)
		re.gir.emitToFileBuffer(str, "")
	}
	str := re.emitAsString(";", 0)
	re.gir.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PreVisitReturnStmtResult(node ast.Expr, index int, indent int) {
	if index > 0 {
		str := re.emitAsString(", ", 0)
		re.gir.emitToFileBuffer(str, "")
	}
}

func (re *RustEmitter) PreVisitCallExpr(node *ast.CallExpr, indent int) {
	re.shouldGenerate = true
}

func (re *RustEmitter) PostVisitCallExpr(node *ast.CallExpr, indent int) {
	re.shouldGenerate = false
}

func (re *RustEmitter) PreVisitDeclStmt(node *ast.DeclStmt, indent int) {
	re.shouldGenerate = true
	str := re.emitAsString("let ", indent)
	re.gir.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PostVisitDeclStmt(node *ast.DeclStmt, indent int) {
	p1 := SearchPointerReverse("@PreVisitDeclStmtValueSpecType", re.gir.pointerAndPositionVec)
	p2 := SearchPointerReverse("@PostVisitDeclStmtValueSpecType", re.gir.pointerAndPositionVec)
	p3 := SearchPointerReverse("@PreVisitDeclStmtValueSpecNames", re.gir.pointerAndPositionVec)
	p4 := SearchPointerReverse("@PostVisitDeclStmtValueSpecNames", re.gir.pointerAndPositionVec)
	if p1 != nil && p2 != nil && p3 != nil && p4 != nil {
		// Extract the substring between the positions of the pointers
		fieldType, err := ExtractSubstringBetween(p1.Position, p2.Position, re.gir.fileBuffer)
		if err != nil {
			fmt.Println("Error extracting field type:", err)
			return
		}
		fieldName, err := ExtractSubstringBetween(p3.Position, p4.Position, re.gir.fileBuffer)
		if err != nil {
			fmt.Println("Error extracting field name:", err)
			return
		}
		re.gir.fileBuffer, err = RewriteFileBufferBetween(re.gir.fileBuffer, p1.Position, p4.Position, fieldName+":"+fieldType)
		if err != nil {
			fmt.Println("Error rewriting file buffer:", err)
			return
		}
	}
	re.shouldGenerate = false

}

func (re *RustEmitter) PreVisitAssignStmt(node *ast.AssignStmt, indent int) {
	re.shouldGenerate = true
	str := re.emitAsString("", indent)
	re.gir.emitToFileBuffer(str, "")
}
func (re *RustEmitter) PostVisitAssignStmt(node *ast.AssignStmt, indent int) {
	str := re.emitAsString(";", 0)
	re.gir.emitToFileBuffer(str, "")
	re.shouldGenerate = false
}

func (re *RustEmitter) PreVisitAssignStmtRhs(node *ast.AssignStmt, indent int) {
	re.shouldGenerate = true
	str := re.emitAsString(re.assignmentToken+" ", indent+1)
	re.gir.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PostVisitAssignStmtRhs(node *ast.AssignStmt, indent int) {
	re.shouldGenerate = false
	re.isTuple = false
}

func (re *RustEmitter) PreVisitAssignStmtLhsExpr(node ast.Expr, index int, indent int) {
	if index > 0 {
		str := re.emitAsString(", ", indent)
		re.gir.emitToFileBuffer(str, "")
	}
}

func (re *RustEmitter) PreVisitAssignStmtLhs(node *ast.AssignStmt, indent int) {
	re.shouldGenerate = true
	assignmentToken := node.Tok.String()
	if assignmentToken == ":=" && len(node.Lhs) == 1 {
		str := re.emitAsString("let ", indent)
		re.gir.emitToFileBuffer(str, "")
	} else if assignmentToken == ":=" && len(node.Lhs) > 1 {
		str := re.emitAsString("let (", indent)
		re.gir.emitToFileBuffer(str, "")
	} else if assignmentToken == "=" && len(node.Lhs) > 1 {
		str := re.emitAsString("(", indent)
		re.gir.emitToFileBuffer(str, "")
		re.isTuple = true
	}
	if assignmentToken != "+=" {
		assignmentToken = "="
	}
	re.assignmentToken = assignmentToken
}

func (re *RustEmitter) PostVisitAssignStmtLhs(node *ast.AssignStmt, indent int) {
	if node.Tok.String() == ":=" && len(node.Lhs) > 1 {
		str := re.emitAsString(")", indent)
		re.gir.emitToFileBuffer(str, "")
	} else if node.Tok.String() == "=" && len(node.Lhs) > 1 {
		str := re.emitAsString(")", indent)
		re.gir.emitToFileBuffer(str, "")
	}
	re.shouldGenerate = false

}

func (re *RustEmitter) PreVisitIndexExprIndex(node *ast.IndexExpr, indent int) {
	re.shouldGenerate = true
	str := re.emitAsString("[", 0)
	re.gir.emitToFileBuffer(str, "")

}
func (re *RustEmitter) PostVisitIndexExprIndex(node *ast.IndexExpr, indent int) {
	str := re.emitAsString("]", 0)
	re.gir.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PreVisitBinaryExpr(node *ast.BinaryExpr, indent int) {
	re.shouldGenerate = true
	str := re.emitAsString("(", 1)
	re.gir.emitToFileBuffer(str, "")
}
func (re *RustEmitter) PostVisitBinaryExpr(node *ast.BinaryExpr, indent int) {
	str := re.emitAsString(")", 1)
	re.gir.emitToFileBuffer(str, "")
	re.shouldGenerate = false
}

func (re *RustEmitter) PreVisitBinaryExprOperator(op token.Token, indent int) {
	str := re.emitAsString(op.String()+" ", 1)
	re.gir.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PreVisitCallExprArg(node ast.Expr, index int, indent int) {
	if index > 0 {
		str := re.emitAsString(", ", 0)
		re.gir.emitToFileBuffer(str, "")
	}
}
func (re *RustEmitter) PostVisitExprStmtX(node ast.Expr, indent int) {
	str := re.emitAsString(";", 0)
	re.gir.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PreVisitIfStmt(node *ast.IfStmt, indent int) {
	re.shouldGenerate = true
}
func (re *RustEmitter) PostVisitIfStmt(node *ast.IfStmt, indent int) {
	re.shouldGenerate = false
}

func (re *RustEmitter) PreVisitIfStmtCond(node *ast.IfStmt, indent int) {
	str := re.emitAsString("if (", 1)
	re.gir.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PostVisitIfStmtCond(node *ast.IfStmt, indent int) {
	str := re.emitAsString(")\n", 0)
	re.gir.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PreVisitForStmt(node *ast.ForStmt, indent int) {
	re.insideForPostCond = true
	str := re.emitAsString("for (", indent)
	re.gir.emitToFileBuffer(str, "")
	re.shouldGenerate = true
}

func (re *RustEmitter) PostVisitForStmtInit(node ast.Stmt, indent int) {
	if node == nil {
		str := re.emitAsString(";", 0)
		re.gir.emitToFileBuffer(str, "")
	}
}

func (re *RustEmitter) PostVisitForStmtPost(node ast.Stmt, indent int) {
	if node != nil {
		re.insideForPostCond = false
	}
	str := re.emitAsString(")\n", 0)
	re.gir.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PreVisitIfStmtElse(node *ast.IfStmt, indent int) {
	str := re.emitAsString("else", 1)
	re.gir.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PostVisitForStmtCond(node ast.Expr, indent int) {
	str := re.emitAsString(";", 0)
	re.gir.emitToFileBuffer(str, "")
	re.shouldGenerate = false
}

func (re *RustEmitter) PostVisitForStmt(node *ast.ForStmt, indent int) {
	re.shouldGenerate = false
	re.insideForPostCond = false
}

func (re *RustEmitter) PreVisitRangeStmt(node *ast.RangeStmt, indent int) {
	re.shouldGenerate = true
	str := re.emitAsString("foreach (var ", indent)
	re.gir.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PostVisitRangeStmtValue(node ast.Expr, indent int) {
	str := re.emitAsString(" in ", 0)
	re.gir.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PostVisitRangeStmtX(node ast.Expr, indent int) {
	str := re.emitAsString(")\n", 0)
	re.gir.emitToFileBuffer(str, "")
	re.shouldGenerate = false
}

func (re *RustEmitter) PreVisitIncDecStmt(node *ast.IncDecStmt, indent int) {
	re.shouldGenerate = true
}

func (re *RustEmitter) PostVisitIncDecStmt(node *ast.IncDecStmt, indent int) {
	str := re.emitAsString(node.Tok.String(), 0)
	if !re.insideForPostCond {
		str += re.emitAsString(";", 0)
	}
	re.gir.emitToFileBuffer(str, "")
	re.shouldGenerate = false
}

func (re *RustEmitter) PreVisitCompositeLitType(node ast.Expr, indent int) {
	re.gir.emitToFileBuffer("", "@PreVisitCompositeLitType")
}

func (re *RustEmitter) PostVisitCompositeLitType(node ast.Expr, indent int) {
	pointerAndPosition := SearchPointerReverse("@PreVisitCompositeLitType", re.gir.pointerAndPositionVec)
	if pointerAndPosition != nil {
		// TODO not very effective
		// go through all aliases and check if the underlying type matches
		for aliasName, alias := range re.aliases {
			if alias.UnderlyingType == re.pkg.TypesInfo.Types[node].Type.Underlying().String() {
				re.gir.fileBuffer, _ = RewriteFileBufferBetween(re.gir.fileBuffer, pointerAndPosition.Position, len(re.gir.fileBuffer), aliasName)
			}
		}
		if re.isArray {
			// TODO that's still hack
			// we operate on string representation of the type
			// has to be rewritten to use some kind of IR
			// We are trying to rewrite the type to a vector type
			// let x = Vec<> into let x: Vec<type> = vec![]
			vecTypeStrRepr, _ := ExtractSubstringBetween(pointerAndPosition.Position, len(re.gir.fileBuffer), re.gir.fileBuffer)
			re.gir.fileBuffer, _ = RewriteFileBufferBetween(re.gir.fileBuffer, pointerAndPosition.Position-len(" ="), len(re.gir.fileBuffer), ":"+vecTypeStrRepr+" = vec!")

		}
	}
}

func (re *RustEmitter) PreVisitCompositeLitElts(node []ast.Expr, indent int) {
	str := re.emitAsString("{", 0)
	re.gir.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PostVisitCompositeLitElts(node []ast.Expr, indent int) {
	str := re.emitAsString("}", 0)
	re.gir.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PreVisitCompositeLitElt(node ast.Expr, index int, indent int) {
	if index > 0 {
		str := re.emitAsString(", ", 0)
		re.gir.emitToFileBuffer(str, "")
	}
}

func (re *RustEmitter) PostVisitSliceExprX(node ast.Expr, indent int) {
	str := re.emitAsString("[", 0)
	re.gir.emitToFileBuffer(str, "")
	re.shouldGenerate = false
}

func (re *RustEmitter) PostVisitSliceExpr(node *ast.SliceExpr, indent int) {
	str := re.emitAsString("]", 0)
	re.gir.emitToFileBuffer(str, "")
	re.shouldGenerate = true
}

func (re *RustEmitter) PostVisitSliceExprLow(node ast.Expr, indent int) {
	re.gir.emitToFileBuffer("..", "")
}

func (re *RustEmitter) PreVisitFuncLit(node *ast.FuncLit, indent int) {
	str := re.emitAsString("(", indent)
	re.gir.emitToFileBuffer(str, "")
}
func (re *RustEmitter) PostVisitFuncLit(node *ast.FuncLit, indent int) {
	str := re.emitAsString("}", 0)
	re.gir.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PostVisitFuncLitTypeParams(node *ast.FieldList, indent int) {
	str := re.emitAsString(")", 0)
	str += re.emitAsString("=>", 0)
	re.gir.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PreVisitFuncLitTypeParam(node *ast.Field, index int, indent int) {
	str := ""
	if index > 0 {
		str += re.emitAsString(", ", 0)
	}
	re.gir.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PostVisitFuncLitTypeParam(node *ast.Field, index int, indent int) {
	str := re.emitAsString(" ", 0)
	str += re.emitAsString(node.Names[0].Name, indent)
	re.gir.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PreVisitFuncLitBody(node *ast.BlockStmt, indent int) {
	str := re.emitAsString("{\n", 0)
	re.gir.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PreVisitFuncLitTypeResults(node *ast.FieldList, indent int) {
	re.shouldGenerate = false
}

func (re *RustEmitter) PreVisitInterfaceType(node *ast.InterfaceType, indent int) {
	str := re.emitAsString("object", indent)
	re.gir.stack = append(re.gir.stack, str)
}

func (re *RustEmitter) PostVisitInterfaceType(node *ast.InterfaceType, indent int) {
	// emit only if it's not a complex type
	if len(re.gir.stack) == 1 {
		re.gir.emitToFileBuffer(re.gir.stack[len(re.gir.stack)-1], "")
		re.gir.stack = re.gir.stack[:len(re.gir.stack)-1]
	}
}

func (re *RustEmitter) PreVisitKeyValueExprValue(node ast.Expr, indent int) {
	str := re.emitAsString("= ", 0)
	re.gir.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PreVisitUnaryExpr(node *ast.UnaryExpr, indent int) {
	str := re.emitAsString("(", 0)
	str += re.emitAsString(node.Op.String(), 0)
	re.gir.emitToFileBuffer(str, "")
}
func (re *RustEmitter) PostVisitUnaryExpr(node *ast.UnaryExpr, indent int) {
	str := re.emitAsString(")", 0)
	re.gir.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PreVisitGenDeclConstName(node *ast.Ident, indent int) {
	// TODO dummy implementation
	// not very well performed
	for constIdent, obj := range re.pkg.TypesInfo.Defs {
		if obj == nil {
			continue
		}
		if con, ok := obj.(*types.Const); ok {
			if constIdent.Name != node.Name {
				continue
			}
			constType := con.Type().String()
			constType = strings.TrimPrefix(constType, "untyped ")
			if constType == re.pkg.TypesInfo.Defs[node].Type().String() {
				constType = trimBeforeChar(constType, '.')
			}
			str := re.emitAsString(fmt.Sprintf("public const %s %s = ", constType, node.Name), 0)

			re.gir.emitToFileBuffer(str, "")
		}
	}
}
func (re *RustEmitter) PostVisitGenDeclConstName(node *ast.Ident, indent int) {
	str := re.emitAsString(";\n", 0)
	re.gir.emitToFileBuffer(str, "")
}
func (re *RustEmitter) PostVisitGenDeclConst(node *ast.GenDecl, indent int) {
	str := re.emitAsString("\n", 0)
	re.gir.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PreVisitSwitchStmt(node *ast.SwitchStmt, indent int) {
	re.shouldGenerate = true
	str := re.emitAsString("switch (", indent)
	re.gir.emitToFileBuffer(str, "")
}
func (re *RustEmitter) PostVisitSwitchStmt(node *ast.SwitchStmt, indent int) {
	str := re.emitAsString("}", indent)
	re.gir.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PostVisitSwitchStmtTag(node ast.Expr, indent int) {
	str := re.emitAsString(") {\n", 0)
	re.gir.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PostVisitCaseClause(node *ast.CaseClause, indent int) {
	re.gir.emitToFileBuffer("\n", "")
	str := re.emitAsString("break;\n", indent+4)
	re.gir.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PostVisitCaseClauseList(node []ast.Expr, indent int) {
	if len(node) == 0 {
		str := re.emitAsString("default:\n", indent+2)
		re.gir.emitToFileBuffer(str, "")
	}
}

func (re *RustEmitter) PreVisitCaseClauseListExpr(node ast.Expr, index int, indent int) {
	str := re.emitAsString("case ", indent+2)
	tv := re.pkg.TypesInfo.Types[node]
	if typeVal, ok := rustTypesMap[tv.Type.String()]; ok {
		str += "(" + typeVal + ")"
	}
	re.gir.emitToFileBuffer(str, "")
	re.shouldGenerate = true
}

func (re *RustEmitter) PostVisitCaseClauseListExpr(node ast.Expr, index int, indent int) {
	str := re.emitAsString(":\n", 0)
	re.gir.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PreVisitTypeAssertExprType(node ast.Expr, indent int) {
	str := re.emitAsString("(", indent)
	re.gir.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PostVisitTypeAssertExprType(node ast.Expr, indent int) {
	str := re.emitAsString(")", indent)
	re.gir.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PreVisitKeyValueExpr(node *ast.KeyValueExpr, indent int) {
	re.shouldGenerate = true
}

func (re *RustEmitter) PreVisitBranchStmt(node *ast.BranchStmt, indent int) {
	str := re.emitAsString(node.Tok.String()+";", indent)
	re.gir.emitToFileBuffer(str, "")
}

func (re *RustEmitter) PreVisitCallExprFun(node ast.Expr, indent int) {
	re.gir.emitToFileBuffer("", "@PreVisitCallExprFun")
}

func (re *RustEmitter) PostVisitCallExprFun(node ast.Expr, indent int) {
	re.gir.emitToFileBuffer("", "@PostVisitCallExprFun")
}

func (re *RustEmitter) PreVisitFuncDeclSignatureTypeParamsListType(node ast.Expr, argName *ast.Ident, index int, indent int) {
	re.gir.emitToFileBuffer("", "@PreVisitFuncDeclSignatureTypeParamsListType")
}

func (re *RustEmitter) PostVisitFuncDeclSignatureTypeParamsListType(node ast.Expr, argName *ast.Ident, index int, indent int) {
	re.gir.emitToFileBuffer("", "@PostVisitFuncDeclSignatureTypeParamsListType")
}

func (re *RustEmitter) PostVisitFuncDeclSignatureTypeParamsArgName(node *ast.Ident, index int, indent int) {
	re.gir.emitToFileBuffer("", "@PostVisitFuncDeclSignatureTypeParamsArgName")
}

func (re *RustEmitter) PostVisitFuncDeclSignatureTypeParamsList(node *ast.Field, index int, indent int) {
	p1 := SearchPointerReverse("@PreVisitFuncDeclSignatureTypeParamsListType", re.gir.pointerAndPositionVec)
	p2 := SearchPointerReverse("@PostVisitFuncDeclSignatureTypeParamsListType", re.gir.pointerAndPositionVec)
	p3 := SearchPointerReverse("@PreVisitFuncDeclSignatureTypeParamsArgName", re.gir.pointerAndPositionVec)
	p4 := SearchPointerReverse("@PostVisitFuncDeclSignatureTypeParamsArgName", re.gir.pointerAndPositionVec)

	if p1 != nil && p2 != nil && p3 != nil && p4 != nil {
		typeStrRepr, err := ExtractSubstringBetween(p1.Position, p2.Position, re.gir.fileBuffer)
		if err != nil {
			fmt.Println("Error extracting type representation:", err)
			return
		}
		nameStrRepr, err := ExtractSubstringBetween(p3.Position, p4.Position, re.gir.fileBuffer)
		if err != nil {
			fmt.Println("Error extracting name representation:", err)
			return
		}
		if containsWhitespace(nameStrRepr) {
			fmt.Println("Error: Type parameter name contains whitespace")
			return
		}
		if containsWhitespace(typeStrRepr) {
			fmt.Println("Error: Type parameter type contains whitespace")
			return
		}
		re.gir.fileBuffer, err = RewriteFileBufferBetween(re.gir.fileBuffer, p1.Position, p4.Position, nameStrRepr+":"+typeStrRepr)
		if err != nil {
			fmt.Println("Error rewriting file buffer:", err)
			return
		}
	}
}

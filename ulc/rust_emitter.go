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

var rustDestTypes = []string{"i8", "i16", "i32", "i64", "u8", "u16", "Box<dyn Any>", "String", "i32"}

var rustTypesMap = map[string]string{
	"int8":    rustDestTypes[0],
	"int16":   rustDestTypes[1],
	"int32":   rustDestTypes[2],
	"int64":   rustDestTypes[3],
	"uint8":   rustDestTypes[4],
	"uint16":  rustDestTypes[5],
	"uint32":  "u32",
	"uint64":  "u64",
	"any":     rustDestTypes[6],
	"string":  rustDestTypes[7],
	"int":     rustDestTypes[8],
	"bool":    "bool",
	"float32": "f32",
	"float64": "f64",
}

// mapGoTypeToRust converts a Go type string to its Rust equivalent
func (re *RustEmitter) mapGoTypeToRust(goType string) string {
	if rustType, ok := rustTypesMap[goType]; ok {
		return rustType
	}
	// Return the original if not found in map
	return goType
}

type RustEmitter struct {
	Output string
	file   *os.File
	BaseEmitter
	pkg                  *packages.Package
	insideForPostCond    bool
	assignmentToken      string
	forwardDecls         bool
	shouldGenerate       bool
	numFuncResults       int
	aliases              map[string]Alias
	currentPackage       string
	isArray              bool
	arrayType            string
	isTuple              bool
	sawIncrement         bool   // Track if we saw ++ in for loop post statement
	declType             string // Store the type for multi-name declarations
	declNameCount        int    // Count of names in current declaration
	declNameIndex        int    // Current name index
	inAssignRhs                       bool   // Track if we're in assignment RHS
	pkgHasInterfaceTypes              bool   // Track if current package has any interface{} types
	currentCompLitTypeNoDefault       bool   // Track if current composite literal's type doesn't derive Default
	processedPkgsInterfaceTypes       map[string]bool // Cache for package interface{} type checks
	inKeyValueExpr                    bool   // Track if we're inside a KeyValueExpr (struct field init)
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

// Helper function to determine token type for Rust specific content
func (re *RustEmitter) getTokenType(content string) TokenType {
	// Check for Rust keywords
	switch content {
	case "fn", "let", "mut", "impl", "trait", "mod", "use", "pub", "struct", "enum", "match", "if", "else", "loop", "while", "for", "in", "return", "break", "continue":
		return RustKeyword
	case "(":
		return LeftParen
	case ")":
		return RightParen
	case "{":
		return LeftBrace
	case "}":
		return RightBrace
	case "[":
		return LeftBracket
	case "]":
		return RightBracket
	case ";":
		return Semicolon
	case ",":
		return Comma
	case ".":
		return Dot
	case "=", "+=", "-=", "*=", "/=":
		return Assignment
	case "+", "-", "*", "/", "%":
		return ArithmeticOperator
	case "==", "!=", "<", ">", "<=", ">=":
		return ComparisonOperator
	case "&&", "||", "!":
		return LogicalOperator
	case "++":
		return UnaryOperator
	case " ", "\t":
		return WhiteSpace
	case "\n":
		return NewLine
	}

	// Check if it's a number
	if len(content) > 0 && (content[0] >= '0' && content[0] <= '9') {
		return NumberLiteral
	}

	// Check if it's a string literal
	if len(content) >= 2 && content[0] == '"' && content[len(content)-1] == '"' {
		return StringLiteral
	}

	// Default to identifier
	return Identifier
}

// Helper function to emit token
func (re *RustEmitter) emitToken(content string, tokenType TokenType, indent int) {
	token := CreateToken(tokenType, re.emitAsString(content, indent))
	_ = re.gir.emitTokenToFileBuffer(token, EmptyVisitMethod)
}

// Helper function to convert []Token to []string for backward compatibility
func tokensToStrings(tokens []Token) []string {
	result := make([]string, len(tokens))
	for i, token := range tokens {
		result[i] = token.Content
	}
	return result
}

func (re *RustEmitter) SetFile(file *os.File) {
	re.file = file
}

func (re *RustEmitter) GetFile() *os.File {
	return re.file
}

func (re *RustEmitter) PreVisitProgram(indent int) {
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

// println equivalents - multiple versions for different arg counts
pub fn println<T: fmt::Display>(val: T) {
    std::println!("{}", val);
}

pub fn println0() {
    std::println!();
}

// printf - multiple versions for different arg counts
pub fn printf<T: fmt::Display>(val: T) {
    print!("{}", val);
}

pub fn printf2<T: fmt::Display>(_fmt: String, val: T) {
    // Simplified: ignoring format string, just printing value
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

pub fn len<T>(slice: &[T]) -> i32 {
    slice.len() as i32
}
`
	str := re.emitAsString(builtin, indent)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)

	re.insideForPostCond = false
}

func (re *RustEmitter) PostVisitProgram(indent int) {
	emitTokensToFile(re.file, re.gir.tokenSlice)
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
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
}

func (re *RustEmitter) PreVisitBlockStmt(node *ast.BlockStmt, indent int) {
	if re.forwardDecls {
		return
	}
	re.emitToken("{", LeftBrace, 1)
	str := re.emitAsString("\n", 0)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
}

func (re *RustEmitter) PostVisitBlockStmt(node *ast.BlockStmt, indent int) {
	if re.forwardDecls {
		return
	}
	re.emitToken("}", RightBrace, 1)
	re.isArray = false
}

func (re *RustEmitter) PreVisitFuncDeclSignatureTypeParams(node *ast.FuncDecl, indent int) {
	if re.forwardDecls {
		return
	}
	re.shouldGenerate = true
	re.emitToken("(", LeftParen, 0)
}

func (re *RustEmitter) PostVisitFuncDeclSignatureTypeParams(node *ast.FuncDecl, indent int) {
	if re.forwardDecls {
		return
	}
	re.shouldGenerate = false
	re.emitToken(")", RightParen, 0)

	p1 := SearchPointerIndexReverse("@PreVisitFuncDeclSignatureTypeResults", re.gir.pointerAndIndexVec)
	p2 := SearchPointerIndexReverse("@PostVisitFuncDeclSignatureTypeResults", re.gir.pointerAndIndexVec)
	if p1 != nil && p2 != nil {
		results, err := ExtractTokensBetween(p1.Index, p2.Index, re.gir.tokenSlice)
		if err != nil {
			fmt.Println("Error extracting results:", err)
			return
		}

		re.gir.tokenSlice, err = RewriteTokensBetween(re.gir.tokenSlice, p1.Index, p2.Index, []string{""})
		if err != nil {
			fmt.Println("Error rewriting file buffer:", err)
			return
		}
		if strings.TrimSpace(strings.Join(tokensToStrings(results), "")) != "" {
			re.gir.tokenSlice = append(re.gir.tokenSlice, CreateToken(RustKeyword, " -> "))
			re.gir.tokenSlice = append(re.gir.tokenSlice, CreateToken(Identifier, strings.Join(tokensToStrings(results), "")))
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
		str = re.emitAsString("None", indent)
	} else {
		if n, ok := rustTypesMap[name]; ok {
			str = re.emitAsString(n, indent)
		} else {
			str = re.emitAsString(name, indent)
		}
	}

	re.emitToken(str, Identifier, 0)

}
func (re *RustEmitter) PreVisitCallExprArgs(node []ast.Expr, indent int) {
	if re.forwardDecls {
		return
	}
	re.emitToken("(", LeftParen, 0)
	p1 := SearchPointerIndexReverse("@PreVisitCallExprFun", re.gir.pointerAndIndexVec)
	p2 := SearchPointerIndexReverse("@PostVisitCallExprFun", re.gir.pointerAndIndexVec)
	if p1 != nil && p2 != nil {
		// Extract the substring between the positions of the pointers
		funName, err := ExtractTokensBetween(p1.Index, p2.Index, re.gir.tokenSlice)
		if err != nil {
			fmt.Println("Error extracting function name:", err)
			return
		}
		funNameStr := strings.Join(tokensToStrings(funName), "")
		if strings.Contains(funNameStr, "len") || strings.Contains(funNameStr, "append") {
			// add & before the first argument for len and append
			str := re.emitAsString("&", 0)
			re.gir.emitToFileBuffer(str, EmptyVisitMethod)
		}
	}
}
func (re *RustEmitter) PostVisitCallExprArgs(node []ast.Expr, indent int) {
	if re.forwardDecls {
		return
	}
	// Rewrite printf/println calls based on argument count
	p1 := SearchPointerIndexReverse("@PreVisitCallExprFun", re.gir.pointerAndIndexVec)
	p2 := SearchPointerIndexReverse("@PostVisitCallExprFun", re.gir.pointerAndIndexVec)
	if p1 != nil && p2 != nil {
		funName, err := ExtractTokensBetween(p1.Index, p2.Index, re.gir.tokenSlice)
		if err == nil {
			funNameStr := strings.Join(tokensToStrings(funName), "")
			// Handle println with 0 args
			if funNameStr == "println" && len(node) == 0 {
				re.gir.tokenSlice, _ = RewriteTokensBetween(re.gir.tokenSlice, p1.Index, p2.Index, []string{"println0"})
			}
			// Handle printf with 2 args (format string + value)
			if funNameStr == "printf" && len(node) == 2 {
				re.gir.tokenSlice, _ = RewriteTokensBetween(re.gir.tokenSlice, p1.Index, p2.Index, []string{"printf2"})
			}
		}
	}
	re.emitToken(")", RightParen, 0)
}

func (re *RustEmitter) PreVisitBasicLit(e *ast.BasicLit, indent int) {
	if re.forwardDecls {
		return
	}
	var str string
	if e.Kind == token.STRING {
		e.Value = strings.Replace(e.Value, "\"", "", -1)
		if e.Value[0] == '`' {
			e.Value = strings.Replace(e.Value, "`", "", -1)
			str = (re.emitAsString(fmt.Sprintf("r#\"%s\"#", e.Value), 0))
		} else {
			str = (re.emitAsString(fmt.Sprintf("\"%s\"", e.Value), 0))
		}
		re.emitToken(str, StringLiteral, 0)
	} else {
		str = (re.emitAsString(e.Value, 0))
		re.emitToken(str, NumberLiteral, 0)
	}
}

func (re *RustEmitter) PostVisitBasicLit(e *ast.BasicLit, indent int) {
	if re.forwardDecls {
		return
	}
	// For string literals, add .to_string() to convert &str to String
	if e.Kind == token.STRING {
		re.gir.emitToFileBuffer(".to_string()", EmptyVisitMethod)
	}
}

func (re *RustEmitter) PreVisitDeclStmtValueSpecType(node *ast.ValueSpec, index int, indent int) {
	if re.forwardDecls {
		return
	}
	// For second and subsequent names, start a new let statement
	if index > 0 {
		re.emitToken(";", Semicolon, 0)
		re.emitToken("\n", NewLine, 0)
		re.emitToken("let", RustKeyword, indent)
		re.emitToken(" ", WhiteSpace, 0)
		re.emitToken("mut", RustKeyword, 0)
		re.emitToken(" ", WhiteSpace, 0)
	}
	re.gir.emitToFileBuffer("", "@PreVisitDeclStmtValueSpecType")
}

func (re *RustEmitter) PostVisitDeclStmtValueSpecType(node *ast.ValueSpec, index int, indent int) {
	if re.forwardDecls {
		return
	}
	pointerAndPosition := SearchPointerIndexReverse("@PreVisitDeclStmtValueSpecType", re.gir.pointerAndIndexVec)
	if pointerAndPosition != nil {
		for aliasName, alias := range re.aliases {
			if alias.UnderlyingType == re.pkg.TypesInfo.Types[node.Type].Type.Underlying().String() {
				re.gir.tokenSlice, _ = RewriteTokensBetween(re.gir.tokenSlice, pointerAndPosition.Index, len(re.gir.tokenSlice), []string{aliasName})
			}
		}
	}
	str := re.emitAsString(" ", 0)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
	re.gir.emitToFileBuffer("", "@PostVisitDeclStmtValueSpecType")
}

func (re *RustEmitter) PreVisitDeclStmtValueSpecNames(node *ast.Ident, index int, indent int) {
	if re.forwardDecls {
		return
	}
	re.declNameIndex = index
	re.gir.emitToFileBuffer("", "@PreVisitDeclStmtValueSpecNames")
}

func (re *RustEmitter) PostVisitDeclStmtValueSpecNames(node *ast.Ident, index int, indent int) {
	if re.forwardDecls {
		return
	}
	// Reorder tokens: swap type and name to get "name: type" format
	// This needs to be done for each name-type pair
	p1 := SearchPointerIndexReverse("@PreVisitDeclStmtValueSpecType", re.gir.pointerAndIndexVec)
	p2 := SearchPointerIndexReverse("@PostVisitDeclStmtValueSpecType", re.gir.pointerAndIndexVec)
	p3 := SearchPointerIndexReverse("@PreVisitDeclStmtValueSpecNames", re.gir.pointerAndIndexVec)
	if p1 != nil && p2 != nil && p3 != nil {
		// Extract the type tokens
		fieldType, err := ExtractTokensBetween(p1.Index, p2.Index, re.gir.tokenSlice)
		if err == nil && len(fieldType) > 0 {
			// Extract the name tokens (from p3 to end)
			fieldName, err := ExtractTokensBetween(p3.Index, len(re.gir.tokenSlice), re.gir.tokenSlice)
			if err == nil && len(fieldName) > 0 {
				// Build new tokens: name: type
				newTokens := []string{}
				newTokens = append(newTokens, tokensToStrings(fieldName)...)
				newTokens = append(newTokens, ":")
				newTokens = append(newTokens, tokensToStrings(fieldType)...)
				re.gir.tokenSlice, _ = RewriteTokensBetween(re.gir.tokenSlice, p1.Index, len(re.gir.tokenSlice), newTokens)
			}
		}
	}
	re.gir.emitToFileBuffer("", "@PostVisitDeclStmtValueSpecNames")
	var str string
	if re.isArray {
		str += " = Vec::new();"
		re.isArray = false
	}
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
}

func (re *RustEmitter) PreVisitGenStructFieldType(node ast.Expr, indent int) {
	if re.forwardDecls {
		return
	}
	str := re.emitAsString("pub ", indent+2)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
	re.gir.emitToFileBuffer("", "@PreVisitGenStructFieldType")
}

func (re *RustEmitter) PostVisitGenStructFieldType(node ast.Expr, indent int) {
	if re.forwardDecls {
		return
	}
	re.gir.emitToFileBuffer("", "@PostVisitGenStructFieldType")
	re.gir.emitToFileBuffer(" ", EmptyVisitMethod)
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
	p1 := SearchPointerIndexReverse("@PreVisitGenStructFieldType", re.gir.pointerAndIndexVec)
	p2 := SearchPointerIndexReverse("@PostVisitGenStructFieldType", re.gir.pointerAndIndexVec)
	p3 := SearchPointerIndexReverse("@PreVisitGenStructFieldName", re.gir.pointerAndIndexVec)
	p4 := SearchPointerIndexReverse("@PostVisitGenStructFieldName", re.gir.pointerAndIndexVec)

	if p1 != nil && p2 != nil && p3 != nil && p4 != nil {
		fieldType, err := ExtractTokensBetween(p1.Index, p2.Index, re.gir.tokenSlice)
		if err != nil {
			fmt.Println("Error extracting field type:", err)
			return
		}
		fieldName, err := ExtractTokensBetween(p3.Index, p4.Index, re.gir.tokenSlice)
		if err != nil {
			fmt.Println("Error extracting field name:", err)
			return
		}
		newTokens := []string{}
		newTokens = append(newTokens, tokensToStrings(fieldName)...)
		newTokens = append(newTokens, ":")
		newTokens = append(newTokens, tokensToStrings(fieldType)...)
		re.gir.tokenSlice, err = RewriteTokensBetween(re.gir.tokenSlice, p1.Index, p4.Index, newTokens)
		if err != nil {
			fmt.Println("Error rewriting file buffer:", err)
			return
		}
	}

	re.gir.emitToFileBuffer(",\n", EmptyVisitMethod)
}

func (re *RustEmitter) PreVisitPackage(pkg *packages.Package, indent int) {
	if re.forwardDecls {
		return
	}
	re.pkg = pkg
	// Initialize the cache if not already done
	if re.processedPkgsInterfaceTypes == nil {
		re.processedPkgsInterfaceTypes = make(map[string]bool)
	}
	// Check if package has any interface{} types
	re.pkgHasInterfaceTypes = re.packageHasInterfaceTypes(pkg)
	// Cache this package's result
	re.processedPkgsInterfaceTypes[pkg.PkgPath] = re.pkgHasInterfaceTypes
}

// packageHasInterfaceTypes scans all structs in the package for interface{} fields
func (re *RustEmitter) packageHasInterfaceTypes(pkg *packages.Package) bool {
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			if genDecl, ok := decl.(*ast.GenDecl); ok {
				for _, spec := range genDecl.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						if structType, ok := typeSpec.Type.(*ast.StructType); ok {
							if structType.Fields != nil {
								for _, field := range structType.Fields.List {
									if field.Type != nil {
										typeStr := pkg.TypesInfo.Types[field.Type].Type.String()
										if strings.Contains(typeStr, "interface{}") || strings.Contains(typeStr, "interface {") {
											return true
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return false
}

// typeHasInterfaceFields checks if a type contains interface{} fields (directly or transitively)
func (re *RustEmitter) typeHasInterfaceFields(t types.Type) bool {
	// Get the underlying type
	underlying := t.Underlying()
	if structType, ok := underlying.(*types.Struct); ok {
		for i := 0; i < structType.NumFields(); i++ {
			field := structType.Field(i)
			fieldTypeStr := field.Type().String()
			if strings.Contains(fieldTypeStr, "interface{}") || strings.Contains(fieldTypeStr, "interface {") {
				return true
			}
			// Check nested structs recursively
			if re.typeHasInterfaceFields(field.Type()) {
				return true
			}
		}
	}
	return false
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
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
}

func (re *RustEmitter) PostVisitFuncDecl(node *ast.FuncDecl, indent int) {
	if re.forwardDecls {
		return
	}
	str := re.emitAsString("\n\n", 0)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
}

func (re *RustEmitter) PreVisitGenStructInfo(node GenTypeInfo, indent int) {
	if re.forwardDecls {
		return
	}
	// Check if this specific struct has interface{} fields (can't derive Clone/Default)
	var str string
	if re.structHasInterfaceFields(node.Name) {
		// Only derive Debug for structs with Any/interface{} fields
		str = re.emitAsString("#[derive(Debug)]\n", indent+2)
	} else {
		// Check if struct only has primitive/Copy types (can derive Copy)
		canCopy := re.structCanDeriveCopy(node.Name)
		if canCopy {
			str = re.emitAsString("#[derive(Default, Clone, Copy, Debug)]\n", indent+2)
		} else {
			// Add derive macros for Default (needed for ..Default::default() in struct init)
			str = re.emitAsString("#[derive(Default, Clone, Debug)]\n", indent+2)
		}
	}
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
	str = re.emitAsString(fmt.Sprintf("pub struct %s\n", node.Name), indent+2)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
	re.emitToken("{", LeftBrace, indent+2)
	str = re.emitAsString("\n", 0)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
	re.shouldGenerate = true
}

// structHasInterfaceFields checks if a struct has interface{} fields (directly or in nested structs)
func (re *RustEmitter) structHasInterfaceFields(structName string) bool {
	return re.structHasInterfaceFieldsRecursive(structName, make(map[string]bool))
}

// structHasInterfaceFieldsRecursive checks recursively if a struct has interface{} fields
func (re *RustEmitter) structHasInterfaceFieldsRecursive(structName string, visited map[string]bool) bool {
	// Prevent infinite recursion
	if visited[structName] {
		return false
	}
	visited[structName] = true

	for _, file := range re.pkg.Syntax {
		for _, decl := range file.Decls {
			if genDecl, ok := decl.(*ast.GenDecl); ok {
				for _, spec := range genDecl.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						if typeSpec.Name.Name == structName {
							if structType, ok := typeSpec.Type.(*ast.StructType); ok {
								if structType.Fields != nil {
									for _, field := range structType.Fields.List {
										fieldType := re.pkg.TypesInfo.Types[field.Type].Type
										typeStr := fieldType.String()
										// Direct interface{} check
										if strings.Contains(typeStr, "interface{}") || strings.Contains(typeStr, "interface {") {
											return true
										}
										// Check nested struct fields recursively
										if named, ok := fieldType.(*types.Named); ok {
											if _, isStruct := named.Underlying().(*types.Struct); isStruct {
												nestedName := named.Obj().Name()
												if re.structHasInterfaceFieldsRecursive(nestedName, visited) {
													return true
												}
											}
										}
										// Check slice element type
										if slice, ok := fieldType.(*types.Slice); ok {
											elemType := slice.Elem()
											if named, ok := elemType.(*types.Named); ok {
												if _, isStruct := named.Underlying().(*types.Struct); isStruct {
													nestedName := named.Obj().Name()
													if re.structHasInterfaceFieldsRecursive(nestedName, visited) {
														return true
													}
												}
											}
										}
									}
								}
								return false
							}
						}
					}
				}
			}
		}
	}
	return false
}

// structCanDeriveCopy checks if a struct only contains primitive/Copy fields
func (re *RustEmitter) structCanDeriveCopy(structName string) bool {
	for _, file := range re.pkg.Syntax {
		for _, decl := range file.Decls {
			if genDecl, ok := decl.(*ast.GenDecl); ok {
				for _, spec := range genDecl.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						if typeSpec.Name.Name == structName {
							if structType, ok := typeSpec.Type.(*ast.StructType); ok {
								if structType.Fields != nil {
									for _, field := range structType.Fields.List {
										fieldType := re.pkg.TypesInfo.Types[field.Type].Type
										typeStr := fieldType.String()
										// If field is a slice/array, String, or interface{}, can't derive Copy
										if strings.HasPrefix(typeStr, "[]") ||
											typeStr == "string" ||
											strings.Contains(typeStr, "interface") {
											return false
										}
										// If field is a struct type, can't safely derive Copy
										// (the nested struct might have non-Copy fields)
										if named, ok := fieldType.(*types.Named); ok {
											if _, isStruct := named.Underlying().(*types.Struct); isStruct {
												return false
											}
										}
									}
								}
								return true
							}
						}
					}
				}
			}
		}
	}
	return false
}

func (re *RustEmitter) PostVisitGenStructInfo(node GenTypeInfo, indent int) {
	if re.forwardDecls {
		return
	}
	re.emitToken("}", RightBrace, indent+2)
	str := re.emitAsString("\n\n", 0)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
	re.shouldGenerate = false
}

func (re *RustEmitter) PreVisitArrayType(node ast.ArrayType, indent int) {
	if re.forwardDecls {
		return
	}
	re.gir.emitToFileBuffer("", "@@PreVisitArrayType")
	re.emitToken("<", LeftAngle, 0)
}
func (re *RustEmitter) PostVisitArrayType(node ast.ArrayType, indent int) {
	if re.forwardDecls {
		return
	}

	re.emitToken(">", RightAngle, 0)

	pointerAndPosition := SearchPointerIndexReverse("@@PreVisitArrayType", re.gir.pointerAndIndexVec)
	if pointerAndPosition != nil {
		tokens, _ := ExtractTokens(pointerAndPosition.Index, re.gir.tokenSlice)
		re.isArray = true
		re.arrayType = strings.Join(tokens, "")
		// Prepend "Vec" before the array type tokens
		re.gir.tokenSlice, _ = RewriteTokens(re.gir.tokenSlice, pointerAndPosition.Index, []string{}, []string{"Vec"})
	}
}

func (re *RustEmitter) PreVisitFuncType(node *ast.FuncType, indent int) {
	if re.forwardDecls {
		return
	}
	re.gir.emitToFileBuffer("", "@@PreVisitFuncType")
	// TODO use Box<dyn Fn> for function types for now
	str := re.emitAsString("Box<dyn Fn(", indent)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
}
func (re *RustEmitter) PostVisitFuncType(node *ast.FuncType, indent int) {
	if re.forwardDecls {
		return
	}

	pointerAndPosition := SearchPointerIndexReverse("@@PreVisitFuncType", re.gir.pointerAndIndexVec)
	if pointerAndPosition != nil && re.numFuncResults > 0 {
		// For function types with return values, we need to reorder tokens
		// to move return type to the end (Rust syntax requirement)
		tokens, _ := ExtractTokens(pointerAndPosition.Index, re.gir.tokenSlice)
		if len(tokens) > 2 {
			// Find and move return type to end with arrow separator
			var reorderedTokens []string
			reorderedTokens = append(reorderedTokens, tokens[0]) // "Box<dyn Fn("
			if len(tokens) > 3 {
				// Skip return type (index 1) and add parameters first
				reorderedTokens = append(reorderedTokens, tokens[2:]...)
				reorderedTokens = append(reorderedTokens, ") -> ")
				reorderedTokens = append(reorderedTokens, tokens[1]) // Add return type at end
				reorderedTokens = append(reorderedTokens, ">")
			} else {
				reorderedTokens = append(reorderedTokens, tokens[1:]...)
				reorderedTokens = append(reorderedTokens, ")>")
			}
			re.gir.tokenSlice, _ = RewriteTokensBetween(re.gir.tokenSlice, pointerAndPosition.Index, len(re.gir.tokenSlice), reorderedTokens)
			return
		}
	}

	re.emitToken(")", RightParen, 0)
	re.emitToken(">", RightAngle, 0)
}

func (re *RustEmitter) PreVisitFuncTypeParam(node *ast.Field, index int, indent int) {
	if re.forwardDecls {
		return
	}
	if index > 0 {
		str := re.emitAsString(", ", 0)
		re.gir.emitToFileBuffer(str, EmptyVisitMethod)
	}
}

func (re *RustEmitter) PreVisitSelectorExprX(node ast.Expr, indent int) {
	// For package names, suppress generation since we're generating single-file output
	if ident, ok := node.(*ast.Ident); ok {
		obj := re.pkg.TypesInfo.Uses[ident]
		if obj != nil {
			if _, ok := obj.(*types.PkgName); ok {
				// Don't generate the package name
				re.shouldGenerate = false
				return
			}
		}
	}
}

func (re *RustEmitter) PostVisitSelectorExprX(node ast.Expr, indent int) {
	if re.forwardDecls {
		return
	}
	var str string
	scopeOperator := "." // Default to dot for field access
	if ident, ok := node.(*ast.Ident); ok {
		if re.lowerToBuiltins(ident.Name) == "" {
			// Re-enable generation for the selector part (e.g., Printf after fmt)
			re.shouldGenerate = true
			return
		}
		// Check if this is a package name - skip operator for single-file output
		obj := re.pkg.TypesInfo.Uses[ident]
		if obj != nil {
			if _, ok := obj.(*types.PkgName); ok {
				// For single-file output, don't emit any scope operator for package references
				// The type/function will be referenced directly
				re.shouldGenerate = true // Re-enable for the selector part
				return
			}
		}
	}

	str = re.emitAsString(scopeOperator, 0)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)

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
		re.gir.emitToFileBuffer(str, EmptyVisitMethod)
	}
}

func (re *RustEmitter) PreVisitFuncDeclSignatureTypeParamsArgName(node *ast.Ident, index int, indent int) {
	if re.forwardDecls {
		return
	}
	re.gir.emitToFileBuffer(" ", EmptyVisitMethod)
	re.gir.emitToFileBuffer("", "@PreVisitFuncDeclSignatureTypeParamsArgName")
}

func (re *RustEmitter) PreVisitFuncDeclSignatureTypeResultsList(node *ast.Field, index int, indent int) {
	if re.forwardDecls {
		return
	}
	if index > 0 {
		str := re.emitAsString(",", 0)
		re.gir.emitToFileBuffer(str, EmptyVisitMethod)
	}
	re.gir.emitToFileBuffer("", "@PreVisitFuncDeclSignatureTypeResultsList")
}

func (re *RustEmitter) PostVisitFuncDeclSignatureTypeResultsList(node *ast.Field, index int, indent int) {
	if re.forwardDecls {
		return
	}
	pointerAndPosition := SearchPointerIndexReverse("@PreVisitFuncDeclSignatureTypeResultsList", re.gir.pointerAndIndexVec)
	if pointerAndPosition != nil {
		for aliasName, alias := range re.aliases {
			if alias.UnderlyingType == re.pkg.TypesInfo.Types[node.Type].Type.Underlying().String() {
				re.gir.tokenSlice, _ = RewriteTokensBetween(re.gir.tokenSlice, pointerAndPosition.Index, len(re.gir.tokenSlice), []string{aliasName})
			}
		}
	}
}

func (re *RustEmitter) PreVisitFuncDeclSignatureTypeResults(node *ast.FuncDecl, indent int) {
	if re.forwardDecls {
		return
	}
	re.gir.emitToFileBuffer("", "@PreVisitFuncDeclSignatureTypeResults")
	re.shouldGenerate = true // Enable generating result types

	if node.Type.Results != nil {
		if len(node.Type.Results.List) > 1 {
			re.emitToken("(", LeftParen, 0)
		}
	}
}

func (re *RustEmitter) PostVisitFuncDeclSignatureTypeResults(node *ast.FuncDecl, indent int) {
	if re.forwardDecls {
		return
	}

	if node.Type.Results != nil {
		if len(node.Type.Results.List) > 1 {
			re.emitToken(")", RightParen, 0)
		}
	}

	str := re.emitAsString("", 1)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
	re.gir.emitToFileBuffer("", "@PostVisitFuncDeclSignatureTypeResults")
}

func (re *RustEmitter) PreVisitTypeAliasName(node *ast.Ident, indent int) {
	if re.forwardDecls {
		return
	}
	re.gir.emitToFileBuffer("", "@@PreVisitTypeAliasName")
	str := re.emitAsString("type ", indent+2)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
	re.shouldGenerate = true
}

func (re *RustEmitter) PostVisitTypeAliasName(node *ast.Ident, indent int) {
	if re.forwardDecls {
		return
	}
	re.gir.emitToFileBuffer(" = ", EmptyVisitMethod)
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
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)

	// Extract tokens for alias processing
	pointerAndPosition := SearchPointerIndexReverse("@@PreVisitTypeAliasName", re.gir.pointerAndIndexVec)
	if pointerAndPosition != nil {
		tokens, _ := ExtractTokens(pointerAndPosition.Index, re.gir.tokenSlice)
		if len(tokens) >= 3 {
			// tokens[0] = "type ", tokens[1] = alias name, tokens[2] = " = ", tokens[3+] = type
			aliasName := tokens[1]
			typeTokens := tokens[3 : len(tokens)-1] // exclude the ";\n\n" at the end
			typeStr := strings.Join(typeTokens, "")
			re.aliases[aliasName] = Alias{
				PackageName:    re.pkg.Name + ".Api",
				representation: ConvertToAliasRepr(ParseNestedTypes(typeStr), []string{"", re.pkg.Name + ".Api"}),
				UnderlyingType: re.pkg.TypesInfo.Types[node].Type.String(),
			}
		}
	}
	re.shouldGenerate = false
}

func (re *RustEmitter) PreVisitReturnStmt(node *ast.ReturnStmt, indent int) {
	if re.forwardDecls {
		return
	}
	re.shouldGenerate = true
	str := re.emitAsString("return ", indent)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)

	if len(node.Results) > 1 {
		re.emitToken("(", LeftParen, 0)
	}
}

func (re *RustEmitter) PostVisitReturnStmt(node *ast.ReturnStmt, indent int) {
	if re.forwardDecls {
		return
	}
	if len(node.Results) > 1 {
		re.emitToken(")", RightParen, 0)
	}
	str := re.emitAsString(";", 0)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
}

func (re *RustEmitter) PreVisitReturnStmtResult(node ast.Expr, index int, indent int) {
	if index > 0 {
		str := re.emitAsString(", ", 0)
		re.gir.emitToFileBuffer(str, EmptyVisitMethod)
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
	// Get info about this declaration for multi-name handling
	if genDecl, ok := node.Decl.(*ast.GenDecl); ok {
		for _, spec := range genDecl.Specs {
			if valueSpec, ok := spec.(*ast.ValueSpec); ok {
				re.declNameCount = len(valueSpec.Names)
				// Store type info for multi-name declarations
				if valueSpec.Type != nil {
					re.declType = re.pkg.TypesInfo.Types[valueSpec.Type].Type.String()
				}
			}
		}
	}
	re.declNameIndex = 0
	// Use "let mut" for var declarations since they may be reassigned
	re.emitToken("let", RustKeyword, indent)
	re.emitToken(" ", WhiteSpace, 0)
	re.emitToken("mut", RustKeyword, 0)
	re.emitToken(" ", WhiteSpace, 0)
}

func (re *RustEmitter) PostVisitDeclStmt(node *ast.DeclStmt, indent int) {
	// Reordering is now done per-name in PostVisitDeclStmtValueSpecNames
	re.emitToken(";", Semicolon, 0)
	re.shouldGenerate = false
}

func (re *RustEmitter) PreVisitAssignStmt(node *ast.AssignStmt, indent int) {
	re.shouldGenerate = true
	str := re.emitAsString("", indent)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
}
func (re *RustEmitter) PostVisitAssignStmt(node *ast.AssignStmt, indent int) {
	re.emitToken(";", Semicolon, 0)
	re.shouldGenerate = false
}

func (re *RustEmitter) PreVisitAssignStmtRhs(node *ast.AssignStmt, indent int) {
	re.shouldGenerate = true
	re.inAssignRhs = true
	opTokenType := re.getTokenType(re.assignmentToken)
	re.emitToken(re.assignmentToken, opTokenType, indent+1)
	re.emitToken(" ", WhiteSpace, 0)
}

func (re *RustEmitter) PostVisitAssignStmtRhs(node *ast.AssignStmt, indent int) {
	re.shouldGenerate = false
	re.isTuple = false
	re.inAssignRhs = false
}

func (re *RustEmitter) PreVisitAssignStmtLhsExpr(node ast.Expr, index int, indent int) {
	if index > 0 {
		str := re.emitAsString(", ", indent)
		re.gir.emitToFileBuffer(str, EmptyVisitMethod)
	}
}

func (re *RustEmitter) PreVisitAssignStmtLhs(node *ast.AssignStmt, indent int) {
	re.shouldGenerate = true
	assignmentToken := node.Tok.String()
	if assignmentToken == ":=" && len(node.Lhs) == 1 {
		re.emitToken("let", RustKeyword, indent)
		re.emitToken(" ", WhiteSpace, 0)
		re.emitToken("mut", RustKeyword, 0)
		re.emitToken(" ", WhiteSpace, 0)
	} else if assignmentToken == ":=" && len(node.Lhs) > 1 {
		re.emitToken("(", LeftParen, 0)
		re.emitToken("let", RustKeyword, indent)
		re.emitToken(" ", WhiteSpace, 0)
		re.emitToken("mut", RustKeyword, 0)
		re.emitToken(" ", WhiteSpace, 0)
	} else if assignmentToken == "=" && len(node.Lhs) > 1 {
		re.emitToken("(", LeftParen, indent)
		re.isTuple = true
	}
	if assignmentToken != "+=" {
		assignmentToken = "="
	}
	re.assignmentToken = assignmentToken
}

func (re *RustEmitter) PostVisitAssignStmtLhs(node *ast.AssignStmt, indent int) {
	if node.Tok.String() == ":=" && len(node.Lhs) > 1 {
		re.emitToken(")", RightParen, indent)
	} else if node.Tok.String() == "=" && len(node.Lhs) > 1 {
		re.emitToken(")", RightParen, indent)
	}
	re.shouldGenerate = false

}

func (re *RustEmitter) PreVisitIndexExpr(node *ast.IndexExpr, indent int) {
	// For assignment RHS, check if the element type is a function (needs borrowing in Rust)
	if re.inAssignRhs {
		tv := re.pkg.TypesInfo.Types[node.X]
		if tv.Type != nil {
			// Check if it's a slice/array of functions
			typeStr := tv.Type.String()
			if strings.Contains(typeStr, "func(") {
				// Add borrow operator for function types
				re.gir.emitToFileBuffer("&", EmptyVisitMethod)
			}
		}
	}
}

func (re *RustEmitter) PreVisitIndexExprIndex(node *ast.IndexExpr, indent int) {
	re.shouldGenerate = true
	re.emitToken("[", LeftBracket, 0)

}
func (re *RustEmitter) PostVisitIndexExprIndex(node *ast.IndexExpr, indent int) {
	// Check if the index type is an integer (not usize) - need to add "as usize"
	if node.Index != nil {
		tv := re.pkg.TypesInfo.Types[node.Index]
		if tv.Type != nil {
			typeStr := tv.Type.String()
			// Go int types need to be cast to usize for Rust indexing
			if typeStr == "int" || typeStr == "int32" || typeStr == "int64" ||
				typeStr == "int8" || typeStr == "int16" {
				re.gir.emitToFileBuffer(" as usize", EmptyVisitMethod)
			}
		}
	}
	re.emitToken("]", RightBracket, 0)
}

func (re *RustEmitter) PreVisitBinaryExpr(node *ast.BinaryExpr, indent int) {
	re.shouldGenerate = true
	re.emitToken("(", LeftParen, 1)
}
func (re *RustEmitter) PostVisitBinaryExpr(node *ast.BinaryExpr, indent int) {
	re.emitToken(")", RightParen, 1)
	re.shouldGenerate = false
}

func (re *RustEmitter) PreVisitBinaryExprOperator(op token.Token, indent int) {
	content := op.String()
	opTokenType := re.getTokenType(content)
	re.emitToken(content, opTokenType, 0)
	re.emitToken(" ", WhiteSpace, 0)
}

func (re *RustEmitter) PreVisitCallExprArg(node ast.Expr, index int, indent int) {
	if index > 0 {
		str := re.emitAsString(", ", 0)
		re.gir.emitToFileBuffer(str, EmptyVisitMethod)
	}
}

func (re *RustEmitter) PostVisitCallExprArg(node ast.Expr, index int, indent int) {
	if re.forwardDecls {
		return
	}
	// Check if the argument is a struct type that needs .clone()
	tv := re.pkg.TypesInfo.Types[node]
	if tv.Type != nil {
		// Check if it's a named type (potential struct)
		if named, ok := tv.Type.(*types.Named); ok {
			// Check if underlying type is a struct
			if _, isStruct := named.Underlying().(*types.Struct); isStruct {
				// Get the struct name and check if it has interface{} fields
				structName := named.Obj().Name()
				if !re.structHasInterfaceFields(structName) {
					// Add .clone() for struct types without interface{} fields
					re.gir.emitToFileBuffer(".clone()", EmptyVisitMethod)
				}
				return
			}
		}
		// Also handle non-named struct types (rare but possible)
		if _, isStruct := tv.Type.(*types.Struct); isStruct {
			re.gir.emitToFileBuffer(".clone()", EmptyVisitMethod)
			return
		}
	}
}

func (re *RustEmitter) PostVisitExprStmtX(node ast.Expr, indent int) {
	str := re.emitAsString(";", 0)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
}

func (re *RustEmitter) PreVisitIfStmt(node *ast.IfStmt, indent int) {
	re.shouldGenerate = true
}
func (re *RustEmitter) PostVisitIfStmt(node *ast.IfStmt, indent int) {
	re.shouldGenerate = false
}

func (re *RustEmitter) PreVisitIfStmtCond(node *ast.IfStmt, indent int) {
	str := re.emitAsString("if ", 1)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
	re.emitToken("(", LeftParen, 0)
}

func (re *RustEmitter) PostVisitIfStmtCond(node *ast.IfStmt, indent int) {
	re.emitToken(")", RightParen, 0)
	str := re.emitAsString("\n", 0)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
}

func (re *RustEmitter) PreVisitForStmt(node *ast.ForStmt, indent int) {
	re.insideForPostCond = true
	re.sawIncrement = false // Reset for this for loop
	str := re.emitAsString("for ", indent)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
	re.shouldGenerate = true
}

func (re *RustEmitter) PostVisitForStmtInit(node ast.Stmt, indent int) {
	if node == nil {
		str := re.emitAsString(";", 0)
		re.gir.emitToFileBuffer(str, EmptyVisitMethod)
	}
}

func (re *RustEmitter) PostVisitForStmtPost(node ast.Stmt, indent int) {
	if node != nil {
		re.insideForPostCond = false
	}
	str := re.emitAsString("\n", 0)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
}

func (re *RustEmitter) PreVisitIfStmtElse(node *ast.IfStmt, indent int) {
	str := re.emitAsString("else", 1)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
}

func (re *RustEmitter) PostVisitForStmtCond(node ast.Expr, indent int) {
	str := re.emitAsString(";", 0)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
	re.shouldGenerate = false
}

func (re *RustEmitter) PostVisitForStmt(node *ast.ForStmt, indent int) {
	re.shouldGenerate = false
	re.insideForPostCond = false

	p1 := SearchPointerIndexReverse(PreVisitForStmtInit, re.gir.pointerAndIndexVec)
	p2 := SearchPointerIndexReverse(PostVisitForStmtInit, re.gir.pointerAndIndexVec)
	var forVars []Token
	var rangeTokens []Token
	hasInit := false
	if p1 != nil && p2 != nil {
		// Extract the substring between the positions of the pointers
		initTokens, err := ExtractTokensBetween(p1.Index, p2.Index, re.gir.tokenSlice)
		if err != nil {
			fmt.Println("Error extracting init statement:", err)
			return
		}
		for i := 0; i < len(initTokens); i++ {
			tok := initTokens[i]
			if tok.Type == WhiteSpace {
				initTokens, _ = RemoveTokenAt(initTokens, i)
				i = i - 1
			}
		}
		// Check if there's actual init content (not just empty)
		hasInit = len(initTokens) > 0 && !(len(initTokens) == 1 && initTokens[0].Content == ";")
		for i, tok := range initTokens {
			if tok.Type == Assignment {
				forVars = append(forVars, initTokens[i-1])
				rangeTokens = append(rangeTokens, initTokens[i+1])
			}
		}
	}

	p3 := SearchPointerIndexReverse(PreVisitForStmtCond, re.gir.pointerAndIndexVec)
	p4 := SearchPointerIndexReverse(PostVisitForStmtCond, re.gir.pointerAndIndexVec)
	var condTokens []Token
	hasCond := false
	if p3 != nil && p4 != nil {
		// Extract the substring between the positions of the pointers
		var err error
		condTokens, err = ExtractTokensBetween(p3.Index, p4.Index, re.gir.tokenSlice)
		if err != nil {
			fmt.Println("Error extracting condition statement:", err)
			return
		}
		for i := 0; i < len(condTokens); i++ {
			tok := condTokens[i]
			if tok.Type == WhiteSpace {
				condTokens, _ = RemoveTokenAt(condTokens, i)
				i = i - 1
			}
		}
		// Check if there's actual condition content
		hasCond = len(condTokens) > 0 && !(len(condTokens) == 1 && condTokens[0].Content == ";")

		for i, tok := range condTokens {
			if tok.Type == ComparisonOperator && tok.Content == "<" {
				rangeTokens = append(rangeTokens, condTokens[i+1])
			}
		}
	}

	p6 := SearchPointerIndexReverse(PostVisitForStmtPost, re.gir.pointerAndIndexVec)
	pFor := SearchPointerIndexReverse(PreVisitForStmt, re.gir.pointerAndIndexVec)

	// Case 1: Condition-only for loop (no init, no post) → while loop
	// Go: for cond { } → Rust: while cond { }
	if pFor != nil && p6 != nil && !hasInit && hasCond && node.Post == nil {
		// Build new tokens for Rust while loop: "while cond\n"
		newTokens := []string{}
		newTokens = append(newTokens, "while ")
		// Remove trailing semicolon from condition tokens
		for _, tok := range condTokens {
			if tok.Content != ";" {
				newTokens = append(newTokens, tok.Content)
			}
		}
		newTokens = append(newTokens, "\n")

		// Rewrite the tokens from PreVisitForStmt to PostVisitForStmtPost
		re.gir.tokenSlice, _ = RewriteTokensBetween(re.gir.tokenSlice, pFor.Index, p6.Index, newTokens)
		return
	}

	// Case 2: Traditional for loop with init, cond, post and increment → for in range
	if pFor != nil && p6 != nil && len(forVars) > 0 && len(rangeTokens) >= 2 && re.sawIncrement {
		// Build new tokens for Rust for loop: "for var in start..end\n"
		newTokens := []string{}
		newTokens = append(newTokens, "for ")
		newTokens = append(newTokens, forVars[0].Content)
		newTokens = append(newTokens, " in ")
		newTokens = append(newTokens, rangeTokens[0].Content)
		newTokens = append(newTokens, "..")
		newTokens = append(newTokens, rangeTokens[1].Content)
		newTokens = append(newTokens, "\n")

		// Rewrite the tokens from PreVisitForStmt to PostVisitForStmtPost
		re.gir.tokenSlice, _ = RewriteTokensBetween(re.gir.tokenSlice, pFor.Index, p6.Index, newTokens)
	}
}

func (re *RustEmitter) PreVisitRangeStmt(node *ast.RangeStmt, indent int) {
	re.shouldGenerate = true
	str := re.emitAsString("for ", indent)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
}

func (re *RustEmitter) PostVisitRangeStmtValue(node ast.Expr, indent int) {
	str := re.emitAsString(" in ", 0)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
}

func (re *RustEmitter) PostVisitRangeStmtX(node ast.Expr, indent int) {
	str := re.emitAsString("\n", 0)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
	re.shouldGenerate = false
}

func (re *RustEmitter) PreVisitIncDecStmt(node *ast.IncDecStmt, indent int) {
	re.shouldGenerate = true
	// Track if we see ++ for for loop rewriting
	if node.Tok.String() == "++" {
		re.sawIncrement = true
	}
}

func (re *RustEmitter) PostVisitIncDecStmt(node *ast.IncDecStmt, indent int) {
	content := node.Tok.String()
	// Rust doesn't support ++ or --, convert to += 1 or -= 1
	if content == "++" {
		re.gir.emitToFileBuffer(" += 1", EmptyVisitMethod)
	} else if content == "--" {
		re.gir.emitToFileBuffer(" -= 1", EmptyVisitMethod)
	}
	if !re.insideForPostCond {
		re.emitToken(";", Semicolon, 0)
	}
	re.shouldGenerate = false
}

func (re *RustEmitter) PreVisitCompositeLit(node *ast.CompositeLit, indent int) {
	// Check if the composite literal's type is from a package with interface{} types
	// If so, that type won't derive Default, so we shouldn't use ..Default::default()
	re.currentCompLitTypeNoDefault = false // Reset for each composite literal

	if node.Type != nil {
		typeInfo := re.pkg.TypesInfo.Types[node.Type]
		if typeInfo.Type != nil {
			// Check if this is a named type and get its package
			if named, ok := typeInfo.Type.(*types.Named); ok {
				if named.Obj() != nil && named.Obj().Pkg() != nil {
					typePkg := named.Obj().Pkg()
					typePkgPath := typePkg.Path()
					// Check if we've already determined this package has interface{} types
					if hasIntf, cached := re.processedPkgsInterfaceTypes[typePkgPath]; cached {
						re.currentCompLitTypeNoDefault = hasIntf
					} else {
						// Not cached, check if the type's package is same as current
						if typePkgPath == re.pkg.PkgPath {
							re.currentCompLitTypeNoDefault = re.pkgHasInterfaceTypes
						} else {
							// For external packages, check if ANY type in that package has interface{} fields
							hasIntf := re.packageScopeHasInterfaceTypes(typePkg)
							re.processedPkgsInterfaceTypes[typePkgPath] = hasIntf
							re.currentCompLitTypeNoDefault = hasIntf
						}
					}
				}
			}
		}
	}
}

// packageScopeHasInterfaceTypes checks if any struct in the package has interface{} fields
func (re *RustEmitter) packageScopeHasInterfaceTypes(pkg *types.Package) bool {
	scope := pkg.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		if typeName, ok := obj.(*types.TypeName); ok {
			if re.typeHasInterfaceFields(typeName.Type()) {
				return true
			}
		}
	}
	return false
}

func (re *RustEmitter) PreVisitCompositeLitType(node ast.Expr, indent int) {
	re.gir.emitToFileBuffer("", "@PreVisitCompositeLitType")
}

func (re *RustEmitter) PostVisitCompositeLitType(node ast.Expr, indent int) {
	pointerAndPosition := SearchPointerIndexReverse("@PreVisitCompositeLitType", re.gir.pointerAndIndexVec)
	if pointerAndPosition != nil {
		// TODO not very effective
		// go through all aliases and check if the underlying type matches
		for aliasName, alias := range re.aliases {
			if alias.UnderlyingType == re.pkg.TypesInfo.Types[node].Type.Underlying().String() {
				re.gir.tokenSlice, _ = RewriteTokensBetween(re.gir.tokenSlice, pointerAndPosition.Index, len(re.gir.tokenSlice), []string{aliasName})
			}
		}
		if re.isArray {
			// TODO that's still hack
			// we operate on string representation of the type
			// has to be rewritten to use some kind of IR
			if re.inKeyValueExpr {
				// Inside struct field initialization: nodes: []Type{} -> nodes: vec![]
				// Just replace the type with vec!, keeping the field name and colon
				newTokens := []string{"vec!"}
				re.gir.tokenSlice, _ = RewriteTokensBetween(re.gir.tokenSlice, pointerAndPosition.Index, len(re.gir.tokenSlice), newTokens)
			} else {
				// Variable declaration: let x = []Type{} -> let x: Vec<type> = vec![]
				vecTypeStrRepr, _ := ExtractTokensBetween(pointerAndPosition.Index, len(re.gir.tokenSlice), re.gir.tokenSlice)
				newTokens := []string{}
				newTokens = append(newTokens, ":")
				newTokens = append(newTokens, tokensToStrings(vecTypeStrRepr)...)
				newTokens = append(newTokens, " = vec!")
				re.gir.tokenSlice, _ = RewriteTokensBetween(re.gir.tokenSlice, pointerAndPosition.Index-len("=")-len(" "), len(re.gir.tokenSlice), newTokens)
			}
		}
	}
}

func (re *RustEmitter) PreVisitCompositeLitElts(node []ast.Expr, indent int) {
	re.emitToken("{", LeftBrace, 0)
}

func (re *RustEmitter) PostVisitCompositeLitElts(node []ast.Expr, indent int) {
	// For empty struct initialization in Rust, add ..Default::default()
	// But NOT for arrays/vectors - those just use {}
	// Don't use Default::default() if the type doesn't derive Default (due to interface{} fields)
	if len(node) == 0 && !re.isArray && !re.currentCompLitTypeNoDefault {
		re.gir.emitToFileBuffer("..Default::default()", EmptyVisitMethod)
	}
	re.emitToken("}", RightBrace, 0)
}

func (re *RustEmitter) PreVisitCompositeLitElt(node ast.Expr, index int, indent int) {
	if index > 0 {
		str := re.emitAsString(", ", 0)
		re.gir.emitToFileBuffer(str, EmptyVisitMethod)
	}
}

func (re *RustEmitter) PreVisitSliceExpr(node *ast.SliceExpr, indent int) {
	// Don't add & - we'll add .to_vec() at the end to get Vec back
}

func (re *RustEmitter) PostVisitSliceExprX(node ast.Expr, indent int) {
	re.emitToken("[", LeftBracket, 0)
	re.shouldGenerate = false
}

func (re *RustEmitter) PostVisitSliceExpr(node *ast.SliceExpr, indent int) {
	re.emitToken("]", RightBracket, 0)
	// Convert slice to Vec to match Go semantics
	re.gir.emitToFileBuffer(".to_vec()", EmptyVisitMethod)
	re.shouldGenerate = true
}

func (re *RustEmitter) PostVisitSliceExprLow(node ast.Expr, indent int) {
	re.gir.emitToFileBuffer("..", EmptyVisitMethod)
}

func (re *RustEmitter) PreVisitFuncLit(node *ast.FuncLit, indent int) {
	// Wrap closure in Box::new() for Rust
	str := re.emitAsString("Box::new(", 0)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
	re.emitToken("|", Identifier, indent)
}
func (re *RustEmitter) PostVisitFuncLit(node *ast.FuncLit, indent int) {
	re.emitToken("}", RightBrace, 0)
	// Close the Box::new() wrapper
	re.emitToken(")", RightParen, 0)
}

func (re *RustEmitter) PostVisitFuncLitTypeParams(node *ast.FieldList, indent int) {
	re.emitToken("|", Identifier, 0)
}

func (re *RustEmitter) PreVisitFuncLitTypeParam(node *ast.Field, index int, indent int) {
	str := ""
	if index > 0 {
		str += re.emitAsString(", ", 0)
	}
	// Emit name first, then colon, then type will follow
	if len(node.Names) > 0 {
		str += re.emitAsString(node.Names[0].Name+": ", 0)
	}
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
}

func (re *RustEmitter) PostVisitFuncLitTypeParam(node *ast.Field, index int, indent int) {
	// Type has already been emitted, nothing to do
}

func (re *RustEmitter) PreVisitFuncLitBody(node *ast.BlockStmt, indent int) {
	re.emitToken("{", LeftBrace, 0)
	str := re.emitAsString("\n", 0)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
}

func (re *RustEmitter) PreVisitFuncLitTypeResults(node *ast.FieldList, indent int) {
	re.shouldGenerate = false
}

func (re *RustEmitter) PreVisitInterfaceType(node *ast.InterfaceType, indent int) {
	str := re.emitAsString("Box<dyn Any>", indent)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
}

func (re *RustEmitter) PostVisitInterfaceType(node *ast.InterfaceType, indent int) {
}

func (re *RustEmitter) PreVisitKeyValueExprValue(node ast.Expr, indent int) {
	// In Rust struct initialization, use `:` not `=`
	str := re.emitAsString(": ", 0)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
}

func (re *RustEmitter) PreVisitUnaryExpr(node *ast.UnaryExpr, indent int) {
	re.emitToken("(", LeftParen, 0)
	str := re.emitAsString(node.Op.String(), 0)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
}
func (re *RustEmitter) PostVisitUnaryExpr(node *ast.UnaryExpr, indent int) {
	re.emitToken(")", RightParen, 0)
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

			re.gir.emitToFileBuffer(str, EmptyVisitMethod)
		}
	}
}
func (re *RustEmitter) PostVisitGenDeclConstName(node *ast.Ident, indent int) {
	str := re.emitAsString(";\n", 0)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
}
func (re *RustEmitter) PostVisitGenDeclConst(node *ast.GenDecl, indent int) {
	str := re.emitAsString("\n", 0)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
}

func (re *RustEmitter) PreVisitSwitchStmt(node *ast.SwitchStmt, indent int) {
	re.shouldGenerate = true
	str := re.emitAsString("switch (", indent)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
}
func (re *RustEmitter) PostVisitSwitchStmt(node *ast.SwitchStmt, indent int) {
	re.emitToken("}", RightBrace, indent)
}

func (re *RustEmitter) PostVisitSwitchStmtTag(node ast.Expr, indent int) {
	re.emitToken(")", RightParen, 0)
	str := re.emitAsString(" ", 0)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
	re.emitToken("{", LeftBrace, 0)
	str = re.emitAsString("\n", 0)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
}

func (re *RustEmitter) PostVisitCaseClause(node *ast.CaseClause, indent int) {
	re.gir.emitToFileBuffer("\n", EmptyVisitMethod)
	str := re.emitAsString("break;\n", indent+4)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
}

func (re *RustEmitter) PostVisitCaseClauseList(node []ast.Expr, indent int) {
	if len(node) == 0 {
		str := re.emitAsString("default:\n", indent+2)
		re.gir.emitToFileBuffer(str, EmptyVisitMethod)
	}
}

func (re *RustEmitter) PreVisitCaseClauseListExpr(node ast.Expr, index int, indent int) {
	str := re.emitAsString("case ", indent+2)
	tv := re.pkg.TypesInfo.Types[node]
	if typeVal, ok := rustTypesMap[tv.Type.String()]; ok {
		str += "(" + typeVal + ")"
	}
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
	re.shouldGenerate = true
}

func (re *RustEmitter) PostVisitCaseClauseListExpr(node ast.Expr, index int, indent int) {
	str := re.emitAsString(":\n", 0)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
}

func (re *RustEmitter) PreVisitTypeAssertExprType(node ast.Expr, indent int) {
	re.emitToken("(", LeftParen, indent)
}

func (re *RustEmitter) PostVisitTypeAssertExprType(node ast.Expr, indent int) {
	re.emitToken(")", RightParen, indent)
}

func (re *RustEmitter) PreVisitKeyValueExpr(node *ast.KeyValueExpr, indent int) {
	re.shouldGenerate = true
	re.inKeyValueExpr = true
}

func (re *RustEmitter) PostVisitKeyValueExpr(node *ast.KeyValueExpr, indent int) {
	re.inKeyValueExpr = false
}

func (re *RustEmitter) PreVisitBranchStmt(node *ast.BranchStmt, indent int) {
	str := re.emitAsString(node.Tok.String()+";", indent)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
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
	p1 := SearchPointerIndexReverse("@PreVisitFuncDeclSignatureTypeParamsListType", re.gir.pointerAndIndexVec)
	p2 := SearchPointerIndexReverse("@PostVisitFuncDeclSignatureTypeParamsListType", re.gir.pointerAndIndexVec)
	p3 := SearchPointerIndexReverse("@PreVisitFuncDeclSignatureTypeParamsArgName", re.gir.pointerAndIndexVec)
	p4 := SearchPointerIndexReverse("@PostVisitFuncDeclSignatureTypeParamsArgName", re.gir.pointerAndIndexVec)

	if p1 != nil && p2 != nil && p3 != nil && p4 != nil {
		typeStrRepr, err := ExtractTokensBetween(p1.Index, p2.Index, re.gir.tokenSlice)
		if err != nil {
			fmt.Println("Error extracting type representation:", err)
			return
		}
		nameStrRepr, err := ExtractTokensBetween(p3.Index, p4.Index, re.gir.tokenSlice)
		if err != nil {
			fmt.Println("Error extracting name representation:", err)
			return
		}
		if containsWhitespace(strings.Join(tokensToStrings(nameStrRepr), "")) {
			fmt.Println("Error: Type parameter name contains whitespace")
			return
		}
		if containsWhitespace(strings.Join(tokensToStrings(typeStrRepr), "")) {
			fmt.Println("Error: Type parameter type contains whitespace")
			return
		}
		newTokens := []string{}
		// Add mut for non-primitive types (struct parameters are often modified in Go)
		typeStr := strings.Join(tokensToStrings(typeStrRepr), "")
		isPrimitive := typeStr == "i8" || typeStr == "i16" || typeStr == "i32" || typeStr == "i64" ||
			typeStr == "u8" || typeStr == "u16" || typeStr == "u32" || typeStr == "u64" ||
			typeStr == "bool" || typeStr == "f32" || typeStr == "f64" || typeStr == "String" ||
			typeStr == "&str" || typeStr == "usize" || typeStr == "isize"
		if !isPrimitive {
			newTokens = append(newTokens, "mut ")
		}
		newTokens = append(newTokens, tokensToStrings(nameStrRepr)...)
		newTokens = append(newTokens, ":")
		newTokens = append(newTokens, tokensToStrings(typeStrRepr)...)
		re.gir.tokenSlice, err = RewriteTokensBetween(re.gir.tokenSlice, p1.Index, p4.Index, newTokens)
		if err != nil {
			fmt.Println("Error rewriting file buffer:", err)
			return
		}
	}
}

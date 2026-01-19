package compiler

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"strings"

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
	Output      string
	OutputDir   string
	OutputName  string
	LinkRuntime string // Path to runtime directory (empty = disabled)
	file        *os.File
	BaseEmitter
	pkg                          *packages.Package
	insideForPostCond            bool
	assignmentToken              string
	forwardDecls                 bool
	shouldGenerate               bool
	numFuncResults               int
	aliases                      map[string]Alias
	currentPackage               string
	isArray                      bool
	arrayType                    string
	isTuple                      bool
	sawIncrement                 bool                    // Track if we saw ++ in for loop post statement
	isInfiniteLoop               bool                    // Track if current for loop is infinite (no init, cond, post)
	declType                     string                  // Store the type for multi-name declarations
	declNameCount                int                     // Count of names in current declaration
	declNameIndex                int                     // Current name index
	inAssignRhs                  bool                    // Track if we're in assignment RHS
	inAssignLhs                  bool                    // Track if we're in assignment LHS
	inFieldAssign                bool                    // Track if we're assigning to a struct field
	isArrayStack                 []bool                  // Stack to save/restore isArray for nested composite literals
	pkgHasInterfaceTypes         bool                    // Track if current package has any interface{} types
	currentCompLitTypeNoDefault  bool                    // Track if current composite literal's type doesn't derive Default
	compLitTypeNoDefaultStack    []bool                  // Stack to save/restore currentCompLitTypeNoDefault for nested composite literals
	currentCompLitType           types.Type              // Track the current composite literal's type for checking at post-visit
	compLitTypeStack             []types.Type            // Stack of composite literal types
	processedPkgsInterfaceTypes  map[string]bool         // Cache for package interface{} type checks
	inKeyValueExpr               bool                    // Track if we're inside a KeyValueExpr (struct field init)
	inMultiValueReturn           bool                    // Track if we're in a multi-value return statement
	multiValueReturnResultIndex  int                     // Current result index in multi-value return
	inReturnStmt                 bool                    // Track if we're inside a return statement
	inMultiValueDecl             bool                    // Track if we're in a multi-value := declaration
	currentFuncReturnsAny        bool                    // Track if current function returns any/interface{}
	callExprFunMarkerStack       []int                   // Stack of indices for nested call markers
	callExprFunEndMarkerStack    []int                   // Stack of end indices for nested call markers
	callExprArgsMarkerStack      []int                   // Stack of indices for nested call arg markers
	localClosureAssign           bool                    // Track if current assignment has a function literal RHS
	localClosures                map[string]*ast.FuncLit // Map of local closure names to their AST
	localClosureBodyTokens       map[string][]Token      // Map of local closure names to their body tokens
	currentClosureName           string                  // Name of the variable being assigned a closure
	inLocalClosureInline         bool                    // Track if we're inlining a local closure
	inLocalClosureBody           bool                    // Track if we're inside a local closure body being processed
	localClosureBodyStartIndex   int                     // Token index where closure body starts (after opening brace)
	localClosureAssignStartIndex int                     // Token index where the assignment statement starts
	currentCompLitIsSlice        bool                    // Track if current composite literal is a slice type alias
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

// escapeRustKeyword escapes Rust reserved keywords with r# prefix
func escapeRustKeyword(name string) string {
	// Rust reserved keywords that might conflict with Go identifiers
	rustKeywords := map[string]bool{
		"as": true, "break": true, "const": true, "continue": true,
		"crate": true, "else": true, "enum": true, "extern": true,
		// Note: "false" and "true" are NOT escaped - they're boolean literals
		"fn": true, "for": true, "if": true,
		"impl": true, "in": true, "let": true, "loop": true,
		"match": true, "mod": true, "move": true, "mut": true,
		"pub": true, "ref": true, "return": true, "self": true,
		"Self": true, "static": true, "struct": true, "super": true,
		"trait": true, "type": true, "unsafe": true,
		"use": true, "where": true, "while": true,
		// Reserved for future use
		"abstract": true, "async": true, "await": true, "become": true,
		"box": true, "do": true, "final": true, "macro": true,
		"override": true, "priv": true, "try": true, "typeof": true,
		"unsized": true, "virtual": true, "yield": true,
	}
	if rustKeywords[name] {
		return "r#" + name
	}
	return name
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

	// For Cargo projects, write to src/main.rs instead
	if re.LinkRuntime != "" {
		srcDir := filepath.Join(re.OutputDir, "src")
		if err := os.MkdirAll(srcDir, 0755); err != nil {
			fmt.Println("Error creating src directory:", err)
			return
		}
		outputFile = filepath.Join(srcDir, "main.rs")
	}

	var err error
	re.file, err = os.Create(outputFile)
	re.SetFile(re.file)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	builtin := `use std::fmt;
use std::any::Any;
use std::rc::Rc;

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

pub fn printf2<T: fmt::Display>(fmt_str: String, val: T) {
    // Convert C-style format to Rust format
    let rust_fmt = fmt_str.replace("%d", "{}").replace("%s", "{}").replace("%v", "{}");
    let result = rust_fmt.replace("{}", &format!("{}", val));
    print!("{}", result);
}

pub fn printf3<T1: fmt::Display, T2: fmt::Display>(fmt_str: String, v1: T1, v2: T2) {
    let rust_fmt = fmt_str.replace("%d", "{}").replace("%s", "{}").replace("%v", "{}");
    let result = rust_fmt.replacen("{}", &format!("{}", v1), 1).replacen("{}", &format!("{}", v2), 1);
    print!("{}", result);
}

pub fn printf4<T1: fmt::Display, T2: fmt::Display, T3: fmt::Display>(fmt_str: String, v1: T1, v2: T2, v3: T3) {
    let rust_fmt = fmt_str.replace("%d", "{}").replace("%s", "{}").replace("%v", "{}");
    let result = rust_fmt.replacen("{}", &format!("{}", v1), 1).replacen("{}", &format!("{}", v2), 1).replacen("{}", &format!("{}", v3), 1);
    print!("{}", result);
}

pub fn printf5<T1: fmt::Display, T2: fmt::Display, T3: fmt::Display, T4: fmt::Display>(fmt_str: String, v1: T1, v2: T2, v3: T3, v4: T4) {
    let rust_fmt = fmt_str.replace("%d", "{}").replace("%s", "{}").replace("%v", "{}");
    let result = rust_fmt.replacen("{}", &format!("{}", v1), 1).replacen("{}", &format!("{}", v2), 1).replacen("{}", &format!("{}", v3), 1).replacen("{}", &format!("{}", v4), 1);
    print!("{}", result);
}

// Print byte as character (for %c format)
pub fn printc(b: i8) {
    print!("{}", b as u8 as char);
}

// Convert byte to character string (for Sprintf %c format)
pub fn byte_to_char(b: i8) -> String {
    (b as u8 as char).to_string()
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

// string_format for 2 args (format string + 1 value)
pub fn string_format2<T: fmt::Display>(fmt_str: &str, val: T) -> String {
    let rust_fmt = fmt_str.replace("%d", "{}").replace("%s", "{}").replace("%v", "{}");
    rust_fmt.replace("{}", &format!("{}", val))
}

pub fn len<T>(slice: &[T]) -> i32 {
    slice.len() as i32
}
`
	str := re.emitAsString(builtin, indent)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)

	// Include runtime module if link-runtime is enabled
	if re.LinkRuntime != "" {
		runtimeInclude := `
// Graphics runtime
mod graphics;
use graphics::*;
`
		re.gir.emitToFileBuffer(runtimeInclude, EmptyVisitMethod)
	}

	re.insideForPostCond = false
}

func (re *RustEmitter) PostVisitProgram(indent int) {
	emitTokensToFile(re.file, re.gir.tokenSlice)
	re.file.Close()

	// Generate Cargo project files if link-runtime is enabled
	if re.LinkRuntime != "" {
		if err := re.GenerateCargoToml(); err != nil {
			log.Printf("Warning: %v", err)
		}
		if err := re.GenerateGraphicsMod(); err != nil {
			log.Printf("Warning: %v", err)
		}
		if err := re.GenerateBuildRs(); err != nil {
			log.Printf("Warning: %v", err)
		}
	}
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
	str = re.emitAsString(fmt.Sprintf("pub fn %s", node.Name), 0)
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
	// Note: removed isArray = false as it interfered with composite literal stack management
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
		// In Go, nil for slices means empty slice - use Vec::new() in Rust
		// For pointers/interfaces, None would be correct, but Vec::new() is safer
		// for the common case of slice assignment
		str = re.emitAsString("Vec::new()", indent)
	} else {
		if n, ok := rustTypesMap[name]; ok {
			str = re.emitAsString(n, indent)
		} else {
			// Escape Rust keywords
			name = escapeRustKeyword(name)
			str = re.emitAsString(name, indent)
		}
	}

	re.emitToken(str, Identifier, 0)

}
func (re *RustEmitter) PreVisitCallExprArgs(node []ast.Expr, indent int) {
	if re.forwardDecls {
		return
	}
	// Push the args start position to the stack for nested call handling
	re.callExprArgsMarkerStack = append(re.callExprArgsMarkerStack, len(re.gir.tokenSlice))
	re.gir.emitToFileBuffer("", "@PreVisitCallExprArgs")
	re.emitToken("(", LeftParen, 0)
	// Use stack indices for function name extraction (top of stacks = current call)
	if len(re.callExprFunMarkerStack) > 0 && len(re.callExprFunEndMarkerStack) > 0 {
		p1Index := re.callExprFunMarkerStack[len(re.callExprFunMarkerStack)-1]
		p2Index := re.callExprFunEndMarkerStack[len(re.callExprFunEndMarkerStack)-1]
		// Extract the substring between the positions of the pointers
		funName, err := ExtractTokensBetween(p1Index, p2Index, re.gir.tokenSlice)
		if err != nil {
			fmt.Println("Error extracting function name:", err)
			return
		}
		funNameStr := strings.Join(tokensToStrings(funName), "")
		// Skip adding & for type conversions
		if isConversion, _ := re.isTypeConversion(funNameStr); !isConversion {
			if strings.Contains(funNameStr, "len") || strings.Contains(funNameStr, "append") {
				// add & before the first argument for len and append
				str := re.emitAsString("&", 0)
				re.gir.emitToFileBuffer(str, EmptyVisitMethod)
			}
		}
	}
}

// isTypeConversion checks if a function name represents a type conversion
func (re *RustEmitter) isTypeConversion(funName string) (bool, string) {
	// Map Go type names and Rust type names to Rust cast targets
	typeConversions := map[string]string{
		// Go type names
		"int8":    "i8",
		"int16":   "i16",
		"int32":   "i32",
		"int64":   "i64",
		"int":     "i32",
		"uint8":   "u8",
		"uint16":  "u16",
		"uint32":  "u32",
		"uint64":  "u64",
		"uint":    "u32",
		"float32": "f32",
		"float64": "f64",
		"byte":    "u8",
		"rune":    "i32",
		// Rust type names (in case they're already converted)
		"i8":  "i8",
		"i16": "i16",
		"i32": "i32",
		"i64": "i64",
		"u8":  "u8",
		"u16": "u16",
		"u32": "u32",
		"u64": "u64",
		"f32": "f32",
		"f64": "f64",
	}
	if rustType, ok := typeConversions[funName]; ok {
		return true, rustType
	}
	return false, ""
}

func (re *RustEmitter) PostVisitCallExprArgs(node []ast.Expr, indent int) {
	if re.forwardDecls {
		return
	}

	// Pop from stacks at the end (defer to ensure it happens even on early returns)
	defer func() {
		if len(re.callExprFunMarkerStack) > 0 {
			re.callExprFunMarkerStack = re.callExprFunMarkerStack[:len(re.callExprFunMarkerStack)-1]
		}
		if len(re.callExprFunEndMarkerStack) > 0 {
			re.callExprFunEndMarkerStack = re.callExprFunEndMarkerStack[:len(re.callExprFunEndMarkerStack)-1]
		}
		if len(re.callExprArgsMarkerStack) > 0 {
			re.callExprArgsMarkerStack = re.callExprArgsMarkerStack[:len(re.callExprArgsMarkerStack)-1]
		}
	}()

	// Use stack indices for the current call (top of stacks)
	if len(re.callExprFunMarkerStack) == 0 || len(re.callExprFunEndMarkerStack) == 0 || len(re.callExprArgsMarkerStack) == 0 {
		re.emitToken(")", RightParen, 0)
		return
	}

	p1Index := re.callExprFunMarkerStack[len(re.callExprFunMarkerStack)-1]
	p2Index := re.callExprFunEndMarkerStack[len(re.callExprFunEndMarkerStack)-1]
	pArgsIndex := re.callExprArgsMarkerStack[len(re.callExprArgsMarkerStack)-1]

	funName, err := ExtractTokensBetween(p1Index, p2Index, re.gir.tokenSlice)
	if err == nil {
		funNameStr := strings.Join(tokensToStrings(funName), "")

		// Handle local closure inlining: addToken() -> { body }
		funNameTrimmedForClosure := strings.TrimSpace(funNameStr)
		if bodyTokens, ok := re.localClosureBodyTokens[funNameTrimmedForClosure]; ok && len(node) == 0 {
			// Replace the entire call with the inlined body wrapped in a block
			newTokens := []string{"{"}
			for _, tok := range bodyTokens {
				newTokens = append(newTokens, tok.Content)
			}
			newTokens = append(newTokens, "}")
			re.gir.tokenSlice, _ = RewriteTokensBetween(re.gir.tokenSlice, p1Index, len(re.gir.tokenSlice), newTokens)
			return // Skip emitting closing paren since we replaced everything
		}

		// Handle type conversions: i8(x) -> (x as i8)
		funNameTrimmed := strings.TrimSpace(funNameStr)
		if isConv, rustType := re.isTypeConversion(funNameTrimmed); isConv && len(node) == 1 {
			// Extract the argument tokens (between @PreVisitCallExprArgs and current position)
			argTokens, err := ExtractTokensBetween(pArgsIndex, len(re.gir.tokenSlice), re.gir.tokenSlice)
			if err == nil && len(argTokens) > 0 {
				// Remove the opening paren from call args (added by PreVisitCallExprArgs)
				// but keep any inner parens (e.g., from binary expressions)
				argStr := strings.TrimSpace(strings.Join(tokensToStrings(argTokens), ""))
				if len(argStr) > 0 && argStr[0] == '(' {
					argStr = argStr[1:]
				}
				argStr = strings.TrimSpace(argStr)
				// Generate: (arg as type)
				newTokens := []string{"(", argStr, " as ", rustType, ")"}
				re.gir.tokenSlice, _ = RewriteTokensBetween(re.gir.tokenSlice, p1Index, len(re.gir.tokenSlice), newTokens)
				return // Skip emitting closing paren since we replaced everything
			}
		}

		// Handle println with 0 args
		if funNameStr == "println" && len(node) == 0 {
			re.gir.tokenSlice, _ = RewriteTokensBetween(re.gir.tokenSlice, p1Index, p2Index, []string{"println0"})
		}
		// Handle printf with different arg counts (count includes format string)
		if funNameStr == "printf" {
			switch len(node) {
			case 2:
				// Special case: printf("%c", byte) -> printc(byte)
				if basicLit, ok := node[0].(*ast.BasicLit); ok && basicLit.Kind == token.STRING {
					fmtStr := strings.Trim(basicLit.Value, "\"")
					if fmtStr == "%c" {
						// Rewrite to printc and remove the format string argument
						// Find the argument tokens
						argTokens, err := ExtractTokensBetween(pArgsIndex, len(re.gir.tokenSlice), re.gir.tokenSlice)
						if err == nil && len(argTokens) > 0 {
							// Find the comma that separates the format string from the actual argument
							argStr := strings.Join(tokensToStrings(argTokens), "")
							// Skip the opening paren
							if len(argStr) > 0 && argStr[0] == '(' {
								argStr = argStr[1:]
							}
							// Find comma and extract just the second argument
							commaIdx := strings.Index(argStr, ",")
							if commaIdx >= 0 {
								secondArg := strings.TrimSpace(argStr[commaIdx+1:])
								// Rewrite: printf("%c", b) -> printc(b)
								newTokens := []string{"printc", "(", secondArg}
								re.gir.tokenSlice, _ = RewriteTokensBetween(re.gir.tokenSlice, p1Index, len(re.gir.tokenSlice), newTokens)
								// Don't return here - let the closing paren be added normally
								break
							}
						}
					}
				}
				re.gir.tokenSlice, _ = RewriteTokensBetween(re.gir.tokenSlice, p1Index, p2Index, []string{"printf2"})
			case 3:
				re.gir.tokenSlice, _ = RewriteTokensBetween(re.gir.tokenSlice, p1Index, p2Index, []string{"printf3"})
			case 4:
				re.gir.tokenSlice, _ = RewriteTokensBetween(re.gir.tokenSlice, p1Index, p2Index, []string{"printf4"})
			case 5:
				re.gir.tokenSlice, _ = RewriteTokensBetween(re.gir.tokenSlice, p1Index, p2Index, []string{"printf5"})
			}
		}
		// Handle string_format with different arg counts
		if funNameStr == "string_format" {
			switch len(node) {
			case 2:
				// Special case: Sprintf("%c", byte) -> byte_to_char(byte)
				if basicLit, ok := node[0].(*ast.BasicLit); ok && basicLit.Kind == token.STRING {
					fmtStr := strings.Trim(basicLit.Value, "\"")
					if fmtStr == "%c" {
						// Rewrite to byte_to_char and remove the format string argument
						argTokens, err := ExtractTokensBetween(pArgsIndex, len(re.gir.tokenSlice), re.gir.tokenSlice)
						if err == nil && len(argTokens) > 0 {
							argStr := strings.Join(tokensToStrings(argTokens), "")
							if len(argStr) > 0 && argStr[0] == '(' {
								argStr = argStr[1:]
							}
							commaIdx := strings.Index(argStr, ",")
							if commaIdx >= 0 {
								secondArg := strings.TrimSpace(argStr[commaIdx+1:])
								newTokens := []string{"byte_to_char", "(", secondArg}
								re.gir.tokenSlice, _ = RewriteTokensBetween(re.gir.tokenSlice, p1Index, len(re.gir.tokenSlice), newTokens)
								break
							}
						}
					}
				}
				re.gir.tokenSlice, _ = RewriteTokensBetween(re.gir.tokenSlice, p1Index, p2Index, []string{"string_format2"})
			}
		}

		// Handle len() on String - convert to method syntax: len(str) -> str.len() as i32
		if funNameStr == "len" && len(node) == 1 {
			argType := re.pkg.TypesInfo.Types[node[0]]
			if argType.Type != nil && argType.Type.String() == "string" {
				// Extract the argument tokens
				argTokens, err := ExtractTokensBetween(pArgsIndex, len(re.gir.tokenSlice), re.gir.tokenSlice)
				if err == nil && len(argTokens) > 0 {
					argStr := strings.TrimSpace(strings.Join(tokensToStrings(argTokens), ""))
					// Remove ( from start and & if present
					if len(argStr) > 0 && argStr[0] == '(' {
						argStr = argStr[1:]
					}
					argStr = strings.TrimPrefix(argStr, "&")
					argStr = strings.TrimSpace(argStr)
					// Generate: str.len() as i32
					newTokens := []string{argStr, ".len() as i32"}
					re.gir.tokenSlice, _ = RewriteTokensBetween(re.gir.tokenSlice, p1Index, len(re.gir.tokenSlice), newTokens)
					return // Skip emitting closing paren
				}
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
	} else if e.Kind == token.CHAR {
		// Character literals in Go are runes - convert to numeric i8 for Rust
		// This allows use in match patterns (which don't allow `as` casts)
		charVal := e.Value
		if len(charVal) >= 3 && charVal[0] == '\'' && charVal[len(charVal)-1] == '\'' {
			inner := charVal[1 : len(charVal)-1]
			var numVal int
			// Handle escape sequences
			if len(inner) >= 2 && inner[0] == '\\' {
				switch inner[1] {
				case 'n':
					numVal = 10 // newline
				case 't':
					numVal = 9 // tab
				case 'r':
					numVal = 13 // carriage return
				case '\\':
					numVal = 92 // backslash
				case '\'':
					numVal = 39 // single quote
				case '0':
					numVal = 0 // null
				default:
					numVal = int(inner[1])
				}
			} else if len(inner) == 1 {
				// Single character - use ASCII value
				numVal = int(inner[0])
			} else {
				// Fallback - just emit as is
				str = re.emitAsString(charVal, 0)
				re.emitToken(str, CharLiteral, 0)
				return
			}
			// Don't add i8 suffix - let Rust infer the type from context
			// This allows character literals to work in match expressions cast to i32
			str = re.emitAsString(fmt.Sprintf("%d", numVal), 0)
		} else {
			str = re.emitAsString(charVal, 0)
		}
		re.emitToken(str, CharLiteral, 0)
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
	// But skip if we're in a += context (Rust's += for String expects &str)
	if e.Kind == token.STRING {
		if re.inAssignRhs && re.assignmentToken == "+=" {
			// Don't add .to_string() for += operations
			return
		}
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
		typeInfo := re.pkg.TypesInfo.Types[node.Type]
		// Only do alias replacement if the type is NOT already a named type (alias)
		// If it's a named type like types.ExprKind, don't replace it with another alias
		if typeInfo.Type != nil {
			if _, isNamed := typeInfo.Type.(*types.Named); !isNamed {
				// Type is a basic/primitive type - check for alias replacement
				for aliasName, alias := range re.aliases {
					if alias.UnderlyingType == typeInfo.Type.Underlying().String() {
						re.gir.tokenSlice, _ = RewriteTokensBetween(re.gir.tokenSlice, pointerAndPosition.Index, len(re.gir.tokenSlice), []string{aliasName})
						break
					}
				}
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

	// Save the type name BEFORE reordering for default initialization
	var typeName string
	if p1 != nil && p2 != nil {
		fieldType, err := ExtractTokensBetween(p1.Index, p2.Index, re.gir.tokenSlice)
		if err == nil && len(fieldType) > 0 {
			typeName = strings.TrimSpace(strings.Join(tokensToStrings(fieldType), ""))
		}
	}

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
		str += " = Vec::new()"
		re.isArray = false
	} else {
		// Add default initialization based on type
		// Primitive numeric types get zero initialization
		primitiveDefaults := map[string]string{
			"i8": "0", "i16": "0", "i32": "0", "i64": "0",
			"u8": "0", "u16": "0", "u32": "0", "u64": "0",
			"f32": "0.0", "f64": "0.0",
			"bool": "false",
		}
		if defaultVal, isPrimitive := primitiveDefaults[typeName]; isPrimitive {
			str += " = " + defaultVal
		} else if typeName == "String" {
			str += " = String::new()"
		} else if len(typeName) > 0 && !strings.Contains(typeName, "Box<dyn") {
			// For struct types declared without value (var x StructType), initialize with default
			// Skip Box<dyn Any> - can't call default() on trait objects
			// Handle module-qualified types like types::Plan by checking the type name part
			typeNamePart := typeName
			if idx := strings.LastIndex(typeName, "::"); idx >= 0 {
				typeNamePart = typeName[idx+2:]
			}
			// Check if type name starts with uppercase (struct type)
			if len(typeNamePart) > 0 && typeNamePart[0] >= 'A' && typeNamePart[0] <= 'Z' {
				str += " = " + typeName + "::default()"
			}
		}
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
	re.currentPackage = pkg.Name
	// Initialize the caches if not already done
	if re.processedPkgsInterfaceTypes == nil {
		re.processedPkgsInterfaceTypes = make(map[string]bool)
	}
	if re.localClosures == nil {
		re.localClosures = make(map[string]*ast.FuncLit)
	}
	if re.localClosureBodyTokens == nil {
		re.localClosureBodyTokens = make(map[string][]Token)
	}
	// Check if package has any interface{} types
	re.pkgHasInterfaceTypes = re.packageHasInterfaceTypes(pkg)
	// Cache this package's result
	re.processedPkgsInterfaceTypes[pkg.PkgPath] = re.pkgHasInterfaceTypes

	// Generate module declaration for non-main packages
	if pkg.Name != "main" {
		str := re.emitAsString(fmt.Sprintf("pub mod %s {\n", pkg.Name), indent)
		re.gir.emitToFileBuffer(str, EmptyVisitMethod)
		// Import crate-root items (helper functions like append, len, println, etc.)
		str = re.emitAsString("use crate::*;\n\n", indent)
		re.gir.emitToFileBuffer(str, EmptyVisitMethod)
	}
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
			// Check for interface{} fields
			if strings.Contains(fieldTypeStr, "interface{}") || strings.Contains(fieldTypeStr, "interface {") {
				return true
			}
			// Check for function fields (Box<dyn Fn> in Rust doesn't implement Default)
			if strings.Contains(fieldTypeStr, "func(") {
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
	// Close the module declaration for non-main packages
	if pkg.Name != "main" {
		str := re.emitAsString(fmt.Sprintf("} // pub mod %s\n\n", pkg.Name), indent)
		re.gir.emitToFileBuffer(str, EmptyVisitMethod)
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
	// Check if this specific struct has interface{} fields or function fields
	// (can't derive Clone/Default/Debug for these)
	var str string
	hasInterfaceFields := re.structHasInterfaceFields(node.Name)
	hasFunctionFields := re.structHasFunctionFields(node.Name)
	if hasFunctionFields {
		// Structs with function fields can derive Clone (Rc implements Clone)
		// but not Default or Debug (dyn Fn doesn't implement these)
		str = re.emitAsString("#[derive(Clone)]\n", indent+2)
	} else if hasInterfaceFields {
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
										// Check for function fields (Box<dyn Fn> in Rust doesn't implement Clone)
										if strings.Contains(typeStr, "func(") {
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
										// If field is a function type (will become Box<dyn Fn...>), can't derive Copy
										if strings.HasPrefix(typeStr, "func(") {
											return false
										}
										// If field is a struct type, can't safely derive Copy
										// (the nested struct might have non-Copy fields)
										if named, ok := fieldType.(*types.Named); ok {
											if _, isStruct := named.Underlying().(*types.Struct); isStruct {
												return false
											}
										}
										// Check underlying type for function signatures
										if _, isSig := fieldType.Underlying().(*types.Signature); isSig {
											return false
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

// structHasFunctionFields checks if a struct has function/closure fields
func (re *RustEmitter) structHasFunctionFields(structName string) bool {
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
										// Check for function types (will become Box<dyn Fn...>)
										if strings.HasPrefix(typeStr, "func(") {
											return true
										}
										// Check underlying type for function signatures
										if _, isSig := fieldType.Underlying().(*types.Signature); isSig {
											return true
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
	// Use Rc<dyn Fn> for function types - Rc implements Clone so structs with function fields can be cloned
	str := re.emitAsString("Rc<dyn Fn(", indent)
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
			reorderedTokens = append(reorderedTokens, tokens[0]) // "Rc<dyn Fn("
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
	// For builtin package names (like fmt), suppress generation
	// For user-defined module names (like types, ast), generate them
	if ident, ok := node.(*ast.Ident); ok {
		obj := re.pkg.TypesInfo.Uses[ident]
		if obj != nil {
			if _, ok := obj.(*types.PkgName); ok {
				// Check if this is a builtin package that gets lowered
				if re.lowerToBuiltins(ident.Name) == "" {
					// Builtin package (fmt) - suppress generation
					re.shouldGenerate = false
					return
				}
				// User-defined module - let it be generated
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
	isBuiltinPackage := false
	if ident, ok := node.(*ast.Ident); ok {
		// Check if this is a builtin package (like fmt) that we lower to crate-level functions
		if re.lowerToBuiltins(ident.Name) == "" {
			// This is a builtin package like "fmt"
			isBuiltinPackage = true
		}

		// Check if this is a package name - use :: for module-qualified access
		obj := re.pkg.TypesInfo.Uses[ident]
		if obj != nil {
			if _, ok := obj.(*types.PkgName); ok {
				// For builtin packages (fmt), don't emit any operator
				// The selector will be lowered to a crate-level function
				if isBuiltinPackage {
					re.shouldGenerate = true
					return
				}
				// Use :: for module-qualified access in Rust
				scopeOperator = "::"
			}
		}
		// Also check if the identifier is a known namespace/module
		if _, found := namespaces[ident.Name]; found {
			scopeOperator = "::"
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
		typeInfo := re.pkg.TypesInfo.Types[node.Type]
		// Only do alias replacement if the type is NOT already a named type (alias)
		// If it's a named type like ast.AST, don't replace it with another alias
		if typeInfo.Type != nil {
			if _, isNamed := typeInfo.Type.(*types.Named); !isNamed {
				// Type is a basic/primitive type - check for alias replacement
				for aliasName, alias := range re.aliases {
					if alias.UnderlyingType == typeInfo.Type.Underlying().String() {
						re.gir.tokenSlice, _ = RewriteTokensBetween(re.gir.tokenSlice, pointerAndPosition.Index, len(re.gir.tokenSlice), []string{aliasName})
						break
					}
				}
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
	str := re.emitAsString("pub type ", indent+2)
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
	re.inReturnStmt = true
	str := re.emitAsString("return ", indent)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)

	if len(node.Results) > 1 {
		re.inMultiValueReturn = true
		re.multiValueReturnResultIndex = 0
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
	re.inMultiValueReturn = false
	re.inReturnStmt = false
	str := re.emitAsString(";", 0)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
}

func (re *RustEmitter) PreVisitReturnStmtResult(node ast.Expr, index int, indent int) {
	if index > 0 {
		str := re.emitAsString(", ", 0)
		re.gir.emitToFileBuffer(str, EmptyVisitMethod)
	}
	re.multiValueReturnResultIndex = index

	// If returning from a function that returns any (Box<dyn Any>),
	// and the return value is a concrete type, wrap in Box::new()
	if re.currentFuncReturnsAny && node != nil {
		nodeType := re.pkg.TypesInfo.Types[node]
		if nodeType.Type != nil {
			typeStr := nodeType.Type.String()
			// Don't wrap if already Box<dyn Any> or interface{}
			if typeStr != "interface{}" && typeStr != "any" && !strings.Contains(typeStr, "Box<dyn Any>") {
				re.gir.emitToFileBuffer("Box::new(", EmptyVisitMethod)
			}
		}
	}
}

func (re *RustEmitter) PostVisitReturnStmtResult(node ast.Expr, index int, indent int) {
	if re.forwardDecls {
		return
	}
	// Add .clone() to the first result in a multi-value return if it's an identifier
	// This prevents "borrow of moved value" errors when subsequent results reference fields
	if re.inMultiValueReturn && index == 0 {
		if _, ok := node.(*ast.Ident); ok {
			re.gir.emitToFileBuffer(".clone()", EmptyVisitMethod)
		}
	}

	// Close Box::new() if we opened it in Pre
	if re.currentFuncReturnsAny && node != nil {
		nodeType := re.pkg.TypesInfo.Types[node]
		if nodeType.Type != nil {
			typeStr := nodeType.Type.String()
			if typeStr != "interface{}" && typeStr != "any" && !strings.Contains(typeStr, "Box<dyn Any>") {
				re.gir.emitToFileBuffer(")", EmptyVisitMethod)
			}
		}
	}
}

func (re *RustEmitter) PreVisitCallExpr(node *ast.CallExpr, indent int) {
	re.shouldGenerate = true
	// In += context, string functions return String but += expects &str
	// Add & before calls to string_format (Sprintf)
	if re.inAssignRhs && re.assignmentToken == "+=" {
		if sel, ok := node.Fun.(*ast.SelectorExpr); ok {
			if sel.Sel.Name == "Sprintf" {
				re.gir.emitToFileBuffer("&", EmptyVisitMethod)
			}
		}
	}
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

	// For += with String, Rust expects &str on RHS
	// Add & before RHS if it's a String variable (not a string literal)
	if re.assignmentToken == "+=" && len(node.Rhs) == 1 {
		rhsType := re.pkg.TypesInfo.Types[node.Rhs[0]]
		if rhsType.Type != nil && rhsType.Type.String() == "string" {
			// Check if RHS is not a string literal (literals are handled separately)
			if _, isBasicLit := node.Rhs[0].(*ast.BasicLit); !isBasicLit {
				re.gir.emitToFileBuffer("&", EmptyVisitMethod)
			}
		}
	}

	// If assigning to a variable of type any (interface{}), wrap RHS in Box::new()
	if len(node.Lhs) == 1 && len(node.Rhs) == 1 && re.assignmentToken == "=" {
		lhsType := re.pkg.TypesInfo.Types[node.Lhs[0]]
		rhsType := re.pkg.TypesInfo.Types[node.Rhs[0]]
		if lhsType.Type != nil && rhsType.Type != nil {
			lhsTypeStr := lhsType.Type.String()
			rhsTypeStr := rhsType.Type.String()
			// If LHS is any/interface{} and RHS is a concrete type, wrap in Box::new()
			if (lhsTypeStr == "interface{}" || lhsTypeStr == "any") &&
				rhsTypeStr != "interface{}" && rhsTypeStr != "any" {
				re.gir.emitToFileBuffer("Box::new(", EmptyVisitMethod)
			}
		}
	}
}

func (re *RustEmitter) PostVisitAssignStmtRhs(node *ast.AssignStmt, indent int) {
	// Check if we need to add a type cast for constant assignments
	// This handles untyped int constants assigned to i8 variables
	if len(node.Lhs) == 1 && len(node.Rhs) == 1 {
		// Get the LHS type
		lhsType := re.pkg.TypesInfo.Types[node.Lhs[0]]
		rhsType := re.pkg.TypesInfo.Types[node.Rhs[0]]
		if lhsType.Type != nil {
			lhsTypeStr := lhsType.Type.String()

			// Close Box::new() if we opened it for any/interface{} assignment
			if rhsType.Type != nil {
				rhsTypeStr := rhsType.Type.String()
				if (lhsTypeStr == "interface{}" || lhsTypeStr == "any") &&
					rhsTypeStr != "interface{}" && rhsTypeStr != "any" &&
					re.assignmentToken == "=" {
					re.gir.emitToFileBuffer(")", EmptyVisitMethod)
				}
			}

			// Check if RHS is a constant identifier
			if rhsIdent, ok := node.Rhs[0].(*ast.Ident); ok {
				if obj := re.pkg.TypesInfo.Uses[rhsIdent]; obj != nil {
					if _, isConst := obj.(*types.Const); isConst {
						// Get the constant type
						constType := obj.Type().String()
						// If assigning int constant to int8 field/variable, add cast
						if (constType == "int" || constType == "untyped int") && lhsTypeStr == "int8" {
							re.gir.emitToFileBuffer(" as i8", EmptyVisitMethod)
						}
					}
				}
			}
		}
	}

	// For local closure assignments, remove the entire statement from token stream
	// The body tokens have already been stored in PostVisitFuncLit
	// Only truncate for the outer closure assignment, not inner assignments
	// inLocalClosureBody is false after PostVisitFuncLit, true while inside closure body
	if re.localClosureAssign && re.currentClosureName != "" && !re.inLocalClosureBody {
		// Remove all tokens from the assignment start to current position
		if re.localClosureAssignStartIndex < len(re.gir.tokenSlice) {
			re.gir.tokenSlice = re.gir.tokenSlice[:re.localClosureAssignStartIndex]
		}
		// Reset flags
		re.localClosureAssign = false
		re.currentClosureName = ""
	}

	re.shouldGenerate = false
	re.isTuple = false
	re.inAssignRhs = false
}

func (re *RustEmitter) PreVisitAssignStmtLhsExpr(node ast.Expr, index int, indent int) {
	if index > 0 {
		str := re.emitAsString(", ", indent)
		re.gir.emitToFileBuffer(str, EmptyVisitMethod)
		// For multi-value declarations, add mut before each subsequent variable
		if re.inMultiValueDecl {
			re.emitToken("mut", RustKeyword, 0)
			re.emitToken(" ", WhiteSpace, 0)
		}
	}
}

func (re *RustEmitter) PreVisitAssignStmtLhs(node *ast.AssignStmt, indent int) {
	re.shouldGenerate = true
	re.inAssignLhs = true // Track that we're in LHS
	assignmentToken := node.Tok.String()
	// Check if LHS is a field access (SelectorExpr)
	re.inFieldAssign = false
	if len(node.Lhs) == 1 {
		if _, ok := node.Lhs[0].(*ast.SelectorExpr); ok {
			re.inFieldAssign = true
		}
	}
	// Check if RHS is a function literal (for local closure inlining)
	// Don't reset if we're inside a local closure body being processed
	if !re.inLocalClosureBody {
		re.localClosureAssign = false
		re.currentClosureName = ""
	}
	if assignmentToken == ":=" && len(node.Rhs) == 1 {
		if funcLit, ok := node.Rhs[0].(*ast.FuncLit); ok {
			if ident, ok := node.Lhs[0].(*ast.Ident); ok {
				re.localClosureAssign = true
				re.currentClosureName = ident.Name
				re.localClosures[ident.Name] = funcLit
				// Record assignment start index for later removal
				re.localClosureAssignStartIndex = len(re.gir.tokenSlice)
				// Skip emitting the assignment - we'll inline the closure body at call sites
				re.shouldGenerate = false
				return
			}
		}
	}
	re.inMultiValueDecl = false
	if assignmentToken == ":=" && len(node.Lhs) == 1 {
		re.emitToken("let", RustKeyword, indent)
		re.emitToken(" ", WhiteSpace, 0)
		re.emitToken("mut", RustKeyword, 0)
		re.emitToken(" ", WhiteSpace, 0)
	} else if assignmentToken == ":=" && len(node.Lhs) > 1 {
		// Multi-value declaration: let (mut a, mut b) = ...
		re.inMultiValueDecl = true
		re.emitToken("let", RustKeyword, indent)
		re.emitToken(" ", WhiteSpace, 0)
		re.emitToken("(", LeftParen, 0)
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
	// For local closure inlining, skip all emission - the assignment will be removed
	if re.localClosureAssign && re.currentClosureName != "" {
		re.shouldGenerate = false
		return
	}
	if node.Tok.String() == ":=" && len(node.Lhs) > 1 {
		re.emitToken(")", RightParen, indent)
	} else if node.Tok.String() == "=" && len(node.Lhs) > 1 {
		re.emitToken(")", RightParen, indent)
	}
	re.inAssignLhs = false // Done with LHS
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
	// If the base expression is a string, we need .as_bytes() for indexing
	if node.X != nil {
		tv := re.pkg.TypesInfo.Types[node.X]
		if tv.Type != nil && tv.Type.String() == "string" {
			re.gir.emitToFileBuffer(".as_bytes()", EmptyVisitMethod)
		}
	}
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

	// Add .clone() for Vec element access when the element type doesn't implement Copy
	// This is needed because Rust doesn't allow moving out of indexed collections
	// BUT: Don't add .clone() when we're in the LHS of an assignment (we're assigning TO it)
	if node.X != nil && !re.inAssignLhs {
		tv := re.pkg.TypesInfo.Types[node.X]
		if tv.Type != nil {
			// Check if it's a slice/array type
			underlying := tv.Type.Underlying()
			if sliceType, ok := underlying.(*types.Slice); ok {
				elemType := sliceType.Elem()
				// Check if element type is a struct (non-Copy type)
				if _, isStruct := elemType.Underlying().(*types.Struct); isStruct {
					re.gir.emitToFileBuffer(".clone()", EmptyVisitMethod)
				}
			}
		}
	}
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
	// Check if the argument type needs .clone()
	tv := re.pkg.TypesInfo.Types[node]
	if tv.Type != nil {
		typeStr := tv.Type.String()

		// Clone Vec/slice types
		if strings.HasPrefix(typeStr, "[]") {
			re.gir.emitToFileBuffer(".clone()", EmptyVisitMethod)
			return
		}

		// Clone String types (but not string literals - those get .to_string() anyway)
		if typeStr == "string" {
			if _, isBasicLit := node.(*ast.BasicLit); !isBasicLit {
				re.gir.emitToFileBuffer(".clone()", EmptyVisitMethod)
				return
			}
		}

		// Check if it's a named type (potential struct)
		if named, ok := tv.Type.(*types.Named); ok {
			// Check if underlying type is a struct
			if underlyingStruct, isStruct := named.Underlying().(*types.Struct); isStruct {
				// Check if any field has interface{} type (Box<dyn Any> doesn't implement Clone)
				// Note: function fields now use Rc<dyn Fn> which implements Clone
				hasNonClonableField := false
				for i := 0; i < underlyingStruct.NumFields(); i++ {
					field := underlyingStruct.Field(i)
					fieldTypeStr := field.Type().String()
					if strings.Contains(fieldTypeStr, "interface{}") || strings.Contains(fieldTypeStr, "interface {") {
						hasNonClonableField = true
						break
					}
				}
				if hasNonClonableField {
					// Don't clone structs with interface fields (Box<dyn Any> doesn't implement Clone)
					return
				}
				// Clone all other structs (including those with function fields - Rc implements Clone)
				re.gir.emitToFileBuffer(".clone()", EmptyVisitMethod)
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

	// Detect loop type upfront to emit correct Rust keyword
	var str string
	if node.Init == nil && node.Cond == nil && node.Post == nil {
		// Infinite loop: for { } -> loop { }
		str = re.emitAsString("loop", indent)
		re.isInfiniteLoop = true
	} else {
		str = re.emitAsString("for ", indent)
		re.isInfiniteLoop = false
	}
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
	re.shouldGenerate = true
}

func (re *RustEmitter) PostVisitForStmtInit(node ast.Stmt, indent int) {
	// Don't emit semicolon for infinite loops (they use `loop` keyword)
	if re.isInfiniteLoop {
		return
	}
	if node == nil {
		str := re.emitAsString(";", 0)
		re.gir.emitToFileBuffer(str, EmptyVisitMethod)
	}
}

func (re *RustEmitter) PostVisitForStmtPost(node ast.Stmt, indent int) {
	re.insideForPostCond = false
	str := re.emitAsString("\n", 0)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
}

func (re *RustEmitter) PreVisitIfStmtElse(node *ast.IfStmt, indent int) {
	str := re.emitAsString("else", 1)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
}

func (re *RustEmitter) PostVisitForStmtCond(node ast.Expr, indent int) {
	// Don't emit semicolon for infinite loops (they use `loop` keyword)
	if !re.isInfiniteLoop {
		str := re.emitAsString(";", 0)
		re.gir.emitToFileBuffer(str, EmptyVisitMethod)
	}
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
				// Extract ALL tokens after < until end (excluding final semicolon)
				// This handles complex expressions like (newState.depth as i32)
				upperBound := condTokens[i+1:]
				// Remove trailing semicolons and whitespace (but not parens yet)
				for len(upperBound) > 0 {
					lastTok := upperBound[len(upperBound)-1]
					trimmed := strings.TrimSpace(lastTok.Content)
					if trimmed == ";" || trimmed == "" {
						upperBound = upperBound[:len(upperBound)-1]
					} else {
						break
					}
				}
				// Combine all upper bound tokens into a single token for simplicity
				upperBoundStr := ""
				for _, t := range upperBound {
					upperBoundStr += t.Content
				}
				upperBoundStr = strings.TrimSpace(upperBoundStr)
				// Count parens and strip only UNMATCHED trailing )
				// The BinaryExpr wrapper adds one extra ) that we need to remove
				for strings.HasSuffix(upperBoundStr, ")") {
					openCount := strings.Count(upperBoundStr, "(")
					closeCount := strings.Count(upperBoundStr, ")")
					if closeCount > openCount {
						upperBoundStr = strings.TrimSuffix(upperBoundStr, ")")
						upperBoundStr = strings.TrimSpace(upperBoundStr)
					} else {
						break
					}
				}
				if upperBoundStr != "" {
					rangeTokens = append(rangeTokens, CreateToken(Identifier, upperBoundStr))
				}
				break // Only process first < operator
			}
		}
	}

	p6 := SearchPointerIndexReverse(PostVisitForStmtPost, re.gir.pointerAndIndexVec)
	pFor := SearchPointerIndexReverse(PreVisitForStmt, re.gir.pointerAndIndexVec)

	// Case 0: Infinite loop (no init, no cond, no post) → loop
	// Go: for { } → Rust: loop { }
	if pFor != nil && p6 != nil && !hasInit && !hasCond && node.Post == nil {
		// Build new tokens for Rust loop: "loop\n"
		newTokens := []string{"loop\n"}
		re.gir.tokenSlice, _ = RewriteTokensBetween(re.gir.tokenSlice, pFor.Index, p6.Index, newTokens)
		return
	}

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
	// Check the type of the expression being ranged over
	tv := re.pkg.TypesInfo.Types[node]
	if tv.Type != nil {
		typeStr := tv.Type.String()
		if typeStr == "string" {
			// String needs .bytes() to iterate and get i8 values
			re.gir.emitToFileBuffer(".bytes()", EmptyVisitMethod)
		} else {
			// Add .clone() to the collection to avoid ownership transfer
			re.gir.emitToFileBuffer(".clone()", EmptyVisitMethod)
		}
	} else {
		re.gir.emitToFileBuffer(".clone()", EmptyVisitMethod)
	}
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
	// Save current isArray state for nested composite literals
	re.isArrayStack = append(re.isArrayStack, re.isArray)
	// Reset for this composite literal
	re.isArray = false
	re.currentCompLitIsSlice = false

	// Push the type to the stack so we can check it in PostVisitCompositeLitElts
	var compLitType types.Type
	if node.Type != nil {
		typeInfo := re.pkg.TypesInfo.Types[node.Type]
		if typeInfo.Type != nil {
			compLitType = typeInfo.Type
			// Check if the underlying type is a slice (for type aliases like AST = []Statement)
			if underlying := compLitType.Underlying(); underlying != nil {
				if _, ok := underlying.(*types.Slice); ok {
					// Only use Vec::new() for empty slice literals
					// For non-empty, set isArray so vec![] syntax is used
					if len(node.Elts) == 0 {
						re.currentCompLitIsSlice = true
					} else {
						re.isArray = true
					}
				}
			}
		}
	}
	re.compLitTypeStack = append(re.compLitTypeStack, compLitType)
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
		// For slice type aliases (like AST = []Statement), replace with Vec::new()
		// The braces will be suppressed in PreVisitCompositeLitElts/PostVisitCompositeLitElts
		if re.currentCompLitIsSlice {
			if re.inKeyValueExpr || re.inFieldAssign || re.inReturnStmt {
				// Inside struct field initialization, field assignment, or return statement
				re.gir.tokenSlice, _ = RewriteTokensBetween(re.gir.tokenSlice, pointerAndPosition.Index, len(re.gir.tokenSlice), []string{"Vec::new()"})
			} else {
				// Variable declaration: let x = []Type{} -> let x: Vec<type> = Vec::new()
				// Extract the type tokens for the type annotation
				vecTypeStrRepr, _ := ExtractTokensBetween(pointerAndPosition.Index, len(re.gir.tokenSlice), re.gir.tokenSlice)
				newTokens := []string{}
				newTokens = append(newTokens, ":")
				newTokens = append(newTokens, tokensToStrings(vecTypeStrRepr)...)
				newTokens = append(newTokens, " = Vec::new()")
				re.gir.tokenSlice, _ = RewriteTokensBetween(re.gir.tokenSlice, pointerAndPosition.Index-len("=")-len(" "), len(re.gir.tokenSlice), newTokens)
			}
			return
		}
		// TODO not very effective
		// go through all aliases and check if the underlying type matches
		// Only do alias replacement if the type is NOT already a named type (alias)
		typeInfo := re.pkg.TypesInfo.Types[node]
		if typeInfo.Type != nil {
			if _, isNamed := typeInfo.Type.(*types.Named); !isNamed {
				// Type is a basic/primitive type - check for alias replacement
				for aliasName, alias := range re.aliases {
					if alias.UnderlyingType == typeInfo.Type.Underlying().String() {
						re.gir.tokenSlice, _ = RewriteTokensBetween(re.gir.tokenSlice, pointerAndPosition.Index, len(re.gir.tokenSlice), []string{aliasName})
						break
					}
				}
			}
		}
		if re.isArray {
			// TODO that's still hack
			// we operate on string representation of the type
			// has to be rewritten to use some kind of IR
			if re.inKeyValueExpr || re.inFieldAssign || re.inReturnStmt {
				// Inside struct field initialization, field assignment, or return statement: []Type{} -> vec![]
				// Just replace the type with vec!, keeping context intact
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
	// Skip braces for slice type aliases - Vec::new() is already emitted
	if re.currentCompLitIsSlice {
		return
	}
	re.emitToken("{", LeftBrace, 0)
}

func (re *RustEmitter) PostVisitCompositeLitElts(node []ast.Expr, indent int) {
	// For struct initialization in Rust, add ..Default::default() to handle missing fields
	// But NOT for arrays/vectors - those just use {}
	// Don't use Default::default() if the type doesn't derive Default (due to interface{}/func fields)

	// Get the current composite literal's type from the stack
	var currentType types.Type
	if len(re.compLitTypeStack) > 0 {
		currentType = re.compLitTypeStack[len(re.compLitTypeStack)-1]
	}

	// Check if this specific type has interface/function fields that prevent Default
	typeHasNoDefault := currentType != nil && re.typeHasInterfaceFields(currentType)

	// Check if the underlying type is a slice (for type aliases like AST = []Statement)
	isSliceType := false
	if currentType != nil {
		underlying := currentType.Underlying()
		if _, ok := underlying.(*types.Slice); ok {
			isSliceType = true
		}
	}

	// Skip braces and default for slice type aliases - Vec::new() is already emitted
	if re.currentCompLitIsSlice {
		re.currentCompLitIsSlice = false
	} else {
		if !re.isArray && !isSliceType && !typeHasNoDefault {
			if len(node) > 0 {
				// Partial struct init - add comma before Default
				re.gir.emitToFileBuffer(", ..Default::default()", EmptyVisitMethod)
			} else {
				// Empty struct init
				re.gir.emitToFileBuffer("..Default::default()", EmptyVisitMethod)
			}
		}
		re.emitToken("}", RightBrace, 0)
	}

	// Restore isArray from stack for nested composite literals
	if len(re.isArrayStack) > 0 {
		re.isArray = re.isArrayStack[len(re.isArrayStack)-1]
		re.isArrayStack = re.isArrayStack[:len(re.isArrayStack)-1]
	}
	// Pop from compLitTypeStack
	if len(re.compLitTypeStack) > 0 {
		re.compLitTypeStack = re.compLitTypeStack[:len(re.compLitTypeStack)-1]
	}
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

func (re *RustEmitter) PreVisitSliceExprLow(node ast.Expr, indent int) {
	// Re-enable generation for the low index expression
	re.shouldGenerate = true
}

func (re *RustEmitter) PostVisitSliceExprLow(node ast.Expr, indent int) {
	// Cast to usize for slice indexing (Rust requires usize for slice indices)
	if node != nil {
		re.gir.emitToFileBuffer(" as usize", EmptyVisitMethod)
	}
	re.gir.emitToFileBuffer("..", EmptyVisitMethod)
	re.shouldGenerate = false
}

func (re *RustEmitter) PreVisitFuncLit(node *ast.FuncLit, indent int) {
	// For local closure inlining, skip wrapper emission
	if re.localClosureAssign && re.currentClosureName != "" {
		return
	}
	// Check if this closure returns any (interface{})
	re.currentFuncReturnsAny = false
	if node.Type != nil && node.Type.Results != nil {
		for _, result := range node.Type.Results.List {
			if result.Type != nil {
				resultType := re.pkg.TypesInfo.Types[result.Type]
				if resultType.Type != nil {
					typeStr := resultType.Type.String()
					if typeStr == "interface{}" || typeStr == "any" {
						re.currentFuncReturnsAny = true
						break
					}
				}
			}
		}
	}
	// Wrap closure with Rc::new() since all function types use Rc<dyn Fn>
	wrapperStr := "Rc::new("
	str := re.emitAsString(wrapperStr, 0)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
	re.emitToken("|", Identifier, indent)
}
func (re *RustEmitter) PostVisitFuncLit(node *ast.FuncLit, indent int) {
	// For local closures being inlined, extract and store body tokens, skip wrapper
	if re.inLocalClosureBody && re.currentClosureName != "" {
		// Extract body tokens (from after { to current position)
		bodyEndIndex := len(re.gir.tokenSlice)
		if re.localClosureBodyStartIndex < bodyEndIndex {
			bodyTokens := make([]Token, bodyEndIndex-re.localClosureBodyStartIndex)
			copy(bodyTokens, re.gir.tokenSlice[re.localClosureBodyStartIndex:bodyEndIndex])
			re.localClosureBodyTokens[re.currentClosureName] = bodyTokens
		}
		// Clear the flag
		re.inLocalClosureBody = false
		// Don't emit the closing braces - the entire assignment will be truncated
		return
	}
	re.emitToken("}", RightBrace, 0)
	// Close the Rc::new() or Box::new() wrapper
	re.emitToken(")", RightParen, 0)
	re.currentFuncReturnsAny = false
}

func (re *RustEmitter) PostVisitFuncLitTypeParams(node *ast.FieldList, indent int) {
	// For local closure inlining, skip wrapper emission
	if re.localClosureAssign && re.currentClosureName != "" {
		return
	}
	re.emitToken("|", Identifier, 0)
}

func (re *RustEmitter) PreVisitFuncLitTypeParam(node *ast.Field, index int, indent int) {
	str := ""
	if index > 0 {
		str += re.emitAsString(", ", 0)
	}
	// Emit name first, then colon, then type will follow
	if len(node.Names) > 0 {
		// Escape Rust keywords in parameter names
		paramName := escapeRustKeyword(node.Names[0].Name)
		str += re.emitAsString(paramName+": ", 0)
	}
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
}

func (re *RustEmitter) PostVisitFuncLitTypeParam(node *ast.Field, index int, indent int) {
	// Type has already been emitted, nothing to do
}

func (re *RustEmitter) PreVisitFuncLitBody(node *ast.BlockStmt, indent int) {
	// For local closures being inlined, skip wrapper emission but record body start
	if re.localClosureAssign && re.currentClosureName != "" {
		re.localClosureBodyStartIndex = len(re.gir.tokenSlice)
		re.inLocalClosureBody = true // Track that we're inside the closure body
		return
	}
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

func (re *RustEmitter) PostVisitKeyValueExprValue(node ast.Expr, indent int) {
	// Add .clone() for non-Copy types in struct field assignments
	// This is needed because Rust closures that move values become FnOnce, not Fn
	// For slices (Vec in Rust), we need to clone to avoid moving the captured variable
	if node != nil {
		tv := re.pkg.TypesInfo.Types[node]
		if tv.Type != nil {
			typeStr := tv.Type.String()
			// Check if it's a slice type (will become Vec in Rust)
			if strings.HasPrefix(typeStr, "[]") {
				re.gir.emitToFileBuffer(".clone()", EmptyVisitMethod)
			}
		}
	}
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
			// Map Go types to Rust types for constants
			rustType := re.mapGoTypeToRust(constType)
			str := re.emitAsString(fmt.Sprintf("pub const %s: %s = ", node.Name, rustType), 0)

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
	str := re.emitAsString("match ", indent)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
}
func (re *RustEmitter) PostVisitSwitchStmt(node *ast.SwitchStmt, indent int) {
	// Check if the switch has a default case
	hasDefault := false
	if node.Body != nil {
		for _, stmt := range node.Body.List {
			if caseClause, ok := stmt.(*ast.CaseClause); ok {
				if len(caseClause.List) == 0 {
					hasDefault = true
					break
				}
			}
		}
	}
	// If no default case, add one for Rust match exhaustiveness
	if !hasDefault {
		str := re.emitAsString("_ => {}\n", indent+2)
		re.gir.emitToFileBuffer(str, EmptyVisitMethod)
	}
	re.emitToken("}", RightBrace, indent)
}

func (re *RustEmitter) PostVisitSwitchStmtTag(node ast.Expr, indent int) {
	// Check if we need to cast to i32 to match constants (which are i32 by default)
	if node != nil {
		tv := re.pkg.TypesInfo.Types[node]
		if tv.Type != nil {
			typeStr := tv.Type.String()
			// Cast smaller integer types to i32 so they match constant types
			if typeStr == "int8" || typeStr == "uint8" ||
				typeStr == "int16" || typeStr == "uint16" {
				str := re.emitAsString(" as i32", 0)
				re.gir.emitToFileBuffer(str, EmptyVisitMethod)
			}
		}
	}
	// Rust match doesn't use parentheses around the tag
	str := re.emitAsString(" ", 0)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
	re.emitToken("{", LeftBrace, 0)
	str = re.emitAsString("\n", 0)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
}

func (re *RustEmitter) PostVisitCaseClause(node *ast.CaseClause, indent int) {
	// In Rust match, close the block for this arm
	str := re.emitAsString("}\n", indent+2)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
}

func (re *RustEmitter) PostVisitCaseClauseList(node []ast.Expr, indent int) {
	if len(node) == 0 {
		// Rust match uses _ for default case
		str := re.emitAsString("_ => {\n", indent+2)
		re.gir.emitToFileBuffer(str, EmptyVisitMethod)
	}
}

func (re *RustEmitter) PreVisitCaseClauseListExpr(node ast.Expr, index int, indent int) {
	// Rust match arms don't need "case" keyword - just the pattern
	str := re.emitAsString("", indent+2)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
	re.shouldGenerate = true
}

func (re *RustEmitter) PostVisitCaseClauseListExpr(node ast.Expr, index int, indent int) {
	// Rust match uses => and block
	str := re.emitAsString(" => {\n", 0)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
}

func (re *RustEmitter) PreVisitTypeAssertExpr(node *ast.TypeAssertExpr, indent int) {
	re.gir.emitToFileBuffer("", "@PreVisitTypeAssertExpr")
}

func (re *RustEmitter) PreVisitTypeAssertExprType(node ast.Expr, indent int) {
	re.gir.emitToFileBuffer("", "@PreVisitTypeAssertExprType")
}

func (re *RustEmitter) PostVisitTypeAssertExprType(node ast.Expr, indent int) {
	re.gir.emitToFileBuffer("", "@PostVisitTypeAssertExprType")
}

func (re *RustEmitter) PreVisitTypeAssertExprX(node ast.Expr, indent int) {
	re.gir.emitToFileBuffer("", "@PreVisitTypeAssertExprX")
}

func (re *RustEmitter) PostVisitTypeAssertExprX(node ast.Expr, indent int) {
	re.gir.emitToFileBuffer("", "@PostVisitTypeAssertExprX")
}

func (re *RustEmitter) PostVisitTypeAssertExpr(node *ast.TypeAssertExpr, indent int) {
	// Reorder type assertion from (Type)X to X.downcast_ref::<Type>().unwrap().clone()
	p1 := SearchPointerIndexReverseString("@PreVisitTypeAssertExprType", re.gir.pointerAndIndexVec)
	p2 := SearchPointerIndexReverseString("@PostVisitTypeAssertExprType", re.gir.pointerAndIndexVec)
	p3 := SearchPointerIndexReverseString("@PreVisitTypeAssertExprX", re.gir.pointerAndIndexVec)
	p4 := SearchPointerIndexReverseString("@PostVisitTypeAssertExprX", re.gir.pointerAndIndexVec)
	p0 := SearchPointerIndexReverseString("@PreVisitTypeAssertExpr", re.gir.pointerAndIndexVec)

	if p0 != nil && p1 != nil && p2 != nil && p3 != nil && p4 != nil {
		typeTokens, err := ExtractTokensBetween(p1.Index, p2.Index, re.gir.tokenSlice)
		if err != nil {
			return
		}
		exprTokens, err := ExtractTokensBetween(p3.Index, p4.Index, re.gir.tokenSlice)
		if err != nil {
			return
		}
		typeStr := strings.TrimSpace(strings.Join(tokensToStrings(typeTokens), ""))
		exprStr := strings.TrimSpace(strings.Join(tokensToStrings(exprTokens), ""))

		// Generate Rust downcast syntax: X.downcast_ref::<Type>().unwrap().clone()
		newTokens := []string{exprStr, ".downcast_ref::<", typeStr, ">().unwrap().clone()"}
		re.gir.tokenSlice, _ = RewriteTokensBetween(re.gir.tokenSlice, p0.Index, p4.Index, newTokens)
	}
}

func (re *RustEmitter) PreVisitKeyValueExpr(node *ast.KeyValueExpr, indent int) {
	re.shouldGenerate = true
	re.inKeyValueExpr = true
}

func (re *RustEmitter) PostVisitKeyValueExpr(node *ast.KeyValueExpr, indent int) {
	re.inKeyValueExpr = false
	// Add type cast if needed for struct field initialization
	// This handles untyped int constants assigned to i8 fields
	if node.Value != nil {
		// Get the key (field name)
		if keyIdent, ok := node.Key.(*ast.Ident); ok {
			fieldName := keyIdent.Name
			// Use heuristic based on field name for common i8 fields
			if fieldName == "Type" {
				// Check if this is an identifier (likely a constant) being assigned
				if valueIdent, ok := node.Value.(*ast.Ident); ok {
					// Check if the identifier refers to an object with int type
					obj := re.pkg.TypesInfo.Uses[valueIdent]
					if obj != nil {
						objType := obj.Type().String()
						// Cast int constants to i8 for Type fields
						if objType == "int" || objType == "untyped int" || strings.HasSuffix(objType, ".int") {
							re.gir.emitToFileBuffer(" as i8", EmptyVisitMethod)
						}
					} else {
						// If obj is nil but it looks like a constant name (starts with uppercase), add cast
						// This handles cross-package constant references that might not be resolved
						if len(valueIdent.Name) > 0 && valueIdent.Name[0] >= 'A' && valueIdent.Name[0] <= 'Z' {
							re.gir.emitToFileBuffer(" as i8", EmptyVisitMethod)
						}
					}
				} else if selExpr, ok := node.Value.(*ast.SelectorExpr); ok {
					// Handle package-qualified constants like ast.StatementTypeFrom
					obj := re.pkg.TypesInfo.Uses[selExpr.Sel]
					if obj != nil {
						if _, isConst := obj.(*types.Const); isConst {
							objType := obj.Type().String()
							// Cast int constants to i8 for Type fields
							if objType == "int" || objType == "untyped int" || strings.HasSuffix(objType, ".int") {
								re.gir.emitToFileBuffer(" as i8", EmptyVisitMethod)
							}
						}
					}
				}
			}
		}
	}
}

func (re *RustEmitter) PreVisitBranchStmt(node *ast.BranchStmt, indent int) {
	str := re.emitAsString(node.Tok.String()+";", indent)
	re.gir.emitToFileBuffer(str, EmptyVisitMethod)
}

func (re *RustEmitter) PreVisitCallExprFun(node ast.Expr, indent int) {
	// Check if this is a selector expression (obj.field) where the field is a function type
	// In Rust, calling a function stored in a struct field requires: (obj.field)(args)
	if sel, ok := node.(*ast.SelectorExpr); ok {
		// Get the type of the selector (the field)
		if tv := re.pkg.TypesInfo.Selections[sel]; tv != nil {
			// Check if the field type is a function type (Signature)
			if _, isSig := tv.Type().Underlying().(*types.Signature); isSig {
				re.gir.emitToFileBuffer("(", EmptyVisitMethod)
			}
		}
	}
	// Push the current position to the stack for nested call handling
	re.callExprFunMarkerStack = append(re.callExprFunMarkerStack, len(re.gir.tokenSlice))
	re.gir.emitToFileBuffer("", "@PreVisitCallExprFun")
}

func (re *RustEmitter) PostVisitCallExprFun(node ast.Expr, indent int) {
	// Push the current position to the stack for nested call handling (end of function name)
	re.callExprFunEndMarkerStack = append(re.callExprFunEndMarkerStack, len(re.gir.tokenSlice))
	re.gir.emitToFileBuffer("", "@PostVisitCallExprFun")
	// Close the paren if we opened one for function field call
	if sel, ok := node.(*ast.SelectorExpr); ok {
		if tv := re.pkg.TypesInfo.Selections[sel]; tv != nil {
			if _, isSig := tv.Type().Underlying().(*types.Signature); isSig {
				re.gir.emitToFileBuffer(")", EmptyVisitMethod)
			}
		}
	}
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
		// Only check name for whitespace - types can have spaces (e.g., Box<dyn Any>)
		nameStr := strings.TrimSpace(strings.Join(tokensToStrings(nameStrRepr), ""))
		if nameStr == "" || containsWhitespace(nameStr) {
			// If name is empty or has whitespace, skip reordering
			return
		}
		newTokens := []string{}
		// Add mut for non-primitive types (struct parameters are often modified in Go)
		typeStr := strings.TrimSpace(strings.Join(tokensToStrings(typeStrRepr), ""))
		isPrimitive := typeStr == "i8" || typeStr == "i16" || typeStr == "i32" || typeStr == "i64" ||
			typeStr == "u8" || typeStr == "u16" || typeStr == "u32" || typeStr == "u64" ||
			typeStr == "bool" || typeStr == "f32" || typeStr == "f64" || typeStr == "String" ||
			typeStr == "&str" || typeStr == "usize" || typeStr == "isize"
		if !isPrimitive {
			newTokens = append(newTokens, "mut ")
		}
		newTokens = append(newTokens, nameStr)
		newTokens = append(newTokens, ": ")
		newTokens = append(newTokens, typeStr)
		re.gir.tokenSlice, err = RewriteTokensBetween(re.gir.tokenSlice, p1.Index, p4.Index, newTokens)
		if err != nil {
			fmt.Println("Error rewriting file buffer:", err)
			return
		}
	}
}

// GenerateCargoToml creates a Cargo.toml for building the Rust project with SDL2
func (re *RustEmitter) GenerateCargoToml() error {
	if re.LinkRuntime == "" {
		return nil
	}

	cargoPath := filepath.Join(re.OutputDir, "Cargo.toml")
	file, err := os.Create(cargoPath)
	if err != nil {
		return fmt.Errorf("failed to create Cargo.toml: %w", err)
	}
	defer file.Close()

	cargoToml := fmt.Sprintf(`[package]
name = "%s"
version = "0.1.0"
edition = "2021"

[dependencies]
sdl2 = "0.36"
`, re.OutputName)

	_, err = file.WriteString(cargoToml)
	if err != nil {
		return fmt.Errorf("failed to write Cargo.toml: %w", err)
	}

	DebugLogPrintf("Generated Cargo.toml at %s", cargoPath)
	return nil
}

// GenerateGraphicsMod creates the graphics.rs module file by copying from runtime
func (re *RustEmitter) GenerateGraphicsMod() error {
	if re.LinkRuntime == "" {
		return nil
	}

	// Create src directory if needed (Cargo convention)
	srcDir := filepath.Join(re.OutputDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		return fmt.Errorf("failed to create src directory: %w", err)
	}

	// Source path: LinkRuntime points to runtime directory, graphics runtime is in graphics/rust/
	runtimeSrcPath := filepath.Join(re.LinkRuntime, "graphics", "rust", "graphics_runtime.rs")
	graphicsRs, err := os.ReadFile(runtimeSrcPath)
	if err != nil {
		return fmt.Errorf("failed to read graphics runtime from %s: %w", runtimeSrcPath, err)
	}

	// Destination path
	graphicsPath := filepath.Join(srcDir, "graphics.rs")
	if err := os.WriteFile(graphicsPath, graphicsRs, 0644); err != nil {
		return fmt.Errorf("failed to write graphics.rs: %w", err)
	}

	DebugLogPrintf("Copied graphics.rs from %s to %s", runtimeSrcPath, graphicsPath)
	return nil
}

// GenerateBuildRs creates a build.rs file that sets library search paths for SDL2
func (re *RustEmitter) GenerateBuildRs() error {
	if re.LinkRuntime == "" {
		return nil
	}

	buildRsPath := filepath.Join(re.OutputDir, "build.rs")
	file, err := os.Create(buildRsPath)
	if err != nil {
		return fmt.Errorf("failed to create build.rs: %w", err)
	}
	defer file.Close()

	buildRs := `fn main() {
    // Add Homebrew library path for macOS
    #[cfg(target_os = "macos")]
    {
        // Apple Silicon Macs
        println!("cargo:rustc-link-search=/opt/homebrew/lib");
        // Intel Macs
        println!("cargo:rustc-link-search=/usr/local/lib");
    }

    // Add common Linux library paths
    #[cfg(target_os = "linux")]
    {
        println!("cargo:rustc-link-search=/usr/lib");
        println!("cargo:rustc-link-search=/usr/local/lib");
    }
}
`

	_, err = file.WriteString(buildRs)
	if err != nil {
		return fmt.Errorf("failed to write build.rs: %w", err)
	}

	DebugLogPrintf("Generated build.rs at %s", buildRsPath)
	return nil
}

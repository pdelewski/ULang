package compiler

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
	"log"
)

var cppTypesMap = map[string]string{
	"int8":   "std::int8_t",
	"int16":  "std::int16_t",
	"int32":  "std::int32_t",
	"int64":  "std::int64_t",
	"uint8":  "std::uint8_t",
	"uint16": "std::uint16_t",
	"any":    "std::any",
	"string": "std::string",
}

type CPPEmitter struct {
	Output          string
	OutputDir       string
	OutputName      string
	LinkRuntime     string // Path to runtime directory (empty = disabled)
	GraphicsRuntime string // Graphics backend: tigr (default), sdl2, none
	file            *os.File
	Emitter
	pkg               *packages.Package
	insideForPostCond bool
	assignmentToken   string
	forwardDecl       bool
	// Key-value range loop support
	isKeyValueRange       bool
	rangeKeyName          string
	rangeValueName        string
	rangeCollectionExpr   string
	captureRangeExpr      bool
	suppressRangeEmit     bool
	rangeStmtIndent       int
	pendingRangeValueDecl bool // True when we need to emit value decl at start of block
	pendingValueName      string
	pendingCollectionExpr string
	pendingKeyName        string
}

func (*CPPEmitter) lowerToBuiltins(selector string) string {
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

func (cppe *CPPEmitter) emitToFile(s string) error {
	// When capturing range expression, append to buffer instead of file
	if cppe.captureRangeExpr {
		cppe.rangeCollectionExpr += s
		return nil
	}
	// When suppressing range emit (key/value identifiers), skip
	if cppe.suppressRangeEmit {
		return nil
	}
	return emitToFile(cppe.file, s)
}

func (cppe *CPPEmitter) emitAsString(s string, indent int) string {
	return strings.Repeat(" ", indent) + s
}

func (cppe *CPPEmitter) SetFile(file *os.File) {
	cppe.file = file
}

func (cppe *CPPEmitter) GetFile() *os.File {
	return cppe.file
}

func (cppe *CPPEmitter) PreVisitProgram(indent int) {
	outputFile := cppe.Output
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
		"#include <functional>\n")
	cppe.file.WriteString(`#include <cstdarg> // For va_start, etc.
#include <initializer_list>
#include <iostream>
`)
	// Include runtime header if link-runtime is enabled
	if cppe.LinkRuntime != "" {
		// Include graphics runtime based on selected backend
		graphicsBackend := cppe.GraphicsRuntime
		if graphicsBackend == "" {
			graphicsBackend = "tigr"
		}
		switch graphicsBackend {
		case "sdl2":
			cppe.file.WriteString("#include \"graphics/cpp/graphics_runtime_sdl2.hpp\"\n")
		case "tigr":
			cppe.file.WriteString("#include \"graphics/cpp/graphics_runtime_tigr.hpp\"\n")
		case "none":
			// No graphics runtime
		}
	}
	cppe.file.WriteString(`
using int8 = int8_t;
using int16 = int16_t;
using int32 = int32_t;
using int64 = int64_t;
using uint8 = uint8_t;
using uint16 = uint16_t;
using uint32 = uint32_t;
using uint64 = uint64_t;

std::string string_format(const std::string fmt, ...) {
  int size =
      ((int)fmt.size()) * 2 + 50; // Use a rubric appropriate for your code
  std::string str;
  va_list ap;
  while (1) { // Maximum two passes on a POSIX system...
    str.resize(size);
    va_start(ap, fmt);
    int n = vsnprintf((char *)str.data(), size, fmt.c_str(), ap);
    va_end(ap);
    if (n > -1 && n < size) { // Everything worked
      str.resize(n);
      return str;
    }
    if (n > -1)     // Needed size returned
      size = n + 1; // For null char
    else
      size *= 2; // Guess at a larger size (OS specific)
  }
  return str;
}

void println() { printf("\n"); }

void println(std::int8_t val) { printf("%d\n", val); }

template<typename T>
void println(const T& val) { std::cout << val << std::endl;}

void printf(const signed char c) {
  std::cout << (int)c;
}

template<typename T>
void printf(const T& val) { std::cout << val;}

// Function to mimic Go's append behavior for std::vector
template <typename T>
std::vector<T>& append(std::vector<T> &vec,
                      const std::initializer_list<T> &elements) {
  std::vector<T>& result = vec;           // Create a copy of the original vector
  result.insert(result.end(), elements); // Append the elements
  return result;                         // Return the new vector
}

// Overload to allow appending another vector
template <typename T>
std::vector<T>& append( std::vector<T> &vec,
                      const std::vector<T> &elements) {
  std::vector<T> result = vec; // Create a copy of the original vector
  result.insert(result.end(), elements.begin(),
                elements.end()); // Append the elements
  return result;                 // Return the new vector
}

template <typename T>
std::vector<T>& append(std::vector<T> &vec, const T &element) {
  std::vector<T>& result = vec; // Create a copy of the original vector
  result.push_back(element);   // Append the single element
  return result;               // Return the new vector
}

// Specialization for appending const char* to vector of strings
std::vector<std::string>& append(std::vector<std::string> &vec, const char *element) {
  std::vector<std::string>& result = vec;
  result.push_back(std::string(element));
  return result;
}
`)
	cppe.file.WriteString("\n\n")
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
	cppe.insideForPostCond = false
}

func (cppe *CPPEmitter) PostVisitProgram(indent int) {
	cppe.file.Close()

	// Generate Makefile if link-runtime is enabled
	if err := cppe.GenerateMakefile(); err != nil {
		log.Printf("Warning: %v", err)
	}
}

func (cppe *CPPEmitter) PreVisitPackage(pkg *packages.Package, indent int) {
	name := pkg.Name
	DebugLogPrintf("CPPEmitter: PreVisitPackage: %s (path: %s)", name, pkg.PkgPath)
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

func (cppe *CPPEmitter) PostVisitPackage(pkg *packages.Package, indent int) {
	name := pkg.Name
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

func (cppe *CPPEmitter) PreVisitBasicLit(e *ast.BasicLit, indent int) {
	if e.Kind == token.STRING {
		// Use a local copy to avoid mutating the AST (which affects other emitters)
		value := strings.Replace(e.Value, "\"", "", -1)
		if len(value) > 0 && value[0] == '`' {
			value = strings.Replace(value, "`", "", -1)
			cppe.emitToFile(cppe.emitAsString(fmt.Sprintf("R\"(%s)\"", value), 0))
		} else {
			cppe.emitToFile(cppe.emitAsString(fmt.Sprintf("\"%s\"", value), 0))
		}
	} else {
		cppe.emitToFile(cppe.emitAsString(e.Value, 0))
	}
}

func (cppe *CPPEmitter) PreVisitIdent(e *ast.Ident, indent int) {
	var str string
	name := e.Name
	name = cppe.lowerToBuiltins(name)
	if name == "nil" {
		str = cppe.emitAsString("{}", indent)
		cppe.emitToFile(str)
	} else {
		if n, ok := cppTypesMap[name]; ok {
			str = cppe.emitAsString(n, indent)
			cppe.emitToFile(str)
		} else {
			str = cppe.emitAsString(name, indent)
			cppe.emitToFile(str)
		}
	}
}

func (cppe *CPPEmitter) PreVisitBinaryExprOperator(op token.Token, indent int) {
	str := cppe.emitAsString(op.String()+" ", 1)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PreVisitCallExprArgs(node []ast.Expr, indent int) {
	str := cppe.emitAsString("(", 0)
	cppe.emitToFile(str)
}
func (cppe *CPPEmitter) PostVisitCallExprArgs(node []ast.Expr, indent int) {
	str := cppe.emitAsString(")", 0)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PreVisitCallExprArg(node ast.Expr, index int, indent int) {
	if index > 0 {
		str := cppe.emitAsString(", ", 0)
		cppe.emitToFile(str)
	}
}

func (cppe *CPPEmitter) PreVisitParenExpr(node *ast.ParenExpr, indent int) {
	str := cppe.emitAsString("(", 0)
	cppe.emitToFile(str)
}
func (cppe *CPPEmitter) PostVisitParenExpr(node *ast.ParenExpr, indent int) {
	str := cppe.emitAsString(")", 0)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PreVisitCompositeLitElts(node []ast.Expr, indent int) {
	str := cppe.emitAsString("{", 0)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PostVisitCompositeLitElts(node []ast.Expr, indent int) {
	str := cppe.emitAsString("}", 0)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PreVisitCompositeLitElt(node ast.Expr, index int, indent int) {
	if index > 0 {
		str := cppe.emitAsString(", ", 0)
		cppe.emitToFile(str)
	}
}

func (cppe *CPPEmitter) PreVisitArrayType(node ast.ArrayType, indent int) {
	str := cppe.emitAsString("std::vector<", indent)
	cppe.emitToFile(str)
}
func (cppe *CPPEmitter) PostVisitArrayType(node ast.ArrayType, indent int) {
	str := cppe.emitAsString(">", 0)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PostVisitSelectorExprX(node ast.Expr, indent int) {
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

func (cppe *CPPEmitter) PreVisitIndexExprIndex(node *ast.IndexExpr, indent int) {
	str := cppe.emitAsString("[", 0)
	cppe.emitToFile(str)
}
func (cppe *CPPEmitter) PostVisitIndexExprIndex(node *ast.IndexExpr, indent int) {
	str := cppe.emitAsString("]", 0)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PreVisitUnaryExpr(node *ast.UnaryExpr, indent int) {
	str := cppe.emitAsString("(", 0)
	str += cppe.emitAsString(node.Op.String(), 0)
	cppe.emitToFile(str)
}
func (cppe *CPPEmitter) PostVisitUnaryExpr(node *ast.UnaryExpr, indent int) {
	str := cppe.emitAsString(")", 0)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PreVisitSliceExpr(node *ast.SliceExpr, indent int) {
	str := cppe.emitAsString("std::vector<std::remove_reference<decltype(", indent)
	cppe.emitToFile(str)
}
func (cppe *CPPEmitter) PostVisitSliceExpr(node *ast.SliceExpr, indent int) {
	str := cppe.emitAsString(")", 0)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PostVisitSliceExprX(node ast.Expr, indent int) {
	str := cppe.emitAsString("[0]", 0)
	str += cppe.emitAsString(")>::type>(", 0)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PreVisitSliceExprLow(node ast.Expr, indent int) {
	str := cppe.emitAsString(".begin() ", 0)
	cppe.emitToFile(str)
	if node != nil {
		str := cppe.emitAsString("+ ", 0)
		cppe.emitToFile(str)
	} else {
		DebugLogPrintf("Low index: <nil>\n")
	}
}

func (cppe *CPPEmitter) PreVisitSliceExprXEnd(node ast.Expr, indent int) {
	str := cppe.emitAsString(", ", 0)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PreVisitSliceExprHigh(node ast.Expr, indent int) {
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

func (cppe *CPPEmitter) PostVisitSliceExprHigh(node ast.Expr, indent int) {

}

func (cppe *CPPEmitter) PreVisitFuncType(node *ast.FuncType, indent int) {
	str := cppe.emitAsString("std::function<", indent)
	cppe.emitToFile(str)
}
func (cppe *CPPEmitter) PostVisitFuncType(node *ast.FuncType, indent int) {
	str := cppe.emitAsString(">", 0)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PreVisitFuncTypeResults(node *ast.FieldList, indent int) {
	if node == nil {
		str := cppe.emitAsString("void", 0)
		cppe.emitToFile(str)
	}
}

func (cppe *CPPEmitter) PreVisitFuncTypeResult(node *ast.Field, index int, indent int) {
	if index > 0 {
		str := cppe.emitAsString(", ", 0)
		cppe.emitToFile(str)
	}
}

func (cppe *CPPEmitter) PreVisitKeyValueExpr(node *ast.KeyValueExpr, indent int) {
	str := cppe.emitAsString(".", 0)
	cppe.emitToFile(str)
}
func (cppe *CPPEmitter) PostVisitKeyValueExpr(node *ast.KeyValueExpr, indent int) {
	str := cppe.emitAsString("\n", 0)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PreVisitFuncTypeParams(node *ast.FieldList, indent int) {
	str := cppe.emitAsString("(", 0)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PostVisitFuncTypeParams(node *ast.FieldList, indent int) {
	str := cppe.emitAsString(")", 0)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PreVisitFuncTypeParam(node *ast.Field, index int, indent int) {
	if index > 0 {
		str := cppe.emitAsString(", ", 0)
		cppe.emitToFile(str)
	}
}

func (cppe *CPPEmitter) PreVisitKeyValueExprValue(node ast.Expr, indent int) {
	str := cppe.emitAsString("= ", 0)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PreVisitFuncLit(node *ast.FuncLit, indent int) {
	str := cppe.emitAsString("[&](", indent)
	cppe.emitToFile(str)
}
func (cppe *CPPEmitter) PostVisitFuncLit(node *ast.FuncLit, indent int) {
	str := cppe.emitAsString("}", 0)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PostVisitFuncLitTypeParams(node *ast.FieldList, indent int) {
	str := cppe.emitAsString(")", 0)
	str += cppe.emitAsString("->", 0)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PreVisitFuncLitTypeParam(node *ast.Field, index int, indent int) {
	str := ""
	if index > 0 {
		str += cppe.emitAsString(", ", 0)
	}
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PostVisitFuncLitTypeParam(node *ast.Field, index int, indent int) {
	str := cppe.emitAsString(" ", 0)
	str += cppe.emitAsString(node.Names[0].Name, indent)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PreVisitFuncLitBody(node *ast.BlockStmt, indent int) {
	str := cppe.emitAsString("{\n", 0)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PreVisitFuncLitTypeResults(node *ast.FieldList, indent int) {
	if node == nil {
		str := cppe.emitAsString("void", 0)
		cppe.emitToFile(str)
	}
}

func (cppe *CPPEmitter) PreVisitFuncLitTypeResult(node *ast.Field, index int, indent int) {
	if index > 0 {
		str := cppe.emitAsString(", ", 0)
		cppe.emitToFile(str)
	}
}

func (cppe *CPPEmitter) PreVisitTypeAssertExprType(node ast.Expr, indent int) {
	str := cppe.emitAsString("std::any_cast<", indent)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PreVisitTypeAssertExprX(node ast.Expr, indent int) {
	str := cppe.emitAsString(">(std::any(", 0)
	cppe.emitToFile(str)
}
func (cppe *CPPEmitter) PostVisitTypeAssertExprX(node ast.Expr, indent int) {
	str := cppe.emitAsString("))", 0)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PreVisitStarExpr(node *ast.StarExpr, indent int) {
	str := cppe.emitAsString("*", 0)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PreVisitInterfaceType(node *ast.InterfaceType, indent int) {
	str := cppe.emitAsString("std::any", indent)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PostVisitExprStmtX(node ast.Expr, indent int) {
	str := cppe.emitAsString(";", 0)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PreVisitDeclStmtValueSpecNames(node *ast.Ident, index int, indent int) {
	str := cppe.emitAsString(" ", 0)
	cppe.emitToFile(str)
}
func (cppe *CPPEmitter) PostVisitDeclStmtValueSpecNames(node *ast.Ident, index int, indent int) {
	cppe.emitToFile(";")
}

func (cppe *CPPEmitter) PreVisitBranchStmt(node *ast.BranchStmt, indent int) {
	str := cppe.emitAsString(node.Tok.String()+";", indent)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PostVisitIncDecStmt(node *ast.IncDecStmt, indent int) {
	str := cppe.emitAsString(node.Tok.String(), 0)
	if !cppe.insideForPostCond {
		str += cppe.emitAsString(";", 0)
	}
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PreVisitForStmtPost(node ast.Stmt, indent int) {
	if node != nil {
		cppe.insideForPostCond = true
	}
}
func (cppe *CPPEmitter) PostVisitForStmtPost(node ast.Stmt, indent int) {
	cppe.insideForPostCond = false
	str := cppe.emitAsString(")\n", 0)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PreVisitAssignStmt(node *ast.AssignStmt, indent int) {
	str := cppe.emitAsString("", indent)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PostVisitAssignStmt(node *ast.AssignStmt, indent int) {
	// Don't emit semicolon inside for loop post statement
	if !cppe.insideForPostCond {
		str := cppe.emitAsString(";", 0)
		cppe.emitToFile(str)
	}
}

func (cppe *CPPEmitter) PreVisitAssignStmtLhs(node *ast.AssignStmt, indent int) {
	assignmentToken := node.Tok.String()
	if assignmentToken == ":=" && len(node.Lhs) == 1 {
		// Check if RHS is a string literal - use std::string instead of auto
		// to avoid const char* which breaks string concatenation
		if len(node.Rhs) == 1 {
			if lit, ok := node.Rhs[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
				str := cppe.emitAsString("std::string ", indent)
				cppe.emitToFile(str)
			} else {
				str := cppe.emitAsString("auto ", indent)
				cppe.emitToFile(str)
			}
		} else {
			str := cppe.emitAsString("auto ", indent)
			cppe.emitToFile(str)
		}
	} else if assignmentToken == ":=" && len(node.Lhs) > 1 {
		str := cppe.emitAsString("auto [", indent)
		cppe.emitToFile(str)
	} else if assignmentToken == "=" && len(node.Lhs) > 1 {
		str := cppe.emitAsString("std::tie(", indent)
		cppe.emitToFile(str)
	}
	// Preserve compound assignment operators, convert := to =
	if assignmentToken != "+=" && assignmentToken != "-=" && assignmentToken != "*=" && assignmentToken != "/=" {
		assignmentToken = "="
	}
	cppe.assignmentToken = assignmentToken
}

func (cppe *CPPEmitter) PostVisitAssignStmtLhs(node *ast.AssignStmt, indent int) {
	if node.Tok.String() == ":=" && len(node.Lhs) > 1 {
		str := cppe.emitAsString("]", indent)
		cppe.emitToFile(str)
	} else if node.Tok.String() == "=" && len(node.Lhs) > 1 {
		str := cppe.emitAsString(")", indent)
		cppe.emitToFile(str)
	}
}

func (cppe *CPPEmitter) PreVisitAssignStmtRhs(node *ast.AssignStmt, indent int) {
	str := cppe.emitAsString(cppe.assignmentToken+" ", indent+1)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PreVisitAssignStmtLhsExpr(node ast.Expr, index int, indent int) {
	if index > 0 {
		str := cppe.emitAsString(", ", indent)
		cppe.emitToFile(str)
	}
}

func (cppe *CPPEmitter) PreVisitReturnStmt(node *ast.ReturnStmt, indent int) {
	str := cppe.emitAsString("return ", indent)
	cppe.emitToFile(str)
	if len(node.Results) > 1 {
		str := cppe.emitAsString("std::make_tuple(", 0)
		cppe.emitToFile(str)
	}
}

func (cppe *CPPEmitter) PostVisitReturnStmt(node *ast.ReturnStmt, indent int) {
	if len(node.Results) > 1 {
		str := cppe.emitAsString(")", 0)
		cppe.emitToFile(str)
	}
	str := cppe.emitAsString(";", 0)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PreVisitReturnStmtResult(node ast.Expr, index int, indent int) {
	if index > 0 {
		str := cppe.emitAsString(", ", 0)
		cppe.emitToFile(str)
	}
}

func (cppe *CPPEmitter) PreVisitIfStmtCond(node *ast.IfStmt, indent int) {
	str := cppe.emitAsString("if (", indent)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PostVisitIfStmtCond(node *ast.IfStmt, indent int) {
	str := cppe.emitAsString(")\n", 0)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PreVisitIfStmtElse(node *ast.IfStmt, indent int) {
	str := cppe.emitAsString("else", 1)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PreVisitForStmt(node *ast.ForStmt, indent int) {
	str := cppe.emitAsString("for (", indent)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PostVisitForStmtInit(node ast.Stmt, indent int) {
	if node == nil {
		str := cppe.emitAsString(";", 0)
		cppe.emitToFile(str)
	}
}

func (cppe *CPPEmitter) PostVisitForStmtCond(node ast.Expr, indent int) {
	str := cppe.emitAsString(";", 0)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PreVisitRangeStmt(node *ast.RangeStmt, indent int) {
	// Check if this is a key-value range (both Key and Value present)
	if node.Key != nil && node.Value != nil {
		cppe.isKeyValueRange = true
		cppe.rangeKeyName = node.Key.(*ast.Ident).Name
		cppe.rangeValueName = node.Value.(*ast.Ident).Name
		cppe.rangeCollectionExpr = ""
		cppe.suppressRangeEmit = true
		cppe.rangeStmtIndent = indent
		// Don't emit anything yet - we'll emit the complete for loop in PostVisitRangeStmtX
	} else {
		cppe.isKeyValueRange = false
		str := cppe.emitAsString("for (auto ", indent)
		cppe.emitToFile(str)
	}
}

func (cppe *CPPEmitter) PreVisitRangeStmtKey(node ast.Expr, indent int) {
	// For key-value range, we've already captured the key name
	// For index-only range (for i := range), let it emit normally
}

func (cppe *CPPEmitter) PostVisitRangeStmtKey(node ast.Expr, indent int) {
	// Nothing special needed here
}

func (cppe *CPPEmitter) PreVisitRangeStmtValue(node ast.Expr, indent int) {
	// For key-value range, we've already captured the value name, keep suppressing
}

func (cppe *CPPEmitter) PostVisitRangeStmtValue(node ast.Expr, indent int) {
	if cppe.isKeyValueRange {
		// Stop suppressing, start capturing collection expression
		cppe.suppressRangeEmit = false
		cppe.captureRangeExpr = true
	} else {
		str := cppe.emitAsString(" : ", 0)
		cppe.emitToFile(str)
	}
}

func (cppe *CPPEmitter) PreVisitRangeStmtX(node ast.Expr, indent int) {
	// For key-value range, we're already in capture mode
}

func (cppe *CPPEmitter) PostVisitRangeStmtX(node ast.Expr, indent int) {
	if cppe.isKeyValueRange {
		// Stop capturing and emit the complete for loop
		cppe.captureRangeExpr = false
		collection := cppe.rangeCollectionExpr
		key := cppe.rangeKeyName
		value := cppe.rangeValueName
		indent := cppe.rangeStmtIndent

		// Emit: for (size_t key = 0; key < collection.size(); key++)
		str := cppe.emitAsString(fmt.Sprintf("for (size_t %s = 0; %s < %s.size(); %s++)\n", key, key, collection, key), indent)
		cppe.emitToFile(str)

		// Set pending value declaration to be emitted at start of body block
		cppe.pendingRangeValueDecl = true
		cppe.pendingValueName = value
		cppe.pendingCollectionExpr = collection
		cppe.pendingKeyName = key

		// Reset range state
		cppe.isKeyValueRange = false
		cppe.rangeKeyName = ""
		cppe.rangeValueName = ""
		cppe.rangeCollectionExpr = ""
	} else {
		str := cppe.emitAsString(")\n", 0)
		cppe.emitToFile(str)
	}
}

func (cppe *CPPEmitter) PreVisitRangeStmtBody(node *ast.BlockStmt, indent int) {
	// For key-value range, emit the value declaration at the start of the body
	// This is called from the base_pass when visiting the body
}

func (cppe *CPPEmitter) PostVisitRangeStmt(node *ast.RangeStmt, indent int) {
	// Reset any range-related state
}
func (cppe *CPPEmitter) PreVisitSwitchStmt(node *ast.SwitchStmt, indent int) {
	str := cppe.emitAsString("switch (", indent)
	cppe.emitToFile(str)
}
func (cppe *CPPEmitter) PostVisitSwitchStmt(node *ast.SwitchStmt, indent int) {
	str := cppe.emitAsString("}", indent)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PostVisitSwitchStmtTag(node ast.Expr, indent int) {
	str := cppe.emitAsString(") {\n", 0)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PostVisitCaseClause(node *ast.CaseClause, indent int) {
	cppe.emitToFile("\n")
	str := cppe.emitAsString("break;\n", indent+4)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PostVisitCaseClauseList(node []ast.Expr, indent int) {
	if len(node) == 0 {
		str := cppe.emitAsString("default:\n", indent+2)
		cppe.emitToFile(str)
	}
}

func (cppe *CPPEmitter) PreVisitCaseClauseListExpr(node ast.Expr, index int, indent int) {
	str := cppe.emitAsString("case ", indent+2)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PostVisitCaseClauseListExpr(node ast.Expr, index int, indent int) {
	str := cppe.emitAsString(":\n", 0)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PreVisitBlockStmt(node *ast.BlockStmt, indent int) {
	str := cppe.emitAsString("{\n", indent)
	cppe.emitToFile(str)

	// If we have a pending value declaration from key-value range, emit it now
	if cppe.pendingRangeValueDecl {
		valueDecl := cppe.emitAsString(fmt.Sprintf("auto %s = %s[%s];\n",
			cppe.pendingValueName, cppe.pendingCollectionExpr, cppe.pendingKeyName), indent+2)
		cppe.emitToFile(valueDecl)
		cppe.pendingRangeValueDecl = false
		cppe.pendingValueName = ""
		cppe.pendingCollectionExpr = ""
		cppe.pendingKeyName = ""
	}
}

func (cppe *CPPEmitter) PostVisitBlockStmt(node *ast.BlockStmt, indent int) {
	str := cppe.emitAsString("}", indent)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PostVisitBlockStmtList(node ast.Stmt, index int, indent int) {
	str := cppe.emitAsString("\n", indent)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PostVisitFuncDecl(node *ast.FuncDecl, indent int) {
	str := cppe.emitAsString("\n\n", 0)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PreVisitFuncDeclBody(node *ast.BlockStmt, indent int) {
	str := cppe.emitAsString("\n", 0)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PreVisitFuncDeclSignatureTypeResults(node *ast.FuncDecl, indent int) {
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

func (cppe *CPPEmitter) PostVisitFuncDeclSignatureTypeParams(node *ast.FuncDecl, indent int) {
	str := cppe.emitAsString(")", 0)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PreVisitFuncDeclName(node *ast.Ident, indent int) {
	DebugLogPrintf("CPPEmitter: PreVisitFuncDeclName: %s", node.Name)
	str := cppe.emitAsString(node.Name, 0)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PostVisitFuncDeclSignatureTypeResults(node *ast.FuncDecl, indent int) {
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
}

func (cppe *CPPEmitter) PreVisitFuncDeclSignatureTypeResultsList(node *ast.Field, index int, indent int) {
	if index > 0 {
		str := cppe.emitAsString(",", 0)
		cppe.emitToFile(str)
	}
}

func (cppe *CPPEmitter) PreVisitFuncDeclSignatureTypeParams(node *ast.FuncDecl, indent int) {
	str := cppe.emitAsString("(", 0)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PreVisitFuncDeclSignatureTypeParamsList(node *ast.Field, index int, indent int) {
	if index > 0 {
		str := cppe.emitAsString(", ", 0)
		cppe.emitToFile(str)
	}
}

func (cppe *CPPEmitter) PreVisitFuncDeclSignatureTypeParamsArgName(node *ast.Ident, index int, indent int) {
	cppe.emitToFile(" ")
}

func (cppe *CPPEmitter) PostVisitFuncDeclSignature(node *ast.FuncDecl, indent int) {
	if cppe.forwardDecl {
		str := cppe.emitAsString(";\n", 0)
		cppe.emitToFile(str)
	}
}

func (cppe *CPPEmitter) PreVisitGenStructInfo(node GenTypeInfo, indent int) {
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
func (cppe *CPPEmitter) PostVisitGenStructInfo(node GenTypeInfo, indent int) {
	str := cppe.emitAsString("};\n\n", 0)
	err := cppe.emitToFile(str)
	if err != nil {
		fmt.Println("Error writing to file:", err)
	}
}

func (cppe *CPPEmitter) PreVisitFuncDeclSignatures(indent int) {
	// Generate forward function declarations
	str := cppe.emitAsString("// Forward declarations\n", 0)
	cppe.emitToFile(str)
	cppe.forwardDecl = true
}

func (cppe *CPPEmitter) PostVisitFuncDeclSignatures(indent int) {
	str := cppe.emitAsString("\n", 0)
	cppe.emitToFile(str)
	cppe.forwardDecl = false
}

func (cppe *CPPEmitter) PostVisitGenDeclConst(node *ast.GenDecl, indent int) {
	str := cppe.emitAsString("\n", 0)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PostVisitGenStructFieldType(node ast.Expr, indent int) {
	cppe.emitToFile(" ")
}

func (cppe *CPPEmitter) PostVisitGenStructFieldName(node *ast.Ident, indent int) {
	cppe.emitToFile(";\n")
}

func (cppe *CPPEmitter) PreVisitGenDeclConstName(node *ast.Ident, indent int) {
	str := cppe.emitAsString(fmt.Sprintf("constexpr auto %s = ", node.Name), 0)
	cppe.emitToFile(str)
}
func (cppe *CPPEmitter) PostVisitGenDeclConstName(node *ast.Ident, indent int) {
	str := cppe.emitAsString(";\n", 0)
	cppe.emitToFile(str)
}

func (cppe *CPPEmitter) PreVisitTypeAliasName(node *ast.Ident, indent int) {
	cppe.emitToFile(fmt.Sprintf("using "))
}

func (cppe *CPPEmitter) PostVisitTypeAliasName(node *ast.Ident, indent int) {
	cppe.emitToFile(" = ")
}

func (cppe *CPPEmitter) PostVisitTypeAliasType(node ast.Expr, indent int) {
	cppe.emitToFile(";\n\n")
}

// GenerateMakefile creates a Makefile for building the C++ project
func (cppe *CPPEmitter) GenerateMakefile() error {
	if cppe.LinkRuntime == "" {
		return nil
	}

	makefilePath := filepath.Join(cppe.OutputDir, "Makefile")
	file, err := os.Create(makefilePath)
	if err != nil {
		return fmt.Errorf("failed to create Makefile: %w", err)
	}
	defer file.Close()

	// Convert LinkRuntime to absolute path so Makefile works from any directory
	absRuntimePath, err := filepath.Abs(cppe.LinkRuntime)
	if err != nil {
		absRuntimePath = cppe.LinkRuntime // Fall back to original if Abs fails
	}

	var makefile string

	// Determine graphics backend (default to tigr)
	graphicsBackend := cppe.GraphicsRuntime
	if graphicsBackend == "" {
		graphicsBackend = "tigr"
	}

	switch graphicsBackend {
	case "sdl2":
		makefile = fmt.Sprintf(`CXX = g++
CXXFLAGS = -O3 -std=c++17 -I%s $(shell sdl2-config --cflags)
LDFLAGS = $(shell sdl2-config --libs)

TARGET = %s
SRCS = %s.cpp

all: $(TARGET)

$(TARGET): $(SRCS)
	$(CXX) $(CXXFLAGS) -o $@ $^ $(LDFLAGS)

clean:
	rm -f $(TARGET)

.PHONY: all clean
`, absRuntimePath, cppe.OutputName, cppe.OutputName)

	case "none":
		makefile = fmt.Sprintf(`CXX = g++
CXXFLAGS = -O3 -std=c++17 -I%s
LDFLAGS =

TARGET = %s
SRCS = %s.cpp

all: $(TARGET)

$(TARGET): $(SRCS)
	$(CXX) $(CXXFLAGS) -o $@ $^ $(LDFLAGS)

clean:
	rm -f $(TARGET)

.PHONY: all clean
`, absRuntimePath, cppe.OutputName, cppe.OutputName)

	default: // tigr (default)
		makefile = fmt.Sprintf(`CXX = g++
CC = gcc
CXXFLAGS = -O3 -std=c++17 -I%s
CFLAGS = -O3

TARGET = %s
SRCS = %s.cpp
TIGR_SRC = %s/graphics/cpp/tigr.c

# Platform-specific flags
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
    LDFLAGS = -framework OpenGL -framework Cocoa
endif
ifeq ($(UNAME_S),Linux)
    LDFLAGS = -lGL -lX11
endif
# Windows (MinGW)
ifeq ($(OS),Windows_NT)
    LDFLAGS = -lopengl32 -lgdi32
endif

all: $(TARGET)

tigr.o: $(TIGR_SRC)
	$(CC) $(CFLAGS) -c $< -o $@

$(TARGET): $(SRCS) tigr.o
	$(CXX) $(CXXFLAGS) -o $@ $(SRCS) tigr.o $(LDFLAGS)

clean:
	rm -f $(TARGET) tigr.o

.PHONY: all clean
`, absRuntimePath, cppe.OutputName, cppe.OutputName, absRuntimePath)
	}

	_, err = file.WriteString(makefile)
	if err != nil {
		return fmt.Errorf("failed to write Makefile: %w", err)
	}

	DebugLogPrintf("Generated Makefile at %s (graphics: %s)", makefilePath, graphicsBackend)
	return nil
}

package compiler

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

var jsTypesMap = map[string]string{
	"int8":   "number",
	"int16":  "number",
	"int32":  "number",
	"int64":  "number",
	"int":    "number",
	"uint8":  "number",
	"uint16": "number",
	"uint32": "number",
	"uint64": "number",
	"float32": "number",
	"float64": "number",
	"bool":   "boolean",
	"string": "string",
	"any":    "any",
}

type JSEmitter struct {
	Output          string
	OutputDir       string
	OutputName      string
	LinkRuntime     string // Path to runtime directory (empty = disabled)
	GraphicsRuntime string // Graphics backend for browser
	file            *os.File
	Emitter
	pkg                   *packages.Package
	insideForPostCond     bool
	assignmentToken       string
	forwardDecl           bool
	currentPackage        string
	// Key-value range loop support
	isKeyValueRange       bool
	rangeKeyName          string
	rangeValueName        string
	rangeCollectionExpr   string
	captureRangeExpr      bool
	suppressRangeEmit     bool
	rangeStmtIndent       int
	pendingRangeValueDecl bool  // Emit value declaration in next block
	// For loop support
	sawIncrement          bool
	sawDecrement          bool
	forLoopInclusive      bool
	forLoopReverse        bool
	isInfiniteLoop        bool
	// Multi-value support
	inMultiValueReturn    bool
	multiValueReturnIndex int
	numFuncResults        int
	// Type suppression for JavaScript (no type annotations)
	suppressTypeEmit      bool
	// For loop init section (suppress semicolon after assignment)
	insideForInit         bool
	// Pending slice/struct initialization
	pendingSliceInit      bool
	pendingStructInit     bool
}

func (*JSEmitter) lowerToBuiltins(selector string) string {
	switch selector {
	case "fmt":
		return ""
	case "types":
		// Local package reference - constants are defined globally
		return ""
	case "Sprintf":
		return "stringFormat"
	case "Println":
		return "console.log"
	case "Printf":
		return "console.log"
	case "Print":
		return "console.log"
	case "len":
		return "len"
	}
	return selector
}

func (jse *JSEmitter) emitToFile(s string) error {
	if jse.captureRangeExpr {
		jse.rangeCollectionExpr += s
		return nil
	}
	if jse.suppressRangeEmit {
		return nil
	}
	return emitToFile(jse.file, s)
}

func (jse *JSEmitter) emitAsString(s string, indent int) string {
	return strings.Repeat("  ", indent) + s
}

func (jse *JSEmitter) SetFile(file *os.File) {
	jse.file = file
}

func (jse *JSEmitter) GetFile() *os.File {
	return jse.file
}

func (jse *JSEmitter) PreVisitProgram(indent int) {
	outputFile := jse.Output
	var err error
	jse.file, err = os.Create(outputFile)
	jse.SetFile(jse.file)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}

	// Write JavaScript header with runtime helpers
	jse.file.WriteString(`// Generated JavaScript code
"use strict";

// Runtime helpers
function len(arr) {
  if (typeof arr === 'string') return arr.length;
  if (Array.isArray(arr)) return arr.length;
  return 0;
}

function append(arr, ...items) {
  return [...arr, ...items];
}

function stringFormat(fmt, ...args) {
  let i = 0;
  return fmt.replace(/%[sdvfx%]/g, (match) => {
    if (match === '%%') return '%';
    if (i >= args.length) return match;
    const arg = args[i++];
    switch (match) {
      case '%s': return String(arg);
      case '%d': return parseInt(arg, 10);
      case '%f': return parseFloat(arg);
      case '%v': return String(arg);
      case '%x': return parseInt(arg, 10).toString(16);
      default: return arg;
    }
  });
}

function make(type, length, capacity) {
  if (Array.isArray(type)) {
    return new Array(length || 0).fill(type[0] === 'number' ? 0 : null);
  }
  return [];
}

// Type conversion functions (no-op in JavaScript - all numbers are float64)
function int8(v) { return v | 0; }
function int16(v) { return v | 0; }
function int32(v) { return v | 0; }
function int64(v) { return v; }  // BigInt not used for simplicity
function int(v) { return v | 0; }
function uint8(v) { return (v | 0) & 0xFF; }
function uint16(v) { return (v | 0) & 0xFFFF; }
function uint32(v) { return (v | 0) >>> 0; }
function uint64(v) { return v; }  // BigInt not used for simplicity
function float32(v) { return v; }
function float64(v) { return v; }
function string(v) { return String(v); }
function bool(v) { return Boolean(v); }

`)

	// Include graphics runtime if enabled
	if jse.LinkRuntime != "" {
		jse.file.WriteString(`// Graphics runtime for Canvas
const graphics = {
  canvas: null,
  ctx: null,
  running: true,
  keys: {},
  mouseX: 0,
  mouseY: 0,
  mouseDown: false,

  CreateWindow: function(title, width, height) {
    this.canvas = document.createElement('canvas');
    this.canvas.width = width;
    this.canvas.height = height;
    this.ctx = this.canvas.getContext('2d');
    document.body.appendChild(this.canvas);
    document.title = title;

    // Event listeners
    window.addEventListener('keydown', (e) => { this.keys[e.key] = true; });
    window.addEventListener('keyup', (e) => { this.keys[e.key] = false; });
    this.canvas.addEventListener('mousemove', (e) => {
      const rect = this.canvas.getBoundingClientRect();
      this.mouseX = e.clientX - rect.left;
      this.mouseY = e.clientY - rect.top;
    });
    this.canvas.addEventListener('mousedown', () => { this.mouseDown = true; });
    this.canvas.addEventListener('mouseup', () => { this.mouseDown = false; });

    return this.canvas;
  },

  NewColor: function(r, g, b, a) {
    return { r, g, b, a: a !== undefined ? a : 255 };
  },

  Clear: function(canvas, color) {
    this.ctx.fillStyle = ` + "`rgba(${color.r}, ${color.g}, ${color.b}, ${color.a / 255})`" + `;
    this.ctx.fillRect(0, 0, canvas.width, canvas.height);
  },

  FillRect: function(canvas, rect, color) {
    this.ctx.fillStyle = ` + "`rgba(${color.r}, ${color.g}, ${color.b}, ${color.a / 255})`" + `;
    this.ctx.fillRect(rect.x, rect.y, rect.width, rect.height);
  },

  NewRect: function(x, y, width, height) {
    return { x, y, width, height };
  },

  DrawLine: function(canvas, x1, y1, x2, y2, color) {
    this.ctx.strokeStyle = ` + "`rgba(${color.r}, ${color.g}, ${color.b}, ${color.a / 255})`" + `;
    this.ctx.beginPath();
    this.ctx.moveTo(x1, y1);
    this.ctx.lineTo(x2, y2);
    this.ctx.stroke();
  },

  SetPixel: function(canvas, x, y, color) {
    this.ctx.fillStyle = ` + "`rgba(${color.r}, ${color.g}, ${color.b}, ${color.a / 255})`" + `;
    this.ctx.fillRect(x, y, 1, 1);
  },

  PollEvents: function(canvas) {
    return [canvas, this.running];
  },

  Update: function(canvas) {
    // Canvas updates automatically
  },

  KeyDown: function(canvas, key) {
    return this.keys[key] || false;
  },

  GetMousePos: function(canvas) {
    return [this.mouseX, this.mouseY];
  },

  MouseDown: function(canvas) {
    return this.mouseDown;
  },

  Closed: function(canvas) {
    return !this.running;
  },

  Free: function(canvas) {
    if (canvas && canvas.parentNode) {
      canvas.parentNode.removeChild(canvas);
    }
  }
};

`)
	}
}

func (jse *JSEmitter) PostVisitProgram(indent int) {
	// Add main() call at the end
	jse.file.WriteString("\n// Run main\nmain();\n")
	jse.file.Close()

	// Create HTML wrapper if graphics runtime is enabled
	if jse.LinkRuntime != "" {
		jse.createHTMLWrapper()
	}
}

func (jse *JSEmitter) createHTMLWrapper() {
	htmlFile := strings.TrimSuffix(jse.Output, ".js") + ".html"
	f, err := os.Create(htmlFile)
	if err != nil {
		fmt.Println("Error creating HTML file:", err)
		return
	}
	defer f.Close()

	jsFileName := filepath.Base(jse.Output)
	f.WriteString(fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>%s</title>
  <style>
    body { margin: 0; display: flex; justify-content: center; align-items: center; min-height: 100vh; background: #1a1a1a; }
    canvas { border: 1px solid #333; }
  </style>
</head>
<body>
  <script src="%s"></script>
</body>
</html>
`, jse.OutputName, jsFileName))
}

func (jse *JSEmitter) PreVisitPackage(pkg *packages.Package, indent int) {
	jse.pkg = pkg
	jse.currentPackage = pkg.Name
}

// PreVisitFuncDecl handles function declarations
func (jse *JSEmitter) PreVisitFuncDecl(node *ast.FuncDecl, indent int) {
	if jse.forwardDecl {
		return
	}
	str := jse.emitAsString("\nfunction ", indent)
	jse.emitToFile(str)
}

func (jse *JSEmitter) PreVisitFuncDeclSignature(node *ast.FuncDecl, indent int) {
	if jse.forwardDecl {
		return
	}
	// Count results for multi-value returns
	if node.Type.Results != nil {
		jse.numFuncResults = len(node.Type.Results.List)
	} else {
		jse.numFuncResults = 0
	}
}

// Suppress return type emission (JavaScript doesn't have return types)
func (jse *JSEmitter) PreVisitFuncDeclSignatureTypeResults(node *ast.FuncDecl, indent int) {
	jse.suppressTypeEmit = true
}

func (jse *JSEmitter) PostVisitFuncDeclSignatureTypeResults(node *ast.FuncDecl, indent int) {
	jse.suppressTypeEmit = false
}

func (jse *JSEmitter) PreVisitFuncDeclName(node *ast.Ident, indent int) {
	if jse.forwardDecl {
		return
	}
	jse.emitToFile(node.Name)
}

func (jse *JSEmitter) PreVisitFuncDeclSignatureTypeParams(node *ast.FuncDecl, indent int) {
	if jse.forwardDecl {
		return
	}
	jse.emitToFile("(")
}

func (jse *JSEmitter) PreVisitFuncDeclSignatureTypeParamsList(node *ast.Field, index int, indent int) {
	if jse.forwardDecl {
		return
	}
	if index > 0 {
		jse.emitToFile(", ")
	}
}

// Suppress parameter type emission (JavaScript doesn't have type annotations)
func (jse *JSEmitter) PreVisitFuncDeclSignatureTypeParamsListType(node ast.Expr, argName *ast.Ident, index int, indent int) {
	jse.suppressTypeEmit = true
}

func (jse *JSEmitter) PostVisitFuncDeclSignatureTypeParamsListType(node ast.Expr, argName *ast.Ident, index int, indent int) {
	jse.suppressTypeEmit = false
}

func (jse *JSEmitter) PreVisitFuncDeclSignatureTypeParamsArgName(node *ast.Ident, index int, indent int) {
	// Arg name is emitted via PreVisitIdent when traversed
}

func (jse *JSEmitter) PostVisitFuncDeclSignatureTypeParams(node *ast.FuncDecl, indent int) {
	if jse.forwardDecl {
		return
	}
	jse.emitToFile(") ")
}

func (jse *JSEmitter) PostVisitFuncDeclSignature(node *ast.FuncDecl, indent int) {
	// Nothing to do - closing paren is in PostVisitFuncDeclSignatureTypeParams
}

func (jse *JSEmitter) PostVisitFuncDecl(node *ast.FuncDecl, indent int) {
	if jse.forwardDecl {
		return
	}
	jse.emitToFile("\n")
}

// JavaScript doesn't need forward declarations (function hoisting handles this)
func (jse *JSEmitter) PreVisitFuncDeclSignatures(indent int) {
	jse.forwardDecl = true
}

func (jse *JSEmitter) PostVisitFuncDeclSignatures(indent int) {
	jse.forwardDecl = false
}

// Block statements
func (jse *JSEmitter) PreVisitBlockStmt(node *ast.BlockStmt, indent int) {
	if jse.forwardDecl {
		return
	}
	jse.emitToFile("{\n")
	// Emit pending range value declaration
	if jse.pendingRangeValueDecl {
		str := jse.emitAsString("let "+jse.rangeValueName+" = "+jse.rangeCollectionExpr+"["+jse.rangeKeyName+"];\n", indent+1)
		jse.emitToFile(str)
		jse.pendingRangeValueDecl = false
	}
}

func (jse *JSEmitter) PostVisitBlockStmt(node *ast.BlockStmt, indent int) {
	if jse.forwardDecl {
		return
	}
	str := jse.emitAsString("}\n", indent)
	jse.emitToFile(str)
}

// Assignment statements
func (jse *JSEmitter) PreVisitAssignStmt(node *ast.AssignStmt, indent int) {
	if jse.forwardDecl {
		return
	}
	// Check if all LHS are blank identifiers - if so, suppress the statement
	allBlank := true
	for _, lhs := range node.Lhs {
		if ident, ok := lhs.(*ast.Ident); ok {
			if ident.Name != "_" {
				allBlank = false
				break
			}
		} else {
			allBlank = false
			break
		}
	}
	if allBlank {
		jse.suppressRangeEmit = true // Reuse this flag to suppress entire statement
	}
}

func (jse *JSEmitter) PreVisitAssignStmtLhs(node *ast.AssignStmt, indent int) {
	if jse.forwardDecl {
		return
	}
	assignmentToken := node.Tok.String()
	if assignmentToken == ":=" && len(node.Lhs) == 1 {
		str := jse.emitAsString("let ", indent)
		jse.emitToFile(str)
	} else if assignmentToken == ":=" && len(node.Lhs) > 1 {
		str := jse.emitAsString("let [", indent)
		jse.emitToFile(str)
	} else {
		str := jse.emitAsString("", indent)
		jse.emitToFile(str)
	}
	// Convert := to =, preserve compound operators
	if assignmentToken != "+=" && assignmentToken != "-=" && assignmentToken != "*=" && assignmentToken != "/=" {
		assignmentToken = "="
	}
	jse.assignmentToken = assignmentToken
}

func (jse *JSEmitter) PostVisitAssignStmtLhs(node *ast.AssignStmt, indent int) {
	if jse.forwardDecl {
		return
	}
	if node.Tok.String() == ":=" && len(node.Lhs) > 1 {
		jse.emitToFile("]")
	}
}

func (jse *JSEmitter) PreVisitAssignStmtRhs(node *ast.AssignStmt, indent int) {
	if jse.forwardDecl {
		return
	}
	jse.emitToFile(" " + jse.assignmentToken + " ")
}

func (jse *JSEmitter) PostVisitAssignStmt(node *ast.AssignStmt, indent int) {
	// Check if this was a blank identifier assignment - reset suppression
	allBlank := true
	for _, lhs := range node.Lhs {
		if ident, ok := lhs.(*ast.Ident); ok {
			if ident.Name != "_" {
				allBlank = false
				break
			}
		} else {
			allBlank = false
			break
		}
	}
	if allBlank {
		jse.suppressRangeEmit = false
		return // Don't emit anything for blank identifier assignments
	}
	if jse.forwardDecl {
		return
	}
	// Don't emit semicolon inside for loop init or post conditions
	if !jse.insideForPostCond && !jse.insideForInit {
		jse.emitToFile(";\n")
	}
}

func (jse *JSEmitter) PreVisitAssignStmtLhsExpr(node ast.Expr, index int, indent int) {
	if jse.forwardDecl {
		return
	}
	if index > 0 {
		jse.emitToFile(", ")
	}
}

// Expression statements
func (jse *JSEmitter) PostVisitExprStmtX(node ast.Expr, indent int) {
	if jse.forwardDecl {
		return
	}
	jse.emitToFile(";\n")
}

// Identifiers
func (jse *JSEmitter) PreVisitIdent(node *ast.Ident, indent int) {
	if jse.forwardDecl {
		return
	}
	// Skip type emissions in JavaScript
	if jse.suppressTypeEmit {
		return
	}
	name := node.Name
	// Apply builtin lowering
	lowered := jse.lowerToBuiltins(name)
	// If lowered to empty string, don't emit (e.g., "fmt" package)
	if lowered == "" {
		return
	}
	// Handle special identifiers
	switch lowered {
	case "true", "false":
		jse.emitToFile(lowered)
	case "nil":
		jse.emitToFile("null")
	default:
		jse.emitToFile(lowered)
	}
}

// Basic literals
func (jse *JSEmitter) PreVisitBasicLit(node *ast.BasicLit, indent int) {
	if jse.forwardDecl {
		return
	}
	switch node.Kind {
	case token.STRING:
		// Handle raw strings
		if strings.HasPrefix(node.Value, "`") {
			// Convert to template literal
			jse.emitToFile(node.Value)
		} else {
			jse.emitToFile(node.Value)
		}
	case token.CHAR:
		// Convert char to string
		jse.emitToFile(node.Value)
	default:
		jse.emitToFile(node.Value)
	}
}

// Binary expressions
func (jse *JSEmitter) PreVisitBinaryExpr(node *ast.BinaryExpr, indent int) {
	if jse.forwardDecl {
		return
	}
	jse.emitToFile("(")
}

func (jse *JSEmitter) PreVisitBinaryExprOperator(op token.Token, indent int) {
	if jse.forwardDecl {
		return
	}
	opStr := op.String()
	// Handle Go operators that need conversion
	switch opStr {
	case "&&":
		jse.emitToFile(" && ")
	case "||":
		jse.emitToFile(" || ")
	default:
		jse.emitToFile(" " + opStr + " ")
	}
}

func (jse *JSEmitter) PostVisitBinaryExpr(node *ast.BinaryExpr, indent int) {
	if jse.forwardDecl {
		return
	}
	jse.emitToFile(")")
}

// Unary expressions
func (jse *JSEmitter) PreVisitUnaryExpr(node *ast.UnaryExpr, indent int) {
	if jse.forwardDecl {
		return
	}
	jse.emitToFile(node.Op.String())
}

// Call expressions
func (jse *JSEmitter) PreVisitCallExpr(node *ast.CallExpr, indent int) {
	if jse.forwardDecl {
		return
	}
}

func (jse *JSEmitter) PreVisitCallExprFun(node ast.Expr, indent int) {
	// Don't emit here - the function name will be emitted by PreVisitIdent
	// through traverseExpression
}

func (jse *JSEmitter) PostVisitCallExprFun(node ast.Expr, indent int) {
	if jse.forwardDecl {
		return
	}
	jse.emitToFile("(")
}

func (jse *JSEmitter) PreVisitCallExprArg(node ast.Expr, index int, indent int) {
	if jse.forwardDecl {
		return
	}
	if index > 0 {
		jse.emitToFile(", ")
	}
}

func (jse *JSEmitter) PostVisitCallExprArgs(node []ast.Expr, indent int) {
	if jse.forwardDecl {
		return
	}
	jse.emitToFile(")")
}

// Return statements
func (jse *JSEmitter) PreVisitReturnStmt(node *ast.ReturnStmt, indent int) {
	if jse.forwardDecl {
		return
	}
	str := jse.emitAsString("return ", indent)
	jse.emitToFile(str)
	if len(node.Results) > 1 {
		jse.emitToFile("[")
		jse.inMultiValueReturn = true
		jse.multiValueReturnIndex = 0
	}
}

func (jse *JSEmitter) PreVisitReturnStmtResult(node ast.Expr, index int, indent int) {
	if jse.forwardDecl {
		return
	}
	if index > 0 {
		jse.emitToFile(", ")
	}
}

func (jse *JSEmitter) PostVisitReturnStmt(node *ast.ReturnStmt, indent int) {
	if jse.forwardDecl {
		return
	}
	if len(node.Results) > 1 {
		jse.emitToFile("]")
		jse.inMultiValueReturn = false
	}
	jse.emitToFile(";\n")
}

// If statements
func (jse *JSEmitter) PreVisitIfStmt(node *ast.IfStmt, indent int) {
	if jse.forwardDecl {
		return
	}
	str := jse.emitAsString("if (", indent)
	jse.emitToFile(str)
}

func (jse *JSEmitter) PostVisitIfStmtCond(node *ast.IfStmt, indent int) {
	if jse.forwardDecl {
		return
	}
	jse.emitToFile(") ")
}

func (jse *JSEmitter) PreVisitIfStmtElse(node *ast.IfStmt, indent int) {
	if jse.forwardDecl {
		return
	}
	str := jse.emitAsString(" else ", indent)
	jse.emitToFile(str)
}

// For statements
func (jse *JSEmitter) PreVisitForStmt(node *ast.ForStmt, indent int) {
	if jse.forwardDecl {
		return
	}
	// Check if infinite loop
	jse.isInfiniteLoop = node.Init == nil && node.Cond == nil && node.Post == nil
	if jse.isInfiniteLoop {
		str := jse.emitAsString("while (true) ", indent)
		jse.emitToFile(str)
	} else {
		str := jse.emitAsString("for (", indent)
		jse.emitToFile(str)
	}
}

func (jse *JSEmitter) PreVisitForStmtInit(node ast.Stmt, indent int) {
	if jse.forwardDecl || jse.isInfiniteLoop {
		return
	}
	jse.insideForInit = true
}

func (jse *JSEmitter) PostVisitForStmtInit(node ast.Stmt, indent int) {
	if jse.forwardDecl || jse.isInfiniteLoop {
		return
	}
	jse.insideForInit = false
	jse.emitToFile("; ")
}

func (jse *JSEmitter) PostVisitForStmtCond(node ast.Expr, indent int) {
	if jse.forwardDecl || jse.isInfiniteLoop {
		return
	}
	jse.emitToFile("; ")
}

func (jse *JSEmitter) PreVisitForStmtPost(node ast.Stmt, indent int) {
	if jse.forwardDecl || jse.isInfiniteLoop {
		return
	}
	jse.insideForPostCond = true
}

func (jse *JSEmitter) PostVisitForStmtPost(node ast.Stmt, indent int) {
	if jse.forwardDecl || jse.isInfiniteLoop {
		return
	}
	jse.insideForPostCond = false
	jse.emitToFile(") ")
}

// Range statements
func (jse *JSEmitter) PreVisitRangeStmt(node *ast.RangeStmt, indent int) {
	if jse.forwardDecl {
		return
	}
	// Handle different range patterns
	// Note: Go AST sets Key=nil when using blank identifier _, so we check Value first
	if node.Value != nil {
		// for key, value := range collection OR for _, value := range collection
		// When key is blank (_), Go AST has Key=nil, not Key=Ident{Name:"_"}
		jse.isKeyValueRange = true
		if node.Key == nil {
			// Key is nil (blank identifier _) - use synthetic index
			jse.rangeKeyName = "_idx"
		} else if keyIdent, ok := node.Key.(*ast.Ident); ok {
			if keyIdent.Name == "_" {
				// Explicit blank key
				jse.rangeKeyName = "_idx"
				DebugLogPrintf("JSEmitter: Range key is blank _, using _idx")
			} else {
				jse.rangeKeyName = keyIdent.Name
				DebugLogPrintf("JSEmitter: Range key is %s", keyIdent.Name)
			}
		} else {
			// Key is not a simple identifier, use synthetic index
			jse.rangeKeyName = "_idx"
			DebugLogPrintf("JSEmitter: Range key not ident, using _idx")
		}
		if valIdent, ok := node.Value.(*ast.Ident); ok {
			jse.rangeValueName = valIdent.Name
		}
		jse.rangeCollectionExpr = ""
		jse.suppressRangeEmit = true
		jse.rangeStmtIndent = indent
	} else if node.Key != nil {
		// for i := range collection (index-only)
		jse.isKeyValueRange = false
		if keyIdent, ok := node.Key.(*ast.Ident); ok {
			if keyIdent.Name == "_" {
				jse.rangeKeyName = "_idx"
			} else {
				jse.rangeKeyName = keyIdent.Name
			}
		} else {
			jse.rangeKeyName = "_idx"
		}
		jse.rangeCollectionExpr = ""
		jse.suppressRangeEmit = true
		jse.rangeStmtIndent = indent
	} else {
		DebugLogPrintf("JSEmitter: Range has nil Key and nil Value")
	}
}

func (jse *JSEmitter) PreVisitRangeStmtKey(node ast.Expr, indent int) {
	// Key is captured in PreVisitRangeStmt, suppress emission
	jse.suppressRangeEmit = true
}

func (jse *JSEmitter) PostVisitRangeStmtKey(node ast.Expr, indent int) {
	// Keep suppressing
}

func (jse *JSEmitter) PreVisitRangeStmtValue(node ast.Expr, indent int) {
	// Value is captured in PreVisitRangeStmt, suppress emission
}

func (jse *JSEmitter) PostVisitRangeStmtValue(node ast.Expr, indent int) {
	// Stop suppressing, start capturing collection expression
	jse.suppressRangeEmit = false
	jse.captureRangeExpr = true
}

func (jse *JSEmitter) PreVisitRangeStmtX(node ast.Expr, indent int) {
	if jse.forwardDecl {
		return
	}
	// Already in capture mode from PostVisitRangeStmtValue
}

func (jse *JSEmitter) PostVisitRangeStmtX(node ast.Expr, indent int) {
	if jse.forwardDecl {
		return
	}
	jse.captureRangeExpr = false
	collection := jse.rangeCollectionExpr
	key := jse.rangeKeyName
	rangeIndent := jse.rangeStmtIndent

	if jse.isKeyValueRange {
		// Emit: for (let key = 0; key < collection.length; key++)
		str := jse.emitAsString(fmt.Sprintf("for (let %s = 0; %s < %s.length; %s++) ", key, key, collection, key), rangeIndent)
		jse.emitToFile(str)
		// Set flag to emit value declaration in the block
		if jse.rangeValueName != "" && jse.rangeValueName != "_" {
			jse.pendingRangeValueDecl = true
		}
	} else {
		// Index-only: for (let i = 0; i < collection.length; i++)
		str := jse.emitAsString(fmt.Sprintf("for (let %s = 0; %s < %s.length; %s++) ", key, key, collection, key), rangeIndent)
		jse.emitToFile(str)
	}
}

func (jse *JSEmitter) PostVisitRangeStmt(node *ast.RangeStmt, indent int) {
	if jse.forwardDecl {
		return
	}
	jse.isKeyValueRange = false
	jse.rangeKeyName = ""
	jse.rangeValueName = ""
	jse.rangeCollectionExpr = ""
}

// Increment/Decrement statements
func (jse *JSEmitter) PreVisitIncDecStmt(node *ast.IncDecStmt, indent int) {
	if jse.forwardDecl {
		return
	}
	if !jse.insideForPostCond {
		str := jse.emitAsString("", indent)
		jse.emitToFile(str)
	}
}

func (jse *JSEmitter) PostVisitIncDecStmtX(node ast.Expr, indent int) {
	if jse.forwardDecl {
		return
	}
}

func (jse *JSEmitter) PostVisitIncDecStmt(node *ast.IncDecStmt, indent int) {
	if jse.forwardDecl {
		return
	}
	jse.emitToFile(node.Tok.String())
	if !jse.insideForPostCond {
		jse.emitToFile(";\n")
	}
}

// Index expressions (array access)
func (jse *JSEmitter) PreVisitIndexExpr(node *ast.IndexExpr, indent int) {
	if jse.forwardDecl {
		return
	}
}

func (jse *JSEmitter) PreVisitIndexExprIndex(node *ast.IndexExpr, indent int) {
	if jse.forwardDecl {
		return
	}
	jse.emitToFile("[")
}

func (jse *JSEmitter) PostVisitIndexExpr(node *ast.IndexExpr, indent int) {
	if jse.forwardDecl {
		return
	}
	jse.emitToFile("]")
}

// Composite literals (arrays, objects)
func (jse *JSEmitter) PreVisitCompositeLit(node *ast.CompositeLit, indent int) {
	if jse.forwardDecl {
		return
	}
	// Check if it's a struct or array
	if node.Type != nil {
		switch node.Type.(type) {
		case *ast.ArrayType:
			jse.emitToFile("[")
			return
		case *ast.Ident, *ast.SelectorExpr:
			// Struct initialization - use object literal syntax
			// Check if it has named fields (KeyValueExpr)
			if len(node.Elts) > 0 {
				if _, hasKeys := node.Elts[0].(*ast.KeyValueExpr); hasKeys {
					jse.emitToFile("{")
					return
				}
			}
			// Empty struct or positional values
			jse.emitToFile("{")
			return
		}
	}
	jse.emitToFile("[")
}

// Suppress type emission in composite literals (already handled in PreVisitCompositeLit)
func (jse *JSEmitter) PreVisitCompositeLitType(node ast.Expr, indent int) {
	jse.suppressTypeEmit = true
}

func (jse *JSEmitter) PostVisitCompositeLitType(node ast.Expr, indent int) {
	jse.suppressTypeEmit = false
}

func (jse *JSEmitter) PreVisitCompositeLitElt(node ast.Expr, index int, indent int) {
	if jse.forwardDecl {
		return
	}
	if index > 0 {
		jse.emitToFile(", ")
	}
}

func (jse *JSEmitter) PostVisitCompositeLit(node *ast.CompositeLit, indent int) {
	if jse.forwardDecl {
		return
	}
	if node.Type != nil {
		switch node.Type.(type) {
		case *ast.Ident, *ast.SelectorExpr:
			// Close struct/object literal
			jse.emitToFile("}")
			return
		}
	}
	jse.emitToFile("]")
}

// Selector expressions (field access)
func (jse *JSEmitter) PreVisitSelectorExpr(node *ast.SelectorExpr, indent int) {
	if jse.forwardDecl {
		return
	}
}

func (jse *JSEmitter) PostVisitSelectorExprX(node ast.Expr, indent int) {
	if jse.forwardDecl {
		return
	}
	// Don't emit dot when suppressing type emission
	if jse.suppressTypeEmit {
		return
	}
	// Only emit dot if X was not lowered to empty (e.g., fmt -> "")
	if ident, ok := node.(*ast.Ident); ok {
		if jse.lowerToBuiltins(ident.Name) == "" {
			return
		}
	}
	jse.emitToFile(".")
}

func (jse *JSEmitter) PreVisitSelectorExprSel(node *ast.Ident, indent int) {
	// Selector name is emitted via PreVisitIdent when Sel is traversed
}

// Type specifications (structs)
func (jse *JSEmitter) PreVisitTypeSpec(node *ast.TypeSpec, indent int) {
	if jse.forwardDecl {
		return
	}
	// Check if it's a struct type
	if structType, ok := node.Type.(*ast.StructType); ok {
		str := jse.emitAsString("class "+node.Name.Name+" {\n", indent)
		jse.emitToFile(str)

		// Generate constructor
		str = jse.emitAsString("constructor(", indent+1)
		jse.emitToFile(str)

		// Collect field names
		var fieldNames []string
		for _, field := range structType.Fields.List {
			for _, name := range field.Names {
				fieldNames = append(fieldNames, name.Name)
			}
		}
		jse.emitToFile(strings.Join(fieldNames, ", "))
		jse.emitToFile(") {\n")

		// Initialize fields
		for _, name := range fieldNames {
			str = jse.emitAsString("this."+name+" = "+name+";\n", indent+2)
			jse.emitToFile(str)
		}

		str = jse.emitAsString("}\n", indent+1)
		jse.emitToFile(str)
		str = jse.emitAsString("}\n\n", indent)
		jse.emitToFile(str)
	} else if _, ok := node.Type.(*ast.ArrayType); ok {
		// Type alias for array - just emit a comment
		str := jse.emitAsString("// type "+node.Name.Name+" = array\n", indent)
		jse.emitToFile(str)
	}
}

// Struct field type and name handling - suppress type emissions
func (jse *JSEmitter) PreVisitGenStructFieldType(node ast.Expr, indent int) {
	jse.suppressTypeEmit = true
}

func (jse *JSEmitter) PostVisitGenStructFieldType(node ast.Expr, indent int) {
	jse.suppressTypeEmit = false
}

func (jse *JSEmitter) PreVisitGenStructFieldName(node *ast.Ident, indent int) {
	// Field names are handled in PreVisitTypeSpec, suppress here
	jse.suppressTypeEmit = true
}

func (jse *JSEmitter) PostVisitGenStructFieldName(node *ast.Ident, indent int) {
	jse.suppressTypeEmit = false
}

// Type alias handling - suppress in JavaScript (no type aliases needed)
func (jse *JSEmitter) PreVisitTypeAliasName(node *ast.Ident, indent int) {
	jse.suppressTypeEmit = true
}

func (jse *JSEmitter) PostVisitTypeAliasName(node *ast.Ident, indent int) {
	// Keep suppressed until after the type
}

func (jse *JSEmitter) PreVisitTypeAliasType(node ast.Expr, indent int) {
	// Type still suppressed
}

func (jse *JSEmitter) PostVisitTypeAliasType(node ast.Expr, indent int) {
	jse.suppressTypeEmit = false
}

// Declaration statements (var a int, var b []string, etc.)
func (jse *JSEmitter) PreVisitDeclStmt(node *ast.DeclStmt, indent int) {
	if jse.forwardDecl {
		return
	}
}

func (jse *JSEmitter) PreVisitDeclStmtValueSpecType(node *ast.ValueSpec, index int, indent int) {
	// Suppress type emission - JavaScript doesn't need type annotations
	jse.suppressTypeEmit = true
	// Emit "let " for variable declaration
	str := jse.emitAsString("let ", indent)
	jse.emitToFile(str)
	// Check if we need to add default initialization
	if len(node.Values) == 0 && node.Type != nil {
		switch t := node.Type.(type) {
		case *ast.ArrayType:
			// Slice/array type - initialize to []
			jse.pendingSliceInit = true
		case *ast.Ident:
			// Custom type (struct) - initialize to {} unless it's a built-in type
			if !isBuiltinType(t.Name) {
				jse.pendingStructInit = true
			}
		case *ast.SelectorExpr:
			// External package type (struct) - initialize to {}
			jse.pendingStructInit = true
		}
	}
}

func (jse *JSEmitter) PostVisitDeclStmtValueSpecType(node *ast.ValueSpec, index int, indent int) {
	jse.suppressTypeEmit = false
}

func (jse *JSEmitter) PreVisitDeclStmtValueSpecNames(node *ast.Ident, index int, indent int) {
	// Variable name is emitted via PreVisitIdent when traversed
}

func (jse *JSEmitter) PostVisitDeclStmtValueSpecNames(node *ast.Ident, index int, indent int) {
	if jse.forwardDecl {
		return
	}
	if jse.pendingSliceInit {
		jse.emitToFile(" = []")
		jse.pendingSliceInit = false
	} else if jse.pendingStructInit {
		jse.emitToFile(" = {}")
		jse.pendingStructInit = false
	}
	jse.emitToFile(";\n")
}

// isBuiltinType returns true if the type name is a Go built-in type
func isBuiltinType(name string) bool {
	switch name {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64", "complex64", "complex128",
		"bool", "string", "byte", "rune", "error", "any":
		return true
	}
	return false
}

func (jse *JSEmitter) PostVisitDeclStmt(node *ast.DeclStmt, indent int) {
	if jse.forwardDecl {
		return
	}
}

// Variable declarations
func (jse *JSEmitter) PreVisitGenDeclVar(node *ast.GenDecl, indent int) {
	if jse.forwardDecl {
		return
	}
}

func (jse *JSEmitter) PreVisitValueSpec(node *ast.ValueSpec, indent int) {
	if jse.forwardDecl {
		return
	}
	str := jse.emitAsString("let ", indent)
	jse.emitToFile(str)
}

func (jse *JSEmitter) PreVisitValueSpecName(node *ast.Ident, index int, indent int) {
	if jse.forwardDecl {
		return
	}
	if index > 0 {
		jse.emitToFile(", ")
	}
	jse.emitToFile(node.Name)
}

func (jse *JSEmitter) PreVisitValueSpecValue(node ast.Expr, index int, indent int) {
	if jse.forwardDecl {
		return
	}
	jse.emitToFile(" = ")
}

func (jse *JSEmitter) PostVisitValueSpec(node *ast.ValueSpec, indent int) {
	if jse.forwardDecl {
		return
	}
	// If no value, initialize with default
	if len(node.Values) == 0 {
		if node.Type != nil {
			jse.emitToFile(" = ")
			jse.emitDefaultValue(node.Type)
		}
	}
	jse.emitToFile(";\n")
}

func (jse *JSEmitter) emitDefaultValue(typeExpr ast.Expr) {
	switch t := typeExpr.(type) {
	case *ast.Ident:
		switch t.Name {
		case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64":
			jse.emitToFile("0")
		case "bool":
			jse.emitToFile("false")
		case "string":
			jse.emitToFile("\"\"")
		default:
			jse.emitToFile("null")
		}
	case *ast.ArrayType:
		jse.emitToFile("[]")
	default:
		jse.emitToFile("null")
	}
}

// Constant declarations
func (jse *JSEmitter) PreVisitGenDeclConst(node *ast.GenDecl, indent int) {
	if jse.forwardDecl {
		return
	}
}

func (jse *JSEmitter) PreVisitGenDeclConstName(node *ast.Ident, indent int) {
	if jse.forwardDecl {
		return
	}
	str := jse.emitAsString("const "+node.Name+" = ", indent)
	jse.emitToFile(str)
}

func (jse *JSEmitter) PostVisitGenDeclConstName(node *ast.Ident, indent int) {
	if jse.forwardDecl {
		return
	}
	jse.emitToFile(";\n")
}

// Parenthesized expressions
func (jse *JSEmitter) PreVisitParenExpr(node *ast.ParenExpr, indent int) {
	if jse.forwardDecl {
		return
	}
	jse.emitToFile("(")
}

func (jse *JSEmitter) PostVisitParenExpr(node *ast.ParenExpr, indent int) {
	if jse.forwardDecl {
		return
	}
	jse.emitToFile(")")
}

// Break and continue
func (jse *JSEmitter) PreVisitBranchStmt(node *ast.BranchStmt, indent int) {
	if jse.forwardDecl {
		return
	}
	str := jse.emitAsString(node.Tok.String()+";\n", indent)
	jse.emitToFile(str)
}

// Switch statements
func (jse *JSEmitter) PreVisitSwitchStmt(node *ast.SwitchStmt, indent int) {
	if jse.forwardDecl {
		return
	}
	str := jse.emitAsString("switch (", indent)
	jse.emitToFile(str)
}

func (jse *JSEmitter) PostVisitSwitchStmtTag(node ast.Expr, indent int) {
	if jse.forwardDecl {
		return
	}
	jse.emitToFile(") {\n")
}

func (jse *JSEmitter) PostVisitSwitchStmt(node *ast.SwitchStmt, indent int) {
	if jse.forwardDecl {
		return
	}
	str := jse.emitAsString("}\n", indent)
	jse.emitToFile(str)
}

func (jse *JSEmitter) PreVisitCaseClause(node *ast.CaseClause, indent int) {
	if jse.forwardDecl {
		return
	}
	if len(node.List) == 0 {
		str := jse.emitAsString("default:\n", indent)
		jse.emitToFile(str)
	} else {
		str := jse.emitAsString("case ", indent)
		jse.emitToFile(str)
	}
}

func (jse *JSEmitter) PreVisitCaseClauseExpr(node ast.Expr, index int, indent int) {
	if jse.forwardDecl {
		return
	}
	if index > 0 {
		jse.emitToFile(", ")
	}
}

func (jse *JSEmitter) PostVisitCaseClauseList(node []ast.Expr, indent int) {
	if jse.forwardDecl {
		return
	}
	if len(node) > 0 {
		jse.emitToFile(":\n")
	}
}

func (jse *JSEmitter) PostVisitCaseClause(node *ast.CaseClause, indent int) {
	if jse.forwardDecl {
		return
	}
	// Add break if no fallthrough (Go's default behavior)
	str := jse.emitAsString("break;\n", indent+1)
	jse.emitToFile(str)
}

// Key-value expressions (struct field initialization)
func (jse *JSEmitter) PreVisitKeyValueExpr(node *ast.KeyValueExpr, indent int) {
	if jse.forwardDecl {
		return
	}
}

func (jse *JSEmitter) PreVisitKeyValueExprValue(node ast.Expr, indent int) {
	if jse.forwardDecl {
		return
	}
	jse.emitToFile(": ")
}

// Slice expressions
// Go: a[low:high] => JS: a.slice(low, high)
// Go: a[low:] => JS: a.slice(low)
// Go: a[:high] => JS: a.slice(0, high)
func (jse *JSEmitter) PreVisitSliceExpr(node *ast.SliceExpr, indent int) {
	if jse.forwardDecl {
		return
	}
}

func (jse *JSEmitter) PostVisitSliceExprX(node ast.Expr, indent int) {
	if jse.forwardDecl {
		return
	}
	// After the array name, emit .slice(
	jse.emitToFile(".slice(")
}

func (jse *JSEmitter) PreVisitSliceExprXBegin(node ast.Expr, indent int) {
	// Suppress the second X visit - it's for Go's internal slice bounds
	jse.suppressRangeEmit = true
}

func (jse *JSEmitter) PostVisitSliceExprXBegin(node ast.Expr, indent int) {
	jse.suppressRangeEmit = false
}

func (jse *JSEmitter) PreVisitSliceExprLow(node ast.Expr, indent int) {
	if jse.forwardDecl {
		return
	}
	// If Low is nil (like a[:high]), emit 0
	if node == nil {
		jse.emitToFile("0")
	}
}

func (jse *JSEmitter) PostVisitSliceExprLow(node ast.Expr, indent int) {
	if jse.forwardDecl {
		return
	}
	// We'll add comma in PreVisitSliceExprHigh if High is not nil
}

func (jse *JSEmitter) PreVisitSliceExprXEnd(node ast.Expr, indent int) {
	// Suppress the third X visit
	jse.suppressRangeEmit = true
}

func (jse *JSEmitter) PostVisitSliceExprXEnd(node ast.Expr, indent int) {
	jse.suppressRangeEmit = false
}

func (jse *JSEmitter) PreVisitSliceExprHigh(node ast.Expr, indent int) {
	if jse.forwardDecl {
		return
	}
	// If High is not nil, emit comma before it
	if node != nil {
		jse.emitToFile(", ")
	}
}

func (jse *JSEmitter) PostVisitSliceExpr(node *ast.SliceExpr, indent int) {
	if jse.forwardDecl {
		return
	}
	jse.emitToFile(")")
}

// Type assertions (limited support)
func (jse *JSEmitter) PreVisitTypeAssertExpr(node *ast.TypeAssertExpr, indent int) {
	if jse.forwardDecl {
		return
	}
	// In JavaScript, just return the value (dynamic typing)
}

func (jse *JSEmitter) PostVisitTypeAssertExpr(node *ast.TypeAssertExpr, indent int) {
	if jse.forwardDecl {
		return
	}
	// No-op for JavaScript
}

// Function literals (closures)
func (jse *JSEmitter) PreVisitFuncLit(node *ast.FuncLit, indent int) {
	if jse.forwardDecl {
		return
	}
	jse.emitToFile("function(")
}

func (jse *JSEmitter) PreVisitFuncLitTypeParams(node *ast.FieldList, indent int) {
	// Start of parameters - suppress type emission
	jse.suppressTypeEmit = true
}

func (jse *JSEmitter) PreVisitFuncLitTypeParam(node *ast.Field, index int, indent int) {
	if jse.forwardDecl {
		return
	}
	if index > 0 {
		jse.emitToFile(", ")
	}
	for i, name := range node.Names {
		if i > 0 {
			jse.emitToFile(", ")
		}
		jse.emitToFile(name.Name)
	}
}

func (jse *JSEmitter) PostVisitFuncLitTypeParams(node *ast.FieldList, indent int) {
	if jse.forwardDecl {
		return
	}
	jse.suppressTypeEmit = false
	jse.emitToFile(") ")
}

// Helper to check if type needs special handling
func (jse *JSEmitter) getJSType(goType string) string {
	if jsType, ok := jsTypesMap[goType]; ok {
		return jsType
	}
	return goType
}

// mapGoTypeToJS converts Go types to JavaScript type comments
func (jse *JSEmitter) mapGoTypeToJS(t types.Type) string {
	if t == nil {
		return "any"
	}
	switch underlying := t.Underlying().(type) {
	case *types.Basic:
		return jse.getJSType(underlying.Name())
	case *types.Slice:
		return "Array"
	case *types.Struct:
		return "Object"
	default:
		return "any"
	}
}

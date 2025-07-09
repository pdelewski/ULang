package main

import (
	"fmt"
	"go/ast"
	"go/token"
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

func (re *RustEmitter) PreVisitIdent(e *ast.Ident, indent int) {
	var str string
	name := e.Name
	name = re.lowerToBuiltins(name)
	if name == "nil" {
		str = re.emitAsString("{}", indent)
		re.emitToFileBuffer(str, "")
	} else {
		if n, ok := rustTypesMap[name]; ok {
			str = re.emitAsString(n, indent)
			re.emitToFileBuffer(str, "")
		} else {
			str = re.emitAsString(name, indent)
			re.emitToFileBuffer(str, "")
		}
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

package main

import (
	"fmt"
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

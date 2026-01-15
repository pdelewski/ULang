package main

// This file contains all supported Go language constructs that compile
// successfully across all backends (C++, C#, Rust).
//
// UNSUPPORTED CONSTRUCTS (not included in this file):
//
// 1. for i, x := range slice - Range loop with both index and value
//    Only "for _, x := range" (value only) is supported
//
// 2. if slice == nil - Nil comparison for slices
//    C++ std::vector cannot be compared to nullptr
//
// 3. len(string) - String length
//    C++ backend uses std::size() which doesn't work on C-style strings
//
// 4. for condition { } - While-style loops
//    C# backend has a bug with semicolons in loop body
//
// 5. iota - Constant enumeration
//    Not yet implemented
//
// 6. fmt.Sprintf - String formatting
//    Rust backend has type mismatch issues with string_format2
//
// 7. for _, x := range []int{1,2,3} - Range over inline slice literal
//    Rust backend generates malformed code
//
// 8. []interface{} - Slice of empty interface (any type)
//    Not supported across backends

import (
	"alltests/types"
	"fmt"
)

// Struct type declaration with slice field
type Composite struct {
	a []int
}

// Struct with multiple field types
type Person struct {
	name string
	age  int
}

// Basic function with single return value
func testBasicConstructs() int8 {
	testSliceOperations()
	testLoopConstructs()
	testBooleanLogic()
	return 5
}

// Function with multiple return values
func testFunctionCalls() (int16, int16) {
	return testFunctionVariables()
}

// Slice operations: nil slice, len, indexing, struct field access
func testSliceOperations() {
	var a []int
	c := Composite{}

	// Slice literal with int type (from slice test)
	intSlice := []int{1, 2, 3}
	fmt.Println(len(intSlice))

	if len(a) == 0 {
	} else {
		if a[0] == 0 {
			a[0] = 1
		}
	}

	if len(c.a) == 0 {
	}
}

// Loop constructs: C-style for, range for
func testLoopConstructs() {
	var a []int

	// C-style for loop
	for x := 0; x < 10; x++ {
		if !(len(a) == 0) {
		} else if len(a) == 0 {
		}
	}

	// Range-based for loop with blank identifier
	for _, x := range a {
		if x == 0 {
		}
	}
}

// Boolean logic: not operator, boolean literals
func testBooleanLogic() {
	b := false
	if !b {
	}

	c := true
	if c {
	}
}

// Function types: slice of functions, closures, calling through variables
func testFunctionVariables() (int16, int16) {
	x := []func(int, int){
		func(a int, b int) {
			fmt.Println(a)
			fmt.Println(b)
		},
	}

	f := x[0]
	f(10, 20)
	x[0](20, 30)

	if len(x) == 0 {
	}

	return 10, 20
}

// Sink function for consuming values
func sink(p int8) {
}

// Empty slice and slice with values initialization
func testArrayInitialization() {
	a := []int8{}
	if len(a) == 0 {
	}

	b := []int8{1, 2, 3}
	if len(b) == 0 {
	}
}

// Slice expressions: slicing with start index
func testSliceExpressions() {
	a := []int8{1, 2, 3}

	// Slice from index to end
	b := a[1:]
	if len(b) == 0 {
	}

	// Slice from start to index
	c := a[:2]
	if len(c) == 0 {
	}

	// Slice with both bounds
	d := a[1:2]
	if len(d) == 0 {
	}
}

// Variable declarations: var, short declaration, multiple on one line
func testVariableDeclarations() {
	var a int8
	var b, c int16

	a = 1
	a = a + 5
	d := 10

	sink(a)
	if b == 0 {
	}
	if c == 0 {
	}
	if d == 10 {
	}
}

// Arithmetic operators
func testArithmeticOperators() {
	a := 10
	b := 3

	sum := a + b
	diff := a - b
	prod := a * b
	quot := a / b
	rem := a % b

	fmt.Println(sum)
	fmt.Println(diff)
	fmt.Println(prod)
	fmt.Println(quot)
	fmt.Println(rem)
}

// Comparison operators
func testComparisonOperators() {
	a := 10
	b := 20

	if a == b {
	}
	if a != b {
	}
	if a < b {
	}
	if a > b {
	}
	if a <= b {
	}
	if a >= b {
	}
}

// Logical operators
func testLogicalOperators() {
	a := true
	b := false

	if a && b {
	}
	if a || b {
	}
	if !a {
	}
}

// Assignment operators
func testAssignmentOperators() {
	a := 10
	a = 20
	a += 5
	a -= 3

	fmt.Println(a)
}

// Increment and decrement
func testIncrementDecrement() {
	a := 0
	a++
	a--
	fmt.Println(a)
}

// String operations
func testStringOperations() {
	s := "hello"
	fmt.Println(s)
}

// Print functions
func testPrintFunctions() {
	// Print with newline
	fmt.Println("Hello")
	fmt.Println(42)
	fmt.Println()

	// Print without newline
	fmt.Print("World")
	fmt.Print("\n")

	// Printf with format specifiers
	fmt.Printf("%d\n", 100)
	fmt.Printf("%s\n", "test")
}

// Type conversions
func testTypeConversions() {
	a := 65
	b := int8(a)
	sink(b)
}

// Append operation
func testAppend() {
	a := []int{}
	a = append(a, 1)
	a = append(a, 2)
	a = append(a, 3)
	fmt.Println(len(a))
}

// Struct initialization
func testStructInitialization() {
	// Empty struct
	c := Composite{}
	if len(c.a) == 0 {
	}

	// Struct with field values
	p := Person{name: "Alice", age: 30}
	fmt.Println(p.name)
	fmt.Println(p.age)
}

// Nested if statements
func testNestedIf() {
	a := 10
	b := 20

	if a == 10 {
		if b == 20 {
			fmt.Println("nested")
		}
	}
}

// Test int32, int64 types (from iceberg)
func testInt32Int64Types() {
	var a int32
	var b int64

	a = 100
	b = 200

	fmt.Println(a)
	fmt.Println(b)

	// Struct with int32/int64 fields
	record := types.DataRecord{
		ID:          1,
		Size:        1024,
		Count:       10,
		SequenceNum: 999,
	}
	fmt.Println(record.ID)
	fmt.Println(record.Size)
}

// Test type aliases (from substrait)
func testTypeAliases() {
	var kind types.ExprKind
	kind = types.ExprLiteral

	if kind == types.ExprLiteral {
		fmt.Println("literal")
	}
	if kind == types.ExprColumn {
		fmt.Println("column")
	}

	var relKind types.RelNodeKind
	relKind = types.RelOpScan
	if relKind == types.RelOpScan {
		fmt.Println("scan")
	}
}

// Test fmt.Printf with multiple arguments (from substrait)
func testPrintfMultipleArgs() {
	fmt.Printf("a=%d, b=%d\n", 10, 20)
	fmt.Printf("name=%s, value=%d\n", "test", 100)
}

// Test zero-value struct declaration (from substrait)
func testZeroValueStruct() {
	var plan types.Plan
	plan.Literals = []string{}
	plan.Root = 0
	fmt.Println(plan.Root)
	fmt.Println(len(plan.Literals))
}

// Test function returning modified struct
func testReturnModifiedStruct() {
	var plan types.Plan
	plan.Literals = []string{}

	idx := 0
	plan, idx = types.AddLiteralToPlan(plan, "first")
	plan, idx = types.AddLiteralToPlan(plan, "second")

	fmt.Println(idx)
	fmt.Println(len(plan.Literals))
}

// Test bool field in struct (from iceberg)
func testBoolFieldInStruct() {
	f := types.Field{
		ID:       1,
		Name:     "column1",
		Required: true,
	}
	fmt.Println(f.ID)
	fmt.Println(f.Name)
	if f.Required {
		fmt.Println("required")
	}

	f2 := types.Field{
		ID:       2,
		Name:     "column2",
		Required: false,
	}
	if !f2.Required {
		fmt.Println("optional")
	}
}

// Test nested struct field (from iceberg)
func testNestedStructField() {
	stats := types.ColumnStats{NullCount: 100}
	dataFile := types.DataFile{
		FilePath:    "/path/to/file",
		RecordCount: 1000,
		Stats:       stats,
	}
	entry := types.ManifestEntry{
		Status:    1,
		DataFileF: dataFile,
	}

	fmt.Println(entry.Status)
	fmt.Println(entry.DataFileF.FilePath)
	fmt.Println(entry.DataFileF.RecordCount)
	fmt.Println(entry.DataFileF.Stats.NullCount)
}

// Test multi-package import (from iceberg pattern)
func testMultiPackageImport() {
	// Use types from the types package
	record := types.DataRecord{
		ID:          42,
		Size:        2048,
		Count:       5,
		SequenceNum: 100,
	}
	types.LoadData(record)

	// Use function from types package
	var plan types.Plan
	plan.Literals = []string{}
	idx := 0
	plan, idx = types.AddLiteralToPlan(plan, "value1")
	plan, idx = types.AddLiteralToPlan(plan, "value2")
	fmt.Println(idx)
	fmt.Println(len(plan.Literals))

	// Use constants from types package
	if types.ExprLiteral == 0 {
		fmt.Println("literal is 0")
	}

	// Use nested struct from types package
	entry := types.ManifestEntry{
		Status: 1,
		DataFileF: types.DataFile{
			FilePath:    "/data/file.parquet",
			RecordCount: 500,
			Stats:       types.ColumnStats{NullCount: 10},
		},
	}
	fmt.Println(entry.DataFileF.FilePath)
}

// Complete language feature test
func testCompleteLanguageFeatures() {
	var a int8
	var b, c int16

	a = 1
	a = a + 5
	d := 10

	a = testBasicConstructs()
	b, c = testFunctionCalls()

	if (a == 1) && (b == 10) {
		a = 2
		var aa int8
		aa = testBasicConstructs()
		sink(aa)

		if a == 5 {
			a = 10
		}
	} else {
		a = 3
	}

	if b == 10 {
	}
	if c == 20 {
	}
	if d == 10 {
	}
}

func main() {
	fmt.Println("=== All Language Constructs Test ===")

	testCompleteLanguageFeatures()
	testArrayInitialization()
	testSliceExpressions()
	testVariableDeclarations()
	testArithmeticOperators()
	testComparisonOperators()
	testLogicalOperators()
	testAssignmentOperators()
	testIncrementDecrement()
	testStringOperations()
	testPrintFunctions()
	testTypeConversions()
	testAppend()
	testStructInitialization()
	testNestedIf()
	testInt32Int64Types()
	testTypeAliases()
	testPrintfMultipleArgs()
	testZeroValueStruct()
	testReturnModifiedStruct()
	testBoolFieldInStruct()
	testNestedStructField()
	testMultiPackageImport()

	fmt.Println("=== Done ===")
}

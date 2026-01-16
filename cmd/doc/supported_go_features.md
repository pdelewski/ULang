# Supported Go Features

This document lists all Go language features supported by the goany transpiler, based on analysis of the test cases and library code.

## 1. Primitive Types

| Type | Description | Example |
|------|-------------|---------|
| `int8` | 8-bit signed integer | `var a int8` |
| `int16` | 16-bit signed integer | `var b int16` |
| `int32` | 32-bit signed integer | `var c int32` |
| `int64` | 64-bit signed integer | `var d int64` |
| `int` | Platform-dependent integer | `value int` |
| `uint8` | 8-bit unsigned integer | `var a uint8` |
| `uint16` | 16-bit unsigned integer | `var b uint16` |
| `uint32` | 32-bit unsigned integer | `var c uint32` |
| `uint64` | 64-bit unsigned integer | `var d uint64` |
| `string` | String type | `var s string` |
| `bool` | Boolean type | `b := false` |
| `rune` | Unicode code point (via range) | `for _, r := range s` |

## 2. Composite Types

### Structs

```go
type ListNode struct {
    value int
    next  int
}

type List struct {
    nodes []ListNode
    head  int
}
```

### Slices

```go
var a []int              // nil slice declaration
b := []int{1, 2, 3}      // slice literal with values
c := []int{}             // empty slice
```

### Type Aliases

```go
type ExprKind int
type AST []Statement     // slice of structs
```

## 3. Variable Declarations

### var Declaration

```go
var a int8               // zero-value initialization
var b, c int16           // multiple declarations
var s string
```

### Short Declaration (:=)

```go
b := len(a)              // type inference
c := []int{1, 2, 3}      // slice literal
d := Composite{}         // struct literal
```

### Constants

```go
const (
    TokenTypeIdent = iota  // iota enumeration
    TokenTypeSpace
    TokenTypeSpecialChar
)

const MaxValue = 100       // explicit value
```

## 4. Control Flow

### if/else

```go
if len(a) == 0 {
    // empty case
} else if a[0] == 0 {
    a[0] = 1
} else {
    // default case
}
```

### switch

```go
switch statement.Type {
case ast.StatementTypeFrom:
    state = ast.WalkFrom(...)
case ast.StatementTypeWhere:
    state = ast.WalkWhere(...)
case ast.StatementTypeSelect:
    state = ast.WalkSelect(...)
}
```

### for Loop - C-style

```go
for x := 0; x < 10; x++ {
    // loop body
}
```

### for Range Loop

```go
// Range over slice with index and value
for i, x := range a {
    // use i and x
}

// Range over slice, ignore index
for _, x := range a {
    // use x
}

// Range over string (rune iteration)
for _, r := range s {
    // r is rune
}
```

### for While-style Loop

```go
for l.nodes[lastNodeIndex].next != -1 {
    lastNodeIndex = l.nodes[lastNodeIndex].next
}
```

### return Statement

```go
return 5                 // single value
return 10, 20            // multiple values
return                   // void return
```

### break and continue

```go
for i := 0; i < n; i++ {
    if condition {
        break
    }
    if otherCondition {
        continue
    }
}
```

## 5. Functions

### Basic Function Declaration

```go
func testBasicConstructs() int8 {
    return 5
}

func sink(p int8) {
    // no return
}
```

### Multiple Return Values

```go
func testFunctionCalls() (int16, int16) {
    return 10, 20
}

func GetNextToken(tokens []Token) (Token, []Token) {
    // return token and remaining tokens
}
```

### Multiple Return Value Assignment

```go
b, c = testFunctionCalls()
token, remaining := GetNextToken(tokens)
```

### Function Parameters

```go
func Add(l List, value int) List {
    // pass by value
    return l
}

func IsDigit(b int8) bool {
    return b >= 48 && b <= 57
}
```

### Closures/Anonymous Functions

```go
x := []func(int, int){
    func(a int, b int) {
        fmt.Println(a)
        fmt.Println(b)
    },
}

addToken := func(t Token) {
    tokens = append(tokens, t)
}
```

### Function Types in Structs

```go
type Visitor struct {
    PreVisitFrom  func(state any, expr From) any
    PostVisitFrom func(state any, expr From) any
}
```

### Function as First-Class Values

```go
f := x[0]        // assign function to variable
f(10, 20)        // call through variable
x[0](20, 30)     // call through slice index
```

## 6. Operators

### Arithmetic Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `+` | Addition | `a + 5` |
| `-` | Subtraction | `a - 1` |
| `*` | Multiplication | `a * 2` |
| `/` | Division | `a / 2` |
| `%` | Modulo | `a % 2` |

### Comparison Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `==` | Equal | `a == 0` |
| `!=` | Not equal | `a != -1` |
| `<` | Less than | `x < 10` |
| `>` | Greater than | `x > 0` |
| `<=` | Less or equal | `x <= n` |
| `>=` | Greater or equal | `x >= 0` |

### Logical Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `&&` | Logical AND | `(a == 1) && (b == 10)` |
| `\|\|` | Logical OR | `a \|\| b` |
| `!` | Logical NOT | `!b` |

### Assignment Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `=` | Assignment | `a = 1` |
| `:=` | Short declaration | `a := 1` |
| `+=` | Add and assign | `a += 5` |
| `-=` | Subtract and assign | `a -= 1` |

### Unary Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `!` | Logical NOT | `!b` |
| `++` | Post-increment | `x++` |
| `--` | Post-decrement | `x--` |

## 7. Expressions

### Index Expressions

```go
a[0]                     // slice/array index
token.Representation[i]  // nested index
```

### Slice Expressions

```go
a[1:]                    // from index 1 to end
a[:n]                    // from start to index n
a[i:j]                   // from index i to j
```

### Composite Literals

```go
// Empty struct
c := Composite{}

// Struct with field initialization
node := ListNode{value: 10, next: -1}

// Struct with partial initialization
return List{
    nodes: []ListNode{},
    head:  -1,
}

// Slice literal
nums := []int{1, 2, 3}

// Slice of structs
tokens := []Token{
    {Type: TokenTypeIdent, Representation: []int8{...}},
}
```

### Type Conversions

```go
int8(r)                  // convert to int8
int8(text[i])            // convert byte to int8
```

### Type Assertions

```go
newState := state.(State)  // assert interface to concrete type
```

## 8. Argument Passing

### Pass by Value

All arguments are passed by value. Functions that need to modify data return the modified copy:

```go
func Add(l List, value int) List {
    // modify l
    return l
}

// Usage:
list = Add(list, 10)
```

### Interface Types (any)

```go
func WalkFrom(expr From, state any, visitor Visitor) any {
    // state can be any type
    return state
}

var state any
state = State{depth: 0}
newState := state.(State)  // type assertion
```

## 9. Built-in Functions

| Function | Description | Example |
|----------|-------------|---------|
| `len()` | Length of slice/string | `len(a)` |
| `append()` | Append to slice | `append(a, x)` |
| `fmt.Println()` | Print with newline | `fmt.Println(a)` |
| `fmt.Printf()` | Formatted print | `fmt.Printf("%d", a)` |
| `fmt.Sprintf()` | Formatted string | `fmt.Sprintf("%d", a)` |

## 10. Package System

### Package Declaration

```go
package main
package containers
package lexer
```

### Import Statements

```go
// Single import
import "fmt"

// Grouped imports
import (
    "fmt"
    "uql/ast"
    "uql/lexer"
)
```

## 11. Other Constructs

### Blank Identifier

```go
for _, x := range a {    // ignore index
    // use x only
}
```

### Nil Values

```go
var a []int              // nil slice
a = nil                  // assign nil
if a == nil { }          // nil check
```

### Comments

```go
// Single line comment

/*
   Multi-line
   comment
*/
```

## Unsupported Features

The following Go features are NOT currently supported:

- Pointers (`*T`, `&x`)
- Methods with receivers
- Goroutines and channels
- Defer statements
- Maps
- Interfaces (except `any`/`interface{}`)
- Error type and error handling patterns
- Named return values
- Variadic functions (`...`)
- Init functions
- Struct embedding
- Type switches
- Select statements
- Goto statements
- Labels

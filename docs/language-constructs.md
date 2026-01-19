# goany Language Constructs

goany is a Go-to-C++/Rust/C# transpiler. This document specifies the supported subset of Go syntax.

## Types

### Primitive Types
- `int`, `int8`, `int16`, `int32`, `int64`
- `uint8`
- `bool`
- `string`

### Composite Types
- `[]T` - Slice of type T
- `struct` - Named struct types
- `func(params) return` - Function types

### Type Aliases
```go
type ExprKind int
```

## Variables

### Declaration
```go
var a int8
var b, c int16
d := 10
```

### Assignment
```go
a = 1
a += 5
a -= 3
a++
a--
```

## Functions

### Definition
```go
func name(param1 int, param2 string) int {
    return 0
}
```

### Multiple Return Values
```go
func multi() (int, int) {
    return 10, 20
}
```

### Function Variables
```go
x := []func(int, int){
    func(a int, b int) { },
}
f := x[0]
f(10, 20)
```

## Structs

### Definition
```go
type Person struct {
    name string
    age  int
}
```

### Initialization
```go
p := Person{name: "Alice", age: 30}
c := Composite{}
```

### Nested Structs
```go
type Inner struct { Value int }
type Outer struct { Data Inner }
o := Outer{Data: Inner{Value: 10}}
```

### Field Access
```go
p.name
o.Data.Value
```

## Slices

### Initialization
```go
a := []int{}
b := []int{1, 2, 3}
var c []int
```

### Operations
```go
len(a)
a[0]
a[0] = 1
a = append(a, value)
```

### Slicing
```go
b := a[1:]
c := a[:2]
d := a[1:2]
```

## Control Flow

### If/Else
```go
if condition {
} else if condition2 {
} else {
}
```

### C-style For Loop
```go
for i := 0; i < 10; i++ {
}
```

### While-style Loop
```go
for condition {
}
```

### Infinite Loop
```go
for {
    if done { break }
}
```

### Range Loop (value only)
```go
for _, x := range slice {
}
```

### Range Loop (index only)
```go
for i := range slice {
}
```

### Break and Continue
```go
break
continue
```

### Switch
```go
switch x {
case 1:
    // ...
case 2, 3:
    // ...
default:
    // ...
}
```

## Operators

### Arithmetic
`+`, `-`, `*`, `/`, `%`

### Comparison
`==`, `!=`, `<`, `>`, `<=`, `>=`

### Logical
`&&`, `||`, `!`

### Bitwise
`&`, `|`, `>>`, `<<`

## Type Conversion
```go
b := int8(a)
```

## Constants
```go
const (
    Value1 TypeName = 0
    Value2 TypeName = 1
)
```

## Packages

### Import
```go
import (
    "fmt"
    "mypackage/subpkg"
)
```

### Package Functions
```go
subpkg.FunctionName(args)
```

## Standard Library

### fmt Package
```go
fmt.Println(value)
fmt.Print(value)
fmt.Printf("%d %s\n", intVal, strVal)
```

## Unsupported Constructs

| Construct | Reason |
|-----------|--------|
| `for i, x := range slice` | Only `_` or index-only supported |
| `slice == nil` | C++ std::vector incompatible |
| `len(string)` | Backend incompatibility |
| `iota` | Not implemented |
| `fmt.Sprintf` | Type mismatch in Rust |
| `for _, x := range []int{1,2,3}` | Inline literal not supported |
| `[]interface{}` | Not supported |
| `map[K]V` | Maps not implemented |
| `chan T` | Channels not implemented |
| `go func()` | Goroutines not implemented |
| `select` | Select statement not implemented |
| `defer` | Defer not implemented |
| `panic/recover` | Not implemented |
| type switch | `switch x.(type)` not supported |
| `interface{}` | Interfaces not implemented |
| `*T` (pointers) | Pointers not implemented |
| `make()` | Use slice literals instead |
| `new()` | Use struct literals instead |
| `copy()` | Not implemented |
| `delete()` | Not implemented (no maps) |
| `cap()` | Slice capacity not implemented |
| named return values | `func f() (x int)` not supported |
| method receivers | `func (t T) Method()` not supported |
| embedded structs | Anonymous struct fields not supported |
| struct tags | `json:"name"` tags not supported |
| variadic functions | `func(args ...int)` not supported |
| blank imports | `import _ "pkg"` not supported |
| init functions | `func init()` not supported |
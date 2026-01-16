# C++ Backend Code Generation

This document describes the translation strategy from Go to C++.

## Overview

The C++ backend generates modern C++17 code that preserves Go's semantics while leveraging C++ standard library containers and idioms. The generated code prioritizes readability and correctness over performance optimization.

## Type System

### Primitive Types

Go's fixed-width integer types map directly to C++ standard integer types via type aliases. This ensures consistent size and behavior across platforms.

| Go | C++ | Rationale |
|----|-----|-----------|
| `int8` | `int8_t` | Exact 8-bit signed integer |
| `int16` | `int16_t` | Exact 16-bit signed integer |
| `int32` | `int32_t` | Exact 32-bit signed integer |
| `int64` | `int64_t` | Exact 64-bit signed integer |
| `int` | `int` | Platform-dependent, matches Go behavior |
| `string` | `std::string` | Dynamic string with similar semantics |
| `bool` | `bool` | Direct mapping |

### Slices

Go slices are translated to `std::vector<T>`. While Go slices are reference types with a backing array, C++ vectors are value types. To preserve Go's pass-by-value-of-reference semantics, the transpiler generates code that copies vectors when needed.

```go
// Go: slice declaration
var a []int
```
```cpp
// C++: vector declaration
std::vector<int> a;
```

The `append` function is implemented as a template that returns a new vector, mimicking Go's append behavior where the result must be assigned back.

### Structs

Go structs map directly to C++ structs. Both languages use value semantics for structs by default.

```go
type Point struct {
    x, y int
}
```
```cpp
struct Point {
    int x;
    int y;
};
```

### Function Types

Go function types are translated to `std::function<>`, which provides type-erased callable wrappers. This allows storing different callable objects (lambdas, function pointers) with the same signature.

```go
var f func(int, int)
```
```cpp
std::function<void(int, int)> f;
```

### Interface Types

Go's `interface{}` (empty interface) maps to `std::any`, which can hold any copyable type and provides runtime type checking.

## Variable Declarations

### Explicit Declarations

Go's `var` declarations with explicit types translate to C++ variable declarations with the corresponding type. Variables are zero-initialized in Go, which is preserved in C++.

### Short Declarations

Go's `:=` operator infers types from the right-hand side. C++ uses `auto` for the same purpose, providing similar type inference behavior.

```go
x := 42          // Go infers int
name := "hello"  // Go infers string
```
```cpp
auto x = 42;           // C++ infers int
auto name = "hello";   // C++ infers const char*, needs std::string wrapper
```

## Functions

### Basic Functions

Go functions translate directly to C++ functions. The syntax differs but semantics are preserved.

### Multiple Return Values

Go supports multiple return values natively. C++ achieves this using `std::tuple<>` for the return type and `std::tie()` for unpacking.

```go
func divide(a, b int) (int, int) {
    return a / b, a % b
}
quotient, remainder := divide(10, 3)
```
```cpp
std::tuple<int, int> divide(int a, int b) {
    return std::make_tuple(a / b, a % b);
}
int quotient, remainder;
std::tie(quotient, remainder) = divide(10, 3);
```

### Closures

Go closures (anonymous functions) translate to C++ lambdas. The capture mode `[&]` captures all referenced variables by reference, which approximates Go's closure behavior for local variables.

```go
x := 10
f := func() { fmt.Println(x) }
```
```cpp
auto x = 10;
auto f = [&]() { println(x); };
```

## Control Flow

### Conditionals

Go's `if` statements translate directly to C++ `if` statements. The only difference is that C++ requires parentheses around the condition.

### Loops

**C-style for loops** translate directly with `auto` for the loop variable:
```go
for i := 0; i < 10; i++ { }
```
```cpp
for (auto i = 0; i < 10; i++) { }
```

**Range-based for loops** translate to C++ range-based for:
```go
for _, x := range items { }
```
```cpp
for (auto x : items) { }
```

**While-style loops** (for with only condition) translate to C++ `while`:
```go
for condition { }
```
```cpp
while (condition) { }
```

### Switch Statements

Go's `switch` translates to C++ `switch`. Note that Go's switch doesn't fall through by default, so `break` statements are added to each case in C++.

## Built-in Functions

### len()

Go's `len()` for slices translates to `std::size()`, which works uniformly on all standard containers.

### append()

Go's `append()` is implemented as a template function that creates a new vector with the appended element(s), preserving Go's semantics where append may return a new backing array.

### fmt.Println / fmt.Printf

These translate to custom `println()` and `printf()` functions that wrap `std::cout` for consistent output behavior.

## Forward Declarations

The C++ backend generates forward declarations for all functions before their definitions. This is necessary because C++ requires functions to be declared before use, unlike Go which allows any order.

## Runtime Support

The generated code includes a runtime header with:
- Type aliases for Go integer types
- `append()` template function for slice operations
- `println()` and `printf()` wrappers for output
- `string_format()` for sprintf-like formatting

## Limitations

- **Goroutines**: Not supported (no concurrency translation)
- **Channels**: Not supported
- **Defer**: Not supported
- **Interfaces**: Only `interface{}` via `std::any`
- **Methods**: Receiver functions not supported
- **Pointers**: Limited support

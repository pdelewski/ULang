# Rust Backend Code Generation

This document describes the translation strategy from Go to Rust.

## Overview

The Rust backend generates safe Rust code that respects Rust's ownership and borrowing rules. This presents unique challenges since Go's garbage-collected runtime allows implicit sharing and mutation, while Rust requires explicit ownership management. The transpiler uses strategic cloning and reference counting to bridge this semantic gap.

## Type System

### Primitive Types

Go's integer types map directly to Rust's primitive types. Rust provides exact equivalents with similar naming conventions.

| Go | Rust | Rationale |
|----|------|-----------|
| `int8` | `i8` | 8-bit signed integer |
| `int16` | `i16` | 16-bit signed integer |
| `int32` | `i32` | 32-bit signed integer |
| `int64` | `i64` | 64-bit signed integer |
| `uint8` | `u8` | 8-bit unsigned integer |
| `uint16` | `u16` | 16-bit unsigned integer |
| `uint32` | `u32` | 32-bit unsigned integer |
| `uint64` | `u64` | 64-bit unsigned integer |
| `int` | `i32` | Default to 32-bit (platform-independent) |
| `string` | `String` | Owned, growable string |
| `bool` | `bool` | Boolean type |

### Slices

Go slices translate to `Vec<T>`. This is one of the most significant semantic differences:

- **Go slices** are reference types pointing to a backing array. Multiple slices can share the same backing array, and modifications through one slice are visible through others.
- **Rust Vec** is an owned collection. To preserve Go's semantics where functions can observe modifications, the transpiler uses `.clone()` to create copies when passing vectors.

```go
// Go: slices share backing array
a := []int{1, 2, 3}
b := a  // b and a share same backing
```
```rust
// Rust: explicit cloning needed
let mut a: Vec<i32> = vec![1, 2, 3];
let mut b = a.clone();  // Explicit copy
```

The `append` function returns a new vector, matching Go's behavior where append may allocate a new backing array.

### Structs

Go structs translate to Rust structs with automatic derive macros for common traits:

```go
type Point struct {
    X, Y int
}
```
```rust
#[derive(Default, Clone, Debug)]
pub struct Point {
    pub x: i32,
    pub y: i32,
}
```

The derived traits serve specific purposes:
- `Default`: Enables `..Default::default()` for partial initialization
- `Clone`: Required for the cloning strategy used throughout
- `Debug`: Allows debug printing
- `Copy`: Added for structs with only primitive fields (enables implicit copying)

### Function Types

Go function types present a challenge in Rust due to the borrow checker. The transpiler uses `Rc<dyn Fn(...)>` (reference-counted trait objects) to allow function values to be cloned and stored.

```go
var handlers []func(int)
```
```rust
let mut handlers: Vec<Rc<dyn Fn(i32)>>;
```

Using `Rc` (reference counting) instead of `Box` allows the function values to be cloned, which is necessary when they're stored in collections or passed around.

### Interface Types

Go's `interface{}` (empty interface) translates to `Box<dyn Any>`, which can hold any type that implements the `Any` trait (most types).

## Variable Declarations

### Mutability

A key difference from Go: all variables are declared as `let mut` (mutable) by default. This is because Go allows reassignment of any variable, and determining immutability would require additional analysis.

```go
x := 10
x = 20  // Reassignment allowed in Go
```
```rust
let mut x = 10;
x = 20;  // Requires mut in Rust
```

### Explicit Type Declarations

Go's `var` declarations with explicit types require type annotations in Rust, with zero-value initialization:

```go
var count int16
```
```rust
let mut count: i16 = 0;
```

### Short Declarations with Type Inference

Go's `:=` allows type inference, which Rust's `let` also supports:

```go
count := 42
name := "hello"
```
```rust
let mut count = 42;
let mut name = "hello".to_string();
```

Note that string literals require `.to_string()` to convert from `&str` to `String`.

### Empty Slice Declarations

Empty slices require explicit type annotations because Rust cannot infer the element type from `Vec::new()`:

```go
a := []int8{}
```
```rust
let mut a: Vec<i8> = Vec::new();
```

## Functions

### Basic Functions

Go functions translate to Rust functions with similar syntax:

```go
func add(a, b int) int {
    return a + b
}
```
```rust
fn add(a: i32, b: i32) -> i32 {
    return a + b;
}
```

### Multiple Return Values

Go's multiple return values translate to Rust tuples:

```go
func divmod(a, b int) (int, int) {
    return a / b, a % b
}
q, r := divmod(10, 3)
```
```rust
fn divmod(a: i32, b: i32) -> (i32, i32) {
    return (a / b, a % b);
}
let (mut q, mut r) = divmod(10, 3);
```

### Closures

Go closures translate to Rust closures wrapped in `Rc::new()`:

```go
x := []func(int, int){
    func(a, b int) { fmt.Println(a + b) },
}
```
```rust
let mut x: Vec<Rc<dyn Fn(i32, i32)>> = vec![
    Rc::new(|a: i32, b: i32| {
        println(a + b);
    })
];
```

### Local Closure Inlining

When a closure is assigned to a local variable and then called (a common pattern in Go), the Rust backend inlines the closure body at the call site. This avoids borrow checker issues where a closure captures mutable variables.

```go
// Go pattern
addItem := func(item Item) {
    items = append(items, item)
}
addItem(newItem)
```
```rust
// Rust: body inlined at call site
{
    items = append(&items.clone(), newItem);
}
```

## Control Flow

### Conditionals

Go's `if` statements translate directly. Rust requires the condition to be a boolean expression (no implicit truthiness).

### Loops

**C-style for loops** translate to Rust range expressions:
```go
for i := 0; i < 10; i++ { }
```
```rust
for i in 0..10 { }
```

**Range-based for loops** require cloning the collection to avoid borrowing issues:
```go
for _, item := range items { }
```
```rust
for item in items.clone() { }
```

**While-style loops** translate to Rust `while`:
```go
for condition { }
```
```rust
while condition { }
```

### Switch Statements

Go's `switch` translates to Rust's `match` expression:

```go
switch x {
case 1:
    // handle 1
case 2:
    // handle 2
default:
    // handle default
}
```
```rust
match x {
    1 => {
        // handle 1
    }
    2 => {
        // handle 2
    }
    _ => {
        // handle default
    }
}
```

## Index and Slice Operations

### Array/Slice Indexing

Rust requires `usize` for indexing, so integer indices are cast:

```go
item := items[i]
items[0] = value
```
```rust
let item = items[i as usize].clone();
items[0 as usize] = value;
```

Note: `.clone()` is added on the right-hand side but NOT on the left-hand side of assignments (that would be an error).

### Slice Expressions

Go's slice expressions translate to Rust range syntax with `.to_vec()`:

```go
sub := items[1:]      // from index 1 to end
sub := items[:5]      // from start to index 5
sub := items[1:5]     // from index 1 to 5
```
```rust
let sub = items[1 as usize..].to_vec();
let sub = items[..5 as usize].to_vec();
let sub = items[1 as usize..5 as usize].to_vec();
```

### Length

Go's `len()` translates to a helper function that returns `i32`:

```go
n := len(items)
```
```rust
let n = len(&items.clone());
```

## Built-in Functions

### fmt.Println

Go's `fmt.Println()` translates to custom `println()` functions:

```go
fmt.Println(value)
fmt.Println()
```
```rust
println(value);
println0();  // No arguments version
```

### fmt.Printf

Go's `fmt.Printf()` translates to numbered `printf` variants based on argument count. Format strings are converted from C-style (`%d`) to Rust-style (`{}`).

```go
fmt.Printf("Value: %d\n", x)
fmt.Printf("%c", b)  // Print byte as char
```
```rust
printf2("Value: %d\n".to_string(), x);
printc(b);  // Special function for %c
```

The `%c` format specifier gets special treatment with a dedicated `printc()` function that handles `i8` to `char` conversion.

### fmt.Sprintf

Similarly, `fmt.Sprintf()` translates to `string_format2()`:

```go
s := fmt.Sprintf("Count: %d", n)
s := fmt.Sprintf("%c", b)
```
```rust
let s = string_format2("Count: %d", n);
let s = byte_to_char(b);  // Special case for %c
```

## Ownership and Cloning Strategy

The Rust backend uses a "clone liberally" strategy to ensure correctness:

1. **Passing to functions**: Collections are cloned and passed as references: `&items.clone()`
2. **Struct field access**: Fields are cloned when read: `node.field.clone()`
3. **Loop iteration**: Collections are cloned: `for item in items.clone()`
4. **Index expressions on RHS**: Elements are cloned: `items[i as usize].clone()`

This approach sacrifices some performance for correctness, as it avoids complex borrow checker issues that would arise from trying to share mutable state.

## Keyword Escaping

Go identifiers that are Rust keywords are escaped with the `r#` prefix:

| Go identifier | Rust escaped |
|--------------|--------------|
| `type` | `r#type` |
| `match` | `r#match` |
| `mod` | `r#mod` |
| `ref` | `r#ref` |

Exception: `true` and `false` are not escaped as they have the same meaning in both languages.

## Code Formatting

The Rust backend uses `rustfmt` for code formatting, which properly handles Rust-specific syntax like type annotations, closures, and match expressions.

## Runtime Support

The generated code includes runtime helper functions:
- `println()`, `println0()`: Print with newline
- `printf()`, `printf2()`, etc.: Formatted printing
- `printc()`: Print byte as character
- `byte_to_char()`: Convert byte to String
- `append()`: Go-style append returning new Vec
- `len()`: Length function returning i32
- `string_format2()`: Sprintf equivalent

## Limitations

- **Goroutines**: Not supported
- **Channels**: Not supported
- **Defer**: Not supported
- **Interfaces**: Only `interface{}` via `Box<dyn Any>`
- **Methods**: Receiver functions not supported
- **Pointers**: Not supported
- **Maps**: Not supported
- **Performance**: Liberal cloning may impact performance

# Rust Backend Implementation Rules

This document describes the main rules and implementation decisions for the Go-to-Rust transpilation backend.

## 1. Closure Handling

### Local Closure Inlining

When a closure is assigned to a local variable and then called, the closure body is inlined at the call site instead of creating a separate closure. This avoids Rust borrow checker conflicts where a closure captures mutable variables.

**Implementation details:**
- Track `localClosureBodyTokens` map to store closure bodies by name
- Use `inLocalClosureBody` flag to track when processing closure body
- Store `currentClosureName` for the closure being processed
- The closure assignment is removed from output, body tokens are stored
- At call sites, stored body tokens are inlined wrapped in `{}`

**Example transformation:**
```go
// Go
addToken := func(t Token) { tokens = append(tokens, t) }
addToken(myToken)
```
```rust
// Rust (inlined)
{ tokens = append(&tokens.clone(), myToken); }
```

### Closure Wrapper Type

All closures use `Rc::new()` wrapper (not `Box::new()`) to allow cloning.

**Rule:** Function types use `Rc<dyn Fn(...)>`

**Location:** `PreVisitFuncLit` in `rust_emitter.go`

## 2. Assignment Handling

### LHS vs RHS Context

Track `inAssignLhs` flag to distinguish left-hand side from right-hand side of assignments.

**Rule:** On LHS of assignments, do NOT add `.clone()` to indexed expressions.

**Correct:**
```rust
l.nodes[i as usize] = tmp;
```

**Incorrect (causes E0070):**
```rust
l.nodes[i as usize].clone() = tmp;  // Error: invalid left-hand side
```

### Clone on RHS

Struct field access on RHS gets `.clone()` for Copy types. `PostVisitIndexExprIndex` checks `!re.inAssignLhs` before adding `.clone()`.

## 3. Type Annotations

### Empty Slice Initialization

`[]Type{}` in Go becomes `Vec::new()` in Rust. Rust cannot infer type for empty `Vec::new()` without context.

**Rule:** For variable declarations, add explicit type annotation.

**Transformation:**
```go
// Go
a := []int8{}
```
```rust
// Rust
let mut a: Vec<i8> = Vec::new();
```

**Context-dependent behavior:**
- In struct field init (`inKeyValueExpr`), return statements (`inReturnStmt`), or field assignment (`inFieldAssign`): just use `Vec::new()`
- In variable declarations: add `: Vec<Type> =` before `Vec::new()`

Non-empty slices use `vec!{...}` macro which can infer types.

## 4. Keyword Handling

### Rust Keyword Escaping

Go identifiers that are Rust keywords get `r#` prefix.

**Examples:**
- `type` → `r#type`
- `match` → `r#match`
- `mod` → `r#mod`

**Exception:** `true` and `false` are NOT escaped (they're boolean literals in both languages).

**Location:** `escapeRustKeyword()` function in `rust_emitter.go`

## 5. Printf/Sprintf Handling

### Format String Processing

The `printf2`, `printf3`, etc. functions convert C-style format specifiers to Rust format:
- `%d` → `{}`
- `%s` → `{}`
- `%v` → `{}`

The format string is actually processed with value substitution.

### Character Printing (%c)

Special handling for printing bytes as characters:

| Go Code | Rust Code |
|---------|-----------|
| `Printf("%c", byte)` | `printc(byte)` |
| `Sprintf("%c", byte)` | `byte_to_char(byte)` |

**Runtime functions:**
```rust
pub fn printc(b: i8) {
    print!("{}", b as u8 as char);
}

pub fn byte_to_char(b: i8) -> String {
    (b as u8 as char).to_string()
}
```

**Location:** Special case detection in `PostVisitCallExprArgs`

## 6. Type Mappings

| Go Type | Rust Type |
|---------|-----------|
| `int8` | `i8` |
| `int16` | `i16` |
| `int32` | `i32` |
| `int64` | `i64` |
| `uint8` | `u8` |
| `uint16` | `u16` |
| `uint32` | `u32` |
| `uint64` | `u64` |
| `int` | `i32` |
| `string` | `String` |
| `[]T` | `Vec<T>` |
| `func(...)` | `Rc<dyn Fn(...)>` |
| `interface{}` | `Box<dyn Any>` |

## 7. Slice Operations

### Indexing

Rust requires `usize` for array/slice indexing.

**Rule:** `a[i]` → `a[i as usize]`

### Slicing

| Go | Rust |
|----|------|
| `a[i:]` | `a[i as usize..].to_vec()` |
| `a[:j]` | `a[..j as usize].to_vec()` |
| `a[i:j]` | `a[i as usize..j as usize].to_vec()` |

### Length

| Go | Rust |
|----|------|
| `len(slice)` | `len(&slice.clone())` (returns `i32`) |
| `len(string)` | `string.len() as i32` |

## 8. Struct Handling

### Default Initialization

Structs automatically derive necessary traits:
```rust
#[derive(Default, Clone, Debug)]
pub struct MyStruct { ... }
```

Partial initialization uses `..Default::default()`:
```rust
MyStruct { field: val, ..Default::default() }
```

### Copy vs Clone

- Structs with only primitive fields: derive `Copy`
- Structs with `Vec`, `String`, or function fields: only derive `Clone`

## 9. For Loop Handling

### Range-based For

| Go | Rust |
|----|------|
| `for i := 0; i < n; i++` | `for i in 0..n` |
| `for _, v := range slice` | `for v in slice.clone()` |
| `for i, v := range slice` | `for (i, v) in slice.clone().iter().enumerate()` |

## 10. String Handling

### String Literals

- `"text"` → `"text".to_string()` when used as `String` type
- Raw strings: `` `text` `` → `r#"text"#`

### String Concatenation

**Rule:** `str += other` → `str += &other`

Rust expects `&str` on RHS of `+=` for String.

## 11. Control Flow

### If Statements

Parentheses around conditions are preserved (generates warnings but valid Rust).

Boolean negation works the same: `!cond`

### Return Statements

Multiple return values use tuples:
```go
// Go
return a, b
```
```rust
// Rust
return (a, b)
```

Function signature: `fn foo() -> (T1, T2)`

## 12. Token Stream Manipulation

The Rust emitter uses a token-based approach for code generation and transformation.

### Markers

- `@PreVisitXxx` / `@PostVisitXxx` markers inserted during AST traversal
- Used to locate positions for token rewriting

### Key Functions

| Function | Purpose |
|----------|---------|
| `RewriteTokensBetween()` | Replace token range with new tokens |
| `ExtractTokensBetween()` | Get tokens for analysis |
| `SearchPointerIndexReverse()` | Find markers in token stream |

This enables complex transformations like:
- Closure inlining
- Type annotation insertion
- Format specifier handling

## 13. Runtime Library

The Rust backend includes a runtime library embedded in generated code with:

- `println`, `println0` - printing functions
- `printf`, `printf2`, `printf3`, etc. - formatted printing
- `printc` - print byte as character
- `byte_to_char` - convert byte to String
- `append`, `append_many` - Go-style slice append
- `string_format`, `string_format2` - Sprintf equivalents
- `len` - generic length function

Type aliases for Go compatibility:
```rust
type Int8 = i8;
type Int16 = i16;
// ... etc
```

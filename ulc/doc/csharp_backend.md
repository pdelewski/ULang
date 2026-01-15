# C# Backend Code Generation

This document describes the translation strategy from Go to C#.

## Overview

The C# backend generates modern C# code targeting .NET Core/.NET 5+. It leverages C#'s strong type system, value tuples for multiple returns, and LINQ-style operations where appropriate. The generated code is wrapped in a namespace and static class structure for proper encapsulation.

## Type System

### Primitive Types

Go's integer types map to C#'s built-in numeric types. C# provides exact equivalents for all fixed-width integer types.

| Go | C# | Rationale |
|----|-----|-----------|
| `int8` | `sbyte` | 8-bit signed integer (-128 to 127) |
| `int16` | `short` | 16-bit signed integer |
| `int32` | `int` | 32-bit signed integer |
| `int64` | `long` | 64-bit signed integer |
| `uint8` | `byte` | 8-bit unsigned integer |
| `uint16` | `ushort` | 16-bit unsigned integer |
| `uint32` | `uint` | 32-bit unsigned integer |
| `uint64` | `ulong` | 64-bit unsigned integer |
| `string` | `string` | Immutable string type |
| `bool` | `bool` | Boolean type |

Note that `sbyte` in C# is the signed 8-bit type, while `byte` is unsigned—the opposite naming convention from what might be expected.

### Slices

Go slices are translated to `List<T>`. While Go slices have reference semantics with a backing array, C# `List<T>` is a reference type that behaves similarly. The `Append` extension method is provided to mimic Go's append semantics.

```go
// Go
var items []int
items = append(items, 42)
```
```csharp
// C#
List<int> items = new List<int>();
items = items.Append(42);
```

The key difference is that Go's append may or may not create a new backing array, while the C# implementation always creates a new list to maintain functional semantics.

### Structs

Go structs translate to C# structs. Both are value types with similar semantics. C# structs are declared with `public` fields to match Go's exported field behavior.

```go
type Point struct {
    X, Y int
}
```
```csharp
public struct Point {
    public int X;
    public int Y;
}
```

### Function Types

Go function types translate to C# delegate types. For functions that don't return a value, `Action<>` is used. For functions with return values, `Func<>` would be used.

```go
var handler func(int, string)
```
```csharp
Action<int, string> handler;
```

### Interface Types

Go's empty interface `interface{}` translates to C#'s `object` type, which is the base type of all types in C# and can hold any value.

## Variable Declarations

### Explicit Declarations

Go's `var` declarations translate to C# variable declarations. C# requires initialization, so `default` is used for zero-initialization.

```go
var count int16
```
```csharp
short count = default;
```

### Short Declarations

Go's `:=` operator maps to C#'s `var` keyword, which provides type inference.

```go
name := "hello"
count := 42
```
```csharp
var name = "hello";
var count = 42;
```

### Type Casting

C# is stricter about implicit numeric conversions than Go. When assigning to smaller integer types like `sbyte` or `short`, explicit casts are required.

```go
var a int8
a = 5
a = a + 1
```
```csharp
sbyte a = default;
a = (sbyte)5;
a = (sbyte)(a + 1);
```

## Functions

### Basic Functions

Go functions translate to C# static methods. All generated methods are `public static` to allow access from anywhere.

```go
func Add(a, b int) int {
    return a + b
}
```
```csharp
public static int Add(int a, int b) {
    return a + b;
}
```

### Multiple Return Values

Go's multiple return values translate to C# value tuples, introduced in C# 7.0. This provides a clean syntax for both returning and unpacking multiple values.

```go
func divmod(a, b int) (int, int) {
    return a / b, a % b
}
q, r := divmod(10, 3)
```
```csharp
public static (int, int) divmod(int a, int b) {
    return (a / b, a % b);
}
(var q, var r) = divmod(10, 3);
```

### Closures

Go closures translate to C# lambda expressions. C# lambdas capture variables by reference for reference types and by value for value types, which closely matches Go's closure semantics.

```go
multiplier := 2
double := func(x int) int { return x * multiplier }
```
```csharp
var multiplier = 2;
Func<int, int> double = (int x) => x * multiplier;
```

## Control Flow

### Conditionals

Go's `if` statements translate directly to C# `if` statements. The syntax is nearly identical.

### Loops

**C-style for loops** translate directly:
```go
for i := 0; i < 10; i++ { }
```
```csharp
for (var i = 0; i < 10; i++) { }
```

**Range-based for loops** translate to C#'s `foreach`:
```go
for _, item := range items { }
```
```csharp
foreach (var item in items) { }
```

**While-style loops** translate to C# `while`:
```go
for condition { }
```
```csharp
while (condition) { }
```

### Switch Statements

Go's `switch` translates to C# `switch`. Both languages support switching on values, but Go's switch doesn't fall through by default while C#'s does. The transpiler adds `break` statements to prevent fallthrough.

## Slice Operations

### Length

Go's `len()` function translates to a `SliceBuiltins.Length()` helper method that works with both collections and strings.

```go
n := len(items)
```
```csharp
var n = SliceBuiltins.Length(items);
```

### Slicing

Go's slice expressions translate to C# range syntax (C# 8.0+):

```go
sub := items[1:]      // from index 1 to end
sub := items[:5]      // from start to index 5
sub := items[1:5]     // from index 1 to 5
```
```csharp
var sub = items[1..];     // Range syntax
var sub = items[..5];
var sub = items[1..5];
```

### Append

Go's `append()` is implemented as an extension method on `List<T>`:

```go
items = append(items, newItem)
```
```csharp
items = items.Append(newItem);
```

## Built-in Functions

### fmt.Println

Go's `fmt.Println()` translates to `Console.WriteLine()`:

```go
fmt.Println("Hello")
fmt.Println(42)
```
```csharp
Console.WriteLine("Hello");
Console.WriteLine(42);
```

### fmt.Printf

Go's `fmt.Printf()` translates to a custom `Formatter.Printf()` method that converts Go format specifiers to C# format strings:

- `%d` → `{0}`, `{1}`, etc.
- `%s` → `{0}`, `{1}`, etc.
- `%c` → character conversion from `sbyte`

```go
fmt.Printf("Value: %d\n", x)
```
```csharp
Formatter.Printf(@"Value: %d\n", x);
```

### fmt.Sprintf

Similarly, `fmt.Sprintf()` translates to `Formatter.Sprintf()`:

```go
s := fmt.Sprintf("Count: %d", n)
```
```csharp
var s = Formatter.Sprintf("Count: %d", n);
```

## Code Structure

The generated C# code follows this structure:

```csharp
using System;
using System.Collections.Generic;

// Runtime helpers (SliceBuiltins, Formatter)

namespace MainClass {
    public struct Api {
        // All generated code here
        public static void Main() { }
    }
}
```

All code is placed in a static struct to provide a namespace for the functions without requiring instantiation.

## Runtime Support

The generated code includes runtime helper classes:

- **SliceBuiltins**: Extension methods for `Append()` and `Length()` operations
- **Formatter**: `Printf()` and `Sprintf()` implementations with Go-style format string conversion

## Limitations

- **Goroutines**: Not supported
- **Channels**: Not supported
- **Defer**: Not supported
- **Interfaces**: Only `interface{}` via `object`
- **Methods**: Receiver functions not supported
- **Pointers**: Not supported
- **Maps**: Not supported

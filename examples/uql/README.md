# UQL (Universal Query Language)

A SQL-like query language parser and AST implementation in goany (Go subset that transpiles to C++/C#/Rust).

## Overview

UQL demonstrates building a complete language frontend including:
- **Lexer** - Tokenizes input strings into tokens
- **Parser** - Builds an Abstract Syntax Tree (AST) from tokens
- **AST Walker** - Visitor pattern for traversing the AST

## Query Syntax

UQL uses a pipe-based syntax where intermediate results are assigned to variables:

```sql
t1 = from table1;
t2 = where t1.field1 > 10 && t1.field2 < 20;
t3 = select t2.field1;
```

### Supported Statements

- `from <table>` - Select data from a table
- `where <condition>` - Filter rows based on conditions
- `select <fields>` - Project specific fields

### Supported Operators

- Comparison: `>`, `<`, `>=`, `<=`, `==`, `!=`
- Logical: `&&`, `||`

## Building

```bash
# From the cmd directory
./goany --source=../examples/uql --output=build/uql --link-runtime=../runtime

# Compile C++
cd build/uql && make

# Or compile C#
cd build/uql && dotnet build

# Or compile Rust
cd build/uql && cargo build
```

## Running

```bash
# C++
./uql

# C#
dotnet run

# Rust
cargo run
```

## Structure

```
uql/
├── main.go      # Main program with visitor implementation
├── go.mod       # Go module file
├── lexer/       # Lexical analysis
│   └── lexer.go # Token definitions and tokenizer
├── parser/      # Syntax analysis
│   └── parser.go # Parser implementation
└── ast/         # Abstract Syntax Tree
    └── ast.go   # AST node types and walker
```

## AST Visitor Pattern

The example demonstrates using the visitor pattern to traverse the AST:

```go
visitor := ast.Visitor{
    PreVisitFrom: func(state any, expr ast.From) any {
        // Called before visiting From node children
        return state
    },
    PostVisitFrom: func(state any, from ast.From) any {
        // Called after visiting From node children
        return state
    },
    // ... similar for Where, Select, LogicalExpr
}
```

## Example Output

```
1
From:
  t1
  table1
2
Where:
  t2
  t2
  t1.field1 > 10
  t1.field2 < 20
3
Select:
  t3
  t2.field1
```

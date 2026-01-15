# ULang

A Go language transpiler that generates portable code for multiple target platforms. Write your code once in Go and transpile it to C++, C#, or Rust.

## Overview

This project provides a foundation for writing portable libraries in Go that can be transpiled to multiple backend languages. Currently supported backends:

- **C++** - generates `.cpp` files
- **C#** - generates `.cs` files
- **Rust** - generates `.rs` files

## Building

To build the compiler:

```bash
cd ulc
make
```

### Make Targets

| Target | Description |
|--------|-------------|
| `make` | Build the project (default) |
| `make build` | Generate code, build astyle, and build ULC binary |
| `make clean` | Clean all build artifacts |
| `make rebuild` | Clean and rebuild everything |
| `make dev` | Development build (with debug info) |
| `make prod` | Production build (optimized) |
| `make test` | Run tests |
| `make help` | Show all available targets |

## Usage

```bash
./ulc -source=[directory] -output=[name] -backend=[backend]
```

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-source` | Source directory containing Go files | (required) |
| `-output` | Output file name (without extension) | (required) |
| `-backend` | Backend(s) to use: `all`, `cpp`, `cs`, `rust` | `all` |

The `-backend` flag accepts comma-separated values for multiple backends.

### Examples

Transpile to all backends:
```bash
./ulc -source=./libs/uql -output=uql
```

Transpile to Rust only:
```bash
./ulc -source=./libs/uql -output=uql -backend=rust
```

Transpile to C# and Rust:
```bash
./ulc -source=./libs/uql -output=uql -backend=cs,rust
```

## Supported Features

### Types
- Primitive types: `int8`, `int16`, `int32`, `int64`, `uint8`, `uint16`, `uint32`, `uint64`
- `string`
- Slices: `[]T`
- Structs
- Function types
- `interface{}`

### Language Constructs
- Variable declarations and assignments
- Functions with multiple return values
- Structs with methods
- For loops (C-style and range-based)
- If/else statements
- Switch statements

### Limitations

Some Go features may not be fully supported due to differences in target platforms. See `ulc/doc/rust_backend_rules.md` for detailed implementation notes on the Rust backend.

## Project Structure

```
ULang/
├── ulc/                    # Compiler source code
│   ├── main.go            # Entry point
│   ├── rust_emitter.go    # Rust backend
│   ├── csharp_emitter.go  # C# backend
│   ├── cpp_emitter.go     # C++ backend
│   └── doc/               # Documentation
├── libs/                   # Example libraries
│   ├── uql/               # SQL query parser
│   └── contlib/           # Container library
└── tests/                  # Test cases
```

## License

This project is a personal experiment for exploring language transpilation concepts.
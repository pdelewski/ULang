# goany

![goany banner](docs/goany-banner.png)

A Go language transpiler that generates portable code for multiple target platforms. Write your code once in Go and transpile it to C++, C#, Rust, or JavaScript.

[**▶ Try the C64 Emulator Demo WIP (runs in browser)**](https://pdelewski.github.io/goany/demos/c64.html)

[**▶ MOS 6502 Text Sprites Demo**](https://pdelewski.github.io/goany/demos/text.html) | [**▶ MOS 6502 Basic Graphics Shapes Demo**](https://pdelewski.github.io/goany/demos/graphic.html)

## Overview

This project provides a foundation for writing portable libraries in Go that can be transpiled to multiple backend languages. Currently supported backends:

- **C++** - generates `.cpp` files
- **C#** (.NET) - generates `.cs` files
- **Rust** - generates `.rs` files
- **JavaScript** - generates `.js` files (runs in browser with Canvas API)

## Project Goals

The main aim of goany is to provide a tool for writing **portable applications and libraries** that work across different programming languages and platforms.

### Key Objectives

1. **Cross-language portability** - Write code once and transpile it to C++, Rust, C#/.NET, and JavaScript, enabling code reuse across different ecosystems and platforms.

2. **Near 1-to-1 translation** - The generated code maintains almost direct correspondence to the original source, making it readable, debuggable, and easy to understand.

3. **Focused feature set** - goany intentionally does not support all Go language features. The goal is to support the subset needed to write reusable libraries across languages, not to be a complete Go transpiler.

### Important Note

> **All valid goany programs are valid Go programs, but not vice-versa.**

This means:
- You can compile and run any goany program with the standard Go toolchain
- Not every Go program can be transpiled by goany (only the supported subset)
- goany source files are regular `.go` files that follow Go syntax

This design allows you to develop and test your code using Go's excellent tooling, then transpile to other languages when ready for deployment.

## Building

To build the compiler:

```bash
cd cmd
make
```

### Make Targets

| Target | Description |
|--------|-------------|
| `make` | Build the project (default) |
| `make build` | Generate code, build astyle, and build goany binary |
| `make clean` | Clean all build artifacts |
| `make rebuild` | Clean and rebuild everything |
| `make dev` | Development build (with debug info) |
| `make prod` | Production build (optimized) |
| `make test` | Run tests |
| `make help` | Show all available targets |

## Usage

```bash
./goany -source=[directory] -output=[name] -backend=[backend]
```

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-source` | Source directory containing Go files | (required) |
| `-output` | Output file name (without extension) | (required) |
| `-backend` | Backend(s) to use: `all`, `cpp`, `cs`, `rust`, `js` | `all` |
| `-link-runtime` | Path to runtime for linking (generates build files with include paths) | (none) |
| `-graphics-runtime` | Graphics backend: `tigr`, `sdl2`, `none` | `tigr` |
| `-debug` | Enable debug output | `false` |

The `-backend` flag accepts comma-separated values for multiple backends.

The `-graphics-runtime` flag selects the graphics library:
- `tigr` - Bundled, header-only library (C++ only, no external dependencies)
- `sdl2` - SDL2 library (requires SDL2 installed, supports all backends)
- `none` - No graphics support (for CLI applications)

### Examples

Transpile to all backends:
```bash
./goany -source=../examples/uql -output=uql
```

Transpile to Rust only:
```bash
./goany -source=../examples/uql -output=uql -backend=rust
```

Transpile to C# and Rust:
```bash
./goany -source=../examples/uql -output=uql -backend=cs,rust
```

Transpile to JavaScript (runs in browser):
```bash
./goany -source=../examples/mos6502/cmd/c64 -output=c64 -backend=js -link-runtime=../runtime
```

Transpile graphics demo with tigr (default):
```bash
./goany -source=../examples/graphics-demo -output=./build/graphics-demo -backend=cpp -link-runtime=../runtime
```

Transpile graphics demo with SDL2 (all backends):
```bash
./goany -source=../examples/graphics-demo -output=./build/graphics-demo -backend=rust -link-runtime=../runtime -graphics-runtime=sdl2
```

Transpile CLI app without graphics:
```bash
./goany -source=../examples/uql -output=./build/uql -link-runtime=../runtime -graphics-runtime=none
```

## Supported Features

goany supports a subset of Go designed for cross-platform portability:

- **Types**: primitives (`int8`-`int64`, `uint8`-`uint64`, `float32`, `float64`, `bool`, `string`), slices, structs, function types
- **Constructs**: variables, functions (multiple returns), methods, for loops, if/else, switch
- **Graphics**: cross-platform 2D graphics library with tigr, SDL2, and Canvas backends

See [docs/supported-features.md](docs/supported-features.md) for detailed documentation, limitations, and known issues.

## Project Structure

| Directory | Description |
|-----------|-------------|
| `cmd/` | CLI entry point (`main.go`, `Makefile`) |
| `compiler/` | Backend emitters (`cpp_emitter.go`, `csharp_emitter.go`, `rust_emitter.go`, `js_emitter.go`) |
| `runtime/` | Runtime libraries (graphics for Canvas/SDL2/TIGR) |
| `examples/` | Example projects (`uql`, `contlib`, `graphics-demo`, `mos6502`) |
| `scripts/` | Utility scripts (`setup-deps.sh`) |
| `tests/` | Test cases |

## License

This project is licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for details.

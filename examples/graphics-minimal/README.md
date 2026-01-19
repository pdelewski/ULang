# Graphics Minimal Example

The simplest possible SDL2 graphics application in goany (Go subset that transpiles to C++/C#/Rust).

## Overview

This example demonstrates the minimal code needed to create and close an SDL2 window. It's useful as a starting point for graphics applications and to verify the graphics runtime is working correctly.

## Code

```go
package main

import "runtime/graphics"

func main() {
    w := graphics.CreateWindow("Minimal", 400, 300)
    graphics.CloseWindow(w)
}
```

## Prerequisites

- SDL2 library installed on your system
  - macOS: `brew install sdl2`
  - Ubuntu: `apt-get install libsdl2-dev`
  - Windows: Download from libsdl.org

## Building

```bash
# From the cmd directory
./goany --source=../examples/graphics-minimal --output=build/graphics-minimal --link-runtime=../runtime

# Compile C++
cd build/graphics-minimal && make

# Or compile C#
cd build/graphics-minimal && dotnet build

# Or compile Rust
cd build/graphics-minimal && cargo build
```

## Running

```bash
# C++
./graphics-minimal

# C#
dotnet run

# Rust
cargo run
```

## Expected Behavior

The program creates a 400x300 pixel window titled "Minimal" and immediately closes it. You may see the window flash briefly.

## See Also

- [graphics-demo](../graphics-demo/) - A more complete graphics example with shapes and an event loop

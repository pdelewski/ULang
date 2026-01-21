# goany Graphics Runtime

Cross-platform 2D graphics library for goany transpiled applications.

## Graphics Backends

Two backends are available, selected via `--graphics-runtime` flag:

| Backend | Flag | C++ | C# | Rust | Dependencies |
|---------|------|-----|-----|------|--------------|
| **tigr** | `--graphics-runtime=tigr` | ✅ | ✅ | ✅ | C compiler (bundled source) |
| **SDL2** | `--graphics-runtime=sdl2` | ✅ | ✅ | ✅ | SDL2 library |

**tigr** is the default for all backends - source is bundled with the transpiler and compiled at build time. Requires a C compiler (gcc/clang on macOS/Linux, MSVC on Windows).

**SDL2** is available for all backends when hardware acceleration or additional features are needed.

## Quick Start

### Option A: Using tigr (C++ only, no dependencies)

```bash
# Transpile with tigr (default)
./goany -source=./myapp -output=./build/myapp -backend=cpp -link-runtime=../runtime

# Build
cd build && make
```

### Option B: Using SDL2 (all backends)

#### 1. Install SDL2 Dependencies

Run the setup script:
```bash
./scripts/setup-deps.sh
```

Or install manually:

| Platform | Command |
|----------|---------|
| macOS | `brew install sdl2` |
| Ubuntu/Debian | `sudo apt install libsdl2-dev` |
| Fedora | `sudo dnf install SDL2-devel` |
| Arch | `sudo pacman -S sdl2` |
| Windows (MSYS2) | `pacman -S mingw-w64-x86_64-SDL2` |

#### 2. Transpile with SDL2

```bash
./goany -source=./myapp -output=./build/myapp -link-runtime=../runtime -graphics-runtime=sdl2
```

### 2. Write Your Code

```go
package main

import "myapp/graphics"

func main() {
    // Create window
    w := graphics.CreateWindow("My App", 800, 600)

    // Main loop
    running := true
    for running {
        w, running = graphics.PollEvents(w)

        graphics.Clear(w, graphics.Black())
        graphics.FillRect(w, graphics.NewRect(100, 100, 200, 150), graphics.Red())
        graphics.DrawCircle(w, 400, 300, 50, graphics.White())
        graphics.Present(w)
    }

    graphics.CloseWindow(w)
}
```

### 3. Transpile and Compile

```bash
# Transpile
./goany -source=./myapp -output=myapp

# Compile C++
g++ -std=c++17 myapp.cpp $(sdl2-config --cflags --libs)

# Or C#
dotnet add package SDL2-CS
dotnet build

# Or Rust
cargo build
```

## API Reference

### Types

```go
type Window struct { ... }  // Window handle
type Color struct { R, G, B, A uint8 }
type Rect struct { X, Y, Width, Height int32 }
```

### Window Management

| Function | Description |
|----------|-------------|
| `CreateWindow(title string, width int32, height int32) Window` | Create a new window |
| `CloseWindow(w Window)` | Close and destroy window |
| `PollEvents(w Window) (Window, bool)` | Process events, returns false on quit |
| `IsRunning(w Window) bool` | Check if window is still open |
| `GetWidth(w Window) int32` | Get window width |
| `GetHeight(w Window) int32` | Get window height |

### Rendering

| Function | Description |
|----------|-------------|
| `Clear(w Window, c Color)` | Clear screen with color |
| `Present(w Window)` | Display rendered frame |

### Drawing Primitives

| Function | Description |
|----------|-------------|
| `DrawRect(w Window, rect Rect, c Color)` | Draw rectangle outline |
| `FillRect(w Window, rect Rect, c Color)` | Draw filled rectangle |
| `DrawLine(w Window, x1, y1, x2, y2 int32, c Color)` | Draw line |
| `DrawPoint(w Window, x, y int32, c Color)` | Draw single pixel |
| `DrawCircle(w Window, cx, cy, r int32, c Color)` | Draw circle outline |
| `FillCircle(w Window, cx, cy, r int32, c Color)` | Draw filled circle |

### Helpers

| Function | Description |
|----------|-------------|
| `NewColor(r, g, b, a uint8) Color` | Create custom color |
| `NewRect(x, y, w, h int32) Rect` | Create rectangle |
| `Black()`, `White()`, `Red()`, `Green()`, `Blue()` | Predefined colors |

## Backend Compilation

### C++
```bash
g++ -std=c++17 output.cpp $(sdl2-config --cflags --libs)
```

### C#
```bash
dotnet add package SDL2-CS
dotnet build
```

### Rust
```toml
# Cargo.toml
[dependencies]
sdl2 = "0.36"
```

## Go Backend Selection

For Go execution, two backends are available:

```bash
# Default: tigr backend (no external dependencies)
go run .

# SDL2 backend (requires SDL2 installed)
go build -tags sdl2
```

## Directory Structure

```
runtime/graphics/
├── go.mod
├── graphics.go                    # Go API (tigr backend, default)
├── graphics_sdl2.go               # Go API (SDL2 backend, build tag: sdl2)
├── README.md
├── go/
│   ├── tigr/
│   │   ├── graphics_runtime_tigr.go  # Go CGO tigr bindings
│   │   ├── tigr.c                    # tigr library (bundled)
│   │   └── tigr.h                    # tigr header (bundled)
│   └── sdl2/
│       └── graphics_runtime_sdl2.go  # Go CGO SDL2 bindings
├── cpp/
│   ├── tigr.h                     # tigr library header (bundled)
│   ├── tigr.c                     # tigr library implementation (bundled)
│   ├── graphics_runtime_tigr.hpp  # C++ tigr backend
│   └── graphics_runtime_sdl2.hpp  # C++ SDL2 backend
├── csharp/
│   ├── GraphicsRuntimeSDL2.cs     # C# SDL2 implementation
│   └── GraphicsRuntimeTigr.cs     # C# tigr implementation
└── rust/
    ├── graphics_runtime_sdl2.rs   # Rust SDL2 implementation
    └── graphics_runtime_tigr.rs   # Rust tigr implementation
```

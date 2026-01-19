# Graphics Demo

A demonstration of SDL2 graphics capabilities in goany (Go subset that transpiles to C++/C#/Rust).

## Overview

This example shows how to create interactive graphics applications using the goany graphics runtime. It demonstrates:

- Creating a window with a main loop
- Event handling (window close)
- Drawing various shapes
- Screen clearing and presenting

## Features

The demo draws several shapes on screen:

- **Red filled rectangle** - Demonstrates `FillRect`
- **Green rectangle outline** - Demonstrates `DrawRect`
- **Blue filled circle** - Demonstrates `FillCircle`
- **White circle outline** - Demonstrates `DrawCircle`
- **Yellow crossing lines** - Demonstrates `DrawLine`
- **White dotted line** - Demonstrates `DrawPoint`

## Prerequisites

- SDL2 library installed on your system
  - macOS: `brew install sdl2`
  - Ubuntu: `apt-get install libsdl2-dev`
  - Windows: Download from libsdl.org

## Building

```bash
# From the cmd directory
./goany --source=../examples/graphics-demo --output=build/graphics-demo --link-runtime=../runtime

# Compile C++
cd build/graphics-demo && make

# Or compile C#
cd build/graphics-demo && dotnet build

# Or compile Rust
cd build/graphics-demo && cargo build
```

## Running

```bash
# C++
./graphics-demo

# C#
dotnet run

# Rust
cargo run
```

## Graphics API

### Window Management

```go
w := graphics.CreateWindow("Title", width, height)
w, running := graphics.PollEvents(w)
graphics.CloseWindow(w)
```

### Drawing Functions

```go
graphics.Clear(w, color)           // Clear screen with color
graphics.Present(w)                // Display the frame

graphics.FillRect(w, rect, color)  // Filled rectangle
graphics.DrawRect(w, rect, color)  // Rectangle outline
graphics.FillCircle(w, x, y, r, c) // Filled circle
graphics.DrawCircle(w, x, y, r, c) // Circle outline
graphics.DrawLine(w, x1, y1, x2, y2, color)
graphics.DrawPoint(w, x, y, color)
```

### Colors

```go
graphics.NewColor(r, g, b, a)  // Create custom color
graphics.Red()                  // Predefined colors
graphics.Green()
graphics.Blue()
graphics.White()
```

### Geometry

```go
graphics.NewRect(x, y, width, height)
```

## Expected Output

An 800x600 window displaying:
- Dark blue background
- Red filled rectangle (top-left area)
- Green rectangle outline (top-center area)
- Blue filled circle (bottom-left area)
- White circle outline (bottom-center area)
- Yellow X pattern (right side)
- Dotted white line

Close the window to exit the program.

## See Also

- [graphics-minimal](../graphics-minimal/) - Minimal graphics example
- [mos6502](../mos6502/) - Graphics used for 6502 emulator display

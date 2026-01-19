package main

import (
	"mos6502/cpu"
	"runtime/graphics"
)

// createDrawRectProgram creates a 6502 program that draws a filled rectangle
func createDrawRectProgram(startX int, startY int, endX int, endY int, color uint8) []uint8 {
	program := []uint8{}

	row := startY
	for {
		if row > endY {
			break
		}
		col := startX
		for {
			if col > endX {
				break
			}
			addr := 0x0200 + (row * 32) + col
			addrLow := uint8(addr & 0xFF)
			addrHigh := uint8((addr >> 8) & 0xFF)
			program = append(program, uint8(cpu.OpLDAImm))
			program = append(program, color)
			program = append(program, uint8(cpu.OpSTAAbs))
			program = append(program, addrLow)
			program = append(program, addrHigh)
			col = col + 1
		}
		row = row + 1
	}

	return program
}

// createDiagonalLineProgram creates a program that draws a diagonal line
func createDiagonalLineProgram(x1 int, y1 int, x2 int, y2 int, color uint8) []uint8 {
	program := []uint8{}

	steps := x2 - x1
	if steps < 0 {
		steps = -steps
	}
	dy := y2 - y1
	if dy < 0 {
		dy = -dy
	}
	if dy > steps {
		steps = dy
	}

	if steps == 0 {
		steps = 1
	}

	i := 0
	for {
		if i > steps {
			break
		}
		x := x1 + (i * (x2 - x1) / steps)
		y := y1 + (i * (y2 - y1) / steps)

		if x >= 0 {
			if x < 32 {
				if y >= 0 {
					if y < 32 {
						addr := 0x0200 + (y * 32) + x
						addrLow := uint8(addr & 0xFF)
						addrHigh := uint8((addr >> 8) & 0xFF)
						program = append(program, uint8(cpu.OpLDAImm))
						program = append(program, color)
						program = append(program, uint8(cpu.OpSTAAbs))
						program = append(program, addrLow)
						program = append(program, addrHigh)
					}
				}
			}
		}
		i = i + 1
	}

	return program
}

func main() {
	// Create window (32x32 screen scaled up)
	scale := int32(16)
	windowWidth := int32(cpu.ScreenWidth) * scale
	windowHeight := int32(cpu.ScreenHeight) * scale
	w := graphics.CreateWindow("MOS 6502 Emulator", windowWidth, windowHeight)

	// Create CPU
	c := cpu.NewCPU()

	// Build the demo program
	program := []uint8{}

	// Draw a red filled rectangle (8,4) to (24,12)
	rectProg := createDrawRectProgram(8, 4, 24, 12, uint8(2))
	idx := 0
	for {
		if idx >= len(rectProg) {
			break
		}
		program = append(program, rectProg[idx])
		idx = idx + 1
	}

	// Draw a green filled rectangle (4,16) to (12,28)
	rectProg2 := createDrawRectProgram(4, 16, 12, 28, uint8(3))
	idx = 0
	for {
		if idx >= len(rectProg2) {
			break
		}
		program = append(program, rectProg2[idx])
		idx = idx + 1
	}

	// Draw a blue filled rectangle (18,18) to (28,26)
	rectProg3 := createDrawRectProgram(18, 18, 28, 26, uint8(4))
	idx = 0
	for {
		if idx >= len(rectProg3) {
			break
		}
		program = append(program, rectProg3[idx])
		idx = idx + 1
	}

	// Draw yellow diagonal line
	lineProg := createDiagonalLineProgram(2, 2, 30, 14, uint8(5))
	idx = 0
	for {
		if idx >= len(lineProg) {
			break
		}
		program = append(program, lineProg[idx])
		idx = idx + 1
	}

	// Draw white diagonal line
	lineProg2 := createDiagonalLineProgram(30, 2, 2, 14, uint8(1))
	idx = 0
	for {
		if idx >= len(lineProg2) {
			break
		}
		program = append(program, lineProg2[idx])
		idx = idx + 1
	}

	// Add halt instruction
	program = append(program, uint8(cpu.OpBRK))

	// Load and run program
	c = cpu.LoadProgram(c, program, 0x0600)
	c = cpu.SetPC(c, 0x0600)
	c = cpu.Run(c, 100000)

	// Main display loop
	for {
		var running bool
		w, running = graphics.PollEvents(w)
		if !running {
			break
		}

		// Clear screen with dark background
		graphics.Clear(w, graphics.NewColor(16, 16, 32, 255))

		// Render the 6502 screen memory
		y := 0
		for {
			if y >= cpu.ScreenHeight {
				break
			}
			x := 0
			for {
				if x >= cpu.ScreenWidth {
					break
				}
				colorVal := cpu.GetScreenPixel(c, x, y)
				if colorVal != 0 {
					px := int32(x) * scale
					py := int32(y) * scale
					// Map color value to actual color
					if colorVal == 1 {
						graphics.FillRect(w, graphics.NewRect(px, py, scale, scale), graphics.White())
					} else if colorVal == 2 {
						graphics.FillRect(w, graphics.NewRect(px, py, scale, scale), graphics.Red())
					} else if colorVal == 3 {
						graphics.FillRect(w, graphics.NewRect(px, py, scale, scale), graphics.Green())
					} else if colorVal == 4 {
						graphics.FillRect(w, graphics.NewRect(px, py, scale, scale), graphics.Blue())
					} else if colorVal == 5 {
						graphics.FillRect(w, graphics.NewRect(px, py, scale, scale), graphics.NewColor(255, 255, 0, 255))
					} else if colorVal == 6 {
						graphics.FillRect(w, graphics.NewRect(px, py, scale, scale), graphics.NewColor(255, 0, 255, 255))
					} else if colorVal == 7 {
						graphics.FillRect(w, graphics.NewRect(px, py, scale, scale), graphics.NewColor(0, 255, 255, 255))
					} else if colorVal == 8 {
						graphics.FillRect(w, graphics.NewRect(px, py, scale, scale), graphics.NewColor(128, 128, 128, 255))
					} else if colorVal == 9 {
						graphics.FillRect(w, graphics.NewRect(px, py, scale, scale), graphics.NewColor(255, 128, 0, 255))
					} else {
						graphics.FillRect(w, graphics.NewRect(px, py, scale, scale), graphics.White())
					}
				}
				x = x + 1
			}
			y = y + 1
		}

		// Present frame
		graphics.Present(w)
	}

	graphics.CloseWindow(w)
}

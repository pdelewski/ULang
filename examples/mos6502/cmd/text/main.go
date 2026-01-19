package main

import (
	"mos6502lib/assembler"
	"mos6502lib/cpu"
	"mos6502lib/font"
	"runtime/graphics"
)

// Text screen constants
// Screen is 4 characters wide x 4 characters tall (uses 32x32 pixel display)
// Each character is 8x8 pixels
const TextCols = 4
const TextRows = 4

// Text screen memory starts at $0200
const TextScreenBase = 0x0200

// hexDigit converts a value 0-15 to a hex character string
func hexDigit(n int) string {
	if n == 0 {
		return "0"
	} else if n == 1 {
		return "1"
	} else if n == 2 {
		return "2"
	} else if n == 3 {
		return "3"
	} else if n == 4 {
		return "4"
	} else if n == 5 {
		return "5"
	} else if n == 6 {
		return "6"
	} else if n == 7 {
		return "7"
	} else if n == 8 {
		return "8"
	} else if n == 9 {
		return "9"
	} else if n == 10 {
		return "A"
	} else if n == 11 {
		return "B"
	} else if n == 12 {
		return "C"
	} else if n == 13 {
		return "D"
	} else if n == 14 {
		return "E"
	} else if n == 15 {
		return "F"
	}
	return "0"
}

// toHex2 converts a byte to 2-digit hex string
func toHex2(n int) string {
	high := (n >> 4) & 0x0F
	low := n & 0x0F
	return hexDigit(high) + hexDigit(low)
}

// toHex4 converts a 16-bit value to 4-digit hex string
func toHex4(n int) string {
	return toHex2((n>>8)&0xFF) + toHex2(n&0xFF)
}

// makeLdaImm creates "LDA #$XX" instruction
func makeLdaImm(value int) string {
	return "LDA #$" + toHex2(value)
}

// makeStaAbs creates "STA $XXXX" instruction
func makeStaAbs(addr int) string {
	return "STA $" + toHex4(addr)
}

// createHelloWorldDemo creates a demo that displays text
func createHelloWorldDemo() []uint8 {
	lines := []string{}

	// Row 0: "HI"
	lines = append(lines, "LDA #$48") // H
	lines = append(lines, "STA $0200")
	lines = append(lines, "LDA #$49") // I
	lines = append(lines, "STA $0201")

	// Row 1: "6502"
	lines = append(lines, "LDA #$36") // 6
	lines = append(lines, "STA $0204")
	lines = append(lines, "LDA #$35") // 5
	lines = append(lines, "STA $0205")
	lines = append(lines, "LDA #$30") // 0
	lines = append(lines, "STA $0206")
	lines = append(lines, "LDA #$32") // 2
	lines = append(lines, "STA $0207")

	// Row 2: "CPU"
	lines = append(lines, "LDA #$43") // C
	lines = append(lines, "STA $0208")
	lines = append(lines, "LDA #$50") // P
	lines = append(lines, "STA $0209")
	lines = append(lines, "LDA #$55") // U
	lines = append(lines, "STA $020A")

	// Row 3: "TEST"
	lines = append(lines, "LDA #$54") // T
	lines = append(lines, "STA $020C")
	lines = append(lines, "LDA #$45") // E
	lines = append(lines, "STA $020D")
	lines = append(lines, "LDA #$53") // S
	lines = append(lines, "STA $020E")
	lines = append(lines, "LDA #$54") // T
	lines = append(lines, "STA $020F")

	lines = append(lines, "BRK")
	return assembler.AssembleLines(lines)
}

func main() {
	// Create window (32x32 screen scaled up)
	scale := int32(16)
	windowWidth := int32(cpu.ScreenWidth) * scale
	windowHeight := int32(cpu.ScreenHeight) * scale
	w := graphics.CreateWindow("MOS 6502 Text Mode", windowWidth, windowHeight)

	// Create CPU
	c := cpu.NewCPU()

	// Load font data
	fontData := font.GetFontData()

	// Create the demo program
	program := createHelloWorldDemo()

	// Load and run program
	c = cpu.LoadProgram(c, program, 0x0600)
	c = cpu.SetPC(c, 0x0600)
	c = cpu.Run(c, 10000)

	// Text color (green like old terminals)
	textColor := graphics.NewColor(0, 255, 0, 255)

	// Main display loop
	for {
		var running bool
		w, running = graphics.PollEvents(w)
		if !running {
			break
		}

		// Clear screen with dark background
		graphics.Clear(w, graphics.NewColor(0, 0, 0, 255))

		// Render the text screen
		charY := 0
		for {
			if charY >= TextRows {
				break
			}
			charX := 0
			for {
				if charX >= TextCols {
					break
				}
				// Get character code from screen memory
				memAddr := TextScreenBase + (charY * TextCols) + charX
				charCode := int(cpu.GetMemory(c, memAddr))

				// Only render printable characters (32-127)
				if charCode >= 32 {
					if charCode <= 127 {
						// Render 8x8 character bitmap
						pixelY := 0
						for {
							if pixelY >= 8 {
								break
							}
							pixelX := 0
							for {
								if pixelX >= 8 {
									break
								}
								// Check if this pixel is set in the font
								if font.GetPixel(fontData, charCode, pixelX, pixelY) {
									screenX := int32(charX*8+pixelX) * scale
									screenY := int32(charY*8+pixelY) * scale
									graphics.FillRect(w, graphics.NewRect(screenX, screenY, scale, scale), textColor)
								}
								pixelX = pixelX + 1
							}
							pixelY = pixelY + 1
						}
					}
				}
				charX = charX + 1
			}
			charY = charY + 1
		}

		// Present frame
		graphics.Present(w)
	}

	graphics.CloseWindow(w)
}

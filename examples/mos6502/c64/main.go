package main

import (
	"mos6502lib/assembler"
	"mos6502lib/cpu"
	"mos6502lib/font"
	"runtime/graphics"
)

// C64-style text screen constants
// Screen is 40 characters wide x 25 characters tall (like real C64)
const TextCols = 40
const TextRows = 25

// Text screen memory starts at $0400 (like real C64)
const TextScreenBase = 0x0400

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

// toHex converts an integer to a hex string with $ prefix
// Uses 2 digits for values <= 255, 4 digits otherwise
func toHex(n int) string {
	if n > 255 {
		return "$" + toHex4(n)
	}
	return "$" + toHex2(n)
}

// addStringToScreen generates assembly to write a string to screen memory
func addStringToScreen(lines []string, text string, row int, col int) []string {
	baseAddr := TextScreenBase + (row * TextCols) + col
	i := 0
	for {
		if i >= len(text) {
			break
		}
		charCode := int(text[i])
		addr := baseAddr + i
		lines = append(lines, "LDA #"+toHex(charCode))
		lines = append(lines, "STA "+toHex(addr))
		i = i + 1
	}
	return lines
}

// clearScreen generates assembly to fill screen memory with spaces
func clearScreen(lines []string) []string {
	// Fill all 1000 screen locations (40x25) with space character (0x20)
	addr := TextScreenBase
	i := 0
	for {
		if i >= TextCols*TextRows {
			break
		}
		lines = append(lines, "LDA #$20") // space character
		lines = append(lines, "STA "+toHex(addr+i))
		i = i + 1
	}
	return lines
}

// createC64WelcomeScreen creates the classic C64 boot screen
func createC64WelcomeScreen() []uint8 {
	lines := []string{}

	// First, clear the screen with spaces
	lines = clearScreen(lines)

	// Classic C64 boot screen:
	//
	//     **** COMMODORE 64 BASIC V2 ****
	//
	//  64K RAM SYSTEM  38911 BASIC BYTES FREE
	//
	// READY.
	// _

	// Row 1: "    **** COMMODORE 64 BASIC V2 ****"
	lines = addStringToScreen(lines, "**** COMMODORE 64 BASIC V2 ****", 1, 4)

	// Row 3: " 64K RAM SYSTEM  38911 BASIC BYTES FREE"
	lines = addStringToScreen(lines, "64K RAM SYSTEM  38911 BASIC BYTES FREE", 3, 1)

	// Row 5: "READY."
	lines = addStringToScreen(lines, "READY.", 5, 0)

	// Row 6: Cursor - use underscore as cursor representation
	// Row 6, col 0 = 0x0400 + (6 * 40) = 0x0400 + 240 = 0x04F0
	lines = append(lines, "LDA #$5F") // underscore cursor (ASCII 95)
	lines = append(lines, "STA $04F0")

	lines = append(lines, "BRK")
	return assembler.AssembleLines(lines)
}

func main() {
	// Create window (320x200 C64 resolution scaled up)
	scale := int32(2)
	windowWidth := int32(TextCols*8) * scale
	windowHeight := int32(TextRows*8) * scale
	w := graphics.CreateWindow("Commodore 64", windowWidth, windowHeight)

	// Create CPU
	c := cpu.NewCPU()

	// Load font data
	fontData := font.GetFontData()

	// Create the C64 welcome screen program
	program := createC64WelcomeScreen()

	// Load and run program
	c = cpu.LoadProgram(c, program, 0x0600)
	c = cpu.SetPC(c, 0x0600)
	c = cpu.Run(c, 100000)

	// C64 colors: light blue text on dark blue background
	textColor := graphics.NewColor(134, 122, 222, 255) // C64 light blue
	bgColor := graphics.NewColor(64, 50, 133, 255)     // C64 dark blue

	// Main display loop
	for {
		var running bool
		w, running = graphics.PollEvents(w)
		if !running {
			break
		}

		// Clear screen with C64 dark blue background
		graphics.Clear(w, bgColor)

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

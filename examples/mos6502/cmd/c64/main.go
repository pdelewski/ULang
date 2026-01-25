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
	// Load space character once, then just STA to each location
	lines = append(lines, "LDA #$20") // space character - load once
	addr := TextScreenBase
	i := 0
	for {
		if i >= TextCols*TextRows {
			break
		}
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

	// Row 7: Cursor - use underscore as cursor representation (blank line after READY.)
	// Row 7, col 0 = 0x0400 + (7 * 40) = 0x0400 + 280 = 0x0518
	lines = append(lines, "LDA #$5F") // underscore cursor (ASCII 95)
	lines = append(lines, "STA $0518")

	lines = append(lines, "BRK")
	return assembler.AssembleLines(lines)
}

func main() {
	// Create window (320x200 C64 resolution scaled up)
	scale := int32(4)
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

	// Cursor position (starts on row 7, after blank line below READY.)
	cursorRow := 7
	cursorCol := 0

	// Main display loop
	graphics.RunLoop(w, func(w graphics.Window) bool {
		// Handle keyboard input
		key := graphics.GetLastKey()
		if key != 0 {
			// Clear old cursor BEFORE changing position
			oldCursorAddr := TextScreenBase + (cursorRow * TextCols) + cursorCol
			if c.Memory[oldCursorAddr] == 95 {
				c.Memory[oldCursorAddr] = 32 // clear old cursor
			}

			if key == 13 {
				// Enter - move to next line
				cursorCol = 0
				cursorRow = cursorRow + 1
				if cursorRow >= TextRows {
					cursorRow = TextRows - 1
				}
			} else if key == 8 {
				// Backspace - move back and clear
				if cursorCol > 0 {
					cursorCol = cursorCol - 1
					// Clear the character at cursor position
					addr := TextScreenBase + (cursorRow * TextCols) + cursorCol
					c.Memory[addr] = 32 // space
				} else if cursorRow > 0 {
					// At beginning of line - move to end of previous line
					cursorRow = cursorRow - 1
					cursorCol = TextCols - 1
				}
			} else if key >= 32 && key <= 126 {
				// Printable character
				addr := TextScreenBase + (cursorRow * TextCols) + cursorCol
				c.Memory[addr] = uint8(key)
				cursorCol = cursorCol + 1
				if cursorCol >= TextCols {
					cursorCol = 0
					cursorRow = cursorRow + 1
					if cursorRow >= TextRows {
						cursorRow = TextRows - 1
					}
				}
			}
			// Draw cursor at new position
			cursorAddr := TextScreenBase + (cursorRow * TextCols) + cursorCol
			if c.Memory[cursorAddr] == 32 {
				c.Memory[cursorAddr] = 95 // underscore cursor
			}
		}

		// Clear screen with C64 dark blue background
		graphics.Clear(w, bgColor)

		// Render the text screen
		memAddr := TextScreenBase
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
				charCode := int(cpu.GetMemory(c, memAddr))
				memAddr = memAddr + 1

				// Only render printable non-space characters (33-127)
				// Skip spaces (32) as they have no pixels to render
				if charCode > 32 && charCode <= 127 {
					// Render 8x8 character bitmap
					baseScreenX := int32(charX * 8)
					baseScreenY := int32(charY * 8)
					pixelY := 0
					for {
						if pixelY >= 8 {
							break
						}
						// Get entire row byte once (reduces function calls from 64 to 8)
						rowByte := font.GetRow(fontData, charCode, pixelY)
						if rowByte != 0 {
							// Only iterate if row has any pixels
							mask := uint8(0x80)
							pixelX := 0
							for {
								if pixelX >= 8 {
									break
								}
								if (rowByte & mask) != 0 {
									screenX := (baseScreenX + int32(pixelX)) * scale
									screenY := (baseScreenY + int32(pixelY)) * scale
									graphics.FillRect(w, graphics.NewRect(screenX, screenY, scale, scale), textColor)
								}
								mask = mask >> 1
								pixelX = pixelX + 1
							}
						}
						pixelY = pixelY + 1
					}
				}
				charX = charX + 1
			}
			charY = charY + 1
		}

		// Present frame
		graphics.Present(w)

		return true
	})

	graphics.CloseWindow(w)
}

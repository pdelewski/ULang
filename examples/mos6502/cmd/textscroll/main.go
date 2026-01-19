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

// getCharAt returns the character at position in the message with scrolling
// The message scrolls horizontally across a 4-character window
func getCharAt(message string, scrollOffset int, position int) int {
	// Calculate the index in the message
	msgLen := len(message)
	if msgLen == 0 {
		return 32 // space
	}
	idx := (scrollOffset + position) % msgLen
	return int(message[idx])
}

// createScrollingTextProgram creates a 6502 program that displays scrolling text
// message: the text to scroll
// scrollOffset: current scroll position
// row: which row (0-3) to display the text on
func createScrollingTextProgram(message string, scrollOffset int, row int) []uint8 {
	lines := []string{}

	// Write 4 characters for this row
	col := 0
	for {
		if col >= TextCols {
			break
		}
		charCode := getCharAt(message, scrollOffset, col)
		memAddr := TextScreenBase + (row * TextCols) + col
		lines = append(lines, makeLdaImm(charCode))
		lines = append(lines, makeStaAbs(memAddr))
		col = col + 1
	}

	return assembler.AssembleLines(lines)
}

// createMultiLineTextProgram creates a program displaying multiple lines of scrolling text
func createMultiLineTextProgram(messages []string, scrollOffsets []int) []uint8 {
	lines := []string{}

	row := 0
	for {
		if row >= TextRows {
			break
		}
		if row >= len(messages) {
			break
		}

		message := messages[row]
		scrollOffset := 0
		if row < len(scrollOffsets) {
			scrollOffset = scrollOffsets[row]
		}

		// Write 4 characters for this row
		col := 0
		for {
			if col >= TextCols {
				break
			}
			charCode := getCharAt(message, scrollOffset, col)
			memAddr := TextScreenBase + (row * TextCols) + col
			lines = append(lines, makeLdaImm(charCode))
			lines = append(lines, makeStaAbs(memAddr))
			col = col + 1
		}

		row = row + 1
	}

	lines = append(lines, "BRK")
	return assembler.AssembleLines(lines)
}

// clearScreenMemory clears the text screen area in CPU memory
func clearScreenMemory(c cpu.CPU) cpu.CPU {
	addr := TextScreenBase
	for {
		if addr >= TextScreenBase+(TextCols*TextRows) {
			break
		}
		c.Memory[addr] = 32 // space character
		addr = addr + 1
	}
	return c
}

func main() {
	// Create window (32x32 screen scaled up)
	scale := int32(16)
	windowWidth := int32(cpu.ScreenWidth) * scale
	windowHeight := int32(cpu.ScreenHeight) * scale
	w := graphics.CreateWindow("MOS 6502 Scrolling Text", windowWidth, windowHeight)

	// Load font data
	fontData := font.GetFontData()

	// Define the scrolling messages for each row
	// Adding spaces for smoother scrolling effect
	messages := []string{
		"HELLO MOS6502!  ",
		"TEXT MODE DEMO  ",
		"SCROLLING TEXT  ",
		"<< RETRO CPU >> ",
	}

	// Scroll offsets for each row (they scroll at different speeds for effect)
	scrollOffsets := []int{0, 0, 0, 0}

	// Frame counter for animation timing
	frameCount := 0

	// Text color (green like old terminals)
	textColor := graphics.NewColor(0, 255, 0, 255)

	// Background color (dark)
	bgColor := graphics.NewColor(0, 0, 0, 255)

	// Main display loop
	for {
		var running bool
		w, running = graphics.PollEvents(w)
		if !running {
			break
		}

		// Update scroll positions every few frames
		frameCount = frameCount + 1
		if frameCount >= 8 {
			frameCount = 0

			// Scroll each row at different speeds
			scrollOffsets[0] = scrollOffsets[0] + 1
			if scrollOffsets[0] >= len(messages[0]) {
				scrollOffsets[0] = 0
			}

			// Row 1 scrolls every other update
			if scrollOffsets[0]%2 == 0 {
				scrollOffsets[1] = scrollOffsets[1] + 1
				if scrollOffsets[1] >= len(messages[1]) {
					scrollOffsets[1] = 0
				}
			}

			// Row 2 scrolls in opposite direction
			scrollOffsets[2] = scrollOffsets[2] - 1
			if scrollOffsets[2] < 0 {
				scrollOffsets[2] = len(messages[2]) - 1
			}

			// Row 3 scrolls at same speed as row 0
			scrollOffsets[3] = scrollOffsets[3] + 1
			if scrollOffsets[3] >= len(messages[3]) {
				scrollOffsets[3] = 0
			}
		}

		// Create CPU and run the program to update screen memory
		c := cpu.NewCPU()
		c = clearScreenMemory(c)

		// Create the program with current scroll offsets
		program := createMultiLineTextProgram(messages, scrollOffsets)

		// Load and run program
		c = cpu.LoadProgram(c, program, 0x0600)
		c = cpu.SetPC(c, 0x0600)
		c = cpu.Run(c, 10000)

		// Clear screen with dark background
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
	}

	graphics.CloseWindow(w)
}
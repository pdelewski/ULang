package basic

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
func toHex(n int) string {
	if n > 255 {
		return "$" + toHex4(n)
	}
	return "$" + toHex2(n)
}

// genPrint generates assembly to print text to screen
// cursorRow and cursorCol specify where to start printing
func genPrint(args string, cursorRow int, cursorCol int) []string {
	lines := []string{}

	// Parse the string to print
	text := parseString(args)

	// Calculate base address
	baseAddr := ScreenBase + (cursorRow * TextCols) + cursorCol

	// Generate LDA/STA for each character
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

// genPoke generates assembly for POKE addr, value
func genPoke(addr int, value int) []string {
	lines := []string{}

	// Clamp value to byte range
	if value > 255 {
		value = value & 0xFF
	}
	if value < 0 {
		value = 0
	}

	lines = append(lines, "LDA #"+toHex(value))
	lines = append(lines, "STA "+toHex(addr))

	return lines
}

// genClear generates assembly to clear the screen
func genClear() []string {
	lines := []string{}

	// Load space character once
	lines = append(lines, "LDA #$20")

	// Store to each screen location (40x25 = 1000 locations)
	i := 0
	for {
		if i >= TextCols*TextRows {
			break
		}
		addr := ScreenBase + i
		lines = append(lines, "STA "+toHex(addr))
		i = i + 1
	}

	return lines
}

// genList generates assembly to display program listing
func genList(state BasicState, startRow int) []string {
	lines := []string{}

	row := startRow
	i := 0
	for {
		if i >= len(state.Lines) {
			break
		}
		if row >= TextRows {
			break
		}

		// Format: "lineNum text"
		lineNum := state.Lines[i].LineNum
		text := state.Lines[i].Text

		// Convert line number to string
		numStr := intToString(lineNum)

		// Build full line: "10 PRINT HELLO"
		fullLine := numStr + " " + text

		// Generate assembly to print this line
		baseAddr := ScreenBase + (row * TextCols)
		j := 0
		for {
			if j >= len(fullLine) {
				break
			}
			if j >= TextCols {
				break
			}
			charCode := int(fullLine[j])
			addr := baseAddr + j
			lines = append(lines, "LDA #"+toHex(charCode))
			lines = append(lines, "STA "+toHex(addr))
			j = j + 1
		}

		row = row + 1
		i = i + 1
	}

	return lines
}

// intToString converts an integer to a string (goany compatible)
func intToString(n int) string {
	if n == 0 {
		return "0"
	}

	// Handle negative numbers
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}

	// Build digits in reverse
	digits := ""
	for {
		if n == 0 {
			break
		}
		digit := n % 10
		digits = digitToChar(digit) + digits
		n = n / 10
	}

	if neg {
		digits = "-" + digits
	}

	return digits
}

// digitToChar converts a digit 0-9 to its character
func digitToChar(d int) string {
	if d == 0 {
		return "0"
	} else if d == 1 {
		return "1"
	} else if d == 2 {
		return "2"
	} else if d == 3 {
		return "3"
	} else if d == 4 {
		return "4"
	} else if d == 5 {
		return "5"
	} else if d == 6 {
		return "6"
	} else if d == 7 {
		return "7"
	} else if d == 8 {
		return "8"
	} else if d == 9 {
		return "9"
	}
	return "0"
}

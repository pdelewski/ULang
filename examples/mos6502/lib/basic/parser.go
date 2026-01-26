package basic

// parseLine extracts the command and arguments from a BASIC line
// Returns command name (uppercase) and the rest of the line as arguments
func parseLine(line string) (string, string) {
	// Skip leading whitespace
	pos := 0
	for {
		if pos >= len(line) {
			break
		}
		if line[pos] != ' ' && line[pos] != '\t' {
			break
		}
		pos = pos + 1
	}

	// Extract command (letters only)
	cmd := ""
	for {
		if pos >= len(line) {
			break
		}
		ch := int(line[pos])
		if !isLetterCode(ch) {
			break
		}
		cmd = cmd + toUpperChar(ch)
		pos = pos + 1
	}

	// Skip whitespace after command
	for {
		if pos >= len(line) {
			break
		}
		if line[pos] != ' ' && line[pos] != '\t' {
			break
		}
		pos = pos + 1
	}

	// Rest is arguments - build manually
	args := ""
	for {
		if pos >= len(line) {
			break
		}
		args = args + charToString(int(line[pos]))
		pos = pos + 1
	}

	return cmd, args
}

// parseLineNumber extracts a line number from the beginning of a line
// Returns lineNum, rest of line, and whether a line number was found
func parseLineNumber(line string) (int, string, bool) {
	// Skip leading whitespace
	pos := 0
	for {
		if pos >= len(line) {
			break
		}
		if line[pos] != ' ' && line[pos] != '\t' {
			break
		}
		pos = pos + 1
	}

	// Check if starts with digit
	if pos >= len(line) || !isDigitCode(int(line[pos])) {
		return 0, line, false
	}

	// Extract number - build digit string manually
	lineNum := 0
	for {
		if pos >= len(line) {
			break
		}
		ch := int(line[pos])
		if !isDigitCode(ch) {
			break
		}
		lineNum = lineNum*10 + (ch - int('0'))
		pos = pos + 1
	}

	// Skip whitespace after number
	for {
		if pos >= len(line) {
			break
		}
		if line[pos] != ' ' && line[pos] != '\t' {
			break
		}
		pos = pos + 1
	}

	// Build rest manually
	rest := ""
	for {
		if pos >= len(line) {
			break
		}
		rest = rest + charToString(int(line[pos]))
		pos = pos + 1
	}

	return lineNum, rest, true
}

// parsePoke extracts address and value from POKE arguments
// Format: "address, value" or "address,value"
func parsePoke(args string) (int, int) {
	// Find comma and parse address
	addr := 0
	i := 0

	// Skip leading spaces
	for {
		if i >= len(args) {
			break
		}
		if args[i] != ' ' && args[i] != '\t' {
			break
		}
		i = i + 1
	}

	// Parse address number
	for {
		if i >= len(args) {
			break
		}
		ch := int(args[i])
		if ch == int(',') {
			break
		}
		if ch == int(' ') || ch == int('\t') {
			i = i + 1
			continue
		}
		if isDigitCode(ch) {
			addr = addr*10 + (ch - int('0'))
		}
		i = i + 1
	}

	// Skip comma
	if i < len(args) && args[i] == ',' {
		i = i + 1
	}

	// Skip spaces after comma
	for {
		if i >= len(args) {
			break
		}
		if args[i] != ' ' && args[i] != '\t' {
			break
		}
		i = i + 1
	}

	// Parse value number
	value := 0
	for {
		if i >= len(args) {
			break
		}
		ch := int(args[i])
		if ch == int(' ') || ch == int('\t') {
			break
		}
		if isDigitCode(ch) {
			value = value*10 + (ch - int('0'))
		}
		i = i + 1
	}

	return addr, value
}

// parseString extracts a string literal from arguments
// Handles "text" format, returns the text without quotes
func parseString(args string) string {
	// Find first quote
	start := -1
	i := 0
	for {
		if i >= len(args) {
			break
		}
		if args[i] == '"' {
			start = i + 1
			break
		}
		i = i + 1
	}

	if start < 0 {
		// No quote found, return args as-is (for PRINT without quotes)
		return trimSpacesStr(args)
	}

	// Extract string content until closing quote
	result := ""
	j := start
	for {
		if j >= len(args) {
			break
		}
		if args[j] == '"' {
			break
		}
		result = result + charToString(int(args[j]))
		j = j + 1
	}

	return result
}

// charToString converts a character code to a single-character string
func charToString(ch int) string {
	if ch >= 32 && ch <= 126 {
		// Printable ASCII - build lookup
		if ch == 32 {
			return " "
		} else if ch == 33 {
			return "!"
		} else if ch == 34 {
			// double quote - skip for now
			return ""
		} else if ch == 35 {
			return "#"
		} else if ch == 36 {
			return "$"
		} else if ch == 37 {
			return "%"
		} else if ch == 38 {
			return "&"
		} else if ch == 39 {
			return "'"
		} else if ch == 40 {
			return "("
		} else if ch == 41 {
			return ")"
		} else if ch == 42 {
			return "*"
		} else if ch == 43 {
			return "+"
		} else if ch == 44 {
			return ","
		} else if ch == 45 {
			return "-"
		} else if ch == 46 {
			return "."
		} else if ch == 47 {
			return "/"
		} else if ch == 48 {
			return "0"
		} else if ch == 49 {
			return "1"
		} else if ch == 50 {
			return "2"
		} else if ch == 51 {
			return "3"
		} else if ch == 52 {
			return "4"
		} else if ch == 53 {
			return "5"
		} else if ch == 54 {
			return "6"
		} else if ch == 55 {
			return "7"
		} else if ch == 56 {
			return "8"
		} else if ch == 57 {
			return "9"
		} else if ch == 58 {
			return ":"
		} else if ch == 59 {
			return ";"
		} else if ch == 60 {
			return "<"
		} else if ch == 61 {
			return "="
		} else if ch == 62 {
			return ">"
		} else if ch == 63 {
			return "?"
		} else if ch == 64 {
			return "@"
		} else if ch == 65 {
			return "A"
		} else if ch == 66 {
			return "B"
		} else if ch == 67 {
			return "C"
		} else if ch == 68 {
			return "D"
		} else if ch == 69 {
			return "E"
		} else if ch == 70 {
			return "F"
		} else if ch == 71 {
			return "G"
		} else if ch == 72 {
			return "H"
		} else if ch == 73 {
			return "I"
		} else if ch == 74 {
			return "J"
		} else if ch == 75 {
			return "K"
		} else if ch == 76 {
			return "L"
		} else if ch == 77 {
			return "M"
		} else if ch == 78 {
			return "N"
		} else if ch == 79 {
			return "O"
		} else if ch == 80 {
			return "P"
		} else if ch == 81 {
			return "Q"
		} else if ch == 82 {
			return "R"
		} else if ch == 83 {
			return "S"
		} else if ch == 84 {
			return "T"
		} else if ch == 85 {
			return "U"
		} else if ch == 86 {
			return "V"
		} else if ch == 87 {
			return "W"
		} else if ch == 88 {
			return "X"
		} else if ch == 89 {
			return "Y"
		} else if ch == 90 {
			return "Z"
		} else if ch == 91 {
			return "["
		} else if ch == 92 {
			// backslash - return empty for now
			return ""
		} else if ch == 93 {
			return "]"
		} else if ch == 94 {
			return "^"
		} else if ch == 95 {
			return "_"
		} else if ch == 96 {
			return "`"
		} else if ch == 97 {
			return "a"
		} else if ch == 98 {
			return "b"
		} else if ch == 99 {
			return "c"
		} else if ch == 100 {
			return "d"
		} else if ch == 101 {
			return "e"
		} else if ch == 102 {
			return "f"
		} else if ch == 103 {
			return "g"
		} else if ch == 104 {
			return "h"
		} else if ch == 105 {
			return "i"
		} else if ch == 106 {
			return "j"
		} else if ch == 107 {
			return "k"
		} else if ch == 108 {
			return "l"
		} else if ch == 109 {
			return "m"
		} else if ch == 110 {
			return "n"
		} else if ch == 111 {
			return "o"
		} else if ch == 112 {
			return "p"
		} else if ch == 113 {
			return "q"
		} else if ch == 114 {
			return "r"
		} else if ch == 115 {
			return "s"
		} else if ch == 116 {
			return "t"
		} else if ch == 117 {
			return "u"
		} else if ch == 118 {
			return "v"
		} else if ch == 119 {
			return "w"
		} else if ch == 120 {
			return "x"
		} else if ch == 121 {
			return "y"
		} else if ch == 122 {
			return "z"
		} else if ch == 123 {
			return "{"
		} else if ch == 124 {
			return "|"
		} else if ch == 125 {
			return "}"
		} else if ch == 126 {
			return "~"
		}
	}
	return ""
}

// isLetterCode checks if a character code is a letter
func isLetterCode(ch int) bool {
	return (ch >= int('a') && ch <= int('z')) || (ch >= int('A') && ch <= int('Z'))
}

// isDigitCode checks if a character code is a digit
func isDigitCode(ch int) bool {
	return ch >= int('0') && ch <= int('9')
}

// toUpperChar converts a character code to uppercase character string
func toUpperChar(ch int) string {
	if ch >= int('a') && ch <= int('z') {
		ch = ch - 32
	}
	// Convert to character
	if ch == 65 {
		return "A"
	} else if ch == 66 {
		return "B"
	} else if ch == 67 {
		return "C"
	} else if ch == 68 {
		return "D"
	} else if ch == 69 {
		return "E"
	} else if ch == 70 {
		return "F"
	} else if ch == 71 {
		return "G"
	} else if ch == 72 {
		return "H"
	} else if ch == 73 {
		return "I"
	} else if ch == 74 {
		return "J"
	} else if ch == 75 {
		return "K"
	} else if ch == 76 {
		return "L"
	} else if ch == 77 {
		return "M"
	} else if ch == 78 {
		return "N"
	} else if ch == 79 {
		return "O"
	} else if ch == 80 {
		return "P"
	} else if ch == 81 {
		return "Q"
	} else if ch == 82 {
		return "R"
	} else if ch == 83 {
		return "S"
	} else if ch == 84 {
		return "T"
	} else if ch == 85 {
		return "U"
	} else if ch == 86 {
		return "V"
	} else if ch == 87 {
		return "W"
	} else if ch == 88 {
		return "X"
	} else if ch == 89 {
		return "Y"
	} else if ch == 90 {
		return "Z"
	}
	return ""
}

// toUpper converts a string to uppercase (not used, but kept for reference)
func toUpper(s string) string {
	result := ""
	i := 0
	for {
		if i >= len(s) {
			break
		}
		ch := int(s[i])
		result = result + toUpperChar(ch)
		i = i + 1
	}
	return result
}

// parseNumber parses a decimal number string
func parseNumber(s string) int {
	result := 0
	i := 0
	for {
		if i >= len(s) {
			break
		}
		ch := s[i]
		if ch >= '0' && ch <= '9' {
			result = result*10 + int(ch-'0')
		}
		i = i + 1
	}
	return result
}

// trimSpacesStr removes leading and trailing spaces
func trimSpacesStr(s string) string {
	// Find start
	start := 0
	for {
		if start >= len(s) {
			break
		}
		if s[start] != ' ' && s[start] != '\t' {
			break
		}
		start = start + 1
	}

	// Find end
	end := len(s)
	for {
		if end <= start {
			break
		}
		if s[end-1] != ' ' && s[end-1] != '\t' {
			break
		}
		end = end - 1
	}

	if start >= end {
		return ""
	}

	// Build result manually without slicing
	result := ""
	i := start
	for {
		if i >= end {
			break
		}
		result = result + charToString(int(s[i]))
		i = i + 1
	}
	return result
}

// HasLineNumber checks if a line starts with a line number
func HasLineNumber(line string) bool {
	lineNum, rest, found := parseLineNumber(line)
	// Use variables to avoid compiler warning
	if lineNum > 0 && len(rest) >= 0 {
		// variables used
	}
	return found
}

// ExtractLineNumber extracts line number and rest of line
func ExtractLineNumber(line string) (int, string) {
	lineNum, rest, found := parseLineNumber(line)
	if !found {
		return 0, line
	}
	return lineNum, rest
}

// IsCommand checks if a line is a specific command (case insensitive)
func IsCommand(line string, cmdName string) bool {
	cmd, args := parseLine(line)
	// Use args to avoid unused variable
	if len(args) >= 0 {
		// variable used
	}
	return cmd == toUpper(cmdName)
}

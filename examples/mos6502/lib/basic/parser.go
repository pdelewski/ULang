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
			return "\""
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

// Variable base address in zero page ($10-$29 for A-Z)
const VarBaseAddr = 0x10

// GetVariableAddress returns the zero page address for a variable name (A-Z)
func GetVariableAddress(name string) int {
	if len(name) == 0 {
		return -1
	}
	ch := int(name[0])
	// Convert to uppercase
	if ch >= int('a') && ch <= int('z') {
		ch = ch - 32
	}
	// Check if valid variable name (A-Z)
	if ch >= int('A') && ch <= int('Z') {
		return VarBaseAddr + (ch - int('A'))
	}
	return -1
}

// IsVariableName checks if a string is a valid variable name (single letter A-Z)
func IsVariableName(s string) bool {
	if len(s) != 1 {
		return false
	}
	ch := int(s[0])
	if ch >= int('a') && ch <= int('z') {
		ch = ch - 32
	}
	return ch >= int('A') && ch <= int('Z')
}

// Expression types
const (
	ExprNumber   = 1
	ExprVariable = 2
	ExprBinaryOp = 3
)

// Expression represents a parsed expression
// Note: For goany compatibility, we don't use pointers.
// Binary ops store left/right as simple values (no deep nesting)
type Expression struct {
	Type       int
	Value      int    // For ExprNumber
	VarName    string // For ExprVariable
	Op         string // For ExprBinaryOp: +, -, *, /
	LeftType   int    // Type of left operand
	LeftValue  int    // Value if left is number
	LeftVar    string // Variable name if left is variable
	RightType  int    // Type of right operand
	RightValue int    // Value if right is number
	RightVar   string // Variable name if right is variable
}

// parseSimpleValue parses a number or variable (no operators)
func parseSimpleValue(args string) (int, int, string) {
	args = trimSpacesStr(args)
	// Check if it's a number
	if len(args) > 0 && isDigitCode(int(args[0])) {
		return ExprNumber, parseNumber(args), ""
	}
	// Check if it's a variable
	if IsVariableName(args) {
		return ExprVariable, 0, toUpperChar(int(args[0]))
	}
	return ExprNumber, 0, ""
}

// ParseExpression parses a simple expression (number, variable, or binary op)
// Note: For goany compatibility, only supports simple binary expressions (no nesting)
func ParseExpression(args string) Expression {
	args = trimSpacesStr(args)

	// Try to find a binary operator (scan from right for left-associativity)
	// Handle + and - first (lower precedence)
	parenDepth := 0
	opPos := -1
	opChar := ""

	i := len(args) - 1
	for {
		if i < 0 {
			break
		}
		ch := args[i]
		if ch == ')' {
			parenDepth = parenDepth + 1
		} else if ch == '(' {
			parenDepth = parenDepth - 1
		} else if parenDepth == 0 {
			if ch == '+' || ch == '-' {
				opPos = i
				opChar = charToString(int(ch))
				break
			}
		}
		i = i - 1
	}

	// If no + or -, look for * and /
	if opPos < 0 {
		i = len(args) - 1
		for {
			if i < 0 {
				break
			}
			ch := args[i]
			if ch == ')' {
				parenDepth = parenDepth + 1
			} else if ch == '(' {
				parenDepth = parenDepth - 1
			} else if parenDepth == 0 {
				if ch == '*' || ch == '/' {
					opPos = i
					opChar = charToString(int(ch))
					break
				}
			}
			i = i - 1
		}
	}

	// If we found an operator, parse left and right as simple values
	if opPos > 0 && opPos < len(args)-1 {
		leftStr := ""
		j := 0
		for {
			if j >= opPos {
				break
			}
			leftStr = leftStr + charToString(int(args[j]))
			j = j + 1
		}
		rightStr := ""
		j = opPos + 1
		for {
			if j >= len(args) {
				break
			}
			rightStr = rightStr + charToString(int(args[j]))
			j = j + 1
		}
		leftType, leftVal, leftVar := parseSimpleValue(leftStr)
		rightType, rightVal, rightVar := parseSimpleValue(rightStr)
		return Expression{
			Type:       ExprBinaryOp,
			Op:         opChar,
			LeftType:   leftType,
			LeftValue:  leftVal,
			LeftVar:    leftVar,
			RightType:  rightType,
			RightValue: rightVal,
			RightVar:   rightVar,
		}
	}

	// No operator found - must be a simple value (number or variable)
	// Check if it's a number
	if len(args) > 0 && isDigitCode(int(args[0])) {
		return Expression{
			Type:  ExprNumber,
			Value: parseNumber(args),
		}
	}

	// Check if it's a variable
	if IsVariableName(args) {
		return Expression{
			Type:    ExprVariable,
			VarName: toUpperChar(int(args[0])),
		}
	}

	// Default to 0
	return Expression{
		Type:  ExprNumber,
		Value: 0,
	}
}

// ParseLet parses a LET statement: LET var = expr or var = expr
// Returns variable name and expression
func ParseLet(args string) (string, Expression) {
	args = trimSpacesStr(args)

	// Find the variable name (first letter)
	varName := ""
	pos := 0

	// Skip LET keyword if present
	if len(args) >= 3 {
		first3 := ""
		j := 0
		for {
			if j >= 3 {
				break
			}
			first3 = first3 + toUpperChar(int(args[j]))
			j = j + 1
		}
		if first3 == "LET" {
			pos = 3
			// Skip whitespace after LET
			for {
				if pos >= len(args) {
					break
				}
				if args[pos] != ' ' && args[pos] != '\t' {
					break
				}
				pos = pos + 1
			}
		}
	}

	// Get variable name
	if pos < len(args) {
		varName = toUpperChar(int(args[pos]))
		pos = pos + 1
	}

	// Skip whitespace and =
	for {
		if pos >= len(args) {
			break
		}
		if args[pos] != ' ' && args[pos] != '\t' && args[pos] != '=' {
			break
		}
		pos = pos + 1
	}

	// Parse the expression
	exprStr := ""
	for {
		if pos >= len(args) {
			break
		}
		exprStr = exprStr + charToString(int(args[pos]))
		pos = pos + 1
	}

	expr := ParseExpression(exprStr)
	return varName, expr
}

// ParseGoto parses a GOTO statement and returns the target line number
func ParseGoto(args string) int {
	return parseNumber(trimSpacesStr(args))
}

// ParseGosub parses a GOSUB statement and returns the target line number
func ParseGosub(args string) int {
	return parseNumber(trimSpacesStr(args))
}

// Condition operators
const (
	CondEq  = 1 // =
	CondNe  = 2 // <> or !=
	CondLt  = 3 // <
	CondGt  = 4 // >
	CondLe  = 5 // <=
	CondGe  = 6 // >=
)

// Condition represents a parsed condition for IF statements
type Condition struct {
	Left  Expression
	Op    int
	Right Expression
}

// ParseCondition parses a condition like "A > 5" or "B = C"
func ParseCondition(args string) Condition {
	args = trimSpacesStr(args)

	// Find the operator
	opPos := -1
	opLen := 1
	op := 0

	i := 0
	for {
		if i >= len(args) {
			break
		}
		ch := args[i]
		if ch == '=' {
			opPos = i
			opLen = 1
			op = CondEq
			break
		} else if ch == '<' {
			opPos = i
			if i+1 < len(args) && args[i+1] == '>' {
				opLen = 2
				op = CondNe
			} else if i+1 < len(args) && args[i+1] == '=' {
				opLen = 2
				op = CondLe
			} else {
				opLen = 1
				op = CondLt
			}
			break
		} else if ch == '>' {
			opPos = i
			if i+1 < len(args) && args[i+1] == '=' {
				opLen = 2
				op = CondGe
			} else {
				opLen = 1
				op = CondGt
			}
			break
		}
		i = i + 1
	}

	if opPos < 0 {
		// No operator found, return default
		return Condition{Op: CondEq}
	}

	// Extract left and right parts
	leftStr := ""
	j := 0
	for {
		if j >= opPos {
			break
		}
		leftStr = leftStr + charToString(int(args[j]))
		j = j + 1
	}

	rightStr := ""
	j = opPos + opLen
	for {
		if j >= len(args) {
			break
		}
		rightStr = rightStr + charToString(int(args[j]))
		j = j + 1
	}

	return Condition{
		Left:  ParseExpression(leftStr),
		Op:    op,
		Right: ParseExpression(rightStr),
	}
}

// ParseIf parses an IF statement: IF condition THEN statement
// Returns the condition and the THEN statement
func ParseIf(args string) (Condition, string) {
	args = trimSpacesStr(args)

	// Find THEN keyword
	thenPos := -1
	i := 0
	for {
		if i >= len(args)-3 {
			break
		}
		// Check for THEN (case insensitive)
		match := true
		thenWord := "THEN"
		k := 0
		for {
			if k >= 4 {
				break
			}
			ch := int(args[i+k])
			if ch >= int('a') && ch <= int('z') {
				ch = ch - 32
			}
			if ch != int(thenWord[k]) {
				match = false
				break
			}
			k = k + 1
		}
		if match {
			thenPos = i
			break
		}
		i = i + 1
	}

	if thenPos < 0 {
		// No THEN found
		return Condition{}, ""
	}

	// Extract condition part
	condStr := ""
	j := 0
	for {
		if j >= thenPos {
			break
		}
		condStr = condStr + charToString(int(args[j]))
		j = j + 1
	}

	// Extract statement after THEN
	stmtStr := ""
	j = thenPos + 4 // Skip "THEN"
	// Skip whitespace
	for {
		if j >= len(args) {
			break
		}
		if args[j] != ' ' && args[j] != '\t' {
			break
		}
		j = j + 1
	}
	for {
		if j >= len(args) {
			break
		}
		stmtStr = stmtStr + charToString(int(args[j]))
		j = j + 1
	}

	return ParseCondition(condStr), stmtStr
}

// ForLoopInfo stores information about a FOR loop
type ForLoopInfo struct {
	VarName   string
	StartVal  int
	EndVal    int
	LoopAddr  int // Address of loop start for NEXT to jump back to
}

// ParseFor parses a FOR statement: FOR var = start TO end
// Returns variable name, start value, end value
func ParseFor(args string) (string, int, int) {
	args = trimSpacesStr(args)

	// Find variable name
	varName := ""
	pos := 0
	if pos < len(args) {
		varName = toUpperChar(int(args[pos]))
		pos = pos + 1
	}

	// Skip to =
	for {
		if pos >= len(args) {
			break
		}
		if args[pos] == '=' {
			pos = pos + 1
			break
		}
		pos = pos + 1
	}

	// Skip whitespace
	for {
		if pos >= len(args) {
			break
		}
		if args[pos] != ' ' && args[pos] != '\t' {
			break
		}
		pos = pos + 1
	}

	// Parse start value
	startStr := ""
	for {
		if pos >= len(args) {
			break
		}
		ch := int(args[pos])
		// Stop at whitespace or T (for TO)
		if args[pos] == ' ' || args[pos] == '\t' {
			break
		}
		if ch == int('T') || ch == int('t') {
			break
		}
		startStr = startStr + charToString(ch)
		pos = pos + 1
	}

	// Skip TO keyword
	for {
		if pos >= len(args) {
			break
		}
		ch := int(args[pos])
		if ch >= int('a') && ch <= int('z') {
			ch = ch - 32
		}
		if ch == int('T') || ch == int('O') || args[pos] == ' ' || args[pos] == '\t' {
			pos = pos + 1
		} else {
			break
		}
	}

	// Parse end value
	endStr := ""
	for {
		if pos >= len(args) {
			break
		}
		endStr = endStr + charToString(int(args[pos]))
		pos = pos + 1
	}

	startVal := parseNumber(trimSpacesStr(startStr))
	endVal := parseNumber(trimSpacesStr(endStr))

	return varName, startVal, endVal
}

// ParseNext parses a NEXT statement: NEXT var
// Returns the variable name
func ParseNext(args string) string {
	args = trimSpacesStr(args)
	if len(args) > 0 {
		return toUpperChar(int(args[0]))
	}
	return ""
}

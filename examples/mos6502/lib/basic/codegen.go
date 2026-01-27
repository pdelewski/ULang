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
// Uses cursor row from $30 to calculate screen address at runtime
// After printing, cursor moves to beginning of next line
// Uses $35:$36 for screen address pointer, $37 for temp high byte
func genPrint(args string, cursorRow int, cursorCol int, ctx CompileContext) ([]string, CompileContext) {
	lines := []string{}

	// Parse the string to print
	text := parseString(args)

	// Calculate base address from cursor row in $30
	// baseAddr = ScreenBase + $30 * 40
	// row * 40 = row * 8 + row * 32
	// Must handle 16-bit arithmetic properly for rows >= 8
	lines = append(lines, "LDA #$00")
	lines = append(lines, "STA $37")      // Initialize high byte temp to 0
	lines = append(lines, "STA $38")      // Clear row*8 storage

	lines = append(lines, "LDA $30")      // Load cursor row
	lines = append(lines, "ASL A")        // * 2
	lines = append(lines, "ASL A")        // * 4
	lines = append(lines, "ASL A")        // * 8
	lines = append(lines, "STA $38")      // Store row * 8 in $38 (for later)

	lines = append(lines, "ASL A")        // * 16 (may set carry for row >= 16)
	lines = append(lines, "ROL $37")      // Rotate carry into high byte
	lines = append(lines, "ASL A")        // * 32 (may set carry for row >= 8)
	lines = append(lines, "ROL $37")      // Rotate carry into high byte

	// Now A = (row * 32) low byte, $37 = (row * 32) high byte
	lines = append(lines, "CLC")
	lines = append(lines, "ADC $38")      // A = (row * 40) low byte (may carry)
	lines = append(lines, "STA $35")      // Store low byte of screen address
	lines = append(lines, "LDA $37")      // Load high byte of row*32
	lines = append(lines, "ADC #$04")     // Add carry + screen base high byte ($0400)
	lines = append(lines, "STA $36")      // High byte of screen address

	// Now $35:$36 contains the screen address for current row
	// Use Y for column offset, always start at column 0
	lines = append(lines, "LDY #$00")

	// Generate LDA/STA for each character using indirect indexed addressing
	i := 0
	for {
		if i >= len(text) {
			break
		}
		charCode := int(text[i])
		lines = append(lines, "LDA #"+toHex(charCode))
		lines = append(lines, "STA ($35),Y")
		lines = append(lines, "INY")
		i = i + 1
	}

	// Increment cursor row in zero page
	lines = append(lines, "INC $30")

	// Check if cursor row >= 25 (need to scroll)
	lines = append(lines, "LDA $30")
	lines = append(lines, "CMP #$19")
	lines = append(lines, "BCC print_no_scroll_"+intToString(ctx.LabelCounter))
	lines = append(lines, "JSR SCROLL_UP")
	lines = append(lines, "print_no_scroll_"+intToString(ctx.LabelCounter)+":")
	ctx.LabelCounter = ctx.LabelCounter + 1

	return lines, ctx
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

// genSimpleValue generates assembly to load a simple value (number or variable) into A
func genSimpleValue(valType int, valNum int, valVar string) []string {
	lines := []string{}
	if valType == ExprNumber {
		lines = append(lines, "LDA #"+toHex(valNum&0xFF))
	} else if valType == ExprVariable {
		addr := GetVariableAddress(valVar)
		if addr >= 0 {
			lines = append(lines, "LDA "+toHex(addr))
		}
	}
	return lines
}

// genExpression generates assembly to evaluate an expression
// Result is left in the A register
func genExpression(expr Expression) []string {
	lines := []string{}

	if expr.Type == ExprNumber {
		// Load immediate value
		lines = append(lines, "LDA #"+toHex(expr.Value&0xFF))
	} else if expr.Type == ExprVariable {
		// Load from variable address
		addr := GetVariableAddress(expr.VarName)
		if addr >= 0 {
			lines = append(lines, "LDA "+toHex(addr))
		}
	} else if expr.Type == ExprBinaryOp {
		// For binary operations with flat structure:
		// 1. Load right operand, push to stack
		// 2. Load left operand (stays in A)
		// 3. Pop right from stack and perform operation

		// Load right operand
		rightLines := genSimpleValue(expr.RightType, expr.RightValue, expr.RightVar)
		i := 0
		for {
			if i >= len(rightLines) {
				break
			}
			lines = append(lines, rightLines[i])
			i = i + 1
		}
		// Push right operand to stack
		lines = append(lines, "PHA")

		// Load left operand
		leftLines := genSimpleValue(expr.LeftType, expr.LeftValue, expr.LeftVar)
		i = 0
		for {
			if i >= len(leftLines) {
				break
			}
			lines = append(lines, leftLines[i])
			i = i + 1
		}

		// Pop right operand and perform operation
		// Store A temporarily, pull right to A, then perform op
		lines = append(lines, "STA $00") // Temp storage (left)
		lines = append(lines, "PLA")     // Right operand now in A
		lines = append(lines, "STA $01") // Store right in temp

		if expr.Op == "+" {
			lines = append(lines, "CLC")
			lines = append(lines, "LDA $00") // Left operand
			lines = append(lines, "ADC $01") // Add right
		} else if expr.Op == "-" {
			lines = append(lines, "SEC")
			lines = append(lines, "LDA $00") // Left operand
			lines = append(lines, "SBC $01") // Subtract right
		} else if expr.Op == "*" {
			// Simple multiplication using repeated addition
			lines = append(lines, "LDA $00") // Left (multiplicand)
			lines = append(lines, "STA $02")
			lines = append(lines, "LDA $01") // Right (multiplier)
			lines = append(lines, "STA $03")
			lines = append(lines, "LDA #$00") // Result = 0
			lines = append(lines, "STA $04")
			lines = append(lines, "LDX $03")     // Counter
			lines = append(lines, "BEQ mul_done") // If 0, skip
			lines = append(lines, "mul_loop:")
			lines = append(lines, "CLC")
			lines = append(lines, "LDA $04")
			lines = append(lines, "ADC $02")
			lines = append(lines, "STA $04")
			lines = append(lines, "DEX")
			lines = append(lines, "BNE mul_loop")
			lines = append(lines, "mul_done:")
			lines = append(lines, "LDA $04")
		} else if expr.Op == "/" {
			// Simple division using repeated subtraction
			lines = append(lines, "LDA $00") // Dividend
			lines = append(lines, "STA $02")
			lines = append(lines, "LDA $01") // Divisor
			lines = append(lines, "STA $03")
			lines = append(lines, "LDA #$00") // Quotient = 0
			lines = append(lines, "STA $04")
			lines = append(lines, "LDA $03")
			lines = append(lines, "BEQ div_done") // Avoid div by 0
			lines = append(lines, "div_loop:")
			lines = append(lines, "LDA $02")
			lines = append(lines, "CMP $03")
			lines = append(lines, "BCC div_done") // If dividend < divisor, done
			lines = append(lines, "SEC")
			lines = append(lines, "SBC $03")
			lines = append(lines, "STA $02")
			lines = append(lines, "INC $04")
			lines = append(lines, "JMP div_loop")
			lines = append(lines, "div_done:")
			lines = append(lines, "LDA $04")
		}
	}

	return lines
}

// genLet generates assembly for a LET statement
func genLet(varName string, expr Expression) []string {
	lines := []string{}

	// Generate code to evaluate expression (result in A)
	exprLines := genExpression(expr)
	i := 0
	for {
		if i >= len(exprLines) {
			break
		}
		lines = append(lines, exprLines[i])
		i = i + 1
	}

	// Store result in variable
	addr := GetVariableAddress(varName)
	if addr >= 0 {
		lines = append(lines, "STA "+toHex(addr))
	}

	return lines
}

// genGoto generates a JMP instruction
// The target address will be resolved in pass 2
// For now, we use a placeholder label format: LINE_xxx
func genGoto(lineNum int) []string {
	lines := []string{}
	lines = append(lines, "JMP LINE_"+intToString(lineNum))
	return lines
}

// genGosub generates a JSR instruction
func genGosub(lineNum int) []string {
	lines := []string{}
	lines = append(lines, "JSR LINE_"+intToString(lineNum))
	return lines
}

// genReturn generates an RTS instruction
func genReturn() []string {
	lines := []string{}
	lines = append(lines, "RTS")
	return lines
}

// genEnd generates a BRK instruction to halt the program
func genEnd() []string {
	lines := []string{}
	lines = append(lines, "BRK")
	return lines
}

// CompileContext holds compilation state that needs to be passed between functions
// (goany doesn't support package-level mutable variables)
type CompileContext struct {
	LabelCounter int
	ForLoopStack []ForLoopInfo
}

// NewCompileContext creates a new compilation context
func NewCompileContext() CompileContext {
	return CompileContext{
		LabelCounter: 0,
		ForLoopStack: []ForLoopInfo{},
	}
}

// nextLabel generates a unique label and returns updated context
func nextLabel(ctx CompileContext) (string, CompileContext) {
	ctx.LabelCounter = ctx.LabelCounter + 1
	return "L" + intToString(ctx.LabelCounter), ctx
}

// genCondition generates code to evaluate a condition
// After execution, the appropriate flags are set for branching
func genCondition(cond Condition) []string {
	lines := []string{}

	// Evaluate right expression, push to stack
	rightLines := genExpression(cond.Right)
	i := 0
	for {
		if i >= len(rightLines) {
			break
		}
		lines = append(lines, rightLines[i])
		i = i + 1
	}
	lines = append(lines, "PHA")

	// Evaluate left expression
	leftLines := genExpression(cond.Left)
	i = 0
	for {
		if i >= len(leftLines) {
			break
		}
		lines = append(lines, leftLines[i])
		i = i + 1
	}

	// Pop right and compare
	lines = append(lines, "STA $00") // Store left
	lines = append(lines, "PLA")     // Right in A
	lines = append(lines, "STA $01") // Store right
	lines = append(lines, "LDA $00") // Load left
	lines = append(lines, "CMP $01") // Compare with right

	return lines
}

// genIf generates code for an IF/THEN statement
func genIf(cond Condition, thenStmt string, ctx CompileContext) ([]string, CompileContext) {
	lines := []string{}

	// Generate condition evaluation
	condLines := genCondition(cond)
	i := 0
	for {
		if i >= len(condLines) {
			break
		}
		lines = append(lines, condLines[i])
		i = i + 1
	}

	// Generate conditional branch
	// CMP sets: Z flag if equal, C flag if A >= operand
	skipLabel := ""
	skipLabel, ctx = nextLabel(ctx)

	if cond.Op == CondEq {
		// If NOT equal, skip the THEN part
		lines = append(lines, "BNE "+skipLabel)
	} else if cond.Op == CondNe {
		// If equal, skip the THEN part
		lines = append(lines, "BEQ "+skipLabel)
	} else if cond.Op == CondLt {
		// If >= (C set), skip
		lines = append(lines, "BCS "+skipLabel)
	} else if cond.Op == CondGe {
		// If < (C clear), skip
		lines = append(lines, "BCC "+skipLabel)
	} else if cond.Op == CondGt {
		// If <= (Z set OR C clear), skip
		lines = append(lines, "BEQ "+skipLabel)
		lines = append(lines, "BCC "+skipLabel)
	} else if cond.Op == CondLe {
		// If > (Z clear AND C set), skip
		// This is tricky - need to check both conditions
		gtLabel := ""
		gtLabel, ctx = nextLabel(ctx)
		lines = append(lines, "BEQ "+gtLabel) // If equal, don't skip (LE includes equal)
		lines = append(lines, "BCS "+skipLabel) // If C set and not equal, it's greater, skip
		lines = append(lines, gtLabel+":")
	}

	// Generate the THEN statement
	thenLines := []string{}
	thenLines, ctx = compileLine(thenStmt, 0, 0, ctx) // Row/col not used for non-PRINT
	i = 0
	for {
		if i >= len(thenLines) {
			break
		}
		lines = append(lines, thenLines[i])
		i = i + 1
	}

	// Skip label
	lines = append(lines, skipLabel+":")

	return lines, ctx
}

// genFor generates code for a FOR statement
func genFor(varName string, startVal int, endVal int, ctx CompileContext) ([]string, CompileContext) {
	lines := []string{}

	// Initialize loop variable
	addr := GetVariableAddress(varName)
	if addr >= 0 {
		lines = append(lines, "LDA #"+toHex(startVal&0xFF))
		lines = append(lines, "STA "+toHex(addr))
	}

	// Generate loop start label (build the full label with colon directly)
	labelNum := ctx.LabelCounter
	loopLabelWithColon := "FOR_"
	loopLabelWithColon = loopLabelWithColon + varName
	loopLabelWithColon = loopLabelWithColon + "_"
	loopLabelWithColon = loopLabelWithColon + intToString(labelNum)
	loopLabelWithColon = loopLabelWithColon + ":"
	ctx.LabelCounter = ctx.LabelCounter + 1
	lines = append(lines, loopLabelWithColon)

	// Push loop info onto stack with the label number
	info := ForLoopInfo{
		VarName:  varName,
		StartVal: startVal,
		EndVal:   endVal,
		LoopAddr: 0, // Will be set during assembly
		LabelNum: labelNum,
	}
	ctx.ForLoopStack = append(ctx.ForLoopStack, info)

	return lines, ctx
}

// genNext generates code for a NEXT statement
func genNext(varName string, ctx CompileContext) ([]string, CompileContext) {
	lines := []string{}

	// Find the matching FOR loop
	loopIdx := -1
	i := len(ctx.ForLoopStack) - 1
	for {
		if i < 0 {
			break
		}
		if ctx.ForLoopStack[i].VarName == varName {
			loopIdx = i
			break
		}
		i = i - 1
	}

	if loopIdx < 0 {
		// No matching FOR found
		return lines, ctx
	}

	info := ctx.ForLoopStack[loopIdx]
	addr := GetVariableAddress(varName)

	if addr >= 0 {
		// Increment the loop variable
		lines = append(lines, "INC "+toHex(addr))

		// Load and compare with end value + 1
		lines = append(lines, "LDA "+toHex(addr))
		endPlusOne := (info.EndVal + 1) & 0xFF
		lines = append(lines, "CMP #"+toHex(endPlusOne))

		// If less than end+1, loop back (build label using +=)
		branchInstr := "BCC FOR_"
		branchInstr = branchInstr + varName
		branchInstr = branchInstr + "_"
		branchInstr = branchInstr + intToString(info.LabelNum)
		lines = append(lines, branchInstr)
	}

	// Pop from stack (rebuild without last element to avoid slice syntax issues)
	if loopIdx == len(ctx.ForLoopStack)-1 {
		newStack := []ForLoopInfo{}
		stackIdx := 0
		for {
			if stackIdx >= len(ctx.ForLoopStack)-1 {
				break
			}
			newStack = append(newStack, ctx.ForLoopStack[stackIdx])
			stackIdx = stackIdx + 1
		}
		ctx.ForLoopStack = newStack
	}

	return lines, ctx
}

// genPrintVar generates assembly to print a variable's value
// Uses cursor row from $30 to calculate screen address at runtime
// After printing, cursor moves to beginning of next line
// Uses $35:$36 for screen address pointer, $37 for temp high byte
func genPrintVar(varName string, cursorRow int, cursorCol int, ctx CompileContext) ([]string, CompileContext) {
	lines := []string{}

	addr := GetVariableAddress(varName)
	if addr < 0 {
		return lines, ctx
	}

	// Calculate base address from cursor row in $30
	// baseAddr = ScreenBase + $30 * 40
	// row * 40 = row * 8 + row * 32
	// Must handle 16-bit arithmetic properly for rows >= 8
	lines = append(lines, "LDA #$00")
	lines = append(lines, "STA $37")      // Initialize high byte temp to 0
	lines = append(lines, "STA $38")      // Clear row*8 storage

	lines = append(lines, "LDA $30")      // Load cursor row
	lines = append(lines, "ASL A")        // * 2
	lines = append(lines, "ASL A")        // * 4
	lines = append(lines, "ASL A")        // * 8
	lines = append(lines, "STA $38")      // Store row * 8 in $38 (for later)

	lines = append(lines, "ASL A")        // * 16 (may set carry for row >= 16)
	lines = append(lines, "ROL $37")      // Rotate carry into high byte
	lines = append(lines, "ASL A")        // * 32 (may set carry for row >= 8)
	lines = append(lines, "ROL $37")      // Rotate carry into high byte

	// Now A = (row * 32) low byte, $37 = (row * 32) high byte
	lines = append(lines, "CLC")
	lines = append(lines, "ADC $38")      // A = (row * 40) low byte (may carry)
	lines = append(lines, "STA $35")      // Store low byte of screen address
	lines = append(lines, "LDA $37")      // Load high byte of row*32
	lines = append(lines, "ADC #$04")     // Add carry + screen base high byte ($0400)
	lines = append(lines, "STA $36")      // High byte of screen address

	// Load variable value
	lines = append(lines, "LDA "+toHex(addr))

	// Convert to ASCII digit (works for 0-9)
	// Add '0' (0x30) to get ASCII
	lines = append(lines, "CLC")
	lines = append(lines, "ADC #$30")

	// Store to screen using indirect indexed addressing
	lines = append(lines, "LDY #$00")
	lines = append(lines, "STA ($35),Y")

	// Increment cursor row in zero page
	lines = append(lines, "INC $30")

	// Check if cursor row >= 25 (need to scroll)
	lines = append(lines, "LDA $30")
	lines = append(lines, "CMP #$19")
	lines = append(lines, "BCC printvar_no_scroll_"+intToString(ctx.LabelCounter))
	lines = append(lines, "JSR SCROLL_UP")
	lines = append(lines, "printvar_no_scroll_"+intToString(ctx.LabelCounter)+":")
	ctx.LabelCounter = ctx.LabelCounter + 1

	return lines, ctx
}

// GenScrollRoutine generates the scroll subroutine that scrolls the screen up by one line
// Uses zero page locations $40-$44 as temporary storage
// This should be placed at the end of the compiled code before BRK
func GenScrollRoutine() []string {
	lines := []string{}

	lines = append(lines, "JMP AFTER_SCROLL") // Skip over subroutine during normal execution

	lines = append(lines, "SCROLL_UP:")

	// Set up source pointer at $40:$41 = $0428 (row 1)
	lines = append(lines, "LDA #$28")
	lines = append(lines, "STA $40")
	lines = append(lines, "LDA #$04")
	lines = append(lines, "STA $41")

	// Set up dest pointer at $42:$43 = $0400 (row 0)
	lines = append(lines, "LDA #$00")
	lines = append(lines, "STA $42")
	lines = append(lines, "LDA #$04")
	lines = append(lines, "STA $43")

	// Counter for rows (24 rows to copy)
	lines = append(lines, "LDA #$18") // 24 decimal
	lines = append(lines, "STA $44")

	lines = append(lines, "SCROLL_ROW:")
	// Copy 40 bytes from source to dest
	lines = append(lines, "LDY #$27") // 39 decimal
	lines = append(lines, "SCROLL_BYTE:")
	lines = append(lines, "LDA ($40),Y")
	lines = append(lines, "STA ($42),Y")
	lines = append(lines, "DEY")
	lines = append(lines, "BPL SCROLL_BYTE")

	// Advance source pointer by 40
	lines = append(lines, "CLC")
	lines = append(lines, "LDA $40")
	lines = append(lines, "ADC #$28") // 40 decimal
	lines = append(lines, "STA $40")
	lines = append(lines, "LDA $41")
	lines = append(lines, "ADC #$00")
	lines = append(lines, "STA $41")

	// Advance dest pointer by 40
	lines = append(lines, "CLC")
	lines = append(lines, "LDA $42")
	lines = append(lines, "ADC #$28") // 40 decimal
	lines = append(lines, "STA $42")
	lines = append(lines, "LDA $43")
	lines = append(lines, "ADC #$00")
	lines = append(lines, "STA $43")

	// Decrement row counter
	lines = append(lines, "DEC $44")
	lines = append(lines, "BNE SCROLL_ROW")

	// Clear last row (row 24)
	// Address = $0400 + 24*40 = $0400 + 960 = $07C0
	lines = append(lines, "LDY #$27") // 39 decimal
	lines = append(lines, "LDA #$20") // Space character
	lines = append(lines, "CLEAR_LAST:")
	lines = append(lines, "STA $07C0,Y")
	lines = append(lines, "DEY")
	lines = append(lines, "BPL CLEAR_LAST")

	// Set cursor row to 24
	lines = append(lines, "LDA #$18") // 24 decimal
	lines = append(lines, "STA $30")

	lines = append(lines, "RTS")

	lines = append(lines, "AFTER_SCROLL:")

	return lines
}

package basic

import (
	"mos6502lib/assembler"
)

// Screen constants (C64 layout)
const TextCols = 40
const TextRows = 25
const ScreenBase = 0x0400
const CodeBase = 0xC000

// Zero page address for cursor row tracking (after variables $10-$29)
const CursorRowAddr = 0x30

// ProgramLine stores one line of BASIC program
type ProgramLine struct {
	LineNum int
	Text    string
}

// BasicState holds interpreter state
type BasicState struct {
	Lines     []ProgramLine
	CursorRow int
	CursorCol int
}

// NewBasicState creates a new interpreter state
func NewBasicState() BasicState {
	return BasicState{
		Lines:     []ProgramLine{},
		CursorRow: 0,
		CursorCol: 0,
	}
}

// SetCursor updates cursor position
func SetCursor(state BasicState, row int, col int) BasicState {
	state.CursorRow = row
	state.CursorCol = col
	return state
}

// GetCursorAddr returns the screen memory address for current cursor position
func GetCursorAddr(state BasicState) int {
	return ScreenBase + (state.CursorRow * TextCols) + state.CursorCol
}

// StoreLine stores or updates a program line
func StoreLine(state BasicState, lineNum int, text string) BasicState {
	// Check if line already exists and rebuild the lines slice
	newLines := []ProgramLine{}
	found := false
	i := 0
	for {
		if i >= len(state.Lines) {
			break
		}
		if state.Lines[i].LineNum == lineNum {
			// Replace with updated line
			newLines = append(newLines, ProgramLine{LineNum: lineNum, Text: text})
			found = true
		} else {
			newLines = append(newLines, state.Lines[i])
		}
		i = i + 1
	}

	if !found {
		// Add new line
		newLines = append(newLines, ProgramLine{LineNum: lineNum, Text: text})
	}

	state.Lines = newLines
	// Sort lines by line number
	state = sortLines(state)

	return state
}

// sortLines sorts program lines by line number
func sortLines(state BasicState) BasicState {
	// Simple bubble sort (goany compatible)
	n := len(state.Lines)
	i := 0
	for {
		if i >= n-1 {
			break
		}
		j := 0
		for {
			if j >= n-i-1 {
				break
			}
			if state.Lines[j].LineNum > state.Lines[j+1].LineNum {
				// Swap
				temp := state.Lines[j]
				state.Lines[j] = state.Lines[j+1]
				state.Lines[j+1] = temp
			}
			j = j + 1
		}
		i = i + 1
	}
	return state
}

// DeleteLine removes a program line
func DeleteLine(state BasicState, lineNum int) BasicState {
	newLines := []ProgramLine{}
	i := 0
	for {
		if i >= len(state.Lines) {
			break
		}
		if state.Lines[i].LineNum != lineNum {
			newLines = append(newLines, state.Lines[i])
		}
		i = i + 1
	}
	state.Lines = newLines
	return state
}

// ClearProgram removes all program lines
func ClearProgram(state BasicState) BasicState {
	state.Lines = []ProgramLine{}
	return state
}

// CompileImmediate compiles a single line for immediate execution
func CompileImmediate(state BasicState, line string) []uint8 {
	ctx := NewCompileContext()
	asmLines := []string{}

	// Initialize cursor row in zero page (used by PRINT for runtime address calculation)
	asmLines = append(asmLines, "LDA #$"+toHex2(state.CursorRow))
	asmLines = append(asmLines, "STA $30")

	// Initialize temporary zero-page locations used by PRINT
	asmLines = append(asmLines, "LDA #$00")
	asmLines = append(asmLines, "STA $35")
	asmLines = append(asmLines, "STA $36")
	asmLines = append(asmLines, "STA $37")
	asmLines = append(asmLines, "STA $38")
	asmLines = append(asmLines, "CLC")

	lineAsm := []string{}
	lineAsm, ctx = compileLine(line, state.CursorRow, state.CursorCol, ctx)
	// Append all generated lines
	j := 0
	for {
		if j >= len(lineAsm) {
			break
		}
		asmLines = append(asmLines, lineAsm[j])
		j = j + 1
	}
	// Use ctx.LabelCounter to avoid unused variable warning (goany compatible)
	if ctx.LabelCounter < 0 {
		ctx.LabelCounter = 0
	}
	asmLines = append(asmLines, "BRK")
	return assembler.AssembleLines(asmLines)
}

// CompileProgram compiles all stored program lines
func CompileProgram(state BasicState) []uint8 {
	// Create fresh compilation context
	ctx := NewCompileContext()

	asmLines := []string{}

	// Initialize X register to 0 for cursor offset
	asmLines = append(asmLines, "LDX #$00")

	// Initialize cursor row in zero page
	asmLines = append(asmLines, "LDA #$"+toHex2(state.CursorRow))
	asmLines = append(asmLines, "STA $30")

	// Initialize temporary zero-page locations used by PRINT
	// Uses $35:$36 for screen address, $37:$38 for temps
	// This ensures clean state between runs (important for JS backend)
	asmLines = append(asmLines, "LDA #$00")
	asmLines = append(asmLines, "STA $35")
	asmLines = append(asmLines, "STA $36")
	asmLines = append(asmLines, "STA $37")
	asmLines = append(asmLines, "STA $38")

	// Clear carry flag to ensure clean arithmetic state
	asmLines = append(asmLines, "CLC")

	row := state.CursorRow
	col := 0

	i := 0
	for {
		if i >= len(state.Lines) {
			break
		}
		// Add line number label for GOTO/GOSUB targets
		lineLabel := "LINE_" + intToString(state.Lines[i].LineNum) + ":"
		asmLines = append(asmLines, lineLabel)

		lineAsm := []string{}
		lineAsm, ctx = compileLine(state.Lines[i].Text, row, col, ctx)
		// Append all assembly lines
		j := 0
		for {
			if j >= len(lineAsm) {
				break
			}
			asmLines = append(asmLines, lineAsm[j])
			j = j + 1
		}
		row = row + 1
		if row >= TextRows {
			row = TextRows - 1
		}
		i = i + 1
	}

	asmLines = append(asmLines, "BRK")
	return assembler.AssembleLines(asmLines)
}

// CompileProgramDebug compiles and returns asm line count, instruction count, and last opcode first byte
func CompileProgramDebug(state BasicState) ([]uint8, int, int, int) {
	// Create fresh compilation context
	ctx := NewCompileContext()

	asmLines := []string{}

	// Initialize X register to 0 for cursor offset
	asmLines = append(asmLines, "LDX #$00")

	row := state.CursorRow
	col := 0

	i := 0
	for {
		if i >= len(state.Lines) {
			break
		}
		// Add line number label for GOTO/GOSUB targets
		lineLabel := "LINE_" + intToString(state.Lines[i].LineNum) + ":"
		asmLines = append(asmLines, lineLabel)

		lineAsm := []string{}
		lineAsm, ctx = compileLine(state.Lines[i].Text, row, col, ctx)
		j := 0
		for {
			if j >= len(lineAsm) {
				break
			}
			asmLines = append(asmLines, lineAsm[j])
			j = j + 1
		}
		row = row + 1
		if row >= TextRows {
			row = TextRows - 1
		}
		i = i + 1
	}

	asmLines = append(asmLines, "BRK")
	code, instrCount := assembler.AssembleLinesWithCount(asmLines)
	lastByte := assembler.GetLastInstrFirstByte(asmLines)
	return code, len(asmLines), instrCount, lastByte
}

// compileLine compiles a single BASIC line to assembly
func compileLine(line string, cursorRow int, cursorCol int, ctx CompileContext) ([]string, CompileContext) {
	// Parse the line
	cmd, args := parseLine(line)

	if cmd == "PRINT" {
		// Check if printing a variable (single letter) or a string
		trimmedArgs := trimSpacesStr(args)
		if IsVariableName(trimmedArgs) {
			return genPrintVar(trimmedArgs, cursorRow, cursorCol, ctx)
		}
		return genPrint(args, cursorRow, cursorCol, ctx)
	} else if cmd == "POKE" {
		addr, value := parsePoke(args)
		return genPoke(addr, value), ctx
	} else if cmd == "CLR" {
		return genClear(), ctx
	} else if cmd == "LET" {
		varName, expr := ParseLet(args)
		return genLet(varName, expr), ctx
	} else if cmd == "GOTO" {
		lineNum := ParseGoto(args)
		return genGoto(lineNum), ctx
	} else if cmd == "GOSUB" {
		lineNum := ParseGosub(args)
		return genGosub(lineNum), ctx
	} else if cmd == "RETURN" {
		return genReturn(), ctx
	} else if cmd == "IF" {
		cond, thenStmt := ParseIf(args)
		return genIf(cond, thenStmt, ctx)
	} else if cmd == "FOR" {
		varName, startVal, endVal := ParseFor(args)
		return genFor(varName, startVal, endVal, ctx)
	} else if cmd == "NEXT" {
		varName := ParseNext(args)
		return genNext(varName, ctx)
	} else if cmd == "REM" {
		// Comment - generate no code
		return []string{}, ctx
	} else if cmd == "END" {
		return genEnd(), ctx
	}

	// Check if it's a variable assignment without LET (e.g., "A = 5")
	if len(line) > 0 {
		trimmed := trimSpacesStr(line)
		if len(trimmed) >= 3 {
			firstChar := trimmed[0]
			if (firstChar >= 'A' && firstChar <= 'Z') || (firstChar >= 'a' && firstChar <= 'z') {
				// Check for = sign
				hasEquals := false
				i := 1
				for {
					if i >= len(trimmed) {
						break
					}
					if trimmed[i] == '=' {
						hasEquals = true
						break
					}
					if trimmed[i] != ' ' && trimmed[i] != '\t' {
						break
					}
					i = i + 1
				}
				if hasEquals {
					varName, expr := ParseLet(trimmed)
					return genLet(varName, expr), ctx
				}
			}
		}
	}

	// Unknown command - return empty
	return []string{}, ctx
}

// GetLineCount returns number of stored program lines
func GetLineCount(state BasicState) int {
	return len(state.Lines)
}

// GetLine returns a program line by index
func GetLine(state BasicState, index int) ProgramLine {
	if index >= 0 && index < len(state.Lines) {
		return state.Lines[index]
	}
	return ProgramLine{LineNum: 0, Text: ""}
}

// GetCursorRowAddr returns the zero page address where cursor row is stored
func GetCursorRowAddr() int {
	return CursorRowAddr
}

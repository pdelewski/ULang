package basic

import (
	"mos6502lib/assembler"
)

// Screen constants (C64 layout)
const TextCols = 40
const TextRows = 25
const ScreenBase = 0x0400
const CodeBase = 0xC000

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
	asmLines := compileLine(line, state.CursorRow, state.CursorCol)
	asmLines = append(asmLines, "BRK")
	return assembler.AssembleLines(asmLines)
}

// CompileProgram compiles all stored program lines
func CompileProgram(state BasicState) []uint8 {
	asmLines := []string{}
	row := state.CursorRow
	col := 0

	i := 0
	for {
		if i >= len(state.Lines) {
			break
		}
		lineAsm := compileLine(state.Lines[i].Text, row, col)
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

// compileLine compiles a single BASIC line to assembly
func compileLine(line string, cursorRow int, cursorCol int) []string {
	// Parse the line
	cmd, args := parseLine(line)

	if cmd == "PRINT" {
		return genPrint(args, cursorRow, cursorCol)
	} else if cmd == "POKE" {
		addr, value := parsePoke(args)
		return genPoke(addr, value)
	} else if cmd == "CLR" {
		return genClear()
	}

	// Unknown command - return empty
	return []string{}
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

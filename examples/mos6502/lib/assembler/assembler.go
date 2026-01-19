package assembler

import (
	"mos6502lib/cpu"
)

// Token types for 6502 assembly
const (
	TokenTypeInstruction = 1
	TokenTypeNumber      = 2
	TokenTypeLabel       = 3
	TokenTypeComma       = 4
	TokenTypeNewline     = 5
	TokenTypeHash        = 6  // # for immediate mode
	TokenTypeDollar      = 7  // $ for hex numbers
	TokenTypeColon       = 8  // : for labels
	TokenTypeIdentifier  = 9  // Generic identifier
	TokenTypeComment     = 10 // ; comment
)

// Addressing modes
const (
	ModeImplied   = 0
	ModeImmediate = 1
	ModeZeroPage  = 2
	ModeAbsolute  = 3
	ModeZeroPageX = 4
)

// Token represents a lexical token
type Token struct {
	Type           int8
	Representation []int8
}

// Instruction represents a parsed assembly instruction
type Instruction struct {
	OpcodeBytes []int8
	Mode        int8 // 0=implied, 1=immediate, 2=zeropage, 3=absolute, 4=zeropage,X
	Operand     int  // The operand value
	LabelBytes  []int8
	HasLabel    bool
}

// IsDigit checks if a character is a digit
func IsDigit(b int8) bool {
	return b >= '0' && b <= '9'
}

// IsHexDigit checks if a character is a hex digit
func IsHexDigit(b int8) bool {
	return IsDigit(b) || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')
}

// IsAlpha checks if a character is alphabetic
func IsAlpha(b int8) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_'
}

// IsWhitespace checks if a character is whitespace (but not newline)
func IsWhitespace(b int8) bool {
	return b == ' ' || b == '\t'
}

// StringToBytes converts a string to int8 slice
func StringToBytes(s string) []int8 {
	result := []int8{}
	i := 0
	for {
		if i >= len(s) {
			break
		}
		result = append(result, int8(s[i]))
		i = i + 1
	}
	return result
}

// ToUpper converts a byte to uppercase
func ToUpper(b int8) int8 {
	if b >= 'a' && b <= 'z' {
		return b - 32
	}
	return b
}

// Tokenize converts assembly text into tokens
func Tokenize(text string) []Token {
	tokens := []Token{}
	bytes := StringToBytes(text)
	i := 0

	for {
		if i >= len(bytes) {
			break
		}
		b := bytes[i]

		// Skip whitespace (not newline)
		if IsWhitespace(b) {
			i = i + 1
			continue
		}

		// Newline
		if b == '\n' {
			tokens = append(tokens, Token{Type: TokenTypeNewline, Representation: []int8{b}})
			i = i + 1
			continue
		}

		// Comment - skip to end of line
		if b == ';' {
			for {
				if i >= len(bytes) {
					break
				}
				if bytes[i] == '\n' {
					break
				}
				i = i + 1
			}
			continue
		}

		// Hash (immediate mode indicator)
		if b == '#' {
			tokens = append(tokens, Token{Type: TokenTypeHash, Representation: []int8{b}})
			i = i + 1
			continue
		}

		// Dollar (hex number indicator)
		if b == '$' {
			tokens = append(tokens, Token{Type: TokenTypeDollar, Representation: []int8{b}})
			i = i + 1
			continue
		}

		// Colon (label definition)
		if b == ':' {
			tokens = append(tokens, Token{Type: TokenTypeColon, Representation: []int8{b}})
			i = i + 1
			continue
		}

		// Comma
		if b == ',' {
			tokens = append(tokens, Token{Type: TokenTypeComma, Representation: []int8{b}})
			i = i + 1
			continue
		}

		// Number (decimal)
		if IsDigit(b) {
			repr := []int8{}
			for {
				if i >= len(bytes) {
					break
				}
				if !IsDigit(bytes[i]) {
					break
				}
				repr = append(repr, bytes[i])
				i = i + 1
			}
			tokens = append(tokens, Token{Type: TokenTypeNumber, Representation: repr})
			continue
		}

		// Hex number (after $)
		if IsHexDigit(b) {
			repr := []int8{}
			for {
				if i >= len(bytes) {
					break
				}
				if !IsHexDigit(bytes[i]) {
					break
				}
				repr = append(repr, bytes[i])
				i = i + 1
			}
			tokens = append(tokens, Token{Type: TokenTypeNumber, Representation: repr})
			continue
		}

		// Identifier (instruction or label)
		if IsAlpha(b) {
			repr := []int8{}
			for {
				if i >= len(bytes) {
					break
				}
				if !IsAlpha(bytes[i]) && !IsDigit(bytes[i]) {
					break
				}
				repr = append(repr, bytes[i])
				i = i + 1
			}
			tokens = append(tokens, Token{Type: TokenTypeIdentifier, Representation: repr})
			continue
		}

		// Unknown character, skip
		i = i + 1
	}

	return tokens
}

// ParseHex parses a hex string to integer
func ParseHex(bytes []int8) int {
	result := 0
	i := 0
	for {
		if i >= len(bytes) {
			break
		}
		b := bytes[i]
		result = result * 16
		if b >= '0' && b <= '9' {
			result = result + int(b-'0')
		} else if b >= 'a' && b <= 'f' {
			result = result + int(b-'a'+10)
		} else if b >= 'A' && b <= 'F' {
			result = result + int(b-'A'+10)
		}
		i = i + 1
	}
	return result
}

// ParseDecimal parses a decimal string to integer
func ParseDecimal(bytes []int8) int {
	result := 0
	i := 0
	for {
		if i >= len(bytes) {
			break
		}
		b := bytes[i]
		result = result*10 + int(b-'0')
		i = i + 1
	}
	return result
}

// MatchToken checks if token matches a string (case insensitive)
func MatchToken(token Token, s string) bool {
	if len(token.Representation) != len(s) {
		return false
	}
	i := 0
	for {
		if i >= len(s) {
			break
		}
		if ToUpper(token.Representation[i]) != ToUpper(int8(s[i])) {
			return false
		}
		i = i + 1
	}
	return true
}

// CopyBytes copies a byte slice
func CopyBytes(src []int8) []int8 {
	dst := []int8{}
	i := 0
	for {
		if i >= len(src) {
			break
		}
		dst = append(dst, src[i])
		i = i + 1
	}
	return dst
}

// Parse converts tokens into instructions
func Parse(tokens []Token) []Instruction {
	instructions := []Instruction{}
	i := 0

	for {
		if i >= len(tokens) {
			break
		}

		// Skip newlines
		if tokens[i].Type == TokenTypeNewline {
			i = i + 1
			continue
		}

		// Check for label definition (identifier followed by colon)
		currentLabelBytes := []int8{}
		hasLabel := false
		if tokens[i].Type == TokenTypeIdentifier && i+1 < len(tokens) && tokens[i+1].Type == TokenTypeColon {
			currentLabelBytes = CopyBytes(tokens[i].Representation)
			hasLabel = true
			i = i + 2 // Skip label and colon
			// Skip any whitespace/newlines after label
			for {
				if i >= len(tokens) {
					break
				}
				if tokens[i].Type != TokenTypeNewline {
					break
				}
				i = i + 1
			}
			if i >= len(tokens) {
				break
			}
		}

		// Expect instruction
		if tokens[i].Type != TokenTypeIdentifier {
			i = i + 1
			continue
		}

		instr := Instruction{
			OpcodeBytes: CopyBytes(tokens[i].Representation),
			Mode:        ModeImplied,
			Operand:     0,
			LabelBytes:  currentLabelBytes,
			HasLabel:    hasLabel,
		}
		i = i + 1

		// Check for operand
		if i < len(tokens) && tokens[i].Type != TokenTypeNewline {
			// Immediate mode: #$XX or #NN
			if tokens[i].Type == TokenTypeHash {
				i = i + 1
				instr.Mode = ModeImmediate
				if i < len(tokens) && tokens[i].Type == TokenTypeDollar {
					i = i + 1
					if i < len(tokens) && tokens[i].Type == TokenTypeNumber {
						instr.Operand = ParseHex(tokens[i].Representation)
						i = i + 1
					}
				} else if i < len(tokens) && tokens[i].Type == TokenTypeNumber {
					instr.Operand = ParseDecimal(tokens[i].Representation)
					i = i + 1
				}
			} else if tokens[i].Type == TokenTypeDollar {
				// Absolute or ZeroPage: $XXXX or $XX
				i = i + 1
				if i < len(tokens) && tokens[i].Type == TokenTypeNumber {
					instr.Operand = ParseHex(tokens[i].Representation)
					// Check for ,X indexing
					if len(tokens[i].Representation) <= 2 {
						instr.Mode = ModeZeroPage
					} else {
						instr.Mode = ModeAbsolute
					}
					i = i + 1
					// Check for ,X
					if i < len(tokens) && tokens[i].Type == TokenTypeComma {
						i = i + 1
						if i < len(tokens) && tokens[i].Type == TokenTypeIdentifier {
							if MatchToken(tokens[i], "X") {
								instr.Mode = ModeZeroPageX
							}
							i = i + 1
						}
					}
				}
			} else if tokens[i].Type == TokenTypeNumber {
				// Decimal number
				instr.Operand = ParseDecimal(tokens[i].Representation)
				if instr.Operand <= 255 {
					instr.Mode = ModeZeroPage
				} else {
					instr.Mode = ModeAbsolute
				}
				i = i + 1
			}
		}

		instructions = append(instructions, instr)
	}

	return instructions
}

// IsOpcode checks if the opcode bytes match a given instruction name
func IsOpcode(opcodeBytes []int8, name string) bool {
	if len(opcodeBytes) != len(name) {
		return false
	}
	i := 0
	for {
		if i >= len(name) {
			break
		}
		ob := opcodeBytes[i]
		// Convert to uppercase
		if ob >= 'a' && ob <= 'z' {
			ob = ob - 32
		}
		nb := int8(name[i])
		if ob != nb {
			return false
		}
		i = i + 1
	}
	return true
}

// Assemble converts parsed instructions to machine code
func Assemble(instructions []Instruction) []uint8 {
	code := []uint8{}

	idx := 0
	for {
		if idx >= len(instructions) {
			break
		}
		instr := instructions[idx]
		opcodeBytes := instr.OpcodeBytes

		// Match instruction and emit bytes
		if IsOpcode(opcodeBytes, "LDA") {
			if instr.Mode == ModeImmediate {
				code = append(code, uint8(cpu.OpLDAImm))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeZeroPage {
				code = append(code, uint8(cpu.OpLDAZp))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeZeroPageX {
				code = append(code, uint8(cpu.OpLDAZpX))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeAbsolute {
				code = append(code, uint8(cpu.OpLDAAbs))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			}
		} else if IsOpcode(opcodeBytes, "LDX") {
			if instr.Mode == ModeImmediate {
				code = append(code, uint8(cpu.OpLDXImm))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeAbsolute {
				code = append(code, uint8(cpu.OpLDXAbs))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			}
		} else if IsOpcode(opcodeBytes, "LDY") {
			if instr.Mode == ModeImmediate {
				code = append(code, uint8(cpu.OpLDYImm))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeAbsolute {
				code = append(code, uint8(cpu.OpLDYAbs))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			}
		} else if IsOpcode(opcodeBytes, "STA") {
			if instr.Mode == ModeZeroPage {
				code = append(code, uint8(cpu.OpSTAZp))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeZeroPageX {
				code = append(code, uint8(cpu.OpSTAZpX))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeAbsolute {
				code = append(code, uint8(cpu.OpSTAAbs))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			}
		} else if IsOpcode(opcodeBytes, "STX") {
			if instr.Mode == ModeAbsolute {
				code = append(code, uint8(cpu.OpSTXAbs))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			}
		} else if IsOpcode(opcodeBytes, "STY") {
			if instr.Mode == ModeAbsolute {
				code = append(code, uint8(cpu.OpSTYAbs))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			}
		} else if IsOpcode(opcodeBytes, "ADC") {
			if instr.Mode == ModeImmediate {
				code = append(code, uint8(cpu.OpADCImm))
				code = append(code, uint8(instr.Operand))
			}
		} else if IsOpcode(opcodeBytes, "SBC") {
			if instr.Mode == ModeImmediate {
				code = append(code, uint8(cpu.OpSBCImm))
				code = append(code, uint8(instr.Operand))
			}
		} else if IsOpcode(opcodeBytes, "INX") {
			code = append(code, uint8(cpu.OpINX))
		} else if IsOpcode(opcodeBytes, "INY") {
			code = append(code, uint8(cpu.OpINY))
		} else if IsOpcode(opcodeBytes, "DEX") {
			code = append(code, uint8(cpu.OpDEX))
		} else if IsOpcode(opcodeBytes, "DEY") {
			code = append(code, uint8(cpu.OpDEY))
		} else if IsOpcode(opcodeBytes, "INC") {
			if instr.Mode == ModeZeroPage {
				code = append(code, uint8(cpu.OpINC))
				code = append(code, uint8(instr.Operand))
			}
		} else if IsOpcode(opcodeBytes, "CMP") {
			if instr.Mode == ModeImmediate {
				code = append(code, uint8(cpu.OpCMPImm))
				code = append(code, uint8(instr.Operand))
			}
		} else if IsOpcode(opcodeBytes, "CPX") {
			if instr.Mode == ModeImmediate {
				code = append(code, uint8(cpu.OpCPXImm))
				code = append(code, uint8(instr.Operand))
			}
		} else if IsOpcode(opcodeBytes, "CPY") {
			if instr.Mode == ModeImmediate {
				code = append(code, uint8(cpu.OpCPYImm))
				code = append(code, uint8(instr.Operand))
			}
		} else if IsOpcode(opcodeBytes, "BNE") {
			code = append(code, uint8(cpu.OpBNE))
			code = append(code, uint8(instr.Operand))
		} else if IsOpcode(opcodeBytes, "BEQ") {
			code = append(code, uint8(cpu.OpBEQ))
			code = append(code, uint8(instr.Operand))
		} else if IsOpcode(opcodeBytes, "BCC") {
			code = append(code, uint8(cpu.OpBCC))
			code = append(code, uint8(instr.Operand))
		} else if IsOpcode(opcodeBytes, "BCS") {
			code = append(code, uint8(cpu.OpBCS))
			code = append(code, uint8(instr.Operand))
		} else if IsOpcode(opcodeBytes, "JMP") {
			code = append(code, uint8(cpu.OpJMP))
			code = append(code, uint8(instr.Operand&0xFF))
			code = append(code, uint8((instr.Operand>>8)&0xFF))
		} else if IsOpcode(opcodeBytes, "JSR") {
			code = append(code, uint8(cpu.OpJSR))
			code = append(code, uint8(instr.Operand&0xFF))
			code = append(code, uint8((instr.Operand>>8)&0xFF))
		} else if IsOpcode(opcodeBytes, "RTS") {
			code = append(code, uint8(cpu.OpRTS))
		} else if IsOpcode(opcodeBytes, "NOP") {
			code = append(code, uint8(cpu.OpNOP))
		} else if IsOpcode(opcodeBytes, "BRK") {
			code = append(code, uint8(cpu.OpBRK))
		}

		idx = idx + 1
	}

	return code
}

// AssembleString is a convenience function that tokenizes, parses, and assembles
func AssembleString(text string) []uint8 {
	tokens := Tokenize(text)
	instructions := Parse(tokens)
	return Assemble(instructions)
}

// AppendLineBytes appends a line's bytes to allBytes
func AppendLineBytes(allBytes []int8, lineBytes []int8) []int8 {
	j := 0
	for {
		if j >= len(lineBytes) {
			break
		}
		allBytes = append(allBytes, lineBytes[j])
		j = j + 1
	}
	return allBytes
}

// TokenizeBytes converts assembly bytes into tokens
func TokenizeBytes(bytes []int8) []Token {
	tokens := []Token{}
	i := 0

	for {
		if i >= len(bytes) {
			break
		}
		b := bytes[i]

		// Skip whitespace (not newline)
		if IsWhitespace(b) {
			i = i + 1
			continue
		}

		// Newline
		if b == '\n' {
			tokens = append(tokens, Token{Type: TokenTypeNewline, Representation: []int8{b}})
			i = i + 1
			continue
		}

		// Comment - skip to end of line
		if b == ';' {
			for {
				if i >= len(bytes) {
					break
				}
				if bytes[i] == '\n' {
					break
				}
				i = i + 1
			}
			continue
		}

		// Hash (immediate mode indicator)
		if b == '#' {
			tokens = append(tokens, Token{Type: TokenTypeHash, Representation: []int8{b}})
			i = i + 1
			continue
		}

		// Dollar (hex number indicator)
		if b == '$' {
			tokens = append(tokens, Token{Type: TokenTypeDollar, Representation: []int8{b}})
			i = i + 1
			continue
		}

		// Colon (label definition)
		if b == ':' {
			tokens = append(tokens, Token{Type: TokenTypeColon, Representation: []int8{b}})
			i = i + 1
			continue
		}

		// Comma
		if b == ',' {
			tokens = append(tokens, Token{Type: TokenTypeComma, Representation: []int8{b}})
			i = i + 1
			continue
		}

		// Number (decimal)
		if IsDigit(b) {
			repr := []int8{}
			for {
				if i >= len(bytes) {
					break
				}
				if !IsDigit(bytes[i]) {
					break
				}
				repr = append(repr, bytes[i])
				i = i + 1
			}
			tokens = append(tokens, Token{Type: TokenTypeNumber, Representation: repr})
			continue
		}

		// Hex number (after $)
		if IsHexDigit(b) {
			repr := []int8{}
			for {
				if i >= len(bytes) {
					break
				}
				if !IsHexDigit(bytes[i]) {
					break
				}
				repr = append(repr, bytes[i])
				i = i + 1
			}
			tokens = append(tokens, Token{Type: TokenTypeNumber, Representation: repr})
			continue
		}

		// Identifier (instruction or label)
		if IsAlpha(b) {
			repr := []int8{}
			for {
				if i >= len(bytes) {
					break
				}
				if !IsAlpha(bytes[i]) && !IsDigit(bytes[i]) {
					break
				}
				repr = append(repr, bytes[i])
				i = i + 1
			}
			tokens = append(tokens, Token{Type: TokenTypeIdentifier, Representation: repr})
			continue
		}

		// Unknown character, skip
		i = i + 1
	}

	return tokens
}

// AssembleLines assembles a slice of assembly lines
func AssembleLines(lines []string) []uint8 {
	// Build combined bytes with newlines
	allBytes := []int8{}
	i := 0
	for {
		if i >= len(lines) {
			break
		}
		lineBytes := StringToBytes(lines[i])
		allBytes = AppendLineBytes(allBytes, lineBytes)
		if i < len(lines)-1 {
			allBytes = append(allBytes, int8(10))
		}
		i = i + 1
	}
	tokens := TokenizeBytes(allBytes)
	instructions := Parse(tokens)
	return Assemble(instructions)
}

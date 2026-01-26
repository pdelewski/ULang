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
	TokenTypeLParen      = 11 // ( for indirect addressing
	TokenTypeRParen      = 12 // ) for indirect addressing
)

// Addressing modes
const (
	ModeImplied     = 0
	ModeImmediate   = 1
	ModeZeroPage    = 2
	ModeAbsolute    = 3
	ModeZeroPageX   = 4
	ModeZeroPageY   = 5
	ModeAbsoluteX   = 6
	ModeAbsoluteY   = 7
	ModeIndirectX   = 8  // (zp,X)
	ModeIndirectY   = 9  // (zp),Y
	ModeAccumulator = 10 // for ASL A, LSR A, etc.
	ModeIndirect    = 11 // for JMP ($addr)
)

// Token represents a lexical token
type Token struct {
	Type           int8
	Representation []int8
}

// Instruction represents a parsed assembly instruction
type Instruction struct {
	OpcodeBytes    []int8
	Mode           int8 // 0=implied, 1=immediate, 2=zeropage, 3=absolute, 4=zeropage,X
	Operand        int  // The operand value
	LabelBytes     []int8
	HasLabel       bool
	TargetLabel    []int8 // Label reference for JMP/JSR/branch
	HasTargetLabel bool
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

		// Left parenthesis (indirect addressing)
		if b == '(' {
			tokens = append(tokens, Token{Type: TokenTypeLParen, Representation: []int8{b}})
			i = i + 1
			continue
		}

		// Right parenthesis (indirect addressing)
		if b == ')' {
			tokens = append(tokens, Token{Type: TokenTypeRParen, Representation: []int8{b}})
			i = i + 1
			continue
		}

		// Number - must start with 0-9 (hex letters A-F only valid after initial digit)
		if IsDigit(b) {
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
			// Indirect addressing: ($XX),Y or ($XX,X)
			if tokens[i].Type == TokenTypeLParen {
				i = i + 1
				if i < len(tokens) && tokens[i].Type == TokenTypeDollar {
					i = i + 1
					if i < len(tokens) && tokens[i].Type == TokenTypeNumber {
						instr.Operand = ParseHex(tokens[i].Representation)
						i = i + 1
						// Check for ,X) - indexed indirect
						if i < len(tokens) && tokens[i].Type == TokenTypeComma {
							i = i + 1
							if i < len(tokens) && tokens[i].Type == TokenTypeIdentifier && MatchToken(tokens[i], "X") {
								i = i + 1
								if i < len(tokens) && tokens[i].Type == TokenTypeRParen {
									instr.Mode = ModeIndirectX
									i = i + 1
								}
							}
						} else if i < len(tokens) && tokens[i].Type == TokenTypeRParen {
							// Check for ),Y - indirect indexed
							i = i + 1
							if i < len(tokens) && tokens[i].Type == TokenTypeComma {
								i = i + 1
								if i < len(tokens) && tokens[i].Type == TokenTypeIdentifier && MatchToken(tokens[i], "Y") {
									instr.Mode = ModeIndirectY
									i = i + 1
								}
							} else {
								// Just ($XX) - indirect mode for JMP
								instr.Mode = ModeIndirect
							}
						}
					}
				}
			} else if tokens[i].Type == TokenTypeHash {
			// Immediate mode: #$XX or #NN
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
					// Determine base mode by operand size
					if len(tokens[i].Representation) <= 2 {
						instr.Mode = ModeZeroPage
					} else {
						instr.Mode = ModeAbsolute
					}
					i = i + 1
					// Check for ,X or ,Y indexing
					if i < len(tokens) && tokens[i].Type == TokenTypeComma {
						i = i + 1
						if i < len(tokens) && tokens[i].Type == TokenTypeIdentifier {
							if MatchToken(tokens[i], "X") {
								if instr.Mode == ModeZeroPage {
									instr.Mode = ModeZeroPageX
								} else {
									instr.Mode = ModeAbsoluteX
								}
							} else if MatchToken(tokens[i], "Y") {
								if instr.Mode == ModeZeroPage {
									instr.Mode = ModeZeroPageY
								} else {
									instr.Mode = ModeAbsoluteY
								}
							}
							i = i + 1
						}
					}
				}
			} else if tokens[i].Type == TokenTypeIdentifier && MatchToken(tokens[i], "A") {
				// Accumulator mode: ASL A, LSR A, ROL A, ROR A
				instr.Mode = ModeAccumulator
				i = i + 1
			} else if tokens[i].Type == TokenTypeNumber {
				// Decimal number
				instr.Operand = ParseDecimal(tokens[i].Representation)
				if instr.Operand <= 255 {
					instr.Mode = ModeZeroPage
				} else {
					instr.Mode = ModeAbsolute
				}
				i = i + 1
			} else if tokens[i].Type == TokenTypeIdentifier {
				// Label reference for JMP/JSR/branch
				instr.TargetLabel = CopyBytes(tokens[i].Representation)
				instr.HasTargetLabel = true
				instr.Mode = ModeAbsolute // Default to absolute for JMP/JSR
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

// BytesMatch checks if two byte slices are equal
func BytesMatch(a []int8, b []int8) bool {
	if len(a) != len(b) {
		return false
	}
	i := 0
	for {
		if i >= len(a) {
			break
		}
		if a[i] != b[i] {
			return false
		}
		i = i + 1
	}
	return true
}

// getInstructionSize returns the size in bytes of an instruction
func getInstructionSize(instr Instruction) int {
	// Most instructions are 1-3 bytes
	opcodeBytes := instr.OpcodeBytes
	mode := instr.Mode

	// Implied mode instructions (1 byte)
	if mode == ModeImplied || mode == ModeAccumulator {
		return 1
	}

	// Branch instructions (2 bytes: opcode + relative offset)
	if IsOpcode(opcodeBytes, "BPL") || IsOpcode(opcodeBytes, "BMI") ||
		IsOpcode(opcodeBytes, "BVC") || IsOpcode(opcodeBytes, "BVS") ||
		IsOpcode(opcodeBytes, "BCC") || IsOpcode(opcodeBytes, "BCS") ||
		IsOpcode(opcodeBytes, "BNE") || IsOpcode(opcodeBytes, "BEQ") {
		return 2
	}

	// Immediate, ZeroPage, and Indirect modes (2 bytes)
	if mode == ModeImmediate || mode == ModeZeroPage ||
		mode == ModeZeroPageX || mode == ModeZeroPageY ||
		mode == ModeIndirectX || mode == ModeIndirectY {
		return 2
	}

	// Absolute modes (3 bytes)
	if mode == ModeAbsolute || mode == ModeAbsoluteX ||
		mode == ModeAbsoluteY || mode == ModeIndirect {
		return 3
	}

	// JMP and JSR with label (3 bytes)
	if instr.HasTargetLabel {
		return 3
	}

	// Default
	return 1
}

// LabelEntry stores a label name and its address
type LabelEntry struct {
	Name []int8
	Addr int
}

// findLabelAddr looks up a label address in the label table
func findLabelAddr(labels []LabelEntry, target []int8) (int, bool) {
	i := 0
	for {
		if i >= len(labels) {
			break
		}
		if BytesMatch(labels[i].Name, target) {
			return labels[i].Addr, true
		}
		i = i + 1
	}
	return 0, false
}

// resolveLabels performs first pass to calculate addresses and resolve labels
func resolveLabels(instructions []Instruction, baseAddr int) []Instruction {
	// Build label table using slices (goany compatible)
	labels := []LabelEntry{}
	currentAddr := baseAddr

	// First pass: calculate addresses
	idx := 0
	for {
		if idx >= len(instructions) {
			break
		}
		instr := instructions[idx]

		// If this instruction has a label definition, record its address
		if instr.HasLabel {
			entry := LabelEntry{
				Name: CopyBytes(instr.LabelBytes),
				Addr: currentAddr,
			}
			labels = append(labels, entry)
		}

		currentAddr = currentAddr + getInstructionSize(instr)
		idx = idx + 1
	}

	// Second pass: resolve label references
	result := []Instruction{}
	currentAddr = baseAddr

	idx = 0
	for {
		if idx >= len(instructions) {
			break
		}
		instr := instructions[idx]

		// Resolve label before appending to avoid C# struct indexer issue
		if instr.HasTargetLabel {
			// Look up address
			addr, found := findLabelAddr(labels, instr.TargetLabel)
			if found {
				// Check if this is a branch instruction (needs relative offset)
				if IsOpcode(instr.OpcodeBytes, "BPL") || IsOpcode(instr.OpcodeBytes, "BMI") ||
					IsOpcode(instr.OpcodeBytes, "BVC") || IsOpcode(instr.OpcodeBytes, "BVS") ||
					IsOpcode(instr.OpcodeBytes, "BCC") || IsOpcode(instr.OpcodeBytes, "BCS") ||
					IsOpcode(instr.OpcodeBytes, "BNE") || IsOpcode(instr.OpcodeBytes, "BEQ") {
					// Calculate relative offset
					// Branch offset is calculated from the address AFTER the branch instruction
					instrSize := getInstructionSize(instr)
					nextAddr := currentAddr + instrSize
					offset := addr - nextAddr
					// Convert to signed byte
					if offset < -128 || offset > 127 {
						offset = 0 // Out of range, use 0
					}
					instr.Operand = offset & 0xFF
					instr.Mode = ModeZeroPage // Branch uses 1-byte operand
				} else {
					// JMP or JSR - use absolute address
					instr.Operand = addr
				}
				instr.HasTargetLabel = false // Mark as resolved
			}
		}

		result = append(result, instr)
		currentAddr = currentAddr + getInstructionSize(instr)
		idx = idx + 1
	}

	return result
}

// CodeBase is the default base address for code
const CodeBaseAddr = 0xC000

// Assemble converts parsed instructions to machine code
func Assemble(instructions []Instruction) []uint8 {
	// Resolve labels before generating code
	instructions = resolveLabels(instructions, CodeBaseAddr)

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
			} else if instr.Mode == ModeAbsoluteX {
				code = append(code, uint8(cpu.OpLDAAbsX))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			} else if instr.Mode == ModeAbsoluteY {
				code = append(code, uint8(cpu.OpLDAAbsY))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			}
		} else if IsOpcode(opcodeBytes, "LDX") {
			if instr.Mode == ModeImmediate {
				code = append(code, uint8(cpu.OpLDXImm))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeZeroPage {
				code = append(code, uint8(cpu.OpLDXZp))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeZeroPageY {
				code = append(code, uint8(cpu.OpLDXZpY))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeAbsolute {
				code = append(code, uint8(cpu.OpLDXAbs))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			} else if instr.Mode == ModeAbsoluteY {
				code = append(code, uint8(cpu.OpLDXAbsY))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			}
		} else if IsOpcode(opcodeBytes, "LDY") {
			if instr.Mode == ModeImmediate {
				code = append(code, uint8(cpu.OpLDYImm))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeZeroPage {
				code = append(code, uint8(cpu.OpLDYZp))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeZeroPageX {
				code = append(code, uint8(cpu.OpLDYZpX))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeAbsolute {
				code = append(code, uint8(cpu.OpLDYAbs))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			} else if instr.Mode == ModeAbsoluteX {
				code = append(code, uint8(cpu.OpLDYAbsX))
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
			} else if instr.Mode == ModeAbsoluteX {
				code = append(code, uint8(cpu.OpSTAAbsX))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			} else if instr.Mode == ModeAbsoluteY {
				code = append(code, uint8(cpu.OpSTAAbsY))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			} else if instr.Mode == ModeIndirectY {
				code = append(code, uint8(cpu.OpSTAIndY))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeIndirectX {
				code = append(code, uint8(cpu.OpSTAIndX))
				code = append(code, uint8(instr.Operand))
			}
		} else if IsOpcode(opcodeBytes, "STX") {
			if instr.Mode == ModeZeroPage {
				code = append(code, uint8(cpu.OpSTXZp))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeZeroPageY {
				code = append(code, uint8(cpu.OpSTXZpY))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeAbsolute {
				code = append(code, uint8(cpu.OpSTXAbs))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			}
		} else if IsOpcode(opcodeBytes, "STY") {
			if instr.Mode == ModeZeroPage {
				code = append(code, uint8(cpu.OpSTYZp))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeZeroPageX {
				code = append(code, uint8(cpu.OpSTYZpX))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeAbsolute {
				code = append(code, uint8(cpu.OpSTYAbs))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			}

		// ADC - Add with Carry
		} else if IsOpcode(opcodeBytes, "ADC") {
			if instr.Mode == ModeImmediate {
				code = append(code, uint8(cpu.OpADCImm))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeZeroPage {
				code = append(code, uint8(cpu.OpADCZp))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeZeroPageX {
				code = append(code, uint8(cpu.OpADCZpX))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeAbsolute {
				code = append(code, uint8(cpu.OpADCAbs))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			} else if instr.Mode == ModeAbsoluteX {
				code = append(code, uint8(cpu.OpADCAbsX))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			} else if instr.Mode == ModeAbsoluteY {
				code = append(code, uint8(cpu.OpADCAbsY))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			}

		// SBC - Subtract with Carry
		} else if IsOpcode(opcodeBytes, "SBC") {
			if instr.Mode == ModeImmediate {
				code = append(code, uint8(cpu.OpSBCImm))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeZeroPage {
				code = append(code, uint8(cpu.OpSBCZp))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeZeroPageX {
				code = append(code, uint8(cpu.OpSBCZpX))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeAbsolute {
				code = append(code, uint8(cpu.OpSBCAbs))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			} else if instr.Mode == ModeAbsoluteX {
				code = append(code, uint8(cpu.OpSBCAbsX))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			} else if instr.Mode == ModeAbsoluteY {
				code = append(code, uint8(cpu.OpSBCAbsY))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			}

		// AND - Logical AND
		} else if IsOpcode(opcodeBytes, "AND") {
			if instr.Mode == ModeImmediate {
				code = append(code, uint8(cpu.OpANDImm))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeZeroPage {
				code = append(code, uint8(cpu.OpANDZp))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeZeroPageX {
				code = append(code, uint8(cpu.OpANDZpX))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeAbsolute {
				code = append(code, uint8(cpu.OpANDAbs))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			} else if instr.Mode == ModeAbsoluteX {
				code = append(code, uint8(cpu.OpANDAbsX))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			} else if instr.Mode == ModeAbsoluteY {
				code = append(code, uint8(cpu.OpANDAbsY))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			}

		// ORA - Logical OR
		} else if IsOpcode(opcodeBytes, "ORA") {
			if instr.Mode == ModeImmediate {
				code = append(code, uint8(cpu.OpORAImm))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeZeroPage {
				code = append(code, uint8(cpu.OpORAZp))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeZeroPageX {
				code = append(code, uint8(cpu.OpORAZpX))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeAbsolute {
				code = append(code, uint8(cpu.OpORAAbs))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			} else if instr.Mode == ModeAbsoluteX {
				code = append(code, uint8(cpu.OpORAAbsX))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			} else if instr.Mode == ModeAbsoluteY {
				code = append(code, uint8(cpu.OpORAAbsY))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			}

		// EOR - Exclusive OR
		} else if IsOpcode(opcodeBytes, "EOR") {
			if instr.Mode == ModeImmediate {
				code = append(code, uint8(cpu.OpEORImm))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeZeroPage {
				code = append(code, uint8(cpu.OpEORZp))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeZeroPageX {
				code = append(code, uint8(cpu.OpEORZpX))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeAbsolute {
				code = append(code, uint8(cpu.OpEORAbs))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			} else if instr.Mode == ModeAbsoluteX {
				code = append(code, uint8(cpu.OpEORAbsX))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			} else if instr.Mode == ModeAbsoluteY {
				code = append(code, uint8(cpu.OpEORAbsY))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			}

		// ASL - Arithmetic Shift Left
		} else if IsOpcode(opcodeBytes, "ASL") {
			if instr.Mode == ModeAccumulator || instr.Mode == ModeImplied {
				code = append(code, uint8(cpu.OpASLA))
			} else if instr.Mode == ModeZeroPage {
				code = append(code, uint8(cpu.OpASLZp))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeZeroPageX {
				code = append(code, uint8(cpu.OpASLZpX))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeAbsolute {
				code = append(code, uint8(cpu.OpASLAbs))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			} else if instr.Mode == ModeAbsoluteX {
				code = append(code, uint8(cpu.OpASLAbsX))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			}

		// LSR - Logical Shift Right
		} else if IsOpcode(opcodeBytes, "LSR") {
			if instr.Mode == ModeAccumulator || instr.Mode == ModeImplied {
				code = append(code, uint8(cpu.OpLSRA))
			} else if instr.Mode == ModeZeroPage {
				code = append(code, uint8(cpu.OpLSRZp))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeZeroPageX {
				code = append(code, uint8(cpu.OpLSRZpX))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeAbsolute {
				code = append(code, uint8(cpu.OpLSRAbs))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			} else if instr.Mode == ModeAbsoluteX {
				code = append(code, uint8(cpu.OpLSRAbsX))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			}

		// ROL - Rotate Left
		} else if IsOpcode(opcodeBytes, "ROL") {
			if instr.Mode == ModeAccumulator || instr.Mode == ModeImplied {
				code = append(code, uint8(cpu.OpROLA))
			} else if instr.Mode == ModeZeroPage {
				code = append(code, uint8(cpu.OpROLZp))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeZeroPageX {
				code = append(code, uint8(cpu.OpROLZpX))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeAbsolute {
				code = append(code, uint8(cpu.OpROLAbs))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			} else if instr.Mode == ModeAbsoluteX {
				code = append(code, uint8(cpu.OpROLAbsX))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			}

		// ROR - Rotate Right
		} else if IsOpcode(opcodeBytes, "ROR") {
			if instr.Mode == ModeAccumulator || instr.Mode == ModeImplied {
				code = append(code, uint8(cpu.OpRORA))
			} else if instr.Mode == ModeZeroPage {
				code = append(code, uint8(cpu.OpRORZp))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeZeroPageX {
				code = append(code, uint8(cpu.OpRORZpX))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeAbsolute {
				code = append(code, uint8(cpu.OpRORAbs))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			} else if instr.Mode == ModeAbsoluteX {
				code = append(code, uint8(cpu.OpRORAbsX))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			}

		// INC - Increment Memory
		} else if IsOpcode(opcodeBytes, "INC") {
			if instr.Mode == ModeZeroPage {
				code = append(code, uint8(cpu.OpINC))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeZeroPageX {
				code = append(code, uint8(cpu.OpINCZpX))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeAbsolute {
				code = append(code, uint8(cpu.OpINCAbs))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			}

		// DEC - Decrement Memory
		} else if IsOpcode(opcodeBytes, "DEC") {
			if instr.Mode == ModeZeroPage {
				code = append(code, uint8(cpu.OpDECZp))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeZeroPageX {
				code = append(code, uint8(cpu.OpDECZpX))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeAbsolute {
				code = append(code, uint8(cpu.OpDECAbs))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			}

		// Register increments/decrements
		} else if IsOpcode(opcodeBytes, "INX") {
			code = append(code, uint8(cpu.OpINX))
		} else if IsOpcode(opcodeBytes, "INY") {
			code = append(code, uint8(cpu.OpINY))
		} else if IsOpcode(opcodeBytes, "DEX") {
			code = append(code, uint8(cpu.OpDEX))
		} else if IsOpcode(opcodeBytes, "DEY") {
			code = append(code, uint8(cpu.OpDEY))

		// CMP - Compare Accumulator
		} else if IsOpcode(opcodeBytes, "CMP") {
			if instr.Mode == ModeImmediate {
				code = append(code, uint8(cpu.OpCMPImm))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeZeroPage {
				code = append(code, uint8(cpu.OpCMPZp))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeZeroPageX {
				code = append(code, uint8(cpu.OpCMPZpX))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeAbsolute {
				code = append(code, uint8(cpu.OpCMPAbs))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			} else if instr.Mode == ModeAbsoluteX {
				code = append(code, uint8(cpu.OpCMPAbsX))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			} else if instr.Mode == ModeAbsoluteY {
				code = append(code, uint8(cpu.OpCMPAbsY))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			}

		// CPX - Compare X Register
		} else if IsOpcode(opcodeBytes, "CPX") {
			if instr.Mode == ModeImmediate {
				code = append(code, uint8(cpu.OpCPXImm))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeZeroPage {
				code = append(code, uint8(cpu.OpCPXZp))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeAbsolute {
				code = append(code, uint8(cpu.OpCPXAbs))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			}

		// CPY - Compare Y Register
		} else if IsOpcode(opcodeBytes, "CPY") {
			if instr.Mode == ModeImmediate {
				code = append(code, uint8(cpu.OpCPYImm))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeZeroPage {
				code = append(code, uint8(cpu.OpCPYZp))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeAbsolute {
				code = append(code, uint8(cpu.OpCPYAbs))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			}

		// BIT - Bit Test
		} else if IsOpcode(opcodeBytes, "BIT") {
			if instr.Mode == ModeZeroPage {
				code = append(code, uint8(cpu.OpBITZp))
				code = append(code, uint8(instr.Operand))
			} else if instr.Mode == ModeAbsolute {
				code = append(code, uint8(cpu.OpBITAbs))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			}

		// Branch Instructions
		} else if IsOpcode(opcodeBytes, "BPL") {
			code = append(code, uint8(cpu.OpBPL))
			code = append(code, uint8(instr.Operand))
		} else if IsOpcode(opcodeBytes, "BMI") {
			code = append(code, uint8(cpu.OpBMI))
			code = append(code, uint8(instr.Operand))
		} else if IsOpcode(opcodeBytes, "BVC") {
			code = append(code, uint8(cpu.OpBVC))
			code = append(code, uint8(instr.Operand))
		} else if IsOpcode(opcodeBytes, "BVS") {
			code = append(code, uint8(cpu.OpBVS))
			code = append(code, uint8(instr.Operand))
		} else if IsOpcode(opcodeBytes, "BCC") {
			code = append(code, uint8(cpu.OpBCC))
			code = append(code, uint8(instr.Operand))
		} else if IsOpcode(opcodeBytes, "BCS") {
			code = append(code, uint8(cpu.OpBCS))
			code = append(code, uint8(instr.Operand))
		} else if IsOpcode(opcodeBytes, "BNE") {
			code = append(code, uint8(cpu.OpBNE))
			code = append(code, uint8(instr.Operand))
		} else if IsOpcode(opcodeBytes, "BEQ") {
			code = append(code, uint8(cpu.OpBEQ))
			code = append(code, uint8(instr.Operand))

		// Jump Instructions
		} else if IsOpcode(opcodeBytes, "JMP") {
			if instr.Mode == ModeIndirect {
				code = append(code, uint8(cpu.OpJMPInd))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			} else {
				code = append(code, uint8(cpu.OpJMP))
				code = append(code, uint8(instr.Operand&0xFF))
				code = append(code, uint8((instr.Operand>>8)&0xFF))
			}
		} else if IsOpcode(opcodeBytes, "JSR") {
			code = append(code, uint8(cpu.OpJSR))
			code = append(code, uint8(instr.Operand&0xFF))
			code = append(code, uint8((instr.Operand>>8)&0xFF))
		} else if IsOpcode(opcodeBytes, "RTS") {
			code = append(code, uint8(cpu.OpRTS))
		} else if IsOpcode(opcodeBytes, "RTI") {
			code = append(code, uint8(cpu.OpRTI))

		// Stack Instructions
		} else if IsOpcode(opcodeBytes, "PHA") {
			code = append(code, uint8(cpu.OpPHA))
		} else if IsOpcode(opcodeBytes, "PHP") {
			code = append(code, uint8(cpu.OpPHP))
		} else if IsOpcode(opcodeBytes, "PLA") {
			code = append(code, uint8(cpu.OpPLA))
		} else if IsOpcode(opcodeBytes, "PLP") {
			code = append(code, uint8(cpu.OpPLP))

		// Transfer Instructions
		} else if IsOpcode(opcodeBytes, "TAX") {
			code = append(code, uint8(cpu.OpTAX))
		} else if IsOpcode(opcodeBytes, "TXA") {
			code = append(code, uint8(cpu.OpTXA))
		} else if IsOpcode(opcodeBytes, "TAY") {
			code = append(code, uint8(cpu.OpTAY))
		} else if IsOpcode(opcodeBytes, "TYA") {
			code = append(code, uint8(cpu.OpTYA))
		} else if IsOpcode(opcodeBytes, "TSX") {
			code = append(code, uint8(cpu.OpTSX))
		} else if IsOpcode(opcodeBytes, "TXS") {
			code = append(code, uint8(cpu.OpTXS))

		// Flag Instructions
		} else if IsOpcode(opcodeBytes, "CLC") {
			code = append(code, uint8(cpu.OpCLC))
		} else if IsOpcode(opcodeBytes, "SEC") {
			code = append(code, uint8(cpu.OpSEC))
		} else if IsOpcode(opcodeBytes, "CLI") {
			code = append(code, uint8(cpu.OpCLI))
		} else if IsOpcode(opcodeBytes, "SEI") {
			code = append(code, uint8(cpu.OpSEI))
		} else if IsOpcode(opcodeBytes, "CLV") {
			code = append(code, uint8(cpu.OpCLV))
		} else if IsOpcode(opcodeBytes, "CLD") {
			code = append(code, uint8(cpu.OpCLD))
		} else if IsOpcode(opcodeBytes, "SED") {
			code = append(code, uint8(cpu.OpSED))

		// Other
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

		// Left parenthesis (indirect addressing)
		if b == '(' {
			tokens = append(tokens, Token{Type: TokenTypeLParen, Representation: []int8{b}})
			i = i + 1
			continue
		}

		// Right parenthesis (indirect addressing)
		if b == ')' {
			tokens = append(tokens, Token{Type: TokenTypeRParen, Representation: []int8{b}})
			i = i + 1
			continue
		}

		// Number - must start with 0-9 (hex letters A-F only valid after initial digit)
		if IsDigit(b) {
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

// AssembleLinesWithCount assembles and returns instruction count for debugging
func AssembleLinesWithCount(lines []string) ([]uint8, int) {
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
	return Assemble(instructions), len(instructions)
}

// GetLastInstrFirstByte returns the first byte of the last instruction's opcode for debugging
func GetLastInstrFirstByte(lines []string) int {
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
	if len(instructions) > 0 {
		lastInstr := instructions[len(instructions)-1]
		if len(lastInstr.OpcodeBytes) > 0 {
			return int(lastInstr.OpcodeBytes[0])
		}
	}
	return -1
}

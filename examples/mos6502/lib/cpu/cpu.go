package cpu

// MOS 6502 CPU Emulator
// A simple interpreter supporting basic instructions for drawing shapes

// Opcodes
const (
	// Load/Store - LDA
	OpLDAImm  = 0xA9 // LDA #value
	OpLDAZp   = 0xA5 // LDA $zp
	OpLDAZpX  = 0xB5 // LDA $zp,X
	OpLDAAbs  = 0xAD // LDA $addr
	OpLDAAbsX = 0xBD // LDA $addr,X
	OpLDAAbsY = 0xB9 // LDA $addr,Y
	OpLDAIndX = 0xA1 // LDA ($zp,X)
	OpLDAIndY = 0xB1 // LDA ($zp),Y

	// Load/Store - LDX
	OpLDXImm  = 0xA2 // LDX #value
	OpLDXZp   = 0xA6 // LDX $zp
	OpLDXZpY  = 0xB6 // LDX $zp,Y
	OpLDXAbs  = 0xAE // LDX $addr
	OpLDXAbsY = 0xBE // LDX $addr,Y

	// Load/Store - LDY
	OpLDYImm  = 0xA0 // LDY #value
	OpLDYZp   = 0xA4 // LDY $zp
	OpLDYZpX  = 0xB4 // LDY $zp,X
	OpLDYAbs  = 0xAC // LDY $addr
	OpLDYAbsX = 0xBC // LDY $addr,X

	// Load/Store - STA
	OpSTAZp   = 0x85 // STA $zp
	OpSTAZpX  = 0x95 // STA $zp,X
	OpSTAAbs  = 0x8D // STA $addr
	OpSTAAbsX = 0x9D // STA $addr,X
	OpSTAAbsY = 0x99 // STA $addr,Y
	OpSTAIndX = 0x81 // STA ($zp,X)
	OpSTAIndY = 0x91 // STA ($zp),Y

	// Load/Store - STX
	OpSTXZp  = 0x86 // STX $zp
	OpSTXZpY = 0x96 // STX $zp,Y
	OpSTXAbs = 0x8E // STX $addr

	// Load/Store - STY
	OpSTYZp  = 0x84 // STY $zp
	OpSTYZpX = 0x94 // STY $zp,X
	OpSTYAbs = 0x8C // STY $addr

	// Arithmetic - ADC
	OpADCImm  = 0x69 // ADC #value
	OpADCZp   = 0x65 // ADC $zp
	OpADCZpX  = 0x75 // ADC $zp,X
	OpADCAbs  = 0x6D // ADC $addr
	OpADCAbsX = 0x7D // ADC $addr,X
	OpADCAbsY = 0x79 // ADC $addr,Y
	OpADCIndX = 0x61 // ADC ($zp,X)
	OpADCIndY = 0x71 // ADC ($zp),Y

	// Arithmetic - SBC
	OpSBCImm  = 0xE9 // SBC #value
	OpSBCZp   = 0xE5 // SBC $zp
	OpSBCZpX  = 0xF5 // SBC $zp,X
	OpSBCAbs  = 0xED // SBC $addr
	OpSBCAbsX = 0xFD // SBC $addr,X
	OpSBCAbsY = 0xF9 // SBC $addr,Y
	OpSBCIndX = 0xE1 // SBC ($zp,X)
	OpSBCIndY = 0xF1 // SBC ($zp),Y

	// Logical - AND
	OpANDImm  = 0x29 // AND #value
	OpANDZp   = 0x25 // AND $zp
	OpANDZpX  = 0x35 // AND $zp,X
	OpANDAbs  = 0x2D // AND $addr
	OpANDAbsX = 0x3D // AND $addr,X
	OpANDAbsY = 0x39 // AND $addr,Y
	OpANDIndX = 0x21 // AND ($zp,X)
	OpANDIndY = 0x31 // AND ($zp),Y

	// Logical - ORA
	OpORAImm  = 0x09 // ORA #value
	OpORAZp   = 0x05 // ORA $zp
	OpORAZpX  = 0x15 // ORA $zp,X
	OpORAAbs  = 0x0D // ORA $addr
	OpORAAbsX = 0x1D // ORA $addr,X
	OpORAAbsY = 0x19 // ORA $addr,Y
	OpORAIndX = 0x01 // ORA ($zp,X)
	OpORAIndY = 0x11 // ORA ($zp),Y

	// Logical - EOR
	OpEORImm  = 0x49 // EOR #value
	OpEORZp   = 0x45 // EOR $zp
	OpEORZpX  = 0x55 // EOR $zp,X
	OpEORAbs  = 0x4D // EOR $addr
	OpEORAbsX = 0x5D // EOR $addr,X
	OpEORAbsY = 0x59 // EOR $addr,Y
	OpEORIndX = 0x41 // EOR ($zp,X)
	OpEORIndY = 0x51 // EOR ($zp),Y

	// Shift - ASL
	OpASLA    = 0x0A // ASL A
	OpASLZp   = 0x06 // ASL $zp
	OpASLZpX  = 0x16 // ASL $zp,X
	OpASLAbs  = 0x0E // ASL $addr
	OpASLAbsX = 0x1E // ASL $addr,X

	// Shift - LSR
	OpLSRA    = 0x4A // LSR A
	OpLSRZp   = 0x46 // LSR $zp
	OpLSRZpX  = 0x56 // LSR $zp,X
	OpLSRAbs  = 0x4E // LSR $addr
	OpLSRAbsX = 0x5E // LSR $addr,X

	// Rotate - ROL
	OpROLA    = 0x2A // ROL A
	OpROLZp   = 0x26 // ROL $zp
	OpROLZpX  = 0x36 // ROL $zp,X
	OpROLAbs  = 0x2E // ROL $addr
	OpROLAbsX = 0x3E // ROL $addr,X

	// Rotate - ROR
	OpRORA    = 0x6A // ROR A
	OpRORZp   = 0x66 // ROR $zp
	OpRORZpX  = 0x76 // ROR $zp,X
	OpRORAbs  = 0x6E // ROR $addr
	OpRORAbsX = 0x7E // ROR $addr,X

	// Increment/Decrement
	OpINX    = 0xE8 // INX
	OpINY    = 0xC8 // INY
	OpDEX    = 0xCA // DEX
	OpDEY    = 0x88 // DEY
	OpINC    = 0xE6 // INC $zp
	OpINCZpX = 0xF6 // INC $zp,X
	OpINCAbs = 0xEE // INC $addr
	OpDECZp  = 0xC6 // DEC $zp
	OpDECZpX = 0xD6 // DEC $zp,X
	OpDECAbs = 0xCE // DEC $addr

	// Compare - CMP
	OpCMPImm  = 0xC9 // CMP #value
	OpCMPZp   = 0xC5 // CMP $zp
	OpCMPZpX  = 0xD5 // CMP $zp,X
	OpCMPAbs  = 0xCD // CMP $addr
	OpCMPAbsX = 0xDD // CMP $addr,X
	OpCMPAbsY = 0xD9 // CMP $addr,Y
	OpCMPIndX = 0xC1 // CMP ($zp,X)
	OpCMPIndY = 0xD1 // CMP ($zp),Y

	// Compare - CPX
	OpCPXImm = 0xE0 // CPX #value
	OpCPXZp  = 0xE4 // CPX $zp
	OpCPXAbs = 0xEC // CPX $addr

	// Compare - CPY
	OpCPYImm = 0xC0 // CPY #value
	OpCPYZp  = 0xC4 // CPY $zp
	OpCPYAbs = 0xCC // CPY $addr

	// Branch
	OpBPL = 0x10 // BPL offset (branch if positive)
	OpBMI = 0x30 // BMI offset (branch if minus)
	OpBVC = 0x50 // BVC offset (branch if overflow clear)
	OpBVS = 0x70 // BVS offset (branch if overflow set)
	OpBCC = 0x90 // BCC offset (branch if carry clear)
	OpBCS = 0xB0 // BCS offset (branch if carry set)
	OpBNE = 0xD0 // BNE offset (branch if not equal)
	OpBEQ = 0xF0 // BEQ offset (branch if equal)

	// Jump
	OpJMP    = 0x4C // JMP $addr
	OpJMPInd = 0x6C // JMP ($addr)
	OpJSR    = 0x20 // JSR $addr
	OpRTS    = 0x60 // RTS
	OpRTI    = 0x40 // RTI

	// Stack
	OpPHA = 0x48 // PHA (push A)
	OpPHP = 0x08 // PHP (push status)
	OpPLA = 0x68 // PLA (pull A)
	OpPLP = 0x28 // PLP (pull status)

	// Transfer
	OpTAX = 0xAA // TAX (A -> X)
	OpTXA = 0x8A // TXA (X -> A)
	OpTAY = 0xA8 // TAY (A -> Y)
	OpTYA = 0x98 // TYA (Y -> A)
	OpTSX = 0xBA // TSX (SP -> X)
	OpTXS = 0x9A // TXS (X -> SP)

	// Flags
	OpCLC = 0x18 // CLC (clear carry)
	OpSEC = 0x38 // SEC (set carry)
	OpCLI = 0x58 // CLI (clear interrupt)
	OpSEI = 0x78 // SEI (set interrupt)
	OpCLV = 0xB8 // CLV (clear overflow)
	OpCLD = 0xD8 // CLD (clear decimal)
	OpSED = 0xF8 // SED (set decimal)

	// Bit Test
	OpBITZp  = 0x24 // BIT $zp
	OpBITAbs = 0x2C // BIT $addr

	// Other
	OpNOP = 0xEA // NOP
	OpBRK = 0x00 // BRK (halt)
)

// Status flags
const (
	FlagC = 0x01 // Carry
	FlagZ = 0x02 // Zero
	FlagI = 0x04 // Interrupt disable
	FlagD = 0x08 // Decimal mode
	FlagB = 0x10 // Break
	FlagV = 0x40 // Overflow
	FlagN = 0x80 // Negative
)

// Screen memory constants
const (
	ScreenBase   = 0x0200 // Screen starts at $0200
	ScreenWidth  = 32     // 32 pixels wide
	ScreenHeight = 32     // 32 pixels tall
	ScreenSize   = 1024   // Total screen memory
)

// CPU represents the 6502 CPU state
type CPU struct {
	A      uint8    // Accumulator
	X      uint8    // X register
	Y      uint8    // Y register
	SP     uint8    // Stack pointer
	PC     int      // Program counter (using int for easier math)
	Status uint8    // Status register
	Memory []uint8  // 64KB memory
	Halted bool     // CPU halted flag
	Cycles int      // Cycle counter
}

// NewCPU creates a new CPU with initialized memory
func NewCPU() CPU {
	mem := []uint8{}
	i := 0
	for {
		if i >= 65536 {
			break
		}
		mem = append(mem, uint8(0))
		i = i + 1
	}
	return CPU{
		A:      0,
		X:      0,
		Y:      0,
		SP:     0xFF,
		PC:     0x0600, // Programs start at $0600
		Status: 0x20,   // Unused bit always set
		Memory: mem,
		Halted: false,
		Cycles: 0,
	}
}

// LoadProgram loads a program into memory at the specified address
func LoadProgram(c CPU, program []uint8, addr int) CPU {
	i := 0
	for {
		if i >= len(program) {
			break
		}
		c.Memory[addr+i] = program[i]
		i = i + 1
	}
	return c
}

// SetPC sets the program counter
func SetPC(c CPU, addr int) CPU {
	c.PC = addr
	return c
}

// ReadByte reads a byte from memory
func ReadByte(c CPU, addr int) uint8 {
	return c.Memory[addr]
}

// WriteByte writes a byte to memory and returns updated CPU
func WriteByte(c CPU, addr int, value uint8) CPU {
	c.Memory[addr] = value
	return c
}

// FetchByte fetches the next byte and increments PC
func FetchByte(c CPU) (CPU, uint8) {
	value := c.Memory[c.PC]
	c.PC = c.PC + 1
	return c, value
}

// FetchWord fetches the next 16-bit word (little endian) and increments PC
func FetchWord(c CPU) (CPU, int) {
	low := int(c.Memory[c.PC])
	high := int(c.Memory[c.PC+1])
	c.PC = c.PC + 2
	return c, low + (high * 256)
}

// SetZN sets Zero and Negative flags based on value
func SetZN(c CPU, value uint8) CPU {
	if value == 0 {
		c.Status = c.Status | FlagZ
	} else {
		c.Status = c.Status & (0xFF - FlagZ)
	}
	if (value & 0x80) != 0 {
		c.Status = c.Status | FlagN
	} else {
		c.Status = c.Status & (0xFF - FlagN)
	}
	return c
}

// SetCarry sets the carry flag
func SetCarry(c CPU, set bool) CPU {
	if set {
		c.Status = c.Status | FlagC
	} else {
		c.Status = c.Status & (0xFF - FlagC)
	}
	return c
}

// GetCarry returns the carry flag
func GetCarry(c CPU) bool {
	return (c.Status & FlagC) != 0
}

// GetZero returns the zero flag
func GetZero(c CPU) bool {
	return (c.Status & FlagZ) != 0
}

// GetNegative returns the negative flag
func GetNegative(c CPU) bool {
	return (c.Status & FlagN) != 0
}

// GetOverflow returns the overflow flag
func GetOverflow(c CPU) bool {
	return (c.Status & FlagV) != 0
}

// SetOverflow sets the overflow flag
func SetOverflow(c CPU, set bool) CPU {
	if set {
		c.Status = c.Status | FlagV
	} else {
		c.Status = c.Status & (0xFF - FlagV)
	}
	return c
}

// ReadIndirectX reads using (zp,X) addressing mode
func ReadIndirectX(c CPU, zp uint8) uint8 {
	addr := int(zp + c.X)
	low := int(c.Memory[addr&0xFF])
	high := int(c.Memory[(addr+1)&0xFF])
	return c.Memory[low+(high*256)]
}

// ReadIndirectY reads using (zp),Y addressing mode
func ReadIndirectY(c CPU, zp uint8) uint8 {
	low := int(c.Memory[int(zp)])
	high := int(c.Memory[(int(zp)+1)&0xFF])
	addr := low + (high * 256) + int(c.Y)
	return c.Memory[addr]
}

// GetIndirectXAddr gets the effective address for (zp,X) mode
func GetIndirectXAddr(c CPU, zp uint8) int {
	addr := int(zp + c.X)
	low := int(c.Memory[addr&0xFF])
	high := int(c.Memory[(addr+1)&0xFF])
	return low + (high * 256)
}

// GetIndirectYAddr gets the effective address for (zp),Y mode
func GetIndirectYAddr(c CPU, zp uint8) int {
	low := int(c.Memory[int(zp)])
	high := int(c.Memory[(int(zp)+1)&0xFF])
	return low + (high * 256) + int(c.Y)
}

// PushByte pushes a byte onto the stack
func PushByte(c CPU, value uint8) CPU {
	c.Memory[0x100+int(c.SP)] = value
	c.SP = c.SP - 1
	return c
}

// PullByte pulls a byte from the stack
func PullByte(c CPU) (CPU, uint8) {
	c.SP = c.SP + 1
	return c, c.Memory[0x100+int(c.SP)]
}

// Branch performs a relative branch
func Branch(c CPU, offset uint8) CPU {
	if offset < 128 {
		c.PC = c.PC + int(offset)
	} else {
		c.PC = c.PC - (256 - int(offset))
	}
	return c
}

// Step executes one instruction and returns updated CPU
func Step(c CPU) CPU {
	if c.Halted {
		return c
	}

	var opcode uint8
	c, opcode = FetchByte(c)
	c.Cycles = c.Cycles + 1

	// LDA - Load Accumulator
	if opcode == OpLDAImm {
		var value uint8
		c, value = FetchByte(c)
		c.A = value
		c = SetZN(c, c.A)
	} else if opcode == OpLDAZp {
		var addr uint8
		c, addr = FetchByte(c)
		c.A = c.Memory[int(addr)]
		c = SetZN(c, c.A)
	} else if opcode == OpLDAZpX {
		var addr uint8
		c, addr = FetchByte(c)
		c.A = c.Memory[int(addr+c.X)&0xFF]
		c = SetZN(c, c.A)
	} else if opcode == OpLDAAbs {
		var addr int
		c, addr = FetchWord(c)
		c.A = c.Memory[addr]
		c = SetZN(c, c.A)
	} else if opcode == OpLDAAbsX {
		var addr int
		c, addr = FetchWord(c)
		c.A = c.Memory[addr+int(c.X)]
		c = SetZN(c, c.A)
	} else if opcode == OpLDAAbsY {
		var addr int
		c, addr = FetchWord(c)
		c.A = c.Memory[addr+int(c.Y)]
		c = SetZN(c, c.A)
	} else if opcode == OpLDAIndX {
		var zp uint8
		c, zp = FetchByte(c)
		c.A = ReadIndirectX(c, zp)
		c = SetZN(c, c.A)
	} else if opcode == OpLDAIndY {
		var zp uint8
		c, zp = FetchByte(c)
		c.A = ReadIndirectY(c, zp)
		c = SetZN(c, c.A)

	// LDX - Load X Register
	} else if opcode == OpLDXImm {
		var value uint8
		c, value = FetchByte(c)
		c.X = value
		c = SetZN(c, c.X)
	} else if opcode == OpLDXZp {
		var addr uint8
		c, addr = FetchByte(c)
		c.X = c.Memory[int(addr)]
		c = SetZN(c, c.X)
	} else if opcode == OpLDXZpY {
		var addr uint8
		c, addr = FetchByte(c)
		c.X = c.Memory[int(addr+c.Y)&0xFF]
		c = SetZN(c, c.X)
	} else if opcode == OpLDXAbs {
		var addr int
		c, addr = FetchWord(c)
		c.X = c.Memory[addr]
		c = SetZN(c, c.X)
	} else if opcode == OpLDXAbsY {
		var addr int
		c, addr = FetchWord(c)
		c.X = c.Memory[addr+int(c.Y)]
		c = SetZN(c, c.X)

	// LDY - Load Y Register
	} else if opcode == OpLDYImm {
		var value uint8
		c, value = FetchByte(c)
		c.Y = value
		c = SetZN(c, c.Y)
	} else if opcode == OpLDYZp {
		var addr uint8
		c, addr = FetchByte(c)
		c.Y = c.Memory[int(addr)]
		c = SetZN(c, c.Y)
	} else if opcode == OpLDYZpX {
		var addr uint8
		c, addr = FetchByte(c)
		c.Y = c.Memory[int(addr+c.X)&0xFF]
		c = SetZN(c, c.Y)
	} else if opcode == OpLDYAbs {
		var addr int
		c, addr = FetchWord(c)
		c.Y = c.Memory[addr]
		c = SetZN(c, c.Y)
	} else if opcode == OpLDYAbsX {
		var addr int
		c, addr = FetchWord(c)
		c.Y = c.Memory[addr+int(c.X)]
		c = SetZN(c, c.Y)

	// STA - Store Accumulator
	} else if opcode == OpSTAZp {
		var addr uint8
		c, addr = FetchByte(c)
		c.Memory[int(addr)] = c.A
	} else if opcode == OpSTAZpX {
		var addr uint8
		c, addr = FetchByte(c)
		c.Memory[int(addr+c.X)&0xFF] = c.A
	} else if opcode == OpSTAAbs {
		var addr int
		c, addr = FetchWord(c)
		c.Memory[addr] = c.A
	} else if opcode == OpSTAAbsX {
		var addr int
		c, addr = FetchWord(c)
		c.Memory[addr+int(c.X)] = c.A
	} else if opcode == OpSTAAbsY {
		var addr int
		c, addr = FetchWord(c)
		c.Memory[addr+int(c.Y)] = c.A
	} else if opcode == OpSTAIndX {
		var zp uint8
		c, zp = FetchByte(c)
		addr := GetIndirectXAddr(c, zp)
		c.Memory[addr] = c.A
	} else if opcode == OpSTAIndY {
		var zp uint8
		c, zp = FetchByte(c)
		addr := GetIndirectYAddr(c, zp)
		c.Memory[addr] = c.A

	// STX - Store X Register
	} else if opcode == OpSTXZp {
		var addr uint8
		c, addr = FetchByte(c)
		c.Memory[int(addr)] = c.X
	} else if opcode == OpSTXZpY {
		var addr uint8
		c, addr = FetchByte(c)
		c.Memory[int(addr+c.Y)&0xFF] = c.X
	} else if opcode == OpSTXAbs {
		var addr int
		c, addr = FetchWord(c)
		c.Memory[addr] = c.X

	// STY - Store Y Register
	} else if opcode == OpSTYZp {
		var addr uint8
		c, addr = FetchByte(c)
		c.Memory[int(addr)] = c.Y
	} else if opcode == OpSTYZpX {
		var addr uint8
		c, addr = FetchByte(c)
		c.Memory[int(addr+c.X)&0xFF] = c.Y
	} else if opcode == OpSTYAbs {
		var addr int
		c, addr = FetchWord(c)
		c.Memory[addr] = c.Y

	// ADC - Add with Carry
	} else if opcode == OpADCImm {
		var value uint8
		c, value = FetchByte(c)
		c = doADC(c, value)
	} else if opcode == OpADCZp {
		var addr uint8
		c, addr = FetchByte(c)
		c = doADC(c, c.Memory[int(addr)])
	} else if opcode == OpADCZpX {
		var addr uint8
		c, addr = FetchByte(c)
		c = doADC(c, c.Memory[int(addr+c.X)&0xFF])
	} else if opcode == OpADCAbs {
		var addr int
		c, addr = FetchWord(c)
		c = doADC(c, c.Memory[addr])
	} else if opcode == OpADCAbsX {
		var addr int
		c, addr = FetchWord(c)
		c = doADC(c, c.Memory[addr+int(c.X)])
	} else if opcode == OpADCAbsY {
		var addr int
		c, addr = FetchWord(c)
		c = doADC(c, c.Memory[addr+int(c.Y)])
	} else if opcode == OpADCIndX {
		var zp uint8
		c, zp = FetchByte(c)
		c = doADC(c, ReadIndirectX(c, zp))
	} else if opcode == OpADCIndY {
		var zp uint8
		c, zp = FetchByte(c)
		c = doADC(c, ReadIndirectY(c, zp))

	// SBC - Subtract with Carry
	} else if opcode == OpSBCImm {
		var value uint8
		c, value = FetchByte(c)
		c = doSBC(c, value)
	} else if opcode == OpSBCZp {
		var addr uint8
		c, addr = FetchByte(c)
		c = doSBC(c, c.Memory[int(addr)])
	} else if opcode == OpSBCZpX {
		var addr uint8
		c, addr = FetchByte(c)
		c = doSBC(c, c.Memory[int(addr+c.X)&0xFF])
	} else if opcode == OpSBCAbs {
		var addr int
		c, addr = FetchWord(c)
		c = doSBC(c, c.Memory[addr])
	} else if opcode == OpSBCAbsX {
		var addr int
		c, addr = FetchWord(c)
		c = doSBC(c, c.Memory[addr+int(c.X)])
	} else if opcode == OpSBCAbsY {
		var addr int
		c, addr = FetchWord(c)
		c = doSBC(c, c.Memory[addr+int(c.Y)])
	} else if opcode == OpSBCIndX {
		var zp uint8
		c, zp = FetchByte(c)
		c = doSBC(c, ReadIndirectX(c, zp))
	} else if opcode == OpSBCIndY {
		var zp uint8
		c, zp = FetchByte(c)
		c = doSBC(c, ReadIndirectY(c, zp))

	// AND - Logical AND
	} else if opcode == OpANDImm {
		var value uint8
		c, value = FetchByte(c)
		c.A = c.A & value
		c = SetZN(c, c.A)
	} else if opcode == OpANDZp {
		var addr uint8
		c, addr = FetchByte(c)
		c.A = c.A & c.Memory[int(addr)]
		c = SetZN(c, c.A)
	} else if opcode == OpANDZpX {
		var addr uint8
		c, addr = FetchByte(c)
		c.A = c.A & c.Memory[int(addr+c.X)&0xFF]
		c = SetZN(c, c.A)
	} else if opcode == OpANDAbs {
		var addr int
		c, addr = FetchWord(c)
		c.A = c.A & c.Memory[addr]
		c = SetZN(c, c.A)
	} else if opcode == OpANDAbsX {
		var addr int
		c, addr = FetchWord(c)
		c.A = c.A & c.Memory[addr+int(c.X)]
		c = SetZN(c, c.A)
	} else if opcode == OpANDAbsY {
		var addr int
		c, addr = FetchWord(c)
		c.A = c.A & c.Memory[addr+int(c.Y)]
		c = SetZN(c, c.A)
	} else if opcode == OpANDIndX {
		var zp uint8
		c, zp = FetchByte(c)
		c.A = c.A & ReadIndirectX(c, zp)
		c = SetZN(c, c.A)
	} else if opcode == OpANDIndY {
		var zp uint8
		c, zp = FetchByte(c)
		c.A = c.A & ReadIndirectY(c, zp)
		c = SetZN(c, c.A)

	// ORA - Logical OR
	} else if opcode == OpORAImm {
		var value uint8
		c, value = FetchByte(c)
		c.A = c.A | value
		c = SetZN(c, c.A)
	} else if opcode == OpORAZp {
		var addr uint8
		c, addr = FetchByte(c)
		c.A = c.A | c.Memory[int(addr)]
		c = SetZN(c, c.A)
	} else if opcode == OpORAZpX {
		var addr uint8
		c, addr = FetchByte(c)
		c.A = c.A | c.Memory[int(addr+c.X)&0xFF]
		c = SetZN(c, c.A)
	} else if opcode == OpORAAbs {
		var addr int
		c, addr = FetchWord(c)
		c.A = c.A | c.Memory[addr]
		c = SetZN(c, c.A)
	} else if opcode == OpORAAbsX {
		var addr int
		c, addr = FetchWord(c)
		c.A = c.A | c.Memory[addr+int(c.X)]
		c = SetZN(c, c.A)
	} else if opcode == OpORAAbsY {
		var addr int
		c, addr = FetchWord(c)
		c.A = c.A | c.Memory[addr+int(c.Y)]
		c = SetZN(c, c.A)
	} else if opcode == OpORAIndX {
		var zp uint8
		c, zp = FetchByte(c)
		c.A = c.A | ReadIndirectX(c, zp)
		c = SetZN(c, c.A)
	} else if opcode == OpORAIndY {
		var zp uint8
		c, zp = FetchByte(c)
		c.A = c.A | ReadIndirectY(c, zp)
		c = SetZN(c, c.A)

	// EOR - Exclusive OR
	} else if opcode == OpEORImm {
		var value uint8
		c, value = FetchByte(c)
		c.A = c.A ^ value
		c = SetZN(c, c.A)
	} else if opcode == OpEORZp {
		var addr uint8
		c, addr = FetchByte(c)
		c.A = c.A ^ c.Memory[int(addr)]
		c = SetZN(c, c.A)
	} else if opcode == OpEORZpX {
		var addr uint8
		c, addr = FetchByte(c)
		c.A = c.A ^ c.Memory[int(addr+c.X)&0xFF]
		c = SetZN(c, c.A)
	} else if opcode == OpEORAbs {
		var addr int
		c, addr = FetchWord(c)
		c.A = c.A ^ c.Memory[addr]
		c = SetZN(c, c.A)
	} else if opcode == OpEORAbsX {
		var addr int
		c, addr = FetchWord(c)
		c.A = c.A ^ c.Memory[addr+int(c.X)]
		c = SetZN(c, c.A)
	} else if opcode == OpEORAbsY {
		var addr int
		c, addr = FetchWord(c)
		c.A = c.A ^ c.Memory[addr+int(c.Y)]
		c = SetZN(c, c.A)
	} else if opcode == OpEORIndX {
		var zp uint8
		c, zp = FetchByte(c)
		c.A = c.A ^ ReadIndirectX(c, zp)
		c = SetZN(c, c.A)
	} else if opcode == OpEORIndY {
		var zp uint8
		c, zp = FetchByte(c)
		c.A = c.A ^ ReadIndirectY(c, zp)
		c = SetZN(c, c.A)

	// ASL - Arithmetic Shift Left
	} else if opcode == OpASLA {
		c = SetCarry(c, (c.A&0x80) != 0)
		c.A = c.A << 1
		c = SetZN(c, c.A)
	} else if opcode == OpASLZp {
		var addr uint8
		c, addr = FetchByte(c)
		val := c.Memory[int(addr)]
		c = SetCarry(c, (val&0x80) != 0)
		val = val << 1
		c.Memory[int(addr)] = val
		c = SetZN(c, val)
	} else if opcode == OpASLZpX {
		var addr uint8
		c, addr = FetchByte(c)
		effAddr := int(addr+c.X) & 0xFF
		val := c.Memory[effAddr]
		c = SetCarry(c, (val&0x80) != 0)
		val = val << 1
		c.Memory[effAddr] = val
		c = SetZN(c, val)
	} else if opcode == OpASLAbs {
		var addr int
		c, addr = FetchWord(c)
		val := c.Memory[addr]
		c = SetCarry(c, (val&0x80) != 0)
		val = val << 1
		c.Memory[addr] = val
		c = SetZN(c, val)
	} else if opcode == OpASLAbsX {
		var addr int
		c, addr = FetchWord(c)
		effAddr := addr + int(c.X)
		val := c.Memory[effAddr]
		c = SetCarry(c, (val&0x80) != 0)
		val = val << 1
		c.Memory[effAddr] = val
		c = SetZN(c, val)

	// LSR - Logical Shift Right
	} else if opcode == OpLSRA {
		c = SetCarry(c, (c.A&0x01) != 0)
		c.A = c.A >> 1
		c = SetZN(c, c.A)
	} else if opcode == OpLSRZp {
		var addr uint8
		c, addr = FetchByte(c)
		val := c.Memory[int(addr)]
		c = SetCarry(c, (val&0x01) != 0)
		val = val >> 1
		c.Memory[int(addr)] = val
		c = SetZN(c, val)
	} else if opcode == OpLSRZpX {
		var addr uint8
		c, addr = FetchByte(c)
		effAddr := int(addr+c.X) & 0xFF
		val := c.Memory[effAddr]
		c = SetCarry(c, (val&0x01) != 0)
		val = val >> 1
		c.Memory[effAddr] = val
		c = SetZN(c, val)
	} else if opcode == OpLSRAbs {
		var addr int
		c, addr = FetchWord(c)
		val := c.Memory[addr]
		c = SetCarry(c, (val&0x01) != 0)
		val = val >> 1
		c.Memory[addr] = val
		c = SetZN(c, val)
	} else if opcode == OpLSRAbsX {
		var addr int
		c, addr = FetchWord(c)
		effAddr := addr + int(c.X)
		val := c.Memory[effAddr]
		c = SetCarry(c, (val&0x01) != 0)
		val = val >> 1
		c.Memory[effAddr] = val
		c = SetZN(c, val)

	// ROL - Rotate Left
	} else if opcode == OpROLA {
		carry := 0
		if GetCarry(c) {
			carry = 1
		}
		c = SetCarry(c, (c.A&0x80) != 0)
		c.A = (c.A << 1) | uint8(carry)
		c = SetZN(c, c.A)
	} else if opcode == OpROLZp {
		var addr uint8
		c, addr = FetchByte(c)
		carry := 0
		if GetCarry(c) {
			carry = 1
		}
		val := c.Memory[int(addr)]
		c = SetCarry(c, (val&0x80) != 0)
		val = (val << 1) | uint8(carry)
		c.Memory[int(addr)] = val
		c = SetZN(c, val)
	} else if opcode == OpROLZpX {
		var addr uint8
		c, addr = FetchByte(c)
		carry := 0
		if GetCarry(c) {
			carry = 1
		}
		effAddr := int(addr+c.X) & 0xFF
		val := c.Memory[effAddr]
		c = SetCarry(c, (val&0x80) != 0)
		val = (val << 1) | uint8(carry)
		c.Memory[effAddr] = val
		c = SetZN(c, val)
	} else if opcode == OpROLAbs {
		var addr int
		c, addr = FetchWord(c)
		carry := 0
		if GetCarry(c) {
			carry = 1
		}
		val := c.Memory[addr]
		c = SetCarry(c, (val&0x80) != 0)
		val = (val << 1) | uint8(carry)
		c.Memory[addr] = val
		c = SetZN(c, val)
	} else if opcode == OpROLAbsX {
		var addr int
		c, addr = FetchWord(c)
		carry := 0
		if GetCarry(c) {
			carry = 1
		}
		effAddr := addr + int(c.X)
		val := c.Memory[effAddr]
		c = SetCarry(c, (val&0x80) != 0)
		val = (val << 1) | uint8(carry)
		c.Memory[effAddr] = val
		c = SetZN(c, val)

	// ROR - Rotate Right
	} else if opcode == OpRORA {
		carry := 0
		if GetCarry(c) {
			carry = 0x80
		}
		c = SetCarry(c, (c.A&0x01) != 0)
		c.A = (c.A >> 1) | uint8(carry)
		c = SetZN(c, c.A)
	} else if opcode == OpRORZp {
		var addr uint8
		c, addr = FetchByte(c)
		carry := 0
		if GetCarry(c) {
			carry = 0x80
		}
		val := c.Memory[int(addr)]
		c = SetCarry(c, (val&0x01) != 0)
		val = (val >> 1) | uint8(carry)
		c.Memory[int(addr)] = val
		c = SetZN(c, val)
	} else if opcode == OpRORZpX {
		var addr uint8
		c, addr = FetchByte(c)
		carry := 0
		if GetCarry(c) {
			carry = 0x80
		}
		effAddr := int(addr+c.X) & 0xFF
		val := c.Memory[effAddr]
		c = SetCarry(c, (val&0x01) != 0)
		val = (val >> 1) | uint8(carry)
		c.Memory[effAddr] = val
		c = SetZN(c, val)
	} else if opcode == OpRORAbs {
		var addr int
		c, addr = FetchWord(c)
		carry := 0
		if GetCarry(c) {
			carry = 0x80
		}
		val := c.Memory[addr]
		c = SetCarry(c, (val&0x01) != 0)
		val = (val >> 1) | uint8(carry)
		c.Memory[addr] = val
		c = SetZN(c, val)
	} else if opcode == OpRORAbsX {
		var addr int
		c, addr = FetchWord(c)
		carry := 0
		if GetCarry(c) {
			carry = 0x80
		}
		effAddr := addr + int(c.X)
		val := c.Memory[effAddr]
		c = SetCarry(c, (val&0x01) != 0)
		val = (val >> 1) | uint8(carry)
		c.Memory[effAddr] = val
		c = SetZN(c, val)

	// INC - Increment Memory
	} else if opcode == OpINC {
		var addr uint8
		c, addr = FetchByte(c)
		val := c.Memory[int(addr)] + 1
		c.Memory[int(addr)] = val
		c = SetZN(c, val)
	} else if opcode == OpINCZpX {
		var addr uint8
		c, addr = FetchByte(c)
		effAddr := int(addr+c.X) & 0xFF
		val := c.Memory[effAddr] + 1
		c.Memory[effAddr] = val
		c = SetZN(c, val)
	} else if opcode == OpINCAbs {
		var addr int
		c, addr = FetchWord(c)
		val := c.Memory[addr] + 1
		c.Memory[addr] = val
		c = SetZN(c, val)

	// DEC - Decrement Memory
	} else if opcode == OpDECZp {
		var addr uint8
		c, addr = FetchByte(c)
		val := c.Memory[int(addr)] - 1
		c.Memory[int(addr)] = val
		c = SetZN(c, val)
	} else if opcode == OpDECZpX {
		var addr uint8
		c, addr = FetchByte(c)
		effAddr := int(addr+c.X) & 0xFF
		val := c.Memory[effAddr] - 1
		c.Memory[effAddr] = val
		c = SetZN(c, val)
	} else if opcode == OpDECAbs {
		var addr int
		c, addr = FetchWord(c)
		val := c.Memory[addr] - 1
		c.Memory[addr] = val
		c = SetZN(c, val)

	// INX, INY, DEX, DEY
	} else if opcode == OpINX {
		c.X = c.X + 1
		c = SetZN(c, c.X)
	} else if opcode == OpINY {
		c.Y = c.Y + 1
		c = SetZN(c, c.Y)
	} else if opcode == OpDEX {
		c.X = c.X - 1
		c = SetZN(c, c.X)
	} else if opcode == OpDEY {
		c.Y = c.Y - 1
		c = SetZN(c, c.Y)

	// CMP - Compare Accumulator
	} else if opcode == OpCMPImm {
		var value uint8
		c, value = FetchByte(c)
		c = doCMP(c, c.A, value)
	} else if opcode == OpCMPZp {
		var addr uint8
		c, addr = FetchByte(c)
		c = doCMP(c, c.A, c.Memory[int(addr)])
	} else if opcode == OpCMPZpX {
		var addr uint8
		c, addr = FetchByte(c)
		c = doCMP(c, c.A, c.Memory[int(addr+c.X)&0xFF])
	} else if opcode == OpCMPAbs {
		var addr int
		c, addr = FetchWord(c)
		c = doCMP(c, c.A, c.Memory[addr])
	} else if opcode == OpCMPAbsX {
		var addr int
		c, addr = FetchWord(c)
		c = doCMP(c, c.A, c.Memory[addr+int(c.X)])
	} else if opcode == OpCMPAbsY {
		var addr int
		c, addr = FetchWord(c)
		c = doCMP(c, c.A, c.Memory[addr+int(c.Y)])
	} else if opcode == OpCMPIndX {
		var zp uint8
		c, zp = FetchByte(c)
		c = doCMP(c, c.A, ReadIndirectX(c, zp))
	} else if opcode == OpCMPIndY {
		var zp uint8
		c, zp = FetchByte(c)
		c = doCMP(c, c.A, ReadIndirectY(c, zp))

	// CPX - Compare X Register
	} else if opcode == OpCPXImm {
		var value uint8
		c, value = FetchByte(c)
		c = doCMP(c, c.X, value)
	} else if opcode == OpCPXZp {
		var addr uint8
		c, addr = FetchByte(c)
		c = doCMP(c, c.X, c.Memory[int(addr)])
	} else if opcode == OpCPXAbs {
		var addr int
		c, addr = FetchWord(c)
		c = doCMP(c, c.X, c.Memory[addr])

	// CPY - Compare Y Register
	} else if opcode == OpCPYImm {
		var value uint8
		c, value = FetchByte(c)
		c = doCMP(c, c.Y, value)
	} else if opcode == OpCPYZp {
		var addr uint8
		c, addr = FetchByte(c)
		c = doCMP(c, c.Y, c.Memory[int(addr)])
	} else if opcode == OpCPYAbs {
		var addr int
		c, addr = FetchWord(c)
		c = doCMP(c, c.Y, c.Memory[addr])

	// BIT - Bit Test
	} else if opcode == OpBITZp {
		var addr uint8
		c, addr = FetchByte(c)
		val := c.Memory[int(addr)]
		c = SetZN(c, uint8(c.A&val))
		c = SetOverflow(c, (val&0x40) != 0)
		// Also set N from memory value, not result
		if (val & 0x80) != 0 {
			c.Status = c.Status | FlagN
		}
	} else if opcode == OpBITAbs {
		var addr int
		c, addr = FetchWord(c)
		val := c.Memory[addr]
		c = SetZN(c, uint8(c.A&val))
		c = SetOverflow(c, (val&0x40) != 0)
		if (val & 0x80) != 0 {
			c.Status = c.Status | FlagN
		}

	// Branch Instructions
	} else if opcode == OpBPL {
		var offset uint8
		c, offset = FetchByte(c)
		if !GetNegative(c) {
			c = Branch(c, offset)
		}
	} else if opcode == OpBMI {
		var offset uint8
		c, offset = FetchByte(c)
		if GetNegative(c) {
			c = Branch(c, offset)
		}
	} else if opcode == OpBVC {
		var offset uint8
		c, offset = FetchByte(c)
		if !GetOverflow(c) {
			c = Branch(c, offset)
		}
	} else if opcode == OpBVS {
		var offset uint8
		c, offset = FetchByte(c)
		if GetOverflow(c) {
			c = Branch(c, offset)
		}
	} else if opcode == OpBCC {
		var offset uint8
		c, offset = FetchByte(c)
		if !GetCarry(c) {
			c = Branch(c, offset)
		}
	} else if opcode == OpBCS {
		var offset uint8
		c, offset = FetchByte(c)
		if GetCarry(c) {
			c = Branch(c, offset)
		}
	} else if opcode == OpBNE {
		var offset uint8
		c, offset = FetchByte(c)
		if !GetZero(c) {
			c = Branch(c, offset)
		}
	} else if opcode == OpBEQ {
		var offset uint8
		c, offset = FetchByte(c)
		if GetZero(c) {
			c = Branch(c, offset)
		}

	// Jump Instructions
	} else if opcode == OpJMP {
		var addr int
		c, addr = FetchWord(c)
		c.PC = addr
	} else if opcode == OpJMPInd {
		var addr int
		c, addr = FetchWord(c)
		// 6502 bug: if addr is $xxFF, high byte is read from $xx00
		low := int(c.Memory[addr])
		high := int(c.Memory[(addr&0xFF00)|((addr+1)&0xFF)])
		c.PC = low + (high * 256)
	} else if opcode == OpJSR {
		var addr int
		c, addr = FetchWord(c)
		retAddr := c.PC - 1
		c = PushByte(c, uint8((retAddr>>8)&0xFF))
		c = PushByte(c, uint8(retAddr&0xFF))
		c.PC = addr
	} else if opcode == OpRTS {
		var low, high uint8
		c, low = PullByte(c)
		c, high = PullByte(c)
		c.PC = int(low) + (int(high) * 256) + 1
	} else if opcode == OpRTI {
		var status uint8
		c, status = PullByte(c)
		c.Status = status | 0x20 // Unused bit always set
		var low, high uint8
		c, low = PullByte(c)
		c, high = PullByte(c)
		c.PC = int(low) + (int(high) * 256)

	// Stack Instructions
	} else if opcode == OpPHA {
		c = PushByte(c, c.A)
	} else if opcode == OpPHP {
		c = PushByte(c, uint8(c.Status|FlagB|0x20)) // B flag set when pushed
	} else if opcode == OpPLA {
		c, c.A = PullByte(c)
		c = SetZN(c, c.A)
	} else if opcode == OpPLP {
		var status uint8
		c, status = PullByte(c)
		c.Status = uint8((status | 0x20) & 0xEF) // Unused bit set, B cleared

	// Transfer Instructions
	} else if opcode == OpTAX {
		c.X = c.A
		c = SetZN(c, c.X)
	} else if opcode == OpTXA {
		c.A = c.X
		c = SetZN(c, c.A)
	} else if opcode == OpTAY {
		c.Y = c.A
		c = SetZN(c, c.Y)
	} else if opcode == OpTYA {
		c.A = c.Y
		c = SetZN(c, c.A)
	} else if opcode == OpTSX {
		c.X = c.SP
		c = SetZN(c, c.X)
	} else if opcode == OpTXS {
		c.SP = c.X

	// Flag Instructions
	} else if opcode == OpCLC {
		c = SetCarry(c, false)
	} else if opcode == OpSEC {
		c = SetCarry(c, true)
	} else if opcode == OpCLI {
		c.Status = c.Status & (0xFF - FlagI)
	} else if opcode == OpSEI {
		c.Status = c.Status | FlagI
	} else if opcode == OpCLV {
		c = SetOverflow(c, false)
	} else if opcode == OpCLD {
		c.Status = c.Status & (0xFF - FlagD)
	} else if opcode == OpSED {
		c.Status = c.Status | FlagD

	// Other
	} else if opcode == OpNOP {
		// Do nothing
	} else if opcode == OpBRK {
		c.Halted = true
	}

	return c
}

// Helper function for ADC
func doADC(c CPU, value uint8) CPU {
	carry := 0
	if GetCarry(c) {
		carry = 1
	}
	result := int(c.A) + int(value) + carry
	// Set overflow flag
	overflow := ((c.A^value)&0x80) == 0 && ((c.A^uint8(result))&0x80) != 0
	c = SetOverflow(c, overflow)
	c = SetCarry(c, result > 255)
	c.A = uint8(result & 0xFF)
	c = SetZN(c, c.A)
	return c
}

// Helper function for SBC
func doSBC(c CPU, value uint8) CPU {
	carry := 0
	if GetCarry(c) {
		carry = 1
	}
	result := int(c.A) - int(value) - (1 - carry)
	// Set overflow flag
	overflow := ((c.A^value)&0x80) != 0 && ((c.A^uint8(result))&0x80) != 0
	c = SetOverflow(c, overflow)
	c = SetCarry(c, result >= 0)
	c.A = uint8(result & 0xFF)
	c = SetZN(c, c.A)
	return c
}

// Helper function for CMP/CPX/CPY
func doCMP(c CPU, reg uint8, value uint8) CPU {
	result := int(reg) - int(value)
	c = SetCarry(c, reg >= value)
	c = SetZN(c, uint8(result&0xFF))
	return c
}

// Run executes instructions until halted or max cycles reached
func Run(c CPU, maxCycles int) CPU {
	for {
		if c.Halted {
			break
		}
		if c.Cycles >= maxCycles {
			break
		}
		c = Step(c)
	}
	return c
}

// GetScreenPixel returns the color value at screen position (x, y)
func GetScreenPixel(c CPU, x int, y int) uint8 {
	if x < 0 {
		return 0
	}
	if x >= ScreenWidth {
		return 0
	}
	if y < 0 {
		return 0
	}
	if y >= ScreenHeight {
		return 0
	}
	addr := ScreenBase + (y * ScreenWidth) + x
	return c.Memory[addr]
}

// IsHalted returns true if the CPU has halted
func IsHalted(c CPU) bool {
	return c.Halted
}

// GetMemory returns the byte at the specified memory address
func GetMemory(c CPU, addr int) uint8 {
	if addr < 0 {
		return 0
	}
	if addr >= 65536 {
		return 0
	}
	return c.Memory[addr]
}

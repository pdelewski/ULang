package cpu

// MOS 6502 CPU Emulator
// A simple interpreter supporting basic instructions for drawing shapes

// Opcodes
const (
	// Load/Store
	OpLDAImm = 0xA9 // LDA #value
	OpLDAZp  = 0xA5 // LDA $zp
	OpLDAZpX = 0xB5 // LDA $zp,X
	OpLDAAbs = 0xAD // LDA $addr
	OpLDXImm = 0xA2 // LDX #value
	OpLDXAbs = 0xAE // LDX $addr
	OpLDYImm = 0xA0 // LDY #value
	OpLDYAbs = 0xAC // LDY $addr
	OpSTAZp  = 0x85 // STA $zp
	OpSTAZpX = 0x95 // STA $zp,X
	OpSTAAbs = 0x8D // STA $addr
	OpSTXAbs = 0x8E // STX $addr
	OpSTYAbs = 0x8C // STY $addr

	// Arithmetic
	OpADCImm = 0x69 // ADC #value
	OpSBCImm = 0xE9 // SBC #value

	// Increment/Decrement
	OpINX = 0xE8 // INX
	OpINY = 0xC8 // INY
	OpDEX = 0xCA // DEX
	OpDEY = 0x88 // DEY
	OpINC = 0xE6 // INC $zp

	// Compare
	OpCMPImm = 0xC9 // CMP #value
	OpCPXImm = 0xE0 // CPX #value
	OpCPYImm = 0xC0 // CPY #value

	// Branch
	OpBNE = 0xD0 // BNE offset
	OpBEQ = 0xF0 // BEQ offset
	OpBCC = 0x90 // BCC offset
	OpBCS = 0xB0 // BCS offset

	// Jump
	OpJMP = 0x4C // JMP $addr
	OpJSR = 0x20 // JSR $addr
	OpRTS = 0x60 // RTS

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

// Step executes one instruction and returns updated CPU
func Step(c CPU) CPU {
	if c.Halted {
		return c
	}

	var opcode uint8
	c, opcode = FetchByte(c)
	c.Cycles = c.Cycles + 1

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
		c.A = c.Memory[int(addr+c.X)]
		c = SetZN(c, c.A)
	} else if opcode == OpLDAAbs {
		var addr int
		c, addr = FetchWord(c)
		c.A = c.Memory[addr]
		c = SetZN(c, c.A)
	} else if opcode == OpLDXImm {
		var value uint8
		c, value = FetchByte(c)
		c.X = value
		c = SetZN(c, c.X)
	} else if opcode == OpLDXAbs {
		var addr int
		c, addr = FetchWord(c)
		c.X = c.Memory[addr]
		c = SetZN(c, c.X)
	} else if opcode == OpLDYImm {
		var value uint8
		c, value = FetchByte(c)
		c.Y = value
		c = SetZN(c, c.Y)
	} else if opcode == OpLDYAbs {
		var addr int
		c, addr = FetchWord(c)
		c.Y = c.Memory[addr]
		c = SetZN(c, c.Y)
	} else if opcode == OpSTAZp {
		var addr uint8
		c, addr = FetchByte(c)
		c.Memory[int(addr)] = c.A
	} else if opcode == OpSTAZpX {
		var addr uint8
		c, addr = FetchByte(c)
		c.Memory[int(addr+c.X)] = c.A
	} else if opcode == OpSTAAbs {
		var addr int
		c, addr = FetchWord(c)
		c.Memory[addr] = c.A
	} else if opcode == OpSTXAbs {
		var addr int
		c, addr = FetchWord(c)
		c.Memory[addr] = c.X
	} else if opcode == OpSTYAbs {
		var addr int
		c, addr = FetchWord(c)
		c.Memory[addr] = c.Y
	} else if opcode == OpADCImm {
		var value uint8
		c, value = FetchByte(c)
		carry := 0
		if GetCarry(c) {
			carry = 1
		}
		result := int(c.A) + int(value) + carry
		c = SetCarry(c, result > 255)
		c.A = uint8(result & 0xFF)
		c = SetZN(c, c.A)
	} else if opcode == OpSBCImm {
		var value uint8
		c, value = FetchByte(c)
		carry := 0
		if GetCarry(c) {
			carry = 1
		}
		result := int(c.A) - int(value) - (1 - carry)
		c = SetCarry(c, result >= 0)
		c.A = uint8(result & 0xFF)
		c = SetZN(c, c.A)
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
	} else if opcode == OpINC {
		var addr uint8
		c, addr = FetchByte(c)
		val := c.Memory[int(addr)] + 1
		c.Memory[int(addr)] = val
		c = SetZN(c, val)
	} else if opcode == OpCMPImm {
		var value uint8
		c, value = FetchByte(c)
		result := int(c.A) - int(value)
		c = SetCarry(c, c.A >= value)
		c = SetZN(c, uint8(result&0xFF))
	} else if opcode == OpCPXImm {
		var value uint8
		c, value = FetchByte(c)
		result := int(c.X) - int(value)
		c = SetCarry(c, c.X >= value)
		c = SetZN(c, uint8(result&0xFF))
	} else if opcode == OpCPYImm {
		var value uint8
		c, value = FetchByte(c)
		result := int(c.Y) - int(value)
		c = SetCarry(c, c.Y >= value)
		c = SetZN(c, uint8(result&0xFF))
	} else if opcode == OpBNE {
		var offset uint8
		c, offset = FetchByte(c)
		if !GetZero(c) {
			if offset < 128 {
				c.PC = c.PC + int(offset)
			} else {
				c.PC = c.PC - (256 - int(offset))
			}
		}
	} else if opcode == OpBEQ {
		var offset uint8
		c, offset = FetchByte(c)
		if GetZero(c) {
			if offset < 128 {
				c.PC = c.PC + int(offset)
			} else {
				c.PC = c.PC - (256 - int(offset))
			}
		}
	} else if opcode == OpBCC {
		var offset uint8
		c, offset = FetchByte(c)
		if !GetCarry(c) {
			if offset < 128 {
				c.PC = c.PC + int(offset)
			} else {
				c.PC = c.PC - (256 - int(offset))
			}
		}
	} else if opcode == OpBCS {
		var offset uint8
		c, offset = FetchByte(c)
		if GetCarry(c) {
			if offset < 128 {
				c.PC = c.PC + int(offset)
			} else {
				c.PC = c.PC - (256 - int(offset))
			}
		}
	} else if opcode == OpJMP {
		var addr int
		c, addr = FetchWord(c)
		c.PC = addr
	} else if opcode == OpJSR {
		var addr int
		c, addr = FetchWord(c)
		// Push return address - 1
		retAddr := c.PC - 1
		c.Memory[0x100+int(c.SP)] = uint8((retAddr >> 8) & 0xFF)
		c.SP = c.SP - 1
		c.Memory[0x100+int(c.SP)] = uint8(retAddr & 0xFF)
		c.SP = c.SP - 1
		c.PC = addr
	} else if opcode == OpRTS {
		c.SP = c.SP + 1
		low := int(c.Memory[0x100+int(c.SP)])
		c.SP = c.SP + 1
		high := int(c.Memory[0x100+int(c.SP)])
		c.PC = (high * 256) + low + 1
	} else if opcode == OpNOP {
		// Do nothing
	} else if opcode == OpBRK {
		c.Halted = true
	}

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

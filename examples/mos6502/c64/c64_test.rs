use std::any::Any;
use std::fmt;
use std::rc::Rc;

// Type aliases (Go-style)
type Int8 = i8;
type Int16 = i16;
type Int32 = i32;
type Int64 = i64;
type Uint8 = u8;
type Uint16 = u16;
type Uint32 = u32;
type Uint64 = u64;

// println equivalents - multiple versions for different arg counts
pub fn println<T: fmt::Display>(val: T) {
    std::println!("{}", val);
}

pub fn println0() {
    std::println!();
}

// printf - multiple versions for different arg counts
pub fn printf<T: fmt::Display>(val: T) {
    print!("{}", val);
}

pub fn printf2<T: fmt::Display>(fmt_str: String, val: T) {
    // Convert C-style format to Rust format
    let rust_fmt = fmt_str
        .replace("%d", "{}")
        .replace("%s", "{}")
        .replace("%v", "{}");
    let result = rust_fmt.replace("{}", &format!("{}", val));
    print!("{}", result);
}

pub fn printf3<T1: fmt::Display, T2: fmt::Display>(fmt_str: String, v1: T1, v2: T2) {
    let rust_fmt = fmt_str
        .replace("%d", "{}")
        .replace("%s", "{}")
        .replace("%v", "{}");
    let result =
        rust_fmt
            .replacen("{}", &format!("{}", v1), 1)
            .replacen("{}", &format!("{}", v2), 1);
    print!("{}", result);
}

pub fn printf4<T1: fmt::Display, T2: fmt::Display, T3: fmt::Display>(
    fmt_str: String,
    v1: T1,
    v2: T2,
    v3: T3,
) {
    let rust_fmt = fmt_str
        .replace("%d", "{}")
        .replace("%s", "{}")
        .replace("%v", "{}");
    let result = rust_fmt
        .replacen("{}", &format!("{}", v1), 1)
        .replacen("{}", &format!("{}", v2), 1)
        .replacen("{}", &format!("{}", v3), 1);
    print!("{}", result);
}

pub fn printf5<T1: fmt::Display, T2: fmt::Display, T3: fmt::Display, T4: fmt::Display>(
    fmt_str: String,
    v1: T1,
    v2: T2,
    v3: T3,
    v4: T4,
) {
    let rust_fmt = fmt_str
        .replace("%d", "{}")
        .replace("%s", "{}")
        .replace("%v", "{}");
    let result = rust_fmt
        .replacen("{}", &format!("{}", v1), 1)
        .replacen("{}", &format!("{}", v2), 1)
        .replacen("{}", &format!("{}", v3), 1)
        .replacen("{}", &format!("{}", v4), 1);
    print!("{}", result);
}

// Print byte as character (for %c format)
pub fn printc(b: i8) {
    print!("{}", b as u8 as char);
}

// Convert byte to character string (for Sprintf %c format)
pub fn byte_to_char(b: i8) -> String {
    (b as u8 as char).to_string()
}

// Go-style append (returns a new Vec)
pub fn append<T: Clone>(vec: &Vec<T>, value: T) -> Vec<T> {
    let mut new_vec = vec.clone();
    new_vec.push(value);
    new_vec
}

pub fn append_many<T: Clone>(vec: &Vec<T>, values: &[T]) -> Vec<T> {
    let mut new_vec = vec.clone();
    new_vec.extend_from_slice(values);
    new_vec
}

// Simple string_format using format!
pub fn string_format(fmt_str: &str, args: &[&dyn fmt::Display]) -> String {
    let mut result = String::new();
    let mut split = fmt_str.split("{}");
    for (i, segment) in split.enumerate() {
        result.push_str(segment);
        if i < args.len() {
            result.push_str(&format!("{}", args[i]));
        }
    }
    result
}

// string_format for 2 args (format string + 1 value)
pub fn string_format2<T: fmt::Display>(fmt_str: &str, val: T) -> String {
    let rust_fmt = fmt_str
        .replace("%d", "{}")
        .replace("%s", "{}")
        .replace("%v", "{}");
    rust_fmt.replace("{}", &format!("{}", val))
}

pub fn len<T>(slice: &[T]) -> i32 {
    slice.len() as i32
}
pub mod cpu {
    use crate::*;

    #[derive(Default, Clone, Debug)]
    pub struct CPU {
        pub A: u8,
        pub X: u8,
        pub Y: u8,
        pub SP: u8,
        pub PC: i32,
        pub Status: u8,
        pub Memory: Vec<u8>,
        pub Halted: bool,
        pub Cycles: i32,
    }

    pub const OpLDAImm: i32 = 0xA9;
    pub const OpLDAZp: i32 = 0xA5;
    pub const OpLDAZpX: i32 = 0xB5;
    pub const OpLDAAbs: i32 = 0xAD;
    pub const OpLDXImm: i32 = 0xA2;
    pub const OpLDXAbs: i32 = 0xAE;
    pub const OpLDYImm: i32 = 0xA0;
    pub const OpLDYAbs: i32 = 0xAC;
    pub const OpSTAZp: i32 = 0x85;
    pub const OpSTAZpX: i32 = 0x95;
    pub const OpSTAAbs: i32 = 0x8D;
    pub const OpSTXAbs: i32 = 0x8E;
    pub const OpSTYAbs: i32 = 0x8C;
    pub const OpADCImm: i32 = 0x69;
    pub const OpSBCImm: i32 = 0xE9;
    pub const OpINX: i32 = 0xE8;
    pub const OpINY: i32 = 0xC8;
    pub const OpDEX: i32 = 0xCA;
    pub const OpDEY: i32 = 0x88;
    pub const OpINC: i32 = 0xE6;
    pub const OpCMPImm: i32 = 0xC9;
    pub const OpCPXImm: i32 = 0xE0;
    pub const OpCPYImm: i32 = 0xC0;
    pub const OpBNE: i32 = 0xD0;
    pub const OpBEQ: i32 = 0xF0;
    pub const OpBCC: i32 = 0x90;
    pub const OpBCS: i32 = 0xB0;
    pub const OpJMP: i32 = 0x4C;
    pub const OpJSR: i32 = 0x20;
    pub const OpRTS: i32 = 0x60;
    pub const OpNOP: i32 = 0xEA;
    pub const OpBRK: i32 = 0x00;

    pub const FlagC: i32 = 0x01;
    pub const FlagZ: i32 = 0x02;
    pub const FlagI: i32 = 0x04;
    pub const FlagD: i32 = 0x08;
    pub const FlagB: i32 = 0x10;
    pub const FlagV: i32 = 0x40;
    pub const FlagN: i32 = 0x80;

    pub const ScreenBase: i32 = 0x0200;
    pub const ScreenWidth: i32 = 32;
    pub const ScreenHeight: i32 = 32;
    pub const ScreenSize: i32 = 1024;

    pub fn NewCPU() -> CPU {
        let mut mem: Vec<u8> = Vec::new();
        let mut i = 0;
        loop {
            if (i >= 65536) {
                break;
            }
            mem = append(&mem.clone(), (0 as u8));
            i = (i + 1);
        }
        return CPU {
            A: 0,
            X: 0,
            Y: 0,
            SP: 0xFF,
            PC: 0x0600,
            Status: 0x20,
            Memory: mem.clone(),
            Halted: false,
            Cycles: 0,
            ..Default::default()
        };
    }

    pub fn LoadProgram(mut c: CPU, mut program: Vec<u8>, addr: i32) -> CPU {
        let mut i = 0;
        loop {
            if (i >= len(&program.clone())) {
                break;
            }
            c.Memory[(addr + i) as usize] = program[i as usize];
            i = (i + 1);
        }
        return c;
    }

    pub fn SetPC(mut c: CPU, addr: i32) -> CPU {
        c.PC = addr;
        return c;
    }

    pub fn ReadByte(mut c: CPU, addr: i32) -> u8 {
        return c.Memory[addr as usize];
    }

    pub fn WriteByte(mut c: CPU, addr: i32, value: u8) -> CPU {
        c.Memory[addr as usize] = value;
        return c;
    }

    pub fn FetchByte(mut c: CPU) -> (CPU, u8) {
        let mut value = c.Memory[c.PC as usize];
        c.PC = (c.PC + 1);
        return (c.clone(), value);
    }

    pub fn FetchWord(mut c: CPU) -> (CPU, i32) {
        let mut low = (c.Memory[c.PC as usize] as i32);
        let mut high = (c.Memory[(c.PC + 1) as usize] as i32);
        c.PC = (c.PC + 2);
        return (c.clone(), (low + (high * 256)));
    }

    pub fn SetZN(mut c: CPU, value: u8) -> CPU {
        if (value == 0) {
            c.Status = (c.Status | FlagZ as u8);
        } else {
            c.Status = (c.Status & (0xFF - FlagZ) as u8);
        }
        if ((value & 0x80) != 0) {
            c.Status = (c.Status | FlagN as u8);
        } else {
            c.Status = (c.Status & (0xFF - FlagN) as u8);
        }
        return c;
    }

    pub fn SetCarry(mut c: CPU, set: bool) -> CPU {
        if (set) {
            c.Status = (c.Status | FlagC as u8);
        } else {
            c.Status = (c.Status & (0xFF - FlagC) as u8);
        }
        return c;
    }

    pub fn GetCarry(mut c: CPU) -> bool {
        return ((c.Status & FlagC as u8) != 0);
    }

    pub fn GetZero(mut c: CPU) -> bool {
        return ((c.Status & FlagZ as u8) != 0);
    }

    pub fn Step(mut c: CPU) -> CPU {
        if (c.Halted) {
            return c;
        }
        let mut opcode: u8 = 0;
        (c, opcode) = FetchByte(c.clone());
        c.Cycles = (c.Cycles + 1);
        if (opcode as i32 == OpLDAImm) {
            let mut value: u8 = 0;
            (c, value) = FetchByte(c.clone());
            c.A = value;
            c = SetZN(c.clone(), c.A);
        } else if (opcode as i32 == OpLDAZp) {
            let mut addr: u8 = 0;
            (c, addr) = FetchByte(c.clone());
            c.A = c.Memory[(addr as i32) as usize];
            c = SetZN(c.clone(), c.A);
        } else if (opcode as i32 == OpLDAZpX) {
            let mut addr: u8 = 0;
            (c, addr) = FetchByte(c.clone());
            c.A = c.Memory[((addr + c.X) as i32) as usize];
            c = SetZN(c.clone(), c.A);
        } else if (opcode as i32 == OpLDAAbs) {
            let mut addr: i32 = 0;
            (c, addr) = FetchWord(c.clone());
            c.A = c.Memory[addr as usize];
            c = SetZN(c.clone(), c.A);
        } else if (opcode as i32 == OpLDXImm) {
            let mut value: u8 = 0;
            (c, value) = FetchByte(c.clone());
            c.X = value;
            c = SetZN(c.clone(), c.X);
        } else if (opcode as i32 == OpLDXAbs) {
            let mut addr: i32 = 0;
            (c, addr) = FetchWord(c.clone());
            c.X = c.Memory[addr as usize];
            c = SetZN(c.clone(), c.X);
        } else if (opcode as i32 == OpLDYImm) {
            let mut value: u8 = 0;
            (c, value) = FetchByte(c.clone());
            c.Y = value;
            c = SetZN(c.clone(), c.Y);
        } else if (opcode as i32 == OpLDYAbs) {
            let mut addr: i32 = 0;
            (c, addr) = FetchWord(c.clone());
            c.Y = c.Memory[addr as usize];
            c = SetZN(c.clone(), c.Y);
        } else if (opcode as i32 == OpSTAZp) {
            let mut addr: u8 = 0;
            (c, addr) = FetchByte(c.clone());
            c.Memory[(addr as i32) as usize] = c.A;
        } else if (opcode as i32 == OpSTAZpX) {
            let mut addr: u8 = 0;
            (c, addr) = FetchByte(c.clone());
            c.Memory[((addr + c.X) as i32) as usize] = c.A;
        } else if (opcode as i32 == OpSTAAbs) {
            let mut addr: i32 = 0;
            (c, addr) = FetchWord(c.clone());
            c.Memory[addr as usize] = c.A;
        } else if (opcode as i32 == OpSTXAbs) {
            let mut addr: i32 = 0;
            (c, addr) = FetchWord(c.clone());
            c.Memory[addr as usize] = c.X;
        } else if (opcode as i32 == OpSTYAbs) {
            let mut addr: i32 = 0;
            (c, addr) = FetchWord(c.clone());
            c.Memory[addr as usize] = c.Y;
        } else if (opcode as i32 == OpADCImm) {
            let mut value: u8 = 0;
            (c, value) = FetchByte(c.clone());
            let mut carry = 0;
            if (GetCarry(c.clone())) {
                carry = 1;
            }
            let mut result = (((c.A as i32) + (value as i32)) + carry);
            c = SetCarry(c.clone(), (result > 255));
            c.A = ((result & 0xFF) as u8);
            c = SetZN(c.clone(), c.A);
        } else if (opcode as i32 == OpSBCImm) {
            let mut value: u8 = 0;
            (c, value) = FetchByte(c.clone());
            let mut carry = 0;
            if (GetCarry(c.clone())) {
                carry = 1;
            }
            let mut result = (((c.A as i32) - (value as i32)) - (1 - carry));
            c = SetCarry(c.clone(), (result >= 0));
            c.A = ((result & 0xFF) as u8);
            c = SetZN(c.clone(), c.A);
        } else if (opcode as i32 == OpINX) {
            c.X = (c.X + 1);
            c = SetZN(c.clone(), c.X);
        } else if (opcode as i32 == OpINY) {
            c.Y = (c.Y + 1);
            c = SetZN(c.clone(), c.Y);
        } else if (opcode as i32 == OpDEX) {
            c.X = (c.X - 1);
            c = SetZN(c.clone(), c.X);
        } else if (opcode as i32 == OpDEY) {
            c.Y = (c.Y - 1);
            c = SetZN(c.clone(), c.Y);
        } else if (opcode as i32 == OpINC) {
            let mut addr: u8 = 0;
            (c, addr) = FetchByte(c.clone());
            let mut val = (c.Memory[(addr as i32) as usize] + 1);
            c.Memory[(addr as i32) as usize] = val;
            c = SetZN(c.clone(), val);
        } else if (opcode as i32 == OpCMPImm) {
            let mut value: u8 = 0;
            (c, value) = FetchByte(c.clone());
            let mut result = ((c.A as i32) - (value as i32));
            c = SetCarry(c.clone(), (c.A >= value));
            c = SetZN(c.clone(), ((result & 0xFF) as u8));
        } else if (opcode as i32 == OpCPXImm) {
            let mut value: u8 = 0;
            (c, value) = FetchByte(c.clone());
            let mut result = ((c.X as i32) - (value as i32));
            c = SetCarry(c.clone(), (c.X >= value));
            c = SetZN(c.clone(), ((result & 0xFF) as u8));
        } else if (opcode as i32 == OpCPYImm) {
            let mut value: u8 = 0;
            (c, value) = FetchByte(c.clone());
            let mut result = ((c.Y as i32) - (value as i32));
            c = SetCarry(c.clone(), (c.Y >= value));
            c = SetZN(c.clone(), ((result & 0xFF) as u8));
        } else if (opcode as i32 == OpBNE) {
            let mut offset: u8 = 0;
            (c, offset) = FetchByte(c.clone());
            if (!GetZero(c.clone())) {
                if (offset < 128) {
                    c.PC = (c.PC + (offset as i32));
                } else {
                    c.PC = (c.PC - (256 - (offset as i32)));
                }
            }
        } else if (opcode as i32 == OpBEQ) {
            let mut offset: u8 = 0;
            (c, offset) = FetchByte(c.clone());
            if (GetZero(c.clone())) {
                if (offset < 128) {
                    c.PC = (c.PC + (offset as i32));
                } else {
                    c.PC = (c.PC - (256 - (offset as i32)));
                }
            }
        } else if (opcode as i32 == OpBCC) {
            let mut offset: u8 = 0;
            (c, offset) = FetchByte(c.clone());
            if (!GetCarry(c.clone())) {
                if (offset < 128) {
                    c.PC = (c.PC + (offset as i32));
                } else {
                    c.PC = (c.PC - (256 - (offset as i32)));
                }
            }
        } else if (opcode as i32 == OpBCS) {
            let mut offset: u8 = 0;
            (c, offset) = FetchByte(c.clone());
            if (GetCarry(c.clone())) {
                if (offset < 128) {
                    c.PC = (c.PC + (offset as i32));
                } else {
                    c.PC = (c.PC - (256 - (offset as i32)));
                }
            }
        } else if (opcode as i32 == OpJMP) {
            let mut addr: i32 = 0;
            (c, addr) = FetchWord(c.clone());
            c.PC = addr;
        } else if (opcode as i32 == OpJSR) {
            let mut addr: i32 = 0;
            (c, addr) = FetchWord(c.clone());
            let mut retAddr = (c.PC - 1);
            c.Memory[(0x100 + (c.SP as i32)) as usize] = (((retAddr >> 8) & 0xFF) as u8);
            c.SP = (c.SP - 1);
            c.Memory[(0x100 + (c.SP as i32)) as usize] = ((retAddr & 0xFF) as u8);
            c.SP = (c.SP - 1);
            c.PC = addr;
        } else if (opcode as i32 == OpRTS) {
            c.SP = (c.SP + 1);
            let mut low = (c.Memory[(0x100 + (c.SP as i32)) as usize] as i32);
            c.SP = (c.SP + 1);
            let mut high = (c.Memory[(0x100 + (c.SP as i32)) as usize] as i32);
            c.PC = (((high * 256) + low) + 1);
        } else if (opcode as i32 == OpNOP) {
        } else if (opcode as i32 == OpBRK) {
            c.Halted = true;
        }
        return c;
    }

    pub fn Run(mut c: CPU, maxCycles: i32) -> CPU {
        loop {
            if (c.Halted) {
                break;
            }
            if (c.Cycles >= maxCycles) {
                break;
            }
            c = Step(c.clone());
        }
        return c;
    }

    pub fn GetScreenPixel(mut c: CPU, x: i32, y: i32) -> u8 {
        if (x < 0) {
            return 0;
        }
        if (x >= ScreenWidth) {
            return 0;
        }
        if (y < 0) {
            return 0;
        }
        if (y >= ScreenHeight) {
            return 0;
        }
        let mut addr = ((ScreenBase + (y * ScreenWidth)) + x);
        return c.Memory[addr as usize];
    }

    pub fn IsHalted(mut c: CPU) -> bool {
        return c.Halted;
    }

    pub fn GetMemory(mut c: CPU, addr: i32) -> u8 {
        if (addr < 0) {
            return 0;
        }
        if (addr >= 65536) {
            return 0;
        }
        return c.Memory[addr as usize];
    }
} // pub mod cpu

pub mod assembler {
    use crate::*;

    #[derive(Default, Clone, Debug)]
    pub struct Instruction {
        pub OpcodeBytes: Vec<i8>,
        pub Mode: i8,
        pub Operand: i32,
        pub LabelBytes: Vec<i8>,
        pub HasLabel: bool,
    }

    #[derive(Default, Clone, Debug)]
    pub struct Token {
        pub Type: i8,
        pub Representation: Vec<i8>,
    }

    pub const TokenTypeInstruction: i32 = 1;
    pub const TokenTypeNumber: i32 = 2;
    pub const TokenTypeLabel: i32 = 3;
    pub const TokenTypeComma: i32 = 4;
    pub const TokenTypeNewline: i32 = 5;
    pub const TokenTypeHash: i32 = 6;
    pub const TokenTypeDollar: i32 = 7;
    pub const TokenTypeColon: i32 = 8;
    pub const TokenTypeIdentifier: i32 = 9;
    pub const TokenTypeComment: i32 = 10;

    pub const ModeImplied: i32 = 0;
    pub const ModeImmediate: i32 = 1;
    pub const ModeZeroPage: i32 = 2;
    pub const ModeAbsolute: i32 = 3;
    pub const ModeZeroPageX: i32 = 4;

    pub fn IsDigit(b: i8) -> bool {
        return ((b >= 48) && (b <= 57));
    }

    pub fn IsHexDigit(b: i8) -> bool {
        return ((IsDigit(b) || ((b >= 97) && (b <= 102))) || ((b >= 65) && (b <= 70)));
    }

    pub fn IsAlpha(b: i8) -> bool {
        return ((((b >= 97) && (b <= 122)) || ((b >= 65) && (b <= 90))) || (b == 95));
    }

    pub fn IsWhitespace(b: i8) -> bool {
        return ((b == 32) || (b == 9));
    }

    pub fn StringToBytes(s: String) -> Vec<i8> {
        let mut result: Vec<i8> = Vec::new();
        let mut i = 0;
        loop {
            if (i >= s.clone().len() as i32) {
                break;
            }
            result = append(&result.clone(), (s.as_bytes()[i as usize] as i8));
            i = (i + 1);
        }
        return result;
    }

    pub fn ToUpper(b: i8) -> i8 {
        if ((b >= 97) && (b <= 122)) {
            return (b - 32);
        }
        return b;
    }

    pub fn Tokenize(text: String) -> Vec<Token> {
        let mut tokens: Vec<Token> = Vec::new();
        let mut bytes = StringToBytes(text.clone());
        let mut i = 0;
        loop {
            if (i >= len(&bytes.clone())) {
                break;
            }
            let mut b = bytes[i as usize];
            if (IsWhitespace(b)) {
                i = (i + 1);
                continue;
            }
            if (b == 10) {
                tokens = append(
                    &tokens.clone(),
                    Token {
                        Type: TokenTypeNewline as i8,
                        Representation: vec![b].clone(),
                        ..Default::default()
                    }
                    .clone(),
                );
                i = (i + 1);
                continue;
            }
            if (b == 59) {
                loop {
                    if (i >= len(&bytes.clone())) {
                        break;
                    }
                    if (bytes[i as usize] == 10) {
                        break;
                    }
                    i = (i + 1);
                }
                continue;
            }
            if (b == 35) {
                tokens = append(
                    &tokens.clone(),
                    Token {
                        Type: TokenTypeHash as i8,
                        Representation: vec![b].clone(),
                        ..Default::default()
                    }
                    .clone(),
                );
                i = (i + 1);
                continue;
            }
            if (b == 36) {
                tokens = append(
                    &tokens.clone(),
                    Token {
                        Type: TokenTypeDollar as i8,
                        Representation: vec![b].clone(),
                        ..Default::default()
                    }
                    .clone(),
                );
                i = (i + 1);
                continue;
            }
            if (b == 58) {
                tokens = append(
                    &tokens.clone(),
                    Token {
                        Type: TokenTypeColon as i8,
                        Representation: vec![b].clone(),
                        ..Default::default()
                    }
                    .clone(),
                );
                i = (i + 1);
                continue;
            }
            if (b == 44) {
                tokens = append(
                    &tokens.clone(),
                    Token {
                        Type: TokenTypeComma as i8,
                        Representation: vec![b].clone(),
                        ..Default::default()
                    }
                    .clone(),
                );
                i = (i + 1);
                continue;
            }
            if (IsHexDigit(b)) {
                let mut repr: Vec<i8> = Vec::new();
                loop {
                    if (i >= len(&bytes.clone())) {
                        break;
                    }
                    if (!IsHexDigit(bytes[i as usize])) {
                        break;
                    }
                    repr = append(&repr.clone(), bytes[i as usize]);
                    i = (i + 1);
                }
                tokens = append(
                    &tokens.clone(),
                    Token {
                        Type: TokenTypeNumber as i8,
                        Representation: repr.clone(),
                        ..Default::default()
                    }
                    .clone(),
                );
                continue;
            }
            if (IsAlpha(b)) {
                let mut repr: Vec<i8> = Vec::new();
                loop {
                    if (i >= len(&bytes.clone())) {
                        break;
                    }
                    if ((!IsAlpha(bytes[i as usize])) && (!IsDigit(bytes[i as usize]))) {
                        break;
                    }
                    repr = append(&repr.clone(), bytes[i as usize]);
                    i = (i + 1);
                }
                tokens = append(
                    &tokens.clone(),
                    Token {
                        Type: TokenTypeIdentifier as i8,
                        Representation: repr.clone(),
                        ..Default::default()
                    }
                    .clone(),
                );
                continue;
            }
            i = (i + 1);
        }
        return tokens;
    }

    pub fn ParseHex(mut bytes: Vec<i8>) -> i32 {
        let mut result = 0;
        let mut i = 0;
        loop {
            if (i >= len(&bytes.clone())) {
                break;
            }
            let mut b = bytes[i as usize];
            result = (result * 16);
            if ((b >= 48) && (b <= 57)) {
                result = (result + ((b - 48) as i32));
            } else if ((b >= 97) && (b <= 102)) {
                result = (result + (((b - 97) + 10) as i32));
            } else if ((b >= 65) && (b <= 70)) {
                result = (result + (((b - 65) + 10) as i32));
            }
            i = (i + 1);
        }
        return result;
    }

    pub fn ParseDecimal(mut bytes: Vec<i8>) -> i32 {
        let mut result = 0;
        let mut i = 0;
        loop {
            if (i >= len(&bytes.clone())) {
                break;
            }
            let mut b = bytes[i as usize];
            result = ((result * 10) + ((b - 48) as i32));
            i = (i + 1);
        }
        return result;
    }

    pub fn MatchToken(mut token: Token, s: String) -> bool {
        if (len(&token.Representation.clone()) != s.clone().len() as i32) {
            return false;
        }
        let mut i = 0;
        loop {
            if (i >= s.clone().len() as i32) {
                break;
            }
            if (ToUpper(token.Representation[i as usize])
                != ToUpper((s.as_bytes()[i as usize] as i8)))
            {
                return false;
            }
            i = (i + 1);
        }
        return true;
    }

    pub fn CopyBytes(mut src: Vec<i8>) -> Vec<i8> {
        let mut dst: Vec<i8> = Vec::new();
        let mut i = 0;
        loop {
            if (i >= len(&src.clone())) {
                break;
            }
            dst = append(&dst.clone(), src[i as usize]);
            i = (i + 1);
        }
        return dst;
    }

    pub fn Parse(mut tokens: Vec<Token>) -> Vec<Instruction> {
        let mut instructions: Vec<Instruction> = Vec::new();
        let mut i = 0;
        loop {
            if (i >= len(&tokens.clone())) {
                break;
            }
            if (tokens[i as usize].clone().Type as i32 == TokenTypeNewline) {
                i = (i + 1);
                continue;
            }
            let mut currentLabelBytes: Vec<i8> = Vec::new();
            let mut hasLabel = false;
            if (((tokens[i as usize].clone().Type as i32 == TokenTypeIdentifier)
                && ((i + 1) < len(&tokens.clone())))
                && (tokens[(i + 1) as usize].clone().Type as i32 == TokenTypeColon))
            {
                currentLabelBytes = CopyBytes(tokens[i as usize].clone().Representation.clone());
                hasLabel = true;
                i = (i + 2);
                loop {
                    if (i >= len(&tokens.clone())) {
                        break;
                    }
                    if (tokens[i as usize].clone().Type as i32 != TokenTypeNewline) {
                        break;
                    }
                    i = (i + 1);
                }
                if (i >= len(&tokens.clone())) {
                    break;
                }
            }
            if (tokens[i as usize].clone().Type as i32 != TokenTypeIdentifier) {
                i = (i + 1);
                continue;
            }
            let mut instr = Instruction {
                OpcodeBytes: CopyBytes(tokens[i as usize].clone().Representation.clone()).clone(),
                Mode: ModeImplied as i8,
                Operand: 0,
                LabelBytes: currentLabelBytes.clone(),
                HasLabel: hasLabel,
                ..Default::default()
            };
            i = (i + 1);
            if ((i < len(&tokens.clone()))
                && (tokens[i as usize].clone().Type as i32 != TokenTypeNewline))
            {
                if (tokens[i as usize].clone().Type as i32 == TokenTypeHash) {
                    i = (i + 1);
                    instr.Mode = ModeImmediate as i8;
                    if ((i < len(&tokens.clone()))
                        && (tokens[i as usize].clone().Type as i32 == TokenTypeDollar))
                    {
                        i = (i + 1);
                        if ((i < len(&tokens.clone()))
                            && (tokens[i as usize].clone().Type as i32 == TokenTypeNumber))
                        {
                            instr.Operand =
                                ParseHex(tokens[i as usize].clone().Representation.clone());
                            i = (i + 1);
                        }
                    } else if ((i < len(&tokens.clone()))
                        && (tokens[i as usize].clone().Type as i32 == TokenTypeNumber))
                    {
                        instr.Operand =
                            ParseDecimal(tokens[i as usize].clone().Representation.clone());
                        i = (i + 1);
                    }
                } else if (tokens[i as usize].clone().Type as i32 == TokenTypeDollar) {
                    i = (i + 1);
                    if ((i < len(&tokens.clone()))
                        && (tokens[i as usize].clone().Type as i32 == TokenTypeNumber))
                    {
                        instr.Operand = ParseHex(tokens[i as usize].clone().Representation.clone());
                        if (len(&tokens[i as usize].clone().Representation.clone()) <= 2) {
                            instr.Mode = ModeZeroPage as i8;
                        } else {
                            instr.Mode = ModeAbsolute as i8;
                        }
                        i = (i + 1);
                        if ((i < len(&tokens.clone()))
                            && (tokens[i as usize].clone().Type as i32 == TokenTypeComma))
                        {
                            i = (i + 1);
                            if ((i < len(&tokens.clone()))
                                && (tokens[i as usize].clone().Type as i32 == TokenTypeIdentifier))
                            {
                                if (MatchToken(tokens[i as usize].clone().clone(), "X".to_string()))
                                {
                                    instr.Mode = ModeZeroPageX as i8;
                                }
                                i = (i + 1);
                            }
                        }
                    }
                } else if (tokens[i as usize].clone().Type as i32 == TokenTypeNumber) {
                    instr.Operand = ParseDecimal(tokens[i as usize].clone().Representation.clone());
                    if (instr.Operand <= 255) {
                        instr.Mode = ModeZeroPage as i8;
                    } else {
                        instr.Mode = ModeAbsolute as i8;
                    }
                    i = (i + 1);
                }
            }
            instructions = append(&instructions.clone(), instr.clone());
        }
        return instructions;
    }

    pub fn IsOpcode(mut opcodeBytes: Vec<i8>, name: String) -> bool {
        if (len(&opcodeBytes.clone()) != name.clone().len() as i32) {
            return false;
        }
        let mut i = 0;
        loop {
            if (i >= name.clone().len() as i32) {
                break;
            }
            let mut ob = opcodeBytes[i as usize];
            if ((ob >= 97) && (ob <= 122)) {
                ob = (ob - 32);
            }
            let mut nb = (name.as_bytes()[i as usize] as i8);
            if (ob != nb) {
                return false;
            }
            i = (i + 1);
        }
        return true;
    }

    pub fn Assemble(mut instructions: Vec<Instruction>) -> Vec<u8> {
        let mut code: Vec<u8> = Vec::new();
        let mut idx = 0;
        loop {
            if (idx >= len(&instructions.clone())) {
                break;
            }
            let mut instr = instructions[idx as usize].clone();
            let mut opcodeBytes = instr.OpcodeBytes;
            if (IsOpcode(opcodeBytes.clone(), "LDA".to_string())) {
                if (instr.Mode as i32 == ModeImmediate) {
                    code = append(&code.clone(), (cpu::OpLDAImm as u8));
                    code = append(&code.clone(), (instr.Operand as u8));
                } else if (instr.Mode as i32 == ModeZeroPage) {
                    code = append(&code.clone(), (cpu::OpLDAZp as u8));
                    code = append(&code.clone(), (instr.Operand as u8));
                } else if (instr.Mode as i32 == ModeZeroPageX) {
                    code = append(&code.clone(), (cpu::OpLDAZpX as u8));
                    code = append(&code.clone(), (instr.Operand as u8));
                } else if (instr.Mode as i32 == ModeAbsolute) {
                    code = append(&code.clone(), (cpu::OpLDAAbs as u8));
                    code = append(&code.clone(), ((instr.Operand & 0xFF) as u8));
                    code = append(&code.clone(), (((instr.Operand >> 8) & 0xFF) as u8));
                }
            } else if (IsOpcode(opcodeBytes.clone(), "LDX".to_string())) {
                if (instr.Mode as i32 == ModeImmediate) {
                    code = append(&code.clone(), (cpu::OpLDXImm as u8));
                    code = append(&code.clone(), (instr.Operand as u8));
                } else if (instr.Mode as i32 == ModeAbsolute) {
                    code = append(&code.clone(), (cpu::OpLDXAbs as u8));
                    code = append(&code.clone(), ((instr.Operand & 0xFF) as u8));
                    code = append(&code.clone(), (((instr.Operand >> 8) & 0xFF) as u8));
                }
            } else if (IsOpcode(opcodeBytes.clone(), "LDY".to_string())) {
                if (instr.Mode as i32 == ModeImmediate) {
                    code = append(&code.clone(), (cpu::OpLDYImm as u8));
                    code = append(&code.clone(), (instr.Operand as u8));
                } else if (instr.Mode as i32 == ModeAbsolute) {
                    code = append(&code.clone(), (cpu::OpLDYAbs as u8));
                    code = append(&code.clone(), ((instr.Operand & 0xFF) as u8));
                    code = append(&code.clone(), (((instr.Operand >> 8) & 0xFF) as u8));
                }
            } else if (IsOpcode(opcodeBytes.clone(), "STA".to_string())) {
                if (instr.Mode as i32 == ModeZeroPage) {
                    code = append(&code.clone(), (cpu::OpSTAZp as u8));
                    code = append(&code.clone(), (instr.Operand as u8));
                } else if (instr.Mode as i32 == ModeZeroPageX) {
                    code = append(&code.clone(), (cpu::OpSTAZpX as u8));
                    code = append(&code.clone(), (instr.Operand as u8));
                } else if (instr.Mode as i32 == ModeAbsolute) {
                    code = append(&code.clone(), (cpu::OpSTAAbs as u8));
                    code = append(&code.clone(), ((instr.Operand & 0xFF) as u8));
                    code = append(&code.clone(), (((instr.Operand >> 8) & 0xFF) as u8));
                }
            } else if (IsOpcode(opcodeBytes.clone(), "STX".to_string())) {
                if (instr.Mode as i32 == ModeAbsolute) {
                    code = append(&code.clone(), (cpu::OpSTXAbs as u8));
                    code = append(&code.clone(), ((instr.Operand & 0xFF) as u8));
                    code = append(&code.clone(), (((instr.Operand >> 8) & 0xFF) as u8));
                }
            } else if (IsOpcode(opcodeBytes.clone(), "STY".to_string())) {
                if (instr.Mode as i32 == ModeAbsolute) {
                    code = append(&code.clone(), (cpu::OpSTYAbs as u8));
                    code = append(&code.clone(), ((instr.Operand & 0xFF) as u8));
                    code = append(&code.clone(), (((instr.Operand >> 8) & 0xFF) as u8));
                }
            } else if (IsOpcode(opcodeBytes.clone(), "ADC".to_string())) {
                if (instr.Mode as i32 == ModeImmediate) {
                    code = append(&code.clone(), (cpu::OpADCImm as u8));
                    code = append(&code.clone(), (instr.Operand as u8));
                }
            } else if (IsOpcode(opcodeBytes.clone(), "SBC".to_string())) {
                if (instr.Mode as i32 == ModeImmediate) {
                    code = append(&code.clone(), (cpu::OpSBCImm as u8));
                    code = append(&code.clone(), (instr.Operand as u8));
                }
            } else if (IsOpcode(opcodeBytes.clone(), "INX".to_string())) {
                code = append(&code.clone(), (cpu::OpINX as u8));
            } else if (IsOpcode(opcodeBytes.clone(), "INY".to_string())) {
                code = append(&code.clone(), (cpu::OpINY as u8));
            } else if (IsOpcode(opcodeBytes.clone(), "DEX".to_string())) {
                code = append(&code.clone(), (cpu::OpDEX as u8));
            } else if (IsOpcode(opcodeBytes.clone(), "DEY".to_string())) {
                code = append(&code.clone(), (cpu::OpDEY as u8));
            } else if (IsOpcode(opcodeBytes.clone(), "INC".to_string())) {
                if (instr.Mode as i32 == ModeZeroPage) {
                    code = append(&code.clone(), (cpu::OpINC as u8));
                    code = append(&code.clone(), (instr.Operand as u8));
                }
            } else if (IsOpcode(opcodeBytes.clone(), "CMP".to_string())) {
                if (instr.Mode as i32 == ModeImmediate) {
                    code = append(&code.clone(), (cpu::OpCMPImm as u8));
                    code = append(&code.clone(), (instr.Operand as u8));
                }
            } else if (IsOpcode(opcodeBytes.clone(), "CPX".to_string())) {
                if (instr.Mode as i32 == ModeImmediate) {
                    code = append(&code.clone(), (cpu::OpCPXImm as u8));
                    code = append(&code.clone(), (instr.Operand as u8));
                }
            } else if (IsOpcode(opcodeBytes.clone(), "CPY".to_string())) {
                if (instr.Mode as i32 == ModeImmediate) {
                    code = append(&code.clone(), (cpu::OpCPYImm as u8));
                    code = append(&code.clone(), (instr.Operand as u8));
                }
            } else if (IsOpcode(opcodeBytes.clone(), "BNE".to_string())) {
                code = append(&code.clone(), (cpu::OpBNE as u8));
                code = append(&code.clone(), (instr.Operand as u8));
            } else if (IsOpcode(opcodeBytes.clone(), "BEQ".to_string())) {
                code = append(&code.clone(), (cpu::OpBEQ as u8));
                code = append(&code.clone(), (instr.Operand as u8));
            } else if (IsOpcode(opcodeBytes.clone(), "BCC".to_string())) {
                code = append(&code.clone(), (cpu::OpBCC as u8));
                code = append(&code.clone(), (instr.Operand as u8));
            } else if (IsOpcode(opcodeBytes.clone(), "BCS".to_string())) {
                code = append(&code.clone(), (cpu::OpBCS as u8));
                code = append(&code.clone(), (instr.Operand as u8));
            } else if (IsOpcode(opcodeBytes.clone(), "JMP".to_string())) {
                code = append(&code.clone(), (cpu::OpJMP as u8));
                code = append(&code.clone(), ((instr.Operand & 0xFF) as u8));
                code = append(&code.clone(), (((instr.Operand >> 8) & 0xFF) as u8));
            } else if (IsOpcode(opcodeBytes.clone(), "JSR".to_string())) {
                code = append(&code.clone(), (cpu::OpJSR as u8));
                code = append(&code.clone(), ((instr.Operand & 0xFF) as u8));
                code = append(&code.clone(), (((instr.Operand >> 8) & 0xFF) as u8));
            } else if (IsOpcode(opcodeBytes.clone(), "RTS".to_string())) {
                code = append(&code.clone(), (cpu::OpRTS as u8));
            } else if (IsOpcode(opcodeBytes.clone(), "NOP".to_string())) {
                code = append(&code.clone(), (cpu::OpNOP as u8));
            } else if (IsOpcode(opcodeBytes.clone(), "BRK".to_string())) {
                code = append(&code.clone(), (cpu::OpBRK as u8));
            }
            idx = (idx + 1);
        }
        return code;
    }

    pub fn AssembleString(text: String) -> Vec<u8> {
        let mut tokens = Tokenize(text.clone());
        let mut instructions = Parse(tokens.clone());
        return Assemble(instructions.clone());
    }

    pub fn AppendLineBytes(mut allBytes: Vec<i8>, mut lineBytes: Vec<i8>) -> Vec<i8> {
        let mut j = 0;
        loop {
            if (j >= len(&lineBytes.clone())) {
                break;
            }
            allBytes = append(&allBytes.clone(), lineBytes[j as usize]);
            j = (j + 1);
        }
        return allBytes;
    }

    pub fn TokenizeBytes(mut bytes: Vec<i8>) -> Vec<Token> {
        let mut tokens: Vec<Token> = Vec::new();
        let mut i = 0;
        loop {
            if (i >= len(&bytes.clone())) {
                break;
            }
            let mut b = bytes[i as usize];
            if (IsWhitespace(b)) {
                i = (i + 1);
                continue;
            }
            if (b == 10) {
                tokens = append(
                    &tokens.clone(),
                    Token {
                        Type: TokenTypeNewline as i8,
                        Representation: vec![b].clone(),
                        ..Default::default()
                    }
                    .clone(),
                );
                i = (i + 1);
                continue;
            }
            if (b == 59) {
                loop {
                    if (i >= len(&bytes.clone())) {
                        break;
                    }
                    if (bytes[i as usize] == 10) {
                        break;
                    }
                    i = (i + 1);
                }
                continue;
            }
            if (b == 35) {
                tokens = append(
                    &tokens.clone(),
                    Token {
                        Type: TokenTypeHash as i8,
                        Representation: vec![b].clone(),
                        ..Default::default()
                    }
                    .clone(),
                );
                i = (i + 1);
                continue;
            }
            if (b == 36) {
                tokens = append(
                    &tokens.clone(),
                    Token {
                        Type: TokenTypeDollar as i8,
                        Representation: vec![b].clone(),
                        ..Default::default()
                    }
                    .clone(),
                );
                i = (i + 1);
                continue;
            }
            if (b == 58) {
                tokens = append(
                    &tokens.clone(),
                    Token {
                        Type: TokenTypeColon as i8,
                        Representation: vec![b].clone(),
                        ..Default::default()
                    }
                    .clone(),
                );
                i = (i + 1);
                continue;
            }
            if (b == 44) {
                tokens = append(
                    &tokens.clone(),
                    Token {
                        Type: TokenTypeComma as i8,
                        Representation: vec![b].clone(),
                        ..Default::default()
                    }
                    .clone(),
                );
                i = (i + 1);
                continue;
            }
            if (IsHexDigit(b)) {
                let mut repr: Vec<i8> = Vec::new();
                loop {
                    if (i >= len(&bytes.clone())) {
                        break;
                    }
                    if (!IsHexDigit(bytes[i as usize])) {
                        break;
                    }
                    repr = append(&repr.clone(), bytes[i as usize]);
                    i = (i + 1);
                }
                tokens = append(
                    &tokens.clone(),
                    Token {
                        Type: TokenTypeNumber as i8,
                        Representation: repr.clone(),
                        ..Default::default()
                    }
                    .clone(),
                );
                continue;
            }
            if (IsAlpha(b)) {
                let mut repr: Vec<i8> = Vec::new();
                loop {
                    if (i >= len(&bytes.clone())) {
                        break;
                    }
                    if ((!IsAlpha(bytes[i as usize])) && (!IsDigit(bytes[i as usize]))) {
                        break;
                    }
                    repr = append(&repr.clone(), bytes[i as usize]);
                    i = (i + 1);
                }
                tokens = append(
                    &tokens.clone(),
                    Token {
                        Type: TokenTypeIdentifier as i8,
                        Representation: repr.clone(),
                        ..Default::default()
                    }
                    .clone(),
                );
                continue;
            }
            i = (i + 1);
        }
        return tokens;
    }

    pub fn AssembleLines(mut lines: Vec<String>) -> Vec<u8> {
        let mut allBytes: Vec<i8> = Vec::new();
        let mut i = 0;
        loop {
            if (i >= len(&lines.clone())) {
                break;
            }
            let mut lineBytes = StringToBytes(lines[i as usize].clone().clone());
            allBytes = AppendLineBytes(allBytes.clone(), lineBytes.clone());
            if (i < (len(&lines.clone()) - 1)) {
                allBytes = append(&allBytes.clone(), (10 as i8));
            }
            i = (i + 1);
        }
        let mut tokens = TokenizeBytes(allBytes.clone());
        let mut instructions = Parse(tokens.clone());
        return Assemble(instructions.clone());
    }
} // pub mod assembler

pub mod font {
    use crate::*;

    pub fn GetFontData() -> Vec<u8> {
        return vec![
            0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x18, 0x18, 0x18, 0x18, 0x18, 0x00,
            0x18, 0x00, 0x6C, 0x6C, 0x24, 0x00, 0x00, 0x00, 0x00, 0x00, 0x6C, 0x6C, 0xFE, 0x6C,
            0xFE, 0x6C, 0x6C, 0x00, 0x18, 0x3E, 0x60, 0x3C, 0x06, 0x7C, 0x18, 0x00, 0x00, 0xC6,
            0xCC, 0x18, 0x30, 0x66, 0xC6, 0x00, 0x38, 0x6C, 0x38, 0x76, 0xDC, 0xCC, 0x76, 0x00,
            0x18, 0x18, 0x30, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0C, 0x18, 0x30, 0x30, 0x30, 0x18,
            0x0C, 0x00, 0x30, 0x18, 0x0C, 0x0C, 0x0C, 0x18, 0x30, 0x00, 0x00, 0x66, 0x3C, 0xFF,
            0x3C, 0x66, 0x00, 0x00, 0x00, 0x18, 0x18, 0x7E, 0x18, 0x18, 0x00, 0x00, 0x00, 0x00,
            0x00, 0x00, 0x00, 0x18, 0x18, 0x30, 0x00, 0x00, 0x00, 0x7E, 0x00, 0x00, 0x00, 0x00,
            0x00, 0x00, 0x00, 0x00, 0x00, 0x18, 0x18, 0x00, 0x06, 0x0C, 0x18, 0x30, 0x60, 0xC0,
            0x80, 0x00, 0x7C, 0xC6, 0xCE, 0xD6, 0xE6, 0xC6, 0x7C, 0x00, 0x18, 0x38, 0x18, 0x18,
            0x18, 0x18, 0x7E, 0x00, 0x7C, 0xC6, 0x06, 0x1C, 0x30, 0x66, 0xFE, 0x00, 0x7C, 0xC6,
            0x06, 0x3C, 0x06, 0xC6, 0x7C, 0x00, 0x1C, 0x3C, 0x6C, 0xCC, 0xFE, 0x0C, 0x1E, 0x00,
            0xFE, 0xC0, 0xC0, 0xFC, 0x06, 0xC6, 0x7C, 0x00, 0x38, 0x60, 0xC0, 0xFC, 0xC6, 0xC6,
            0x7C, 0x00, 0xFE, 0xC6, 0x0C, 0x18, 0x30, 0x30, 0x30, 0x00, 0x7C, 0xC6, 0xC6, 0x7C,
            0xC6, 0xC6, 0x7C, 0x00, 0x7C, 0xC6, 0xC6, 0x7E, 0x06, 0x0C, 0x78, 0x00, 0x00, 0x18,
            0x18, 0x00, 0x00, 0x18, 0x18, 0x00, 0x00, 0x18, 0x18, 0x00, 0x00, 0x18, 0x18, 0x30,
            0x06, 0x0C, 0x18, 0x30, 0x18, 0x0C, 0x06, 0x00, 0x00, 0x00, 0x7E, 0x00, 0x00, 0x7E,
            0x00, 0x00, 0x60, 0x30, 0x18, 0x0C, 0x18, 0x30, 0x60, 0x00, 0x7C, 0xC6, 0x0C, 0x18,
            0x18, 0x00, 0x18, 0x00, 0x7C, 0xC6, 0xDE, 0xDE, 0xDE, 0xC0, 0x78, 0x00, 0x38, 0x6C,
            0xC6, 0xFE, 0xC6, 0xC6, 0xC6, 0x00, 0xFC, 0x66, 0x66, 0x7C, 0x66, 0x66, 0xFC, 0x00,
            0x3C, 0x66, 0xC0, 0xC0, 0xC0, 0x66, 0x3C, 0x00, 0xF8, 0x6C, 0x66, 0x66, 0x66, 0x6C,
            0xF8, 0x00, 0xFE, 0x62, 0x68, 0x78, 0x68, 0x62, 0xFE, 0x00, 0xFE, 0x62, 0x68, 0x78,
            0x68, 0x60, 0xF0, 0x00, 0x3C, 0x66, 0xC0, 0xC0, 0xCE, 0x66, 0x3A, 0x00, 0xC6, 0xC6,
            0xC6, 0xFE, 0xC6, 0xC6, 0xC6, 0x00, 0x3C, 0x18, 0x18, 0x18, 0x18, 0x18, 0x3C, 0x00,
            0x1E, 0x0C, 0x0C, 0x0C, 0xCC, 0xCC, 0x78, 0x00, 0xE6, 0x66, 0x6C, 0x78, 0x6C, 0x66,
            0xE6, 0x00, 0xF0, 0x60, 0x60, 0x60, 0x62, 0x66, 0xFE, 0x00, 0xC6, 0xEE, 0xFE, 0xFE,
            0xD6, 0xC6, 0xC6, 0x00, 0xC6, 0xE6, 0xF6, 0xDE, 0xCE, 0xC6, 0xC6, 0x00, 0x7C, 0xC6,
            0xC6, 0xC6, 0xC6, 0xC6, 0x7C, 0x00, 0xFC, 0x66, 0x66, 0x7C, 0x60, 0x60, 0xF0, 0x00,
            0x7C, 0xC6, 0xC6, 0xC6, 0xD6, 0xDE, 0x7C, 0x06, 0xFC, 0x66, 0x66, 0x7C, 0x6C, 0x66,
            0xE6, 0x00, 0x7C, 0xC6, 0x60, 0x38, 0x0C, 0xC6, 0x7C, 0x00, 0x7E, 0x7E, 0x5A, 0x18,
            0x18, 0x18, 0x3C, 0x00, 0xC6, 0xC6, 0xC6, 0xC6, 0xC6, 0xC6, 0x7C, 0x00, 0xC6, 0xC6,
            0xC6, 0xC6, 0xC6, 0x6C, 0x38, 0x00, 0xC6, 0xC6, 0xC6, 0xD6, 0xD6, 0xFE, 0x6C, 0x00,
            0xC6, 0xC6, 0x6C, 0x38, 0x6C, 0xC6, 0xC6, 0x00, 0x66, 0x66, 0x66, 0x3C, 0x18, 0x18,
            0x3C, 0x00, 0xFE, 0xC6, 0x8C, 0x18, 0x32, 0x66, 0xFE, 0x00, 0x3C, 0x30, 0x30, 0x30,
            0x30, 0x30, 0x3C, 0x00, 0xC0, 0x60, 0x30, 0x18, 0x0C, 0x06, 0x02, 0x00, 0x3C, 0x0C,
            0x0C, 0x0C, 0x0C, 0x0C, 0x3C, 0x00, 0x10, 0x38, 0x6C, 0xC6, 0x00, 0x00, 0x00, 0x00,
            0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0x30, 0x18, 0x0C, 0x00, 0x00, 0x00,
            0x00, 0x00, 0x00, 0x00, 0x78, 0x0C, 0x7C, 0xCC, 0x76, 0x00, 0xE0, 0x60, 0x7C, 0x66,
            0x66, 0x66, 0xDC, 0x00, 0x00, 0x00, 0x7C, 0xC6, 0xC0, 0xC6, 0x7C, 0x00, 0x1C, 0x0C,
            0x7C, 0xCC, 0xCC, 0xCC, 0x76, 0x00, 0x00, 0x00, 0x7C, 0xC6, 0xFE, 0xC0, 0x7C, 0x00,
            0x3C, 0x66, 0x60, 0xF8, 0x60, 0x60, 0xF0, 0x00, 0x00, 0x00, 0x76, 0xCC, 0xCC, 0x7C,
            0x0C, 0xF8, 0xE0, 0x60, 0x6C, 0x76, 0x66, 0x66, 0xE6, 0x00, 0x18, 0x00, 0x38, 0x18,
            0x18, 0x18, 0x3C, 0x00, 0x06, 0x00, 0x06, 0x06, 0x06, 0x66, 0x66, 0x3C, 0xE0, 0x60,
            0x66, 0x6C, 0x78, 0x6C, 0xE6, 0x00, 0x38, 0x18, 0x18, 0x18, 0x18, 0x18, 0x3C, 0x00,
            0x00, 0x00, 0xEC, 0xFE, 0xD6, 0xD6, 0xD6, 0x00, 0x00, 0x00, 0xDC, 0x66, 0x66, 0x66,
            0x66, 0x00, 0x00, 0x00, 0x7C, 0xC6, 0xC6, 0xC6, 0x7C, 0x00, 0x00, 0x00, 0xDC, 0x66,
            0x66, 0x7C, 0x60, 0xF0, 0x00, 0x00, 0x76, 0xCC, 0xCC, 0x7C, 0x0C, 0x1E, 0x00, 0x00,
            0xDC, 0x76, 0x60, 0x60, 0xF0, 0x00, 0x00, 0x00, 0x7E, 0xC0, 0x7C, 0x06, 0xFC, 0x00,
            0x30, 0x30, 0xFC, 0x30, 0x30, 0x36, 0x1C, 0x00, 0x00, 0x00, 0xCC, 0xCC, 0xCC, 0xCC,
            0x76, 0x00, 0x00, 0x00, 0xC6, 0xC6, 0xC6, 0x6C, 0x38, 0x00, 0x00, 0x00, 0xC6, 0xD6,
            0xD6, 0xFE, 0x6C, 0x00, 0x00, 0x00, 0xC6, 0x6C, 0x38, 0x6C, 0xC6, 0x00, 0x00, 0x00,
            0xC6, 0xC6, 0xC6, 0x7E, 0x06, 0xFC, 0x00, 0x00, 0xFE, 0x8C, 0x18, 0x32, 0xFE, 0x00,
            0x0E, 0x18, 0x18, 0x70, 0x18, 0x18, 0x0E, 0x00, 0x18, 0x18, 0x18, 0x18, 0x18, 0x18,
            0x18, 0x00, 0x70, 0x18, 0x18, 0x0E, 0x18, 0x18, 0x70, 0x00, 0x76, 0xDC, 0x00, 0x00,
            0x00, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
        ];
    }

    pub fn GetCharBitmap(mut fontData: Vec<u8>, charCode: i32) -> Vec<u8> {
        let mut result: Vec<u8> = Vec::new();
        let mut code = charCode;
        if (code < 32) {
            code = 32;
        }
        if (code > 127) {
            code = 127;
        }
        let mut offset = ((code - 32) * 8);
        let mut i = 0;
        loop {
            if (i >= 8) {
                break;
            }
            result = append(&result.clone(), fontData[(offset + i) as usize]);
            i = (i + 1);
        }
        return result;
    }

    pub fn GetPixel(mut fontData: Vec<u8>, charCode: i32, x: i32, y: i32) -> bool {
        let mut code = charCode;
        if (code < 32) {
            code = 32;
        }
        if (code > 127) {
            code = 127;
        }
        let mut offset = ((code - 32) * 8);
        let mut row = fontData[(offset + y) as usize];
        let mut mask = ((0x80 >> x) as u8);
        return ((row & mask) != 0);
    }
} // pub mod font

pub const TextCols: i32 = 40;

pub const TextRows: i32 = 25;

pub const TextScreenBase: i32 = 0x0400;

pub fn hexDigit(n: i32) -> String {
    if (n == 0) {
        return "0".to_string();
    } else if (n == 1) {
        return "1".to_string();
    } else if (n == 2) {
        return "2".to_string();
    } else if (n == 3) {
        return "3".to_string();
    } else if (n == 4) {
        return "4".to_string();
    } else if (n == 5) {
        return "5".to_string();
    } else if (n == 6) {
        return "6".to_string();
    } else if (n == 7) {
        return "7".to_string();
    } else if (n == 8) {
        return "8".to_string();
    } else if (n == 9) {
        return "9".to_string();
    } else if (n == 10) {
        return "A".to_string();
    } else if (n == 11) {
        return "B".to_string();
    } else if (n == 12) {
        return "C".to_string();
    } else if (n == 13) {
        return "D".to_string();
    } else if (n == 14) {
        return "E".to_string();
    } else if (n == 15) {
        return "F".to_string();
    }
    return "0".to_string();
}

pub fn toHex2(n: i32) -> String {
    let mut high = ((n >> 4) & 0x0F);
    let mut low = (n & 0x0F);
    return (hexDigit(high) + &hexDigit(low));
}

pub fn toHex4(n: i32) -> String {
    return (toHex2(((n >> 8) & 0xFF)) + &toHex2((n & 0xFF)));
}

pub fn toHex(n: i32) -> String {
    if (n > 255) {
        return ("$".to_string() + &toHex4(n));
    }
    return ("$".to_string() + &toHex2(n));
}

pub fn addStringToScreen(mut lines: Vec<String>, text: String, row: i32, col: i32) -> Vec<String> {
    let mut baseAddr = ((TextScreenBase + (row * TextCols)) + col);
    let mut i = 0;
    loop {
        if (i >= text.clone().len() as i32) {
            break;
        }
        let mut charCode = (text.as_bytes()[i as usize] as i32);
        let mut addr = (baseAddr + i);
        lines = append(
            &lines.clone(),
            ("LDA #".to_string() + &toHex(charCode)).clone(),
        );
        lines = append(&lines.clone(), ("STA ".to_string() + &toHex(addr)).clone());
        i = (i + 1);
    }
    return lines;
}

pub fn clearScreen(mut lines: Vec<String>) -> Vec<String> {
    let mut addr = TextScreenBase;
    let mut i = 0;
    loop {
        if (i >= (TextCols * TextRows)) {
            break;
        }
        lines = append(&lines.clone(), "LDA #$20".to_string());
        lines = append(
            &lines.clone(),
            ("STA ".to_string() + &toHex((addr + i))).clone(),
        );
        i = (i + 1);
    }
    return lines;
}

pub fn createC64WelcomeScreen() -> Vec<u8> {
    let mut lines: Vec<String> = Vec::new();
    lines = clearScreen(lines.clone());
    lines = addStringToScreen(
        lines.clone(),
        "**** COMMODORE 64 BASIC V2 ****".to_string(),
        1,
        4,
    );
    lines = addStringToScreen(
        lines.clone(),
        "64K RAM SYSTEM  38911 BASIC BYTES FREE".to_string(),
        3,
        1,
    );
    lines = addStringToScreen(lines.clone(), "READY.".to_string(), 5, 0);
    lines = append(&lines.clone(), "LDA #$5F".to_string());
    lines = append(&lines.clone(), "STA $04F0".to_string());
    lines = append(&lines.clone(), "BRK".to_string());
    return assembler::AssembleLines(lines.clone());
}

pub fn main() {
    let mut scale = (2 as i32);
    let mut windowWidth = (((TextCols * 8) as i32) * scale);
    let mut windowHeight = (((TextRows * 8) as i32) * scale);
    let mut w = graphics::CreateWindow("Commodore 64".to_string(), windowWidth, windowHeight);
    let mut c = cpu::NewCPU();
    let mut fontData = font::GetFontData();
    let mut program = createC64WelcomeScreen();
    c = cpu::LoadProgram(c.clone(), program.clone(), 0x0600);
    c = cpu::SetPC(c.clone(), 0x0600);
    c = cpu::Run(c.clone(), 100000);
    let mut textColor = graphics::NewColor(134, 122, 222, 255);
    let mut bgColor = graphics::NewColor(64, 50, 133, 255);
    loop {
        let mut running: bool = false;
        (w, running) = graphics::PollEvents(w.clone());
        if (!running) {
            break;
        }
        graphics::Clear(w.clone(), bgColor.clone());
        let mut charY = 0;
        loop {
            if (charY >= TextRows) {
                break;
            }
            let mut charX = 0;
            loop {
                if (charX >= TextCols) {
                    break;
                }
                let mut memAddr = ((TextScreenBase + (charY * TextCols)) + charX);
                let mut charCode = (cpu::GetMemory(c.clone(), memAddr) as i32);
                if (charCode >= 32) {
                    if (charCode <= 127) {
                        let mut pixelY = 0;
                        loop {
                            if (pixelY >= 8) {
                                break;
                            }
                            let mut pixelX = 0;
                            loop {
                                if (pixelX >= 8) {
                                    break;
                                }
                                if (font::GetPixel(fontData.clone(), charCode, pixelX, pixelY)) {
                                    let mut screenX = ((((charX * 8) + pixelX) as i32) * scale);
                                    let mut screenY = ((((charY * 8) + pixelY) as i32) * scale);
                                    graphics::FillRect(
                                        w.clone(),
                                        graphics::NewRect(screenX, screenY, scale, scale).clone(),
                                        textColor.clone(),
                                    );
                                }
                                pixelX = (pixelX + 1);
                            }
                            pixelY = (pixelY + 1);
                        }
                    }
                }
                charX = (charX + 1);
            }
            charY = (charY + 1);
        }
        graphics::Present(w.clone());
    }
    graphics::CloseWindow(w.clone());
}

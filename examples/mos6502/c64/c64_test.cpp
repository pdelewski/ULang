#include <vector>
#include <string>
#include <tuple>
#include <any>
#include <cstdint>
#include <functional>
#include <cstdarg> // For va_start, etc.
#include <initializer_list>
#include <iostream>

using int8 = int8_t;
using int16 = int16_t;
using int32 = int32_t;
using int64 = int64_t;
using uint8 = uint8_t;
using uint16 = uint16_t;
using uint32 = uint32_t;
using uint64 = uint64_t;

std::string string_format(const std::string fmt, ...)
{
    int size =
        ((int)fmt.size()) * 2 + 50; // Use a rubric appropriate for your code
    std::string str;
    va_list ap;
    while (1) { // Maximum two passes on a POSIX system...
        str.resize(size);
        va_start(ap, fmt);
        int n = vsnprintf((char *)str.data(), size, fmt.c_str(), ap);
        va_end(ap);
        if (n > -1 && n < size) { // Everything worked
            str.resize(n);
            return str;
        }
        if (n > -1)     // Needed size returned
            size = n + 1; // For null char
        else
            size *= 2; // Guess at a larger size (OS specific)
    }
    return str;
}

void println()
{
    printf("\n");
}

void println(std::int8_t val)
{
    printf("%d\n", val);
}

template<typename T>
void println(const T& val)
{
    std::cout << val << std::endl;
}

void printf(const signed char c)
{
    std::cout << (int)c;
}

template<typename T>
void printf(const T& val)
{
    std::cout << val;
}

// Function to mimic Go's append behavior for std::vector
template <typename T>
std::vector<T> append(const std::vector<T> &vec,
                      const std::initializer_list<T> &elements)
{
    std::vector<T> result = vec;           // Create a copy of the original vector
    result.insert(result.end(), elements); // Append the elements
    return result;                         // Return the new vector
}

// Overload to allow appending another vector
template <typename T>
std::vector<T> append(const std::vector<T> &vec,
                      const std::vector<T> &elements)
{
    std::vector<T> result = vec; // Create a copy of the original vector
    result.insert(result.end(), elements.begin(),
                  elements.end()); // Append the elements
    return result;                 // Return the new vector
}

template <typename T>
std::vector<T> append(const std::vector<T> &vec, const T &element)
{
    std::vector<T> result = vec; // Create a copy of the original vector
    result.push_back(element);   // Append the single element
    return result;               // Return the new vector
}

// Specialization for appending const char* to vector of strings
std::vector<std::string> append(const std::vector<std::string> &vec, const char *element)
{
    std::vector<std::string> result = vec;
    result.push_back(std::string(element));
    return result;
}


namespace cpu {

struct CPU {
    std::uint8_t A;
    std::uint8_t X;
    std::uint8_t Y;
    std::uint8_t SP;
    int PC;
    std::uint8_t Status;
    std::vector<std::uint8_t> Memory;
    bool Halted;
    int Cycles;
};

constexpr auto OpLDAImm = 0xA9;
constexpr auto OpLDAZp = 0xA5;
constexpr auto OpLDAZpX = 0xB5;
constexpr auto OpLDAAbs = 0xAD;
constexpr auto OpLDXImm = 0xA2;
constexpr auto OpLDXAbs = 0xAE;
constexpr auto OpLDYImm = 0xA0;
constexpr auto OpLDYAbs = 0xAC;
constexpr auto OpSTAZp = 0x85;
constexpr auto OpSTAZpX = 0x95;
constexpr auto OpSTAAbs = 0x8D;
constexpr auto OpSTXAbs = 0x8E;
constexpr auto OpSTYAbs = 0x8C;
constexpr auto OpADCImm = 0x69;
constexpr auto OpSBCImm = 0xE9;
constexpr auto OpINX = 0xE8;
constexpr auto OpINY = 0xC8;
constexpr auto OpDEX = 0xCA;
constexpr auto OpDEY = 0x88;
constexpr auto OpINC = 0xE6;
constexpr auto OpCMPImm = 0xC9;
constexpr auto OpCPXImm = 0xE0;
constexpr auto OpCPYImm = 0xC0;
constexpr auto OpBNE = 0xD0;
constexpr auto OpBEQ = 0xF0;
constexpr auto OpBCC = 0x90;
constexpr auto OpBCS = 0xB0;
constexpr auto OpJMP = 0x4C;
constexpr auto OpJSR = 0x20;
constexpr auto OpRTS = 0x60;
constexpr auto OpNOP = 0xEA;
constexpr auto OpBRK = 0x00;

constexpr auto FlagC = 0x01;
constexpr auto FlagZ = 0x02;
constexpr auto FlagI = 0x04;
constexpr auto FlagD = 0x08;
constexpr auto FlagB = 0x10;
constexpr auto FlagV = 0x40;
constexpr auto FlagN = 0x80;

constexpr auto ScreenBase = 0x0200;
constexpr auto ScreenWidth = 32;
constexpr auto ScreenHeight = 32;
constexpr auto ScreenSize = 1024;

// Forward declarations
CPU NewCPU();
CPU LoadProgram(CPU c, std::vector<std::uint8_t> program, int addr);
CPU SetPC(CPU c, int addr);
std::uint8_t ReadByte(CPU c, int addr);
CPU WriteByte(CPU c, int addr, std::uint8_t value);
std::tuple<CPU,std::uint8_t> FetchByte(CPU c);
std::tuple<CPU,int> FetchWord(CPU c);
CPU SetZN(CPU c, std::uint8_t value);
CPU SetCarry(CPU c, bool set);
bool GetCarry(CPU c);
bool GetZero(CPU c);
CPU Step(CPU c);
CPU Run(CPU c, int maxCycles);
std::uint8_t GetScreenPixel(CPU c, int x, int y);
bool IsHalted(CPU c);
std::uint8_t GetMemory(CPU c, int addr);

CPU NewCPU()
{
    auto mem = std::vector<std::uint8_t> {};
    auto i = 0;
    for (;;) {
        if (i >= 65536) {
            break;
        }
        mem = append(mem, std::uint8_t(0));
        i = i + 1;
    }
    return CPU{.A= 0
                   , .X= 0
               , .Y= 0
               , .SP= 0xFF
               , .PC= 0x0600
               , .Status= 0x20
               , .Memory= mem
               , .Halted= false
               , .Cycles= 0
              };
}

CPU LoadProgram(CPU c, std::vector<std::uint8_t> program, int addr)
{
    auto i = 0;
    for (;;) {
        if (i >= std::size(program)) {
            break;
        }
        c.Memory[addr + i] = program[i];
        i = i + 1;
    }
    return c;
}

CPU SetPC(CPU c, int addr)
{
    c.PC = addr;
    return c;
}

std::uint8_t ReadByte(CPU c, int addr)
{
    return c.Memory[addr];
}

CPU WriteByte(CPU c, int addr, std::uint8_t value)
{
    c.Memory[addr] = value;
    return c;
}

std::tuple<CPU,std::uint8_t> FetchByte(CPU c)
{
    auto value = c.Memory[c.PC];
    c.PC = c.PC + 1;
    return std::make_tuple(c, value);
}

std::tuple<CPU,int> FetchWord(CPU c)
{
    auto low = int(c.Memory[c.PC]);
    auto high = int(c.Memory[c.PC + 1]);
    c.PC = c.PC + 2;
    return std::make_tuple(c, low + (high * 256));
}

CPU SetZN(CPU c, std::uint8_t value)
{
    if (value == 0) {
        c.Status = c.Status | FlagZ;
    } else  {
        c.Status = c.Status & (0xFF - FlagZ);
    }
    if ((value & 0x80) != 0) {
        c.Status = c.Status | FlagN;
    } else  {
        c.Status = c.Status & (0xFF - FlagN);
    }
    return c;
}

CPU SetCarry(CPU c, bool set)
{
    if (set) {
        c.Status = c.Status | FlagC;
    } else  {
        c.Status = c.Status & (0xFF - FlagC);
    }
    return c;
}

bool GetCarry(CPU c)
{
    return (c.Status & FlagC) != 0;
}

bool GetZero(CPU c)
{
    return (c.Status & FlagZ) != 0;
}

CPU Step(CPU c)
{
    if (c.Halted) {
        return c;
    }
    std::uint8_t opcode;
    std::tie(c, opcode) = FetchByte(c);
    c.Cycles = c.Cycles + 1;
    if (opcode == OpLDAImm) {
        std::uint8_t value;
        std::tie(c, value) = FetchByte(c);
        c.A = value;
        c = SetZN(c, c.A);
    } else  if (opcode == OpLDAZp) {
        std::uint8_t addr;
        std::tie(c, addr) = FetchByte(c);
        c.A = c.Memory[int(addr)];
        c = SetZN(c, c.A);
    } else  if (opcode == OpLDAZpX) {
        std::uint8_t addr;
        std::tie(c, addr) = FetchByte(c);
        c.A = c.Memory[int(addr + c.X)];
        c = SetZN(c, c.A);
    } else  if (opcode == OpLDAAbs) {
        int addr;
        std::tie(c, addr) = FetchWord(c);
        c.A = c.Memory[addr];
        c = SetZN(c, c.A);
    } else  if (opcode == OpLDXImm) {
        std::uint8_t value;
        std::tie(c, value) = FetchByte(c);
        c.X = value;
        c = SetZN(c, c.X);
    } else  if (opcode == OpLDXAbs) {
        int addr;
        std::tie(c, addr) = FetchWord(c);
        c.X = c.Memory[addr];
        c = SetZN(c, c.X);
    } else  if (opcode == OpLDYImm) {
        std::uint8_t value;
        std::tie(c, value) = FetchByte(c);
        c.Y = value;
        c = SetZN(c, c.Y);
    } else  if (opcode == OpLDYAbs) {
        int addr;
        std::tie(c, addr) = FetchWord(c);
        c.Y = c.Memory[addr];
        c = SetZN(c, c.Y);
    } else  if (opcode == OpSTAZp) {
        std::uint8_t addr;
        std::tie(c, addr) = FetchByte(c);
        c.Memory[int(addr)] = c.A;
    } else  if (opcode == OpSTAZpX) {
        std::uint8_t addr;
        std::tie(c, addr) = FetchByte(c);
        c.Memory[int(addr + c.X)] = c.A;
    } else  if (opcode == OpSTAAbs) {
        int addr;
        std::tie(c, addr) = FetchWord(c);
        c.Memory[addr] = c.A;
    } else  if (opcode == OpSTXAbs) {
        int addr;
        std::tie(c, addr) = FetchWord(c);
        c.Memory[addr] = c.X;
    } else  if (opcode == OpSTYAbs) {
        int addr;
        std::tie(c, addr) = FetchWord(c);
        c.Memory[addr] = c.Y;
    } else  if (opcode == OpADCImm) {
        std::uint8_t value;
        std::tie(c, value) = FetchByte(c);
        auto carry = 0;
        if (GetCarry(c)) {
            carry = 1;
        }
        auto result = int(c.A) + int(value) + carry;
        c = SetCarry(c, result > 255);
        c.A = std::uint8_t(result & 0xFF);
        c = SetZN(c, c.A);
    } else  if (opcode == OpSBCImm) {
        std::uint8_t value;
        std::tie(c, value) = FetchByte(c);
        auto carry = 0;
        if (GetCarry(c)) {
            carry = 1;
        }
        auto result = int(c.A) - int(value) - (1 - carry);
        c = SetCarry(c, result >= 0);
        c.A = std::uint8_t(result & 0xFF);
        c = SetZN(c, c.A);
    } else  if (opcode == OpINX) {
        c.X = c.X + 1;
        c = SetZN(c, c.X);
    } else  if (opcode == OpINY) {
        c.Y = c.Y + 1;
        c = SetZN(c, c.Y);
    } else  if (opcode == OpDEX) {
        c.X = c.X - 1;
        c = SetZN(c, c.X);
    } else  if (opcode == OpDEY) {
        c.Y = c.Y - 1;
        c = SetZN(c, c.Y);
    } else  if (opcode == OpINC) {
        std::uint8_t addr;
        std::tie(c, addr) = FetchByte(c);
        auto val = c.Memory[int(addr)] + 1;
        c.Memory[int(addr)] = val;
        c = SetZN(c, val);
    } else  if (opcode == OpCMPImm) {
        std::uint8_t value;
        std::tie(c, value) = FetchByte(c);
        auto result = int(c.A) - int(value);
        c = SetCarry(c, c.A >= value);
        c = SetZN(c, std::uint8_t(result & 0xFF));
    } else  if (opcode == OpCPXImm) {
        std::uint8_t value;
        std::tie(c, value) = FetchByte(c);
        auto result = int(c.X) - int(value);
        c = SetCarry(c, c.X >= value);
        c = SetZN(c, std::uint8_t(result & 0xFF));
    } else  if (opcode == OpCPYImm) {
        std::uint8_t value;
        std::tie(c, value) = FetchByte(c);
        auto result = int(c.Y) - int(value);
        c = SetCarry(c, c.Y >= value);
        c = SetZN(c, std::uint8_t(result & 0xFF));
    } else  if (opcode == OpBNE) {
        std::uint8_t offset;
        std::tie(c, offset) = FetchByte(c);
        if ((!GetZero(c))) {
            if (offset < 128) {
                c.PC = c.PC + int(offset);
            } else      {
                c.PC = c.PC - (256 - int(offset));
            }
        }
    } else  if (opcode == OpBEQ) {
        std::uint8_t offset;
        std::tie(c, offset) = FetchByte(c);
        if (GetZero(c)) {
            if (offset < 128) {
                c.PC = c.PC + int(offset);
            } else      {
                c.PC = c.PC - (256 - int(offset));
            }
        }
    } else  if (opcode == OpBCC) {
        std::uint8_t offset;
        std::tie(c, offset) = FetchByte(c);
        if ((!GetCarry(c))) {
            if (offset < 128) {
                c.PC = c.PC + int(offset);
            } else      {
                c.PC = c.PC - (256 - int(offset));
            }
        }
    } else  if (opcode == OpBCS) {
        std::uint8_t offset;
        std::tie(c, offset) = FetchByte(c);
        if (GetCarry(c)) {
            if (offset < 128) {
                c.PC = c.PC + int(offset);
            } else      {
                c.PC = c.PC - (256 - int(offset));
            }
        }
    } else  if (opcode == OpJMP) {
        int addr;
        std::tie(c, addr) = FetchWord(c);
        c.PC = addr;
    } else  if (opcode == OpJSR) {
        int addr;
        std::tie(c, addr) = FetchWord(c);
        auto retAddr = c.PC - 1;
        c.Memory[0x100 + int(c.SP)] = std::uint8_t((retAddr >> 8) & 0xFF);
        c.SP = c.SP - 1;
        c.Memory[0x100 + int(c.SP)] = std::uint8_t(retAddr & 0xFF);
        c.SP = c.SP - 1;
        c.PC = addr;
    } else  if (opcode == OpRTS) {
        c.SP = c.SP + 1;
        auto low = int(c.Memory[0x100 + int(c.SP)]);
        c.SP = c.SP + 1;
        auto high = int(c.Memory[0x100 + int(c.SP)]);
        c.PC = (high * 256) + low + 1;
    } else  if (opcode == OpNOP) {
    } else  if (opcode == OpBRK) {
        c.Halted = true;
    }
    return c;
}

CPU Run(CPU c, int maxCycles)
{
    for (;;) {
        if (c.Halted) {
            break;
        }
        if (c.Cycles >= maxCycles) {
            break;
        }
        c = Step(c);
    }
    return c;
}

std::uint8_t GetScreenPixel(CPU c, int x, int y)
{
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
    auto addr = ScreenBase + (y * ScreenWidth) + x;
    return c.Memory[addr];
}

bool IsHalted(CPU c)
{
    return c.Halted;
}

std::uint8_t GetMemory(CPU c, int addr)
{
    if (addr < 0) {
        return 0;
    }
    if (addr >= 65536) {
        return 0;
    }
    return c.Memory[addr];
}

} // namespace cpu

namespace assembler {

struct Instruction {
    std::vector<std::int8_t> OpcodeBytes;
    std::int8_t Mode;
    int Operand;
    std::vector<std::int8_t> LabelBytes;
    bool HasLabel;
};

struct Token {
    std::int8_t Type;
    std::vector<std::int8_t> Representation;
};

constexpr auto TokenTypeInstruction = 1;
constexpr auto TokenTypeNumber = 2;
constexpr auto TokenTypeLabel = 3;
constexpr auto TokenTypeComma = 4;
constexpr auto TokenTypeNewline = 5;
constexpr auto TokenTypeHash = 6;
constexpr auto TokenTypeDollar = 7;
constexpr auto TokenTypeColon = 8;
constexpr auto TokenTypeIdentifier = 9;
constexpr auto TokenTypeComment = 10;

constexpr auto ModeImplied = 0;
constexpr auto ModeImmediate = 1;
constexpr auto ModeZeroPage = 2;
constexpr auto ModeAbsolute = 3;
constexpr auto ModeZeroPageX = 4;

// Forward declarations
bool IsDigit(std::int8_t b);
bool IsHexDigit(std::int8_t b);
bool IsAlpha(std::int8_t b);
bool IsWhitespace(std::int8_t b);
std::vector<std::int8_t> StringToBytes(std::string s);
std::int8_t ToUpper(std::int8_t b);
std::vector<Token> Tokenize(std::string text);
int ParseHex(std::vector<std::int8_t> bytes);
int ParseDecimal(std::vector<std::int8_t> bytes);
bool MatchToken(Token token, std::string s);
std::vector<std::int8_t> CopyBytes(std::vector<std::int8_t> src);
std::vector<Instruction> Parse(std::vector<Token> tokens);
bool IsOpcode(std::vector<std::int8_t> opcodeBytes, std::string name);
std::vector<std::uint8_t> Assemble(std::vector<Instruction> instructions);
std::vector<std::uint8_t> AssembleString(std::string text);
std::vector<std::int8_t> AppendLineBytes(std::vector<std::int8_t> allBytes, std::vector<std::int8_t> lineBytes);
std::vector<Token> TokenizeBytes(std::vector<std::int8_t> bytes);
std::vector<std::uint8_t> AssembleLines(std::vector<std::string> lines);

bool IsDigit(std::int8_t b)
{
    return b >= '0' && b <= '9';
}

bool IsHexDigit(std::int8_t b)
{
    return IsDigit(b) || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F');
}

bool IsAlpha(std::int8_t b)
{
    return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_';
}

bool IsWhitespace(std::int8_t b)
{
    return b == ' ' || b == '\t';
}

std::vector<std::int8_t> StringToBytes(std::string s)
{
    auto result = std::vector<std::int8_t> {};
    auto i = 0;
    for (;;) {
        if (i >= std::size(s)) {
            break;
        }
        result = append(result, std::int8_t(s[i]));
        i = i + 1;
    }
    return result;
}

std::int8_t ToUpper(std::int8_t b)
{
    if (b >= 'a' && b <= 'z') {
        return b - 32;
    }
    return b;
}

std::vector<Token> Tokenize(std::string text)
{
    auto tokens = std::vector<Token> {};
    auto bytes = StringToBytes(text);
    auto i = 0;
    for (;;) {
        if (i >= std::size(bytes)) {
            break;
        }
        auto b = bytes[i];
        if (IsWhitespace(b)) {
            i = i + 1;
            continue;
        }
        if (b == '\n') {
            tokens = append(tokens, Token{.Type= TokenTypeNewline
                                                 , .Representation= std::vector<std::int8_t>{b}
                                         });
            i = i + 1;
            continue;
        }
        if (b == ';') {
            for (;;) {
                if (i >= std::size(bytes)) {
                    break;
                }
                if (bytes[i] == '\n') {
                    break;
                }
                i = i + 1;
            }
            continue;
        }
        if (b == '#') {
            tokens = append(tokens, Token{.Type= TokenTypeHash
                                                 , .Representation= std::vector<std::int8_t>{b}
                                         });
            i = i + 1;
            continue;
        }
        if (b == '$') {
            tokens = append(tokens, Token{.Type= TokenTypeDollar
                                                 , .Representation= std::vector<std::int8_t>{b}
                                         });
            i = i + 1;
            continue;
        }
        if (b == ':') {
            tokens = append(tokens, Token{.Type= TokenTypeColon
                                                 , .Representation= std::vector<std::int8_t>{b}
                                         });
            i = i + 1;
            continue;
        }
        if (b == ',') {
            tokens = append(tokens, Token{.Type= TokenTypeComma
                                                 , .Representation= std::vector<std::int8_t>{b}
                                         });
            i = i + 1;
            continue;
        }
        if (IsHexDigit(b)) {
            auto repr = std::vector<std::int8_t> {};
            for (;;) {
                if (i >= std::size(bytes)) {
                    break;
                }
                if ((!IsHexDigit(bytes[i]))) {
                    break;
                }
                repr = append(repr, bytes[i]);
                i = i + 1;
            }
            tokens = append(tokens, Token{.Type= TokenTypeNumber
                                                 , .Representation= repr
                                         });
            continue;
        }
        if (IsAlpha(b)) {
            auto repr = std::vector<std::int8_t> {};
            for (;;) {
                if (i >= std::size(bytes)) {
                    break;
                }
                if ((!IsAlpha(bytes[i])) && (!IsDigit(bytes[i]))) {
                    break;
                }
                repr = append(repr, bytes[i]);
                i = i + 1;
            }
            tokens = append(tokens, Token{.Type= TokenTypeIdentifier
                                                 , .Representation= repr
                                         });
            continue;
        }
        i = i + 1;
    }
    return tokens;
}

int ParseHex(std::vector<std::int8_t> bytes)
{
    auto result = 0;
    auto i = 0;
    for (;;) {
        if (i >= std::size(bytes)) {
            break;
        }
        auto b = bytes[i];
        result = result * 16;
        if (b >= '0' && b <= '9') {
            result = result + int(b - '0');
        } else    if (b >= 'a' && b <= 'f') {
            result = result + int(b - 'a' + 10);
        } else    if (b >= 'A' && b <= 'F') {
            result = result + int(b - 'A' + 10);
        }
        i = i + 1;
    }
    return result;
}

int ParseDecimal(std::vector<std::int8_t> bytes)
{
    auto result = 0;
    auto i = 0;
    for (;;) {
        if (i >= std::size(bytes)) {
            break;
        }
        auto b = bytes[i];
        result = result * 10 + int(b - '0');
        i = i + 1;
    }
    return result;
}

bool MatchToken(Token token, std::string s)
{
    if (std::size(token.Representation) != std::size(s)) {
        return false;
    }
    auto i = 0;
    for (;;) {
        if (i >= std::size(s)) {
            break;
        }
        if (ToUpper(token.Representation[i]) != ToUpper(std::int8_t(s[i]))) {
            return false;
        }
        i = i + 1;
    }
    return true;
}

std::vector<std::int8_t> CopyBytes(std::vector<std::int8_t> src)
{
    auto dst = std::vector<std::int8_t> {};
    auto i = 0;
    for (;;) {
        if (i >= std::size(src)) {
            break;
        }
        dst = append(dst, src[i]);
        i = i + 1;
    }
    return dst;
}

std::vector<Instruction> Parse(std::vector<Token> tokens)
{
    auto instructions = std::vector<Instruction> {};
    auto i = 0;
    for (;;) {
        if (i >= std::size(tokens)) {
            break;
        }
        if (tokens[i].Type == TokenTypeNewline) {
            i = i + 1;
            continue;
        }
        auto currentLabelBytes = std::vector<std::int8_t> {};
        auto hasLabel = false;
        if (tokens[i].Type == TokenTypeIdentifier && i + 1 < std::size(tokens) && tokens[i + 1].Type == TokenTypeColon) {
            currentLabelBytes = CopyBytes(tokens[i].Representation);
            hasLabel = true;
            i = i + 2;
            for (;;) {
                if (i >= std::size(tokens)) {
                    break;
                }
                if (tokens[i].Type != TokenTypeNewline) {
                    break;
                }
                i = i + 1;
            }
            if (i >= std::size(tokens)) {
                break;
            }
        }
        if (tokens[i].Type != TokenTypeIdentifier) {
            i = i + 1;
            continue;
        }
        auto instr = Instruction{.OpcodeBytes= CopyBytes(tokens[i].Representation)
                                               , .Mode= ModeImplied
                                 , .Operand= 0
                                 , .LabelBytes= currentLabelBytes
                                 , .HasLabel= hasLabel
                                };
        i = i + 1;
        if (i < std::size(tokens) && tokens[i].Type != TokenTypeNewline) {
            if (tokens[i].Type == TokenTypeHash) {
                i = i + 1;
                instr.Mode = ModeImmediate;
                if (i < std::size(tokens) && tokens[i].Type == TokenTypeDollar) {
                    i = i + 1;
                    if (i < std::size(tokens) && tokens[i].Type == TokenTypeNumber) {
                        instr.Operand = ParseHex(tokens[i].Representation);
                        i = i + 1;
                    }
                } else        if (i < std::size(tokens) && tokens[i].Type == TokenTypeNumber) {
                    instr.Operand = ParseDecimal(tokens[i].Representation);
                    i = i + 1;
                }
            } else      if (tokens[i].Type == TokenTypeDollar) {
                i = i + 1;
                if (i < std::size(tokens) && tokens[i].Type == TokenTypeNumber) {
                    instr.Operand = ParseHex(tokens[i].Representation);
                    if (std::size(tokens[i].Representation) <= 2) {
                        instr.Mode = ModeZeroPage;
                    } else          {
                        instr.Mode = ModeAbsolute;
                    }
                    i = i + 1;
                    if (i < std::size(tokens) && tokens[i].Type == TokenTypeComma) {
                        i = i + 1;
                        if (i < std::size(tokens) && tokens[i].Type == TokenTypeIdentifier) {
                            if (MatchToken(tokens[i], "X")) {
                                instr.Mode = ModeZeroPageX;
                            }
                            i = i + 1;
                        }
                    }
                }
            } else      if (tokens[i].Type == TokenTypeNumber) {
                instr.Operand = ParseDecimal(tokens[i].Representation);
                if (instr.Operand <= 255) {
                    instr.Mode = ModeZeroPage;
                } else        {
                    instr.Mode = ModeAbsolute;
                }
                i = i + 1;
            }
        }
        instructions = append(instructions, instr);
    }
    return instructions;
}

bool IsOpcode(std::vector<std::int8_t> opcodeBytes, std::string name)
{
    if (std::size(opcodeBytes) != std::size(name)) {
        return false;
    }
    auto i = 0;
    for (;;) {
        if (i >= std::size(name)) {
            break;
        }
        auto ob = opcodeBytes[i];
        if (ob >= 'a' && ob <= 'z') {
            ob = ob - 32;
        }
        auto nb = std::int8_t(name[i]);
        if (ob != nb) {
            return false;
        }
        i = i + 1;
    }
    return true;
}

std::vector<std::uint8_t> Assemble(std::vector<Instruction> instructions)
{
    auto code = std::vector<std::uint8_t> {};
    auto idx = 0;
    for (;;) {
        if (idx >= std::size(instructions)) {
            break;
        }
        auto instr = instructions[idx];
        auto opcodeBytes = instr.OpcodeBytes;
        if (IsOpcode(opcodeBytes, "LDA")) {
            if (instr.Mode == ModeImmediate) {
                code = append(code, std::uint8_t(cpu::OpLDAImm));
                code = append(code, std::uint8_t(instr.Operand));
            } else      if (instr.Mode == ModeZeroPage) {
                code = append(code, std::uint8_t(cpu::OpLDAZp));
                code = append(code, std::uint8_t(instr.Operand));
            } else      if (instr.Mode == ModeZeroPageX) {
                code = append(code, std::uint8_t(cpu::OpLDAZpX));
                code = append(code, std::uint8_t(instr.Operand));
            } else      if (instr.Mode == ModeAbsolute) {
                code = append(code, std::uint8_t(cpu::OpLDAAbs));
                code = append(code, std::uint8_t(instr.Operand & 0xFF));
                code = append(code, std::uint8_t((instr.Operand >> 8) & 0xFF));
            }
        } else    if (IsOpcode(opcodeBytes, "LDX")) {
            if (instr.Mode == ModeImmediate) {
                code = append(code, std::uint8_t(cpu::OpLDXImm));
                code = append(code, std::uint8_t(instr.Operand));
            } else      if (instr.Mode == ModeAbsolute) {
                code = append(code, std::uint8_t(cpu::OpLDXAbs));
                code = append(code, std::uint8_t(instr.Operand & 0xFF));
                code = append(code, std::uint8_t((instr.Operand >> 8) & 0xFF));
            }
        } else    if (IsOpcode(opcodeBytes, "LDY")) {
            if (instr.Mode == ModeImmediate) {
                code = append(code, std::uint8_t(cpu::OpLDYImm));
                code = append(code, std::uint8_t(instr.Operand));
            } else      if (instr.Mode == ModeAbsolute) {
                code = append(code, std::uint8_t(cpu::OpLDYAbs));
                code = append(code, std::uint8_t(instr.Operand & 0xFF));
                code = append(code, std::uint8_t((instr.Operand >> 8) & 0xFF));
            }
        } else    if (IsOpcode(opcodeBytes, "STA")) {
            if (instr.Mode == ModeZeroPage) {
                code = append(code, std::uint8_t(cpu::OpSTAZp));
                code = append(code, std::uint8_t(instr.Operand));
            } else      if (instr.Mode == ModeZeroPageX) {
                code = append(code, std::uint8_t(cpu::OpSTAZpX));
                code = append(code, std::uint8_t(instr.Operand));
            } else      if (instr.Mode == ModeAbsolute) {
                code = append(code, std::uint8_t(cpu::OpSTAAbs));
                code = append(code, std::uint8_t(instr.Operand & 0xFF));
                code = append(code, std::uint8_t((instr.Operand >> 8) & 0xFF));
            }
        } else    if (IsOpcode(opcodeBytes, "STX")) {
            if (instr.Mode == ModeAbsolute) {
                code = append(code, std::uint8_t(cpu::OpSTXAbs));
                code = append(code, std::uint8_t(instr.Operand & 0xFF));
                code = append(code, std::uint8_t((instr.Operand >> 8) & 0xFF));
            }
        } else    if (IsOpcode(opcodeBytes, "STY")) {
            if (instr.Mode == ModeAbsolute) {
                code = append(code, std::uint8_t(cpu::OpSTYAbs));
                code = append(code, std::uint8_t(instr.Operand & 0xFF));
                code = append(code, std::uint8_t((instr.Operand >> 8) & 0xFF));
            }
        } else    if (IsOpcode(opcodeBytes, "ADC")) {
            if (instr.Mode == ModeImmediate) {
                code = append(code, std::uint8_t(cpu::OpADCImm));
                code = append(code, std::uint8_t(instr.Operand));
            }
        } else    if (IsOpcode(opcodeBytes, "SBC")) {
            if (instr.Mode == ModeImmediate) {
                code = append(code, std::uint8_t(cpu::OpSBCImm));
                code = append(code, std::uint8_t(instr.Operand));
            }
        } else    if (IsOpcode(opcodeBytes, "INX")) {
            code = append(code, std::uint8_t(cpu::OpINX));
        } else    if (IsOpcode(opcodeBytes, "INY")) {
            code = append(code, std::uint8_t(cpu::OpINY));
        } else    if (IsOpcode(opcodeBytes, "DEX")) {
            code = append(code, std::uint8_t(cpu::OpDEX));
        } else    if (IsOpcode(opcodeBytes, "DEY")) {
            code = append(code, std::uint8_t(cpu::OpDEY));
        } else    if (IsOpcode(opcodeBytes, "INC")) {
            if (instr.Mode == ModeZeroPage) {
                code = append(code, std::uint8_t(cpu::OpINC));
                code = append(code, std::uint8_t(instr.Operand));
            }
        } else    if (IsOpcode(opcodeBytes, "CMP")) {
            if (instr.Mode == ModeImmediate) {
                code = append(code, std::uint8_t(cpu::OpCMPImm));
                code = append(code, std::uint8_t(instr.Operand));
            }
        } else    if (IsOpcode(opcodeBytes, "CPX")) {
            if (instr.Mode == ModeImmediate) {
                code = append(code, std::uint8_t(cpu::OpCPXImm));
                code = append(code, std::uint8_t(instr.Operand));
            }
        } else    if (IsOpcode(opcodeBytes, "CPY")) {
            if (instr.Mode == ModeImmediate) {
                code = append(code, std::uint8_t(cpu::OpCPYImm));
                code = append(code, std::uint8_t(instr.Operand));
            }
        } else    if (IsOpcode(opcodeBytes, "BNE")) {
            code = append(code, std::uint8_t(cpu::OpBNE));
            code = append(code, std::uint8_t(instr.Operand));
        } else    if (IsOpcode(opcodeBytes, "BEQ")) {
            code = append(code, std::uint8_t(cpu::OpBEQ));
            code = append(code, std::uint8_t(instr.Operand));
        } else    if (IsOpcode(opcodeBytes, "BCC")) {
            code = append(code, std::uint8_t(cpu::OpBCC));
            code = append(code, std::uint8_t(instr.Operand));
        } else    if (IsOpcode(opcodeBytes, "BCS")) {
            code = append(code, std::uint8_t(cpu::OpBCS));
            code = append(code, std::uint8_t(instr.Operand));
        } else    if (IsOpcode(opcodeBytes, "JMP")) {
            code = append(code, std::uint8_t(cpu::OpJMP));
            code = append(code, std::uint8_t(instr.Operand & 0xFF));
            code = append(code, std::uint8_t((instr.Operand >> 8) & 0xFF));
        } else    if (IsOpcode(opcodeBytes, "JSR")) {
            code = append(code, std::uint8_t(cpu::OpJSR));
            code = append(code, std::uint8_t(instr.Operand & 0xFF));
            code = append(code, std::uint8_t((instr.Operand >> 8) & 0xFF));
        } else    if (IsOpcode(opcodeBytes, "RTS")) {
            code = append(code, std::uint8_t(cpu::OpRTS));
        } else    if (IsOpcode(opcodeBytes, "NOP")) {
            code = append(code, std::uint8_t(cpu::OpNOP));
        } else    if (IsOpcode(opcodeBytes, "BRK")) {
            code = append(code, std::uint8_t(cpu::OpBRK));
        }
        idx = idx + 1;
    }
    return code;
}

std::vector<std::uint8_t> AssembleString(std::string text)
{
    auto tokens = Tokenize(text);
    auto instructions = Parse(tokens);
    return Assemble(instructions);
}

std::vector<std::int8_t> AppendLineBytes(std::vector<std::int8_t> allBytes, std::vector<std::int8_t> lineBytes)
{
    auto j = 0;
    for (;;) {
        if (j >= std::size(lineBytes)) {
            break;
        }
        allBytes = append(allBytes, lineBytes[j]);
        j = j + 1;
    }
    return allBytes;
}

std::vector<Token> TokenizeBytes(std::vector<std::int8_t> bytes)
{
    auto tokens = std::vector<Token> {};
    auto i = 0;
    for (;;) {
        if (i >= std::size(bytes)) {
            break;
        }
        auto b = bytes[i];
        if (IsWhitespace(b)) {
            i = i + 1;
            continue;
        }
        if (b == '\n') {
            tokens = append(tokens, Token{.Type= TokenTypeNewline
                                                 , .Representation= std::vector<std::int8_t>{b}
                                         });
            i = i + 1;
            continue;
        }
        if (b == ';') {
            for (;;) {
                if (i >= std::size(bytes)) {
                    break;
                }
                if (bytes[i] == '\n') {
                    break;
                }
                i = i + 1;
            }
            continue;
        }
        if (b == '#') {
            tokens = append(tokens, Token{.Type= TokenTypeHash
                                                 , .Representation= std::vector<std::int8_t>{b}
                                         });
            i = i + 1;
            continue;
        }
        if (b == '$') {
            tokens = append(tokens, Token{.Type= TokenTypeDollar
                                                 , .Representation= std::vector<std::int8_t>{b}
                                         });
            i = i + 1;
            continue;
        }
        if (b == ':') {
            tokens = append(tokens, Token{.Type= TokenTypeColon
                                                 , .Representation= std::vector<std::int8_t>{b}
                                         });
            i = i + 1;
            continue;
        }
        if (b == ',') {
            tokens = append(tokens, Token{.Type= TokenTypeComma
                                                 , .Representation= std::vector<std::int8_t>{b}
                                         });
            i = i + 1;
            continue;
        }
        if (IsHexDigit(b)) {
            auto repr = std::vector<std::int8_t> {};
            for (;;) {
                if (i >= std::size(bytes)) {
                    break;
                }
                if ((!IsHexDigit(bytes[i]))) {
                    break;
                }
                repr = append(repr, bytes[i]);
                i = i + 1;
            }
            tokens = append(tokens, Token{.Type= TokenTypeNumber
                                                 , .Representation= repr
                                         });
            continue;
        }
        if (IsAlpha(b)) {
            auto repr = std::vector<std::int8_t> {};
            for (;;) {
                if (i >= std::size(bytes)) {
                    break;
                }
                if ((!IsAlpha(bytes[i])) && (!IsDigit(bytes[i]))) {
                    break;
                }
                repr = append(repr, bytes[i]);
                i = i + 1;
            }
            tokens = append(tokens, Token{.Type= TokenTypeIdentifier
                                                 , .Representation= repr
                                         });
            continue;
        }
        i = i + 1;
    }
    return tokens;
}

std::vector<std::uint8_t> AssembleLines(std::vector<std::string> lines)
{
    auto allBytes = std::vector<std::int8_t> {};
    auto i = 0;
    for (;;) {
        if (i >= std::size(lines)) {
            break;
        }
        auto lineBytes = StringToBytes(lines[i]);
        allBytes = AppendLineBytes(allBytes, lineBytes);
        if (i < std::size(lines) - 1) {
            allBytes = append(allBytes, std::int8_t(10));
        }
        i = i + 1;
    }
    auto tokens = TokenizeBytes(allBytes);
    auto instructions = Parse(tokens);
    return Assemble(instructions);
}

} // namespace assembler

namespace font {

// Forward declarations
std::vector<std::uint8_t> GetFontData();
std::vector<std::uint8_t> GetCharBitmap(std::vector<std::uint8_t> fontData, int charCode);
bool GetPixel(std::vector<std::uint8_t> fontData, int charCode, int x, int y);

std::vector<std::uint8_t> GetFontData()
{
    return std::vector<std::uint8_t> {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x18, 0x18, 0x18, 0x18, 0x18, 0x00, 0x18, 0x00, 0x6C, 0x6C, 0x24, 0x00, 0x00, 0x00, 0x00, 0x00, 0x6C, 0x6C, 0xFE, 0x6C, 0xFE, 0x6C, 0x6C, 0x00, 0x18, 0x3E, 0x60, 0x3C, 0x06, 0x7C, 0x18, 0x00, 0x00, 0xC6, 0xCC, 0x18, 0x30, 0x66, 0xC6, 0x00, 0x38, 0x6C, 0x38, 0x76, 0xDC, 0xCC, 0x76, 0x00, 0x18, 0x18, 0x30, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0C, 0x18, 0x30, 0x30, 0x30, 0x18, 0x0C, 0x00, 0x30, 0x18, 0x0C, 0x0C, 0x0C, 0x18, 0x30, 0x00, 0x00, 0x66, 0x3C, 0xFF, 0x3C, 0x66, 0x00, 0x00, 0x00, 0x18, 0x18, 0x7E, 0x18, 0x18, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x18, 0x18, 0x30, 0x00, 0x00, 0x00, 0x7E, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x18, 0x18, 0x00, 0x06, 0x0C, 0x18, 0x30, 0x60, 0xC0, 0x80, 0x00, 0x7C, 0xC6, 0xCE, 0xD6, 0xE6, 0xC6, 0x7C, 0x00, 0x18, 0x38, 0x18, 0x18, 0x18, 0x18, 0x7E, 0x00, 0x7C, 0xC6, 0x06, 0x1C, 0x30, 0x66, 0xFE, 0x00, 0x7C, 0xC6, 0x06, 0x3C, 0x06, 0xC6, 0x7C, 0x00, 0x1C, 0x3C, 0x6C, 0xCC, 0xFE, 0x0C, 0x1E, 0x00, 0xFE, 0xC0, 0xC0, 0xFC, 0x06, 0xC6, 0x7C, 0x00, 0x38, 0x60, 0xC0, 0xFC, 0xC6, 0xC6, 0x7C, 0x00, 0xFE, 0xC6, 0x0C, 0x18, 0x30, 0x30, 0x30, 0x00, 0x7C, 0xC6, 0xC6, 0x7C, 0xC6, 0xC6, 0x7C, 0x00, 0x7C, 0xC6, 0xC6, 0x7E, 0x06, 0x0C, 0x78, 0x00, 0x00, 0x18, 0x18, 0x00, 0x00, 0x18, 0x18, 0x00, 0x00, 0x18, 0x18, 0x00, 0x00, 0x18, 0x18, 0x30, 0x06, 0x0C, 0x18, 0x30, 0x18, 0x0C, 0x06, 0x00, 0x00, 0x00, 0x7E, 0x00, 0x00, 0x7E, 0x00, 0x00, 0x60, 0x30, 0x18, 0x0C, 0x18, 0x30, 0x60, 0x00, 0x7C, 0xC6, 0x0C, 0x18, 0x18, 0x00, 0x18, 0x00, 0x7C, 0xC6, 0xDE, 0xDE, 0xDE, 0xC0, 0x78, 0x00, 0x38, 0x6C, 0xC6, 0xFE, 0xC6, 0xC6, 0xC6, 0x00, 0xFC, 0x66, 0x66, 0x7C, 0x66, 0x66, 0xFC, 0x00, 0x3C, 0x66, 0xC0, 0xC0, 0xC0, 0x66, 0x3C, 0x00, 0xF8, 0x6C, 0x66, 0x66, 0x66, 0x6C, 0xF8, 0x00, 0xFE, 0x62, 0x68, 0x78, 0x68, 0x62, 0xFE, 0x00, 0xFE, 0x62, 0x68, 0x78, 0x68, 0x60, 0xF0, 0x00, 0x3C, 0x66, 0xC0, 0xC0, 0xCE, 0x66, 0x3A, 0x00, 0xC6, 0xC6, 0xC6, 0xFE, 0xC6, 0xC6, 0xC6, 0x00, 0x3C, 0x18, 0x18, 0x18, 0x18, 0x18, 0x3C, 0x00, 0x1E, 0x0C, 0x0C, 0x0C, 0xCC, 0xCC, 0x78, 0x00, 0xE6, 0x66, 0x6C, 0x78, 0x6C, 0x66, 0xE6, 0x00, 0xF0, 0x60, 0x60, 0x60, 0x62, 0x66, 0xFE, 0x00, 0xC6, 0xEE, 0xFE, 0xFE, 0xD6, 0xC6, 0xC6, 0x00, 0xC6, 0xE6, 0xF6, 0xDE, 0xCE, 0xC6, 0xC6, 0x00, 0x7C, 0xC6, 0xC6, 0xC6, 0xC6, 0xC6, 0x7C, 0x00, 0xFC, 0x66, 0x66, 0x7C, 0x60, 0x60, 0xF0, 0x00, 0x7C, 0xC6, 0xC6, 0xC6, 0xD6, 0xDE, 0x7C, 0x06, 0xFC, 0x66, 0x66, 0x7C, 0x6C, 0x66, 0xE6, 0x00, 0x7C, 0xC6, 0x60, 0x38, 0x0C, 0xC6, 0x7C, 0x00, 0x7E, 0x7E, 0x5A, 0x18, 0x18, 0x18, 0x3C, 0x00, 0xC6, 0xC6, 0xC6, 0xC6, 0xC6, 0xC6, 0x7C, 0x00, 0xC6, 0xC6, 0xC6, 0xC6, 0xC6, 0x6C, 0x38, 0x00, 0xC6, 0xC6, 0xC6, 0xD6, 0xD6, 0xFE, 0x6C, 0x00, 0xC6, 0xC6, 0x6C, 0x38, 0x6C, 0xC6, 0xC6, 0x00, 0x66, 0x66, 0x66, 0x3C, 0x18, 0x18, 0x3C, 0x00, 0xFE, 0xC6, 0x8C, 0x18, 0x32, 0x66, 0xFE, 0x00, 0x3C, 0x30, 0x30, 0x30, 0x30, 0x30, 0x3C, 0x00, 0xC0, 0x60, 0x30, 0x18, 0x0C, 0x06, 0x02, 0x00, 0x3C, 0x0C, 0x0C, 0x0C, 0x0C, 0x0C, 0x3C, 0x00, 0x10, 0x38, 0x6C, 0xC6, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0x30, 0x18, 0x0C, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x78, 0x0C, 0x7C, 0xCC, 0x76, 0x00, 0xE0, 0x60, 0x7C, 0x66, 0x66, 0x66, 0xDC, 0x00, 0x00, 0x00, 0x7C, 0xC6, 0xC0, 0xC6, 0x7C, 0x00, 0x1C, 0x0C, 0x7C, 0xCC, 0xCC, 0xCC, 0x76, 0x00, 0x00, 0x00, 0x7C, 0xC6, 0xFE, 0xC0, 0x7C, 0x00, 0x3C, 0x66, 0x60, 0xF8, 0x60, 0x60, 0xF0, 0x00, 0x00, 0x00, 0x76, 0xCC, 0xCC, 0x7C, 0x0C, 0xF8, 0xE0, 0x60, 0x6C, 0x76, 0x66, 0x66, 0xE6, 0x00, 0x18, 0x00, 0x38, 0x18, 0x18, 0x18, 0x3C, 0x00, 0x06, 0x00, 0x06, 0x06, 0x06, 0x66, 0x66, 0x3C, 0xE0, 0x60, 0x66, 0x6C, 0x78, 0x6C, 0xE6, 0x00, 0x38, 0x18, 0x18, 0x18, 0x18, 0x18, 0x3C, 0x00, 0x00, 0x00, 0xEC, 0xFE, 0xD6, 0xD6, 0xD6, 0x00, 0x00, 0x00, 0xDC, 0x66, 0x66, 0x66, 0x66, 0x00, 0x00, 0x00, 0x7C, 0xC6, 0xC6, 0xC6, 0x7C, 0x00, 0x00, 0x00, 0xDC, 0x66, 0x66, 0x7C, 0x60, 0xF0, 0x00, 0x00, 0x76, 0xCC, 0xCC, 0x7C, 0x0C, 0x1E, 0x00, 0x00, 0xDC, 0x76, 0x60, 0x60, 0xF0, 0x00, 0x00, 0x00, 0x7E, 0xC0, 0x7C, 0x06, 0xFC, 0x00, 0x30, 0x30, 0xFC, 0x30, 0x30, 0x36, 0x1C, 0x00, 0x00, 0x00, 0xCC, 0xCC, 0xCC, 0xCC, 0x76, 0x00, 0x00, 0x00, 0xC6, 0xC6, 0xC6, 0x6C, 0x38, 0x00, 0x00, 0x00, 0xC6, 0xD6, 0xD6, 0xFE, 0x6C, 0x00, 0x00, 0x00, 0xC6, 0x6C, 0x38, 0x6C, 0xC6, 0x00, 0x00, 0x00, 0xC6, 0xC6, 0xC6, 0x7E, 0x06, 0xFC, 0x00, 0x00, 0xFE, 0x8C, 0x18, 0x32, 0xFE, 0x00, 0x0E, 0x18, 0x18, 0x70, 0x18, 0x18, 0x0E, 0x00, 0x18, 0x18, 0x18, 0x18, 0x18, 0x18, 0x18, 0x00, 0x70, 0x18, 0x18, 0x0E, 0x18, 0x18, 0x70, 0x00, 0x76, 0xDC, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF};
}

std::vector<std::uint8_t> GetCharBitmap(std::vector<std::uint8_t> fontData, int charCode)
{
    auto result = std::vector<std::uint8_t> {};
    auto code = charCode;
    if (code < 32) {
        code = 32;
    }
    if (code > 127) {
        code = 127;
    }
    auto offset = (code - 32) * 8;
    auto i = 0;
    for (;;) {
        if (i >= 8) {
            break;
        }
        result = append(result, fontData[offset + i]);
        i = i + 1;
    }
    return result;
}

bool GetPixel(std::vector<std::uint8_t> fontData, int charCode, int x, int y)
{
    auto code = charCode;
    if (code < 32) {
        code = 32;
    }
    if (code > 127) {
        code = 127;
    }
    auto offset = (code - 32) * 8;
    auto row = fontData[offset + y];
    auto mask = std::uint8_t(0x80 >> x);
    return (row & mask) != 0;
}

} // namespace font

constexpr auto TextCols = 40;

constexpr auto TextRows = 25;

constexpr auto TextScreenBase = 0x0400;

// Forward declarations
std::string hexDigit(int n);
std::string toHex2(int n);
std::string toHex4(int n);
std::string toHex(int n);
std::vector<std::string> addStringToScreen(std::vector<std::string> lines, std::string text, int row, int col);
std::vector<std::string> clearScreen(std::vector<std::string> lines);
std::vector<std::uint8_t> createC64WelcomeScreen();
int main();

std::string hexDigit(int n)
{
    if (n == 0) {
        return "0";
    } else  if (n == 1) {
        return "1";
    } else  if (n == 2) {
        return "2";
    } else  if (n == 3) {
        return "3";
    } else  if (n == 4) {
        return "4";
    } else  if (n == 5) {
        return "5";
    } else  if (n == 6) {
        return "6";
    } else  if (n == 7) {
        return "7";
    } else  if (n == 8) {
        return "8";
    } else  if (n == 9) {
        return "9";
    } else  if (n == 10) {
        return "A";
    } else  if (n == 11) {
        return "B";
    } else  if (n == 12) {
        return "C";
    } else  if (n == 13) {
        return "D";
    } else  if (n == 14) {
        return "E";
    } else  if (n == 15) {
        return "F";
    }
    return "0";
}

std::string toHex2(int n)
{
    auto high = (n >> 4) & 0x0F;
    auto low = n & 0x0F;
    return hexDigit(high) + hexDigit(low);
}

std::string toHex4(int n)
{
    return toHex2((n >> 8) & 0xFF) + toHex2(n & 0xFF);
}

std::string toHex(int n)
{
    if (n > 255) {
        return "$" + toHex4(n);
    }
    return "$" + toHex2(n);
}

std::vector<std::string> addStringToScreen(std::vector<std::string> lines, std::string text, int row, int col)
{
    auto baseAddr = TextScreenBase + (row * TextCols) + col;
    auto i = 0;
    for (;;) {
        if (i >= std::size(text)) {
            break;
        }
        auto charCode = int(text[i]);
        auto addr = baseAddr + i;
        lines = append(lines, "LDA #" + toHex(charCode));
        lines = append(lines, "STA " + toHex(addr));
        i = i + 1;
    }
    return lines;
}

std::vector<std::string> clearScreen(std::vector<std::string> lines)
{
    auto addr = TextScreenBase;
    auto i = 0;
    for (;;) {
        if (i >= TextCols * TextRows) {
            break;
        }
        lines = append(lines, "LDA #$20");
        lines = append(lines, "STA " + toHex(addr + i));
        i = i + 1;
    }
    return lines;
}

std::vector<std::uint8_t> createC64WelcomeScreen()
{
    auto lines = std::vector<std::string> {};
    lines = clearScreen(lines);
    lines = addStringToScreen(lines, "**** COMMODORE 64 BASIC V2 ****", 1, 4);
    lines = addStringToScreen(lines, "64K RAM SYSTEM  38911 BASIC BYTES FREE", 3, 1);
    lines = addStringToScreen(lines, "READY.", 5, 0);
    lines = append(lines, "LDA #$5F");
    lines = append(lines, "STA $04F0");
    lines = append(lines, "BRK");
    return assembler::AssembleLines(lines);
}

int main()
{
    auto scale = std::int32_t(2);
    auto windowWidth = std::int32_t(TextCols * 8) * scale;
    auto windowHeight = std::int32_t(TextRows * 8) * scale;
    auto w = graphics::CreateWindow("Commodore 64", windowWidth, windowHeight);
    auto c = cpu::NewCPU();
    auto fontData = font::GetFontData();
    auto program = createC64WelcomeScreen();
    c = cpu::LoadProgram(c, program, 0x0600);
    c = cpu::SetPC(c, 0x0600);
    c = cpu::Run(c, 100000);
    auto textColor = graphics::NewColor(134, 122, 222, 255);
    auto bgColor = graphics::NewColor(64, 50, 133, 255);
    for (;;) {
        bool running;
        std::tie(w, running) = graphics::PollEvents(w);
        if ((!running)) {
            break;
        }
        graphics::Clear(w, bgColor);
        auto charY = 0;
        for (;;) {
            if (charY >= TextRows) {
                break;
            }
            auto charX = 0;
            for (;;) {
                if (charX >= TextCols) {
                    break;
                }
                auto memAddr = TextScreenBase + (charY * TextCols) + charX;
                auto charCode = int(cpu::GetMemory(c, memAddr));
                if (charCode >= 32) {
                    if (charCode <= 127) {
                        auto pixelY = 0;
                        for (;;) {
                            if (pixelY >= 8) {
                                break;
                            }
                            auto pixelX = 0;
                            for (;;) {
                                if (pixelX >= 8) {
                                    break;
                                }
                                if (font::GetPixel(fontData, charCode, pixelX, pixelY)) {
                                    auto screenX = std::int32_t(charX * 8 + pixelX) * scale;
                                    auto screenY = std::int32_t(charY * 8 + pixelY) * scale;
                                    graphics::FillRect(w, graphics::NewRect(screenX, screenY, scale, scale), textColor);
                                }
                                pixelX = pixelX + 1;
                            }
                            pixelY = pixelY + 1;
                        }
                    }
                }
                charX = charX + 1;
            }
            charY = charY + 1;
        }
        graphics::Present(w);
    }
    graphics::CloseWindow(w);
}


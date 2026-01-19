using System;
using System.Collections;
using System.Collections.Generic;

public static class SliceBuiltins {
    public static List<T> Append<T>(this List<T> list, T element)
    {
        var result = list != null ? new List<T>(list) : new List<T>();
        result.Add(element);
        return result;
    }

    public static List<T> Append<T>(this List<T> list, params T[] elements)
    {
        var result = list != null ? new List<T>(list) : new List<T>();
        result.AddRange(elements);
        return result;
    }

    public static List<T> Append<T>(this List<T> list, List<T> elements)
    {
        var result = list != null ? new List<T>(list) : new List<T>();
        result.AddRange(elements);
        return result;
    }

    // Fix: Ensure Length works for collections and not generic T
    public static int Length<T>(ICollection<T> collection)
    {
        return collection == null ? 0 : collection.Count;
    }
    public static int Length(string s)
    {
        return s == null ? 0 : s.Length;
    }
}
public class Formatter {
    public static void Printf(string format, params object[] args)
    {
        int argIndex = 0;
        string converted = "";
        List<object> formattedArgs = new List<object>();

        for (int i = 0; i < format.Length; i++) {
            if (format[i] == '%' && i + 1 < format.Length) {
                char next = format[i + 1];
                switch (next) {
                case 'd':
                case 's':
                case 'f':
                    converted += "{" + argIndex + "}";
                    formattedArgs.Add(args[argIndex]);
                    argIndex++;
                    i++; // skip format char
                    continue;
                case 'c':
                    converted += "{" + argIndex + "}";
                    object arg = args[argIndex];
                    if (arg is sbyte sb)
                        formattedArgs.Add((char)sb); // sbyte to char
                    else if (arg is int iVal)
                        formattedArgs.Add((char)iVal);
                    else if (arg is char cVal)
                        formattedArgs.Add(cVal);
                    else
                        throw new ArgumentException($"Argument {argIndex} for %c must be a char, int, or sbyte");
                    argIndex++;
                    i++; // skip format char
                    continue;
                }
            }

            converted += format[i];
        }

        converted = converted
                    .Replace(@"\n", "\n")
                    .Replace(@"\t", "\t")
                    .Replace(@"\\", "\\");

        Console.Write(string.Format(converted, formattedArgs.ToArray()));
    }

    public static string Sprintf(string format, params object[] args)
    {
        int argIndex = 0;
        string converted = "";
        List<object> formattedArgs = new List<object>();

        for (int i = 0; i < format.Length; i++) {
            if (format[i] == '%' && i + 1 < format.Length) {
                char next = format[i + 1];
                switch (next) {
                case 'd':
                case 's':
                case 'f':
                    converted += "{" + argIndex + "}";
                    formattedArgs.Add(args[argIndex]);
                    argIndex++;
                    i++; // skip format char
                    continue;
                case 'c':
                    converted += "{" + argIndex + "}";
                    object arg = args[argIndex];
                    if (arg is sbyte sb)
                        formattedArgs.Add((char)sb); // sbyte to char
                    else if (arg is int iVal)
                        formattedArgs.Add((char)iVal);
                    else if (arg is char cVal)
                        formattedArgs.Add(cVal);
                    else
                        throw new ArgumentException($"Argument {argIndex} for %c must be a char, int, or sbyte");
                    argIndex++;
                    i++; // skip format char
                    continue;
                }
            }

            converted += format[i];
        }
        converted = converted
                    .Replace(@"\n", "\n")
                    .Replace(@"\t", "\t")
                    .Replace(@"\\", "\\");

        return string.Format(converted, formattedArgs.ToArray());
    }
}

namespace cpu {


public struct Api {

    public struct CPU {
        public  byte A;
        public  byte X;
        public  byte Y;
        public  byte SP;
        public  int PC;
        public  byte Status;
        public  List<byte> Memory;
        public  bool Halted;
        public  int Cycles;
    };

    public const int OpLDAImm = 0xA9;
    public const int OpLDAZp = 0xA5;
    public const int OpLDAZpX = 0xB5;
    public const int OpLDAAbs = 0xAD;
    public const int OpLDXImm = 0xA2;
    public const int OpLDXAbs = 0xAE;
    public const int OpLDYImm = 0xA0;
    public const int OpLDYAbs = 0xAC;
    public const int OpSTAZp = 0x85;
    public const int OpSTAZpX = 0x95;
    public const int OpSTAAbs = 0x8D;
    public const int OpSTXAbs = 0x8E;
    public const int OpSTYAbs = 0x8C;
    public const int OpADCImm = 0x69;
    public const int OpSBCImm = 0xE9;
    public const int OpINX = 0xE8;
    public const int OpINY = 0xC8;
    public const int OpDEX = 0xCA;
    public const int OpDEY = 0x88;
    public const int OpINC = 0xE6;
    public const int OpCMPImm = 0xC9;
    public const int OpCPXImm = 0xE0;
    public const int OpCPYImm = 0xC0;
    public const int OpBNE = 0xD0;
    public const int OpBEQ = 0xF0;
    public const int OpBCC = 0x90;
    public const int OpBCS = 0xB0;
    public const int OpJMP = 0x4C;
    public const int OpJSR = 0x20;
    public const int OpRTS = 0x60;
    public const int OpNOP = 0xEA;
    public const int OpBRK = 0x00;

    public const int FlagC = 0x01;
    public const int FlagZ = 0x02;
    public const int FlagI = 0x04;
    public const int FlagD = 0x08;
    public const int FlagB = 0x10;
    public const int FlagV = 0x40;
    public const int FlagN = 0x80;

    public const int ScreenBase = 0x0200;
    public const int ScreenWidth = 32;
    public const int ScreenHeight = 32;
    public const int ScreenSize = 1024;

    public static CPU NewCPU()
    {
        var mem = new List<byte> {};
        var i = 0;
        for (;;) {
            if ( (i >= 65536 )) {
                break;
            }
            mem = SliceBuiltins.Append(mem, (byte)(0));
            i =  (i + 1 );
        }
        return new CPU{A= 0, X= 0, Y= 0, SP= 0xFF, PC= 0x0600, Status= 0x20, Memory= mem, Halted= false, Cycles= 0};
    }

    public static CPU LoadProgram(CPU c, List<byte> program, int addr)
    {
        var i = 0;
        for (;;) {
            if ( (i >= SliceBuiltins.Length(program) )) {
                break;
            }
            c.Memory[ (addr + i )] = (byte)program[i];
            i =  (i + 1 );
        }
        return c;
    }

    public static CPU SetPC(CPU c, int addr)
    {
        c.PC = addr;
        return c;
    }

    public static byte ReadByte(CPU c, int addr)
    {
        return (byte)c.Memory[addr];
    }

    public static CPU WriteByte(CPU c, int addr, byte value)
    {
        c.Memory[addr] = (byte)value;
        return c;
    }

    public static (CPU,byte) FetchByte(CPU c)
    {
        var value = (byte)c.Memory[c.PC];
        c.PC =  (c.PC + 1 );
        return (c, value);
    }

    public static (CPU,int) FetchWord(CPU c)
    {
        var low = (int)(c.Memory[c.PC]);
        var high = (int)(c.Memory[ (c.PC + 1 )]);
        c.PC =  (c.PC + 2 );
        return (c,  (low +  (high * 256 ) ));
    }

    public static CPU SetZN(CPU c, byte value)
    {
        if ( (value == 0 )) {
            c.Status = (byte) (c.Status | FlagZ );
        } else {
            c.Status = (byte) (c.Status &  (0xFF - FlagZ ) );
        }
        if ( ( (value & 0x80 ) != 0 )) {
            c.Status = (byte) (c.Status | FlagN );
        } else {
            c.Status = (byte) (c.Status &  (0xFF - FlagN ) );
        }
        return c;
    }

    public static CPU SetCarry(CPU c, bool set)
    {
        if (set) {
            c.Status = (byte) (c.Status | FlagC );
        } else {
            c.Status = (byte) (c.Status &  (0xFF - FlagC ) );
        }
        return c;
    }

    public static bool GetCarry(CPU c)
    {
        return  ( (c.Status & FlagC ) != 0 );
    }

    public static bool GetZero(CPU c)
    {
        return  ( (c.Status & FlagZ ) != 0 );
    }

    public static CPU Step(CPU c)
    {
        if (c.Halted) {
            return c;
        }
        byte opcode = default;
        (c, opcode) = FetchByte(c);
        c.Cycles =  (c.Cycles + 1 );
        if ( (opcode == OpLDAImm )) {
            byte value = default;
            (c, value) = FetchByte(c);
            c.A = (byte)value;
            c = SetZN(c, c.A);
        } else if ( (opcode == OpLDAZp )) {
            byte addr = default;
            (c, addr) = FetchByte(c);
            c.A = (byte)c.Memory[(int)(addr)];
            c = SetZN(c, c.A);
        } else if ( (opcode == OpLDAZpX )) {
            byte addr = default;
            (c, addr) = FetchByte(c);
            c.A = (byte)c.Memory[(int)( (addr + c.X ))];
            c = SetZN(c, c.A);
        } else if ( (opcode == OpLDAAbs )) {
            int addr = default;
            (c, addr) = FetchWord(c);
            c.A = (byte)c.Memory[addr];
            c = SetZN(c, c.A);
        } else if ( (opcode == OpLDXImm )) {
            byte value = default;
            (c, value) = FetchByte(c);
            c.X = (byte)value;
            c = SetZN(c, c.X);
        } else if ( (opcode == OpLDXAbs )) {
            int addr = default;
            (c, addr) = FetchWord(c);
            c.X = (byte)c.Memory[addr];
            c = SetZN(c, c.X);
        } else if ( (opcode == OpLDYImm )) {
            byte value = default;
            (c, value) = FetchByte(c);
            c.Y = (byte)value;
            c = SetZN(c, c.Y);
        } else if ( (opcode == OpLDYAbs )) {
            int addr = default;
            (c, addr) = FetchWord(c);
            c.Y = (byte)c.Memory[addr];
            c = SetZN(c, c.Y);
        } else if ( (opcode == OpSTAZp )) {
            byte addr = default;
            (c, addr) = FetchByte(c);
            c.Memory[(int)(addr)] = (byte)c.A;
        } else if ( (opcode == OpSTAZpX )) {
            byte addr = default;
            (c, addr) = FetchByte(c);
            c.Memory[(int)( (addr + c.X ))] = (byte)c.A;
        } else if ( (opcode == OpSTAAbs )) {
            int addr = default;
            (c, addr) = FetchWord(c);
            c.Memory[addr] = (byte)c.A;
        } else if ( (opcode == OpSTXAbs )) {
            int addr = default;
            (c, addr) = FetchWord(c);
            c.Memory[addr] = (byte)c.X;
        } else if ( (opcode == OpSTYAbs )) {
            int addr = default;
            (c, addr) = FetchWord(c);
            c.Memory[addr] = (byte)c.Y;
        } else if ( (opcode == OpADCImm )) {
            byte value = default;
            (c, value) = FetchByte(c);
            var carry = 0;
            if (GetCarry(c)) {
                carry = 1;
            }
            var result =  ( ((int)(c.A) + (int)(value) ) + carry );
            c = SetCarry(c,  (result > 255 ));
            c.A = (byte)(byte)( (result & 0xFF ));
            c = SetZN(c, c.A);
        } else if ( (opcode == OpSBCImm )) {
            byte value = default;
            (c, value) = FetchByte(c);
            var carry = 0;
            if (GetCarry(c)) {
                carry = 1;
            }
            var result =  ( ((int)(c.A) - (int)(value) ) -  (1 - carry ) );
            c = SetCarry(c,  (result >= 0 ));
            c.A = (byte)(byte)( (result & 0xFF ));
            c = SetZN(c, c.A);
        } else if ( (opcode == OpINX )) {
            c.X = (byte) (c.X + 1 );
            c = SetZN(c, c.X);
        } else if ( (opcode == OpINY )) {
            c.Y = (byte) (c.Y + 1 );
            c = SetZN(c, c.Y);
        } else if ( (opcode == OpDEX )) {
            c.X = (byte) (c.X - 1 );
            c = SetZN(c, c.X);
        } else if ( (opcode == OpDEY )) {
            c.Y = (byte) (c.Y - 1 );
            c = SetZN(c, c.Y);
        } else if ( (opcode == OpINC )) {
            byte addr = default;
            (c, addr) = FetchByte(c);
            var val = (byte) (c.Memory[(int)(addr)] + 1 );
            c.Memory[(int)(addr)] = (byte)val;
            c = SetZN(c, val);
        } else if ( (opcode == OpCMPImm )) {
            byte value = default;
            (c, value) = FetchByte(c);
            var result =  ((int)(c.A) - (int)(value) );
            c = SetCarry(c,  (c.A >= value ));
            c = SetZN(c, (byte)( (result & 0xFF )));
        } else if ( (opcode == OpCPXImm )) {
            byte value = default;
            (c, value) = FetchByte(c);
            var result =  ((int)(c.X) - (int)(value) );
            c = SetCarry(c,  (c.X >= value ));
            c = SetZN(c, (byte)( (result & 0xFF )));
        } else if ( (opcode == OpCPYImm )) {
            byte value = default;
            (c, value) = FetchByte(c);
            var result =  ((int)(c.Y) - (int)(value) );
            c = SetCarry(c,  (c.Y >= value ));
            c = SetZN(c, (byte)( (result & 0xFF )));
        } else if ( (opcode == OpBNE )) {
            byte offset = default;
            (c, offset) = FetchByte(c);
            if ((!GetZero(c))) {
                if ( (offset < 128 )) {
                    c.PC =  (c.PC + (int)(offset) );
                } else {
                    c.PC =  (c.PC -  (256 - (int)(offset) ) );
                }
            }
        } else if ( (opcode == OpBEQ )) {
            byte offset = default;
            (c, offset) = FetchByte(c);
            if (GetZero(c)) {
                if ( (offset < 128 )) {
                    c.PC =  (c.PC + (int)(offset) );
                } else {
                    c.PC =  (c.PC -  (256 - (int)(offset) ) );
                }
            }
        } else if ( (opcode == OpBCC )) {
            byte offset = default;
            (c, offset) = FetchByte(c);
            if ((!GetCarry(c))) {
                if ( (offset < 128 )) {
                    c.PC =  (c.PC + (int)(offset) );
                } else {
                    c.PC =  (c.PC -  (256 - (int)(offset) ) );
                }
            }
        } else if ( (opcode == OpBCS )) {
            byte offset = default;
            (c, offset) = FetchByte(c);
            if (GetCarry(c)) {
                if ( (offset < 128 )) {
                    c.PC =  (c.PC + (int)(offset) );
                } else {
                    c.PC =  (c.PC -  (256 - (int)(offset) ) );
                }
            }
        } else if ( (opcode == OpJMP )) {
            int addr = default;
            (c, addr) = FetchWord(c);
            c.PC = addr;
        } else if ( (opcode == OpJSR )) {
            int addr = default;
            (c, addr) = FetchWord(c);
            var retAddr =  (c.PC - 1 );
            c.Memory[ (0x100 + (int)(c.SP) )] = (byte)(byte)( ( (retAddr >> 8 ) & 0xFF ));
            c.SP = (byte) (c.SP - 1 );
            c.Memory[ (0x100 + (int)(c.SP) )] = (byte)(byte)( (retAddr & 0xFF ));
            c.SP = (byte) (c.SP - 1 );
            c.PC = addr;
        } else if ( (opcode == OpRTS )) {
            c.SP = (byte) (c.SP + 1 );
            var low = (int)(c.Memory[ (0x100 + (int)(c.SP) )]);
            c.SP = (byte) (c.SP + 1 );
            var high = (int)(c.Memory[ (0x100 + (int)(c.SP) )]);
            c.PC =  ( ( (high * 256 ) + low ) + 1 );
        } else if ( (opcode == OpNOP )) {
        } else if ( (opcode == OpBRK )) {
            c.Halted = true;
        }
        return c;
    }

    public static CPU Run(CPU c, int maxCycles)
    {
        for (;;) {
            if (c.Halted) {
                break;
            }
            if ( (c.Cycles >= maxCycles )) {
                break;
            }
            c = Step(c);
        }
        return c;
    }

    public static byte GetScreenPixel(CPU c, int x, int y)
    {
        if ( (x < 0 )) {
            return (byte)0;
        }
        if ( (x >= ScreenWidth )) {
            return (byte)0;
        }
        if ( (y < 0 )) {
            return (byte)0;
        }
        if ( (y >= ScreenHeight )) {
            return (byte)0;
        }
        var addr =  ( (ScreenBase +  (y * ScreenWidth ) ) + x );
        return (byte)c.Memory[addr];
    }

    public static bool IsHalted(CPU c)
    {
        return c.Halted;
    }

    public static byte GetMemory(CPU c, int addr)
    {
        if ( (addr < 0 )) {
            return (byte)0;
        }
        if ( (addr >= 65536 )) {
            return (byte)0;
        }
        return (byte)c.Memory[addr];
    }

}
}
namespace assembler {


public struct Api {

    public struct Instruction {
        public  List<sbyte> OpcodeBytes;
        public  sbyte Mode;
        public  int Operand;
        public  List<sbyte> LabelBytes;
        public  bool HasLabel;
    };

    public struct Token {
        public  sbyte Type;
        public  List<sbyte> Representation;
    };

    public const int TokenTypeInstruction = 1;
    public const int TokenTypeNumber = 2;
    public const int TokenTypeLabel = 3;
    public const int TokenTypeComma = 4;
    public const int TokenTypeNewline = 5;
    public const int TokenTypeHash = 6;
    public const int TokenTypeDollar = 7;
    public const int TokenTypeColon = 8;
    public const int TokenTypeIdentifier = 9;
    public const int TokenTypeComment = 10;

    public const int ModeImplied = 0;
    public const int ModeImmediate = 1;
    public const int ModeZeroPage = 2;
    public const int ModeAbsolute = 3;
    public const int ModeZeroPageX = 4;

    public static bool IsDigit(sbyte b)
    {
        return  ( (b >= '0' ) &&  (b <= '9' ) );
    }

    public static bool IsHexDigit(sbyte b)
    {
        return  ( (IsDigit(b) ||  ( (b >= 'a' ) &&  (b <= 'f' ) ) ) ||  ( (b >= 'A' ) &&  (b <= 'F' ) ) );
    }

    public static bool IsAlpha(sbyte b)
    {
        return  ( ( ( (b >= 'a' ) &&  (b <= 'z' ) ) ||  ( (b >= 'A' ) &&  (b <= 'Z' ) ) ) ||  (b == '_' ) );
    }

    public static bool IsWhitespace(sbyte b)
    {
        return  ( (b == ' ' ) ||  (b == '\t' ) );
    }

    public static List<sbyte> StringToBytes(string s)
    {
        var result = new List<sbyte> {};
        var i = 0;
        for (;;) {
            if ( (i >= SliceBuiltins.Length(s) )) {
                break;
            }
            result = SliceBuiltins.Append(result, (sbyte)(s[i]));
            i =  (i + 1 );
        }
        return result;
    }

    public static sbyte ToUpper(sbyte b)
    {
        if ( ( (b >= 'a' ) &&  (b <= 'z' ) )) {
            return (sbyte) (b - 32 );
        }
        return (sbyte)b;
    }

    public static List<Token> Tokenize(string text)
    {
        var tokens = new List<Token> {};
        var bytes = StringToBytes(text);
        var i = 0;
        for (;;) {
            if ( (i >= SliceBuiltins.Length(bytes) )) {
                break;
            }
            var b = (sbyte)bytes[i];
            if (IsWhitespace(b)) {
                i =  (i + 1 );
                continue;
            }
            if ( (b == '\n' )) {
                tokens = SliceBuiltins.Append(tokens, new Token{Type= TokenTypeNewline, Representation= new List<sbyte>{b}});
                i =  (i + 1 );
                continue;
            }
            if ( (b == ';' )) {
                for (;;) {
                    if ( (i >= SliceBuiltins.Length(bytes) )) {
                        break;
                    }
                    if ( (bytes[i] == '\n' )) {
                        break;
                    }
                    i =  (i + 1 );
                }
                continue;
            }
            if ( (b == '#' )) {
                tokens = SliceBuiltins.Append(tokens, new Token{Type= TokenTypeHash, Representation= new List<sbyte>{b}});
                i =  (i + 1 );
                continue;
            }
            if ( (b == '$' )) {
                tokens = SliceBuiltins.Append(tokens, new Token{Type= TokenTypeDollar, Representation= new List<sbyte>{b}});
                i =  (i + 1 );
                continue;
            }
            if ( (b == ':' )) {
                tokens = SliceBuiltins.Append(tokens, new Token{Type= TokenTypeColon, Representation= new List<sbyte>{b}});
                i =  (i + 1 );
                continue;
            }
            if ( (b == ',' )) {
                tokens = SliceBuiltins.Append(tokens, new Token{Type= TokenTypeComma, Representation= new List<sbyte>{b}});
                i =  (i + 1 );
                continue;
            }
            if (IsHexDigit(b)) {
                var repr = new List<sbyte> {};
                for (;;) {
                    if ( (i >= SliceBuiltins.Length(bytes) )) {
                        break;
                    }
                    if ((!IsHexDigit(bytes[i]))) {
                        break;
                    }
                    repr = SliceBuiltins.Append(repr, bytes[i]);
                    i =  (i + 1 );
                }
                tokens = SliceBuiltins.Append(tokens, new Token{Type= TokenTypeNumber, Representation= repr});
                continue;
            }
            if (IsAlpha(b)) {
                var repr = new List<sbyte> {};
                for (;;) {
                    if ( (i >= SliceBuiltins.Length(bytes) )) {
                        break;
                    }
                    if ( ((!IsAlpha(bytes[i])) && (!IsDigit(bytes[i])) )) {
                        break;
                    }
                    repr = SliceBuiltins.Append(repr, bytes[i]);
                    i =  (i + 1 );
                }
                tokens = SliceBuiltins.Append(tokens, new Token{Type= TokenTypeIdentifier, Representation= repr});
                continue;
            }
            i =  (i + 1 );
        }
        return tokens;
    }

    public static int ParseHex(List<sbyte> bytes)
    {
        var result = 0;
        var i = 0;
        for (;;) {
            if ( (i >= SliceBuiltins.Length(bytes) )) {
                break;
            }
            var b = (sbyte)bytes[i];
            result =  (result * 16 );
            if ( ( (b >= '0' ) &&  (b <= '9' ) )) {
                result =  (result + (int)( (b - '0' )) );
            } else if ( ( (b >= 'a' ) &&  (b <= 'f' ) )) {
                result =  (result + (int)( ( (b - 'a' ) + 10 )) );
            } else if ( ( (b >= 'A' ) &&  (b <= 'F' ) )) {
                result =  (result + (int)( ( (b - 'A' ) + 10 )) );
            }
            i =  (i + 1 );
        }
        return result;
    }

    public static int ParseDecimal(List<sbyte> bytes)
    {
        var result = 0;
        var i = 0;
        for (;;) {
            if ( (i >= SliceBuiltins.Length(bytes) )) {
                break;
            }
            var b = (sbyte)bytes[i];
            result =  ( (result * 10 ) + (int)( (b - '0' )) );
            i =  (i + 1 );
        }
        return result;
    }

    public static bool MatchToken(Token token, string s)
    {
        if ( (SliceBuiltins.Length(token.Representation) != SliceBuiltins.Length(s) )) {
            return false;
        }
        var i = 0;
        for (;;) {
            if ( (i >= SliceBuiltins.Length(s) )) {
                break;
            }
            if ( (ToUpper(token.Representation[i]) != ToUpper((sbyte)(s[i])) )) {
                return false;
            }
            i =  (i + 1 );
        }
        return true;
    }

    public static List<sbyte> CopyBytes(List<sbyte> src)
    {
        var dst = new List<sbyte> {};
        var i = 0;
        for (;;) {
            if ( (i >= SliceBuiltins.Length(src) )) {
                break;
            }
            dst = SliceBuiltins.Append(dst, src[i]);
            i =  (i + 1 );
        }
        return dst;
    }

    public static List<Instruction> Parse(List<Token> tokens)
    {
        var instructions = new List<Instruction> {};
        var i = 0;
        for (;;) {
            if ( (i >= SliceBuiltins.Length(tokens) )) {
                break;
            }
            if ( (tokens[i].Type == TokenTypeNewline )) {
                i =  (i + 1 );
                continue;
            }
            var currentLabelBytes = new List<sbyte> {};
            var hasLabel = false;
            if ( ( ( (tokens[i].Type == TokenTypeIdentifier ) &&  ( (i + 1 ) < SliceBuiltins.Length(tokens) ) ) &&  (tokens[ (i + 1 )].Type == TokenTypeColon ) )) {
                currentLabelBytes = CopyBytes(tokens[i].Representation);
                hasLabel = true;
                i =  (i + 2 );
                for (;;) {
                    if ( (i >= SliceBuiltins.Length(tokens) )) {
                        break;
                    }
                    if ( (tokens[i].Type != TokenTypeNewline )) {
                        break;
                    }
                    i =  (i + 1 );
                }
                if ( (i >= SliceBuiltins.Length(tokens) )) {
                    break;
                }
            }
            if ( (tokens[i].Type != TokenTypeIdentifier )) {
                i =  (i + 1 );
                continue;
            }
            var instr = new Instruction{OpcodeBytes= CopyBytes(tokens[i].Representation), Mode= ModeImplied, Operand= 0, LabelBytes= currentLabelBytes, HasLabel= hasLabel};
            i =  (i + 1 );
            if ( ( (i < SliceBuiltins.Length(tokens) ) &&  (tokens[i].Type != TokenTypeNewline ) )) {
                if ( (tokens[i].Type == TokenTypeHash )) {
                    i =  (i + 1 );
                    instr.Mode = (sbyte)ModeImmediate;
                    if ( ( (i < SliceBuiltins.Length(tokens) ) &&  (tokens[i].Type == TokenTypeDollar ) )) {
                        i =  (i + 1 );
                        if ( ( (i < SliceBuiltins.Length(tokens) ) &&  (tokens[i].Type == TokenTypeNumber ) )) {
                            instr.Operand = ParseHex(tokens[i].Representation);
                            i =  (i + 1 );
                        }
                    } else if ( ( (i < SliceBuiltins.Length(tokens) ) &&  (tokens[i].Type == TokenTypeNumber ) )) {
                        instr.Operand = ParseDecimal(tokens[i].Representation);
                        i =  (i + 1 );
                    }
                } else if ( (tokens[i].Type == TokenTypeDollar )) {
                    i =  (i + 1 );
                    if ( ( (i < SliceBuiltins.Length(tokens) ) &&  (tokens[i].Type == TokenTypeNumber ) )) {
                        instr.Operand = ParseHex(tokens[i].Representation);
                        if ( (SliceBuiltins.Length(tokens[i].Representation) <= 2 )) {
                            instr.Mode = (sbyte)ModeZeroPage;
                        } else {
                            instr.Mode = (sbyte)ModeAbsolute;
                        }
                        i =  (i + 1 );
                        if ( ( (i < SliceBuiltins.Length(tokens) ) &&  (tokens[i].Type == TokenTypeComma ) )) {
                            i =  (i + 1 );
                            if ( ( (i < SliceBuiltins.Length(tokens) ) &&  (tokens[i].Type == TokenTypeIdentifier ) )) {
                                if (MatchToken(tokens[i], @"X")) {
                                    instr.Mode = (sbyte)ModeZeroPageX;
                                }
                                i =  (i + 1 );
                            }
                        }
                    }
                } else if ( (tokens[i].Type == TokenTypeNumber )) {
                    instr.Operand = ParseDecimal(tokens[i].Representation);
                    if ( (instr.Operand <= 255 )) {
                        instr.Mode = (sbyte)ModeZeroPage;
                    } else {
                        instr.Mode = (sbyte)ModeAbsolute;
                    }
                    i =  (i + 1 );
                }
            }
            instructions = SliceBuiltins.Append(instructions, instr);
        }
        return instructions;
    }

    public static bool IsOpcode(List<sbyte> opcodeBytes, string name)
    {
        if ( (SliceBuiltins.Length(opcodeBytes) != SliceBuiltins.Length(name) )) {
            return false;
        }
        var i = 0;
        for (;;) {
            if ( (i >= SliceBuiltins.Length(name) )) {
                break;
            }
            var ob = (sbyte)opcodeBytes[i];
            if ( ( (ob >= 'a' ) &&  (ob <= 'z' ) )) {
                ob = (sbyte) (ob - 32 );
            }
            var nb = (sbyte)(sbyte)(name[i]);
            if ( (ob != nb )) {
                return false;
            }
            i =  (i + 1 );
        }
        return true;
    }

    public static List<byte> Assemble(List<Instruction> instructions)
    {
        var code = new List<byte> {};
        var idx = 0;
        for (;;) {
            if ( (idx >= SliceBuiltins.Length(instructions) )) {
                break;
            }
            var instr = instructions[idx];
            var opcodeBytes = instr.OpcodeBytes;
            if (IsOpcode(opcodeBytes, @"LDA")) {
                if ( (instr.Mode == ModeImmediate )) {
                    code = SliceBuiltins.Append(code, (byte)(cpu.Api.OpLDAImm));
                    code = SliceBuiltins.Append(code, (byte)(instr.Operand));
                } else if ( (instr.Mode == ModeZeroPage )) {
                    code = SliceBuiltins.Append(code, (byte)(cpu.Api.OpLDAZp));
                    code = SliceBuiltins.Append(code, (byte)(instr.Operand));
                } else if ( (instr.Mode == ModeZeroPageX )) {
                    code = SliceBuiltins.Append(code, (byte)(cpu.Api.OpLDAZpX));
                    code = SliceBuiltins.Append(code, (byte)(instr.Operand));
                } else if ( (instr.Mode == ModeAbsolute )) {
                    code = SliceBuiltins.Append(code, (byte)(cpu.Api.OpLDAAbs));
                    code = SliceBuiltins.Append(code, (byte)( (instr.Operand & 0xFF )));
                    code = SliceBuiltins.Append(code, (byte)( ( (instr.Operand >> 8 ) & 0xFF )));
                }
            } else if (IsOpcode(opcodeBytes, @"LDX")) {
                if ( (instr.Mode == ModeImmediate )) {
                    code = SliceBuiltins.Append(code, (byte)(cpu.Api.OpLDXImm));
                    code = SliceBuiltins.Append(code, (byte)(instr.Operand));
                } else if ( (instr.Mode == ModeAbsolute )) {
                    code = SliceBuiltins.Append(code, (byte)(cpu.Api.OpLDXAbs));
                    code = SliceBuiltins.Append(code, (byte)( (instr.Operand & 0xFF )));
                    code = SliceBuiltins.Append(code, (byte)( ( (instr.Operand >> 8 ) & 0xFF )));
                }
            } else if (IsOpcode(opcodeBytes, @"LDY")) {
                if ( (instr.Mode == ModeImmediate )) {
                    code = SliceBuiltins.Append(code, (byte)(cpu.Api.OpLDYImm));
                    code = SliceBuiltins.Append(code, (byte)(instr.Operand));
                } else if ( (instr.Mode == ModeAbsolute )) {
                    code = SliceBuiltins.Append(code, (byte)(cpu.Api.OpLDYAbs));
                    code = SliceBuiltins.Append(code, (byte)( (instr.Operand & 0xFF )));
                    code = SliceBuiltins.Append(code, (byte)( ( (instr.Operand >> 8 ) & 0xFF )));
                }
            } else if (IsOpcode(opcodeBytes, @"STA")) {
                if ( (instr.Mode == ModeZeroPage )) {
                    code = SliceBuiltins.Append(code, (byte)(cpu.Api.OpSTAZp));
                    code = SliceBuiltins.Append(code, (byte)(instr.Operand));
                } else if ( (instr.Mode == ModeZeroPageX )) {
                    code = SliceBuiltins.Append(code, (byte)(cpu.Api.OpSTAZpX));
                    code = SliceBuiltins.Append(code, (byte)(instr.Operand));
                } else if ( (instr.Mode == ModeAbsolute )) {
                    code = SliceBuiltins.Append(code, (byte)(cpu.Api.OpSTAAbs));
                    code = SliceBuiltins.Append(code, (byte)( (instr.Operand & 0xFF )));
                    code = SliceBuiltins.Append(code, (byte)( ( (instr.Operand >> 8 ) & 0xFF )));
                }
            } else if (IsOpcode(opcodeBytes, @"STX")) {
                if ( (instr.Mode == ModeAbsolute )) {
                    code = SliceBuiltins.Append(code, (byte)(cpu.Api.OpSTXAbs));
                    code = SliceBuiltins.Append(code, (byte)( (instr.Operand & 0xFF )));
                    code = SliceBuiltins.Append(code, (byte)( ( (instr.Operand >> 8 ) & 0xFF )));
                }
            } else if (IsOpcode(opcodeBytes, @"STY")) {
                if ( (instr.Mode == ModeAbsolute )) {
                    code = SliceBuiltins.Append(code, (byte)(cpu.Api.OpSTYAbs));
                    code = SliceBuiltins.Append(code, (byte)( (instr.Operand & 0xFF )));
                    code = SliceBuiltins.Append(code, (byte)( ( (instr.Operand >> 8 ) & 0xFF )));
                }
            } else if (IsOpcode(opcodeBytes, @"ADC")) {
                if ( (instr.Mode == ModeImmediate )) {
                    code = SliceBuiltins.Append(code, (byte)(cpu.Api.OpADCImm));
                    code = SliceBuiltins.Append(code, (byte)(instr.Operand));
                }
            } else if (IsOpcode(opcodeBytes, @"SBC")) {
                if ( (instr.Mode == ModeImmediate )) {
                    code = SliceBuiltins.Append(code, (byte)(cpu.Api.OpSBCImm));
                    code = SliceBuiltins.Append(code, (byte)(instr.Operand));
                }
            } else if (IsOpcode(opcodeBytes, @"INX")) {
                code = SliceBuiltins.Append(code, (byte)(cpu.Api.OpINX));
            } else if (IsOpcode(opcodeBytes, @"INY")) {
                code = SliceBuiltins.Append(code, (byte)(cpu.Api.OpINY));
            } else if (IsOpcode(opcodeBytes, @"DEX")) {
                code = SliceBuiltins.Append(code, (byte)(cpu.Api.OpDEX));
            } else if (IsOpcode(opcodeBytes, @"DEY")) {
                code = SliceBuiltins.Append(code, (byte)(cpu.Api.OpDEY));
            } else if (IsOpcode(opcodeBytes, @"INC")) {
                if ( (instr.Mode == ModeZeroPage )) {
                    code = SliceBuiltins.Append(code, (byte)(cpu.Api.OpINC));
                    code = SliceBuiltins.Append(code, (byte)(instr.Operand));
                }
            } else if (IsOpcode(opcodeBytes, @"CMP")) {
                if ( (instr.Mode == ModeImmediate )) {
                    code = SliceBuiltins.Append(code, (byte)(cpu.Api.OpCMPImm));
                    code = SliceBuiltins.Append(code, (byte)(instr.Operand));
                }
            } else if (IsOpcode(opcodeBytes, @"CPX")) {
                if ( (instr.Mode == ModeImmediate )) {
                    code = SliceBuiltins.Append(code, (byte)(cpu.Api.OpCPXImm));
                    code = SliceBuiltins.Append(code, (byte)(instr.Operand));
                }
            } else if (IsOpcode(opcodeBytes, @"CPY")) {
                if ( (instr.Mode == ModeImmediate )) {
                    code = SliceBuiltins.Append(code, (byte)(cpu.Api.OpCPYImm));
                    code = SliceBuiltins.Append(code, (byte)(instr.Operand));
                }
            } else if (IsOpcode(opcodeBytes, @"BNE")) {
                code = SliceBuiltins.Append(code, (byte)(cpu.Api.OpBNE));
                code = SliceBuiltins.Append(code, (byte)(instr.Operand));
            } else if (IsOpcode(opcodeBytes, @"BEQ")) {
                code = SliceBuiltins.Append(code, (byte)(cpu.Api.OpBEQ));
                code = SliceBuiltins.Append(code, (byte)(instr.Operand));
            } else if (IsOpcode(opcodeBytes, @"BCC")) {
                code = SliceBuiltins.Append(code, (byte)(cpu.Api.OpBCC));
                code = SliceBuiltins.Append(code, (byte)(instr.Operand));
            } else if (IsOpcode(opcodeBytes, @"BCS")) {
                code = SliceBuiltins.Append(code, (byte)(cpu.Api.OpBCS));
                code = SliceBuiltins.Append(code, (byte)(instr.Operand));
            } else if (IsOpcode(opcodeBytes, @"JMP")) {
                code = SliceBuiltins.Append(code, (byte)(cpu.Api.OpJMP));
                code = SliceBuiltins.Append(code, (byte)( (instr.Operand & 0xFF )));
                code = SliceBuiltins.Append(code, (byte)( ( (instr.Operand >> 8 ) & 0xFF )));
            } else if (IsOpcode(opcodeBytes, @"JSR")) {
                code = SliceBuiltins.Append(code, (byte)(cpu.Api.OpJSR));
                code = SliceBuiltins.Append(code, (byte)( (instr.Operand & 0xFF )));
                code = SliceBuiltins.Append(code, (byte)( ( (instr.Operand >> 8 ) & 0xFF )));
            } else if (IsOpcode(opcodeBytes, @"RTS")) {
                code = SliceBuiltins.Append(code, (byte)(cpu.Api.OpRTS));
            } else if (IsOpcode(opcodeBytes, @"NOP")) {
                code = SliceBuiltins.Append(code, (byte)(cpu.Api.OpNOP));
            } else if (IsOpcode(opcodeBytes, @"BRK")) {
                code = SliceBuiltins.Append(code, (byte)(cpu.Api.OpBRK));
            }
            idx =  (idx + 1 );
        }
        return code;
    }

    public static List<byte> AssembleString(string text)
    {
        var tokens = Tokenize(text);
        var instructions = Parse(tokens);
        return Assemble(instructions);
    }

    public static List<sbyte> AppendLineBytes(List<sbyte> allBytes, List<sbyte> lineBytes)
    {
        var j = 0;
        for (;;) {
            if ( (j >= SliceBuiltins.Length(lineBytes) )) {
                break;
            }
            allBytes = SliceBuiltins.Append(allBytes, lineBytes[j]);
            j =  (j + 1 );
        }
        return allBytes;
    }

    public static List<Token> TokenizeBytes(List<sbyte> bytes)
    {
        var tokens = new List<Token> {};
        var i = 0;
        for (;;) {
            if ( (i >= SliceBuiltins.Length(bytes) )) {
                break;
            }
            var b = (sbyte)bytes[i];
            if (IsWhitespace(b)) {
                i =  (i + 1 );
                continue;
            }
            if ( (b == '\n' )) {
                tokens = SliceBuiltins.Append(tokens, new Token{Type= TokenTypeNewline, Representation= new List<sbyte>{b}});
                i =  (i + 1 );
                continue;
            }
            if ( (b == ';' )) {
                for (;;) {
                    if ( (i >= SliceBuiltins.Length(bytes) )) {
                        break;
                    }
                    if ( (bytes[i] == '\n' )) {
                        break;
                    }
                    i =  (i + 1 );
                }
                continue;
            }
            if ( (b == '#' )) {
                tokens = SliceBuiltins.Append(tokens, new Token{Type= TokenTypeHash, Representation= new List<sbyte>{b}});
                i =  (i + 1 );
                continue;
            }
            if ( (b == '$' )) {
                tokens = SliceBuiltins.Append(tokens, new Token{Type= TokenTypeDollar, Representation= new List<sbyte>{b}});
                i =  (i + 1 );
                continue;
            }
            if ( (b == ':' )) {
                tokens = SliceBuiltins.Append(tokens, new Token{Type= TokenTypeColon, Representation= new List<sbyte>{b}});
                i =  (i + 1 );
                continue;
            }
            if ( (b == ',' )) {
                tokens = SliceBuiltins.Append(tokens, new Token{Type= TokenTypeComma, Representation= new List<sbyte>{b}});
                i =  (i + 1 );
                continue;
            }
            if (IsHexDigit(b)) {
                var repr = new List<sbyte> {};
                for (;;) {
                    if ( (i >= SliceBuiltins.Length(bytes) )) {
                        break;
                    }
                    if ((!IsHexDigit(bytes[i]))) {
                        break;
                    }
                    repr = SliceBuiltins.Append(repr, bytes[i]);
                    i =  (i + 1 );
                }
                tokens = SliceBuiltins.Append(tokens, new Token{Type= TokenTypeNumber, Representation= repr});
                continue;
            }
            if (IsAlpha(b)) {
                var repr = new List<sbyte> {};
                for (;;) {
                    if ( (i >= SliceBuiltins.Length(bytes) )) {
                        break;
                    }
                    if ( ((!IsAlpha(bytes[i])) && (!IsDigit(bytes[i])) )) {
                        break;
                    }
                    repr = SliceBuiltins.Append(repr, bytes[i]);
                    i =  (i + 1 );
                }
                tokens = SliceBuiltins.Append(tokens, new Token{Type= TokenTypeIdentifier, Representation= repr});
                continue;
            }
            i =  (i + 1 );
        }
        return tokens;
    }

    public static List<byte> AssembleLines(List<string> lines)
    {
        var allBytes = new List<sbyte> {};
        var i = 0;
        for (;;) {
            if ( (i >= SliceBuiltins.Length(lines) )) {
                break;
            }
            var lineBytes = StringToBytes(lines[i]);
            allBytes = AppendLineBytes(allBytes, lineBytes);
            if ( (i <  (SliceBuiltins.Length(lines) - 1 ) )) {
                allBytes = SliceBuiltins.Append(allBytes, (sbyte)(10));
            }
            i =  (i + 1 );
        }
        var tokens = TokenizeBytes(allBytes);
        var instructions = Parse(tokens);
        return Assemble(instructions);
    }

}
}
namespace font {


public struct Api {

    public static List<byte> GetFontData()
    {
        return new List<byte> {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x18, 0x18, 0x18, 0x18, 0x18, 0x00, 0x18, 0x00, 0x6C, 0x6C, 0x24, 0x00, 0x00, 0x00, 0x00, 0x00, 0x6C, 0x6C, 0xFE, 0x6C, 0xFE, 0x6C, 0x6C, 0x00, 0x18, 0x3E, 0x60, 0x3C, 0x06, 0x7C, 0x18, 0x00, 0x00, 0xC6, 0xCC, 0x18, 0x30, 0x66, 0xC6, 0x00, 0x38, 0x6C, 0x38, 0x76, 0xDC, 0xCC, 0x76, 0x00, 0x18, 0x18, 0x30, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0C, 0x18, 0x30, 0x30, 0x30, 0x18, 0x0C, 0x00, 0x30, 0x18, 0x0C, 0x0C, 0x0C, 0x18, 0x30, 0x00, 0x00, 0x66, 0x3C, 0xFF, 0x3C, 0x66, 0x00, 0x00, 0x00, 0x18, 0x18, 0x7E, 0x18, 0x18, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x18, 0x18, 0x30, 0x00, 0x00, 0x00, 0x7E, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x18, 0x18, 0x00, 0x06, 0x0C, 0x18, 0x30, 0x60, 0xC0, 0x80, 0x00, 0x7C, 0xC6, 0xCE, 0xD6, 0xE6, 0xC6, 0x7C, 0x00, 0x18, 0x38, 0x18, 0x18, 0x18, 0x18, 0x7E, 0x00, 0x7C, 0xC6, 0x06, 0x1C, 0x30, 0x66, 0xFE, 0x00, 0x7C, 0xC6, 0x06, 0x3C, 0x06, 0xC6, 0x7C, 0x00, 0x1C, 0x3C, 0x6C, 0xCC, 0xFE, 0x0C, 0x1E, 0x00, 0xFE, 0xC0, 0xC0, 0xFC, 0x06, 0xC6, 0x7C, 0x00, 0x38, 0x60, 0xC0, 0xFC, 0xC6, 0xC6, 0x7C, 0x00, 0xFE, 0xC6, 0x0C, 0x18, 0x30, 0x30, 0x30, 0x00, 0x7C, 0xC6, 0xC6, 0x7C, 0xC6, 0xC6, 0x7C, 0x00, 0x7C, 0xC6, 0xC6, 0x7E, 0x06, 0x0C, 0x78, 0x00, 0x00, 0x18, 0x18, 0x00, 0x00, 0x18, 0x18, 0x00, 0x00, 0x18, 0x18, 0x00, 0x00, 0x18, 0x18, 0x30, 0x06, 0x0C, 0x18, 0x30, 0x18, 0x0C, 0x06, 0x00, 0x00, 0x00, 0x7E, 0x00, 0x00, 0x7E, 0x00, 0x00, 0x60, 0x30, 0x18, 0x0C, 0x18, 0x30, 0x60, 0x00, 0x7C, 0xC6, 0x0C, 0x18, 0x18, 0x00, 0x18, 0x00, 0x7C, 0xC6, 0xDE, 0xDE, 0xDE, 0xC0, 0x78, 0x00, 0x38, 0x6C, 0xC6, 0xFE, 0xC6, 0xC6, 0xC6, 0x00, 0xFC, 0x66, 0x66, 0x7C, 0x66, 0x66, 0xFC, 0x00, 0x3C, 0x66, 0xC0, 0xC0, 0xC0, 0x66, 0x3C, 0x00, 0xF8, 0x6C, 0x66, 0x66, 0x66, 0x6C, 0xF8, 0x00, 0xFE, 0x62, 0x68, 0x78, 0x68, 0x62, 0xFE, 0x00, 0xFE, 0x62, 0x68, 0x78, 0x68, 0x60, 0xF0, 0x00, 0x3C, 0x66, 0xC0, 0xC0, 0xCE, 0x66, 0x3A, 0x00, 0xC6, 0xC6, 0xC6, 0xFE, 0xC6, 0xC6, 0xC6, 0x00, 0x3C, 0x18, 0x18, 0x18, 0x18, 0x18, 0x3C, 0x00, 0x1E, 0x0C, 0x0C, 0x0C, 0xCC, 0xCC, 0x78, 0x00, 0xE6, 0x66, 0x6C, 0x78, 0x6C, 0x66, 0xE6, 0x00, 0xF0, 0x60, 0x60, 0x60, 0x62, 0x66, 0xFE, 0x00, 0xC6, 0xEE, 0xFE, 0xFE, 0xD6, 0xC6, 0xC6, 0x00, 0xC6, 0xE6, 0xF6, 0xDE, 0xCE, 0xC6, 0xC6, 0x00, 0x7C, 0xC6, 0xC6, 0xC6, 0xC6, 0xC6, 0x7C, 0x00, 0xFC, 0x66, 0x66, 0x7C, 0x60, 0x60, 0xF0, 0x00, 0x7C, 0xC6, 0xC6, 0xC6, 0xD6, 0xDE, 0x7C, 0x06, 0xFC, 0x66, 0x66, 0x7C, 0x6C, 0x66, 0xE6, 0x00, 0x7C, 0xC6, 0x60, 0x38, 0x0C, 0xC6, 0x7C, 0x00, 0x7E, 0x7E, 0x5A, 0x18, 0x18, 0x18, 0x3C, 0x00, 0xC6, 0xC6, 0xC6, 0xC6, 0xC6, 0xC6, 0x7C, 0x00, 0xC6, 0xC6, 0xC6, 0xC6, 0xC6, 0x6C, 0x38, 0x00, 0xC6, 0xC6, 0xC6, 0xD6, 0xD6, 0xFE, 0x6C, 0x00, 0xC6, 0xC6, 0x6C, 0x38, 0x6C, 0xC6, 0xC6, 0x00, 0x66, 0x66, 0x66, 0x3C, 0x18, 0x18, 0x3C, 0x00, 0xFE, 0xC6, 0x8C, 0x18, 0x32, 0x66, 0xFE, 0x00, 0x3C, 0x30, 0x30, 0x30, 0x30, 0x30, 0x3C, 0x00, 0xC0, 0x60, 0x30, 0x18, 0x0C, 0x06, 0x02, 0x00, 0x3C, 0x0C, 0x0C, 0x0C, 0x0C, 0x0C, 0x3C, 0x00, 0x10, 0x38, 0x6C, 0xC6, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0x30, 0x18, 0x0C, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x78, 0x0C, 0x7C, 0xCC, 0x76, 0x00, 0xE0, 0x60, 0x7C, 0x66, 0x66, 0x66, 0xDC, 0x00, 0x00, 0x00, 0x7C, 0xC6, 0xC0, 0xC6, 0x7C, 0x00, 0x1C, 0x0C, 0x7C, 0xCC, 0xCC, 0xCC, 0x76, 0x00, 0x00, 0x00, 0x7C, 0xC6, 0xFE, 0xC0, 0x7C, 0x00, 0x3C, 0x66, 0x60, 0xF8, 0x60, 0x60, 0xF0, 0x00, 0x00, 0x00, 0x76, 0xCC, 0xCC, 0x7C, 0x0C, 0xF8, 0xE0, 0x60, 0x6C, 0x76, 0x66, 0x66, 0xE6, 0x00, 0x18, 0x00, 0x38, 0x18, 0x18, 0x18, 0x3C, 0x00, 0x06, 0x00, 0x06, 0x06, 0x06, 0x66, 0x66, 0x3C, 0xE0, 0x60, 0x66, 0x6C, 0x78, 0x6C, 0xE6, 0x00, 0x38, 0x18, 0x18, 0x18, 0x18, 0x18, 0x3C, 0x00, 0x00, 0x00, 0xEC, 0xFE, 0xD6, 0xD6, 0xD6, 0x00, 0x00, 0x00, 0xDC, 0x66, 0x66, 0x66, 0x66, 0x00, 0x00, 0x00, 0x7C, 0xC6, 0xC6, 0xC6, 0x7C, 0x00, 0x00, 0x00, 0xDC, 0x66, 0x66, 0x7C, 0x60, 0xF0, 0x00, 0x00, 0x76, 0xCC, 0xCC, 0x7C, 0x0C, 0x1E, 0x00, 0x00, 0xDC, 0x76, 0x60, 0x60, 0xF0, 0x00, 0x00, 0x00, 0x7E, 0xC0, 0x7C, 0x06, 0xFC, 0x00, 0x30, 0x30, 0xFC, 0x30, 0x30, 0x36, 0x1C, 0x00, 0x00, 0x00, 0xCC, 0xCC, 0xCC, 0xCC, 0x76, 0x00, 0x00, 0x00, 0xC6, 0xC6, 0xC6, 0x6C, 0x38, 0x00, 0x00, 0x00, 0xC6, 0xD6, 0xD6, 0xFE, 0x6C, 0x00, 0x00, 0x00, 0xC6, 0x6C, 0x38, 0x6C, 0xC6, 0x00, 0x00, 0x00, 0xC6, 0xC6, 0xC6, 0x7E, 0x06, 0xFC, 0x00, 0x00, 0xFE, 0x8C, 0x18, 0x32, 0xFE, 0x00, 0x0E, 0x18, 0x18, 0x70, 0x18, 0x18, 0x0E, 0x00, 0x18, 0x18, 0x18, 0x18, 0x18, 0x18, 0x18, 0x00, 0x70, 0x18, 0x18, 0x0E, 0x18, 0x18, 0x70, 0x00, 0x76, 0xDC, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF};
    }

    public static List<byte> GetCharBitmap(List<byte> fontData, int charCode)
    {
        var result = new List<byte> {};
        var code = charCode;
        if ( (code < 32 )) {
            code = 32;
        }
        if ( (code > 127 )) {
            code = 127;
        }
        var offset =  ( (code - 32 ) * 8 );
        var i = 0;
        for (;;) {
            if ( (i >= 8 )) {
                break;
            }
            result = SliceBuiltins.Append(result, fontData[ (offset + i )]);
            i =  (i + 1 );
        }
        return result;
    }

    public static bool GetPixel(List<byte> fontData, int charCode, int x, int y)
    {
        var code = charCode;
        if ( (code < 32 )) {
            code = 32;
        }
        if ( (code > 127 )) {
            code = 127;
        }
        var offset =  ( (code - 32 ) * 8 );
        var row = (byte)fontData[ (offset + y )];
        var mask = (byte)(byte)( (0x80 >> x ));
        return  ( (row & mask ) != 0 );
    }

}
}
namespace MainClass {


public struct Api {

    public const int TextCols = 40;

    public const int TextRows = 25;

    public const int TextScreenBase = 0x0400;

    public static string hexDigit(int n)
    {
        if ( (n == 0 )) {
            return (string)@"0";
        } else if ( (n == 1 )) {
            return (string)@"1";
        } else if ( (n == 2 )) {
            return (string)@"2";
        } else if ( (n == 3 )) {
            return (string)@"3";
        } else if ( (n == 4 )) {
            return (string)@"4";
        } else if ( (n == 5 )) {
            return (string)@"5";
        } else if ( (n == 6 )) {
            return (string)@"6";
        } else if ( (n == 7 )) {
            return (string)@"7";
        } else if ( (n == 8 )) {
            return (string)@"8";
        } else if ( (n == 9 )) {
            return (string)@"9";
        } else if ( (n == 10 )) {
            return (string)@"A";
        } else if ( (n == 11 )) {
            return (string)@"B";
        } else if ( (n == 12 )) {
            return (string)@"C";
        } else if ( (n == 13 )) {
            return (string)@"D";
        } else if ( (n == 14 )) {
            return (string)@"E";
        } else if ( (n == 15 )) {
            return (string)@"F";
        }
        return (string)@"0";
    }

    public static string toHex2(int n)
    {
        var high =  ( (n >> 4 ) & 0x0F );
        var low =  (n & 0x0F );
        return (string) (hexDigit(high) + hexDigit(low) );
    }

    public static string toHex4(int n)
    {
        return (string) (toHex2( ( (n >> 8 ) & 0xFF )) + toHex2( (n & 0xFF )) );
    }

    public static string toHex(int n)
    {
        if ( (n > 255 )) {
            return (string) (@"$" + toHex4(n) );
        }
        return (string) (@"$" + toHex2(n) );
    }

    public static List<string> addStringToScreen(List<string> lines, string text, int row, int col)
    {
        var baseAddr =  ( (TextScreenBase +  (row * TextCols ) ) + col );
        var i = 0;
        for (;;) {
            if ( (i >= SliceBuiltins.Length(text) )) {
                break;
            }
            var charCode = (int)(text[i]);
            var addr =  (baseAddr + i );
            lines = SliceBuiltins.Append(lines,  (@"LDA #" + toHex(charCode) ));
            lines = SliceBuiltins.Append(lines,  (@"STA " + toHex(addr) ));
            i =  (i + 1 );
        }
        return lines;
    }

    public static List<string> clearScreen(List<string> lines)
    {
        var addr = TextScreenBase;
        var i = 0;
        for (;;) {
            if ( (i >=  (TextCols * TextRows ) )) {
                break;
            }
            lines = SliceBuiltins.Append(lines, @"LDA #$20");
            lines = SliceBuiltins.Append(lines,  (@"STA " + toHex( (addr + i )) ));
            i =  (i + 1 );
        }
        return lines;
    }

    public static List<byte> createC64WelcomeScreen()
    {
        var lines = new List<string> {};
        lines = clearScreen(lines);
        lines = addStringToScreen(lines, @"**** COMMODORE 64 BASIC V2 ****", 1, 4);
        lines = addStringToScreen(lines, @"64K RAM SYSTEM  38911 BASIC BYTES FREE", 3, 1);
        lines = addStringToScreen(lines, @"READY.", 5, 0);
        lines = SliceBuiltins.Append(lines, @"LDA #$5F");
        lines = SliceBuiltins.Append(lines, @"STA $04F0");
        lines = SliceBuiltins.Append(lines, @"BRK");
        return assembler.Api.AssembleLines(lines);
    }

    public static void Main()
    {
        var scale = (int)(int)(2);
        var windowWidth = (int) ((int)( (TextCols * 8 )) * scale );
        var windowHeight = (int) ((int)( (TextRows * 8 )) * scale );
        var w = graphics.Api.CreateWindow(@"Commodore 64", windowWidth, windowHeight);
        var c = cpu.Api.NewCPU();
        var fontData = font.Api.GetFontData();
        var program = createC64WelcomeScreen();
        c = cpu.Api.LoadProgram(c, program, 0x0600);
        c = cpu.Api.SetPC(c, 0x0600);
        c = cpu.Api.Run(c, 100000);
        var textColor = graphics.Api.NewColor(134, 122, 222, 255);
        var bgColor = graphics.Api.NewColor(64, 50, 133, 255);
        for (;;) {
            bool running = default;
            (w, running) = graphics.Api.PollEvents(w);
            if ((!running)) {
                break;
            }
            graphics.Api.Clear(w, bgColor);
            var charY = 0;
            for (;;) {
                if ( (charY >= TextRows )) {
                    break;
                }
                var charX = 0;
                for (;;) {
                    if ( (charX >= TextCols )) {
                        break;
                    }
                    var memAddr =  ( (TextScreenBase +  (charY * TextCols ) ) + charX );
                    var charCode = (int)(cpu.Api.GetMemory(c, memAddr));
                    if ( (charCode >= 32 )) {
                        if ( (charCode <= 127 )) {
                            var pixelY = 0;
                            for (;;) {
                                if ( (pixelY >= 8 )) {
                                    break;
                                }
                                var pixelX = 0;
                                for (;;) {
                                    if ( (pixelX >= 8 )) {
                                        break;
                                    }
                                    if (font.Api.GetPixel(fontData, charCode, pixelX, pixelY)) {
                                        var screenX = (int) ((int)( ( (charX * 8 ) + pixelX )) * scale );
                                        var screenY = (int) ((int)( ( (charY * 8 ) + pixelY )) * scale );
                                        graphics.Api.FillRect(w, graphics.Api.NewRect(screenX, screenY, scale, scale), textColor);
                                    }
                                    pixelX =  (pixelX + 1 );
                                }
                                pixelY =  (pixelY + 1 );
                            }
                        }
                    }
                    charX =  (charX + 1 );
                }
                charY =  (charY + 1 );
            }
            graphics.Api.Present(w);
        }
        graphics.Api.CloseWindow(w);
    }

}
}

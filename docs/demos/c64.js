// Generated JavaScript code
"use strict";

// Runtime helpers
function len(arr) {
  if (typeof arr === "string") return arr.length;
  if (Array.isArray(arr)) return arr.length;
  return 0;
}

function append(arr, ...items) {
  // Handle nil/undefined slices like Go does
  if (arr == null) arr = [];
  // Clone plain objects to preserve Go's value semantics for structs
  const clonedItems = items.map((item) => {
    if (item && typeof item === "object" && !Array.isArray(item)) {
      return { ...item };
    }
    return item;
  });
  return [...arr, ...clonedItems];
}

function stringFormat(fmt, ...args) {
  let i = 0;
  return fmt.replace(/%[sdvfxc%]/g, (match) => {
    if (match === "%%") return "%";
    if (i >= args.length) return match;
    const arg = args[i++];
    switch (match) {
      case "%s":
        return String(arg);
      case "%d":
        return parseInt(arg, 10);
      case "%f":
        return parseFloat(arg);
      case "%v":
        return String(arg);
      case "%x":
        return parseInt(arg, 10).toString(16);
      case "%c":
        return String.fromCharCode(arg);
      default:
        return arg;
    }
  });
}

// printf - like fmt.Printf (no newline)
function printf(fmt, ...args) {
  const str = stringFormat(fmt, ...args);
  if (typeof process !== "undefined" && process.stdout) {
    process.stdout.write(str);
  } else {
    // Browser fallback - accumulate output
    if (typeof window !== "undefined") {
      window._printBuffer = (window._printBuffer || "") + str;
    }
  }
}

// print - like fmt.Print (no newline)
function print(...args) {
  const str = args.map((a) => String(a)).join(" ");
  if (typeof process !== "undefined" && process.stdout) {
    process.stdout.write(str);
  } else {
    if (typeof window !== "undefined") {
      window._printBuffer = (window._printBuffer || "") + str;
    }
  }
}

function make(type, length, capacity) {
  if (Array.isArray(type)) {
    return new Array(length || 0).fill(type[0] === "number" ? 0 : null);
  }
  return [];
}

// Type conversion functions
// Handle string-to-int conversion for character codes (Go rune semantics)
function int8(v) {
  return typeof v === "string" ? v.charCodeAt(0) | 0 : v | 0;
}
function int16(v) {
  return typeof v === "string" ? v.charCodeAt(0) | 0 : v | 0;
}
function int32(v) {
  return typeof v === "string" ? v.charCodeAt(0) | 0 : v | 0;
}
function int64(v) {
  return typeof v === "string" ? v.charCodeAt(0) : v;
} // BigInt not used for simplicity
function int(v) {
  return typeof v === "string" ? v.charCodeAt(0) | 0 : v | 0;
}
function uint8(v) {
  return typeof v === "string" ? v.charCodeAt(0) & 0xff : (v | 0) & 0xff;
}
function uint16(v) {
  return typeof v === "string" ? v.charCodeAt(0) & 0xffff : (v | 0) & 0xffff;
}
function uint32(v) {
  return typeof v === "string" ? v.charCodeAt(0) >>> 0 : (v | 0) >>> 0;
}
function uint64(v) {
  return typeof v === "string" ? v.charCodeAt(0) : v;
} // BigInt not used for simplicity
function float32(v) {
  return v;
}
function float64(v) {
  return v;
}
function string(v) {
  return String(v);
}
function bool(v) {
  return Boolean(v);
}

// Graphics runtime for Canvas
const graphics = {
  canvas: null,
  ctx: null,
  running: true,
  keys: {},
  lastKey: 0,
  mouseX: 0,
  mouseY: 0,
  mouseDown: false,

  CreateWindow: function (title, width, height) {
    this.canvas = document.createElement("canvas");
    this.canvas.width = width;
    this.canvas.height = height;
    this.ctx = this.canvas.getContext("2d");
    document.body.appendChild(this.canvas);
    document.title = title;

    // Event listeners
    window.addEventListener("keydown", (e) => {
      this.keys[e.key] = true;
      // Store ASCII code for GetLastKey
      if (e.key.length === 1) {
        this.lastKey = e.key.charCodeAt(0);
      } else {
        // Map special keys to ASCII codes
        const specialKeys = {
          Enter: 13,
          Backspace: 8,
          Tab: 9,
          Escape: 27,
          ArrowUp: 38,
          ArrowDown: 40,
          ArrowLeft: 37,
          ArrowRight: 39,
          Delete: 127,
          Space: 32,
        };
        if (specialKeys[e.key]) {
          this.lastKey = specialKeys[e.key];
        }
      }
    });
    window.addEventListener("keyup", (e) => {
      this.keys[e.key] = false;
    });
    this.canvas.addEventListener("mousemove", (e) => {
      const rect = this.canvas.getBoundingClientRect();
      this.mouseX = e.clientX - rect.left;
      this.mouseY = e.clientY - rect.top;
    });
    this.canvas.addEventListener("mousedown", () => {
      this.mouseDown = true;
    });
    this.canvas.addEventListener("mouseup", () => {
      this.mouseDown = false;
    });

    return this.canvas;
  },

  NewColor: function (r, g, b, a) {
    return { r, g, b, a: a !== undefined ? a : 255 };
  },

  // Color helper functions
  Red: function () {
    return { r: 255, g: 0, b: 0, a: 255 };
  },
  Green: function () {
    return { r: 0, g: 255, b: 0, a: 255 };
  },
  Blue: function () {
    return { r: 0, g: 0, b: 255, a: 255 };
  },
  White: function () {
    return { r: 255, g: 255, b: 255, a: 255 };
  },
  Black: function () {
    return { r: 0, g: 0, b: 0, a: 255 };
  },

  Clear: function (canvas, color) {
    this.ctx.fillStyle = `rgba(${color.r}, ${color.g}, ${color.b}, ${color.a / 255})`;
    this.ctx.fillRect(0, 0, canvas.width, canvas.height);
  },

  FillRect: function (canvas, rect, color) {
    this.ctx.fillStyle = `rgba(${color.r}, ${color.g}, ${color.b}, ${color.a / 255})`;
    this.ctx.fillRect(rect.x, rect.y, rect.width, rect.height);
  },

  DrawRect: function (canvas, rect, color) {
    this.ctx.strokeStyle = `rgba(${color.r}, ${color.g}, ${color.b}, ${color.a / 255})`;
    this.ctx.strokeRect(rect.x, rect.y, rect.width, rect.height);
  },

  NewRect: function (x, y, width, height) {
    return { x, y, width, height };
  },

  FillCircle: function (canvas, centerX, centerY, radius, color) {
    this.ctx.fillStyle = `rgba(${color.r}, ${color.g}, ${color.b}, ${color.a / 255})`;
    this.ctx.beginPath();
    this.ctx.arc(centerX, centerY, radius, 0, Math.PI * 2);
    this.ctx.fill();
  },

  DrawCircle: function (canvas, centerX, centerY, radius, color) {
    this.ctx.strokeStyle = `rgba(${color.r}, ${color.g}, ${color.b}, ${color.a / 255})`;
    this.ctx.beginPath();
    this.ctx.arc(centerX, centerY, radius, 0, Math.PI * 2);
    this.ctx.stroke();
  },

  DrawPoint: function (canvas, x, y, color) {
    this.ctx.fillStyle = `rgba(${color.r}, ${color.g}, ${color.b}, ${color.a / 255})`;
    this.ctx.fillRect(x, y, 1, 1);
  },

  DrawLine: function (canvas, x1, y1, x2, y2, color) {
    this.ctx.strokeStyle = `rgba(${color.r}, ${color.g}, ${color.b}, ${color.a / 255})`;
    this.ctx.beginPath();
    this.ctx.moveTo(x1, y1);
    this.ctx.lineTo(x2, y2);
    this.ctx.stroke();
  },

  SetPixel: function (canvas, x, y, color) {
    this.ctx.fillStyle = `rgba(${color.r}, ${color.g}, ${color.b}, ${color.a / 255})`;
    this.ctx.fillRect(x, y, 1, 1);
  },

  PollEvents: function (canvas) {
    return [canvas, this.running];
  },

  Update: function (canvas) {
    // Canvas updates automatically
  },

  KeyDown: function (canvas, key) {
    return this.keys[key] || false;
  },

  GetLastKey: function () {
    const key = this.lastKey;
    this.lastKey = 0; // Clear after reading (like native backends)
    return key;
  },

  GetMousePos: function (canvas) {
    return [this.mouseX, this.mouseY];
  },

  MouseDown: function (canvas) {
    return this.mouseDown;
  },

  Closed: function (canvas) {
    return !this.running;
  },

  Free: function (canvas) {
    if (canvas && canvas.parentNode) {
      canvas.parentNode.removeChild(canvas);
    }
  },

  Present: function (canvas) {
    // Canvas updates automatically, no-op
  },

  CloseWindow: function (canvas) {
    // In browser context, don't immediately close - RunLoop is async
    // The canvas will remain until the page is closed
  },

  RunLoop: function (canvas, frameFunc) {
    const self = this;
    function loop() {
      if (!self.running) return;
      // Poll events (update internal state)
      const result = frameFunc(canvas);
      if (result === false) {
        self.running = false;
        return;
      }
      requestAnimationFrame(loop);
    }
    requestAnimationFrame(loop);
  },
};

const cpu = {
  OpLDAImm: 0xa9,
  OpLDAZp: 0xa5,
  OpLDAZpX: 0xb5,
  OpLDAAbs: 0xad,
  OpLDAAbsX: 0xbd,
  OpLDAAbsY: 0xb9,
  OpLDAIndX: 0xa1,
  OpLDAIndY: 0xb1,
  OpLDXImm: 0xa2,
  OpLDXZp: 0xa6,
  OpLDXZpY: 0xb6,
  OpLDXAbs: 0xae,
  OpLDXAbsY: 0xbe,
  OpLDYImm: 0xa0,
  OpLDYZp: 0xa4,
  OpLDYZpX: 0xb4,
  OpLDYAbs: 0xac,
  OpLDYAbsX: 0xbc,
  OpSTAZp: 0x85,
  OpSTAZpX: 0x95,
  OpSTAAbs: 0x8d,
  OpSTAAbsX: 0x9d,
  OpSTAAbsY: 0x99,
  OpSTAIndX: 0x81,
  OpSTAIndY: 0x91,
  OpSTXZp: 0x86,
  OpSTXZpY: 0x96,
  OpSTXAbs: 0x8e,
  OpSTYZp: 0x84,
  OpSTYZpX: 0x94,
  OpSTYAbs: 0x8c,
  OpADCImm: 0x69,
  OpADCZp: 0x65,
  OpADCZpX: 0x75,
  OpADCAbs: 0x6d,
  OpADCAbsX: 0x7d,
  OpADCAbsY: 0x79,
  OpADCIndX: 0x61,
  OpADCIndY: 0x71,
  OpSBCImm: 0xe9,
  OpSBCZp: 0xe5,
  OpSBCZpX: 0xf5,
  OpSBCAbs: 0xed,
  OpSBCAbsX: 0xfd,
  OpSBCAbsY: 0xf9,
  OpSBCIndX: 0xe1,
  OpSBCIndY: 0xf1,
  OpANDImm: 0x29,
  OpANDZp: 0x25,
  OpANDZpX: 0x35,
  OpANDAbs: 0x2d,
  OpANDAbsX: 0x3d,
  OpANDAbsY: 0x39,
  OpANDIndX: 0x21,
  OpANDIndY: 0x31,
  OpORAImm: 0x09,
  OpORAZp: 0x05,
  OpORAZpX: 0x15,
  OpORAAbs: 0x0d,
  OpORAAbsX: 0x1d,
  OpORAAbsY: 0x19,
  OpORAIndX: 0x01,
  OpORAIndY: 0x11,
  OpEORImm: 0x49,
  OpEORZp: 0x45,
  OpEORZpX: 0x55,
  OpEORAbs: 0x4d,
  OpEORAbsX: 0x5d,
  OpEORAbsY: 0x59,
  OpEORIndX: 0x41,
  OpEORIndY: 0x51,
  OpASLA: 0x0a,
  OpASLZp: 0x06,
  OpASLZpX: 0x16,
  OpASLAbs: 0x0e,
  OpASLAbsX: 0x1e,
  OpLSRA: 0x4a,
  OpLSRZp: 0x46,
  OpLSRZpX: 0x56,
  OpLSRAbs: 0x4e,
  OpLSRAbsX: 0x5e,
  OpROLA: 0x2a,
  OpROLZp: 0x26,
  OpROLZpX: 0x36,
  OpROLAbs: 0x2e,
  OpROLAbsX: 0x3e,
  OpRORA: 0x6a,
  OpRORZp: 0x66,
  OpRORZpX: 0x76,
  OpRORAbs: 0x6e,
  OpRORAbsX: 0x7e,
  OpINX: 0xe8,
  OpINY: 0xc8,
  OpDEX: 0xca,
  OpDEY: 0x88,
  OpINC: 0xe6,
  OpINCZpX: 0xf6,
  OpINCAbs: 0xee,
  OpDECZp: 0xc6,
  OpDECZpX: 0xd6,
  OpDECAbs: 0xce,
  OpCMPImm: 0xc9,
  OpCMPZp: 0xc5,
  OpCMPZpX: 0xd5,
  OpCMPAbs: 0xcd,
  OpCMPAbsX: 0xdd,
  OpCMPAbsY: 0xd9,
  OpCMPIndX: 0xc1,
  OpCMPIndY: 0xd1,
  OpCPXImm: 0xe0,
  OpCPXZp: 0xe4,
  OpCPXAbs: 0xec,
  OpCPYImm: 0xc0,
  OpCPYZp: 0xc4,
  OpCPYAbs: 0xcc,
  OpBPL: 0x10,
  OpBMI: 0x30,
  OpBVC: 0x50,
  OpBVS: 0x70,
  OpBCC: 0x90,
  OpBCS: 0xb0,
  OpBNE: 0xd0,
  OpBEQ: 0xf0,
  OpJMP: 0x4c,
  OpJMPInd: 0x6c,
  OpJSR: 0x20,
  OpRTS: 0x60,
  OpRTI: 0x40,
  OpPHA: 0x48,
  OpPHP: 0x08,
  OpPLA: 0x68,
  OpPLP: 0x28,
  OpTAX: 0xaa,
  OpTXA: 0x8a,
  OpTAY: 0xa8,
  OpTYA: 0x98,
  OpTSX: 0xba,
  OpTXS: 0x9a,
  OpCLC: 0x18,
  OpSEC: 0x38,
  OpCLI: 0x58,
  OpSEI: 0x78,
  OpCLV: 0xb8,
  OpCLD: 0xd8,
  OpSED: 0xf8,
  OpBITZp: 0x24,
  OpBITAbs: 0x2c,
  OpNOP: 0xea,
  OpBRK: 0x00,
  FlagC: 0x01,
  FlagZ: 0x02,
  FlagI: 0x04,
  FlagD: 0x08,
  FlagB: 0x10,
  FlagV: 0x40,
  FlagN: 0x80,
  ScreenBase: 0x0200,
  ScreenWidth: 32,
  ScreenHeight: 32,
  ScreenSize: 1024,

  NewCPU: function () {
    let mem = [];
    let i = 0;
    while (true) {
      if (i >= 65536) {
        break;
      }
      mem = append(mem, uint8(0));
      i = i + 1;
    }
    return {
      A: 0,
      X: 0,
      Y: 0,
      SP: 0xff,
      PC: 0x0600,
      Status: 0x20,
      Memory: mem,
      Halted: false,
      Cycles: 0,
    };
  },
  LoadProgram: function (c, program, addr) {
    let i = 0;
    while (true) {
      if (i >= len(program)) {
        break;
      }
      c.Memory[addr + i] = program[i];
      i = i + 1;
    }
    return c;
  },
  SetPC: function (c, addr) {
    c.PC = addr;
    return c;
  },
  ClearHalted: function (c) {
    c.Halted = false;
    c.Cycles = 0;
    return c;
  },
  ReadByte: function (c, addr) {
    return c.Memory[addr];
  },
  WriteByte: function (c, addr, value) {
    c.Memory[addr] = value;
    return c;
  },
  FetchByte: function (c) {
    let value = c.Memory[c.PC];
    c.PC = c.PC + 1;
    return [c, value];
  },
  FetchWord: function (c) {
    let low = int(c.Memory[c.PC]);
    let high = int(c.Memory[c.PC + 1]);
    c.PC = c.PC + 2;
    return [c, low + high * 256];
  },
  SetZN: function (c, value) {
    if (value == 0) {
      c.Status = c.Status | this.FlagZ;
    } else {
      c.Status = c.Status & (0xff - this.FlagZ);
    }
    if ((value & 0x80) != 0) {
      c.Status = c.Status | this.FlagN;
    } else {
      c.Status = c.Status & (0xff - this.FlagN);
    }
    return c;
  },
  SetCarry: function (c, set) {
    if (set) {
      c.Status = c.Status | this.FlagC;
    } else {
      c.Status = c.Status & (0xff - this.FlagC);
    }
    return c;
  },
  GetCarry: function (c) {
    return (c.Status & this.FlagC) != 0;
  },
  GetZero: function (c) {
    return (c.Status & this.FlagZ) != 0;
  },
  GetNegative: function (c) {
    return (c.Status & this.FlagN) != 0;
  },
  GetOverflow: function (c) {
    return (c.Status & this.FlagV) != 0;
  },
  SetOverflow: function (c, set) {
    if (set) {
      c.Status = c.Status | this.FlagV;
    } else {
      c.Status = c.Status & (0xff - this.FlagV);
    }
    return c;
  },
  ReadIndirectX: function (c, zp) {
    let addr = int(zp + c.X);
    let low = int(c.Memory[addr & 0xff]);
    let high = int(c.Memory[(addr + 1) & 0xff]);
    return c.Memory[low + high * 256];
  },
  ReadIndirectY: function (c, zp) {
    let low = int(c.Memory[int(zp)]);
    let high = int(c.Memory[(int(zp) + 1) & 0xff]);
    let addr = low + high * 256 + int(c.Y);
    return c.Memory[addr];
  },
  GetIndirectXAddr: function (c, zp) {
    let addr = int(zp + c.X);
    let low = int(c.Memory[addr & 0xff]);
    let high = int(c.Memory[(addr + 1) & 0xff]);
    return low + high * 256;
  },
  GetIndirectYAddr: function (c, zp) {
    let low = int(c.Memory[int(zp)]);
    let high = int(c.Memory[(int(zp) + 1) & 0xff]);
    return low + high * 256 + int(c.Y);
  },
  PushByte: function (c, value) {
    c.Memory[0x100 + int(c.SP)] = value;
    c.SP = c.SP - 1;
    return c;
  },
  PullByte: function (c) {
    c.SP = c.SP + 1;
    return [c, c.Memory[0x100 + int(c.SP)]];
  },
  Branch: function (c, offset) {
    if (offset < 128) {
      c.PC = c.PC + int(offset);
    } else {
      c.PC = c.PC - (256 - int(offset));
    }
    return c;
  },
  Step: function (c) {
    if (c.Halted) {
      return c;
    }
    let opcode = 0;
    [c, opcode] = this.FetchByte(c);
    c.Cycles = c.Cycles + 1;
    if (opcode == this.OpLDAImm) {
      let value = 0;
      [c, value] = this.FetchByte(c);
      c.A = value;
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpLDAZp) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      c.A = c.Memory[int(addr)];
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpLDAZpX) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      c.A = c.Memory[int(addr + c.X) & 0xff];
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpLDAAbs) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      c.A = c.Memory[addr];
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpLDAAbsX) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      c.A = c.Memory[addr + int(c.X)];
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpLDAAbsY) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      c.A = c.Memory[addr + int(c.Y)];
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpLDAIndX) {
      let zp = 0;
      [c, zp] = this.FetchByte(c);
      c.A = this.ReadIndirectX(c, zp);
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpLDAIndY) {
      let zp = 0;
      [c, zp] = this.FetchByte(c);
      c.A = this.ReadIndirectY(c, zp);
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpLDXImm) {
      let value = 0;
      [c, value] = this.FetchByte(c);
      c.X = value;
      c = this.SetZN(c, c.X);
    } else if (opcode == this.OpLDXZp) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      c.X = c.Memory[int(addr)];
      c = this.SetZN(c, c.X);
    } else if (opcode == this.OpLDXZpY) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      c.X = c.Memory[int(addr + c.Y) & 0xff];
      c = this.SetZN(c, c.X);
    } else if (opcode == this.OpLDXAbs) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      c.X = c.Memory[addr];
      c = this.SetZN(c, c.X);
    } else if (opcode == this.OpLDXAbsY) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      c.X = c.Memory[addr + int(c.Y)];
      c = this.SetZN(c, c.X);
    } else if (opcode == this.OpLDYImm) {
      let value = 0;
      [c, value] = this.FetchByte(c);
      c.Y = value;
      c = this.SetZN(c, c.Y);
    } else if (opcode == this.OpLDYZp) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      c.Y = c.Memory[int(addr)];
      c = this.SetZN(c, c.Y);
    } else if (opcode == this.OpLDYZpX) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      c.Y = c.Memory[int(addr + c.X) & 0xff];
      c = this.SetZN(c, c.Y);
    } else if (opcode == this.OpLDYAbs) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      c.Y = c.Memory[addr];
      c = this.SetZN(c, c.Y);
    } else if (opcode == this.OpLDYAbsX) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      c.Y = c.Memory[addr + int(c.X)];
      c = this.SetZN(c, c.Y);
    } else if (opcode == this.OpSTAZp) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      c.Memory[int(addr)] = c.A;
    } else if (opcode == this.OpSTAZpX) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      c.Memory[int(addr + c.X) & 0xff] = c.A;
    } else if (opcode == this.OpSTAAbs) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      c.Memory[addr] = c.A;
    } else if (opcode == this.OpSTAAbsX) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      c.Memory[addr + int(c.X)] = c.A;
    } else if (opcode == this.OpSTAAbsY) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      c.Memory[addr + int(c.Y)] = c.A;
    } else if (opcode == this.OpSTAIndX) {
      let zp = 0;
      [c, zp] = this.FetchByte(c);
      let addr = this.GetIndirectXAddr(c, zp);
      c.Memory[addr] = c.A;
    } else if (opcode == this.OpSTAIndY) {
      let zp = 0;
      [c, zp] = this.FetchByte(c);
      let addr = this.GetIndirectYAddr(c, zp);
      c.Memory[addr] = c.A;
    } else if (opcode == this.OpSTXZp) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      c.Memory[int(addr)] = c.X;
    } else if (opcode == this.OpSTXZpY) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      c.Memory[int(addr + c.Y) & 0xff] = c.X;
    } else if (opcode == this.OpSTXAbs) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      c.Memory[addr] = c.X;
    } else if (opcode == this.OpSTYZp) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      c.Memory[int(addr)] = c.Y;
    } else if (opcode == this.OpSTYZpX) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      c.Memory[int(addr + c.X) & 0xff] = c.Y;
    } else if (opcode == this.OpSTYAbs) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      c.Memory[addr] = c.Y;
    } else if (opcode == this.OpADCImm) {
      let value = 0;
      [c, value] = this.FetchByte(c);
      c = this.doADC(c, value);
    } else if (opcode == this.OpADCZp) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      c = this.doADC(c, c.Memory[int(addr)]);
    } else if (opcode == this.OpADCZpX) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      c = this.doADC(c, c.Memory[int(addr + c.X) & 0xff]);
    } else if (opcode == this.OpADCAbs) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      c = this.doADC(c, c.Memory[addr]);
    } else if (opcode == this.OpADCAbsX) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      c = this.doADC(c, c.Memory[addr + int(c.X)]);
    } else if (opcode == this.OpADCAbsY) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      c = this.doADC(c, c.Memory[addr + int(c.Y)]);
    } else if (opcode == this.OpADCIndX) {
      let zp = 0;
      [c, zp] = this.FetchByte(c);
      c = this.doADC(c, this.ReadIndirectX(c, zp));
    } else if (opcode == this.OpADCIndY) {
      let zp = 0;
      [c, zp] = this.FetchByte(c);
      c = this.doADC(c, this.ReadIndirectY(c, zp));
    } else if (opcode == this.OpSBCImm) {
      let value = 0;
      [c, value] = this.FetchByte(c);
      c = this.doSBC(c, value);
    } else if (opcode == this.OpSBCZp) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      c = this.doSBC(c, c.Memory[int(addr)]);
    } else if (opcode == this.OpSBCZpX) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      c = this.doSBC(c, c.Memory[int(addr + c.X) & 0xff]);
    } else if (opcode == this.OpSBCAbs) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      c = this.doSBC(c, c.Memory[addr]);
    } else if (opcode == this.OpSBCAbsX) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      c = this.doSBC(c, c.Memory[addr + int(c.X)]);
    } else if (opcode == this.OpSBCAbsY) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      c = this.doSBC(c, c.Memory[addr + int(c.Y)]);
    } else if (opcode == this.OpSBCIndX) {
      let zp = 0;
      [c, zp] = this.FetchByte(c);
      c = this.doSBC(c, this.ReadIndirectX(c, zp));
    } else if (opcode == this.OpSBCIndY) {
      let zp = 0;
      [c, zp] = this.FetchByte(c);
      c = this.doSBC(c, this.ReadIndirectY(c, zp));
    } else if (opcode == this.OpANDImm) {
      let value = 0;
      [c, value] = this.FetchByte(c);
      c.A = c.A & value;
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpANDZp) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      c.A = c.A & c.Memory[int(addr)];
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpANDZpX) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      c.A = c.A & c.Memory[int(addr + c.X) & 0xff];
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpANDAbs) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      c.A = c.A & c.Memory[addr];
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpANDAbsX) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      c.A = c.A & c.Memory[addr + int(c.X)];
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpANDAbsY) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      c.A = c.A & c.Memory[addr + int(c.Y)];
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpANDIndX) {
      let zp = 0;
      [c, zp] = this.FetchByte(c);
      c.A = c.A & this.ReadIndirectX(c, zp);
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpANDIndY) {
      let zp = 0;
      [c, zp] = this.FetchByte(c);
      c.A = c.A & this.ReadIndirectY(c, zp);
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpORAImm) {
      let value = 0;
      [c, value] = this.FetchByte(c);
      c.A = c.A | value;
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpORAZp) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      c.A = c.A | c.Memory[int(addr)];
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpORAZpX) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      c.A = c.A | c.Memory[int(addr + c.X) & 0xff];
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpORAAbs) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      c.A = c.A | c.Memory[addr];
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpORAAbsX) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      c.A = c.A | c.Memory[addr + int(c.X)];
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpORAAbsY) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      c.A = c.A | c.Memory[addr + int(c.Y)];
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpORAIndX) {
      let zp = 0;
      [c, zp] = this.FetchByte(c);
      c.A = c.A | this.ReadIndirectX(c, zp);
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpORAIndY) {
      let zp = 0;
      [c, zp] = this.FetchByte(c);
      c.A = c.A | this.ReadIndirectY(c, zp);
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpEORImm) {
      let value = 0;
      [c, value] = this.FetchByte(c);
      c.A = c.A ^ value;
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpEORZp) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      c.A = c.A ^ c.Memory[int(addr)];
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpEORZpX) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      c.A = c.A ^ c.Memory[int(addr + c.X) & 0xff];
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpEORAbs) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      c.A = c.A ^ c.Memory[addr];
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpEORAbsX) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      c.A = c.A ^ c.Memory[addr + int(c.X)];
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpEORAbsY) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      c.A = c.A ^ c.Memory[addr + int(c.Y)];
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpEORIndX) {
      let zp = 0;
      [c, zp] = this.FetchByte(c);
      c.A = c.A ^ this.ReadIndirectX(c, zp);
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpEORIndY) {
      let zp = 0;
      [c, zp] = this.FetchByte(c);
      c.A = c.A ^ this.ReadIndirectY(c, zp);
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpASLA) {
      c = this.SetCarry(c, (c.A & 0x80) != 0);
      c.A = c.A << 1;
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpASLZp) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      let val = c.Memory[int(addr)];
      c = this.SetCarry(c, (val & 0x80) != 0);
      val = val << 1;
      c.Memory[int(addr)] = val;
      c = this.SetZN(c, val);
    } else if (opcode == this.OpASLZpX) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      let effAddr = int(addr + c.X) & 0xff;
      let val = c.Memory[effAddr];
      c = this.SetCarry(c, (val & 0x80) != 0);
      val = val << 1;
      c.Memory[effAddr] = val;
      c = this.SetZN(c, val);
    } else if (opcode == this.OpASLAbs) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      let val = c.Memory[addr];
      c = this.SetCarry(c, (val & 0x80) != 0);
      val = val << 1;
      c.Memory[addr] = val;
      c = this.SetZN(c, val);
    } else if (opcode == this.OpASLAbsX) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      let effAddr = addr + int(c.X);
      let val = c.Memory[effAddr];
      c = this.SetCarry(c, (val & 0x80) != 0);
      val = val << 1;
      c.Memory[effAddr] = val;
      c = this.SetZN(c, val);
    } else if (opcode == this.OpLSRA) {
      c = this.SetCarry(c, (c.A & 0x01) != 0);
      c.A = c.A >> 1;
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpLSRZp) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      let val = c.Memory[int(addr)];
      c = this.SetCarry(c, (val & 0x01) != 0);
      val = val >> 1;
      c.Memory[int(addr)] = val;
      c = this.SetZN(c, val);
    } else if (opcode == this.OpLSRZpX) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      let effAddr = int(addr + c.X) & 0xff;
      let val = c.Memory[effAddr];
      c = this.SetCarry(c, (val & 0x01) != 0);
      val = val >> 1;
      c.Memory[effAddr] = val;
      c = this.SetZN(c, val);
    } else if (opcode == this.OpLSRAbs) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      let val = c.Memory[addr];
      c = this.SetCarry(c, (val & 0x01) != 0);
      val = val >> 1;
      c.Memory[addr] = val;
      c = this.SetZN(c, val);
    } else if (opcode == this.OpLSRAbsX) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      let effAddr = addr + int(c.X);
      let val = c.Memory[effAddr];
      c = this.SetCarry(c, (val & 0x01) != 0);
      val = val >> 1;
      c.Memory[effAddr] = val;
      c = this.SetZN(c, val);
    } else if (opcode == this.OpROLA) {
      let carry = 0;
      if (this.GetCarry(c)) {
        carry = 1;
      }
      c = this.SetCarry(c, (c.A & 0x80) != 0);
      c.A = (c.A << 1) | uint8(carry);
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpROLZp) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      let carry = 0;
      if (this.GetCarry(c)) {
        carry = 1;
      }
      let val = c.Memory[int(addr)];
      c = this.SetCarry(c, (val & 0x80) != 0);
      val = (val << 1) | uint8(carry);
      c.Memory[int(addr)] = val;
      c = this.SetZN(c, val);
    } else if (opcode == this.OpROLZpX) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      let carry = 0;
      if (this.GetCarry(c)) {
        carry = 1;
      }
      let effAddr = int(addr + c.X) & 0xff;
      let val = c.Memory[effAddr];
      c = this.SetCarry(c, (val & 0x80) != 0);
      val = (val << 1) | uint8(carry);
      c.Memory[effAddr] = val;
      c = this.SetZN(c, val);
    } else if (opcode == this.OpROLAbs) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      let carry = 0;
      if (this.GetCarry(c)) {
        carry = 1;
      }
      let val = c.Memory[addr];
      c = this.SetCarry(c, (val & 0x80) != 0);
      val = (val << 1) | uint8(carry);
      c.Memory[addr] = val;
      c = this.SetZN(c, val);
    } else if (opcode == this.OpROLAbsX) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      let carry = 0;
      if (this.GetCarry(c)) {
        carry = 1;
      }
      let effAddr = addr + int(c.X);
      let val = c.Memory[effAddr];
      c = this.SetCarry(c, (val & 0x80) != 0);
      val = (val << 1) | uint8(carry);
      c.Memory[effAddr] = val;
      c = this.SetZN(c, val);
    } else if (opcode == this.OpRORA) {
      let carry = 0;
      if (this.GetCarry(c)) {
        carry = 0x80;
      }
      c = this.SetCarry(c, (c.A & 0x01) != 0);
      c.A = (c.A >> 1) | uint8(carry);
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpRORZp) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      let carry = 0;
      if (this.GetCarry(c)) {
        carry = 0x80;
      }
      let val = c.Memory[int(addr)];
      c = this.SetCarry(c, (val & 0x01) != 0);
      val = (val >> 1) | uint8(carry);
      c.Memory[int(addr)] = val;
      c = this.SetZN(c, val);
    } else if (opcode == this.OpRORZpX) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      let carry = 0;
      if (this.GetCarry(c)) {
        carry = 0x80;
      }
      let effAddr = int(addr + c.X) & 0xff;
      let val = c.Memory[effAddr];
      c = this.SetCarry(c, (val & 0x01) != 0);
      val = (val >> 1) | uint8(carry);
      c.Memory[effAddr] = val;
      c = this.SetZN(c, val);
    } else if (opcode == this.OpRORAbs) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      let carry = 0;
      if (this.GetCarry(c)) {
        carry = 0x80;
      }
      let val = c.Memory[addr];
      c = this.SetCarry(c, (val & 0x01) != 0);
      val = (val >> 1) | uint8(carry);
      c.Memory[addr] = val;
      c = this.SetZN(c, val);
    } else if (opcode == this.OpRORAbsX) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      let carry = 0;
      if (this.GetCarry(c)) {
        carry = 0x80;
      }
      let effAddr = addr + int(c.X);
      let val = c.Memory[effAddr];
      c = this.SetCarry(c, (val & 0x01) != 0);
      val = (val >> 1) | uint8(carry);
      c.Memory[effAddr] = val;
      c = this.SetZN(c, val);
    } else if (opcode == this.OpINC) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      let val = c.Memory[int(addr)] + 1;
      c.Memory[int(addr)] = val;
      c = this.SetZN(c, val);
    } else if (opcode == this.OpINCZpX) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      let effAddr = int(addr + c.X) & 0xff;
      let val = c.Memory[effAddr] + 1;
      c.Memory[effAddr] = val;
      c = this.SetZN(c, val);
    } else if (opcode == this.OpINCAbs) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      let val = c.Memory[addr] + 1;
      c.Memory[addr] = val;
      c = this.SetZN(c, val);
    } else if (opcode == this.OpDECZp) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      let val = c.Memory[int(addr)] - 1;
      c.Memory[int(addr)] = val;
      c = this.SetZN(c, val);
    } else if (opcode == this.OpDECZpX) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      let effAddr = int(addr + c.X) & 0xff;
      let val = c.Memory[effAddr] - 1;
      c.Memory[effAddr] = val;
      c = this.SetZN(c, val);
    } else if (opcode == this.OpDECAbs) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      let val = c.Memory[addr] - 1;
      c.Memory[addr] = val;
      c = this.SetZN(c, val);
    } else if (opcode == this.OpINX) {
      c.X = c.X + 1;
      c = this.SetZN(c, c.X);
    } else if (opcode == this.OpINY) {
      c.Y = c.Y + 1;
      c = this.SetZN(c, c.Y);
    } else if (opcode == this.OpDEX) {
      c.X = c.X - 1;
      c = this.SetZN(c, c.X);
    } else if (opcode == this.OpDEY) {
      c.Y = c.Y - 1;
      c = this.SetZN(c, c.Y);
    } else if (opcode == this.OpCMPImm) {
      let value = 0;
      [c, value] = this.FetchByte(c);
      c = this.doCMP(c, c.A, value);
    } else if (opcode == this.OpCMPZp) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      c = this.doCMP(c, c.A, c.Memory[int(addr)]);
    } else if (opcode == this.OpCMPZpX) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      c = this.doCMP(c, c.A, c.Memory[int(addr + c.X) & 0xff]);
    } else if (opcode == this.OpCMPAbs) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      c = this.doCMP(c, c.A, c.Memory[addr]);
    } else if (opcode == this.OpCMPAbsX) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      c = this.doCMP(c, c.A, c.Memory[addr + int(c.X)]);
    } else if (opcode == this.OpCMPAbsY) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      c = this.doCMP(c, c.A, c.Memory[addr + int(c.Y)]);
    } else if (opcode == this.OpCMPIndX) {
      let zp = 0;
      [c, zp] = this.FetchByte(c);
      c = this.doCMP(c, c.A, this.ReadIndirectX(c, zp));
    } else if (opcode == this.OpCMPIndY) {
      let zp = 0;
      [c, zp] = this.FetchByte(c);
      c = this.doCMP(c, c.A, this.ReadIndirectY(c, zp));
    } else if (opcode == this.OpCPXImm) {
      let value = 0;
      [c, value] = this.FetchByte(c);
      c = this.doCMP(c, c.X, value);
    } else if (opcode == this.OpCPXZp) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      c = this.doCMP(c, c.X, c.Memory[int(addr)]);
    } else if (opcode == this.OpCPXAbs) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      c = this.doCMP(c, c.X, c.Memory[addr]);
    } else if (opcode == this.OpCPYImm) {
      let value = 0;
      [c, value] = this.FetchByte(c);
      c = this.doCMP(c, c.Y, value);
    } else if (opcode == this.OpCPYZp) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      c = this.doCMP(c, c.Y, c.Memory[int(addr)]);
    } else if (opcode == this.OpCPYAbs) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      c = this.doCMP(c, c.Y, c.Memory[addr]);
    } else if (opcode == this.OpBITZp) {
      let addr = 0;
      [c, addr] = this.FetchByte(c);
      let val = c.Memory[int(addr)];
      c = this.SetZN(c, uint8(c.A & val));
      c = this.SetOverflow(c, (val & 0x40) != 0);
      if ((val & 0x80) != 0) {
        c.Status = c.Status | this.FlagN;
      }
    } else if (opcode == this.OpBITAbs) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      let val = c.Memory[addr];
      c = this.SetZN(c, uint8(c.A & val));
      c = this.SetOverflow(c, (val & 0x40) != 0);
      if ((val & 0x80) != 0) {
        c.Status = c.Status | this.FlagN;
      }
    } else if (opcode == this.OpBPL) {
      let offset = 0;
      [c, offset] = this.FetchByte(c);
      if (!this.GetNegative(c)) {
        c = this.Branch(c, offset);
      }
    } else if (opcode == this.OpBMI) {
      let offset = 0;
      [c, offset] = this.FetchByte(c);
      if (this.GetNegative(c)) {
        c = this.Branch(c, offset);
      }
    } else if (opcode == this.OpBVC) {
      let offset = 0;
      [c, offset] = this.FetchByte(c);
      if (!this.GetOverflow(c)) {
        c = this.Branch(c, offset);
      }
    } else if (opcode == this.OpBVS) {
      let offset = 0;
      [c, offset] = this.FetchByte(c);
      if (this.GetOverflow(c)) {
        c = this.Branch(c, offset);
      }
    } else if (opcode == this.OpBCC) {
      let offset = 0;
      [c, offset] = this.FetchByte(c);
      if (!this.GetCarry(c)) {
        c = this.Branch(c, offset);
      }
    } else if (opcode == this.OpBCS) {
      let offset = 0;
      [c, offset] = this.FetchByte(c);
      if (this.GetCarry(c)) {
        c = this.Branch(c, offset);
      }
    } else if (opcode == this.OpBNE) {
      let offset = 0;
      [c, offset] = this.FetchByte(c);
      if (!this.GetZero(c)) {
        c = this.Branch(c, offset);
      }
    } else if (opcode == this.OpBEQ) {
      let offset = 0;
      [c, offset] = this.FetchByte(c);
      if (this.GetZero(c)) {
        c = this.Branch(c, offset);
      }
    } else if (opcode == this.OpJMP) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      c.PC = addr;
    } else if (opcode == this.OpJMPInd) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      let low = int(c.Memory[addr]);
      let high = int(c.Memory[(addr & 0xff00) | ((addr + 1) & 0xff)]);
      c.PC = low + high * 256;
    } else if (opcode == this.OpJSR) {
      let addr = 0;
      [c, addr] = this.FetchWord(c);
      let retAddr = c.PC - 1;
      c = this.PushByte(c, uint8((retAddr >> 8) & 0xff));
      c = this.PushByte(c, uint8(retAddr & 0xff));
      c.PC = addr;
    } else if (opcode == this.OpRTS) {
      let low = 0;
      let high = 0;
      [c, low] = this.PullByte(c);
      [c, high] = this.PullByte(c);
      c.PC = int(low) + int(high) * 256 + 1;
    } else if (opcode == this.OpRTI) {
      let status = 0;
      [c, status] = this.PullByte(c);
      c.Status = status | 0x20;
      let low = 0;
      let high = 0;
      [c, low] = this.PullByte(c);
      [c, high] = this.PullByte(c);
      c.PC = int(low) + int(high) * 256;
    } else if (opcode == this.OpPHA) {
      c = this.PushByte(c, c.A);
    } else if (opcode == this.OpPHP) {
      c = this.PushByte(c, uint8(c.Status | this.FlagB | 0x20));
    } else if (opcode == this.OpPLA) {
      [c, c.A] = this.PullByte(c);
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpPLP) {
      let status = 0;
      [c, status] = this.PullByte(c);
      c.Status = uint8((status | 0x20) & 0xef);
    } else if (opcode == this.OpTAX) {
      c.X = c.A;
      c = this.SetZN(c, c.X);
    } else if (opcode == this.OpTXA) {
      c.A = c.X;
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpTAY) {
      c.Y = c.A;
      c = this.SetZN(c, c.Y);
    } else if (opcode == this.OpTYA) {
      c.A = c.Y;
      c = this.SetZN(c, c.A);
    } else if (opcode == this.OpTSX) {
      c.X = c.SP;
      c = this.SetZN(c, c.X);
    } else if (opcode == this.OpTXS) {
      c.SP = c.X;
    } else if (opcode == this.OpCLC) {
      c = this.SetCarry(c, false);
    } else if (opcode == this.OpSEC) {
      c = this.SetCarry(c, true);
    } else if (opcode == this.OpCLI) {
      c.Status = c.Status & (0xff - this.FlagI);
    } else if (opcode == this.OpSEI) {
      c.Status = c.Status | this.FlagI;
    } else if (opcode == this.OpCLV) {
      c = this.SetOverflow(c, false);
    } else if (opcode == this.OpCLD) {
      c.Status = c.Status & (0xff - this.FlagD);
    } else if (opcode == this.OpSED) {
      c.Status = c.Status | this.FlagD;
    } else if (opcode == this.OpNOP) {
    } else if (opcode == this.OpBRK) {
      c.Halted = true;
    }
    return c;
  },
  doADC: function (c, value) {
    let carry = 0;
    if (this.GetCarry(c)) {
      carry = 1;
    }
    let result = int(c.A) + int(value) + carry;
    let overflow =
      ((c.A ^ value) & 0x80) == 0 && ((c.A ^ uint8(result)) & 0x80) != 0;
    c = this.SetOverflow(c, overflow);
    c = this.SetCarry(c, result > 255);
    c.A = uint8(result & 0xff);
    c = this.SetZN(c, c.A);
    return c;
  },
  doSBC: function (c, value) {
    let carry = 0;
    if (this.GetCarry(c)) {
      carry = 1;
    }
    let result = int(c.A) - int(value) - (1 - carry);
    let overflow =
      ((c.A ^ value) & 0x80) != 0 && ((c.A ^ uint8(result)) & 0x80) != 0;
    c = this.SetOverflow(c, overflow);
    c = this.SetCarry(c, result >= 0);
    c.A = uint8(result & 0xff);
    c = this.SetZN(c, c.A);
    return c;
  },
  doCMP: function (c, reg, value) {
    let result = int(reg) - int(value);
    c = this.SetCarry(c, reg >= value);
    c = this.SetZN(c, uint8(result & 0xff));
    return c;
  },
  Run: function (c, maxCycles) {
    while (true) {
      if (c.Halted) {
        break;
      }
      if (c.Cycles >= maxCycles) {
        break;
      }
      c = this.Step(c);
    }
    return c;
  },
  GetScreenPixel: function (c, x, y) {
    if (x < 0) {
      return 0;
    }
    if (x >= this.ScreenWidth) {
      return 0;
    }
    if (y < 0) {
      return 0;
    }
    if (y >= this.ScreenHeight) {
      return 0;
    }
    let addr = this.ScreenBase + y * this.ScreenWidth + x;
    return c.Memory[addr];
  },
  IsHalted: function (c) {
    return c.Halted;
  },
  GetMemory: function (c, addr) {
    if (addr < 0) {
      return 0;
    }
    if (addr >= 65536) {
      return 0;
    }
    return c.Memory[addr];
  },
};

const assembler = {
  TokenTypeInstruction: 1,
  TokenTypeNumber: 2,
  TokenTypeLabel: 3,
  TokenTypeComma: 4,
  TokenTypeNewline: 5,
  TokenTypeHash: 6,
  TokenTypeDollar: 7,
  TokenTypeColon: 8,
  TokenTypeIdentifier: 9,
  TokenTypeComment: 10,
  ModeImplied: 0,
  ModeImmediate: 1,
  ModeZeroPage: 2,
  ModeAbsolute: 3,
  ModeZeroPageX: 4,
  ModeZeroPageY: 5,
  ModeAbsoluteX: 6,
  ModeAbsoluteY: 7,
  ModeIndirectX: 8,
  ModeIndirectY: 9,
  ModeAccumulator: 10,
  ModeIndirect: 11,

  IsDigit: function (b) {
    return b >= 48 && b <= 57;
  },
  IsHexDigit: function (b) {
    return this.IsDigit(b) || (b >= 97 && b <= 102) || (b >= 65 && b <= 70);
  },
  IsAlpha: function (b) {
    return (b >= 97 && b <= 122) || (b >= 65 && b <= 90) || b == 95;
  },
  IsWhitespace: function (b) {
    return b == 32 || b == 9;
  },
  StringToBytes: function (s) {
    let result = [];
    let i = 0;
    while (true) {
      if (i >= len(s)) {
        break;
      }
      result = append(result, int8(s.charCodeAt(i)));
      i = i + 1;
    }
    return result;
  },
  ToUpper: function (b) {
    if (b >= 97 && b <= 122) {
      return b - 32;
    }
    return b;
  },
  Tokenize: function (text) {
    let tokens = [];
    let bytes = this.StringToBytes(text);
    let i = 0;
    while (true) {
      if (i >= len(bytes)) {
        break;
      }
      let b = bytes[i];
      if (this.IsWhitespace(b)) {
        i = i + 1;
        continue;
      }
      if (b == 10) {
        tokens = append(tokens, {
          Type: this.TokenTypeNewline,
          Representation: [b],
        });
        i = i + 1;
        continue;
      }
      if (b == 59) {
        while (true) {
          if (i >= len(bytes)) {
            break;
          }
          if (bytes[i] == 10) {
            break;
          }
          i = i + 1;
        }
        continue;
      }
      if (b == 35) {
        tokens = append(tokens, {
          Type: this.TokenTypeHash,
          Representation: [b],
        });
        i = i + 1;
        continue;
      }
      if (b == 36) {
        tokens = append(tokens, {
          Type: this.TokenTypeDollar,
          Representation: [b],
        });
        i = i + 1;
        continue;
      }
      if (b == 58) {
        tokens = append(tokens, {
          Type: this.TokenTypeColon,
          Representation: [b],
        });
        i = i + 1;
        continue;
      }
      if (b == 44) {
        tokens = append(tokens, {
          Type: this.TokenTypeComma,
          Representation: [b],
        });
        i = i + 1;
        continue;
      }
      if (this.IsDigit(b)) {
        let repr = [];
        while (true) {
          if (i >= len(bytes)) {
            break;
          }
          if (!this.IsHexDigit(bytes[i])) {
            break;
          }
          repr = append(repr, bytes[i]);
          i = i + 1;
        }
        tokens = append(tokens, {
          Type: this.TokenTypeNumber,
          Representation: repr,
        });
        continue;
      }
      if (this.IsAlpha(b)) {
        let repr = [];
        while (true) {
          if (i >= len(bytes)) {
            break;
          }
          if (!this.IsAlpha(bytes[i]) && !this.IsDigit(bytes[i])) {
            break;
          }
          repr = append(repr, bytes[i]);
          i = i + 1;
        }
        tokens = append(tokens, {
          Type: this.TokenTypeIdentifier,
          Representation: repr,
        });
        continue;
      }
      i = i + 1;
    }
    return tokens;
  },
  ParseHex: function (bytes) {
    let result = 0;
    let i = 0;
    while (true) {
      if (i >= len(bytes)) {
        break;
      }
      let b = bytes[i];
      result = result * 16;
      if (b >= 48 && b <= 57) {
        result = result + int(b - 48);
      } else if (b >= 97 && b <= 102) {
        result = result + int(b - 97 + 10);
      } else if (b >= 65 && b <= 70) {
        result = result + int(b - 65 + 10);
      }
      i = i + 1;
    }
    return result;
  },
  ParseDecimal: function (bytes) {
    let result = 0;
    let i = 0;
    while (true) {
      if (i >= len(bytes)) {
        break;
      }
      let b = bytes[i];
      result = result * 10 + int(b - 48);
      i = i + 1;
    }
    return result;
  },
  MatchToken: function (token, s) {
    if (len(token.Representation) != len(s)) {
      return false;
    }
    let i = 0;
    while (true) {
      if (i >= len(s)) {
        break;
      }
      if (
        this.ToUpper(token.Representation[i]) !=
        this.ToUpper(int8(s.charCodeAt(i)))
      ) {
        return false;
      }
      i = i + 1;
    }
    return true;
  },
  CopyBytes: function (src) {
    let dst = [];
    let i = 0;
    while (true) {
      if (i >= len(src)) {
        break;
      }
      dst = append(dst, src[i]);
      i = i + 1;
    }
    return dst;
  },
  Parse: function (tokens) {
    let instructions = [];
    let i = 0;
    while (true) {
      if (i >= len(tokens)) {
        break;
      }
      if (tokens[i].Type == this.TokenTypeNewline) {
        i = i + 1;
        continue;
      }
      let currentLabelBytes = [];
      let hasLabel = false;
      if (
        tokens[i].Type == this.TokenTypeIdentifier &&
        i + 1 < len(tokens) &&
        tokens[i + 1].Type == this.TokenTypeColon
      ) {
        currentLabelBytes = this.CopyBytes(tokens[i].Representation);
        hasLabel = true;
        i = i + 2;
        while (true) {
          if (i >= len(tokens)) {
            break;
          }
          if (tokens[i].Type != this.TokenTypeNewline) {
            break;
          }
          i = i + 1;
        }
        if (i >= len(tokens)) {
          break;
        }
      }
      if (tokens[i].Type != this.TokenTypeIdentifier) {
        i = i + 1;
        continue;
      }
      let instr = {
        OpcodeBytes: this.CopyBytes(tokens[i].Representation),
        Mode: this.ModeImplied,
        Operand: 0,
        LabelBytes: currentLabelBytes,
        HasLabel: hasLabel,
      };
      i = i + 1;
      if (i < len(tokens) && tokens[i].Type != this.TokenTypeNewline) {
        if (tokens[i].Type == this.TokenTypeHash) {
          i = i + 1;
          instr.Mode = this.ModeImmediate;
          if (i < len(tokens) && tokens[i].Type == this.TokenTypeDollar) {
            i = i + 1;
            if (i < len(tokens) && tokens[i].Type == this.TokenTypeNumber) {
              instr.Operand = this.ParseHex(tokens[i].Representation);
              i = i + 1;
            }
          } else if (
            i < len(tokens) &&
            tokens[i].Type == this.TokenTypeNumber
          ) {
            instr.Operand = this.ParseDecimal(tokens[i].Representation);
            i = i + 1;
          }
        } else if (tokens[i].Type == this.TokenTypeDollar) {
          i = i + 1;
          if (i < len(tokens) && tokens[i].Type == this.TokenTypeNumber) {
            instr.Operand = this.ParseHex(tokens[i].Representation);
            if (len(tokens[i].Representation) <= 2) {
              instr.Mode = this.ModeZeroPage;
            } else {
              instr.Mode = this.ModeAbsolute;
            }
            i = i + 1;
            if (i < len(tokens) && tokens[i].Type == this.TokenTypeComma) {
              i = i + 1;
              if (
                i < len(tokens) &&
                tokens[i].Type == this.TokenTypeIdentifier
              ) {
                if (this.MatchToken(tokens[i], "X")) {
                  if (instr.Mode == this.ModeZeroPage) {
                    instr.Mode = this.ModeZeroPageX;
                  } else {
                    instr.Mode = this.ModeAbsoluteX;
                  }
                } else if (this.MatchToken(tokens[i], "Y")) {
                  if (instr.Mode == this.ModeZeroPage) {
                    instr.Mode = this.ModeZeroPageY;
                  } else {
                    instr.Mode = this.ModeAbsoluteY;
                  }
                }
                i = i + 1;
              }
            }
          }
        } else if (
          tokens[i].Type == this.TokenTypeIdentifier &&
          this.MatchToken(tokens[i], "A")
        ) {
          instr.Mode = this.ModeAccumulator;
          i = i + 1;
        } else if (tokens[i].Type == this.TokenTypeNumber) {
          instr.Operand = this.ParseDecimal(tokens[i].Representation);
          if (instr.Operand <= 255) {
            instr.Mode = this.ModeZeroPage;
          } else {
            instr.Mode = this.ModeAbsolute;
          }
          i = i + 1;
        }
      }
      instructions = append(instructions, instr);
    }
    return instructions;
  },
  IsOpcode: function (opcodeBytes, name) {
    if (len(opcodeBytes) != len(name)) {
      return false;
    }
    let i = 0;
    while (true) {
      if (i >= len(name)) {
        break;
      }
      let ob = opcodeBytes[i];
      if (ob >= 97 && ob <= 122) {
        ob = ob - 32;
      }
      let nb = int8(name.charCodeAt(i));
      if (ob != nb) {
        return false;
      }
      i = i + 1;
    }
    return true;
  },
  Assemble: function (instructions) {
    let code = [];
    let idx = 0;
    while (true) {
      if (idx >= len(instructions)) {
        break;
      }
      let instr = instructions[idx];
      let opcodeBytes = instr.OpcodeBytes;
      if (this.IsOpcode(opcodeBytes, "LDA")) {
        if (instr.Mode == this.ModeImmediate) {
          code = append(code, uint8(cpu.OpLDAImm));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeZeroPage) {
          code = append(code, uint8(cpu.OpLDAZp));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeZeroPageX) {
          code = append(code, uint8(cpu.OpLDAZpX));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeAbsolute) {
          code = append(code, uint8(cpu.OpLDAAbs));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        } else if (instr.Mode == this.ModeAbsoluteX) {
          code = append(code, uint8(cpu.OpLDAAbsX));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        } else if (instr.Mode == this.ModeAbsoluteY) {
          code = append(code, uint8(cpu.OpLDAAbsY));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        }
      } else if (this.IsOpcode(opcodeBytes, "LDX")) {
        if (instr.Mode == this.ModeImmediate) {
          code = append(code, uint8(cpu.OpLDXImm));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeZeroPage) {
          code = append(code, uint8(cpu.OpLDXZp));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeZeroPageY) {
          code = append(code, uint8(cpu.OpLDXZpY));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeAbsolute) {
          code = append(code, uint8(cpu.OpLDXAbs));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        } else if (instr.Mode == this.ModeAbsoluteY) {
          code = append(code, uint8(cpu.OpLDXAbsY));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        }
      } else if (this.IsOpcode(opcodeBytes, "LDY")) {
        if (instr.Mode == this.ModeImmediate) {
          code = append(code, uint8(cpu.OpLDYImm));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeZeroPage) {
          code = append(code, uint8(cpu.OpLDYZp));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeZeroPageX) {
          code = append(code, uint8(cpu.OpLDYZpX));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeAbsolute) {
          code = append(code, uint8(cpu.OpLDYAbs));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        } else if (instr.Mode == this.ModeAbsoluteX) {
          code = append(code, uint8(cpu.OpLDYAbsX));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        }
      } else if (this.IsOpcode(opcodeBytes, "STA")) {
        if (instr.Mode == this.ModeZeroPage) {
          code = append(code, uint8(cpu.OpSTAZp));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeZeroPageX) {
          code = append(code, uint8(cpu.OpSTAZpX));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeAbsolute) {
          code = append(code, uint8(cpu.OpSTAAbs));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        } else if (instr.Mode == this.ModeAbsoluteX) {
          code = append(code, uint8(cpu.OpSTAAbsX));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        } else if (instr.Mode == this.ModeAbsoluteY) {
          code = append(code, uint8(cpu.OpSTAAbsY));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        }
      } else if (this.IsOpcode(opcodeBytes, "STX")) {
        if (instr.Mode == this.ModeZeroPage) {
          code = append(code, uint8(cpu.OpSTXZp));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeZeroPageY) {
          code = append(code, uint8(cpu.OpSTXZpY));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeAbsolute) {
          code = append(code, uint8(cpu.OpSTXAbs));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        }
      } else if (this.IsOpcode(opcodeBytes, "STY")) {
        if (instr.Mode == this.ModeZeroPage) {
          code = append(code, uint8(cpu.OpSTYZp));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeZeroPageX) {
          code = append(code, uint8(cpu.OpSTYZpX));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeAbsolute) {
          code = append(code, uint8(cpu.OpSTYAbs));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        }
      } else if (this.IsOpcode(opcodeBytes, "ADC")) {
        if (instr.Mode == this.ModeImmediate) {
          code = append(code, uint8(cpu.OpADCImm));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeZeroPage) {
          code = append(code, uint8(cpu.OpADCZp));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeZeroPageX) {
          code = append(code, uint8(cpu.OpADCZpX));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeAbsolute) {
          code = append(code, uint8(cpu.OpADCAbs));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        } else if (instr.Mode == this.ModeAbsoluteX) {
          code = append(code, uint8(cpu.OpADCAbsX));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        } else if (instr.Mode == this.ModeAbsoluteY) {
          code = append(code, uint8(cpu.OpADCAbsY));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        }
      } else if (this.IsOpcode(opcodeBytes, "SBC")) {
        if (instr.Mode == this.ModeImmediate) {
          code = append(code, uint8(cpu.OpSBCImm));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeZeroPage) {
          code = append(code, uint8(cpu.OpSBCZp));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeZeroPageX) {
          code = append(code, uint8(cpu.OpSBCZpX));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeAbsolute) {
          code = append(code, uint8(cpu.OpSBCAbs));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        } else if (instr.Mode == this.ModeAbsoluteX) {
          code = append(code, uint8(cpu.OpSBCAbsX));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        } else if (instr.Mode == this.ModeAbsoluteY) {
          code = append(code, uint8(cpu.OpSBCAbsY));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        }
      } else if (this.IsOpcode(opcodeBytes, "AND")) {
        if (instr.Mode == this.ModeImmediate) {
          code = append(code, uint8(cpu.OpANDImm));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeZeroPage) {
          code = append(code, uint8(cpu.OpANDZp));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeZeroPageX) {
          code = append(code, uint8(cpu.OpANDZpX));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeAbsolute) {
          code = append(code, uint8(cpu.OpANDAbs));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        } else if (instr.Mode == this.ModeAbsoluteX) {
          code = append(code, uint8(cpu.OpANDAbsX));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        } else if (instr.Mode == this.ModeAbsoluteY) {
          code = append(code, uint8(cpu.OpANDAbsY));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        }
      } else if (this.IsOpcode(opcodeBytes, "ORA")) {
        if (instr.Mode == this.ModeImmediate) {
          code = append(code, uint8(cpu.OpORAImm));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeZeroPage) {
          code = append(code, uint8(cpu.OpORAZp));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeZeroPageX) {
          code = append(code, uint8(cpu.OpORAZpX));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeAbsolute) {
          code = append(code, uint8(cpu.OpORAAbs));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        } else if (instr.Mode == this.ModeAbsoluteX) {
          code = append(code, uint8(cpu.OpORAAbsX));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        } else if (instr.Mode == this.ModeAbsoluteY) {
          code = append(code, uint8(cpu.OpORAAbsY));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        }
      } else if (this.IsOpcode(opcodeBytes, "EOR")) {
        if (instr.Mode == this.ModeImmediate) {
          code = append(code, uint8(cpu.OpEORImm));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeZeroPage) {
          code = append(code, uint8(cpu.OpEORZp));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeZeroPageX) {
          code = append(code, uint8(cpu.OpEORZpX));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeAbsolute) {
          code = append(code, uint8(cpu.OpEORAbs));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        } else if (instr.Mode == this.ModeAbsoluteX) {
          code = append(code, uint8(cpu.OpEORAbsX));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        } else if (instr.Mode == this.ModeAbsoluteY) {
          code = append(code, uint8(cpu.OpEORAbsY));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        }
      } else if (this.IsOpcode(opcodeBytes, "ASL")) {
        if (
          instr.Mode == this.ModeAccumulator ||
          instr.Mode == this.ModeImplied
        ) {
          code = append(code, uint8(cpu.OpASLA));
        } else if (instr.Mode == this.ModeZeroPage) {
          code = append(code, uint8(cpu.OpASLZp));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeZeroPageX) {
          code = append(code, uint8(cpu.OpASLZpX));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeAbsolute) {
          code = append(code, uint8(cpu.OpASLAbs));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        } else if (instr.Mode == this.ModeAbsoluteX) {
          code = append(code, uint8(cpu.OpASLAbsX));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        }
      } else if (this.IsOpcode(opcodeBytes, "LSR")) {
        if (
          instr.Mode == this.ModeAccumulator ||
          instr.Mode == this.ModeImplied
        ) {
          code = append(code, uint8(cpu.OpLSRA));
        } else if (instr.Mode == this.ModeZeroPage) {
          code = append(code, uint8(cpu.OpLSRZp));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeZeroPageX) {
          code = append(code, uint8(cpu.OpLSRZpX));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeAbsolute) {
          code = append(code, uint8(cpu.OpLSRAbs));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        } else if (instr.Mode == this.ModeAbsoluteX) {
          code = append(code, uint8(cpu.OpLSRAbsX));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        }
      } else if (this.IsOpcode(opcodeBytes, "ROL")) {
        if (
          instr.Mode == this.ModeAccumulator ||
          instr.Mode == this.ModeImplied
        ) {
          code = append(code, uint8(cpu.OpROLA));
        } else if (instr.Mode == this.ModeZeroPage) {
          code = append(code, uint8(cpu.OpROLZp));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeZeroPageX) {
          code = append(code, uint8(cpu.OpROLZpX));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeAbsolute) {
          code = append(code, uint8(cpu.OpROLAbs));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        } else if (instr.Mode == this.ModeAbsoluteX) {
          code = append(code, uint8(cpu.OpROLAbsX));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        }
      } else if (this.IsOpcode(opcodeBytes, "ROR")) {
        if (
          instr.Mode == this.ModeAccumulator ||
          instr.Mode == this.ModeImplied
        ) {
          code = append(code, uint8(cpu.OpRORA));
        } else if (instr.Mode == this.ModeZeroPage) {
          code = append(code, uint8(cpu.OpRORZp));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeZeroPageX) {
          code = append(code, uint8(cpu.OpRORZpX));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeAbsolute) {
          code = append(code, uint8(cpu.OpRORAbs));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        } else if (instr.Mode == this.ModeAbsoluteX) {
          code = append(code, uint8(cpu.OpRORAbsX));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        }
      } else if (this.IsOpcode(opcodeBytes, "INC")) {
        if (instr.Mode == this.ModeZeroPage) {
          code = append(code, uint8(cpu.OpINC));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeZeroPageX) {
          code = append(code, uint8(cpu.OpINCZpX));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeAbsolute) {
          code = append(code, uint8(cpu.OpINCAbs));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        }
      } else if (this.IsOpcode(opcodeBytes, "DEC")) {
        if (instr.Mode == this.ModeZeroPage) {
          code = append(code, uint8(cpu.OpDECZp));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeZeroPageX) {
          code = append(code, uint8(cpu.OpDECZpX));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeAbsolute) {
          code = append(code, uint8(cpu.OpDECAbs));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        }
      } else if (this.IsOpcode(opcodeBytes, "INX")) {
        code = append(code, uint8(cpu.OpINX));
      } else if (this.IsOpcode(opcodeBytes, "INY")) {
        code = append(code, uint8(cpu.OpINY));
      } else if (this.IsOpcode(opcodeBytes, "DEX")) {
        code = append(code, uint8(cpu.OpDEX));
      } else if (this.IsOpcode(opcodeBytes, "DEY")) {
        code = append(code, uint8(cpu.OpDEY));
      } else if (this.IsOpcode(opcodeBytes, "CMP")) {
        if (instr.Mode == this.ModeImmediate) {
          code = append(code, uint8(cpu.OpCMPImm));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeZeroPage) {
          code = append(code, uint8(cpu.OpCMPZp));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeZeroPageX) {
          code = append(code, uint8(cpu.OpCMPZpX));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeAbsolute) {
          code = append(code, uint8(cpu.OpCMPAbs));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        } else if (instr.Mode == this.ModeAbsoluteX) {
          code = append(code, uint8(cpu.OpCMPAbsX));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        } else if (instr.Mode == this.ModeAbsoluteY) {
          code = append(code, uint8(cpu.OpCMPAbsY));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        }
      } else if (this.IsOpcode(opcodeBytes, "CPX")) {
        if (instr.Mode == this.ModeImmediate) {
          code = append(code, uint8(cpu.OpCPXImm));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeZeroPage) {
          code = append(code, uint8(cpu.OpCPXZp));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeAbsolute) {
          code = append(code, uint8(cpu.OpCPXAbs));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        }
      } else if (this.IsOpcode(opcodeBytes, "CPY")) {
        if (instr.Mode == this.ModeImmediate) {
          code = append(code, uint8(cpu.OpCPYImm));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeZeroPage) {
          code = append(code, uint8(cpu.OpCPYZp));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeAbsolute) {
          code = append(code, uint8(cpu.OpCPYAbs));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        }
      } else if (this.IsOpcode(opcodeBytes, "BIT")) {
        if (instr.Mode == this.ModeZeroPage) {
          code = append(code, uint8(cpu.OpBITZp));
          code = append(code, uint8(instr.Operand));
        } else if (instr.Mode == this.ModeAbsolute) {
          code = append(code, uint8(cpu.OpBITAbs));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        }
      } else if (this.IsOpcode(opcodeBytes, "BPL")) {
        code = append(code, uint8(cpu.OpBPL));
        code = append(code, uint8(instr.Operand));
      } else if (this.IsOpcode(opcodeBytes, "BMI")) {
        code = append(code, uint8(cpu.OpBMI));
        code = append(code, uint8(instr.Operand));
      } else if (this.IsOpcode(opcodeBytes, "BVC")) {
        code = append(code, uint8(cpu.OpBVC));
        code = append(code, uint8(instr.Operand));
      } else if (this.IsOpcode(opcodeBytes, "BVS")) {
        code = append(code, uint8(cpu.OpBVS));
        code = append(code, uint8(instr.Operand));
      } else if (this.IsOpcode(opcodeBytes, "BCC")) {
        code = append(code, uint8(cpu.OpBCC));
        code = append(code, uint8(instr.Operand));
      } else if (this.IsOpcode(opcodeBytes, "BCS")) {
        code = append(code, uint8(cpu.OpBCS));
        code = append(code, uint8(instr.Operand));
      } else if (this.IsOpcode(opcodeBytes, "BNE")) {
        code = append(code, uint8(cpu.OpBNE));
        code = append(code, uint8(instr.Operand));
      } else if (this.IsOpcode(opcodeBytes, "BEQ")) {
        code = append(code, uint8(cpu.OpBEQ));
        code = append(code, uint8(instr.Operand));
      } else if (this.IsOpcode(opcodeBytes, "JMP")) {
        if (instr.Mode == this.ModeIndirect) {
          code = append(code, uint8(cpu.OpJMPInd));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        } else {
          code = append(code, uint8(cpu.OpJMP));
          code = append(code, uint8(instr.Operand & 0xff));
          code = append(code, uint8((instr.Operand >> 8) & 0xff));
        }
      } else if (this.IsOpcode(opcodeBytes, "JSR")) {
        code = append(code, uint8(cpu.OpJSR));
        code = append(code, uint8(instr.Operand & 0xff));
        code = append(code, uint8((instr.Operand >> 8) & 0xff));
      } else if (this.IsOpcode(opcodeBytes, "RTS")) {
        code = append(code, uint8(cpu.OpRTS));
      } else if (this.IsOpcode(opcodeBytes, "RTI")) {
        code = append(code, uint8(cpu.OpRTI));
      } else if (this.IsOpcode(opcodeBytes, "PHA")) {
        code = append(code, uint8(cpu.OpPHA));
      } else if (this.IsOpcode(opcodeBytes, "PHP")) {
        code = append(code, uint8(cpu.OpPHP));
      } else if (this.IsOpcode(opcodeBytes, "PLA")) {
        code = append(code, uint8(cpu.OpPLA));
      } else if (this.IsOpcode(opcodeBytes, "PLP")) {
        code = append(code, uint8(cpu.OpPLP));
      } else if (this.IsOpcode(opcodeBytes, "TAX")) {
        code = append(code, uint8(cpu.OpTAX));
      } else if (this.IsOpcode(opcodeBytes, "TXA")) {
        code = append(code, uint8(cpu.OpTXA));
      } else if (this.IsOpcode(opcodeBytes, "TAY")) {
        code = append(code, uint8(cpu.OpTAY));
      } else if (this.IsOpcode(opcodeBytes, "TYA")) {
        code = append(code, uint8(cpu.OpTYA));
      } else if (this.IsOpcode(opcodeBytes, "TSX")) {
        code = append(code, uint8(cpu.OpTSX));
      } else if (this.IsOpcode(opcodeBytes, "TXS")) {
        code = append(code, uint8(cpu.OpTXS));
      } else if (this.IsOpcode(opcodeBytes, "CLC")) {
        code = append(code, uint8(cpu.OpCLC));
      } else if (this.IsOpcode(opcodeBytes, "SEC")) {
        code = append(code, uint8(cpu.OpSEC));
      } else if (this.IsOpcode(opcodeBytes, "CLI")) {
        code = append(code, uint8(cpu.OpCLI));
      } else if (this.IsOpcode(opcodeBytes, "SEI")) {
        code = append(code, uint8(cpu.OpSEI));
      } else if (this.IsOpcode(opcodeBytes, "CLV")) {
        code = append(code, uint8(cpu.OpCLV));
      } else if (this.IsOpcode(opcodeBytes, "CLD")) {
        code = append(code, uint8(cpu.OpCLD));
      } else if (this.IsOpcode(opcodeBytes, "SED")) {
        code = append(code, uint8(cpu.OpSED));
      } else if (this.IsOpcode(opcodeBytes, "NOP")) {
        code = append(code, uint8(cpu.OpNOP));
      } else if (this.IsOpcode(opcodeBytes, "BRK")) {
        code = append(code, uint8(cpu.OpBRK));
      }
      idx = idx + 1;
    }
    return code;
  },
  AssembleString: function (text) {
    let tokens = this.Tokenize(text);
    let instructions = this.Parse(tokens);
    return this.Assemble(instructions);
  },
  AppendLineBytes: function (allBytes, lineBytes) {
    let j = 0;
    while (true) {
      if (j >= len(lineBytes)) {
        break;
      }
      allBytes = append(allBytes, lineBytes[j]);
      j = j + 1;
    }
    return allBytes;
  },
  TokenizeBytes: function (bytes) {
    let tokens = [];
    let i = 0;
    while (true) {
      if (i >= len(bytes)) {
        break;
      }
      let b = bytes[i];
      if (this.IsWhitespace(b)) {
        i = i + 1;
        continue;
      }
      if (b == 10) {
        tokens = append(tokens, {
          Type: this.TokenTypeNewline,
          Representation: [b],
        });
        i = i + 1;
        continue;
      }
      if (b == 59) {
        while (true) {
          if (i >= len(bytes)) {
            break;
          }
          if (bytes[i] == 10) {
            break;
          }
          i = i + 1;
        }
        continue;
      }
      if (b == 35) {
        tokens = append(tokens, {
          Type: this.TokenTypeHash,
          Representation: [b],
        });
        i = i + 1;
        continue;
      }
      if (b == 36) {
        tokens = append(tokens, {
          Type: this.TokenTypeDollar,
          Representation: [b],
        });
        i = i + 1;
        continue;
      }
      if (b == 58) {
        tokens = append(tokens, {
          Type: this.TokenTypeColon,
          Representation: [b],
        });
        i = i + 1;
        continue;
      }
      if (b == 44) {
        tokens = append(tokens, {
          Type: this.TokenTypeComma,
          Representation: [b],
        });
        i = i + 1;
        continue;
      }
      if (this.IsDigit(b)) {
        let repr = [];
        while (true) {
          if (i >= len(bytes)) {
            break;
          }
          if (!this.IsHexDigit(bytes[i])) {
            break;
          }
          repr = append(repr, bytes[i]);
          i = i + 1;
        }
        tokens = append(tokens, {
          Type: this.TokenTypeNumber,
          Representation: repr,
        });
        continue;
      }
      if (this.IsAlpha(b)) {
        let repr = [];
        while (true) {
          if (i >= len(bytes)) {
            break;
          }
          if (!this.IsAlpha(bytes[i]) && !this.IsDigit(bytes[i])) {
            break;
          }
          repr = append(repr, bytes[i]);
          i = i + 1;
        }
        tokens = append(tokens, {
          Type: this.TokenTypeIdentifier,
          Representation: repr,
        });
        continue;
      }
      i = i + 1;
    }
    return tokens;
  },
  AssembleLines: function (lines) {
    let allBytes = [];
    let i = 0;
    while (true) {
      if (i >= len(lines)) {
        break;
      }
      let lineBytes = this.StringToBytes(lines[i]);
      allBytes = this.AppendLineBytes(allBytes, lineBytes);
      if (i < len(lines) - 1) {
        allBytes = append(allBytes, int8(10));
      }
      i = i + 1;
    }
    let tokens = this.TokenizeBytes(allBytes);
    let instructions = this.Parse(tokens);
    return this.Assemble(instructions);
  },
  AssembleLinesWithCount: function (lines) {
    let allBytes = [];
    let i = 0;
    while (true) {
      if (i >= len(lines)) {
        break;
      }
      let lineBytes = this.StringToBytes(lines[i]);
      allBytes = this.AppendLineBytes(allBytes, lineBytes);
      if (i < len(lines) - 1) {
        allBytes = append(allBytes, int8(10));
      }
      i = i + 1;
    }
    let tokens = this.TokenizeBytes(allBytes);
    let instructions = this.Parse(tokens);
    return [this.Assemble(instructions), len(instructions)];
  },
  GetLastInstrFirstByte: function (lines) {
    let allBytes = [];
    let i = 0;
    while (true) {
      if (i >= len(lines)) {
        break;
      }
      let lineBytes = this.StringToBytes(lines[i]);
      allBytes = this.AppendLineBytes(allBytes, lineBytes);
      if (i < len(lines) - 1) {
        allBytes = append(allBytes, int8(10));
      }
      i = i + 1;
    }
    let tokens = this.TokenizeBytes(allBytes);
    let instructions = this.Parse(tokens);
    if (len(instructions) > 0) {
      let lastInstr = instructions[len(instructions) - 1];
      if (len(lastInstr.OpcodeBytes) > 0) {
        return int(lastInstr.OpcodeBytes[0]);
      }
    }
    return -1;
  },
};

const basic = {
  TextCols: 40,
  TextRows: 25,
  ScreenBase: 0x0400,
  CodeBase: 0xc000,

  NewBasicState: function () {
    return { Lines: [], CursorRow: 0, CursorCol: 0 };
  },
  SetCursor: function (state, row, col) {
    state.CursorRow = row;
    state.CursorCol = col;
    return state;
  },
  GetCursorAddr: function (state) {
    return this.ScreenBase + state.CursorRow * this.TextCols + state.CursorCol;
  },
  StoreLine: function (state, lineNum, text) {
    let newLines = [];
    let found = false;
    let i = 0;
    while (true) {
      if (i >= len(state.Lines)) {
        break;
      }
      if (state.Lines[i].LineNum == lineNum) {
        newLines = append(newLines, { LineNum: lineNum, Text: text });
        found = true;
      } else {
        newLines = append(newLines, state.Lines[i]);
      }
      i = i + 1;
    }
    if (!found) {
      newLines = append(newLines, { LineNum: lineNum, Text: text });
    }
    state.Lines = newLines;
    state = this.sortLines(state);
    return state;
  },
  sortLines: function (state) {
    let n = len(state.Lines);
    let i = 0;
    while (true) {
      if (i >= n - 1) {
        break;
      }
      let j = 0;
      while (true) {
        if (j >= n - i - 1) {
          break;
        }
        if (state.Lines[j].LineNum > state.Lines[j + 1].LineNum) {
          let temp = state.Lines[j];
          state.Lines[j] = state.Lines[j + 1];
          state.Lines[j + 1] = temp;
        }
        j = j + 1;
      }
      i = i + 1;
    }
    return state;
  },
  DeleteLine: function (state, lineNum) {
    let newLines = [];
    let i = 0;
    while (true) {
      if (i >= len(state.Lines)) {
        break;
      }
      if (state.Lines[i].LineNum != lineNum) {
        newLines = append(newLines, state.Lines[i]);
      }
      i = i + 1;
    }
    state.Lines = newLines;
    return state;
  },
  ClearProgram: function (state) {
    state.Lines = [];
    return state;
  },
  CompileImmediate: function (state, line) {
    let asmLines = this.compileLine(line, state.CursorRow, state.CursorCol);
    asmLines = append(asmLines, "BRK");
    return assembler.AssembleLines(asmLines);
  },
  CompileProgram: function (state) {
    let asmLines = [];
    let row = state.CursorRow;
    let col = 0;
    let i = 0;
    while (true) {
      if (i >= len(state.Lines)) {
        break;
      }
      let lineAsm = this.compileLine(state.Lines[i].Text, row, col);
      let j = 0;
      while (true) {
        if (j >= len(lineAsm)) {
          break;
        }
        asmLines = append(asmLines, lineAsm[j]);
        j = j + 1;
      }
      row = row + 1;
      if (row >= this.TextRows) {
        row = this.TextRows - 1;
      }
      i = i + 1;
    }
    asmLines = append(asmLines, "BRK");
    return assembler.AssembleLines(asmLines);
  },
  CompileProgramDebug: function (state) {
    let asmLines = [];
    let row = state.CursorRow;
    let col = 0;
    let i = 0;
    while (true) {
      if (i >= len(state.Lines)) {
        break;
      }
      let lineAsm = this.compileLine(state.Lines[i].Text, row, col);
      let j = 0;
      while (true) {
        if (j >= len(lineAsm)) {
          break;
        }
        asmLines = append(asmLines, lineAsm[j]);
        j = j + 1;
      }
      row = row + 1;
      if (row >= this.TextRows) {
        row = this.TextRows - 1;
      }
      i = i + 1;
    }
    asmLines = append(asmLines, "BRK");
    let [code, instrCount] = assembler.AssembleLinesWithCount(asmLines);
    let lastByte = assembler.GetLastInstrFirstByte(asmLines);
    return [code, len(asmLines), instrCount, lastByte];
  },
  compileLine: function (line, cursorRow, cursorCol) {
    let [cmd, args] = this.parseLine(line);
    if (cmd == "PRINT") {
      return this.genPrint(args, cursorRow, cursorCol);
    } else if (cmd == "POKE") {
      let [addr, value] = this.parsePoke(args);
      return this.genPoke(addr, value);
    } else if (cmd == "CLR") {
      return this.genClear();
    }
    return [];
  },
  GetLineCount: function (state) {
    return len(state.Lines);
  },
  GetLine: function (state, index) {
    if (index >= 0 && index < len(state.Lines)) {
      return state.Lines[index];
    }
    return { LineNum: 0, Text: "" };
  },
  hexDigit: function (n) {
    if (n == 0) {
      return "0";
    } else if (n == 1) {
      return "1";
    } else if (n == 2) {
      return "2";
    } else if (n == 3) {
      return "3";
    } else if (n == 4) {
      return "4";
    } else if (n == 5) {
      return "5";
    } else if (n == 6) {
      return "6";
    } else if (n == 7) {
      return "7";
    } else if (n == 8) {
      return "8";
    } else if (n == 9) {
      return "9";
    } else if (n == 10) {
      return "A";
    } else if (n == 11) {
      return "B";
    } else if (n == 12) {
      return "C";
    } else if (n == 13) {
      return "D";
    } else if (n == 14) {
      return "E";
    } else if (n == 15) {
      return "F";
    }
    return "0";
  },
  toHex2: function (n) {
    let high = (n >> 4) & 0x0f;
    let low = n & 0x0f;
    return this.hexDigit(high) + this.hexDigit(low);
  },
  toHex4: function (n) {
    return this.toHex2((n >> 8) & 0xff) + this.toHex2(n & 0xff);
  },
  toHex: function (n) {
    if (n > 255) {
      return "$" + this.toHex4(n);
    }
    return "$" + this.toHex2(n);
  },
  genPrint: function (args, cursorRow, cursorCol) {
    let lines = [];
    let text = this.parseString(args);
    let baseAddr = this.ScreenBase + cursorRow * this.TextCols + cursorCol;
    let i = 0;
    while (true) {
      if (i >= len(text)) {
        break;
      }
      let charCode = int(text.charCodeAt(i));
      let addr = baseAddr + i;
      lines = append(lines, "LDA #" + this.toHex(charCode));
      lines = append(lines, "STA " + this.toHex(addr));
      i = i + 1;
    }
    return lines;
  },
  genPoke: function (addr, value) {
    let lines = [];
    if (value > 255) {
      value = value & 0xff;
    }
    if (value < 0) {
      value = 0;
    }
    lines = append(lines, "LDA #" + this.toHex(value));
    lines = append(lines, "STA " + this.toHex(addr));
    return lines;
  },
  genClear: function () {
    let lines = [];
    lines = append(lines, "LDA #$20");
    let i = 0;
    while (true) {
      if (i >= this.TextCols * this.TextRows) {
        break;
      }
      let addr = this.ScreenBase + i;
      lines = append(lines, "STA " + this.toHex(addr));
      i = i + 1;
    }
    return lines;
  },
  genList: function (state, startRow) {
    let lines = [];
    let row = startRow;
    let i = 0;
    while (true) {
      if (i >= len(state.Lines)) {
        break;
      }
      if (row >= this.TextRows) {
        break;
      }
      let lineNum = state.Lines[i].LineNum;
      let text = state.Lines[i].Text;
      let numStr = this.intToString(lineNum);
      let fullLine = numStr + " " + text;
      let baseAddr = this.ScreenBase + row * this.TextCols;
      let j = 0;
      while (true) {
        if (j >= len(fullLine)) {
          break;
        }
        if (j >= this.TextCols) {
          break;
        }
        let charCode = int(fullLine.charCodeAt(j));
        let addr = baseAddr + j;
        lines = append(lines, "LDA #" + this.toHex(charCode));
        lines = append(lines, "STA " + this.toHex(addr));
        j = j + 1;
      }
      row = row + 1;
      i = i + 1;
    }
    return lines;
  },
  intToString: function (n) {
    if (n == 0) {
      return "0";
    }
    let neg = false;
    if (n < 0) {
      neg = true;
      n = -n;
    }
    let digits = "";
    while (true) {
      if (n == 0) {
        break;
      }
      let digit = n % 10;
      digits = this.digitToChar(digit) + digits;
      n = (n / 10) | 0;
    }
    if (neg) {
      digits = "-" + digits;
    }
    return digits;
  },
  digitToChar: function (d) {
    if (d == 0) {
      return "0";
    } else if (d == 1) {
      return "1";
    } else if (d == 2) {
      return "2";
    } else if (d == 3) {
      return "3";
    } else if (d == 4) {
      return "4";
    } else if (d == 5) {
      return "5";
    } else if (d == 6) {
      return "6";
    } else if (d == 7) {
      return "7";
    } else if (d == 8) {
      return "8";
    } else if (d == 9) {
      return "9";
    }
    return "0";
  },
  parseLine: function (line) {
    let pos = 0;
    while (true) {
      if (pos >= len(line)) {
        break;
      }
      if (line.charCodeAt(pos) != 32 && line.charCodeAt(pos) != 9) {
        break;
      }
      pos = pos + 1;
    }
    let cmd = "";
    while (true) {
      if (pos >= len(line)) {
        break;
      }
      let ch = int(line.charCodeAt(pos));
      if (!this.isLetterCode(ch)) {
        break;
      }
      cmd = cmd + this.toUpperChar(ch);
      pos = pos + 1;
    }
    while (true) {
      if (pos >= len(line)) {
        break;
      }
      if (line.charCodeAt(pos) != 32 && line.charCodeAt(pos) != 9) {
        break;
      }
      pos = pos + 1;
    }
    let args = "";
    while (true) {
      if (pos >= len(line)) {
        break;
      }
      args = args + this.charToString(int(line.charCodeAt(pos)));
      pos = pos + 1;
    }
    return [cmd, args];
  },
  parseLineNumber: function (line) {
    let pos = 0;
    while (true) {
      if (pos >= len(line)) {
        break;
      }
      if (line.charCodeAt(pos) != 32 && line.charCodeAt(pos) != 9) {
        break;
      }
      pos = pos + 1;
    }
    if (pos >= len(line) || !this.isDigitCode(int(line.charCodeAt(pos)))) {
      return [0, line, false];
    }
    let lineNum = 0;
    while (true) {
      if (pos >= len(line)) {
        break;
      }
      let ch = int(line.charCodeAt(pos));
      if (!this.isDigitCode(ch)) {
        break;
      }
      lineNum = lineNum * 10 + (ch - int(48));
      pos = pos + 1;
    }
    while (true) {
      if (pos >= len(line)) {
        break;
      }
      if (line.charCodeAt(pos) != 32 && line.charCodeAt(pos) != 9) {
        break;
      }
      pos = pos + 1;
    }
    let rest = "";
    while (true) {
      if (pos >= len(line)) {
        break;
      }
      rest = rest + this.charToString(int(line.charCodeAt(pos)));
      pos = pos + 1;
    }
    return [lineNum, rest, true];
  },
  parsePoke: function (args) {
    let addr = 0;
    let i = 0;
    while (true) {
      if (i >= len(args)) {
        break;
      }
      if (args.charCodeAt(i) != 32 && args.charCodeAt(i) != 9) {
        break;
      }
      i = i + 1;
    }
    while (true) {
      if (i >= len(args)) {
        break;
      }
      let ch = int(args.charCodeAt(i));
      if (ch == int(44)) {
        break;
      }
      if (ch == int(32) || ch == int(9)) {
        i = i + 1;
        continue;
      }
      if (this.isDigitCode(ch)) {
        addr = addr * 10 + (ch - int(48));
      }
      i = i + 1;
    }
    if (i < len(args) && args.charCodeAt(i) == 44) {
      i = i + 1;
    }
    while (true) {
      if (i >= len(args)) {
        break;
      }
      if (args.charCodeAt(i) != 32 && args.charCodeAt(i) != 9) {
        break;
      }
      i = i + 1;
    }
    let value = 0;
    while (true) {
      if (i >= len(args)) {
        break;
      }
      let ch = int(args.charCodeAt(i));
      if (ch == int(32) || ch == int(9)) {
        break;
      }
      if (this.isDigitCode(ch)) {
        value = value * 10 + (ch - int(48));
      }
      i = i + 1;
    }
    return [addr, value];
  },
  parseString: function (args) {
    let start = -1;
    let i = 0;
    while (true) {
      if (i >= len(args)) {
        break;
      }
      if (args.charCodeAt(i) == 34) {
        start = i + 1;
        break;
      }
      i = i + 1;
    }
    if (start < 0) {
      return this.trimSpacesStr(args);
    }
    let result = "";
    let j = start;
    while (true) {
      if (j >= len(args)) {
        break;
      }
      if (args.charCodeAt(j) == 34) {
        break;
      }
      result = result + this.charToString(int(args.charCodeAt(j)));
      j = j + 1;
    }
    return result;
  },
  charToString: function (ch) {
    if (ch >= 32 && ch <= 126) {
      if (ch == 32) {
        return " ";
      } else if (ch == 33) {
        return "!";
      } else if (ch == 34) {
        return '"';
      } else if (ch == 35) {
        return "#";
      } else if (ch == 36) {
        return "$";
      } else if (ch == 37) {
        return "%";
      } else if (ch == 38) {
        return "&";
      } else if (ch == 39) {
        return "'";
      } else if (ch == 40) {
        return "(";
      } else if (ch == 41) {
        return ")";
      } else if (ch == 42) {
        return "*";
      } else if (ch == 43) {
        return "+";
      } else if (ch == 44) {
        return ",";
      } else if (ch == 45) {
        return "-";
      } else if (ch == 46) {
        return ".";
      } else if (ch == 47) {
        return "/";
      } else if (ch == 48) {
        return "0";
      } else if (ch == 49) {
        return "1";
      } else if (ch == 50) {
        return "2";
      } else if (ch == 51) {
        return "3";
      } else if (ch == 52) {
        return "4";
      } else if (ch == 53) {
        return "5";
      } else if (ch == 54) {
        return "6";
      } else if (ch == 55) {
        return "7";
      } else if (ch == 56) {
        return "8";
      } else if (ch == 57) {
        return "9";
      } else if (ch == 58) {
        return ":";
      } else if (ch == 59) {
        return ";";
      } else if (ch == 60) {
        return "<";
      } else if (ch == 61) {
        return "=";
      } else if (ch == 62) {
        return ">";
      } else if (ch == 63) {
        return "?";
      } else if (ch == 64) {
        return "@";
      } else if (ch == 65) {
        return "A";
      } else if (ch == 66) {
        return "B";
      } else if (ch == 67) {
        return "C";
      } else if (ch == 68) {
        return "D";
      } else if (ch == 69) {
        return "E";
      } else if (ch == 70) {
        return "F";
      } else if (ch == 71) {
        return "G";
      } else if (ch == 72) {
        return "H";
      } else if (ch == 73) {
        return "I";
      } else if (ch == 74) {
        return "J";
      } else if (ch == 75) {
        return "K";
      } else if (ch == 76) {
        return "L";
      } else if (ch == 77) {
        return "M";
      } else if (ch == 78) {
        return "N";
      } else if (ch == 79) {
        return "O";
      } else if (ch == 80) {
        return "P";
      } else if (ch == 81) {
        return "Q";
      } else if (ch == 82) {
        return "R";
      } else if (ch == 83) {
        return "S";
      } else if (ch == 84) {
        return "T";
      } else if (ch == 85) {
        return "U";
      } else if (ch == 86) {
        return "V";
      } else if (ch == 87) {
        return "W";
      } else if (ch == 88) {
        return "X";
      } else if (ch == 89) {
        return "Y";
      } else if (ch == 90) {
        return "Z";
      } else if (ch == 91) {
        return "[";
      } else if (ch == 92) {
        return "";
      } else if (ch == 93) {
        return "]";
      } else if (ch == 94) {
        return "^";
      } else if (ch == 95) {
        return "_";
      } else if (ch == 96) {
        return "`";
      } else if (ch == 97) {
        return "a";
      } else if (ch == 98) {
        return "b";
      } else if (ch == 99) {
        return "c";
      } else if (ch == 100) {
        return "d";
      } else if (ch == 101) {
        return "e";
      } else if (ch == 102) {
        return "f";
      } else if (ch == 103) {
        return "g";
      } else if (ch == 104) {
        return "h";
      } else if (ch == 105) {
        return "i";
      } else if (ch == 106) {
        return "j";
      } else if (ch == 107) {
        return "k";
      } else if (ch == 108) {
        return "l";
      } else if (ch == 109) {
        return "m";
      } else if (ch == 110) {
        return "n";
      } else if (ch == 111) {
        return "o";
      } else if (ch == 112) {
        return "p";
      } else if (ch == 113) {
        return "q";
      } else if (ch == 114) {
        return "r";
      } else if (ch == 115) {
        return "s";
      } else if (ch == 116) {
        return "t";
      } else if (ch == 117) {
        return "u";
      } else if (ch == 118) {
        return "v";
      } else if (ch == 119) {
        return "w";
      } else if (ch == 120) {
        return "x";
      } else if (ch == 121) {
        return "y";
      } else if (ch == 122) {
        return "z";
      } else if (ch == 123) {
        return "{";
      } else if (ch == 124) {
        return "|";
      } else if (ch == 125) {
        return "}";
      } else if (ch == 126) {
        return "~";
      }
    }
    return "";
  },
  isLetterCode: function (ch) {
    return (
      (ch >= int(97) && ch <= int(122)) || (ch >= int(65) && ch <= int(90))
    );
  },
  isDigitCode: function (ch) {
    return ch >= int(48) && ch <= int(57);
  },
  toUpperChar: function (ch) {
    if (ch >= int(97) && ch <= int(122)) {
      ch = ch - 32;
    }
    if (ch == 65) {
      return "A";
    } else if (ch == 66) {
      return "B";
    } else if (ch == 67) {
      return "C";
    } else if (ch == 68) {
      return "D";
    } else if (ch == 69) {
      return "E";
    } else if (ch == 70) {
      return "F";
    } else if (ch == 71) {
      return "G";
    } else if (ch == 72) {
      return "H";
    } else if (ch == 73) {
      return "I";
    } else if (ch == 74) {
      return "J";
    } else if (ch == 75) {
      return "K";
    } else if (ch == 76) {
      return "L";
    } else if (ch == 77) {
      return "M";
    } else if (ch == 78) {
      return "N";
    } else if (ch == 79) {
      return "O";
    } else if (ch == 80) {
      return "P";
    } else if (ch == 81) {
      return "Q";
    } else if (ch == 82) {
      return "R";
    } else if (ch == 83) {
      return "S";
    } else if (ch == 84) {
      return "T";
    } else if (ch == 85) {
      return "U";
    } else if (ch == 86) {
      return "V";
    } else if (ch == 87) {
      return "W";
    } else if (ch == 88) {
      return "X";
    } else if (ch == 89) {
      return "Y";
    } else if (ch == 90) {
      return "Z";
    }
    return "";
  },
  toUpper: function (s) {
    let result = "";
    let i = 0;
    while (true) {
      if (i >= len(s)) {
        break;
      }
      let ch = int(s.charCodeAt(i));
      result = result + this.toUpperChar(ch);
      i = i + 1;
    }
    return result;
  },
  parseNumber: function (s) {
    let result = 0;
    let i = 0;
    while (true) {
      if (i >= len(s)) {
        break;
      }
      let ch = s.charCodeAt(i);
      if (ch >= 48 && ch <= 57) {
        result = result * 10 + int(ch - 48);
      }
      i = i + 1;
    }
    return result;
  },
  trimSpacesStr: function (s) {
    let start = 0;
    while (true) {
      if (start >= len(s)) {
        break;
      }
      if (s.charCodeAt(start) != 32 && s.charCodeAt(start) != 9) {
        break;
      }
      start = start + 1;
    }
    let end = len(s);
    while (true) {
      if (end <= start) {
        break;
      }
      if (s.charCodeAt(end - 1) != 32 && s.charCodeAt(end - 1) != 9) {
        break;
      }
      end = end - 1;
    }
    if (start >= end) {
      return "";
    }
    let result = "";
    let i = start;
    while (true) {
      if (i >= end) {
        break;
      }
      result = result + this.charToString(int(s.charCodeAt(i)));
      i = i + 1;
    }
    return result;
  },
  HasLineNumber: function (line) {
    let [lineNum, rest, found] = this.parseLineNumber(line);
    if (lineNum > 0 && len(rest) >= 0) {
    }
    return found;
  },
  ExtractLineNumber: function (line) {
    let [lineNum, rest, found] = this.parseLineNumber(line);
    if (!found) {
      return [0, line];
    }
    return [lineNum, rest];
  },
  IsCommand: function (line, cmdName) {
    let [cmd, args] = this.parseLine(line);
    if (len(args) >= 0) {
    }
    return cmd == this.toUpper(cmdName);
  },
};

const font = {
  GetFontData: function () {
    return [
      0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x18, 0x18, 0x18, 0x18,
      0x18, 0x00, 0x18, 0x00, 0x6c, 0x6c, 0x24, 0x00, 0x00, 0x00, 0x00, 0x00,
      0x6c, 0x6c, 0xfe, 0x6c, 0xfe, 0x6c, 0x6c, 0x00, 0x18, 0x3e, 0x60, 0x3c,
      0x06, 0x7c, 0x18, 0x00, 0x00, 0xc6, 0xcc, 0x18, 0x30, 0x66, 0xc6, 0x00,
      0x38, 0x6c, 0x38, 0x76, 0xdc, 0xcc, 0x76, 0x00, 0x18, 0x18, 0x30, 0x00,
      0x00, 0x00, 0x00, 0x00, 0x0c, 0x18, 0x30, 0x30, 0x30, 0x18, 0x0c, 0x00,
      0x30, 0x18, 0x0c, 0x0c, 0x0c, 0x18, 0x30, 0x00, 0x00, 0x66, 0x3c, 0xff,
      0x3c, 0x66, 0x00, 0x00, 0x00, 0x18, 0x18, 0x7e, 0x18, 0x18, 0x00, 0x00,
      0x00, 0x00, 0x00, 0x00, 0x00, 0x18, 0x18, 0x30, 0x00, 0x00, 0x00, 0x7e,
      0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x18, 0x18, 0x00,
      0x06, 0x0c, 0x18, 0x30, 0x60, 0xc0, 0x80, 0x00, 0x7c, 0xc6, 0xce, 0xd6,
      0xe6, 0xc6, 0x7c, 0x00, 0x18, 0x38, 0x18, 0x18, 0x18, 0x18, 0x7e, 0x00,
      0x7c, 0xc6, 0x06, 0x1c, 0x30, 0x66, 0xfe, 0x00, 0x7c, 0xc6, 0x06, 0x3c,
      0x06, 0xc6, 0x7c, 0x00, 0x1c, 0x3c, 0x6c, 0xcc, 0xfe, 0x0c, 0x1e, 0x00,
      0xfe, 0xc0, 0xc0, 0xfc, 0x06, 0xc6, 0x7c, 0x00, 0x38, 0x60, 0xc0, 0xfc,
      0xc6, 0xc6, 0x7c, 0x00, 0xfe, 0xc6, 0x0c, 0x18, 0x30, 0x30, 0x30, 0x00,
      0x7c, 0xc6, 0xc6, 0x7c, 0xc6, 0xc6, 0x7c, 0x00, 0x7c, 0xc6, 0xc6, 0x7e,
      0x06, 0x0c, 0x78, 0x00, 0x00, 0x18, 0x18, 0x00, 0x00, 0x18, 0x18, 0x00,
      0x00, 0x18, 0x18, 0x00, 0x00, 0x18, 0x18, 0x30, 0x06, 0x0c, 0x18, 0x30,
      0x18, 0x0c, 0x06, 0x00, 0x00, 0x00, 0x7e, 0x00, 0x00, 0x7e, 0x00, 0x00,
      0x60, 0x30, 0x18, 0x0c, 0x18, 0x30, 0x60, 0x00, 0x7c, 0xc6, 0x0c, 0x18,
      0x18, 0x00, 0x18, 0x00, 0x7c, 0xc6, 0xde, 0xde, 0xde, 0xc0, 0x78, 0x00,
      0x38, 0x6c, 0xc6, 0xfe, 0xc6, 0xc6, 0xc6, 0x00, 0xfc, 0x66, 0x66, 0x7c,
      0x66, 0x66, 0xfc, 0x00, 0x3c, 0x66, 0xc0, 0xc0, 0xc0, 0x66, 0x3c, 0x00,
      0xf8, 0x6c, 0x66, 0x66, 0x66, 0x6c, 0xf8, 0x00, 0xfe, 0x62, 0x68, 0x78,
      0x68, 0x62, 0xfe, 0x00, 0xfe, 0x62, 0x68, 0x78, 0x68, 0x60, 0xf0, 0x00,
      0x3c, 0x66, 0xc0, 0xc0, 0xce, 0x66, 0x3a, 0x00, 0xc6, 0xc6, 0xc6, 0xfe,
      0xc6, 0xc6, 0xc6, 0x00, 0x3c, 0x18, 0x18, 0x18, 0x18, 0x18, 0x3c, 0x00,
      0x1e, 0x0c, 0x0c, 0x0c, 0xcc, 0xcc, 0x78, 0x00, 0xe6, 0x66, 0x6c, 0x78,
      0x6c, 0x66, 0xe6, 0x00, 0xf0, 0x60, 0x60, 0x60, 0x62, 0x66, 0xfe, 0x00,
      0xc6, 0xee, 0xfe, 0xfe, 0xd6, 0xc6, 0xc6, 0x00, 0xc6, 0xe6, 0xf6, 0xde,
      0xce, 0xc6, 0xc6, 0x00, 0x7c, 0xc6, 0xc6, 0xc6, 0xc6, 0xc6, 0x7c, 0x00,
      0xfc, 0x66, 0x66, 0x7c, 0x60, 0x60, 0xf0, 0x00, 0x7c, 0xc6, 0xc6, 0xc6,
      0xd6, 0xde, 0x7c, 0x06, 0xfc, 0x66, 0x66, 0x7c, 0x6c, 0x66, 0xe6, 0x00,
      0x7c, 0xc6, 0x60, 0x38, 0x0c, 0xc6, 0x7c, 0x00, 0x7e, 0x7e, 0x5a, 0x18,
      0x18, 0x18, 0x3c, 0x00, 0xc6, 0xc6, 0xc6, 0xc6, 0xc6, 0xc6, 0x7c, 0x00,
      0xc6, 0xc6, 0xc6, 0xc6, 0xc6, 0x6c, 0x38, 0x00, 0xc6, 0xc6, 0xc6, 0xd6,
      0xd6, 0xfe, 0x6c, 0x00, 0xc6, 0xc6, 0x6c, 0x38, 0x6c, 0xc6, 0xc6, 0x00,
      0x66, 0x66, 0x66, 0x3c, 0x18, 0x18, 0x3c, 0x00, 0xfe, 0xc6, 0x8c, 0x18,
      0x32, 0x66, 0xfe, 0x00, 0x3c, 0x30, 0x30, 0x30, 0x30, 0x30, 0x3c, 0x00,
      0xc0, 0x60, 0x30, 0x18, 0x0c, 0x06, 0x02, 0x00, 0x3c, 0x0c, 0x0c, 0x0c,
      0x0c, 0x0c, 0x3c, 0x00, 0x10, 0x38, 0x6c, 0xc6, 0x00, 0x00, 0x00, 0x00,
      0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0x30, 0x18, 0x0c, 0x00,
      0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x78, 0x0c, 0x7c, 0xcc, 0x76, 0x00,
      0xe0, 0x60, 0x7c, 0x66, 0x66, 0x66, 0xdc, 0x00, 0x00, 0x00, 0x7c, 0xc6,
      0xc0, 0xc6, 0x7c, 0x00, 0x1c, 0x0c, 0x7c, 0xcc, 0xcc, 0xcc, 0x76, 0x00,
      0x00, 0x00, 0x7c, 0xc6, 0xfe, 0xc0, 0x7c, 0x00, 0x3c, 0x66, 0x60, 0xf8,
      0x60, 0x60, 0xf0, 0x00, 0x00, 0x00, 0x76, 0xcc, 0xcc, 0x7c, 0x0c, 0xf8,
      0xe0, 0x60, 0x6c, 0x76, 0x66, 0x66, 0xe6, 0x00, 0x18, 0x00, 0x38, 0x18,
      0x18, 0x18, 0x3c, 0x00, 0x06, 0x00, 0x06, 0x06, 0x06, 0x66, 0x66, 0x3c,
      0xe0, 0x60, 0x66, 0x6c, 0x78, 0x6c, 0xe6, 0x00, 0x38, 0x18, 0x18, 0x18,
      0x18, 0x18, 0x3c, 0x00, 0x00, 0x00, 0xec, 0xfe, 0xd6, 0xd6, 0xd6, 0x00,
      0x00, 0x00, 0xdc, 0x66, 0x66, 0x66, 0x66, 0x00, 0x00, 0x00, 0x7c, 0xc6,
      0xc6, 0xc6, 0x7c, 0x00, 0x00, 0x00, 0xdc, 0x66, 0x66, 0x7c, 0x60, 0xf0,
      0x00, 0x00, 0x76, 0xcc, 0xcc, 0x7c, 0x0c, 0x1e, 0x00, 0x00, 0xdc, 0x76,
      0x60, 0x60, 0xf0, 0x00, 0x00, 0x00, 0x7e, 0xc0, 0x7c, 0x06, 0xfc, 0x00,
      0x30, 0x30, 0xfc, 0x30, 0x30, 0x36, 0x1c, 0x00, 0x00, 0x00, 0xcc, 0xcc,
      0xcc, 0xcc, 0x76, 0x00, 0x00, 0x00, 0xc6, 0xc6, 0xc6, 0x6c, 0x38, 0x00,
      0x00, 0x00, 0xc6, 0xd6, 0xd6, 0xfe, 0x6c, 0x00, 0x00, 0x00, 0xc6, 0x6c,
      0x38, 0x6c, 0xc6, 0x00, 0x00, 0x00, 0xc6, 0xc6, 0xc6, 0x7e, 0x06, 0xfc,
      0x00, 0x00, 0xfe, 0x8c, 0x18, 0x32, 0xfe, 0x00, 0x0e, 0x18, 0x18, 0x70,
      0x18, 0x18, 0x0e, 0x00, 0x18, 0x18, 0x18, 0x18, 0x18, 0x18, 0x18, 0x00,
      0x70, 0x18, 0x18, 0x0e, 0x18, 0x18, 0x70, 0x00, 0x76, 0xdc, 0x00, 0x00,
      0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
    ];
  },
  GetCharBitmap: function (fontData, charCode) {
    let result = [];
    let code = charCode;
    if (code < 32) {
      code = 32;
    }
    if (code > 127) {
      code = 127;
    }
    let offset = (code - 32) * 8;
    let i = 0;
    while (true) {
      if (i >= 8) {
        break;
      }
      result = append(result, fontData[offset + i]);
      i = i + 1;
    }
    return result;
  },
  GetRow: function (fontData, charCode, y) {
    let code = charCode;
    if (code < 32) {
      code = 32;
    }
    if (code > 127) {
      code = 127;
    }
    let offset = (code - 32) * 8;
    return fontData[offset + y];
  },
  GetPixel: function (fontData, charCode, x, y) {
    let row = this.GetRow(fontData, charCode, y);
    let mask = uint8(0x80 >> x);
    return (row & mask) != 0;
  },
};
const TextCols = 40;
const TextRows = 25;
const TextScreenBase = 0x0400;

function hexDigit(n) {
  if (n == 0) {
    return "0";
  } else if (n == 1) {
    return "1";
  } else if (n == 2) {
    return "2";
  } else if (n == 3) {
    return "3";
  } else if (n == 4) {
    return "4";
  } else if (n == 5) {
    return "5";
  } else if (n == 6) {
    return "6";
  } else if (n == 7) {
    return "7";
  } else if (n == 8) {
    return "8";
  } else if (n == 9) {
    return "9";
  } else if (n == 10) {
    return "A";
  } else if (n == 11) {
    return "B";
  } else if (n == 12) {
    return "C";
  } else if (n == 13) {
    return "D";
  } else if (n == 14) {
    return "E";
  } else if (n == 15) {
    return "F";
  }
  return "0";
}

function toHex2(n) {
  let high = (n >> 4) & 0x0f;
  let low = n & 0x0f;
  return hexDigit(high) + hexDigit(low);
}

function toHex4(n) {
  return toHex2((n >> 8) & 0xff) + toHex2(n & 0xff);
}

function toHex(n) {
  if (n > 255) {
    return "$" + toHex4(n);
  }
  return "$" + toHex2(n);
}

function addStringToScreen(lines, text, row, col) {
  let baseAddr = TextScreenBase + row * TextCols + col;
  let i = 0;
  while (true) {
    if (i >= len(text)) {
      break;
    }
    let charCode = int(text.charCodeAt(i));
    let addr = baseAddr + i;
    lines = append(lines, "LDA #" + toHex(charCode));
    lines = append(lines, "STA " + toHex(addr));
    i = i + 1;
  }
  return lines;
}

function clearScreen(lines) {
  lines = append(lines, "LDA #$20");
  let addr = TextScreenBase;
  let i = 0;
  while (true) {
    if (i >= TextCols * TextRows) {
      break;
    }
    lines = append(lines, "STA " + toHex(addr + i));
    i = i + 1;
  }
  return lines;
}

function charFromCodeMain(ch) {
  if (ch == 32) {
    return " ";
  } else if (ch == 33) {
    return "!";
  } else if (ch == 34) {
    return '"';
  } else if (ch == 35) {
    return "#";
  } else if (ch == 36) {
    return "$";
  } else if (ch == 37) {
    return "%";
  } else if (ch == 38) {
    return "&";
  } else if (ch == 39) {
    return "'";
  } else if (ch == 40) {
    return "(";
  } else if (ch == 41) {
    return ")";
  } else if (ch == 42) {
    return "*";
  } else if (ch == 43) {
    return "+";
  } else if (ch == 44) {
    return ",";
  } else if (ch == 45) {
    return "-";
  } else if (ch == 46) {
    return ".";
  } else if (ch == 47) {
    return "/";
  } else if (ch == 48) {
    return "0";
  } else if (ch == 49) {
    return "1";
  } else if (ch == 50) {
    return "2";
  } else if (ch == 51) {
    return "3";
  } else if (ch == 52) {
    return "4";
  } else if (ch == 53) {
    return "5";
  } else if (ch == 54) {
    return "6";
  } else if (ch == 55) {
    return "7";
  } else if (ch == 56) {
    return "8";
  } else if (ch == 57) {
    return "9";
  } else if (ch == 58) {
    return ":";
  } else if (ch == 59) {
    return ";";
  } else if (ch == 60) {
    return "<";
  } else if (ch == 61) {
    return "=";
  } else if (ch == 62) {
    return ">";
  } else if (ch == 63) {
    return "?";
  } else if (ch == 64) {
    return "@";
  } else if (ch == 65) {
    return "A";
  } else if (ch == 66) {
    return "B";
  } else if (ch == 67) {
    return "C";
  } else if (ch == 68) {
    return "D";
  } else if (ch == 69) {
    return "E";
  } else if (ch == 70) {
    return "F";
  } else if (ch == 71) {
    return "G";
  } else if (ch == 72) {
    return "H";
  } else if (ch == 73) {
    return "I";
  } else if (ch == 74) {
    return "J";
  } else if (ch == 75) {
    return "K";
  } else if (ch == 76) {
    return "L";
  } else if (ch == 77) {
    return "M";
  } else if (ch == 78) {
    return "N";
  } else if (ch == 79) {
    return "O";
  } else if (ch == 80) {
    return "P";
  } else if (ch == 81) {
    return "Q";
  } else if (ch == 82) {
    return "R";
  } else if (ch == 83) {
    return "S";
  } else if (ch == 84) {
    return "T";
  } else if (ch == 85) {
    return "U";
  } else if (ch == 86) {
    return "V";
  } else if (ch == 87) {
    return "W";
  } else if (ch == 88) {
    return "X";
  } else if (ch == 89) {
    return "Y";
  } else if (ch == 90) {
    return "Z";
  } else if (ch == 91) {
    return "[";
  } else if (ch == 93) {
    return "]";
  } else if (ch == 94) {
    return "^";
  } else if (ch == 95) {
    return "_";
  } else if (ch == 96) {
    return "`";
  } else if (ch == 97) {
    return "a";
  } else if (ch == 98) {
    return "b";
  } else if (ch == 99) {
    return "c";
  } else if (ch == 100) {
    return "d";
  } else if (ch == 101) {
    return "e";
  } else if (ch == 102) {
    return "f";
  } else if (ch == 103) {
    return "g";
  } else if (ch == 104) {
    return "h";
  } else if (ch == 105) {
    return "i";
  } else if (ch == 106) {
    return "j";
  } else if (ch == 107) {
    return "k";
  } else if (ch == 108) {
    return "l";
  } else if (ch == 109) {
    return "m";
  } else if (ch == 110) {
    return "n";
  } else if (ch == 111) {
    return "o";
  } else if (ch == 112) {
    return "p";
  } else if (ch == 113) {
    return "q";
  } else if (ch == 114) {
    return "r";
  } else if (ch == 115) {
    return "s";
  } else if (ch == 116) {
    return "t";
  } else if (ch == 117) {
    return "u";
  } else if (ch == 118) {
    return "v";
  } else if (ch == 119) {
    return "w";
  } else if (ch == 120) {
    return "x";
  } else if (ch == 121) {
    return "y";
  } else if (ch == 122) {
    return "z";
  } else if (ch == 123) {
    return "{";
  } else if (ch == 124) {
    return "|";
  } else if (ch == 125) {
    return "}";
  } else if (ch == 126) {
    return "~";
  }
  return "";
}

function digitToCharMain(d) {
  if (d == 0) {
    return "0";
  } else if (d == 1) {
    return "1";
  } else if (d == 2) {
    return "2";
  } else if (d == 3) {
    return "3";
  } else if (d == 4) {
    return "4";
  } else if (d == 5) {
    return "5";
  } else if (d == 6) {
    return "6";
  } else if (d == 7) {
    return "7";
  } else if (d == 8) {
    return "8";
  } else if (d == 9) {
    return "9";
  }
  return "0";
}

function intToString(n) {
  if (n == 0) {
    return "0";
  }
  let neg = false;
  if (n < 0) {
    neg = true;
    n = -n;
  }
  let digits = "";
  while (true) {
    if (n == 0) {
      break;
    }
    let digit = n % 10;
    digits = digitToCharMain(digit) + digits;
    n = (n / 10) | 0;
  }
  if (neg) {
    digits = "-" + digits;
  }
  return digits;
}

function readLineFromScreen(c, row) {
  let result = "";
  let baseAddr = TextScreenBase + row * TextCols;
  let col = 0;
  while (true) {
    if (col >= TextCols) {
      break;
    }
    let ch = int(c.Memory[baseAddr + col]);
    if (ch >= 32 && ch <= 126) {
      result = result + charFromCodeMain(ch);
    }
    col = col + 1;
  }
  let end = len(result);
  while (true) {
    if (end <= 0) {
      break;
    }
    if (result.charCodeAt(end - 1) != 32) {
      break;
    }
    end = end - 1;
  }
  if (end <= 0) {
    return "";
  }
  let trimmed = "";
  let i = 0;
  while (true) {
    if (i >= end) {
      break;
    }
    trimmed = trimmed + charFromCodeMain(int(result.charCodeAt(i)));
    i = i + 1;
  }
  return trimmed;
}

function printReady(c, row) {
  let text = "READY.";
  let baseAddr = TextScreenBase + row * TextCols;
  let i = 0;
  while (true) {
    if (i >= len(text)) {
      break;
    }
    c.Memory[baseAddr + i] = uint8(text.charCodeAt(i));
    i = i + 1;
  }
  return c;
}

function createC64WelcomeScreen() {
  let lines = [];
  lines = clearScreen(lines);
  lines = addStringToScreen(lines, "**** COMMODORE 64 BASIC V2 ****", 1, 4);
  lines = addStringToScreen(
    lines,
    "64K RAM SYSTEM  38911 BASIC BYTES FREE",
    3,
    1,
  );
  lines = addStringToScreen(lines, "READY.", 5, 0);
  lines = append(lines, "LDA #$5F");
  lines = append(lines, "STA $0518");
  lines = append(lines, "BRK");
  return assembler.AssembleLines(lines);
}

function main() {
  let scale = int32(4);
  let windowWidth = int32(TextCols * 8) * scale;
  let windowHeight = int32(TextRows * 8) * scale;
  let w = graphics.CreateWindow("Commodore 64", windowWidth, windowHeight);
  let c = cpu.NewCPU();
  let fontData = font.GetFontData();
  let program = createC64WelcomeScreen();
  c = cpu.LoadProgram(c, program, 0x0600);
  c = cpu.SetPC(c, 0x0600);
  c = cpu.ClearHalted(c);
  c = cpu.Run(c, 100000);
  let textColor = graphics.NewColor(134, 122, 222, 255);
  let bgColor = graphics.NewColor(64, 50, 133, 255);
  let cursorRow = 7;
  let cursorCol = 0;
  let basicState = basic.NewBasicState();
  basicState = basic.SetCursor(basicState, cursorRow, cursorCol);
  let inputStartRow = cursorRow;
  graphics.RunLoop(w, function (w) {
    let key = graphics.GetLastKey();
    if (key != 0) {
      let oldCursorAddr = TextScreenBase + cursorRow * TextCols + cursorCol;
      if (c.Memory[oldCursorAddr] == 95) {
        c.Memory[oldCursorAddr] = 32;
      }
      if (key == 13) {
        let line = readLineFromScreen(c, inputStartRow);
        cursorCol = 0;
        cursorRow = cursorRow + 1;
        if (cursorRow >= TextRows) {
          cursorRow = TextRows - 1;
        }
        if (len(line) > 0) {
          if (basic.HasLineNumber(line)) {
            let [lineNum, rest] = basic.ExtractLineNumber(line);
            basicState = basic.StoreLine(basicState, lineNum, rest);
          } else if (basic.IsCommand(line, "RUN")) {
            let lineCount = basic.GetLineCount(basicState);
            if (lineCount > 0) {
              basicState = basic.SetCursor(basicState, cursorRow, 0);
              let code = basic.CompileProgram(basicState);
              cursorRow = cursorRow + 1;
              if (cursorRow >= TextRows) {
                cursorRow = TextRows - 1;
              }
              c = cpu.LoadProgram(c, code, 0xc000);
              c = cpu.SetPC(c, 0xc000);
              c = cpu.ClearHalted(c);
              c = cpu.Run(c, 100000);
            }
            cursorRow = cursorRow + lineCount;
            if (cursorRow >= TextRows) {
              cursorRow = TextRows - 1;
            }
            c = printReady(c, cursorRow);
            cursorRow = cursorRow + 1;
            if (cursorRow >= TextRows) {
              cursorRow = TextRows - 1;
            }
          } else if (basic.IsCommand(line, "LIST")) {
            let i = 0;
            let listRow = cursorRow;
            while (true) {
              if (i >= basic.GetLineCount(basicState)) {
                break;
              }
              if (listRow >= TextRows) {
                break;
              }
              let pl = basic.GetLine(basicState, i);
              let numStr = intToString(pl.LineNum);
              let listLine = numStr + " " + pl.Text;
              let baseAddr = TextScreenBase + listRow * TextCols;
              let j = 0;
              while (true) {
                if (j >= len(listLine)) {
                  break;
                }
                if (j >= TextCols) {
                  break;
                }
                c.Memory[baseAddr + j] = uint8(listLine.charCodeAt(j));
                j = j + 1;
              }
              listRow = listRow + 1;
              i = i + 1;
            }
            cursorRow = listRow;
            c = printReady(c, cursorRow);
            cursorRow = cursorRow + 1;
            if (cursorRow >= TextRows) {
              cursorRow = TextRows - 1;
            }
          } else if (basic.IsCommand(line, "NEW")) {
            basicState = basic.ClearProgram(basicState);
            c = printReady(c, cursorRow);
            cursorRow = cursorRow + 1;
            if (cursorRow >= TextRows) {
              cursorRow = TextRows - 1;
            }
          } else if (basic.IsCommand(line, "CLR")) {
            basicState = basic.SetCursor(basicState, 0, 0);
            let code = basic.CompileImmediate(basicState, line);
            c = cpu.LoadProgram(c, code, 0xc000);
            c = cpu.SetPC(c, 0xc000);
            c = cpu.ClearHalted(c);
            c = cpu.Run(c, 100000);
            cursorRow = 0;
            cursorCol = 0;
            c = printReady(c, cursorRow);
            cursorRow = cursorRow + 1;
          } else {
            basicState = basic.SetCursor(basicState, cursorRow, 0);
            let code = basic.CompileImmediate(basicState, line);
            if (len(code) > 1) {
              c = cpu.LoadProgram(c, code, 0xc000);
              c = cpu.SetPC(c, 0xc000);
              c = cpu.ClearHalted(c);
              c = cpu.Run(c, 10000);
            }
            if (basic.IsCommand(line, "PRINT")) {
              cursorRow = cursorRow + 1;
              if (cursorRow >= TextRows) {
                cursorRow = TextRows - 1;
              }
            }
            c = printReady(c, cursorRow);
            cursorRow = cursorRow + 1;
            if (cursorRow >= TextRows) {
              cursorRow = TextRows - 1;
            }
          }
        }
        inputStartRow = cursorRow;
      } else if (key == 8) {
        if (cursorCol > 0) {
          cursorCol = cursorCol - 1;
          let addr = TextScreenBase + cursorRow * TextCols + cursorCol;
          c.Memory[addr] = 32;
        } else if (cursorRow > 0) {
          cursorRow = cursorRow - 1;
          cursorCol = TextCols - 1;
        }
      } else if (key >= 32 && key <= 126) {
        let addr = TextScreenBase + cursorRow * TextCols + cursorCol;
        c.Memory[addr] = uint8(key);
        cursorCol = cursorCol + 1;
        if (cursorCol >= TextCols) {
          cursorCol = 0;
          cursorRow = cursorRow + 1;
          if (cursorRow >= TextRows) {
            cursorRow = TextRows - 1;
          }
        }
      }
      let cursorAddr = TextScreenBase + cursorRow * TextCols + cursorCol;
      if (c.Memory[cursorAddr] == 32) {
        c.Memory[cursorAddr] = 95;
      }
    }
    graphics.Clear(w, bgColor);
    let memAddr = TextScreenBase;
    let charY = 0;
    while (true) {
      if (charY >= TextRows) {
        break;
      }
      let charX = 0;
      while (true) {
        if (charX >= TextCols) {
          break;
        }
        let charCode = int(cpu.GetMemory(c, memAddr));
        memAddr = memAddr + 1;
        if (charCode > 32 && charCode <= 127) {
          let baseScreenX = int32(charX * 8);
          let baseScreenY = int32(charY * 8);
          let pixelY = 0;
          while (true) {
            if (pixelY >= 8) {
              break;
            }
            let rowByte = font.GetRow(fontData, charCode, pixelY);
            if (rowByte != 0) {
              let mask = uint8(0x80);
              let pixelX = 0;
              while (true) {
                if (pixelX >= 8) {
                  break;
                }
                if ((rowByte & mask) != 0) {
                  let screenX = (baseScreenX + int32(pixelX)) * scale;
                  let screenY = (baseScreenY + int32(pixelY)) * scale;
                  graphics.FillRect(
                    w,
                    graphics.NewRect(screenX, screenY, scale, scale),
                    textColor,
                  );
                }
                mask = mask >> 1;
                pixelX = pixelX + 1;
              }
            }
            pixelY = pixelY + 1;
          }
        }
        charX = charX + 1;
      }
      charY = charY + 1;
    }
    graphics.Present(w);
    return true;
  });
  graphics.CloseWindow(w);
}

// Run main
main();

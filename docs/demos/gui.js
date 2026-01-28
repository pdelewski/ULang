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
  // Use push for O(1) amortized instead of spread which is O(n)
  for (const item of items) {
    if (item && typeof item === "object" && !Array.isArray(item)) {
      arr.push({ ...item });
    } else {
      arr.push(item);
    }
  }
  return arr;
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
    // Make canvas fill entire browser window
    document.body.style.margin = "0";
    document.body.style.padding = "0";
    document.body.style.overflow = "hidden";

    this.canvas = document.createElement("canvas");
    this.canvas.width = window.innerWidth;
    this.canvas.height = window.innerHeight;
    this.canvas.style.display = "block";
    this.ctx = this.canvas.getContext("2d");
    document.body.appendChild(this.canvas);
    document.title = title;

    // Create window object with dimensions that can be updated
    this.windowObj = {
      canvas: this.canvas,
      width: this.canvas.width,
      height: this.canvas.height,
    };

    // Resize canvas when browser window resizes
    window.addEventListener("resize", () => {
      this.canvas.width = window.innerWidth;
      this.canvas.height = window.innerHeight;
      this.windowObj.width = this.canvas.width;
      this.windowObj.height = this.canvas.height;
    });

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

    return this.windowObj;
  },

  // CreateWindowFullscreen - same as CreateWindow since JS canvas already fills the browser window
  CreateWindowFullscreen: function (title, width, height) {
    return this.CreateWindow(title, width, height);
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

  GetMouse: function (canvas) {
    // Returns [x, y, buttons] like other backends
    return [this.mouseX, this.mouseY, this.mouseDown ? 1 : 0];
  },

  GetWidth: function (w) {
    return w.width;
  },

  GetHeight: function (w) {
    return w.height;
  },

  GetScreenSize: function () {
    return [window.screen.width, window.screen.height];
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

const gui = {
  NewWindowState: function (x, y, width, height) {
    return {
      X: x,
      Y: y,
      Width: width,
      Height: height,
      Dragging: false,
      DragOffsetX: 0,
      DragOffsetY: 0,
    };
  },
  NewMenuState: function () {
    return {
      OpenMenuID: 0,
      MenuBarX: 0,
      MenuBarY: 0,
      MenuBarH: 0,
      CurrentMenuX: 0,
      CurrentMenuW: 0,
      ClickedOutside: false,
    };
  },
  DefaultStyle: function () {
    return {
      BackgroundColor: graphics.NewColor(15, 15, 15, 240),
      TextColor: graphics.NewColor(255, 255, 255, 255),
      ButtonColor: graphics.NewColor(66, 150, 250, 102),
      ButtonHoverColor: graphics.NewColor(66, 150, 250, 200),
      ButtonActiveColor: graphics.NewColor(15, 135, 250, 255),
      CheckboxColor: graphics.NewColor(41, 74, 122, 255),
      CheckmarkColor: graphics.NewColor(66, 150, 250, 255),
      SliderTrackColor: graphics.NewColor(66, 150, 250, 171),
      SliderKnobColor: graphics.NewColor(66, 150, 250, 200),
      BorderColor: graphics.NewColor(80, 80, 80, 255),
      FrameBgColor: graphics.NewColor(29, 47, 73, 138),
      TitleBgColor: graphics.NewColor(41, 74, 137, 255),
      FontSize: 1,
      Padding: 6,
      ButtonHeight: 20,
      SliderHeight: 18,
      CheckboxSize: 16,
      FrameRounding: 2,
    };
  },
  NewContext: function () {
    return {
      Style: this.DefaultStyle(),
      MouseX: 0,
      MouseY: 0,
      MouseDown: false,
      MouseClicked: false,
      MouseReleased: false,
      HotID: 0,
      ActiveID: 0,
      ReleasedID: 0,
      CursorX: 0,
      CursorY: 0,
      Spacing: 0,
    };
  },
  GenID: function (label) {
    let hash = int32(5381);
    for (let i = 0; i < len(label); i++) {
      hash = (hash << 5) + hash + int32(label.charCodeAt(i));
    }
    if (hash < 0) {
      hash = -hash;
    }
    return hash;
  },
  UpdateInput: function (ctx, w) {
    let prevDown = ctx.MouseDown;
    let [x, y, buttons] = graphics.GetMouse(w);
    ctx.MouseX = x;
    ctx.MouseY = y;
    ctx.MouseDown = (buttons & 1) != 0;
    ctx.MouseClicked = ctx.MouseDown && !prevDown;
    ctx.MouseReleased = !ctx.MouseDown && prevDown;
    ctx.ReleasedID = 0;
    if (ctx.MouseReleased) {
      ctx.ReleasedID = ctx.ActiveID;
      ctx.ActiveID = 0;
    }
    ctx.HotID = 0;
    return ctx;
  },
  drawChar: function (w, charCode, x, y, scale, color) {
    if (charCode < 32 || charCode > 127) {
      charCode = 32;
    }
    let offset = (charCode - 32) * 8;
    let font = this.getFontData();
    for (let row = int32(0); row < 8; row++) {
      let rowData = font[offset + int(row)];
      for (let col = int32(0); col < 8; col++) {
        if ((rowData & (0x80 >> col)) != 0) {
          for (let sy = int32(0); sy < scale; sy++) {
            for (let sx = int32(0); sx < scale; sx++) {
              graphics.DrawPoint(
                w,
                x + col * scale + sx,
                y + row * scale + sy,
                color,
              );
            }
          }
        }
      }
    }
  },
  DrawText: function (w, text, x, y, scale, color) {
    let curX = x;
    for (let i = 0; i < len(text); i++) {
      let ch = int(text.charCodeAt(i));
      this.drawChar(w, ch, curX, y, scale, color);
      curX = curX + 8 * scale;
    }
  },
  TextWidth: function (text, scale) {
    return int32(len(text)) * 8 * scale;
  },
  TextHeight: function (scale) {
    return 8 * scale;
  },
  pointInRect: function (px, py, x, y, w, h) {
    return px >= x && px < x + w && py >= y && py < y + h;
  },
  Label: function (ctx, w, text, x, y) {
    this.DrawText(w, text, x, y, ctx.Style.FontSize, ctx.Style.TextColor);
  },
  Button: function (ctx, w, label, x, y, width, height) {
    let id = this.GenID(label);
    let hovered = this.pointInRect(ctx.MouseX, ctx.MouseY, x, y, width, height);
    if (hovered) {
      ctx.HotID = id;
      if (ctx.MouseClicked) {
        ctx.ActiveID = id;
      }
    }
    let bgColor = { R: 0, G: 0, B: 0, A: 0 };
    let pressOffset = int32(0);
    if (ctx.ActiveID == id && hovered) {
      bgColor = ctx.Style.ButtonActiveColor;
      pressOffset = 1;
    } else if (ctx.HotID == id) {
      bgColor = ctx.Style.ButtonHoverColor;
    } else {
      bgColor = ctx.Style.ButtonColor;
    }
    graphics.DrawLine(
      w,
      x,
      y + height - 1,
      x,
      y,
      graphics.NewColor(0, 0, 0, 80),
    );
    graphics.DrawLine(
      w,
      x,
      y,
      x + width - 1,
      y,
      graphics.NewColor(0, 0, 0, 80),
    );
    graphics.FillRect(
      w,
      graphics.NewRect(x + 1, y + 1, width - 2, height - 2),
      bgColor,
    );
    graphics.DrawLine(
      w,
      x + 1,
      y + height - 1,
      x + width - 1,
      y + height - 1,
      graphics.NewColor(255, 255, 255, 30),
    );
    graphics.DrawLine(
      w,
      x + width - 1,
      y + 1,
      x + width - 1,
      y + height - 1,
      graphics.NewColor(255, 255, 255, 30),
    );
    let textW = this.TextWidth(label, ctx.Style.FontSize);
    let textH = this.TextHeight(ctx.Style.FontSize);
    let textX = x + (((width - textW) / 2) | 0) + pressOffset;
    let textY = y + (((height - textH) / 2) | 0) + pressOffset;
    this.DrawText(
      w,
      label,
      textX,
      textY,
      ctx.Style.FontSize,
      ctx.Style.TextColor,
    );
    let clicked = ctx.ReleasedID == id && ctx.MouseReleased && hovered;
    return [ctx, clicked];
  },
  Checkbox: function (ctx, w, label, x, y, value) {
    let id = this.GenID(label);
    let boxSize = ctx.Style.CheckboxSize;
    let labelW = this.TextWidth(label, ctx.Style.FontSize);
    let totalW = boxSize + ctx.Style.Padding + labelW;
    let hovered = this.pointInRect(
      ctx.MouseX,
      ctx.MouseY,
      x,
      y,
      totalW,
      boxSize,
    );
    if (hovered) {
      ctx.HotID = id;
      if (ctx.MouseClicked) {
        ctx.ActiveID = id;
      }
    }
    let boxColor = { R: 0, G: 0, B: 0, A: 0 };
    if (ctx.HotID == id) {
      boxColor = ctx.Style.ButtonHoverColor;
    } else {
      boxColor = ctx.Style.FrameBgColor;
    }
    graphics.FillRect(w, graphics.NewRect(x, y, boxSize, boxSize), boxColor);
    graphics.DrawLine(
      w,
      x,
      y,
      x + boxSize - 1,
      y,
      graphics.NewColor(0, 0, 0, 100),
    );
    graphics.DrawLine(
      w,
      x,
      y,
      x,
      y + boxSize - 1,
      graphics.NewColor(0, 0, 0, 100),
    );
    if (value) {
      let checkColor = ctx.Style.CheckmarkColor;
      let cx = x + ((boxSize / 2) | 0);
      let cy = y + ((boxSize / 2) | 0);
      graphics.DrawLine(w, cx - 5, cy, cx - 2, cy + 4, checkColor);
      graphics.DrawLine(w, cx - 4, cy, cx - 1, cy + 4, checkColor);
      graphics.DrawLine(w, cx - 2, cy + 4, cx + 5, cy - 3, checkColor);
      graphics.DrawLine(w, cx - 1, cy + 4, cx + 6, cy - 3, checkColor);
    }
    let labelX = x + boxSize + ctx.Style.Padding;
    let labelY =
      y + (((boxSize - this.TextHeight(ctx.Style.FontSize)) / 2) | 0);
    this.DrawText(
      w,
      label,
      labelX,
      labelY,
      ctx.Style.FontSize,
      ctx.Style.TextColor,
    );
    let newValue = value;
    if (ctx.ReleasedID == id && ctx.MouseReleased && hovered) {
      newValue = !value;
    }
    return [ctx, newValue];
  },
  Slider: function (ctx, w, label, x, y, width, min, max, value) {
    let id = this.GenID(label);
    let height = ctx.Style.SliderHeight;
    let grabW = int32(12);
    let labelW = this.TextWidth(label, ctx.Style.FontSize);
    let labelY = y + (((height - this.TextHeight(ctx.Style.FontSize)) / 2) | 0);
    this.DrawText(w, label, x, labelY, ctx.Style.FontSize, ctx.Style.TextColor);
    let trackX = x + labelW + ctx.Style.Padding;
    let trackW = width - labelW - ctx.Style.Padding;
    if (value < min) {
      value = min;
    }
    if (value > max) {
      value = max;
    }
    let rangeVal = max - min;
    if (rangeVal == 0) {
      rangeVal = 1;
    }
    let t = (value - min) / rangeVal;
    let grabRange = trackW - grabW;
    let grabX = trackX + int32(float64(grabRange) * t);
    let hovered = this.pointInRect(
      ctx.MouseX,
      ctx.MouseY,
      trackX,
      y,
      trackW,
      height,
    );
    if (hovered) {
      ctx.HotID = id;
      if (ctx.MouseClicked) {
        ctx.ActiveID = id;
      }
    }
    graphics.FillRect(
      w,
      graphics.NewRect(trackX, y, trackW, height),
      ctx.Style.FrameBgColor,
    );
    graphics.DrawLine(
      w,
      trackX,
      y,
      trackX + trackW - 1,
      y,
      graphics.NewColor(0, 0, 0, 100),
    );
    graphics.DrawLine(
      w,
      trackX,
      y,
      trackX,
      y + height - 1,
      graphics.NewColor(0, 0, 0, 100),
    );
    let fillW = grabX - trackX + ((grabW / 2) | 0);
    if (fillW > 0) {
      graphics.FillRect(
        w,
        graphics.NewRect(trackX + 1, y + 1, fillW, height - 2),
        ctx.Style.SliderTrackColor,
      );
    }
    let grabColor = { R: 0, G: 0, B: 0, A: 0 };
    if (ctx.ActiveID == id) {
      grabColor = ctx.Style.ButtonActiveColor;
    } else if (ctx.HotID == id) {
      grabColor = ctx.Style.ButtonHoverColor;
    } else {
      grabColor = ctx.Style.SliderKnobColor;
    }
    graphics.FillRect(w, graphics.NewRect(grabX, y, grabW, height), grabColor);
    if (ctx.ActiveID == id && ctx.MouseDown) {
      let mouseT =
        float64(ctx.MouseX - trackX - ((grabW / 2) | 0)) / float64(grabRange);
      if (mouseT < 0) {
        mouseT = 0;
      }
      if (mouseT > 1) {
        mouseT = 1;
      }
      value = min + mouseT * rangeVal;
    }
    return [ctx, value];
  },
  Panel: function (ctx, w, title, x, y, width, height) {
    let titleH = this.TextHeight(ctx.Style.FontSize) + ctx.Style.Padding * 2;
    graphics.FillRect(
      w,
      graphics.NewRect(x, y, width, titleH),
      ctx.Style.TitleBgColor,
    );
    this.DrawText(
      w,
      title,
      x + ctx.Style.Padding,
      y + (((titleH - this.TextHeight(ctx.Style.FontSize)) / 2) | 0),
      ctx.Style.FontSize,
      ctx.Style.TextColor,
    );
    graphics.FillRect(
      w,
      graphics.NewRect(x, y + titleH, width, height - titleH),
      ctx.Style.BackgroundColor,
    );
    graphics.DrawRect(
      w,
      graphics.NewRect(x, y, width, height),
      ctx.Style.BorderColor,
    );
    graphics.DrawLine(
      w,
      x + 1,
      y + titleH - 1,
      x + width - 2,
      y + titleH - 1,
      graphics.NewColor(255, 255, 255, 20),
    );
  },
  DraggablePanel: function (ctx, w, title, state) {
    let idStr = title;
    idStr += "_panel";
    let id = this.GenID(idStr);
    let titleH = this.TextHeight(ctx.Style.FontSize) + ctx.Style.Padding * 2;
    let inTitleBar = this.pointInRect(
      ctx.MouseX,
      ctx.MouseY,
      state.X,
      state.Y,
      state.Width,
      titleH,
    );
    if (inTitleBar && ctx.MouseClicked) {
      state.Dragging = true;
      state.DragOffsetX = ctx.MouseX - state.X;
      state.DragOffsetY = ctx.MouseY - state.Y;
      ctx.ActiveID = id;
    }
    if (state.Dragging && ctx.MouseDown) {
      state.X = ctx.MouseX - state.DragOffsetX;
      state.Y = ctx.MouseY - state.DragOffsetY;
    }
    if (state.Dragging && ctx.MouseReleased) {
      state.Dragging = false;
      if (ctx.ActiveID == id) {
        ctx.ActiveID = 0;
      }
    }
    this.Panel(ctx, w, title, state.X, state.Y, state.Width, state.Height);
    return [ctx, state];
  },
  Separator: function (ctx, w, x, y, width) {
    graphics.DrawLine(w, x, y, x + width, y, ctx.Style.BorderColor);
  },
  BeginMenuBar: function (ctx, w, state, x, y, width) {
    let height = this.TextHeight(ctx.Style.FontSize) + ctx.Style.Padding * 2;
    graphics.FillRect(
      w,
      graphics.NewRect(x, y, width, height),
      ctx.Style.TitleBgColor,
    );
    graphics.DrawLine(
      w,
      x,
      y + height - 1,
      x + width,
      y + height - 1,
      ctx.Style.BorderColor,
    );
    state.MenuBarX = x;
    state.MenuBarY = y;
    state.MenuBarH = height;
    state.CurrentMenuX = x + ctx.Style.Padding;
    state.ClickedOutside = ctx.MouseClicked;
    return [ctx, state];
  },
  EndMenuBar: function (ctx, state) {
    if (state.ClickedOutside && state.OpenMenuID != 0) {
      state.OpenMenuID = 0;
    }
    return [ctx, state];
  },
  Menu: function (ctx, w, state, label) {
    let id = this.GenID(label);
    let padding = ctx.Style.Padding;
    let textW = this.TextWidth(label, ctx.Style.FontSize);
    let textH = this.TextHeight(ctx.Style.FontSize);
    let menuW = textW + padding * 2;
    let menuH = state.MenuBarH;
    let x = state.CurrentMenuX;
    let y = state.MenuBarY;
    let hovered = this.pointInRect(ctx.MouseX, ctx.MouseY, x, y, menuW, menuH);
    let isOpen = state.OpenMenuID == id;
    if (hovered || isOpen) {
      graphics.FillRect(
        w,
        graphics.NewRect(x, y, menuW, menuH - 1),
        ctx.Style.ButtonHoverColor,
      );
      if (ctx.MouseClicked) {
        if (isOpen) {
          state.OpenMenuID = 0;
          isOpen = false;
        } else {
          state.OpenMenuID = id;
          isOpen = true;
        }
        state.ClickedOutside = false;
      }
    }
    let textY = y + (((menuH - textH) / 2) | 0);
    this.DrawText(
      w,
      label,
      x + padding,
      textY,
      ctx.Style.FontSize,
      ctx.Style.TextColor,
    );
    state.CurrentMenuW = menuW;
    state.CurrentMenuX = x + menuW;
    return [ctx, state, isOpen];
  },
  BeginDropdown: function (ctx, w, state) {
    let dropY = state.MenuBarY + state.MenuBarH;
    return [ctx, dropY];
  },
  MenuItem: function (ctx, w, state, label, dropX, dropY, itemIndex) {
    let padding = ctx.Style.Padding;
    let textH = this.TextHeight(ctx.Style.FontSize);
    let itemH = textH + padding * 2;
    let itemW = int32(150);
    let y = dropY + itemIndex * itemH;
    graphics.FillRect(
      w,
      graphics.NewRect(dropX, y, itemW, itemH),
      ctx.Style.BackgroundColor,
    );
    let hovered = this.pointInRect(
      ctx.MouseX,
      ctx.MouseY,
      dropX,
      y,
      itemW,
      itemH,
    );
    let clicked = false;
    if (hovered) {
      graphics.FillRect(
        w,
        graphics.NewRect(dropX, y, itemW, itemH),
        ctx.Style.ButtonHoverColor,
      );
      state.ClickedOutside = false;
      if (ctx.MouseClicked) {
        clicked = true;
        state.OpenMenuID = 0;
      }
    }
    let textY = y + (((itemH - textH) / 2) | 0);
    this.DrawText(
      w,
      label,
      dropX + padding,
      textY,
      ctx.Style.FontSize,
      ctx.Style.TextColor,
    );
    graphics.DrawRect(
      w,
      graphics.NewRect(dropX, dropY, itemW, (itemIndex + 1) * itemH),
      ctx.Style.BorderColor,
    );
    return [ctx, state, clicked];
  },
  MenuItemSeparator: function (ctx, w, dropX, dropY, itemIndex) {
    let padding = ctx.Style.Padding;
    let textH = this.TextHeight(ctx.Style.FontSize);
    let itemH = textH + padding * 2;
    let itemW = int32(150);
    let y = dropY + itemIndex * itemH + ((itemH / 2) | 0);
    graphics.FillRect(
      w,
      graphics.NewRect(dropX, dropY + itemIndex * itemH, itemW, itemH),
      ctx.Style.BackgroundColor,
    );
    graphics.DrawLine(
      w,
      dropX + padding,
      y,
      dropX + itemW - padding,
      y,
      ctx.Style.BorderColor,
    );
  },
  BeginLayout: function (ctx, x, y, spacing) {
    ctx.CursorX = x;
    ctx.CursorY = y;
    ctx.Spacing = spacing;
    return ctx;
  },
  NextRow: function (ctx, height) {
    ctx.CursorY = ctx.CursorY + height + ctx.Spacing;
    return ctx;
  },
  AutoLabel: function (ctx, w, text) {
    this.Label(ctx, w, text, ctx.CursorX, ctx.CursorY);
    ctx = this.NextRow(ctx, this.TextHeight(ctx.Style.FontSize));
    return ctx;
  },
  AutoButton: function (ctx, w, label, width, height) {
    let result = false;
    [ctx, result] = this.Button(
      ctx,
      w,
      label,
      ctx.CursorX,
      ctx.CursorY,
      width,
      height,
    );
    ctx = this.NextRow(ctx, height);
    return [ctx, result];
  },
  AutoCheckbox: function (ctx, w, label, value) {
    let result = false;
    [ctx, result] = this.Checkbox(
      ctx,
      w,
      label,
      ctx.CursorX,
      ctx.CursorY,
      value,
    );
    ctx = this.NextRow(ctx, ctx.Style.CheckboxSize);
    return [ctx, result];
  },
  AutoSlider: function (ctx, w, label, width, min, max, value) {
    let result = 0;
    [ctx, result] = this.Slider(
      ctx,
      w,
      label,
      ctx.CursorX,
      ctx.CursorY,
      width,
      min,
      max,
      value,
    );
    ctx = this.NextRow(ctx, ctx.Style.SliderHeight);
    return [ctx, result];
  },
  getFontData: function () {
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
};

function main() {
  let w = graphics.CreateWindow("ImGui-like Demo", 1280, 960);
  let ctx = gui.NewContext();
  let showDemo = true;
  let showAnother = false;
  let enabled = true;
  let volume = 0.5;
  let brightness = 75.0;
  let counter = 0;
  let menuState = gui.NewMenuState();
  let demoWin = gui.NewWindowState(20, 45, 350, 400);
  let anotherWin = gui.NewWindowState(400, 45, 350, 200);
  let infoWin = gui.NewWindowState(400, 270, 350, 170);
  let clicked = false;
  let menuOpen = false;
  let dropY = 0;
  let dropX = 0;
  graphics.RunLoop(w, function (w) {
    ctx = gui.UpdateInput(ctx, w);
    graphics.Clear(w, graphics.NewColor(30, 30, 30, 255));
    [ctx, menuState] = gui.BeginMenuBar(
      ctx,
      w,
      menuState,
      0,
      0,
      graphics.GetWidth(w),
    );
    [ctx, menuState, menuOpen] = gui.Menu(ctx, w, menuState, "File");
    if (menuOpen) {
      dropX = menuState.CurrentMenuX - menuState.CurrentMenuW;
      [ctx, dropY] = gui.BeginDropdown(ctx, w, menuState);
      [ctx, menuState, clicked] = gui.MenuItem(
        ctx,
        w,
        menuState,
        "New",
        dropX,
        dropY,
        0,
      );
      if (clicked) {
        counter = 0;
      }
      [ctx, menuState, clicked] = gui.MenuItem(
        ctx,
        w,
        menuState,
        "Open",
        dropX,
        dropY,
        1,
      );
      [ctx, menuState, clicked] = gui.MenuItem(
        ctx,
        w,
        menuState,
        "Save",
        dropX,
        dropY,
        2,
      );
      gui.MenuItemSeparator(ctx, w, dropX, dropY, 3);
      [ctx, menuState, clicked] = gui.MenuItem(
        ctx,
        w,
        menuState,
        "Exit",
        dropX,
        dropY,
        4,
      );
      if (clicked) {
        return false;
      }
    }
    [ctx, menuState, menuOpen] = gui.Menu(ctx, w, menuState, "Edit");
    if (menuOpen) {
      dropX = menuState.CurrentMenuX - menuState.CurrentMenuW;
      [ctx, dropY] = gui.BeginDropdown(ctx, w, menuState);
      [ctx, menuState, clicked] = gui.MenuItem(
        ctx,
        w,
        menuState,
        "Undo",
        dropX,
        dropY,
        0,
      );
      [ctx, menuState, clicked] = gui.MenuItem(
        ctx,
        w,
        menuState,
        "Redo",
        dropX,
        dropY,
        1,
      );
      gui.MenuItemSeparator(ctx, w, dropX, dropY, 2);
      [ctx, menuState, clicked] = gui.MenuItem(
        ctx,
        w,
        menuState,
        "Cut",
        dropX,
        dropY,
        3,
      );
      [ctx, menuState, clicked] = gui.MenuItem(
        ctx,
        w,
        menuState,
        "Copy",
        dropX,
        dropY,
        4,
      );
      [ctx, menuState, clicked] = gui.MenuItem(
        ctx,
        w,
        menuState,
        "Paste",
        dropX,
        dropY,
        5,
      );
    }
    [ctx, menuState, menuOpen] = gui.Menu(ctx, w, menuState, "View");
    if (menuOpen) {
      dropX = menuState.CurrentMenuX - menuState.CurrentMenuW;
      [ctx, dropY] = gui.BeginDropdown(ctx, w, menuState);
      [ctx, menuState, clicked] = gui.MenuItem(
        ctx,
        w,
        menuState,
        "Demo Window",
        dropX,
        dropY,
        0,
      );
      if (clicked) {
        showDemo = !showDemo;
      }
      [ctx, menuState, clicked] = gui.MenuItem(
        ctx,
        w,
        menuState,
        "Another Window",
        dropX,
        dropY,
        1,
      );
      if (clicked) {
        showAnother = !showAnother;
      }
      [ctx, menuState, clicked] = gui.MenuItem(
        ctx,
        w,
        menuState,
        "Info Panel",
        dropX,
        dropY,
        2,
      );
    }
    [ctx, menuState] = gui.EndMenuBar(ctx, menuState);
    [ctx, demoWin] = gui.DraggablePanel(ctx, w, "Demo Window", demoWin);
    ctx = gui.BeginLayout(ctx, demoWin.X + 10, demoWin.Y + 50, 6);
    ctx = gui.AutoLabel(ctx, w, "Hello from goany GUI!");
    gui.Separator(ctx, w, demoWin.X + 10, ctx.CursorY - 2, 330);
    ctx.CursorY = ctx.CursorY + 4;
    [ctx, clicked] = gui.Button(
      ctx,
      w,
      "Click",
      demoWin.X + 10,
      ctx.CursorY,
      80,
      26,
    );
    if (clicked) {
      counter = counter + 1;
    }
    [ctx, clicked] = gui.Button(
      ctx,
      w,
      "Reset",
      demoWin.X + 100,
      ctx.CursorY,
      80,
      26,
    );
    if (clicked) {
      counter = 0;
      volume = 0.5;
      brightness = 75.0;
    }
    gui.Label(
      ctx,
      w,
      "Count: " + intToString(counter),
      demoWin.X + 190,
      ctx.CursorY + 4,
    );
    ctx = gui.NextRow(ctx, 26);
    gui.Separator(ctx, w, demoWin.X + 10, ctx.CursorY - 2, 330);
    ctx.CursorY = ctx.CursorY + 4;
    [ctx, showDemo] = gui.AutoCheckbox(ctx, w, "Show Demo Window", showDemo);
    [ctx, showAnother] = gui.AutoCheckbox(
      ctx,
      w,
      "Show Another Window",
      showAnother,
    );
    [ctx, enabled] = gui.AutoCheckbox(ctx, w, "Enable Feature", enabled);
    gui.Separator(ctx, w, demoWin.X + 10, ctx.CursorY - 2, 330);
    ctx.CursorY = ctx.CursorY + 4;
    [ctx, volume] = gui.AutoSlider(ctx, w, "Volume", 320, 0.0, 1.0, volume);
    [ctx, brightness] = gui.AutoSlider(
      ctx,
      w,
      "Bright",
      320,
      0.0,
      100.0,
      brightness,
    );
    if (showAnother) {
      [ctx, anotherWin] = gui.DraggablePanel(
        ctx,
        w,
        "Another Window",
        anotherWin,
      );
      gui.Label(
        ctx,
        w,
        "This is another panel!",
        anotherWin.X + 10,
        anotherWin.Y + 50,
      );
      [ctx, clicked] = gui.Button(
        ctx,
        w,
        "Close",
        anotherWin.X + 10,
        anotherWin.Y + 90,
        100,
        26,
      );
      if (clicked) {
        showAnother = false;
      }
    }
    [ctx, infoWin] = gui.DraggablePanel(ctx, w, "Info", infoWin);
    ctx = gui.BeginLayout(ctx, infoWin.X + 10, infoWin.Y + 50, 4);
    ctx = gui.AutoLabel(ctx, w, "Application Stats:");
    ctx = gui.AutoLabel(ctx, w, "  Volume: " + floatToString(volume));
    ctx = gui.AutoLabel(ctx, w, "  Brightness: " + floatToString(brightness));
    ctx = gui.AutoLabel(ctx, w, "  Clicks: " + intToString(counter));
    [ctx, clicked] = gui.Button(ctx, w, "Quit", 680, 550, 100, 30);
    if (clicked) {
      return false;
    }
    graphics.Present(w);
    return true;
  });
  graphics.CloseWindow(w);
}

function intToString(n) {
  let digitStrings = ["0", "1", "2", "3", "4", "5", "6", "7", "8", "9"];
  if (n == 0) {
    return "0";
  }
  let negative = false;
  if (n < 0) {
    negative = true;
    n = -n;
  }
  let result = "";
  for (; n > 0; ) {
    let digit = n % 10;
    result = digitStrings[digit] + result;
    n = (n / 10) | 0;
  }
  if (negative) {
    result = "-" + result;
  }
  return result;
}

function floatToString(f) {
  let intPart = int(f);
  let fracPart = int((f - float64(intPart)) * 100);
  if (fracPart < 0) {
    fracPart = -fracPart;
  }
  let fracStr = intToString(fracPart);
  if (fracPart < 10) {
    fracStr = "0" + fracStr;
  }
  return intToString(intPart) + "." + fracStr;
}

// Run main
main();

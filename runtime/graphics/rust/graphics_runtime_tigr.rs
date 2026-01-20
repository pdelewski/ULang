// graphics_runtime_tigr.rs - tigr runtime for goany graphics package
// Requires: tigr.c to be compiled via build.rs
//
// Add to Cargo.toml:
// [build-dependencies]
// cc = "1.0"
//
// Add build.rs:
// fn main() {
//     cc::Build::new()
//         .file("tigr.c")
//         .compile("tigr");
// }

use std::cell::RefCell;
use std::ffi::CString;
use std::os::raw::{c_char, c_int};

// --- FFI bindings for tigr ---

#[repr(C)]
#[derive(Clone, Copy)]
pub struct TPixel {
    pub r: u8,
    pub g: u8,
    pub b: u8,
    pub a: u8,
}

#[repr(C)]
pub struct Tigr {
    pub w: c_int,
    pub h: c_int,
    pub cx: c_int,
    pub cy: c_int,
    pub cw: c_int,
    pub ch: c_int,
    pub pix: *mut TPixel,
    pub handle: *mut std::ffi::c_void,
    pub blitMode: c_int,
}

// tigr key constants
pub const TK_RETURN: c_int = 162;
pub const TK_BACKSPACE: c_int = 161;
pub const TK_ESCAPE: c_int = 175;
pub const TK_SPACE: c_int = 173;
pub const TK_LEFT: c_int = 182;
pub const TK_UP: c_int = 183;
pub const TK_RIGHT: c_int = 184;
pub const TK_DOWN: c_int = 185;
pub const TK_SHIFT: c_int = 163;

extern "C" {
    fn tigrWindow(w: c_int, h: c_int, title: *const c_char, flags: c_int) -> *mut Tigr;
    fn tigrFree(bmp: *mut Tigr);
    fn tigrClosed(bmp: *mut Tigr) -> c_int;
    fn tigrUpdate(bmp: *mut Tigr);
    fn tigrClear(bmp: *mut Tigr, color: TPixel);
    fn tigrPlot(bmp: *mut Tigr, x: c_int, y: c_int, pix: TPixel);
    fn tigrLine(bmp: *mut Tigr, x0: c_int, y0: c_int, x1: c_int, y1: c_int, color: TPixel);
    fn tigrRect(bmp: *mut Tigr, x: c_int, y: c_int, w: c_int, h: c_int, color: TPixel);
    fn tigrFillRect(bmp: *mut Tigr, x: c_int, y: c_int, w: c_int, h: c_int, color: TPixel);
    fn tigrCircle(bmp: *mut Tigr, x: c_int, y: c_int, r: c_int, color: TPixel);
    fn tigrFillCircle(bmp: *mut Tigr, x: c_int, y: c_int, r: c_int, color: TPixel);
    fn tigrReadChar(bmp: *mut Tigr) -> c_int;
    fn tigrKeyDown(bmp: *mut Tigr, key: c_int) -> c_int;
    fn tigrKeyHeld(bmp: *mut Tigr, key: c_int) -> c_int;
}

// --- Thread-local storage ---

thread_local! {
    static LAST_KEY: RefCell<i32> = RefCell::new(0);
}

// --- Public API types ---

#[derive(Clone, Copy)]
pub struct Color {
    pub R: u8,
    pub G: u8,
    pub B: u8,
    pub A: u8,
}

#[derive(Clone, Copy)]
pub struct Rect {
    pub X: i32,
    pub Y: i32,
    pub Width: i32,
    pub Height: i32,
}

#[derive(Clone, Copy)]
pub struct Window {
    pub handle: i64,    // Tigr*
    pub renderer: i64,  // Same as handle for tigr
    pub width: i32,
    pub height: i32,
    pub running: bool,
}

// --- Helper functions ---

fn color_to_tpixel(c: Color) -> TPixel {
    TPixel { r: c.R, g: c.G, b: c.B, a: c.A }
}

// --- Color constructors ---

pub fn NewColor(r: u8, g: u8, b: u8, a: u8) -> Color {
    Color { R: r, G: g, B: b, A: a }
}

pub fn Black() -> Color { Color { R: 0, G: 0, B: 0, A: 255 } }
pub fn White() -> Color { Color { R: 255, G: 255, B: 255, A: 255 } }
pub fn Red() -> Color { Color { R: 255, G: 0, B: 0, A: 255 } }
pub fn Green() -> Color { Color { R: 0, G: 255, B: 0, A: 255 } }
pub fn Blue() -> Color { Color { R: 0, G: 0, B: 255, A: 255 } }

// --- Rect constructor ---

pub fn NewRect(x: i32, y: i32, width: i32, height: i32) -> Rect {
    Rect { X: x, Y: y, Width: width, Height: height }
}

// --- Window management ---

pub fn CreateWindow(title: String, width: i32, height: i32) -> Window {
    let c_title = CString::new(title).unwrap_or_else(|_| CString::new("Window").unwrap());

    let win = unsafe {
        tigrWindow(width as c_int, height as c_int, c_title.as_ptr(), 0)
    };

    if win.is_null() {
        return Window {
            handle: 0,
            renderer: 0,
            width,
            height,
            running: false,
        };
    }

    Window {
        handle: win as i64,
        renderer: win as i64,
        width,
        height,
        running: true,
    }
}

pub fn CloseWindow(w: Window) {
    if w.handle != 0 {
        unsafe {
            tigrFree(w.handle as *mut Tigr);
        }
    }
}

pub fn IsRunning(w: Window) -> bool {
    w.running
}

pub fn PollEvents(mut w: Window) -> (Window, bool) {
    if w.handle == 0 {
        return (w, false);
    }

    let win = w.handle as *mut Tigr;

    // Reset last key
    LAST_KEY.with(|k| *k.borrow_mut() = 0);

    // Use tigrReadChar for character input
    let ch = unsafe { tigrReadChar(win) };
    if ch > 0 && ch < 128 {
        LAST_KEY.with(|k| *k.borrow_mut() = ch);
    }

    // Check special keys that don't produce characters
    LAST_KEY.with(|k| {
        if *k.borrow() == 0 {
            unsafe {
                if tigrKeyDown(win, TK_RETURN) != 0 {
                    *k.borrow_mut() = 13;
                } else if tigrKeyDown(win, TK_BACKSPACE) != 0 {
                    *k.borrow_mut() = 8;
                } else if tigrKeyDown(win, TK_ESCAPE) != 0 {
                    *k.borrow_mut() = 27;
                } else if tigrKeyDown(win, TK_LEFT) != 0 {
                    *k.borrow_mut() = 256;
                } else if tigrKeyDown(win, TK_RIGHT) != 0 {
                    *k.borrow_mut() = 257;
                } else if tigrKeyDown(win, TK_UP) != 0 {
                    *k.borrow_mut() = 258;
                } else if tigrKeyDown(win, TK_DOWN) != 0 {
                    *k.borrow_mut() = 259;
                }
            }
        }
    });

    // Check if window should close
    let closed = unsafe { tigrClosed(win) };
    if closed != 0 {
        w.running = false;
        return (w, false);
    }

    (w, true)
}

pub fn GetLastKey() -> i32 {
    LAST_KEY.with(|k| *k.borrow())
}

pub fn GetWidth(w: Window) -> i32 { w.width }
pub fn GetHeight(w: Window) -> i32 { w.height }

// --- Rendering ---

pub fn Clear(w: Window, c: Color) {
    if w.handle != 0 {
        unsafe {
            tigrClear(w.handle as *mut Tigr, color_to_tpixel(c));
        }
    }
}

pub fn Present(w: Window) {
    if w.handle != 0 {
        unsafe {
            tigrUpdate(w.handle as *mut Tigr);
        }
    }
}

// --- Drawing primitives ---

pub fn DrawRect(w: Window, rect: Rect, c: Color) {
    if w.handle != 0 {
        unsafe {
            tigrRect(
                w.handle as *mut Tigr,
                rect.X as c_int,
                rect.Y as c_int,
                rect.Width as c_int,
                rect.Height as c_int,
                color_to_tpixel(c),
            );
        }
    }
}

pub fn FillRect(w: Window, rect: Rect, c: Color) {
    if w.handle != 0 {
        unsafe {
            tigrFillRect(
                w.handle as *mut Tigr,
                rect.X as c_int,
                rect.Y as c_int,
                rect.Width as c_int,
                rect.Height as c_int,
                color_to_tpixel(c),
            );
        }
    }
}

pub fn DrawLine(w: Window, x1: i32, y1: i32, x2: i32, y2: i32, c: Color) {
    if w.handle != 0 {
        unsafe {
            tigrLine(
                w.handle as *mut Tigr,
                x1 as c_int,
                y1 as c_int,
                x2 as c_int,
                y2 as c_int,
                color_to_tpixel(c),
            );
        }
    }
}

pub fn DrawPoint(w: Window, x: i32, y: i32, c: Color) {
    if w.handle != 0 {
        unsafe {
            tigrPlot(w.handle as *mut Tigr, x as c_int, y as c_int, color_to_tpixel(c));
        }
    }
}

pub fn DrawCircle(w: Window, centerX: i32, centerY: i32, radius: i32, c: Color) {
    if w.handle != 0 {
        unsafe {
            tigrCircle(
                w.handle as *mut Tigr,
                centerX as c_int,
                centerY as c_int,
                radius as c_int,
                color_to_tpixel(c),
            );
        }
    }
}

pub fn FillCircle(w: Window, centerX: i32, centerY: i32, radius: i32, c: Color) {
    if w.handle != 0 {
        unsafe {
            tigrFillCircle(
                w.handle as *mut Tigr,
                centerX as c_int,
                centerY as c_int,
                radius as c_int,
                color_to_tpixel(c),
            );
        }
    }
}

// graphics_runtime_tigr.hpp - tigr runtime for goany graphics package
// This file provides the native implementations for the graphics package using tigr.
// tigr is a tiny graphics library (https://github.com/erkkah/tigr)

#ifndef GRAPHICS_RUNTIME_TIGR_HPP
#define GRAPHICS_RUNTIME_TIGR_HPP

#include "tigr.h"
#include <string>
#include <tuple>
#include <cstdint>

namespace graphics {

struct Color {
    uint8_t R;
    uint8_t G;
    uint8_t B;
    uint8_t A;
};

struct Rect {
    int32_t X;
    int32_t Y;
    int32_t Width;
    int32_t Height;
};

struct Window {
    int64_t handle;   // Tigr*
    int64_t renderer; // Not used with tigr (same as handle)
    int32_t width;
    int32_t height;
    bool running;
};

// --- Color constructors ---

inline Color NewColor(uint8_t r, uint8_t g, uint8_t b, uint8_t a) {
    return Color{r, g, b, a};
}

inline Color Black() { return Color{0, 0, 0, 255}; }
inline Color White() { return Color{255, 255, 255, 255}; }
inline Color Red() { return Color{255, 0, 0, 255}; }
inline Color Green() { return Color{0, 255, 0, 255}; }
inline Color Blue() { return Color{0, 0, 255, 255}; }

// --- Rect constructor ---

inline Rect NewRect(int32_t x, int32_t y, int32_t width, int32_t height) {
    return Rect{x, y, width, height};
}

// --- Helper to convert Color to TPixel ---
namespace detail {
    inline TPixel toTPixel(const Color& c) {
        return tigrRGBA(c.R, c.G, c.B, c.A);
    }
}

// Global to store last key pressed
static int lastKeyPressed = 0;

// --- Window management ---

inline Window CreateWindow(const std::string& title, int32_t width, int32_t height) {
    Tigr* win = tigrWindow(width, height, title.c_str(), TIGR_FIXED);

    if (!win) {
        return Window{0, 0, width, height, false};
    }

    return Window{
        reinterpret_cast<int64_t>(win),
        reinterpret_cast<int64_t>(win),  // renderer same as handle for tigr
        width,
        height,
        true
    };
}

inline void CloseWindow(Window w) {
    if (w.handle) {
        tigrFree(reinterpret_cast<Tigr*>(w.handle));
    }
}

inline bool IsRunning(Window w) {
    return w.running;
}

inline std::tuple<Window, bool> PollEvents(Window w) {
    Tigr* win = reinterpret_cast<Tigr*>(w.handle);

    // Reset last key
    lastKeyPressed = 0;

    // Use tigrReadChar for character input (letters, numbers, space, etc.)
    // This is the primary method and avoids double-detection issues
    int ch = tigrReadChar(win);
    if (ch > 0 && ch < 128) {
        lastKeyPressed = ch;
    }

    // Check special keys that don't produce characters (only if no char was read)
    if (lastKeyPressed == 0) {
        if (tigrKeyDown(win, TK_RETURN)) lastKeyPressed = 13;
        else if (tigrKeyDown(win, TK_BACKSPACE)) lastKeyPressed = 8;
        else if (tigrKeyDown(win, TK_ESCAPE)) lastKeyPressed = 27;
        else if (tigrKeyDown(win, TK_LEFT)) lastKeyPressed = 256;  // Extended key codes
        else if (tigrKeyDown(win, TK_RIGHT)) lastKeyPressed = 257;
        else if (tigrKeyDown(win, TK_UP)) lastKeyPressed = 258;
        else if (tigrKeyDown(win, TK_DOWN)) lastKeyPressed = 259;
    }

    // Check if window should close
    if (tigrClosed(win)) {
        w.running = false;
        return std::make_tuple(w, false);
    }

    return std::make_tuple(w, true);
}

inline int GetLastKey() {
    return lastKeyPressed;
}

inline int32_t GetWidth(Window w) { return w.width; }
inline int32_t GetHeight(Window w) { return w.height; }

// --- Rendering ---

inline void Clear(Window w, Color c) {
    Tigr* win = reinterpret_cast<Tigr*>(w.handle);
    tigrClear(win, detail::toTPixel(c));
}

inline void Present(Window w) {
    Tigr* win = reinterpret_cast<Tigr*>(w.handle);
    tigrUpdate(win);
}

// --- Drawing primitives ---

inline void DrawRect(Window w, Rect rect, Color c) {
    Tigr* win = reinterpret_cast<Tigr*>(w.handle);
    tigrRect(win, rect.X, rect.Y, rect.Width, rect.Height, detail::toTPixel(c));
}

inline void FillRect(Window w, Rect rect, Color c) {
    Tigr* win = reinterpret_cast<Tigr*>(w.handle);
    tigrFillRect(win, rect.X, rect.Y, rect.Width, rect.Height, detail::toTPixel(c));
}

inline void DrawLine(Window w, int32_t x1, int32_t y1, int32_t x2, int32_t y2, Color c) {
    Tigr* win = reinterpret_cast<Tigr*>(w.handle);
    tigrLine(win, x1, y1, x2, y2, detail::toTPixel(c));
}

inline void DrawPoint(Window w, int32_t x, int32_t y, Color c) {
    Tigr* win = reinterpret_cast<Tigr*>(w.handle);
    tigrPlot(win, x, y, detail::toTPixel(c));
}

inline void DrawCircle(Window w, int32_t centerX, int32_t centerY, int32_t radius, Color c) {
    Tigr* win = reinterpret_cast<Tigr*>(w.handle);
    tigrCircle(win, centerX, centerY, radius, detail::toTPixel(c));
}

inline void FillCircle(Window w, int32_t centerX, int32_t centerY, int32_t radius, Color c) {
    Tigr* win = reinterpret_cast<Tigr*>(w.handle);
    tigrFillCircle(win, centerX, centerY, radius, detail::toTPixel(c));
}

} // namespace graphics

#endif // GRAPHICS_RUNTIME_TIGR_HPP

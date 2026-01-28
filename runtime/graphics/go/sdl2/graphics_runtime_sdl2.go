// Package sdl provides the CGO SDL2 implementation for the graphics package.
package sdl

/*
#cgo CFLAGS: -I/opt/homebrew/include -I/usr/include -I/usr/local/include
#cgo LDFLAGS: -L/opt/homebrew/lib -L/usr/lib -L/usr/local/lib -lSDL2
#cgo linux pkg-config: sdl2
#cgo darwin LDFLAGS: -L/opt/homebrew/lib -lSDL2
#cgo windows LDFLAGS: -lSDL2

#include <SDL2/SDL.h>
#include <stdlib.h>

// Helper to create window and renderer
static int createWindowAndRenderer(const char* title, int width, int height,
                                    SDL_Window** window, SDL_Renderer** renderer) {
    if (SDL_Init(SDL_INIT_VIDEO) < 0) {
        return -1;
    }

    *window = SDL_CreateWindow(title,
        SDL_WINDOWPOS_CENTERED, SDL_WINDOWPOS_CENTERED,
        width, height, SDL_WINDOW_SHOWN);
    if (!*window) {
        return -1;
    }

    *renderer = SDL_CreateRenderer(*window, -1,
        SDL_RENDERER_ACCELERATED | SDL_RENDERER_PRESENTVSYNC);
    if (!*renderer) {
        SDL_DestroyWindow(*window);
        return -1;
    }

    return 0;
}

// Global to store last key pressed
static int lastKeyPressed = 0;

// Poll events and return 1 if quit requested
static int pollEvents() {
    SDL_Event event;
    lastKeyPressed = 0;
    while (SDL_PollEvent(&event)) {
        if (event.type == SDL_QUIT) {
            return 1;
        }
        if (event.type == SDL_KEYDOWN) {
            SDL_Keycode key = event.key.keysym.sym;
            // Convert SDL keycode to ASCII for printable characters
            if (key >= SDLK_SPACE && key <= SDLK_z) {
                // Check for shift modifier for uppercase
                if (event.key.keysym.mod & KMOD_SHIFT) {
                    if (key >= SDLK_a && key <= SDLK_z) {
                        lastKeyPressed = key - 32; // Convert to uppercase
                    } else {
                        lastKeyPressed = key;
                    }
                } else {
                    lastKeyPressed = key;
                }
            } else if (key == SDLK_RETURN) {
                lastKeyPressed = 13; // Enter
            } else if (key == SDLK_BACKSPACE) {
                lastKeyPressed = 8; // Backspace
            }
        }
    }
    return 0;
}

// Get the last key pressed (0 if none)
static int getLastKey() {
    return lastKeyPressed;
}

// Draw circle using Bresenham's algorithm
static void drawCircle(SDL_Renderer* renderer, int centerX, int centerY, int radius) {
    int x = radius;
    int y = 0;
    int err = 0;

    while (x >= y) {
        SDL_RenderDrawPoint(renderer, centerX + x, centerY + y);
        SDL_RenderDrawPoint(renderer, centerX + y, centerY + x);
        SDL_RenderDrawPoint(renderer, centerX - y, centerY + x);
        SDL_RenderDrawPoint(renderer, centerX - x, centerY + y);
        SDL_RenderDrawPoint(renderer, centerX - x, centerY - y);
        SDL_RenderDrawPoint(renderer, centerX - y, centerY - x);
        SDL_RenderDrawPoint(renderer, centerX + y, centerY - x);
        SDL_RenderDrawPoint(renderer, centerX + x, centerY - y);

        y++;
        err += 1 + 2 * y;
        if (2 * (err - x) + 1 > 0) {
            x--;
            err += 1 - 2 * x;
        }
    }
}

// Fill circle
static void fillCircle(SDL_Renderer* renderer, int centerX, int centerY, int radius) {
    for (int y = -radius; y <= radius; y++) {
        for (int x = -radius; x <= radius; x++) {
            if (x * x + y * y <= radius * radius) {
                SDL_RenderDrawPoint(renderer, centerX + x, centerY + y);
            }
        }
    }
}
*/
import "C"
import (
	"runtime"
	"unsafe"
)

func init() {
	// On macOS, SDL2 requires event handling on the main thread.
	// This locks the main goroutine to the OS main thread.
	runtime.LockOSThread()
}

// CreateWindow creates a new window with the specified title and dimensions.
// Returns handle, renderer, success
func CreateWindow(title string, width int32, height int32) (int64, int64, bool) {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))

	var window *C.SDL_Window
	var renderer *C.SDL_Renderer

	result := C.createWindowAndRenderer(cTitle, C.int(width), C.int(height), &window, &renderer)
	if result < 0 {
		return 0, 0, false
	}

	return int64(uintptr(unsafe.Pointer(window))),
		int64(uintptr(unsafe.Pointer(renderer))),
		true
}

// CloseWindow closes the window and releases resources.
func CloseWindow(handle int64, renderer int64) {
	if renderer != 0 {
		C.SDL_DestroyRenderer((*C.SDL_Renderer)(unsafe.Pointer(uintptr(renderer))))
	}
	if handle != 0 {
		C.SDL_DestroyWindow((*C.SDL_Window)(unsafe.Pointer(uintptr(handle))))
	}
	C.SDL_Quit()
}

// PollEvents processes pending events and returns true if quit requested.
// The handle parameter is ignored for SDL2 (uses global event queue).
func PollEvents(handle int64) bool {
	return C.pollEvents() != 0
}

// GetLastKey returns the ASCII code of the last key pressed (0 if none).
func GetLastKey() int {
	return int(C.getLastKey())
}

// Clear clears the window with the specified color.
func Clear(renderer int64, r uint8, g uint8, b uint8, a uint8) {
	rend := (*C.SDL_Renderer)(unsafe.Pointer(uintptr(renderer)))
	C.SDL_SetRenderDrawColor(rend, C.Uint8(r), C.Uint8(g), C.Uint8(b), C.Uint8(a))
	C.SDL_RenderClear(rend)
}

// Present presents the rendered content to the screen.
func Present(renderer int64) {
	rend := (*C.SDL_Renderer)(unsafe.Pointer(uintptr(renderer)))
	C.SDL_RenderPresent(rend)
}

// DrawRect draws a rectangle outline.
func DrawRect(renderer int64, x int32, y int32, width int32, height int32, r uint8, g uint8, b uint8, a uint8) {
	rend := (*C.SDL_Renderer)(unsafe.Pointer(uintptr(renderer)))
	C.SDL_SetRenderDrawColor(rend, C.Uint8(r), C.Uint8(g), C.Uint8(b), C.Uint8(a))
	sdlRect := C.SDL_Rect{
		x: C.int(x),
		y: C.int(y),
		w: C.int(width),
		h: C.int(height),
	}
	C.SDL_RenderDrawRect(rend, &sdlRect)
}

// FillRect draws a filled rectangle.
func FillRect(renderer int64, x int32, y int32, width int32, height int32, r uint8, g uint8, b uint8, a uint8) {
	rend := (*C.SDL_Renderer)(unsafe.Pointer(uintptr(renderer)))
	C.SDL_SetRenderDrawColor(rend, C.Uint8(r), C.Uint8(g), C.Uint8(b), C.Uint8(a))
	sdlRect := C.SDL_Rect{
		x: C.int(x),
		y: C.int(y),
		w: C.int(width),
		h: C.int(height),
	}
	C.SDL_RenderFillRect(rend, &sdlRect)
}

// DrawLine draws a line between two points.
func DrawLine(renderer int64, x1 int32, y1 int32, x2 int32, y2 int32, r uint8, g uint8, b uint8, a uint8) {
	rend := (*C.SDL_Renderer)(unsafe.Pointer(uintptr(renderer)))
	C.SDL_SetRenderDrawColor(rend, C.Uint8(r), C.Uint8(g), C.Uint8(b), C.Uint8(a))
	C.SDL_RenderDrawLine(rend, C.int(x1), C.int(y1), C.int(x2), C.int(y2))
}

// DrawPoint draws a single pixel.
func DrawPoint(renderer int64, x int32, y int32, r uint8, g uint8, b uint8, a uint8) {
	rend := (*C.SDL_Renderer)(unsafe.Pointer(uintptr(renderer)))
	C.SDL_SetRenderDrawColor(rend, C.Uint8(r), C.Uint8(g), C.Uint8(b), C.Uint8(a))
	C.SDL_RenderDrawPoint(rend, C.int(x), C.int(y))
}

// DrawCircle draws a circle outline.
func DrawCircle(renderer int64, centerX int32, centerY int32, radius int32, r uint8, g uint8, b uint8, a uint8) {
	rend := (*C.SDL_Renderer)(unsafe.Pointer(uintptr(renderer)))
	C.SDL_SetRenderDrawColor(rend, C.Uint8(r), C.Uint8(g), C.Uint8(b), C.Uint8(a))
	C.drawCircle(rend, C.int(centerX), C.int(centerY), C.int(radius))
}

// FillCircle draws a filled circle.
func FillCircle(renderer int64, centerX int32, centerY int32, radius int32, r uint8, g uint8, b uint8, a uint8) {
	rend := (*C.SDL_Renderer)(unsafe.Pointer(uintptr(renderer)))
	C.SDL_SetRenderDrawColor(rend, C.Uint8(r), C.Uint8(g), C.Uint8(b), C.Uint8(a))
	C.fillCircle(rend, C.int(centerX), C.int(centerY), C.int(radius))
}

// GetMouse returns mouse position and button state.
// Returns: x, y, buttons (bit 0=left, bit 1=right, bit 2=middle)
func GetMouse(handle int64) (int32, int32, int32) {
	var x, y C.int
	buttons := C.SDL_GetMouseState(&x, &y)
	// SDL_BUTTON(X) = (1 << ((X)-1)), so:
	// SDL_BUTTON(1) = 1 (left), SDL_BUTTON(2) = 2 (middle), SDL_BUTTON(3) = 4 (right)
	// Convert to our format: left=1, right=2, middle=4
	result := int32(0)
	if buttons&1 != 0 { // SDL_BUTTON_LEFT
		result |= 1
	}
	if buttons&4 != 0 { // SDL_BUTTON_RIGHT
		result |= 2
	}
	if buttons&2 != 0 { // SDL_BUTTON_MIDDLE
		result |= 4
	}
	return int32(x), int32(y), result
}

// GetScreenSize returns the screen resolution using SDL2's cross-platform API.
func GetScreenSize() (int32, int32) {
	var mode C.SDL_DisplayMode
	if C.SDL_GetCurrentDisplayMode(0, &mode) == 0 {
		return int32(mode.w), int32(mode.h)
	}
	return 1920, 1080 // fallback
}

// CreateWindowFullscreen creates a fullscreen window.
func CreateWindowFullscreen(title string, width int32, height int32) (int64, int64, bool) {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))

	if C.SDL_Init(C.SDL_INIT_VIDEO) < 0 {
		return 0, 0, false
	}

	window := C.SDL_CreateWindow(cTitle,
		C.SDL_WINDOWPOS_CENTERED, C.SDL_WINDOWPOS_CENTERED,
		C.int(width), C.int(height),
		C.SDL_WINDOW_SHOWN|C.SDL_WINDOW_FULLSCREEN_DESKTOP)
	if window == nil {
		return 0, 0, false
	}

	renderer := C.SDL_CreateRenderer(window, -1,
		C.SDL_RENDERER_ACCELERATED|C.SDL_RENDERER_PRESENTVSYNC)
	if renderer == nil {
		C.SDL_DestroyWindow(window)
		return 0, 0, false
	}

	return int64(uintptr(unsafe.Pointer(window))),
		int64(uintptr(unsafe.Pointer(renderer))),
		true
}

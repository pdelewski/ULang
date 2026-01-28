package main

import (
	"runtime/graphics"
	"runtime/gui"
)

func main() {
	w := graphics.CreateWindow("GUI Demo", 640, 480)
	ctx := gui.NewContext()

	// State variables
	darkMode := false
	volume := 0.5
	brightness := 75.0
	clickCount := 0

	graphics.RunLoop(w, func(w graphics.Window) bool {
		// Update input first
		gui.UpdateInput(&ctx, w)

		// Clear background
		var bgColor graphics.Color
		if darkMode {
			bgColor = graphics.NewColor(20, 20, 30, 255)
		} else {
			bgColor = graphics.NewColor(60, 60, 70, 255)
		}
		graphics.Clear(w, bgColor)

		// Manual positioning example
		gui.Label(&ctx, w, "GUI Demo - Manual Layout", 20, 20)

		if gui.Button(&ctx, w, "Click Me", 20, 50, 120, 28) {
			clickCount = clickCount + 1
		}

		// Show click count
		countText := "Clicks: " + intToString(clickCount)
		gui.Label(&ctx, w, countText, 160, 58)

		darkMode = gui.Checkbox(&ctx, w, "Dark Mode", 20, 100, darkMode)

		volume = gui.Slider(&ctx, w, "Volume", 20, 140, 200, 0.0, 1.0, volume)
		brightness = gui.Slider(&ctx, w, "Brightness", 20, 200, 200, 0.0, 100.0, brightness)

		// Auto-layout example
		gui.BeginLayout(&ctx, 300, 50, 8)

		gui.AutoLabel(&ctx, w, "Auto Layout Section")

		if gui.AutoButton(&ctx, w, "Reset", 100, 28) {
			volume = 0.5
			brightness = 75.0
			clickCount = 0
		}

		if gui.AutoButton(&ctx, w, "Quit", 100, 28) {
			return false
		}

		// Draw some visual feedback
		graphics.FillRect(w, graphics.NewRect(300, 200, int32(volume*200), 20), graphics.NewColor(100, 200, 100, 255))
		graphics.DrawRect(w, graphics.NewRect(300, 200, 200, 20), graphics.White())
		gui.Label(&ctx, w, "Volume Bar", 300, 225)

		graphics.FillRect(w, graphics.NewRect(300, 260, int32(brightness*2), 20), graphics.NewColor(200, 200, 100, 255))
		graphics.DrawRect(w, graphics.NewRect(300, 260, 200, 20), graphics.White())
		gui.Label(&ctx, w, "Brightness Bar", 300, 285)

		graphics.Present(w)
		return true
	})

	graphics.CloseWindow(w)
}

// intToString converts an integer to a string (simple implementation)
func intToString(n int) string {
	if n == 0 {
		return "0"
	}
	negative := false
	if n < 0 {
		negative = true
		n = -n
	}
	result := ""
	for n > 0 {
		digit := n % 10
		result = string(rune('0'+digit)) + result
		n = n / 10
	}
	if negative {
		result = "-" + result
	}
	return result
}

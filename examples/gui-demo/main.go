package main

import (
	"libs/gui"
	"runtime/graphics"
)

func main() {
	w := graphics.CreateWindow("ImGui-like Demo", 800, 600)
	ctx := gui.NewContext()

	// State variables
	showDemo := true
	showAnother := false
	enabled := true
	volume := 0.5
	brightness := 75.0
	counter := 0

	var clicked bool

	graphics.RunLoop(w, func(w graphics.Window) bool {
		// Update input first
		ctx = gui.UpdateInput(ctx, w)

		// Clear with very dark background (like ImGui demo)
		graphics.Clear(w, graphics.NewColor(30, 30, 30, 255))

		// Main demo panel
		gui.Panel(ctx, w, "Demo Window", 20, 20, 350, 400)

		// Content inside panel
		ctx = gui.BeginLayout(ctx, 30, 70, 6)

		ctx = gui.AutoLabel(ctx, w, "Hello from goany GUI!")

		gui.Separator(ctx, w, 30, ctx.CursorY-2, 330)
		ctx.CursorY = ctx.CursorY + 4

		// Buttons in a row
		ctx, clicked = gui.Button(ctx, w, "Click", 30, ctx.CursorY, 80, 26)
		if clicked {
			counter = counter + 1
		}
		// Same row button
		ctx, clicked = gui.Button(ctx, w, "Reset", 120, ctx.CursorY, 80, 26)
		if clicked {
			counter = 0
			volume = 0.5
			brightness = 75.0
		}
		gui.Label(ctx, w, "Count: "+intToString(counter), 210, ctx.CursorY+4)
		ctx = gui.NextRow(ctx, 26)

		gui.Separator(ctx, w, 30, ctx.CursorY-2, 330)
		ctx.CursorY = ctx.CursorY + 4

		// Checkboxes
		ctx, showDemo = gui.AutoCheckbox(ctx, w, "Show Demo Window", showDemo)
		ctx, showAnother = gui.AutoCheckbox(ctx, w, "Show Another Window", showAnother)
		ctx, enabled = gui.AutoCheckbox(ctx, w, "Enable Feature", enabled)

		gui.Separator(ctx, w, 30, ctx.CursorY-2, 330)
		ctx.CursorY = ctx.CursorY + 4

		// Sliders
		ctx, volume = gui.AutoSlider(ctx, w, "Volume", 320, 0.0, 1.0, volume)
		ctx, brightness = gui.AutoSlider(ctx, w, "Bright", 320, 0.0, 100.0, brightness)

		// Second panel if enabled
		if showAnother {
			gui.Panel(ctx, w, "Another Window", 400, 20, 350, 200)
			gui.Label(ctx, w, "This is another panel!", 410, 70)
			ctx, clicked = gui.Button(ctx, w, "Close", 410, 110, 100, 26)
			if clicked {
				showAnother = false
			}
		}

		// Info panel
		gui.Panel(ctx, w, "Info", 400, 250, 350, 170)
		ctx = gui.BeginLayout(ctx, 410, 300, 4)
		ctx = gui.AutoLabel(ctx, w, "Application Stats:")
		ctx = gui.AutoLabel(ctx, w, "  Volume: "+floatToString(volume))
		ctx = gui.AutoLabel(ctx, w, "  Brightness: "+floatToString(brightness))
		ctx = gui.AutoLabel(ctx, w, "  Clicks: "+intToString(counter))

		// Quit button at bottom
		ctx, clicked = gui.Button(ctx, w, "Quit", 680, 550, 100, 30)
		if clicked {
			return false
		}

		graphics.Present(w)
		return true
	})

	graphics.CloseWindow(w)
}

// intToString converts an integer to a string
func intToString(n int) string {
	digitStrings := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
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
		result = digitStrings[digit] + result
		n = n / 10
	}
	if negative {
		result = "-" + result
	}
	return result
}

// floatToString converts a float to a string with 2 decimal places
func floatToString(f float64) string {
	// Integer part
	intPart := int(f)
	// Fractional part (2 decimals)
	fracPart := int((f - float64(intPart)) * 100)
	if fracPart < 0 {
		fracPart = -fracPart
	}
	fracStr := intToString(fracPart)
	if fracPart < 10 {
		fracStr = "0" + fracStr
	}
	return intToString(intPart) + "." + fracStr
}

package main

import (
	"libs/gui"
	"runtime/graphics"
)

func main() {
	w := graphics.CreateWindow("ImGui-like Demo", 1280, 960)
	ctx := gui.NewContext()

	// State variables
	showDemo := true
	showAnother := false
	enabled := true
	volume := 0.5
	brightness := 75.0
	counter := 0

	// Menu state
	menuState := gui.NewMenuState()

	// Draggable window states (offset by menu bar height ~20px)
	demoWin := gui.NewWindowState(20, 45, 350, 400)
	anotherWin := gui.NewWindowState(400, 45, 350, 200)
	infoWin := gui.NewWindowState(400, 270, 350, 170)

	var clicked bool
	var menuOpen bool
	var dropY int32
	var dropX int32

	graphics.RunLoop(w, func(w graphics.Window) bool {
		// Update input first
		ctx = gui.UpdateInput(ctx, w)

		// Clear with very dark background (like ImGui demo)
		graphics.Clear(w, graphics.NewColor(30, 30, 30, 255))

		// Menu bar at top (use window width for full-width menu bar)
		ctx, menuState = gui.BeginMenuBar(ctx, w, menuState, 0, 0, graphics.GetWidth(w))

		// File menu
		ctx, menuState, menuOpen = gui.Menu(ctx, w, menuState, "File")
		if menuOpen {
			dropX = menuState.CurrentMenuX - menuState.CurrentMenuW
			ctx, dropY = gui.BeginDropdown(ctx, w, menuState)
			ctx, menuState, clicked = gui.MenuItem(ctx, w, menuState, "New", dropX, dropY, 0)
			if clicked {
				counter = 0
			}
			ctx, menuState, clicked = gui.MenuItem(ctx, w, menuState, "Open", dropX, dropY, 1)
			ctx, menuState, clicked = gui.MenuItem(ctx, w, menuState, "Save", dropX, dropY, 2)
			gui.MenuItemSeparator(ctx, w, dropX, dropY, 3)
			ctx, menuState, clicked = gui.MenuItem(ctx, w, menuState, "Exit", dropX, dropY, 4)
			if clicked {
				return false
			}
		}

		// Edit menu
		ctx, menuState, menuOpen = gui.Menu(ctx, w, menuState, "Edit")
		if menuOpen {
			dropX = menuState.CurrentMenuX - menuState.CurrentMenuW
			ctx, dropY = gui.BeginDropdown(ctx, w, menuState)
			ctx, menuState, clicked = gui.MenuItem(ctx, w, menuState, "Undo", dropX, dropY, 0)
			ctx, menuState, clicked = gui.MenuItem(ctx, w, menuState, "Redo", dropX, dropY, 1)
			gui.MenuItemSeparator(ctx, w, dropX, dropY, 2)
			ctx, menuState, clicked = gui.MenuItem(ctx, w, menuState, "Cut", dropX, dropY, 3)
			ctx, menuState, clicked = gui.MenuItem(ctx, w, menuState, "Copy", dropX, dropY, 4)
			ctx, menuState, clicked = gui.MenuItem(ctx, w, menuState, "Paste", dropX, dropY, 5)
		}

		// View menu
		ctx, menuState, menuOpen = gui.Menu(ctx, w, menuState, "View")
		if menuOpen {
			dropX = menuState.CurrentMenuX - menuState.CurrentMenuW
			ctx, dropY = gui.BeginDropdown(ctx, w, menuState)
			ctx, menuState, clicked = gui.MenuItem(ctx, w, menuState, "Demo Window", dropX, dropY, 0)
			if clicked {
				showDemo = !showDemo
			}
			ctx, menuState, clicked = gui.MenuItem(ctx, w, menuState, "Another Window", dropX, dropY, 1)
			if clicked {
				showAnother = !showAnother
			}
			ctx, menuState, clicked = gui.MenuItem(ctx, w, menuState, "Info Panel", dropX, dropY, 2)
		}

		ctx, menuState = gui.EndMenuBar(ctx, menuState)

		// Main demo panel (draggable)
		ctx, demoWin = gui.DraggablePanel(ctx, w, "Demo Window", demoWin)

		// Content inside panel (relative to panel position)
		ctx = gui.BeginLayout(ctx, demoWin.X+10, demoWin.Y+50, 6)

		ctx = gui.AutoLabel(ctx, w, "Hello from goany GUI!")

		gui.Separator(ctx, w, demoWin.X+10, ctx.CursorY-2, 330)
		ctx.CursorY = ctx.CursorY + 4

		// Buttons in a row
		ctx, clicked = gui.Button(ctx, w, "Click", demoWin.X+10, ctx.CursorY, 80, 26)
		if clicked {
			counter = counter + 1
		}
		// Same row button
		ctx, clicked = gui.Button(ctx, w, "Reset", demoWin.X+100, ctx.CursorY, 80, 26)
		if clicked {
			counter = 0
			volume = 0.5
			brightness = 75.0
		}
		gui.Label(ctx, w, "Count: "+intToString(counter), demoWin.X+190, ctx.CursorY+4)
		ctx = gui.NextRow(ctx, 26)

		gui.Separator(ctx, w, demoWin.X+10, ctx.CursorY-2, 330)
		ctx.CursorY = ctx.CursorY + 4

		// Checkboxes
		ctx, showDemo = gui.AutoCheckbox(ctx, w, "Show Demo Window", showDemo)
		ctx, showAnother = gui.AutoCheckbox(ctx, w, "Show Another Window", showAnother)
		ctx, enabled = gui.AutoCheckbox(ctx, w, "Enable Feature", enabled)

		gui.Separator(ctx, w, demoWin.X+10, ctx.CursorY-2, 330)
		ctx.CursorY = ctx.CursorY + 4

		// Sliders
		ctx, volume = gui.AutoSlider(ctx, w, "Volume", 320, 0.0, 1.0, volume)
		ctx, brightness = gui.AutoSlider(ctx, w, "Bright", 320, 0.0, 100.0, brightness)

		// Second panel if enabled (draggable)
		if showAnother {
			ctx, anotherWin = gui.DraggablePanel(ctx, w, "Another Window", anotherWin)
			gui.Label(ctx, w, "This is another panel!", anotherWin.X+10, anotherWin.Y+50)
			ctx, clicked = gui.Button(ctx, w, "Close", anotherWin.X+10, anotherWin.Y+90, 100, 26)
			if clicked {
				showAnother = false
			}
		}

		// Info panel (draggable)
		ctx, infoWin = gui.DraggablePanel(ctx, w, "Info", infoWin)
		ctx = gui.BeginLayout(ctx, infoWin.X+10, infoWin.Y+50, 4)
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

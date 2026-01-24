package main

import "runtime/graphics"

func main() {
	// Create a window
	w := graphics.CreateWindow("goany Graphics Demo", 800, 600)

	// Run main loop with frame callback
	graphics.RunLoop(w, func(w graphics.Window) bool {
		// Clear screen with dark blue
		graphics.Clear(w, graphics.NewColor(20, 20, 40, 255))

		// Draw some shapes

		// Red filled rectangle
		graphics.FillRect(w, graphics.NewRect(50, 50, 200, 150), graphics.Red())

		// Green rectangle outline
		graphics.DrawRect(w, graphics.NewRect(300, 50, 200, 150), graphics.Green())

		// Blue filled circle
		graphics.FillCircle(w, 150, 400, 80, graphics.Blue())

		// White circle outline
		graphics.DrawCircle(w, 400, 400, 80, graphics.White())

		// Yellow lines
		graphics.DrawLine(w, 550, 50, 750, 200, graphics.NewColor(255, 255, 0, 255))
		graphics.DrawLine(w, 550, 200, 750, 50, graphics.NewColor(255, 255, 0, 255))

		// Draw some points (small dots)
		x := int32(600)
		for {
			if x >= 700 {
				break
			}
			graphics.DrawPoint(w, x, 300, graphics.White())
			x = x + 5
		}

		// Present the frame
		graphics.Present(w)

		return true // Continue running
	})

	// Cleanup
	graphics.CloseWindow(w)
}

package main

import "runtime/graphics"

func main() {
	// Create a window
	w := graphics.CreateWindow("goany Graphics Demo", 800, 600)

	// Main loop
	running := true
	for running {
		// Poll events
		w, running = graphics.PollEvents(w)

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
		var x int32
		x = 600
		for x < 700 {
			graphics.DrawPoint(w, x, 300, graphics.White())
			x = x + 5
		}

		// Present the frame
		graphics.Present(w)
	}

	// Cleanup
	graphics.CloseWindow(w)
}

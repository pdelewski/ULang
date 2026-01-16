package main

import "runtime/graphics"

func main() {
		w := graphics.CreateWindow("Minimal", 400, 300)
		graphics.CloseWindow(w)
}

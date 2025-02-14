package main

import "fmt"

func main() {
	list := NewList()

	// Add some elements to the list
	list = Add(list, 10)
	list = Add(list, 20)
	list = Add(list, 30)
	list = Add(list, 40)

	fmt.Println("List after adding elements:")
	Print(list) // Output: 10 -> 20 -> 30 -> 40 -> nil

	// Remove an element from the list
	list = Remove(list, 20)
	fmt.Println("List after removing 20:")
	Print(list) // Output: 10 -> 30 -> 40 -> nil

	// Remove the head element
	list = Remove(list, 10)
	fmt.Println("List after removing 10:")
	Print(list) // Output: 30 -> 40 -> nil
}

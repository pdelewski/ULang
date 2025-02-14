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

	tree := NewBinaryTree()

	// Insert some elements into the binary tree
	tree = Insert(tree, 10)
	tree = Insert(tree, 20)
	tree = Insert(tree, 30)
	tree = Insert(tree, 40)
	tree = Insert(tree, 50)

	fmt.Println("Binary tree after inserting elements:")
	PrintTree(tree) // Expected output: 10 20 30 40 50

	// Remove an element from the tree
	tree = RemoveFromTree(tree, 30)
	fmt.Println("Binary tree after removing 30:")
	PrintTree(tree) // Expected output: 10 20 50 40

	// Remove the root element
	tree = RemoveFromTree(tree, 10)
	fmt.Println("Binary tree after removing 10:")
	PrintTree(tree) // Expected output: 40 20 50
}

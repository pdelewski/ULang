package main

import "fmt"
import "lib/ulangstd"

func main() {
	list := ulangstd.NewList()

	// Add some elements to the list
	list = ulangstd.Add(list, 10)
	list = ulangstd.Add(list, 20)
	list = ulangstd.Add(list, 30)
	list = ulangstd.Add(list, 40)

	fmt.Println("List after adding elements:")
	ulangstd.PrintList(list) // Output: 10 -> 20 -> 30 -> 40 -> nil

	// Remove an element from the list
	list = ulangstd.Remove(list, 20)
	fmt.Println("List after removing 20:")
	ulangstd.PrintList(list) // Output: 10 -> 30 -> 40 -> nil

	// Remove the head element
	list = ulangstd.Remove(list, 10)
	fmt.Println("List after removing 10:")
	ulangstd.PrintList(list) // Output: 30 -> 40 -> nil

	tree := ulangstd.NewBinaryTree()

	// Insert some elements into the binary tree
	tree = ulangstd.Insert(tree, 10)
	tree = ulangstd.Insert(tree, 20)
	tree = ulangstd.Insert(tree, 30)
	tree = ulangstd.Insert(tree, 40)
	tree = ulangstd.Insert(tree, 50)

	fmt.Println("Binary tree after inserting elements:")
	ulangstd.PrintTree(tree) // Expected output: 10 20 30 40 50

	// Remove an element from the tree
	tree = ulangstd.RemoveFromTree(tree, 30)
	fmt.Println("Binary tree after removing 30:")
	ulangstd.PrintTree(tree) // Expected output: 10 20 50 40

	// Remove the root element
	tree = ulangstd.RemoveFromTree(tree, 10)
	fmt.Println("Binary tree after removing 10:")
	ulangstd.PrintTree(tree) // Expected output: 40 20 50
}

package main

import "fmt"
import "containers_tests/containers"

func main() {
	list := containers.NewList()

	// Add some elements to the list
	list = containers.Add(list, 10)
	list = containers.Add(list, 20)
	list = containers.Add(list, 30)
	list = containers.Add(list, 40)

	fmt.Println("List after adding elements:")
	containers.PrintList(list) // Output: 10 -> 20 -> 30 -> 40 -> nil

	// Remove an element from the list
	list = containers.Remove(list, 20)
	fmt.Println("List after removing 20:")
	containers.PrintList(list) // Output: 10 -> 30 -> 40 -> nil

	// Remove the head element
	list = containers.Remove(list, 10)
	fmt.Println("List after removing 10:")
	containers.PrintList(list) // Output: 30 -> 40 -> nil

	tree := containers.NewBinaryTree()

	// Insert some elements into the binary tree
	tree = containers.Insert(tree, 10)
	tree = containers.Insert(tree, 20)
	tree = containers.Insert(tree, 30)
	tree = containers.Insert(tree, 40)
	tree = containers.Insert(tree, 50)

	fmt.Println("Binary tree after inserting elements:")
	containers.PrintTree(tree) // Expected output: 10 20 30 40 50

	// Remove an element from the tree
	tree = containers.RemoveFromTree(tree, 30)
	fmt.Println("Binary tree after removing 30:")
	containers.PrintTree(tree) // Expected output: 10 20 50 40

	// Remove the root element
	tree = containers.RemoveFromTree(tree, 10)
	fmt.Println("Binary tree after removing 10:")
	containers.PrintTree(tree) // Expected output: 40 20 50
}

package containers

import "fmt"

// BinaryTreeNode represents a node in the binary tree using an array-based approach
type BinaryTreeNode struct {
	value int // The value of the node
	left  int // The index of the left child (-1 if no left child)
	right int // The index of the right child (-1 if no right child)
}

// BinaryTree represents the array-based binary tree
type BinaryTree struct {
	nodes []BinaryTreeNode // The array storing the tree nodes
	root  int              // The index of the root node (-1 if the tree is empty)
}

// NewBinaryTree creates a new empty binary tree and returns it as a value
func NewBinaryTree() BinaryTree {
	return BinaryTree{
		nodes: []BinaryTreeNode{}, // Initialize with an empty slice of nodes
		root:  -1,                 // -1 indicates the tree is empty
	}
}

// Insert inserts a value into the binary tree, maintaining a complete binary tree, and returns the modified tree
func Insert(t BinaryTree, value int) BinaryTree {
	newNode := BinaryTreeNode{
		value: value,
		left:  -1, // No left child
		right: -1, // No right child
	}

	// Add the new node to the array
	t.nodes = append(t.nodes, newNode)

	// If the tree is empty, set the root to the new node's index
	if t.root == -1 {
		t.root = 0
	} else {
		// Otherwise, insert the node as a child of the existing tree
		insertIndex := len(t.nodes) - 1
		parentIndex := (insertIndex - 1) / 2 // Find parent index

		if insertIndex%2 == 1 {
			// Left child
			// TODO original code  code was about mutating this in-place
			// this is problematic now from c# perspective
			// as it causes Cannot modify the return value of 'List<Api.ListNode>.this[int]' because it is not a variable
			// to fix that we have to discover this case and transform code as below
			tmp := t.nodes[parentIndex]
			tmp.left = insertIndex
			t.nodes[parentIndex] = tmp
		} else {
			// Right child
			// TODO original code  code was about mutating this in-place
			// this is problematic now from c# perspective
			// as it causes Cannot modify the return value of 'List<Api.ListNode>.this[int]' because it is not a variable
			// to fix that we have to discover this case and transform code as below
			tmp := t.nodes[parentIndex]
			tmp.right = insertIndex // Update the parent's right child index
			t.nodes[parentIndex] = tmp
		}
	}

	return t // Return the updated tree
}

// Print prints the binary tree in a level-order traversal
func PrintTree(t BinaryTree) {
	if t.root == -1 {
		fmt.Println("The tree is empty.")
		return
	}

	// Start from the root and print nodes in level-order (breadth-first)
	queue := []int{t.root}
	for len(queue) > 0 {
		index := queue[0]
		queue = queue[1:]
		node := t.nodes[index]
		fmt.Printf("%d:", node.value)

		if node.left != -1 {
			queue = append(queue, node.left)
		}
		if node.right != -1 {
			queue = append(queue, node.right)
		}
	}
	fmt.Println()
}

// Remove removes a node by value in the binary tree and returns the modified tree
func RemoveFromTree(t BinaryTree, value int) BinaryTree {
	/*
		if t.root == -1 {
			fmt.Println("The tree is empty.")
			return t
		}

		// Find the index of the node to remove
		indexToRemove := -1
		for i := 0; i < len(t.nodes); i++ {
			if t.nodes[i].value == value {
				indexToRemove = i
				break
			}
		}

		if indexToRemove == -1 {
			fmt.Println("Value not found in the tree.")
			return t
		}

		// Swap the node to be removed with the last node in the array
		lastNodeIndex := len(t.nodes) - 1
		t.nodes[indexToRemove], t.nodes[lastNodeIndex] = t.nodes[lastNodeIndex], t.nodes[indexToRemove]

		// Remove the last node from the array
		t.nodes = t.nodes[:lastNodeIndex]

		// If the removed node had children, we need to fix the parent-child relationships
		if indexToRemove < len(t.nodes) {
			parentIndex := (indexToRemove - 1) / 2
			if indexToRemove%2 == 1 {
				// Update parent's left child index
				t.nodes[parentIndex].left = indexToRemove
			} else {
				// Update parent's right child index
				t.nodes[parentIndex].right = indexToRemove
			}
		}
	*/
	return t
}

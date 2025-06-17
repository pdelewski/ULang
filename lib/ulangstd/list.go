package ulangstd

import "fmt"

// ListNode represents a node in the list using an array-based approach
type ListNode struct {
	value int // The value of the node
	next  int // The index of the next node (-1 if no next node)
}

// List represents the array-based list
type List struct {
	nodes []ListNode // The array storing the list nodes
	head  int        // The index of the head node (-1 if the list is empty)
}

// NewList creates a new empty list and returns it as a value
func NewList() List {
	return List{
		nodes: []ListNode{}, // Initialize with an empty slice of nodes
		head:  -1,           // -1 indicates the list is empty
	}
}

// Add adds a new value to the end of the list and returns the modified list
func Add(l List, value int) List {
	newNode := ListNode{
		value: value,
		next:  -1, // No next node as it will be the last one
	}

	// Add the new node to the array
	l.nodes = append(l.nodes, newNode)

	// If the list is empty, set the head to the new node's index
	if l.head == -1 {
		l.head = len(l.nodes) - 1
	} else {
		// Otherwise, find the last node and link it to the new node
		lastNodeIndex := l.head
		for l.nodes[lastNodeIndex].next != -1 {
			lastNodeIndex = l.nodes[lastNodeIndex].next
		}
		// TODO original code was about mutating this in-place l.nodes[lastNodeIndex].next = len(l.nodes) - 1
		// this is problematic now from c# perspective
		// as it causes Cannot modify the return value of 'List<Api.ListNode>.this[int]' because it is not a variable
		// to fix that we have to discover this case and transform code as below
		tmp := l.nodes[lastNodeIndex]
		tmp.next = len(l.nodes) - 1 // Set the next of the last node to the new node's index
		l.nodes[lastNodeIndex] = tmp
	}

	return l // Return the updated list
}

// Remove removes the first occurrence of a value in the list and returns the modified list
func Remove(l List, value int) List {
	if l.head == -1 {
		fmt.Println("The list is empty.")
		return l
	}

	// If the head node is the one to be removed
	if l.nodes[l.head].value == value {
		l.head = l.nodes[l.head].next
		return l
	}

	prevIndex := -1
	currIndex := l.head
	for currIndex != -1 {
		if l.nodes[currIndex].value == value {
			if prevIndex == -1 {
				l.head = l.nodes[currIndex].next
			} else {
				// TODO original code  code was about mutating this in-place l.nodes[currIndex].next = l.nodes[prevIndex]
				// this is problematic now from c# perspective
				// as it causes Cannot modify the return value of 'List<Api.ListNode>.this[int]' because it is not a variable
				// to fix that we have to discover this case and transform code as below
				tmp := l.nodes[prevIndex]
				tmp.next = l.nodes[currIndex].next
				l.nodes[prevIndex] = tmp // Update the previous node's next pointer
			}
			return l
		}
		prevIndex = currIndex
		currIndex = l.nodes[currIndex].next
	}

	fmt.Println("Value not found in the list.")
	return l
}

// Print prints the list
func PrintList(l List) {
	if l.head == -1 {
		fmt.Println("The list is empty.")
		return
	}

	currIndex := l.head
	for currIndex != -1 {
		fmt.Printf("%d -> ", l.nodes[currIndex].value)
		currIndex = l.nodes[currIndex].next
	}
	fmt.Println("nil")
}

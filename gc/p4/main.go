package main

import "fmt"

// Node represents an object in our allocator that can reference other objects
type Node struct {
	Value    interface{} // Arbitrary data (can be any type)
	RefCount int         // Number of references this node holds
	Refs     []int       // Indices of other nodes this node references (simulating pointers)
}

// Allocator manages arena-backed storage with support for representing pointers as indices
type Allocator struct {
	arena     []Node       // Storage for our nodes
	inUse     []bool       // Tracks whether a slot is currently allocated
	freeList  []int        // List of freed indices ready for reuse
	nextFree  int          // Next index to allocate if freeList is empty
	gcCounter int          // Counter to trigger GC
	rootSet   map[int]bool // Set of root object indices
}

// NewAllocator creates an allocator with a fixed arena size
func NewAllocator(capacity int) *Allocator {
	return &Allocator{
		arena:     make([]Node, capacity),
		inUse:     make([]bool, capacity),
		freeList:  make([]int, 0, capacity),
		nextFree:  0,
		gcCounter: 0,
		rootSet:   make(map[int]bool),
	}
}

// Alloc returns a free index from the arena
func (a *Allocator) Alloc() int {
	var index int
	if len(a.freeList) > 0 {
		// Reuse from freelist
		index = a.freeList[len(a.freeList)-1]
		a.freeList = a.freeList[:len(a.freeList)-1]
	} else if a.nextFree < len(a.arena) {
		// Allocate new slot
		index = a.nextFree
		a.nextFree++
	} else {
		fmt.Println("Allocator out of memory")
		return -1
	}

	// Initialize node and mark as in use
	a.arena[index] = Node{
		Value:    nil,
		RefCount: 0,
		Refs:     make([]int, 0),
	}
	a.inUse[index] = true

	return index
}

// Mark traverses the object graph starting from root indices and marks reachable nodes
func (a *Allocator) Mark(roots []int) {
	// Reset all nodes to not in use first
	for i := range a.inUse {
		a.inUse[i] = false
	}

	// Create a stack for DFS traversal of the object graph
	var stack []int
	for _, root := range roots {
		if root >= 0 && root < len(a.arena) {
			stack = append(stack, root)
		}
	}

	// If no roots provided, we can't mark anything
	if len(stack) == 0 {
		fmt.Println("No root objects to mark")
		return
	}

	// Depth-first search to mark all reachable objects
	visited := make(map[int]bool)
	markCount := 0

	for len(stack) > 0 {
		// Pop from stack
		curr := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		// Skip if already visited or invalid index
		if visited[curr] || curr < 0 || curr >= len(a.arena) {
			continue
		}

		// Mark as visited and in-use
		visited[curr] = true
		a.inUse[curr] = true
		markCount++

		// Add all references to the stack
		for _, ref := range a.arena[curr].Refs {
			if !visited[ref] && ref >= 0 && ref < len(a.arena) {
				stack = append(stack, ref)
			}
		}
	}

	fmt.Printf("Mark phase complete. Marked %d objects as in-use.\n", markCount)
}

// Sweep frees all unmarked nodes
func (a *Allocator) Sweep() {
	sweepCount := 0

	for i := 0; i < a.nextFree; i++ {
		if !a.inUse[i] { // If not marked during mark phase
			// Add to free list for reuse
			a.freeList = append(a.freeList, i)
			// Clear the node
			a.arena[i] = Node{}
			sweepCount++
		}
	}

	fmt.Printf("Sweep phase complete. Freed %d objects.\n", sweepCount)
}

// Free explicitly deallocates a node
func (a *Allocator) Free(index int) {
	if index >= 0 && index < len(a.arena) {
		a.inUse[index] = false
		a.freeList = append(a.freeList, index)
		// Clear the node
		a.arena[index] = Node{}
	}
}

// AddReference adds a reference from one node to another (simulating pointer assignment)
func (a *Allocator) AddReference(fromIdx, toIdx int) bool {
	if fromIdx < 0 || fromIdx >= len(a.arena) || toIdx < 0 || toIdx >= len(a.arena) {
		return false
	}

	// Check if reference already exists
	for _, ref := range a.arena[fromIdx].Refs {
		if ref == toIdx {
			return true // Reference already exists
		}
	}

	// Add new reference
	a.arena[fromIdx].Refs = append(a.arena[fromIdx].Refs, toIdx)
	a.arena[fromIdx].RefCount++
	return true
}

// RemoveReference removes a reference from one node to another (simulating pointer nullification)
func (a *Allocator) RemoveReference(fromIdx, toIdx int) bool {
	if fromIdx < 0 || fromIdx >= len(a.arena) || toIdx < 0 || toIdx >= len(a.arena) {
		return false
	}

	// Find and remove the reference
	for i, ref := range a.arena[fromIdx].Refs {
		if ref == toIdx {
			// Remove by swapping with last element and truncating
			lastIdx := len(a.arena[fromIdx].Refs) - 1
			a.arena[fromIdx].Refs[i] = a.arena[fromIdx].Refs[lastIdx]
			a.arena[fromIdx].Refs = a.arena[fromIdx].Refs[:lastIdx]
			a.arena[fromIdx].RefCount--
			return true
		}
	}

	return false // Reference not found
}

// AddRoot marks an object as a root that should not be garbage collected
func (a *Allocator) AddRoot(index int) bool {
	if index < 0 || index >= len(a.arena) || !a.inUse[index] {
		return false
	}
	a.rootSet[index] = true
	return true
}

// RemoveRoot removes an object from the root set
func (a *Allocator) RemoveRoot(index int) bool {
	if _, exists := a.rootSet[index]; exists {
		delete(a.rootSet, index)
		return true
	}
	return false
}

// GetRoots returns the current root set as a slice
func (a *Allocator) GetRoots() []int {
	roots := make([]int, 0, len(a.rootSet))
	for root := range a.rootSet {
		roots = append(roots, root)
	}
	return roots
}

// New allocates a new node with initial value and performs GC if needed
func (a *Allocator) New(value interface{}, refs []int) int {
	// Run GC every 10 allocations
	a.gcCounter++
	if a.gcCounter%10 == 0 {
		fmt.Println("\nPerforming garbage collection...")
		// Use all current roots for GC
		roots := a.GetRoots()
		if len(roots) == 0 {
			fmt.Println("Warning: No root objects defined, GC won't collect anything")
		}
		a.Mark(roots)
		a.Sweep()
	}

	index := a.Alloc()
	if index == -1 {
		return -1
	}

	// Set the node's value
	a.arena[index].Value = value

	// Add initial references if provided
	for _, ref := range refs {
		a.AddReference(index, ref)
	}

	return index
}

// GetValue returns the value at the given index
func (a *Allocator) GetValue(index int) (interface{}, bool) {
	if index < 0 || index >= len(a.arena) || !a.inUse[index] {
		return nil, false
	}
	return a.arena[index].Value, true
}

// SetValue sets the value at the given index
func (a *Allocator) SetValue(index int, value interface{}) bool {
	if index < 0 || index >= len(a.arena) || !a.inUse[index] {
		return false
	}
	a.arena[index].Value = value
	return true
}

// Main demonstrates use of the allocator with various data types
func main() {
	// Create an allocator with 20 slots
	alloc := NewAllocator(20)

	fmt.Println("Creating an object graph with various data types...")

	// Create a root object (using a string)
	rootIdx := alloc.New("Root Object", nil)
	// Mark it as a root for GC
	alloc.AddRoot(rootIdx)
	fmt.Printf("Created root object at index %d\n", rootIdx)

	// Create a second root object
	root2Idx := alloc.New("Second Root", nil)
	alloc.AddRoot(root2Idx)
	fmt.Printf("Created second root object at index %d\n", root2Idx)

	// Create child objects with different types
	childA := alloc.New(42, nil)                                               // int
	childB := alloc.New(3.14159, nil)                                          // float64
	childC := alloc.New(map[string]string{"name": "node", "type": "map"}, nil) // map
	childD := alloc.New([]string{"Go", "is", "awesome"}, nil)                  // slice

	fmt.Printf("Created child objects with different types at indices %d, %d, %d, %d\n",
		childA, childB, childC, childD)

	// Add references from first root to children A and B
	alloc.AddReference(rootIdx, childA)
	alloc.AddReference(rootIdx, childB)
	fmt.Printf("Added references from first root to children A and B\n")

	// Add references from second root to children C and D
	alloc.AddReference(root2Idx, childC)
	alloc.AddReference(root2Idx, childD)
	fmt.Printf("Added references from second root to children C and D\n")

	// Create a more complex object: a struct
	type Person struct {
		Name string
		Age  int
	}

	personIdx := alloc.New(Person{Name: "Alice", Age: 30}, nil)
	fmt.Printf("Created a struct object at index %d\n", personIdx)

	// Link the Person to the first root
	alloc.AddReference(rootIdx, personIdx)

	// Create some unreachable objects
	unreachable1 := alloc.New("Unreachable String", nil)
	unreachable2 := alloc.New(true, nil) // boolean
	alloc.AddReference(unreachable1, unreachable2)
	fmt.Printf("Created unreachable objects at indices %d and %d\n", unreachable1, unreachable2)

	// Allocate more objects to trigger GC
	fmt.Println("\nAllocating more objects to trigger garbage collection...")
	for i := 0; i < 5; i++ {
		alloc.New(fmt.Sprintf("Auto-created object %d", i), nil)
	}

	// Print the state after GC
	fmt.Println("\nAllocator state after GC:")
	for i := 0; i < alloc.nextFree; i++ {
		if alloc.inUse[i] {
			node := alloc.arena[i]
			fmt.Printf("Index %d: Value=%v, References=%v\n", i, node.Value, node.Refs)
		}
	}

	// Remove the second root, making children C and D unreachable
	fmt.Println("\nRemoving second root from root set")
	alloc.RemoveRoot(root2Idx)

	// Trigger GC again
	fmt.Println("\nManually triggering garbage collection...")
	alloc.Mark(alloc.GetRoots())
	alloc.Sweep()

	// Print final state
	fmt.Println("\nFinal allocator state:")
	for i := 0; i < alloc.nextFree; i++ {
		if alloc.inUse[i] {
			node := alloc.arena[i]
			fmt.Printf("Index %d: Value=%v, References=%v\n", i, node.Value, node.Refs)
		}
	}
	fmt.Printf("Free list: %v\n", alloc.freeList)
	fmt.Printf("Root set: %v\n", alloc.GetRoots())
}

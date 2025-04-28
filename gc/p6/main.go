package main

import (
	"fmt"
	"time"
)

// Node represents an object in our allocator that can reference other objects
// It's now generic with type parameter T
type Node[T any] struct {
	Value    T     // Typed data
	RefCount int   // Number of references this node holds
	Refs     []int // Indices of other nodes this node references (simulating pointers)
}

// Allocator is now generic and can be specialized for different node types
type Allocator[T any] struct {
	arena     []Node[T]    // Storage for our nodes of type T
	inUse     []bool       // Tracks whether a slot is currently allocated
	freeList  []int        // List of freed indices ready for reuse
	nextFree  int          // Next index to allocate if freeList is empty
	gcCounter int          // Counter to trigger GC
	rootSet   map[int]bool // Set of root object indices
}

// NewAllocator creates an allocator with a fixed arena size for type T
func NewAllocator[T any](capacity int) *Allocator[T] {
	return &Allocator[T]{
		arena:     make([]Node[T], capacity),
		inUse:     make([]bool, capacity),
		freeList:  make([]int, 0, capacity),
		nextFree:  0,
		gcCounter: 0,
		rootSet:   make(map[int]bool),
	}
}

// Alloc returns a free index from the arena
func (a *Allocator[T]) Alloc() int {
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
	a.arena[index] = Node[T]{
		RefCount: 0,
		Refs:     make([]int, 0),
	}
	a.inUse[index] = true

	return index
}

// Mark traverses the object graph starting from root indices and marks reachable nodes
func (a *Allocator[T]) Mark(roots []int) {
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
func (a *Allocator[T]) Sweep() {
	sweepCount := 0

	for i := 0; i < a.nextFree; i++ {
		if !a.inUse[i] { // If not marked during mark phase
			// Add to free list for reuse
			a.freeList = append(a.freeList, i)
			// Clear the node
			a.arena[i] = Node[T]{}
			sweepCount++
		}
	}

	fmt.Printf("Sweep phase complete. Freed %d objects.\n", sweepCount)
}

// Free explicitly deallocates a node
func (a *Allocator[T]) Free(index int) {
	if index >= 0 && index < len(a.arena) {
		a.inUse[index] = false
		a.freeList = append(a.freeList, index)
		// Clear the node
		a.arena[index] = Node[T]{}
	}
}

// AddReference adds a reference from one node to another (simulating pointer assignment)
func (a *Allocator[T]) AddReference(fromIdx, toIdx int) bool {
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
func (a *Allocator[T]) RemoveReference(fromIdx, toIdx int) bool {
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
func (a *Allocator[T]) AddRoot(index int) bool {
	if index < 0 || index >= len(a.arena) || !a.inUse[index] {
		return false
	}
	a.rootSet[index] = true
	return true
}

// RemoveRoot removes an object from the root set
func (a *Allocator[T]) RemoveRoot(index int) bool {
	if _, exists := a.rootSet[index]; exists {
		delete(a.rootSet, index)
		return true
	}
	return false
}

// GetRoots returns the current root set as a slice
func (a *Allocator[T]) GetRoots() []int {
	roots := make([]int, 0, len(a.rootSet))
	for root := range a.rootSet {
		roots = append(roots, root)
	}
	return roots
}

// New allocates a new node with initial value and performs GC if needed
func (a *Allocator[T]) New(value T, refs []int) int {
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
func (a *Allocator[T]) GetValue(index int) (T, bool) {
	var zero T
	if index < 0 || index >= len(a.arena) || !a.inUse[index] {
		return zero, false
	}
	return a.arena[index].Value, true
}

// SetValue sets the value at the given index
func (a *Allocator[T]) SetValue(index int, value T) bool {
	if index < 0 || index >= len(a.arena) || !a.inUse[index] {
		return false
	}
	a.arena[index].Value = value
	return true
}

// CreateRegion creates a new allocation region for grouping objects
func (a *Allocator[T]) CreateRegion(name string) int {
	// For this generic version, we need to convert the string to T
	// This is a simplified version - in real code, you'd need proper conversion
	regionIdx := a.New(any(name).(T), nil)
	a.AddRoot(regionIdx)
	return regionIdx
}

// AllocInRegion allocates an object in a specific region
func (a *Allocator[T]) AllocInRegion(regionIdx int, value T) int {
	objIdx := a.New(value, nil)
	a.AddReference(regionIdx, objIdx)
	return objIdx
}

// DestroyRegion removes a region and all its objects
func (a *Allocator[T]) DestroyRegion(regionIdx int) {
	a.RemoveRoot(regionIdx)
	// Will be collected on next GC
}

// RunGC explicitly triggers garbage collection
func (a *Allocator[T]) RunGC() {
	fmt.Println("\nExplicitly triggering garbage collection...")
	roots := a.GetRoots()
	if len(roots) == 0 {
		fmt.Println("Warning: No root objects defined, GC won't collect anything")
	}
	a.Mark(roots)
	a.Sweep()
}

// Main demonstrates four different examples of garbage collection
func main() {
	fmt.Println("FOUR ENHANCED EXAMPLES OF GARBAGE COLLECTION")

	// ============================================================
	// EXAMPLE 1: BASIC GARBAGE COLLECTION
	// ============================================================
	fmt.Println("\n======== EXAMPLE 1: BASIC GARBAGE COLLECTION ========")

	// Create a string allocator
	basicAlloc := NewAllocator[string](20)

	// Create a root object
	root := basicAlloc.New("Root", nil)
	basicAlloc.AddRoot(root)
	fmt.Printf("Created root at index %d\n", root)

	// Create reachable objects
	reachable1 := basicAlloc.New("Reachable 1", nil)
	reachable2 := basicAlloc.New("Reachable 2", nil)

	// Link reachable objects to root
	basicAlloc.AddReference(root, reachable1)
	basicAlloc.AddReference(root, reachable2)
	fmt.Printf("Created reachable objects at indices %d and %d\n",
		reachable1, reachable2)

	// Create unreachable objects
	unreachable1 := basicAlloc.New("Unreachable 1", nil)
	unreachable2 := basicAlloc.New("Unreachable 2", nil)
	fmt.Printf("Created unreachable objects at indices %d and %d\n",
		unreachable1, unreachable2)

	// Show state before GC
	fmt.Println("\nAllocator state BEFORE garbage collection:")
	for i := 0; i < basicAlloc.nextFree; i++ {
		if basicAlloc.inUse[i] {
			val, _ := basicAlloc.GetValue(i)
			fmt.Printf("Index %d: %s\n", i, val)
		}
	}

	// Run garbage collection
	basicAlloc.RunGC()

	// Show state after GC
	fmt.Println("\nAllocator state AFTER garbage collection:")
	for i := 0; i < basicAlloc.nextFree; i++ {
		if basicAlloc.inUse[i] {
			val, _ := basicAlloc.GetValue(i)
			fmt.Printf("Index %d: %s\n", i, val)
		}
	}
	fmt.Printf("Free list: %v\n", basicAlloc.freeList)

	// ============================================================
	// EXAMPLE 2: REFERENCE CYCLES
	// ============================================================
	fmt.Println("\n======== EXAMPLE 2: REFERENCE CYCLES ========")

	// Create allocator for cycle demonstration
	cycleAlloc := NewAllocator[string](20)

	// Create a root
	cycleRoot := cycleAlloc.New("Cycle Root", nil)
	cycleAlloc.AddRoot(cycleRoot)

	// Create a cycle: A -> B -> C -> A
	nodeA := cycleAlloc.New("Node A", nil)
	nodeB := cycleAlloc.New("Node B", nil)
	nodeC := cycleAlloc.New("Node C", nil)

	// Set up the cycle
	cycleAlloc.AddReference(nodeA, nodeB)
	cycleAlloc.AddReference(nodeB, nodeC)
	cycleAlloc.AddReference(nodeC, nodeA)

	// Connect root to the cycle
	cycleAlloc.AddReference(cycleRoot, nodeA)

	fmt.Println("Created a reference cycle: A -> B -> C -> A")
	fmt.Printf("Cycle nodes at indices: A=%d, B=%d, C=%d\n", nodeA, nodeB, nodeC)

	// State before breaking connection
	fmt.Println("\nAllocator state with reachable cycle:")
	for i := 0; i < cycleAlloc.nextFree; i++ {
		if cycleAlloc.inUse[i] {
			val, _ := cycleAlloc.GetValue(i)
			refs := cycleAlloc.arena[i].Refs
			fmt.Printf("Index %d: %s, References: %v\n", i, val, refs)
		}
	}

	// Break connection from root to cycle
	fmt.Println("\nBreaking connection from root to cycle...")
	cycleAlloc.RemoveReference(cycleRoot, nodeA)

	// Run GC
	cycleAlloc.RunGC()

	// State after GC
	fmt.Println("\nAllocator state after breaking connection and GC:")
	for i := 0; i < cycleAlloc.nextFree; i++ {
		if cycleAlloc.inUse[i] {
			val, _ := cycleAlloc.GetValue(i)
			refs := cycleAlloc.arena[i].Refs
			fmt.Printf("Index %d: %s, References: %v\n", i, val, refs)
		}
	}
	fmt.Printf("Free list: %v\n", cycleAlloc.freeList)

	// ============================================================
	// EXAMPLE 3: REGION-BASED MEMORY MANAGEMENT
	// ============================================================
	fmt.Println("\n======== EXAMPLE 3: REGION-BASED MEMORY MANAGEMENT ========")

	// Create person allocator
	type Person struct {
		Name string
		Age  int
	}

	personAlloc := NewAllocator[Person](20)

	// Create a permanent region
	permanent := personAlloc.New(Person{Name: "Permanent Region", Age: 0}, nil)
	personAlloc.AddRoot(permanent)

	// Create a temporary region
	temporary := personAlloc.New(Person{Name: "Temporary Region", Age: 0}, nil)
	personAlloc.AddRoot(temporary)

	fmt.Printf("Created permanent region at index %d\n", permanent)
	fmt.Printf("Created temporary region at index %d\n", temporary)

	// Add objects to permanent region
	for i := 0; i < 3; i++ {
		p := personAlloc.New(Person{
			Name: fmt.Sprintf("Permanent Person %d", i+1),
			Age:  30 + i,
		}, nil)
		personAlloc.AddReference(permanent, p)
		fmt.Printf("Added permanent person at index %d\n", p)
	}

	// Add objects to temporary region
	for i := 0; i < 5; i++ {
		p := personAlloc.New(Person{
			Name: fmt.Sprintf("Temporary Person %d", i+1),
			Age:  20 + i,
		}, nil)
		personAlloc.AddReference(temporary, p)
		fmt.Printf("Added temporary person at index %d\n", p)
	}

	// Show all objects
	fmt.Println("\nAll objects before region deallocation:")
	for i := 0; i < personAlloc.nextFree; i++ {
		if personAlloc.inUse[i] {
			person, _ := personAlloc.GetValue(i)
			fmt.Printf("Index %d: %s (age %d)\n", i, person.Name, person.Age)
		}
	}

	// Remove temporary region
	fmt.Println("\nRemoving temporary region...")
	personAlloc.RemoveRoot(temporary)

	// Run GC
	personAlloc.RunGC()

	// Show remaining objects
	fmt.Println("\nRemaining objects after region deallocation and GC:")
	for i := 0; i < personAlloc.nextFree; i++ {
		if personAlloc.inUse[i] {
			person, _ := personAlloc.GetValue(i)
			fmt.Printf("Index %d: %s (age %d)\n", i, person.Name, person.Age)
		}
	}
	fmt.Printf("Free list: %v\n", personAlloc.freeList)

	// ============================================================
	// EXAMPLE 4: OBJECT POOLING AND REUSE
	// ============================================================
	fmt.Println("\n======== EXAMPLE 4: OBJECT POOLING AND REUSE ========")

	// Create a pool allocator for integers
	poolAlloc := NewAllocator[int](30)

	// Create a root for active objects
	active := poolAlloc.New(0, nil)
	poolAlloc.AddRoot(active)

	// Create a root for the free pool
	pool := poolAlloc.New(0, nil)
	poolAlloc.AddRoot(pool)

	fmt.Printf("Created active list at index %d\n", active)
	fmt.Printf("Created object pool at index %d\n", pool)

	// Fill the pool with some initial objects
	fmt.Println("\nPrefilling the object pool:")
	for i := 0; i < 10; i++ {
		obj := poolAlloc.New(-1, nil) // -1 means available
		poolAlloc.AddReference(pool, obj)
		fmt.Printf("Added object to pool at index %d\n", obj)
	}

	// Use some objects from the pool
	fmt.Println("\nUsing objects from the pool:")

	// Keep track of used objects
	used := make([]int, 0, 5)

	// Use 5 objects
	poolRefs := poolAlloc.arena[pool].Refs
	for i := 0; i < 5 && i < len(poolRefs); i++ {
		objIdx := poolRefs[i]

		// Remove from pool
		poolAlloc.RemoveReference(pool, objIdx)

		// Add to active list with a value
		poolAlloc.SetValue(objIdx, 100+i)
		poolAlloc.AddReference(active, objIdx)

		used = append(used, objIdx)
		fmt.Printf("Used object at index %d, set value to %d\n", objIdx, 100+i)
	}

	// Run GC (should not affect anything)
	poolAlloc.RunGC()

	// Return some objects to the pool
	fmt.Println("\nReturning objects to the pool:")
	for i := 0; i < 3 && i < len(used); i++ {
		objIdx := used[i]

		// Remove from active
		poolAlloc.RemoveReference(active, objIdx)

		// Reset value and return to pool
		poolAlloc.SetValue(objIdx, -1)
		poolAlloc.AddReference(pool, objIdx)

		fmt.Printf("Returned object at index %d to the pool\n", objIdx)
	}

	// Run GC again
	poolAlloc.RunGC()

	// Show final state
	fmt.Println("\nFinal state of object pool system:")
	fmt.Printf("Objects in active list (%d): ", len(poolAlloc.arena[active].Refs))
	for _, idx := range poolAlloc.arena[active].Refs {
		val, _ := poolAlloc.GetValue(idx)
		fmt.Printf("%d(%d) ", idx, val)
	}
	fmt.Println()

	fmt.Printf("Objects in pool (%d): ", len(poolAlloc.arena[pool].Refs))
	for _, idx := range poolAlloc.arena[pool].Refs {
		val, _ := poolAlloc.GetValue(idx)
		fmt.Printf("%d(%d) ", idx, val)
	}
	fmt.Println()

	fmt.Printf("Free list: %v\n", poolAlloc.freeList)

	// ============================================================
	// EXAMPLE 5: AUTOMATIC REGION-BASED CLEANUP WITH WithRegion
	// ============================================================
	fmt.Println("\n======== EXAMPLE 5: AUTOMATIC REGION-BASED CLEANUP WITH WithRegion ========")

	// Create an allocator for processing structured data
	type DataItem struct {
		Name  string
		Value int
	}

	regionAlloc := NewAllocator[DataItem](50)
	fmt.Println("Created a new allocator for demonstrating WithRegion")

	// Track objects allocated outside regions for comparison
	persistentObj := regionAlloc.New(DataItem{Name: "Persistent Object", Value: 1000}, nil)
	regionAlloc.AddRoot(persistentObj)
	fmt.Printf("Created persistent object at index %d\n", persistentObj)

	// First: Create and use a temporary region that will auto-cleanup
	fmt.Println("\n1. Creating first region with automatic cleanup...")
	regionAlloc.WithRegion(func(region int) {
		fmt.Printf("Inside first region (index: %d)\n", region)

		// Create several objects in this region
		for i := 0; i < 5; i++ {
			obj := regionAlloc.New(DataItem{
				Name:  fmt.Sprintf("Temp Object %d", i+1),
				Value: i * 100,
			}, nil)

			// Connect to the region (this is crucial for auto-cleanup to work)
			regionAlloc.AddReference(region, obj)
			fmt.Printf("  Created temporary object at index %d: %s\n",
				obj, regionAlloc.arena[obj].Value.Name)
		}

		// Show all objects while inside region
		fmt.Println("\n  Objects inside first region scope:")
		for i := 0; i < regionAlloc.nextFree; i++ {
			if regionAlloc.inUse[i] {
				val, _ := regionAlloc.GetValue(i)
				fmt.Printf("  Index %d: %s (value: %d)\n", i, val.Name, val.Value)
			}
		}

		fmt.Println("  Exiting first region scope - objects will be automatically freed")
	}) // Region automatically cleaned up here

	// Check what objects remain after region cleanup
	fmt.Println("\nAfter first region cleanup - only persistent objects should remain:")
	for i := 0; i < regionAlloc.nextFree; i++ {
		if regionAlloc.inUse[i] {
			val, _ := regionAlloc.GetValue(i)
			fmt.Printf("Index %d: %s (value: %d)\n", i, val.Name, val.Value)
		}
	}
	fmt.Printf("Free list: %v\n", regionAlloc.freeList)

	// Second: Demonstrate nested regions
	fmt.Println("\n2. Demonstrating nested regions...")
	regionAlloc.WithRegion(func(outerRegion int) {
		fmt.Printf("Inside outer region (index: %d)\n", outerRegion)

		// Create some objects in outer region
		outerObj := regionAlloc.New(DataItem{Name: "Outer Region Object", Value: 500}, nil)
		regionAlloc.AddReference(outerRegion, outerObj)
		fmt.Printf("  Created outer region object at index %d\n", outerObj)

		// Create a nested inner region
		fmt.Println("\n  Creating nested inner region...")
		regionAlloc.WithRegion(func(innerRegion int) {
			fmt.Printf("  Inside inner region (index: %d)\n", innerRegion)

			// Create objects in inner region
			for i := 0; i < 3; i++ {
				innerObj := regionAlloc.New(DataItem{
					Name:  fmt.Sprintf("Inner Object %d", i+1),
					Value: 200 + i,
				}, nil)
				regionAlloc.AddReference(innerRegion, innerObj)
				fmt.Printf("    Created inner region object at index %d\n", innerObj)
			}

			// Show all objects at inner scope
			fmt.Println("\n    Objects inside inner region scope:")
			for i := 0; i < regionAlloc.nextFree; i++ {
				if regionAlloc.inUse[i] {
					val, _ := regionAlloc.GetValue(i)
					fmt.Printf("    Index %d: %s (value: %d)\n", i, val.Name, val.Value)
				}
			}

			fmt.Println("    Exiting inner region scope - inner objects will be freed")
		}) // Inner region automatically cleaned up here

		// Show objects after inner region is cleaned up
		fmt.Println("\n  After inner region cleanup - outer objects should remain:")
		for i := 0; i < regionAlloc.nextFree; i++ {
			if regionAlloc.inUse[i] {
				val, _ := regionAlloc.GetValue(i)
				fmt.Printf("  Index %d: %s (value: %d)\n", i, val.Name, val.Value)
			}
		}

		fmt.Println("  Exiting outer region scope - outer objects will be freed")
	}) // Outer region automatically cleaned up here

	// Final state
	fmt.Println("\nFinal state after all regions - only persistent object should remain:")
	for i := 0; i < regionAlloc.nextFree; i++ {
		if regionAlloc.inUse[i] {
			val, _ := regionAlloc.GetValue(i)
			fmt.Printf("Index %d: %s (value: %d)\n", i, val.Name, val.Value)
		}
	}
	fmt.Printf("Free list: %v\n", regionAlloc.freeList)

	// ============================================================
	// EXAMPLE 6: OPTIMIZED GARBAGE COLLECTION FOR ACYCLIC GRAPHS
	// ============================================================
	fmt.Println("\n======== EXAMPLE 6: OPTIMIZED GARBAGE COLLECTION FOR ACYCLIC GRAPHS ========")

	// Create a tree-based allocator (acyclic structure)
	type TreeNode struct {
		Label string
		Depth int
	}

	treeAlloc := NewAllocator[TreeNode](50)
	fmt.Println("Created a tree allocator for demonstrating MarkNoCycles")

	// Create tree structure (acyclic)
	// Root
	root = treeAlloc.New(TreeNode{Label: "Root", Depth: 0}, nil)
	treeAlloc.AddRoot(root)
	fmt.Printf("Created tree root at index %d\n", root)

	// Level 1 nodes
	level1Nodes := make([]int, 3)
	for i := 0; i < 3; i++ {
		level1Nodes[i] = treeAlloc.New(TreeNode{
			Label: fmt.Sprintf("Level 1 Node %d", i+1),
			Depth: 1,
		}, nil)
		treeAlloc.AddReference(root, level1Nodes[i])
		fmt.Printf("Created level 1 node at index %d\n", level1Nodes[i])
	}

	// Level 2 nodes (children of level 1)
	for i, parentIdx := range level1Nodes {
		childCount := 2 + i // Different number of children for each parent

		for j := 0; j < childCount; j++ {
			childIdx := treeAlloc.New(TreeNode{
				Label: fmt.Sprintf("L2 Child %d of Parent %d", j+1, i+1),
				Depth: 2,
			}, nil)
			treeAlloc.AddReference(parentIdx, childIdx)
			fmt.Printf("Created level 2 node at index %d (parent: %d)\n", childIdx, parentIdx)

			// Some level 2 nodes get level 3 children
			if j%2 == 0 {
				leafIdx := treeAlloc.New(TreeNode{
					Label: fmt.Sprintf("Leaf child of L2 Node %d", childIdx),
					Depth: 3,
				}, nil)
				treeAlloc.AddReference(childIdx, leafIdx)
				fmt.Printf("Created leaf node at index %d (parent: %d)\n", leafIdx, childIdx)
			}
		}
	}

	// Print the tree structure using the print helper
	fmt.Println("\nTree structure before garbage collection:")
	printFileSystem(treeAlloc, root, 0)

	// Run standard GC and measure
	fmt.Println("\nRunning standard garbage collection for comparison:")
	var standardStart time.Time
	var standardDuration time.Duration
	standardStart = time.Now()
	treeAlloc.RunGC()
	standardDuration = time.Since(standardStart)
	fmt.Printf("Standard GC took %v\n", standardDuration)

	// Remove some nodes to create unreachable subtrees
	fmt.Println("\nRemoving references to create unreachable subtrees...")
	treeAlloc.RemoveReference(root, level1Nodes[1])
	fmt.Printf("Removed reference from root to level 1 node %d\n", level1Nodes[1])

	// Run optimized GC for acyclic graphs and measure
	fmt.Println("\nRunning optimized garbage collection for acyclic graphs:")
	var optimizedStart time.Time
	var optimizedDuration time.Duration
	optimizedStart = time.Now()
	treeAlloc.RunGCNoCycles()
	optimizedDuration = time.Since(optimizedStart)
	fmt.Printf("Optimized GC took %v\n", optimizedDuration)

	// Print the tree structure after GC
	fmt.Println("\nTree structure after garbage collection:")
	printFileSystem(treeAlloc, root, 0)
	fmt.Printf("Free list: %v\n", treeAlloc.freeList)
}

// Helper function to print a file system hierarchy
func printFileSystem[T any](alloc *Allocator[T], nodeIdx int, depth int) {
	// Skip if invalid index
	if nodeIdx < 0 || nodeIdx >= len(alloc.arena) || !alloc.inUse[nodeIdx] {
		return
	}

	// Get the node
	node, _ := alloc.GetValue(nodeIdx)

	// Print with indentation based on depth
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}

	// Print node
	fmt.Printf("%s|-- %v\n", indent, node)

	// Print children
	for _, childIdx := range alloc.arena[nodeIdx].Refs {
		printFileSystem(alloc, childIdx, depth+1)
	}
}

// Create a region and automatically manage its lifecycle
func (a *Allocator[T]) WithRegion(f func(regionIdx int)) {
	// Create a new "region" object with appropriate zero value of type T
	var zeroValue T
	regionIdx := a.New(zeroValue, nil)
	f(regionIdx) // Use the region

	// When function ends, free the entire region and its contents
	for _, childIdx := range a.arena[regionIdx].Refs {
		a.FreeTree(childIdx) // Using the tree-based freeing from above
	}
	a.Free(regionIdx)
}

func (a *Allocator[T]) FreeTree(rootIdx int) {
	if !a.inUse[rootIdx] {
		return
	}

	// Free all children first (depth-first)
	for _, childIdx := range a.arena[rootIdx].Refs {
		a.FreeTree(childIdx)
	}

	// Then free the root node itself
	a.Free(rootIdx)
}

// MarkNoCycles provides optimized marking for acyclic object graphs
// This is more efficient than the regular Mark method when you know there are no cycles
func (a *Allocator[T]) MarkNoCycles(roots []int) {
	// Reset all nodes to not in use first
	for i := range a.inUse {
		a.inUse[i] = false
	}

	// For each root, mark it and its entire subtree
	markCount := 0
	for _, root := range roots {
		count := a.markTree(root)
		markCount += count
	}

	fmt.Printf("MarkNoCycles complete. Marked %d objects as in-use.\n", markCount)
}

// markTree recursively marks a node and all its children
// Returns the number of nodes marked
func (a *Allocator[T]) markTree(idx int) int {
	if idx < 0 || idx >= len(a.arena) || a.inUse[idx] {
		return 0
	}

	// Mark this node
	a.inUse[idx] = true
	count := 1

	// Mark all children (no need to track visited nodes since we know there are no cycles)
	for _, childIdx := range a.arena[idx].Refs {
		count += a.markTree(childIdx)
	}

	return count
}

// RunGCNoCycles runs an optimized garbage collection for acyclic object graphs
func (a *Allocator[T]) RunGCNoCycles() {
	fmt.Println("\nRunning optimized garbage collection for acyclic graphs...")
	roots := a.GetRoots()
	if len(roots) == 0 {
		fmt.Println("Warning: No root objects defined, GC won't collect anything")
	}
	a.MarkNoCycles(roots)
	a.Sweep()
}

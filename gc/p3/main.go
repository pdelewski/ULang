package main

import "fmt"

// Allocator manages arena-backed int storage via index
type Allocator struct {
    arena    []int
    freeList []int
    nextFree int
}

// NewAllocator creates an allocator with a fixed arena size
func NewAllocator(capacity int) *Allocator {
    return &Allocator{
	arena:    make([]int, capacity),
	freeList: make([]int, 0, capacity),
	nextFree: 0,
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
    return index
}

func (a *Allocator) AllocAndArena() (int, []int) {
    index := a.Alloc()
    if index == -1 {
	return -1, nil
    }
    return index, a.Arena()
}

func new_(a *Allocator) (int, []int) {
    return a.AllocAndArena()
}

// Free returns an index to the free list for reuse
func (a *Allocator) Free(index int) {
    if index >= 0 && index < len(a.arena) {
	a.freeList = append(a.freeList, index)
    }
}

// Arena gives access to the underlying data
func (a *Allocator) Arena() []int {
    return a.arena
}

// Example function using the allocator
func foo(alloc *Allocator) {
    x_slot, x := new_(alloc)
    if x_slot == -1 {
	return
    }
    x[x_slot] = 42
    fmt.Printf("Allocated value at index %d: %d\n", x_slot, x[x_slot])

    // Simulate we're done with this value
    alloc.Free(x_slot)
    fmt.Printf("Freed index %d\n", x_slot)
}

func bar(alloc *Allocator) {
    x := new(int)
    *x = 42
    fmt.Println("Value inside foo:", *x)
}

// Main test
func main() {
    alloc := NewAllocator(5)

    for i := 0; i < 7; i++ {
	foo(alloc)
    }
}

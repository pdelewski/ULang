package main

import "fmt"

// --- Allocator ---

type Allocator struct {
    freeList []int
    nextFree int
    maxSize  int
}

func NewAllocator(size int) *Allocator {
    free := make([]int, size)
    for i := 0; i < size; i++ {
    free[i] = size - 1 - i
    }
    return &Allocator{
    freeList: free,
    nextFree: size,
    maxSize:  size,
    }
}

func (a *Allocator) Alloc() int {
    if a.nextFree == 0 {
    return -1
    }
    a.nextFree--
    return a.freeList[a.nextFree]
}

func (a *Allocator) Free(index int) {
    if a.nextFree < a.maxSize {
    a.freeList[a.nextFree] = index
    a.nextFree++
    }
}

func foo(allocator *Allocator) {
    x := new(int)
    *x = 42
    fmt.Println("Value inside foo:", *x)
}

var allocator = NewAllocator(100)

func main() {
    foo(allocator)
}

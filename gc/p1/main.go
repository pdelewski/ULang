package main

import (
    "fmt"
)

const MaxNodes = 10

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

// --- Node and List ---

type Node struct {
    value int
    next  int
}

type List struct {
    nodes     []Node
    head      int
    allocator *Allocator
}

func NewList(nodes []Node, allocator *Allocator) *List {
    return &List{
	nodes:     nodes,
	head:      -1,
	allocator: allocator,
    }
}

func (l *List) Insert(value int) {
    idx := l.allocator.Alloc()
    if idx == -1 {
	fmt.Println("List is full")
	return
    }
    l.nodes[idx] = Node{value: value, next: -1}

    if l.head == -1 {
	l.head = idx
	return
    }

    current := l.head
    for l.nodes[current].next != -1 {
	current = l.nodes[current].next
    }
    l.nodes[current].next = idx
}

func (l *List) Delete(value int) {
    if l.head == -1 {
	return
    }

    if l.nodes[l.head].value == value {
	old := l.head
	l.head = l.nodes[l.head].next
	l.allocator.Free(old)
	return
    }

    prev := l.head
    curr := l.nodes[prev].next

    for curr != -1 {
	if l.nodes[curr].value == value {
	    l.nodes[prev].next = l.nodes[curr].next
	    l.allocator.Free(curr)
	    return
	}
	prev = curr
	curr = l.nodes[curr].next
    }
}

func (l *List) Print() {
    for i := l.head; i != -1; i = l.nodes[i].next {
	fmt.Printf("%d -> ", l.nodes[i].value)
    }
    fmt.Println("nil")
}

// --- Main ---

func main() {
    nodes := make([]Node, MaxNodes)
    alloc := NewAllocator(MaxNodes)
    list := NewList(nodes, alloc)

    list.Insert(1)
    list.Insert(2)
    list.Insert(3)
    list.Print() // 1 -> 2 -> 3 -> nil

    list.Delete(2)
    list.Print() // 1 -> 3 -> nil

    list.Insert(4)
    list.Print() // 1 -> 3 -> 4 -> nil
}

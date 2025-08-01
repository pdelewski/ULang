package main

import (
	"errors"
	"fmt"
	"strings"
)

// TopologicalSort performs a topological sort on the given graph.
// The input graph is a map where keys are nodes and values are slices of their dependencies.
func TopologicalSort(graph map[string][]string) ([]string, error) {
	// Track the state of each node: 0 = unvisited, 1 = visiting, 2 = visited
	visited := make(map[string]int)
	result := []string{}

	// Helper function for depth-first search (DFS)
	var visit func(string) error
	visit = func(node string) error {
		state := visited[node]

		// If the node is already visited, return
		if state == 2 {
			return nil
		}
		// If we find a node in "visiting" state, there is a cycle
		if state == 1 {
			return errors.New("cycle detected in the graph")
		}

		// Mark the node as visiting
		visited[node] = 1

		// Visit all the dependencies of the current node, if any
		if deps, exists := graph[node]; exists {
			for _, dep := range deps {
				if err := visit(dep); err != nil {
					return err // propagate the cycle detection error
				}
			}
		}

		// Mark the node as visited and add it to the result
		visited[node] = 2
		result = append(result, node)

		return nil
	}

	// Visit all nodes in the graph (including those without dependencies)
	for node := range graph {
		if visited[node] == 0 {
			if err := visit(node); err != nil {
				return nil, err
			}
		}
	}

	// Ensure we include nodes without outgoing edges
	// For example, in a graph {A -> B}, if C has no dependencies, it should also be in the result.
	for node := range visited {
		if visited[node] == 0 {
			if err := visit(node); err != nil {
				return nil, err
			}
		}
	}

	// Reverse the result because nodes are added in post-order
	reverse(result)

	return result, nil
}

// reverse reverses a slice of strings in place
func reverse(arr []string) {
	for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
		arr[i], arr[j] = arr[j], arr[i]
	}
}

func SliceToMap(slice []string) map[string]int {
	// Create a map to store the string and its index
	result := make(map[string]int)

	// Loop over the slice and fill the map
	for index, value := range slice {
		result[value] = index
	}

	return result
}

func mergeStackElements(marker string, stack []string) []string {
	var merged strings.Builder

	// Process the stack in reverse until we find a marker
	for len(stack) > 0 {
		top := stack[len(stack)-1]
		stack = stack[:len(stack)-1] // Pop element

		// Stop merging when we find a marker
		if strings.HasPrefix(top, marker) {
			stack = append(stack, merged.String()) // Push merged string
			return stack
		}

		// Prepend the element to the merged string (reverse order)
		mergedString := top + merged.String() // Prepend instead of append
		merged.Reset()
		merged.WriteString(mergedString)
	}
	return stack
}

func SearchPointerReverse(target string, pointerAndPositionVec []PointerAndPosition) *PointerAndPosition {
	for i := len(pointerAndPositionVec) - 1; i >= 0; i-- {
		if pointerAndPositionVec[i].Pointer == target {
			return &pointerAndPositionVec[i]
		}
	}
	return nil // Return nil if the pointer is not found
}

func ExtractSubstring(position int, fileBuffer string) (string, error) {
	if position < 0 || position >= len(fileBuffer) {
		return "", fmt.Errorf("position %d is out of bounds", position)
	}
	return fileBuffer[position:], nil
}

func ExtractSubstringBetween(begin int, end int, fileBuffer string) (string, error) {
	if begin < 0 || end > len(fileBuffer) || begin > end {
		return "", fmt.Errorf("invalid range: begin %d, end %d", begin, end)
	}
	return fileBuffer[begin:end], nil
}

func RewriteFileBufferBetween(fileBuffer string, begin int, end int, content string) (string, error) {
	if begin < 0 || end > len(fileBuffer) || begin > end {
		return fileBuffer, fmt.Errorf("invalid range: begin %d, end %d", begin, end)
	}
	return fileBuffer[:begin] + content + fileBuffer[end:], nil
}

func RewriteFileBuffer(fileBuffer string, position int, oldContent, newContent string) (string, error) {
	if position < 0 || position+len(oldContent) > len(fileBuffer) {
		return fileBuffer, fmt.Errorf("position %d is out of bounds or oldContent does not match", position)
	}
	if fileBuffer[position:position+len(oldContent)] != oldContent {
		return fileBuffer, fmt.Errorf("oldContent does not match the existing content at position %d", position)
	}
	return fileBuffer[:position] + newContent + fileBuffer[position+len(oldContent):], nil
}

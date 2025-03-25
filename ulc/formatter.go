package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// formatCode applies basic formatting rules
func formatCode(lines []string) []string {
	var formatted []string
	indentLevel := 0
	indentation := "    " // 4 spaces

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Decrease indent if line closes a block
		if strings.HasPrefix(trimmed, "}") || strings.HasPrefix(trimmed, "];") {
			indentLevel--
		}

		// Format the line with correct indentation
		if indentLevel < 0 {
			indentLevel = 0
		}
		formatted = append(formatted, strings.Repeat(indentation, indentLevel)+trimmed)

		// Increase indent if line opens a block
		if strings.HasSuffix(trimmed, "{") || strings.HasSuffix(trimmed, "[") {
			indentLevel++
		}
	}

	return formatted
}

// readLines reads a file into a slice of strings
func readLines(filename string) ([]string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Join the current directory with the filename to get the full path
	filepath := filepath.Join(wd, filename)

	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

// writeLines writes a slice of strings to a file
func writeLines(filename string, lines []string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(writer, line)
	}
	return writer.Flush()
}

func format(in string, out string) {
	// Read file content
	lines, err := readLines("./" + in)
	if err != nil {
		fmt.Println("Error reading file:", err)
		os.Exit(1)
	}

	// Format code
	formatted := formatCode(lines)

	// Write back to file
	err = writeLines("./"+out, formatted)
	if err != nil {
		fmt.Println("Error writing file:", err)
		os.Exit(1)
	}

	fmt.Println("File formatted successfully!")

}

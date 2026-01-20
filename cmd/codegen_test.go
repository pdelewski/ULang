package main

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestAnnotation represents a parsed @test annotation
type TestAnnotation struct {
	Line        int
	Description string            // The comment line before @test (context)
	Patterns    map[string][]string // backend -> expected patterns
}

// parseTestAnnotations parses @test annotations from a Go source file
// Format: // @test cpp="pattern1" cpp="pattern2" cs="pattern" rust="pattern"
func parseTestAnnotations(filePath string) ([]TestAnnotation, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var annotations []TestAnnotation
	var prevComment string
	lineNum := 0

	// Regex to match @test annotations
	// Matches: backend="pattern" with support for multiple patterns per backend
	patternRegex := regexp.MustCompile(`(cpp|cs|rust)="([^"]*)"`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "//") {
			if strings.Contains(line, "@test") {
				// Parse the annotation
				matches := patternRegex.FindAllStringSubmatch(line, -1)
				if len(matches) > 0 {
					annotation := TestAnnotation{
						Line:        lineNum,
						Description: prevComment,
						Patterns:    make(map[string][]string),
					}
					for _, match := range matches {
						backend := match[1]
						pattern := match[2]
						annotation.Patterns[backend] = append(annotation.Patterns[backend], pattern)
					}
					annotations = append(annotations, annotation)
				}
			} else {
				// Save as potential description for next @test
				prevComment = strings.TrimPrefix(line, "//")
				prevComment = strings.TrimSpace(prevComment)
			}
		}
	}

	return annotations, scanner.Err()
}

// TestCodeGenAnnotations tests generated code against @test annotations in source files
func TestCodeGenAnnotations(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Test cases: source directories containing @test annotations
	testCases := []struct {
		name      string
		sourceDir string
	}{
		{"lang-constructs", filepath.Join(wd, "..", "tests", "lang-constructs")},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			runCodeGenAnnotationTest(t, wd, tc.name, tc.sourceDir)
		})
	}
}

func runCodeGenAnnotationTest(t *testing.T, wd, name, sourceDir string) {
	// Find main.go in the source directory
	mainGoPath := filepath.Join(sourceDir, "main.go")
	if _, err := os.Stat(mainGoPath); os.IsNotExist(err) {
		t.Fatalf("main.go not found in %s", sourceDir)
	}

	// Parse annotations from main.go
	annotations, err := parseTestAnnotations(mainGoPath)
	if err != nil {
		t.Fatalf("Failed to parse annotations: %v", err)
	}

	if len(annotations) == 0 {
		t.Log("No @test annotations found, skipping pattern verification")
		return
	}

	t.Logf("Found %d @test annotations", len(annotations))

	// Create output directory
	outputDir := filepath.Join(os.TempDir(), "codegen_annotation_test_"+name)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}
	defer os.RemoveAll(outputDir)

	// Run the compiler
	outputPath := filepath.Join(outputDir, name)
	cmd := exec.Command("go", "run", ".", "--source="+sourceDir, "--output="+outputPath)
	cmd.Dir = wd
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Code generation failed: %v\nOutput: %s", err, output)
	}

	// Read generated files
	cppContent, err := os.ReadFile(outputPath + ".cpp")
	if err != nil {
		t.Fatalf("Failed to read C++ output: %v", err)
	}

	csContent, err := os.ReadFile(outputPath + ".cs")
	if err != nil {
		t.Fatalf("Failed to read C# output: %v", err)
	}

	rustContent, err := os.ReadFile(outputPath + ".rs")
	if err != nil {
		t.Fatalf("Failed to read Rust output: %v", err)
	}

	// Verify each annotation
	for _, ann := range annotations {
		desc := ann.Description
		if desc == "" {
			desc = "line " + string(rune(ann.Line))
		}

		// Check C++ patterns
		for _, pattern := range ann.Patterns["cpp"] {
			if !strings.Contains(string(cppContent), pattern) {
				t.Errorf("[%s] C++ missing pattern: %q", desc, pattern)
			}
		}

		// Check C# patterns
		for _, pattern := range ann.Patterns["cs"] {
			if !strings.Contains(string(csContent), pattern) {
				t.Errorf("[%s] C# missing pattern: %q", desc, pattern)
			}
		}

		// Check Rust patterns
		for _, pattern := range ann.Patterns["rust"] {
			if !strings.Contains(string(rustContent), pattern) {
				t.Errorf("[%s] Rust missing pattern: %q", desc, pattern)
			}
		}
	}
}

// TestCodeGenAnnotationsVerbose prints all found annotations and generated patterns
func TestCodeGenAnnotationsVerbose(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping verbose test in short mode")
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	sourceDir := filepath.Join(wd, "..", "tests", "lang-constructs")
	mainGoPath := filepath.Join(sourceDir, "main.go")

	annotations, err := parseTestAnnotations(mainGoPath)
	if err != nil {
		t.Fatalf("Failed to parse annotations: %v", err)
	}

	t.Logf("=== Found %d @test annotations ===", len(annotations))
	for i, ann := range annotations {
		t.Logf("%d. [Line %d] %s", i+1, ann.Line, ann.Description)
		for backend, patterns := range ann.Patterns {
			t.Logf("   %s: %v", backend, patterns)
		}
	}
}

package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

// TestCase represents a test case from the e2e workflow
type TestCase struct {
	Name          string
	SourceDir     string
	DotnetEnabled bool
	RustEnabled   bool
}

// E2E test cases matching the workflow
var e2eTestCases = []TestCase{
	{"basic", "tests/basic", true, true},
	{"slice", "tests/slice", true, true},
	{"complex", "tests/complex", true, true},
	{"iceberg", "libs/iceberg", true, false},
	{"contlib", "libs/contlib", true, false},
	{"uql", "libs/uql", true, false},
	{"substrait", "libs/substrait", true, false},
}

func TestE2ECodeGeneration(t *testing.T) {
	// Skip if running short tests
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	for _, tc := range e2eTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			testCodeGeneration(t, tc)
		})
	}
}

func testCodeGeneration(t *testing.T, tc TestCase) {
	// Setup temporary directory for outputs
	tempDir, err := ioutil.TempDir("", "ulc_test_"+tc.Name)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	// Get the source directory path relative to ulc
	sourceDir := filepath.Join(oldDir, "..", tc.SourceDir)
	outputName := tc.Name

	// Test compilation
	testCompilation(t, sourceDir, outputName)

	// Test C++ compilation
	testCppCompilation(t, outputName)

	// Test .NET compilation if enabled
	if tc.DotnetEnabled {
		testDotnetCompilation(t, outputName)
	}

	// Test Rust compilation if enabled
	if tc.RustEnabled {
		testRustCompilation(t, outputName)
	}
}

func testCompilation(t *testing.T, sourceDir, outputName string) {
	// Load packages
	cfg := &packages.Config{
		Mode:  packages.LoadSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedDeps,
		Dir:   sourceDir,
		Tests: false,
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Error loading packages: %v", err)
	}

	if len(pkgs) == 0 {
		t.Fatal("No packages found")
	}

	// Check for package errors
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			t.Logf("Package %s has errors:", pkg.PkgPath)
			for _, err := range pkg.Errors {
				t.Logf("  %v", err)
			}
		}
	}

	// Create emitters
	sema := &BasePass{PassName: "Sema", emitter: &SemaChecker{Emitter: &BaseEmitter{}}}
	cppBackend := &BasePass{PassName: "CppGen", emitter: &CPPEmitter{Emitter: &BaseEmitter{}, Output: outputName + ".cpp"}}
	csBackend := &BasePass{PassName: "CsGen", emitter: &CSharpEmitter{BaseEmitter: BaseEmitter{}, Output: outputName + ".cs"}}
	rustBackend := &BasePass{PassName: "RustGen", emitter: &RustEmitter{BaseEmitter: BaseEmitter{}, Output: outputName + ".rs"}}

	// Run passes
	passManager := &PassManager{
		pkgs: pkgs,
		passes: []Pass{
			sema,
			cppBackend,
			csBackend,
			rustBackend,
		},
	}

	passManager.RunPasses()

	// Verify output files exist and show content on failure
	expectedFiles := []string{outputName + ".cpp", outputName + ".cs", outputName + ".rs"}
	for _, file := range expectedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Expected output file %s does not exist", file)
		} else {
			// Log the generated content for debugging
			if content, err := ioutil.ReadFile(file); err == nil {
				t.Logf("Generated file %s content:\n%s", file, string(content))
			}
		}
	}
}

func testCppCompilation(t *testing.T, outputName string) {
	cppFile := outputName + ".cpp"

	// Check if file exists
	if _, err := os.Stat(cppFile); os.IsNotExist(err) {
		t.Skipf("C++ file %s does not exist, skipping compilation test", cppFile)
		return
	}

	// Try to compile with g++
	cmd := exec.Command("g++", "-std=c++17", cppFile, "-o", outputName+"_cpp")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("C++ compilation failed for %s: %v\nOutput: %s", cppFile, err, string(output))
	}
}

func testDotnetCompilation(t *testing.T, outputName string) {
	csFile := outputName + ".cs"

	// Check if file exists
	if _, err := os.Stat(csFile); os.IsNotExist(err) {
		t.Skipf("C# file %s does not exist, skipping compilation test", csFile)
		return
	}

	// Create .NET project
	projectDir := "dotnet_project"
	appDir := filepath.Join(projectDir, "app")

	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatalf("Failed to create .NET project directory: %v", err)
	}
	defer os.RemoveAll(projectDir)

	// Create new console app
	cmd := exec.Command("dotnet", "new", "console", "--output", appDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Skipf(".NET SDK not available or dotnet new failed: %v\nOutput: %s", err, string(output))
		return
	}

	// Copy generated C# file
	programCs := filepath.Join(appDir, "Program.cs")
	if err := copyFile(csFile, programCs); err != nil {
		t.Fatalf("Failed to copy C# file: %v", err)
	}

	// Build the project
	cmd = exec.Command("dotnet", "build")
	cmd.Dir = appDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf(".NET compilation failed for %s: %v\nOutput: %s", csFile, err, string(output))
	}
}

func testRustCompilation(t *testing.T, outputName string) {
	rsFile := outputName + ".rs"

	// Check if file exists
	if _, err := os.Stat(rsFile); os.IsNotExist(err) {
		t.Skipf("Rust file %s does not exist, skipping compilation test", rsFile)
		return
	}

	// Create Rust project
	projectDir := "rust_project"
	srcDir := filepath.Join(projectDir, "src")

	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("Failed to create Rust project directory: %v", err)
	}
	defer os.RemoveAll(projectDir)

	// Copy Rust file
	mainRs := filepath.Join(srcDir, "main.rs")
	if err := copyFile(rsFile, mainRs); err != nil {
		t.Fatalf("Failed to copy Rust file: %v", err)
	}

	// Create Cargo.toml
	cargoToml := filepath.Join(projectDir, "Cargo.toml")
	cargoContent := `[package]
name = "` + outputName + `"
version = "0.1.0"
edition = "2021"
`
	if err := ioutil.WriteFile(cargoToml, []byte(cargoContent), 0644); err != nil {
		t.Fatalf("Failed to create Cargo.toml: %v", err)
	}

	// Build the project
	cmd := exec.Command("cargo", "build")
	cmd.Dir = projectDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Rust compilation failed for %s: %v\nOutput: %s", rsFile, err, string(output))
	}
}

func copyFile(src, dst string) error {
	data, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dst, data, 0644)
}

// TestCSharpEmitterTokens tests the token functionality specifically
func TestCSharpEmitterTokens(t *testing.T) {
	emitter := &CSharpEmitter{BaseEmitter: BaseEmitter{}}

	// Test token type detection
	testCases := []struct {
		content  string
		expected TokenType
	}{
		{"(", LeftParen},
		{")", RightParen},
		{"[", LeftBracket},
		{"]", RightBracket},
		{"{", LeftBrace},
		{"}", RightBrace},
		{";", Semicolon},
		{",", Comma},
		{"=", Assignment},
		{"+", ArithmeticOperator},
		{"==", ComparisonOperator},
		{"&&", LogicalOperator},
		{"123", NumberLiteral},
		{"\"hello\"", StringLiteral},
		{"variable", Identifier},
		{"class", CSharpKeyword},
		{"if", IfKeyword},
	}

	for _, tc := range testCases {
		t.Run(tc.content, func(t *testing.T) {
			result := emitter.getTokenType(tc.content)
			if result != tc.expected {
				t.Errorf("getTokenType(%q) = %v, expected %v", tc.content, result, tc.expected)
			}
		})
	}
}

// TestCompilerFlags tests various compiler flag combinations
func TestCompilerFlags(t *testing.T) {
	// Test with empty source directory
	t.Run("EmptySourceDir", func(t *testing.T) {
		cfg := &packages.Config{
			Mode: packages.LoadSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedDeps,
			Dir:  "/nonexistent",
		}

		_, err := packages.Load(cfg, "./...")
		if err == nil {
			t.Error("Expected error for nonexistent directory, got nil")
		}
	})
}

// TestOutputFileGeneration tests that all expected output files are generated
func TestOutputFileGeneration(t *testing.T) {
	// Use the basic test case as it's simple
	sourceDir := "../tests/basic"
	outputName := "test_output"

	// Create temporary directory
	tempDir, err := ioutil.TempDir("", "ulc_output_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	// Load packages
	cfg := &packages.Config{
		Mode:  packages.LoadSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedDeps,
		Dir:   filepath.Join(oldDir, sourceDir),
		Tests: false,
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		t.Skipf("Failed to load packages (may be missing): %v", err)
		return
	}

	if len(pkgs) == 0 {
		t.Skip("No packages found")
		return
	}

	// Run compilation
	sema := &BasePass{PassName: "Sema", emitter: &SemaChecker{Emitter: &BaseEmitter{}}}
	cppBackend := &BasePass{PassName: "CppGen", emitter: &CPPEmitter{Emitter: &BaseEmitter{}, Output: outputName + ".cpp"}}
	csBackend := &BasePass{PassName: "CsGen", emitter: &CSharpEmitter{BaseEmitter: BaseEmitter{}, Output: outputName + ".cs"}}
	rustBackend := &BasePass{PassName: "RustGen", emitter: &RustEmitter{BaseEmitter: BaseEmitter{}, Output: outputName + ".rs"}}

	passManager := &PassManager{
		pkgs: pkgs,
		passes: []Pass{
			sema,
			cppBackend,
			csBackend,
			rustBackend,
		},
	}

	passManager.RunPasses()

	// Check that files were created and contain expected content
	expectedFiles := map[string][]string{
		outputName + ".cpp": {"#include", "main", "Hello"},
		outputName + ".cs":  {"using System", "namespace", "Console.WriteLine"},
		outputName + ".rs":  {"fn main", "println!"},
	}

	for filename, expectedContent := range expectedFiles {
		t.Run(filename, func(t *testing.T) {
			if _, err := os.Stat(filename); os.IsNotExist(err) {
				t.Errorf("Expected file %s was not created", filename)
				return
			}

			content, err := ioutil.ReadFile(filename)
			if err != nil {
				t.Errorf("Failed to read file %s: %v", filename, err)
				return
			}

			contentStr := string(content)
			t.Logf("Generated file %s content:\n%s", filename, contentStr)

			hasFailures := false
			for _, expected := range expectedContent {
				if !strings.Contains(contentStr, expected) {
					t.Errorf("File %s does not contain expected content: %s", filename, expected)
					hasFailures = true
				}
			}

			// If there were failures, dump the entire generated content for debugging
			if hasFailures {
				t.Logf("=== FULL GENERATED CONTENT FOR DEBUGGING ===")
				t.Logf("File: %s", filename)
				t.Logf("Content:\n%s", contentStr)
				t.Logf("=== END GENERATED CONTENT ===")
			}
		})
	}
}

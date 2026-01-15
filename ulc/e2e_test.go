package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

type TestCase struct {
	Name          string
	SourceDir     string
	DotnetEnabled bool
	RustEnabled   bool
}

var e2eTestCases = []TestCase{
	{"basic", "../tests/basic", true, true},
	{"slice", "../tests/slice", true, true},
	{"complex", "../tests/complex", true, true},
	{"all", "../tests/all", true, true},
	{"contlib", "../libs/contlib", true, true},
	{"uql", "../libs/uql", true, true},
	{"iceberg", "../libs/iceberg", true, true},
	{"substrait", "../libs/substrait", true, true},
}

func TestE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	for _, tc := range e2eTestCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			runE2ETest(t, wd, tc)
		})
	}
}

func runE2ETest(t *testing.T, wd string, tc TestCase) {
	// Step 1: Generate code using go run
	t.Logf("Generating code for %s", tc.Name)
	cmd := exec.Command("go", "run", ".", fmt.Sprintf("--source=%s", tc.SourceDir), fmt.Sprintf("--output=%s", tc.Name))
	cmd.Dir = wd
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Code generation failed: %v\nOutput: %s", err, output)
	}
	t.Logf("Code generation output: %s", output)

	// Step 2: Compile C++
	t.Logf("Compiling C++ for %s", tc.Name)
	cppFile := filepath.Join(wd, tc.Name+".cpp")
	cmd = exec.Command("g++", "-std=c++17", cppFile)
	cmd.Dir = wd
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("C++ compilation failed: %v\nOutput: %s", err, output)
	}

	// Step 3: Compile C# if enabled
	if tc.DotnetEnabled {
		t.Logf("Compiling C# for %s", tc.Name)
		if err := compileDotnet(t, wd, tc.Name); err != nil {
			t.Fatalf("C# compilation failed: %v", err)
		}
	}

	// Step 4: Compile Rust if enabled
	if tc.RustEnabled {
		t.Logf("Compiling Rust for %s", tc.Name)
		if err := compileRust(t, wd, tc.Name); err != nil {
			t.Fatalf("Rust compilation failed: %v", err)
		}
	}

	// Cleanup
	cleanup(wd, tc.Name)
	t.Logf("Done with %s", tc.Name)
}

func compileDotnet(t *testing.T, wd, name string) error {
	projectDir := filepath.Join(wd, "dotnet_temp_"+name)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("failed to create project dir: %w", err)
	}
	defer os.RemoveAll(projectDir)

	// Create new console project
	cmd := exec.Command("dotnet", "new", "console", "--output", projectDir, "--force")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("dotnet new failed: %w\nOutput: %s", err, output)
	}

	// Copy generated .cs file to Program.cs
	csFile := filepath.Join(wd, name+".cs")
	programCs := filepath.Join(projectDir, "Program.cs")

	content, err := os.ReadFile(csFile)
	if err != nil {
		return fmt.Errorf("failed to read cs file: %w", err)
	}
	if err := os.WriteFile(programCs, content, 0644); err != nil {
		return fmt.Errorf("failed to write Program.cs: %w", err)
	}

	// Build the project
	cmd = exec.Command("dotnet", "build")
	cmd.Dir = projectDir
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("dotnet build failed: %w\nOutput: %s", err, output)
	}

	return nil
}

func compileRust(t *testing.T, wd, name string) error {
	projectDir := filepath.Join(wd, "rust_temp_"+name)
	srcDir := filepath.Join(projectDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		return fmt.Errorf("failed to create src dir: %w", err)
	}
	defer os.RemoveAll(projectDir)

	// Copy generated .rs file to src/main.rs
	rsFile := filepath.Join(wd, name+".rs")
	mainRs := filepath.Join(srcDir, "main.rs")

	content, err := os.ReadFile(rsFile)
	if err != nil {
		return fmt.Errorf("failed to read rs file: %w", err)
	}
	if err := os.WriteFile(mainRs, content, 0644); err != nil {
		return fmt.Errorf("failed to write main.rs: %w", err)
	}

	// Create Cargo.toml
	cargoToml := filepath.Join(projectDir, "Cargo.toml")
	cargoContent := fmt.Sprintf(`[package]
name = "%s"
version = "0.1.0"
edition = "2021"
`, name)
	if err := os.WriteFile(cargoToml, []byte(cargoContent), 0644); err != nil {
		return fmt.Errorf("failed to write Cargo.toml: %w", err)
	}

	// Build the project
	cmd := exec.Command("cargo", "build")
	cmd.Dir = projectDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("cargo build failed: %w\nOutput: %s", err, output)
	}

	return nil
}

func cleanup(wd, name string) {
	os.Remove(filepath.Join(wd, name+".cpp"))
	os.Remove(filepath.Join(wd, name+".cs"))
	os.Remove(filepath.Join(wd, name+".rs"))
	os.Remove(filepath.Join(wd, "a.out"))
}

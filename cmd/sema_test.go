package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// SemaTestCase represents a test case for unsupported construct detection
type SemaTestCase struct {
	Name           string
	Code           string
	ExpectedError  string
}

var semaTestCases = []SemaTestCase{
	{
		Name: "iota",
		Code: `package main

const (
	A = iota
	B
	C
)

func main() {
}
`,
		ExpectedError: "iota is not allowed",
	},
	{
		Name: "range_over_inline_literal",
		Code: `package main

func main() {
	for _, x := range []int{1, 2, 3} {
		_ = x
	}
}
`,
		ExpectedError: "range over inline slice literal",
	},
	{
		Name: "nil_comparison_eq",
		Code: `package main

func main() {
	var a []int
	if a == nil {
	}
}
`,
		ExpectedError: "nil comparison",
	},
	{
		Name: "nil_comparison_neq",
		Code: `package main

func main() {
	var a []int
	if a != nil {
	}
}
`,
		ExpectedError: "nil comparison",
	},
	// Note: empty_interface and slice_of_empty_interface tests removed
	// interface{} is now supported (maps to std::any/Box<dyn Any>/object)
	{
		Name: "string_variable_reuse_after_concat",
		Code: `package main

func main() {
	indent := "  "
	result := indent + "hello"
	result = result + indent
	_ = result
}
`,
		ExpectedError: "string variable reuse after concatenation",
	},
	{
		Name: "string_plusequal_self_concat",
		Code: `package main

func main() {
	result := "hello"
	indent := "  "
	result += result + indent
	_ = result
}
`,
		ExpectedError: "self-referencing string concatenation",
	},
	{
		Name: "struct_field_init_order",
		Code: `package main

type Person struct {
	Name string
	Age  int
	City string
}

func main() {
	p := Person{
		Age:  30,
		Name: "Alice",
		City: "NYC",
	}
	_ = p
}
`,
		ExpectedError: "struct field initialization order does not match declaration order",
	},
}

// SemaValidTestCase represents code that SHOULD compile successfully
type SemaValidTestCase struct {
	Name string
	Code string
}

var semaValidTestCases = []SemaValidTestCase{
	{
		Name: "string_reassign_then_reuse",
		Code: `package main

func main() {
	x := "hello"
	x = x + " world"
	y := x + "!"
	_ = y
}
`,
	},
	{
		Name: "string_plusequal_then_reuse",
		Code: `package main

func main() {
	x := "hello"
	x += " world"
	y := x + "!"
	_ = y
}
`,
	},
	{
		Name: "struct_field_init_correct_order",
		Code: `package main

type Person struct {
	Name string
	Age  int
	City string
}

func main() {
	p := Person{
		Name: "Alice",
		Age:  30,
		City: "NYC",
	}
	_ = p
}
`,
	},
}

func TestSemaValidConstructs(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	for _, tc := range semaValidTestCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			runSemaValidTest(t, wd, tc)
		})
	}
}

func runSemaValidTest(t *testing.T, wd string, tc SemaValidTestCase) {
	// Create temporary directory for test
	testDir := filepath.Join(os.TempDir(), "sema_valid_test_"+tc.Name)
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Create go.mod
	goMod := filepath.Join(testDir, "go.mod")
	if err := os.WriteFile(goMod, []byte("module sematest\ngo 1.24.4\n"), 0644); err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Create main.go with test code
	mainGo := filepath.Join(testDir, "main.go")
	if err := os.WriteFile(mainGo, []byte(tc.Code), 0644); err != nil {
		t.Fatalf("Failed to write main.go: %v", err)
	}

	// Run the compiler - it should succeed
	cmd := exec.Command("go", "run", ".", "--source="+testDir, "--output="+filepath.Join(testDir, "out"))
	cmd.Dir = wd
	output, err := cmd.CombinedOutput()

	// We expect the command to succeed
	if err != nil {
		t.Fatalf("Expected compilation to succeed for %s, but it failed.\nOutput: %s", tc.Name, output)
	}

	t.Logf("Correctly accepted %s", tc.Name)
}

func TestSemaUnsupportedConstructs(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	for _, tc := range semaTestCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			runSemaTest(t, wd, tc)
		})
	}
}

func runSemaTest(t *testing.T, wd string, tc SemaTestCase) {
	// Create temporary directory for test
	testDir := filepath.Join(os.TempDir(), "sema_test_"+tc.Name)
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Create go.mod
	goMod := filepath.Join(testDir, "go.mod")
	if err := os.WriteFile(goMod, []byte("module sematest\ngo 1.24.4\n"), 0644); err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Create main.go with test code
	mainGo := filepath.Join(testDir, "main.go")
	if err := os.WriteFile(mainGo, []byte(tc.Code), 0644); err != nil {
		t.Fatalf("Failed to write main.go: %v", err)
	}

	// Run the compiler - it should fail
	cmd := exec.Command("go", "run", ".", "--source="+testDir, "--output="+filepath.Join(testDir, "out"))
	cmd.Dir = wd
	output, err := cmd.CombinedOutput()

	// We expect the command to fail
	if err == nil {
		t.Fatalf("Expected compilation to fail for %s, but it succeeded.\nOutput: %s", tc.Name, output)
	}

	// Check that the expected error message is in the output
	if !strings.Contains(string(output), tc.ExpectedError) {
		t.Fatalf("Expected error containing %q for %s, but got:\n%s", tc.ExpectedError, tc.Name, output)
	}

	t.Logf("Correctly rejected %s with error: %s", tc.Name, tc.ExpectedError)
}

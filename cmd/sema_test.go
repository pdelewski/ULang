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
		Name: "for_key_value_range",
		Code: `package main

func main() {
	a := []int{1, 2, 3}
	for i, x := range a {
		_ = i
		_ = x
	}
}
`,
		ExpectedError: "for key, value := range is not allowed",
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
	{
		Name: "empty_interface",
		Code: `package main

func foo(x interface{}) {
}

func main() {
	foo(1)
}
`,
		ExpectedError: "empty interface",
	},
	{
		Name: "slice_of_empty_interface",
		Code: `package main

func main() {
	var a []interface{}
	_ = a
}
`,
		ExpectedError: "empty interface",
	},
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

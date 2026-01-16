package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Global debug flag
var DebugMode bool

// DebugPrintf prints only when debug mode is enabled
func DebugPrintf(format string, args ...interface{}) {
	if DebugMode {
		fmt.Printf(format, args...)
	}
}

// DebugLogPrintf logs only when debug mode is enabled
func DebugLogPrintf(format string, args ...interface{}) {
	if DebugMode {
		log.Printf(format, args...)
	}
}

func main() {
	var sourceDir string
	var output string
	var backend string
	var linkRuntime string
	flag.StringVar(&sourceDir, "source", "", "Source directory")
	flag.StringVar(&output, "output", "", "Output program name (can include path, e.g., ./build/project)")
	flag.StringVar(&backend, "backend", "all", "Backend to use: all, cpp, cs, rust (comma-separated for multiple)")
	flag.StringVar(&linkRuntime, "link-runtime", "", "Path to runtime for linking (generates Makefile with -I flag)")
	flag.BoolVar(&DebugMode, "debug", false, "Enable debug output")
	flag.Parse()
	if sourceDir == "" {
		fmt.Println("Please provide a source directory")
		return
	}

	// Parse output directory and name
	outputDir := filepath.Dir(output)
	outputName := filepath.Base(output)

	// Create output directory if it doesn't exist
	if outputDir != "." && outputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			log.Fatalf("Failed to create output directory: %v", err)
		}
	}

	// Note: We allow overwriting existing build files (Cargo.toml, Makefile, etc.)
	// to support iterative development
	cfg := &packages.Config{
		Mode:  packages.LoadSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedDeps | packages.NeedImports,
		Dir:   sourceDir,
		Tests: false,
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		fmt.Println("Error loading packages:", err)
		return
	}

	if len(pkgs) == 0 {
		fmt.Println("No packages found")
		return
	}

	// Parse backend selection
	backends := strings.Split(strings.ToLower(backend), ",")
	backendSet := make(map[string]bool)
	for _, b := range backends {
		backendSet[strings.TrimSpace(b)] = true
	}
	useAll := backendSet["all"]
	useCpp := useAll || backendSet["cpp"]
	useCs := useAll || backendSet["cs"]
	useRust := useAll || backendSet["rust"]

	// Build passes list
	sema := &BasePass{PassName: "Sema", emitter: &SemaChecker{Emitter: &BaseEmitter{}}}
	passes := []Pass{sema}
	var programFiles []string

	if useCpp {
		cppBackend := &BasePass{PassName: "CppGen", emitter: &CPPEmitter{
			Emitter:     &BaseEmitter{},
			Output:      output + ".cpp",
			LinkRuntime: linkRuntime,
			OutputDir:   outputDir,
			OutputName:  outputName,
		}}
		passes = append(passes, cppBackend)
		programFiles = append(programFiles, "cpp")
	}
	if useCs {
		csBackend := &BasePass{PassName: "CsGen", emitter: &CSharpEmitter{
			BaseEmitter: BaseEmitter{},
			Output:      output + ".cs",
			LinkRuntime: linkRuntime,
			OutputDir:   outputDir,
			OutputName:  outputName,
		}}
		passes = append(passes, csBackend)
		programFiles = append(programFiles, "cs")
	}
	if useRust {
		rustBackend := &BasePass{PassName: "RustGen", emitter: &RustEmitter{
			BaseEmitter: BaseEmitter{},
			Output:      output + ".rs",
			LinkRuntime: linkRuntime,
			OutputDir:   outputDir,
			OutputName:  outputName,
		}}
		passes = append(passes, rustBackend)
		programFiles = append(programFiles, "rs")
	}

	passManager := &PassManager{
		pkgs:   pkgs,
		passes: passes,
	}

	passManager.RunPasses()

	// Format generated files
	// Use astyle for C++/C#, rustfmt for Rust
	hasAstyleFiles := useCpp || useCs
	if hasAstyleFiles {
		DebugLogPrintf("Using astyle version: %s\n", GetAStyleVersion())
		const astyleOptions = "--style=webkit"

		if useCpp {
			filePath := fmt.Sprintf("%s.cpp", output)
			err = FormatFile(filePath, astyleOptions)
			if err != nil {
				log.Fatalf("Failed to format %s: %v", filePath, err)
			}
		}
		if useCs {
			filePath := fmt.Sprintf("%s.cs", output)
			err = FormatFile(filePath, astyleOptions)
			if err != nil {
				log.Fatalf("Failed to format %s: %v", filePath, err)
			}
		}
	}

	// Use rustfmt for Rust files
	if useRust {
		var rustFile string
		if linkRuntime != "" {
			// For Cargo projects, the file is in src/main.rs
			rustFile = filepath.Join(outputDir, "src", "main.rs")
		} else {
			rustFile = fmt.Sprintf("%s.rs", output)
		}
		cmd := exec.Command("rustfmt", rustFile)
		if err := cmd.Run(); err != nil {
			// rustfmt not available or failed - just log warning, don't fail
			log.Printf("Warning: rustfmt failed for %s: %v (install with: rustup component add rustfmt)", rustFile, err)
		} else {
			DebugLogPrintf("Successfully formatted: %s", rustFile)
		}
	}

	// Print colorful success message
	green := "\033[32m"
	bold := "\033[1m"
	reset := "\033[0m"
	checkmark := "✓"

	fmt.Printf("\n%s%s%s Transpilation successful!%s\n", bold, green, checkmark, reset)
	fmt.Printf("%s  Generated:%s %s\n", green, reset, strings.Join(programFiles, ", "))
}

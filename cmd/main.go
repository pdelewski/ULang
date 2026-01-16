package main

import (
	"flag"
	"fmt"
	"golang.org/x/tools/go/packages"
	"log"
	"os/exec"
	"strings"
)

func main() {
	var sourceDir string
	var output string
	var backend string
	flag.StringVar(&sourceDir, "source", "", "Source directory")
	flag.StringVar(&output, "output", "", "Output program name")
	flag.StringVar(&backend, "backend", "all", "Backend to use: all, cpp, cs, rust (comma-separated for multiple)")
	flag.Parse()
	if sourceDir == "" {
		fmt.Println("Please provide a source directory")
		return
	}
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
		cppBackend := &BasePass{PassName: "CppGen", emitter: &CPPEmitter{Emitter: &BaseEmitter{}, Output: output + ".cpp"}}
		passes = append(passes, cppBackend)
		programFiles = append(programFiles, "cpp")
	}
	if useCs {
		csBackend := &BasePass{PassName: "CsGen", emitter: &CSharpEmitter{BaseEmitter: BaseEmitter{}, Output: output + ".cs"}}
		passes = append(passes, csBackend)
		programFiles = append(programFiles, "cs")
	}
	if useRust {
		rustBackend := &BasePass{PassName: "RustGen", emitter: &RustEmitter{BaseEmitter: BaseEmitter{}, Output: output + ".rs"}}
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
		log.Printf("Using astyle version: %s\n", GetAStyleVersion())
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
		rustFile := fmt.Sprintf("%s.rs", output)
		cmd := exec.Command("rustfmt", rustFile)
		if err := cmd.Run(); err != nil {
			// rustfmt not available or failed - just log warning, don't fail
			log.Printf("Warning: rustfmt failed for %s: %v (install with: rustup component add rustfmt)", rustFile, err)
		} else {
			log.Printf("Successfully formatted: %s", rustFile)
		}
	}
}

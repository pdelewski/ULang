package main

import (
	"flag"
	"fmt"
	"golang.org/x/tools/go/packages"
	"log"
)

func main() {
	var sourceDir string
	var output string
	flag.StringVar(&sourceDir, "source", "", "./../uql")
	flag.StringVar(&output, "output", "", "Output program name")
	flag.Parse()
	if sourceDir == "" {
		fmt.Println("Please provide a source directory")
		return
	}
	cfg := &packages.Config{
		Mode:  packages.LoadSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedDeps,
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

	sema := &BasePass{PassName: "Sema", emitter: &SemaChecker{Emitter: &BaseEmitter{}}}
	cppBackend := &BasePass{PassName: "CppGen", emitter: &CPPEmitter{Emitter: &BaseEmitter{}, Output: output + ".cpp"}}
	csBackend := &BasePass{PassName: "CsGen", emitter: &CSharpEmitter{BaseEmitter: BaseEmitter{}, Output: output + ".cs"}}
	rustBackend := &BasePass{PassName: "RustGen", emitter: &RustEmitter{BaseEmitter: BaseEmitter{}, Output: output + ".rs"}}

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

	// Format generated files using astyle C API
	log.Printf("Using astyle version: %s\n", GetAStyleVersion())

	const astyleOptions = "--style=webkit"
	var programFiles = []string{
		"cpp",
		"cs",
		"rs",
	}

	for _, fileExt := range programFiles {
		filePath := fmt.Sprintf("%s.%s", output, fileExt)
		err = FormatFile(filePath, astyleOptions)
		if err != nil {
			log.Fatalf("Failed to format %s: %v", filePath, err)
		}
	}

}

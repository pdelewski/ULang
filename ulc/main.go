//go:generate make -C astyle astyle

package main

import (
	"flag"
	"fmt"
	"golang.org/x/tools/go/packages"
	"log"
	"os"
	"os/exec"
	"strings"
)

func RunCommand(command string) error {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return nil
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf("Running: %s\n", command)
	return cmd.Run()
}

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
	csBackend := &BasePass{PassName: "CsGen", emitter: &CSharpEmitter{Emitter: &BaseEmitter{}, Output: output + ".cs"}}
	rustBackend := &BasePass{PassName: "RustGen", emitter: &RustEmitter{Emitter: &BaseEmitter{}, Output: output + ".rs"}}

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

	const formatCmd = "./astyle/astyle --style=webkit"
	var programFiles = []string{
		"cpp",
		"cs",
		"rs",
	}

	for _, file := range programFiles {
		err = RunCommand(fmt.Sprintf("%s %s.%s", formatCmd, output, file))
		if err != nil {
			log.Fatalf("Command failed: %v", err)
		}
	}

}

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
	flag.StringVar(&sourceDir, "source", "", "./../uql")
	flag.Parse()
	if sourceDir == "" {
		fmt.Println("Please provide a source directory")
		return
	}
	cfg := &packages.Config{
		Mode:  packages.LoadSyntax,
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

	sema := &Sema{
		structs: make(map[string]StructInfo),
	}

	passManager := &PassManager{
		pkgs: pkgs,
		passes: []Pass{
			sema,
		},
	}

	cppBackend := &BasePass{PassName: "CppGen", emitter: &CPPEmitter{Emitter: &BaseEmitter{}}}
	csBackend := &BasePass{PassName: "CsGen", emitter: &CSharpEmitter{Emitter: &BaseEmitter{}}}
	passManager.passes = append(passManager.passes, cppBackend)
	passManager.passes = append(passManager.passes, csBackend)

	passManager.RunPasses()
	err = RunCommand("./astyle/astyle --style=webkit output.cpp")
	if err != nil {
		log.Fatalf("Command failed: %v", err)
	}
	err = RunCommand("./astyle/astyle --style=webkit Program.cs")
	if err != nil {
		log.Fatalf("Command failed: %v", err)
	}

}

package main

import (
	"flag"
	"fmt"
	"golang.org/x/tools/go/packages"
)

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

	format("output.cpp", "formatted.xx")
	format("Program.cs", "formatted.cs")
}

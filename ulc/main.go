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
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedSyntax |
			packages.NeedTypes,
		Dir: sourceDir,
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

	passManager := &PassManager{
		pkgs: pkgs,
		passes: []Pass{
			{
				name:    "Sema",
				visitor: &Sema{},
			},
		},
	}

	passManager.passes = append(passManager.passes, Pass{
		name:    "CppGen",
		visitor: &CppBackend{},
	})

	passManager.RunPasses()
}

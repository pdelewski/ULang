package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"golang.org/x/tools/go/packages"
	"os"
)

type Pass interface {
	Name() string
	Visitors() []ast.Visitor
	Finish()
}

type PassManager struct {
	pkgs   []*packages.Package
	passes []Pass
}

func (pm *PassManager) RunPasses() {
	for _, pass := range pm.passes {
		fmt.Printf("Running pass: %s\n", pass.Name())
		for _, pkg := range pm.pkgs {
			fmt.Printf("Package: %s\n", pkg.Name)
			fmt.Println("Files:")
			for _, file := range pkg.GoFiles {
				fmt.Printf("  %s\n", file)
				fset := token.NewFileSet()

				parsedFile, err := parser.ParseFile(fset, file, nil, parser.AllErrors)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error parsing file %s: %v\n", file, err)
					continue
				}
				for _, visitor := range pass.Visitors() {
					ast.Walk(visitor, parsedFile)
				}
			}
		}
		pass.Finish()
	}
}

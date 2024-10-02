package main

import (
	"fmt"
	"go/ast"
	"golang.org/x/tools/go/packages"
)

type Pass interface {
	Name() string
	Visitors(pkg *packages.Package) []ast.Visitor
	PostVisit(visitor ast.Visitor)
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
			visitors := pass.Visitors(pkg)

			for _, visitor := range visitors {
				for _, file := range pkg.Syntax {
					ast.Walk(visitor, file)
				}
			}
			for _, visitor := range visitors {
				pass.PostVisit(visitor)
			}
		}

		pass.Finish()
	}
}

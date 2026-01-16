package main

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/packages"
)

type Pass interface {
	ProLog()
	Name() string
	Visitors(pkg *packages.Package) []ast.Visitor
	PreVisit(visitor ast.Visitor)
	PostVisit(visitor ast.Visitor, visited map[string]struct{})
	EpiLog()
}

type PassManager struct {
	pkgs   []*packages.Package
	passes []Pass
}

func (pm *PassManager) RunPasses() {
	totalPasses := len(pm.passes)
	for i, pass := range pm.passes {
		DebugPrintf("Running pass: %s\n", pass.Name())
		// Show progress in normal mode
		if !DebugMode {
			fmt.Printf("\r[%d/%d] %s...", i+1, totalPasses, pass.Name())
		}
		visited := make(map[string]struct{})
		pass.ProLog()
		for _, pkg := range pm.pkgs {
			DebugPrintf("Package: %s\n", pkg.Name)
			DebugPrintf("Types Topological Sort: %v\n", pkg.TypesInfo)
			visitors := pass.Visitors(pkg)

			for _, visitor := range visitors {
				pass.PreVisit(visitor)
			}

			for _, visitor := range visitors {
				for _, file := range pkg.Syntax {
					ast.Walk(visitor, file)
				}
			}
			for _, visitor := range visitors {
				pass.PostVisit(visitor, visited)
			}
		}
		pass.EpiLog()
	}
	// Clear the progress line and show completion
	if !DebugMode {
		fmt.Printf("\r[%d/%d] Done.                    \n", totalPasses, totalPasses)
	}
}

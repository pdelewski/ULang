package main

import (
	"go/ast"
	"golang.org/x/tools/go/packages"
)

type CppBackend struct {
}

func (v *CppBackend) Name() string {
	return "CppGen"
}

func (v *CppBackend) Visitors(pkg *packages.Package) []ast.Visitor {
	return []ast.Visitor{v}
}

func (v *CppBackend) Visit(node ast.Node) ast.Visitor {
	return v
}

func (v *CppBackend) Finish() {
}

func (v *CppBackend) PostVisit(visitor ast.Visitor) {

}

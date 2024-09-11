package main

import (
	"go/ast"
)

type CppBackend struct {
}

func (v *CppBackend) Name() string {
	return "CppGen"
}

func (v *CppBackend) Visitors() []ast.Visitor {
	return []ast.Visitor{v}
}

func (v *CppBackend) Visit(node ast.Node) ast.Visitor {
	return v
}

func (v *CppBackend) Finish() {
}

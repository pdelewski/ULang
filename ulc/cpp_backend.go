package main

import (
	"go/ast"
)

type CppBackend struct{}

func (v *CppBackend) Visit(node ast.Node) ast.Visitor {
	return v
}

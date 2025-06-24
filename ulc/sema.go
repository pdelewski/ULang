package main

import (
	"go/ast"
	"golang.org/x/tools/go/packages"
)

type SemaChecker struct {
	Emitter
	pkg      *packages.Package
	constCtx bool
}

func (sema *SemaChecker) PreVisitGenDeclConstName(node *ast.Ident, indent int) {
	sema.constCtx = true
}

func (sema *SemaChecker) PreVisitIdent(node *ast.Ident, indent int) {
	if sema.constCtx {
		if node.String() == "iota" {
			panic("\033[31m\033[1miota is not allowed for now\033[0m")
		}
	}
}

func (sema *SemaChecker) PostVisitGenDeclConstName(node *ast.Ident, indent int) {
	sema.constCtx = false
}

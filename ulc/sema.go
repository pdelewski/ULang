package main

import (
	"fmt"
	"go/ast"
	"golang.org/x/tools/go/packages"
	"os"
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
			fmt.Println("\033[31m\033[1mCompilation error : iota is not allowed for now\033[0m")
			os.Exit(-1)
		}
	}
}

func (sema *SemaChecker) PostVisitGenDeclConstName(node *ast.Ident, indent int) {
	sema.constCtx = false
}

func (sema *SemaChecker) PreVisitRangeStmt(node *ast.RangeStmt, indent int) {
	if node.Key != nil {
		if node.Key.(*ast.Ident).Name != "_" {
			fmt.Println("\033[31m\033[1mCompilation error : for key, value := range is not allowed for now\033[0m")
			os.Exit(-1)
		}
	}
	node.Key = nil
}

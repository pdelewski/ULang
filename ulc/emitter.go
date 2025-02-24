package main

import (
	"go/ast"
	"os"
)

type Emitter interface {
	SetFile(file *os.File)
	PreVisitBasicLit(node *ast.BasicLit, indent int)
	PostVisitBasicLit(node *ast.BasicLit, indent int)
	PreVisitIdent(node *ast.Ident, indent int)
	PostVisitIdent(node *ast.Ident, indent int)
}

package main

import (
	"fmt"
	"go/ast"
	"go/token"
)

type Sema struct {
}

func (v *Sema) Name() string {
	return "Sema"
}

func (v *Sema) Visitor() ast.Visitor {
	return v
}

func (v *Sema) Visit(node ast.Node) ast.Visitor {
	switch decl := node.(type) {
	case *ast.GenDecl:
		// Check if it's a type declaration
		if decl.Tok != token.TYPE {
			return v
		}
		for _, spec := range decl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			// Check if it's a struct
			if structType, ok := typeSpec.Type.(*ast.StructType); ok {
				fmt.Printf("Found a struct: %s\n", typeSpec.Name.Name)

				for _, field := range structType.Fields.List {
					for _, fieldName := range field.Names {
						fmt.Printf("  Field: %s, Type: %s\n", fieldName.Name, field.Type)
					}
				}
			}
		}
	case *ast.FuncDecl:
		// Function declarations
		fmt.Printf("Found a function declaration: %s\n", decl.Name.Name)
	}
	// Continue traversing the AST
	return v
}

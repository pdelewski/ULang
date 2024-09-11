package main

import (
	"fmt"
	"go/ast"
	"go/token"
)

func (v *TypeVisitor) exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", v.exprToString(t.X), t.Sel.Name)
	case *ast.StarExpr:
		return "*" + v.exprToString(t.X)
	case *ast.ArrayType:
		return "[]" + v.exprToString(t.Elt)
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", v.exprToString(t.Key), v.exprToString(t.Value))
	case *ast.FuncType:
		v.dumpFuncType(t)
		return fmt.Sprintf("%T", expr)
	default:
		return fmt.Sprintf("%T", expr)
	}
}

type Sema struct {
}

func (s *Sema) Name() string {
	return "Sema"
}

func (s *Sema) Visitors() []ast.Visitor {
	return []ast.Visitor{&TypeVisitor{}}
}

type TypeVisitor struct {
}

func (v *TypeVisitor) dumpField(field *ast.Field) {
	for _, fieldName := range field.Names {
		fmt.Printf("  Field: %s, Type: %s\n", fieldName.Name, v.exprToString(field.Type))
	}
}

func (v *TypeVisitor) dumpFuncDecl(decl *ast.FuncDecl) {
	fmt.Printf("Found a function declaration: %s\n", decl.Name.Name)
	for _, param := range decl.Type.Params.List {
		v.dumpField(param)
	}
}

func (v *TypeVisitor) dumpFuncType(funcType *ast.FuncType) {
	// Dump parameters
	fmt.Printf("Function type:\n")
	fmt.Println("Parameters:")
	if funcType.Params != nil {
		for _, param := range funcType.Params.List {
			v.dumpField(param)
		}
	}

	// Dump return values
	fmt.Println("Return values:")
	if funcType.Results != nil {
		for _, result := range funcType.Results.List {
			if len(result.Names) > 0 {
				for _, resultName := range result.Names {
					fmt.Printf("  Return: %s, Type: %s\n", resultName.Name, v.exprToString(result.Type))
				}
			} else {
				fmt.Printf("  Type: %s\n", v.exprToString(result.Type))
			}
		}
	}
}

func (v *TypeVisitor) Visit(node ast.Node) ast.Visitor {
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
					v.dumpField(field)
				}
			}
		}
	case *ast.FuncDecl:
		v.dumpFuncDecl(decl)
	}
	// Continue traversing the AST
	return v
}

func (s *Sema) Finish() {
}

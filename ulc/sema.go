package main

import (
	"fmt"
	"go/ast"
)

var allowedTypes = map[string]struct{}{
	"int8":    {},
	"int16":   {},
	"int32":   {},
	"int64":   {},
	"uint8":   {},
	"uint16":  {},
	"uint32":  {},
	"uint64":  {},
	"float32": {},
	"float64": {},
}

type StructDescriptor struct {
	StructType *ast.StructType
	PkgName    string
}

type StructVisitor struct {
	structs map[string]StructDescriptor // Stores all struct definitions
}

// Visit implements the ast.Visitor interface, called for each node
func (v *StructVisitor) Visit(n ast.Node) ast.Visitor {
	switch ts := n.(type) {
	case *ast.TypeSpec:
		// Check if the type is a struct and store it
		if st, ok := ts.Type.(*ast.StructType); ok {
			v.structs[ts.Name.Name] = StructDescriptor{StructType: st}
		}
	}
	return v
}

func (s *Sema) inspectStruct(name string, indent int) {
	st, ok := s.structs[name]
	if !ok {
		fmt.Printf("%sUnknown struct: %s\n", indentStr(indent), name)
		return
	}

	for _, field := range st.StructType.Fields.List {
		for _, fieldName := range field.Names {
			fieldTypeStr := fieldType(field.Type)
			fmt.Printf("%sField: %s, Type: %s\n", indentStr(indent), fieldName.Name, fieldTypeStr)

			// If the field is a struct, recursively inspect it
			if structType, isStruct := field.Type.(*ast.Ident); isStruct && s.isStructType(structType) {
				s.inspectStruct(structType.Name, indent+2)
			}
			if selType, isSelector := field.Type.(*ast.SelectorExpr); isSelector {
				fmt.Printf("%sExternal type: %s.%s\n", indentStr(indent+2), selType.X, selType.Sel)
			}
		}
	}
}

// isStructType checks if a field type is a struct defined in the same file
func (s *Sema) isStructType(ident *ast.Ident) bool {
	_, exists := s.structs[ident.Name]
	return exists
}

func fieldType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.ArrayType:
		return "[]" + fieldType(t.Elt)
	case *ast.StarExpr:
		return "*" + fieldType(t.X)
	case *ast.SelectorExpr:
		return fieldType(t.X) + "." + t.Sel.Name
	default:
		return fmt.Sprintf("%T", expr)
	}
}

// indentStr returns a string of spaces for indentation
func indentStr(indent int) string {
	return fmt.Sprintf("%s", " "+fmt.Sprintf("%*s", indent, ""))
}

type Sema struct {
	structs map[string]StructDescriptor
}

func (s *Sema) Visitors() []ast.Visitor {
	return []ast.Visitor{&StructVisitor{s.structs}}
}

func (s *Sema) Name() string {
	return "Sema"
}

func (s *Sema) Finish() {
	for structName := range s.structs {
		fmt.Printf("Inspecting struct: %s\n", structName)
		s.inspectStruct(structName, 0)
	}
}

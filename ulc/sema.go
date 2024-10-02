package main

import (
	"fmt"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/packages"
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

// StructInfo holds the name and the type of a struct for recursive inspection
type StructInfo struct {
	Name       string
	Struct     *types.Struct
	IsExternal bool // Whether this struct is external or local
	Pkg        string
}

type SemaVisitor struct {
	pkg     *packages.Package
	structs map[string]StructInfo // Map to store structs by name
}

func (v *SemaVisitor) Visit(n ast.Node) ast.Visitor {
	switch node := n.(type) {
	case *ast.TypeSpec:
		if _, ok := node.Type.(*ast.StructType); ok {
			if _, ok := v.structs[node.Name.Name]; !ok {
				v.structs[node.Name.Name] = StructInfo{
					Name:       node.Name.Name,
					Struct:     nil, // We don't have type info for local structs yet
					IsExternal: false,
					Pkg:        v.pkg.Name,
				}
			}
		}
	}
	return v
}

func (v *SemaVisitor) inspectStruct(name string, indent int) {
	info, ok := v.structs[name]
	if !ok {
		fmt.Printf("%sUnknown struct: %s\n", indentStr(indent), name)
		return
	}

	// If the struct is local, print its fields
	if !info.IsExternal {
		fmt.Printf("%sInspecting struct: %s\n", indentStr(indent), name)
		for _, field := range v.pkg.Syntax {
			ast.Inspect(field, func(n ast.Node) bool {
				switch ts := n.(type) {
				case *ast.TypeSpec:
					if ts.Name.Name == name {
						if st, ok := ts.Type.(*ast.StructType); ok {
							for _, field := range st.Fields.List {
								fieldType := v.inspectFieldType(field.Type)
								for _, fieldName := range field.Names {
									fmt.Printf("%sField: %s, Type: %s\n", indentStr(indent+2), fieldName.Name, fieldType)
									v.inspectFieldRecursively(field.Type, indent+2)
								}
							}
						}
					}
				}
				return true
			})
		}
	} else {
		// For external structs, use the type information to print the fields
		fmt.Printf("%sInspecting struct: %s\n", indentStr(indent), name)
		v.inspectExternalStruct(info.Struct, indent+2)
	}
}

// inspectFieldRecursively inspects the fields of the struct recursively
func (v *SemaVisitor) inspectFieldRecursively(expr ast.Expr, indent int) {
	switch typ := expr.(type) {
	case *ast.Ident: // Local struct or basic types
		if obj := v.pkg.TypesInfo.Uses[typ]; obj != nil {
			if named, ok := obj.Type().(*types.Named); ok {
				if structType, ok := named.Underlying().(*types.Struct); ok {
					fmt.Printf("%sRecursively inspecting struct: %s from package: %s\n", indentStr(indent), named.Obj().Name(), named.Obj().Pkg().Path())
					v.structs[named.Obj().Name()] = StructInfo{
						Name:       named.Obj().Name(),
						Struct:     structType,
						IsExternal: true,
						Pkg:        named.Obj().Pkg().Name(),
					}
					v.inspectExternalStruct(structType, indent+2)
				}
			}
		}

	case *ast.SelectorExpr: // External struct from another package
		if obj := v.pkg.TypesInfo.Uses[typ.Sel]; obj != nil {
			if named, ok := obj.Type().(*types.Named); ok {
				if structType, ok := named.Underlying().(*types.Struct); ok {
					fmt.Printf("%sRecursively inspecting external struct: %s from package: %s\n", indentStr(indent), named.Obj().Name(), named.Obj().Pkg().Path())
					v.structs[named.Obj().Name()] = StructInfo{
						Name:       named.Obj().Name(),
						Struct:     structType,
						IsExternal: true,
						Pkg:        named.Obj().Pkg().Name(),
					}
					v.inspectExternalStruct(structType, indent+2)
				}
			}
		}
	}
}

func (v *SemaVisitor) inspectExternalStruct(structType *types.Struct, indent int) {
	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		fieldType := field.Type().String()
		fmt.Printf("%sField: %s, Type: %s\n", indentStr(indent), field.Name(), fieldType)

		if named, ok := field.Type().(*types.Named); ok {
			if nestedStruct, ok := named.Underlying().(*types.Struct); ok {
				fmt.Printf("%sRecursively inspecting nested struct: %s\n", indentStr(indent+2), named.Obj().Name())
				v.inspectExternalStruct(nestedStruct, indent+4)
			}
		}
	}
}

func (v *SemaVisitor) inspectFieldType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident: // Basic types or local structs
		return t.Name
	case *ast.SelectorExpr: // Imported types
		if pkgIdent, ok := t.X.(*ast.Ident); ok {
			return fmt.Sprintf("%s.%s", pkgIdent.Name, t.Sel.Name)
		}
	case *ast.StarExpr: // Pointer to a type
		return "*" + v.inspectFieldType(t.X)
	case *ast.ArrayType: // Array of types
		return "[]" + v.inspectFieldType(t.Elt)
	default:
		return fmt.Sprintf("%T", expr)
	}
	return "unknown"
}

func indentStr(indent int) string {
	return fmt.Sprintf("%s", " "+fmt.Sprintf("%*s", indent, ""))
}

type Sema struct {
	structs  map[string]StructInfo
	visitors []ast.Visitor
}

func (s *Sema) Visitors(pkg *packages.Package) []ast.Visitor {
	return []ast.Visitor{&SemaVisitor{structs: s.structs, pkg: pkg}}

}

func (s *Sema) Name() string {
	return "Sema"
}

func (s *Sema) PostVisit(visitor ast.Visitor) {
	semaVisitor := visitor.(*SemaVisitor)
	for name, val := range semaVisitor.structs {
		if val.Pkg != semaVisitor.pkg.Name {
			continue
		}
		semaVisitor.inspectStruct(name, 0)
	}
}

func (s *Sema) Finish() {
}

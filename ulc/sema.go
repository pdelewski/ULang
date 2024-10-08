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

func (v *SemaVisitor) inspectStruct(name string, indent int, visited map[string]struct{}) {
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
								//fieldType := v.inspectFieldType(field.Type)
								for _, fieldName := range field.Names {
									v.inspectFieldRecursively(fieldName.Name, field.Type, indent+2, visited)
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
		v.inspectExternalStruct(name, info.Struct, indent+2, visited)
	}
}

// inspectFieldRecursively inspects the fields of the struct recursively
func (v *SemaVisitor) inspectFieldRecursively(fieldName string, expr ast.Expr, indent int, visited map[string]struct{}) {
	switch typ := expr.(type) {
	case *ast.Ident: // Local struct or basic types
		if obj := v.pkg.TypesInfo.Uses[typ]; obj != nil {
			if named, ok := obj.Type().(*types.Named); ok {
				if structType, ok := named.Underlying().(*types.Struct); ok {
					//fmt.Printf("%sRecursively inspecting struct: %s from package: %s\n", indentStr(indent), named.Obj().Name(), named.Obj().Pkg().Path())
					v.structs[named.Obj().Name()] = StructInfo{
						Name:       named.Obj().Name(),
						Struct:     structType,
						IsExternal: true,
						Pkg:        named.Obj().Pkg().Name(),
					}
					v.inspectExternalStruct(named.Obj().Name(), structType, indent+2, visited)
				}
			} else {
				//				fmt.Printf("%sField: %s, Type: %s\n", indentStr(indent), fieldName, obj.Type().String())
				if _, ok := allowedTypes[obj.Type().String()]; !ok {
					fmt.Printf("Sema failed: %sField: %s, Type: %s\n", indentStr(indent), fieldName, obj.Type().String())
				}
			}
		}

	case *ast.SelectorExpr: // External struct from another package
		if obj := v.pkg.TypesInfo.Uses[typ.Sel]; obj != nil {
			if named, ok := obj.Type().(*types.Named); ok {
				if structType, ok := named.Underlying().(*types.Struct); ok {
					fmt.Printf("%sInspecting struct: %s from package: %s\n", indentStr(indent), named.Obj().Name(), named.Obj().Pkg().Path())
					v.structs[named.Obj().Name()] = StructInfo{
						Name:       named.Obj().Name(),
						Struct:     structType,
						IsExternal: true,
						Pkg:        named.Obj().Pkg().Name(),
					}
					v.inspectExternalStruct(named.Obj().Name(), structType, indent+2, visited)
				}
			} else {
				//fmt.Printf("%sField: %s, Type: %s\n", indentStr(indent), fieldName, named.Obj().Name())
				if _, ok := allowedTypes[obj.Type().String()]; !ok {
					fmt.Printf("Sema failed: %sField: %s, Type: %s\n", indentStr(indent), fieldName, obj.Type().String())
				}
			}
		}
	case *ast.FuncType:
		//fmt.Printf("%sField: %s\n", indentStr(indent), "func")
	}

}

func (v *SemaVisitor) inspectExternalStruct(structName string, structType *types.Struct, indent int, visited map[string]struct{}) {
	if _, ok := visited[structName]; ok {
		return
	}
	visited[structName] = struct{}{}
	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		fieldType := field.Type().String()
		//fmt.Printf("%sField: %s, Type: %s\n", indentStr(indent), field.Name(), fieldType)

		// Handle named types
		if named, ok := field.Type().(*types.Named); ok {
			if nestedStruct, ok := named.Underlying().(*types.Struct); ok {
				fmt.Printf("%sInspecting struct: %s\n", indentStr(indent+2), named.Obj().Name())
				v.inspectExternalStruct(structName, nestedStruct, indent+4, visited)
			}
		} else if slice, ok := field.Type().(*types.Slice); ok {
			elemType := slice.Elem()
			//fmt.Printf("%sField: %s is a slice of: %s\n", indentStr(indent), field.Name(), elemType.String())
			if namedElem, ok := elemType.(*types.Named); ok {
				if nestedStruct, ok := namedElem.Underlying().(*types.Struct); ok {
					v.inspectExternalStruct(structName, nestedStruct, indent+4, visited)
				}
			}
		} else {
			//fmt.Printf("%sField: %s, Type: %s\n", indentStr(indent), field.Name(), fieldType)
			if _, ok := allowedTypes[fieldType]; !ok {
				fmt.Printf("Sema failed: %sField: %s, Type: %s\n", indentStr(indent), field.Name(), fieldType)
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

func (s *Sema) PreVisit(visitor ast.Visitor) {
}

func (s *Sema) PostVisit(visitor ast.Visitor, visited map[string]struct{}) {
	semaVisitor := visitor.(*SemaVisitor)
	for name, val := range semaVisitor.structs {
		if val.Pkg != semaVisitor.pkg.Name {
			continue
		}
		if _, ok := visited[name]; ok {
			return
		}
		semaVisitor.inspectStruct(name, 0, visited)
	}
}

func (s *Sema) ProLog() {
}

func (s *Sema) EpiLog() {
}

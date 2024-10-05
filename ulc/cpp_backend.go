package main

import (
	"fmt"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/packages"
	"os"
)

type CppBackend struct {
	outputFile string
	file       *os.File
	visitor    *CppBackendVisitor
}

type CppBackendVisitor struct {
	pkg  *packages.Package
	pass *CppBackend
}

func (v *CppBackend) Name() string {
	return "CppGen"
}

func (v *CppBackend) Visitors(pkg *packages.Package) []ast.Visitor {
	v.visitor = &CppBackendVisitor{pkg: pkg}
	v.visitor.pass = v
	return []ast.Visitor{v.visitor}
}

func (v *CppBackendVisitor) generateFields(st *ast.StructType) {
	for _, field := range st.Fields.List {
		for _, fieldName := range field.Names {
			switch typ := field.Type.(type) {
			case *ast.Ident:
				_, err := v.pass.file.WriteString(fmt.Sprintf("  %s %s;\n", typ.Name, fieldName.Name))
				if err != nil {
					fmt.Println("Error writing to file:", err)
				}
			case *ast.SelectorExpr: // External struct from another package
				if obj := v.pkg.TypesInfo.Uses[typ.Sel]; obj != nil {
					if named, ok := obj.Type().(*types.Named); ok {
						if _, ok := named.Underlying().(*types.Struct); ok {
							_, err := v.pass.file.WriteString(fmt.Sprintf("  %s.%s %s;\n", named.Obj().Pkg().Name(), named.Obj().Name(), fieldName.Name))
							if err != nil {
								fmt.Println("Error writing to file:", err)
							}
						}
					}
				}
			case *ast.ArrayType:
				switch elt := typ.Elt.(type) {
				case *ast.Ident:
					_, err := v.pass.file.WriteString(fmt.Sprintf("  std::vector<%s> %s;\n", elt.Name, fieldName.Name))
					if err != nil {
						fmt.Println("Error writing to file:", err)
					}
				}
			}
		}
	}
}

func (v *CppBackendVisitor) Visit(node ast.Node) ast.Visitor {
	switch node := node.(type) {
	case *ast.TypeSpec:
		if st, ok := node.Type.(*ast.StructType); ok {
			structInfo := StructInfo{
				Name:       node.Name.Name,
				Struct:     nil, // We don't have type info for local structs yet
				IsExternal: false,
				Pkg:        v.pkg.Name,
			}
			//fmt.Println("struct", structInfo.Name)
			//fmt.Println("{")
			_, err := v.pass.file.WriteString(fmt.Sprintf("struct %s\n", structInfo.Name))
			if err != nil {
				fmt.Println("Error writing to file:", err)
				return v
			}
			_, err = v.pass.file.WriteString("{\n")
			if err != nil {
				fmt.Println("Error writing to file:", err)
				return v
			}
			v.generateFields(st)
			//fmt.Println("};")
			_, err = v.pass.file.WriteString("};\n\n")
			if err != nil {
				fmt.Println("Error writing to file:", err)
				return v
			}
		}
	}
	return v
}

func (v *CppBackend) ProLog() {
	v.outputFile = "./output.cpp"
	var err error
	v.file, err = os.Create(v.outputFile)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
}

func (v *CppBackend) EpiLog() {
	v.file.Close()
}

func (v *CppBackend) PreVisit(visitor ast.Visitor) {
	cppVisitor := visitor.(*CppBackendVisitor)
	//fmt.Println("namespace", cppBackend.pkg.Name)
	//fmt.Println("{")
	_, err := cppVisitor.pass.file.WriteString(fmt.Sprintf("namespace %s\n", cppVisitor.pkg.Name))
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
	_, err = cppVisitor.pass.file.WriteString("{\n\n")
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}

func (v *CppBackend) PostVisit(visitor ast.Visitor, visited map[string]struct{}) {
	cppVisitor := visitor.(*CppBackendVisitor)
	//fmt.Println("} // namespace", cppBackend.pkg.Name)
	_, err := cppVisitor.pass.file.WriteString(fmt.Sprintf("} // namespace %s\n\n", cppVisitor.pkg.Name))
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}

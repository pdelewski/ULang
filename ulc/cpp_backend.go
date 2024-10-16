package main

import (
	"fmt"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/packages"
	"os"
)

var typesMap = map[string]string{
	"int8":   "std::int8_t",
	"int16":  "std::int16_t",
	"int32":  "std::int32_t",
	"int64":  "std::int64_t",
	"uint8":  "std::uint8_t",
	"uint16": "std::uint16_t",
	"any":    "std::any",
	"string": "std::string",
}

type ArrayTypeGen int

const (
	ArrayStructField ArrayTypeGen = iota
	ArrayArgument
	ArrayAlias
	ArrayReturn
)

type CppBackend struct {
	outputFile string
	file       *os.File
	visitor    *CppBackendVisitor
}

type CppBackendVisitor struct {
	pkg   *packages.Package
	pass  *CppBackend
	nodes []ast.Node
}

func (v *CppBackend) Name() string {
	return "CppGen"
}

func (v *CppBackend) Visitors(pkg *packages.Package) []ast.Visitor {
	v.visitor = &CppBackendVisitor{pkg: pkg}
	v.visitor.pass = v
	return []ast.Visitor{v.visitor}
}

func (v *CppBackendVisitor) emit(s string) error {
	_, err := v.pass.file.WriteString(s)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return err
	}
	return nil
}

func (v *CppBackendVisitor) generateArrayType(typ *ast.ArrayType, fieldName string, arrayType ArrayTypeGen) {
	var err error
	switch elt := typ.Elt.(type) {
	case *ast.Ident:
		cppType := elt.Name
		if val, ok := typesMap[elt.Name]; ok {
			cppType = val
		}
		switch arrayType {
		case ArrayStructField:
			if len(fieldName) == 0 {
				panic("expected field")
			}
			err = v.emit(fmt.Sprintf("  std::vector<%s> %s;\n", cppType, fieldName))
		case ArrayArgument:
			if len(fieldName) == 0 {
				panic("expected field")
			}
			err = v.emit(fmt.Sprintf("std::vector<%s> %s", cppType, fieldName))
		case ArrayReturn:
			err = v.emit(fmt.Sprintf("std::vector<%s>", cppType))
		case ArrayAlias:
			if len(fieldName) == 0 {
				panic("expected field")
			}
			err = v.emit(fmt.Sprintf("using %s = std::vector<%s>;\n\n", fieldName, cppType))
		}
		if err != nil {
			fmt.Println("Error writing to file:", err)
		}
	case *ast.SelectorExpr: // Imported types
		if pkgIdent, ok := elt.X.(*ast.Ident); ok {
			cppType := elt.Sel.Name
			if val, ok := typesMap[elt.Sel.Name]; ok {
				cppType = val
			}
			switch arrayType {
			case ArrayStructField:
				if len(fieldName) == 0 {
					panic("expected field")
				}
				err = v.emit(fmt.Sprintf("  std::vector<%s::%s> %s;\n", pkgIdent.Name, cppType, fieldName))
			case ArrayArgument:
				if len(fieldName) == 0 {
					panic("expected field")
				}
				err = v.emit(fmt.Sprintf("std::vector<%s::%s> %s", pkgIdent.Name, cppType, fieldName))
			case ArrayReturn:
				err = v.emit(fmt.Sprintf("std::vector<%s::%s>", pkgIdent.Name, cppType))
			case ArrayAlias:
				if len(fieldName) == 0 {
					panic("expected field")
				}
				err = v.emit(fmt.Sprintf("using %s = std::vector<%s::%s>\n\n", fieldName, pkgIdent.Name, cppType))
			}
		}
		if err != nil {
			fmt.Println("Error writing to file:", err)
		}
	}
}

func (v *CppBackendVisitor) generatePrimType(cppType string, pkg string, fieldName string) {
	if val, ok := typesMap[cppType]; ok {
		cppType = val
	}
	var err error
	if fieldName == "" {
		err = v.emit(cppType)
	} else if pkg != "" {
		err = v.emit(fmt.Sprintf("%s::%s %s", pkg, cppType, fieldName))
	} else {
		err = v.emit(fmt.Sprintf("%s %s", cppType, fieldName))
	}
	if err != nil {
		fmt.Println("Error writing to file:", err)
	}
}

func (v *CppBackendVisitor) generateFields(st *ast.StructType) {
	for _, field := range st.Fields.List {
		for _, fieldName := range field.Names {
			switch typ := field.Type.(type) {
			case *ast.Ident:
				cppType := typ.Name
				if val, ok := typesMap[typ.Name]; ok {
					cppType = val
				}
				err := v.emit(fmt.Sprintf("  %s %s;\n", cppType, fieldName.Name))
				if err != nil {
					fmt.Println("Error writing to file:", err)
				}
			case *ast.SelectorExpr: // External struct from another package
				if obj := v.pkg.TypesInfo.Uses[typ.Sel]; obj != nil {
					if named, ok := obj.Type().(*types.Named); ok {
						if _, ok := named.Underlying().(*types.Struct); ok {
							v.emit("  ")
							v.generatePrimType(named.Obj().Name(), named.Obj().Pkg().Name(), fieldName.Name)
							v.emit(";\n")
						}
					}
				}
			case *ast.ArrayType:
				v.generateArrayType(typ, fieldName.Name, ArrayStructField)
			}
		}
	}
}

func (v *CppBackendVisitor) inspectType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident: // Basic types or local structs
		return t.Name
	case *ast.SelectorExpr: // Imported types
		if pkgIdent, ok := t.X.(*ast.Ident); ok {
			return fmt.Sprintf("%s::%s", pkgIdent.Name, t.Sel.Name)
		}
	case *ast.StarExpr: // Pointer to a type
		return "*" + v.inspectType(t.X)
	case *ast.ArrayType: // Array of types
		return "[]" + v.inspectType(t.Elt)
	default:
		return fmt.Sprintf("%T", expr)
	}
	return "unknown"
}

func (v *CppBackendVisitor) generateFuncDecl(node *ast.FuncDecl) ast.Visitor {
	if node.Type.Results != nil {
		resultArgIndex := 0
		if len(node.Type.Results.List) > 0 {
			err := v.emit("std::tuple<")
			if err != nil {
				fmt.Println("Error writing to file:", err)
				return v
			}
		}
		for _, result := range node.Type.Results.List {
			if resultArgIndex > 0 {
				err := v.emit(",")
				if err != nil {
					fmt.Println("Error writing to file:", err)
					return v
				}
			}
			if arrayArg, ok := result.Type.(*ast.ArrayType); ok {
				v.generateArrayType(arrayArg, "", ArrayReturn)
			} else {
				v.generatePrimType(v.inspectType(result.Type), "", "")
			}
			resultArgIndex++
		}
		if len(node.Type.Results.List) > 0 {
			err := v.emit(">")
			if err != nil {
				fmt.Println("Error writing to file:", err)
				return v
			}
		}
	} else if node.Name.Name == "main" {
		err := v.emit("int")
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return v
		}
	} else {
		err := v.emit("void")
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return v
		}
	}
	err := v.emit(" ")
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return v
	}
	err = v.emit(node.Name.Name + "(")
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return v
	}
	argIndex := 0
	for _, arg := range node.Type.Params.List {
		if argIndex > 0 {
			err = v.emit(", ")
			if err != nil {
				fmt.Println("Error writing to file:", err)
				return v
			}
		}
		for _, argName := range arg.Names {
			if arrayArg, ok := arg.Type.(*ast.ArrayType); ok {
				v.generateArrayType(arrayArg, argName.Name, ArrayArgument)
			} else {
				v.generatePrimType(v.inspectType(arg.Type), "", argName.Name)
			}
		}
		argIndex++
	}
	err = v.emit(")\n")
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return v
	}
	err = v.emit("{\n")
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return v
	}
	err = v.emit("}\n")
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return v
	}
	return v
}

func (v *CppBackendVisitor) gen() {
	for _, node := range v.nodes {
		switch node := node.(type) {
		case *ast.TypeSpec:
			if st, ok := node.Type.(*ast.StructType); ok {
				structInfo := StructInfo{
					Name:       node.Name.Name,
					Struct:     nil, // We don't have type info for local structs yet
					IsExternal: false,
					Pkg:        v.pkg.Name,
				}
				err := v.emit(fmt.Sprintf("struct %s\n", structInfo.Name))
				if err != nil {
					fmt.Println("Error writing to file:", err)
				}
				err = v.emit("{\n")
				if err != nil {
					fmt.Println("Error writing to file:", err)
				}
				v.generateFields(st)
				err = v.emit("};\n\n")
				if err != nil {
					fmt.Println("Error writing to file:", err)
				}
			} else {
				if arrayArg, ok := node.Type.(*ast.ArrayType); ok {
					v.generateArrayType(arrayArg, node.Name.Name, ArrayAlias)
				} else {
					err := v.emit(fmt.Sprintf("using %s = %s;\n\n", node.Name.Name, v.inspectType(node.Type)))
					if err != nil {
						fmt.Println("Error writing to file:", err)
					}
				}
			}
		case *ast.FuncDecl:
			v.generateFuncDecl(node)
		}
	}
}

func (v *CppBackendVisitor) Visit(node ast.Node) ast.Visitor {
	v.nodes = append(v.nodes, node)
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
	_, err = v.file.WriteString("#include <vector>\n#include <string>\n#include <tuple>\n#include <any>\n#include <cstdint>\n\n")
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}

func (v *CppBackend) EpiLog() {
	v.file.Close()
}

func (v *CppBackend) PreVisit(visitor ast.Visitor) {
	cppVisitor := visitor.(*CppBackendVisitor)
	if cppVisitor.pkg.Name == "main" {
		return
	}
	err := cppVisitor.emit(fmt.Sprintf("namespace %s\n", cppVisitor.pkg.Name))
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
	cppVisitor.gen()
	if cppVisitor.pkg.Name == "main" {
		return
	}
	err := cppVisitor.emit(fmt.Sprintf("} // namespace %s\n\n", cppVisitor.pkg.Name))
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}

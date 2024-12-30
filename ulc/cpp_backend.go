package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/packages"
	"log"
	"os"
	"sort"
	"strings"
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

var primTypes = map[string]struct{}{
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

type ArrayTypeGen int

const (
	ArrayStructField ArrayTypeGen = iota
	ArrayArgument
	ArrayAlias
	ArrayReturn
)

type GenStructInfo struct {
	Name       string
	Struct     *ast.StructType
	IsExternal bool // Whether this struct is external or local
	Pkg        string
}

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

func getFunctionName(callExpr *ast.CallExpr) string {
	switch fun := callExpr.Fun.(type) {
	case *ast.Ident:
		// Direct function call, e.g., `foo()`
		return fun.Name
	case *ast.SelectorExpr:
		// Selector expression, e.g., `pkg.Func()`
		return fun.Sel.Name
	default:
		return "<unknown>"
	}
}

func (v *CppBackendVisitor) generateCallExpr(node *ast.CallExpr) error {
	var err error
	err = v.emit("  " + getFunctionName(node))
	if err != nil {
		return err
	}
	err = v.emit("(")
	if err != nil {
		return err
	}
	for i, arg := range node.Args {
		if i > 0 {
			err = v.emit(", ")
			if err != nil {
				return err
			}
		}
		switch arg := arg.(type) {
		case *ast.Ident:
			err = v.emit(arg.Name)
			if err != nil {
				return err
			}
		case *ast.BasicLit:
			err = v.emit(arg.Value)
			if err != nil {
				return err
			}
		case *ast.SelectorExpr:
			err = v.emit(arg.Sel.Name)
			if err != nil {
				return err
			}
		case *ast.CallExpr:
			err = v.generateCallExpr(arg)
			if err != nil {
				return err
			}
		}
	}
	return v.emit(");\n")
}

func (v *CppBackendVisitor) emitExpression(expr ast.Expr) {
	switch e := expr.(type) {
	case *ast.BasicLit:
		v.emit(e.Value) // Basic literals like numbers or strings
	case *ast.Ident:
		v.emit(e.Name) // Variables or identifiers
	case *ast.BinaryExpr:
		v.emit("(")
		v.emitExpression(e.X) // Left operand
		v.emit(" " + e.Op.String() + " ")
		v.emitExpression(e.Y) // Right operand
		v.emit(")")
	case *ast.CallExpr:
		if fun, ok := e.Fun.(*ast.Ident); ok {
			v.emit(fun.Name + "(")
			for i, arg := range e.Args {
				if i > 0 {
					v.emit(", ")
				}
				v.emitExpression(arg) // Function arguments
			}
			v.emit(")")
		} else {
			fmt.Println("<complex call expression>")
		}
	default:
		fmt.Println("<unknown expression>")
	}
}

func (v *CppBackendVisitor) emitAssignment(assignStmt *ast.AssignStmt) {
	assignmentToken := assignStmt.Tok.String()
	v.emit("  ")
	if assignmentToken == ":=" {
		v.emit("auto ")
		assignmentToken = "="
	}
	if len(assignStmt.Lhs) > 1 {
		v.emit("std::tie(")
	}
	first := true
	for _, lhs := range assignStmt.Lhs {
		if !first {
			v.emit(", ")
		}
		first = false
		if ident, ok := lhs.(*ast.Ident); ok {
			v.emit(ident.Name)
		}
	}

	if len(assignStmt.Lhs) > 1 {
		v.emit(")")
	}

	v.emit(" " + assignmentToken + " ")

	for _, rhs := range assignStmt.Rhs {
		v.emitExpression(rhs)
	}
	v.emit(";\n")
}

func (v *CppBackendVisitor) emitReturnStmt(retStmt *ast.ReturnStmt) {
	v.emit("  return ")
	if len(retStmt.Results) > 1 {
		v.emit("std::make_tuple(")
	}
	first := true
	for _, result := range retStmt.Results {
		if !first {
			v.emit(", ")
		}
		first = false
		v.emitExpression(result)
	}
	if len(retStmt.Results) > 1 {
		v.emit(")")
	}
	v.emit(";\n")
}

func (v *CppBackendVisitor) generateIndent(level int) {
	v.emit(strings.Repeat("  ", level))
}

func (v *CppBackendVisitor) emitBlockStmt(block *ast.BlockStmt, indent int) {
	for _, stmt := range block.List {
		switch stmt := stmt.(type) {
		case *ast.ExprStmt:
			if callExpr, ok := stmt.X.(*ast.CallExpr); ok {
				err := v.generateCallExpr(callExpr)
				if err != nil {
					fmt.Println("Error writing to file:", err)
				}
			}
		case *ast.DeclStmt:
			var variables []Variable
			if genDecl, ok := stmt.Decl.(*ast.GenDecl); ok && genDecl.Tok == token.VAR {
				for _, spec := range genDecl.Specs {
					if valueSpec, ok := spec.(*ast.ValueSpec); ok {
						// Iterate through all variables declared
						for _, ident := range valueSpec.Names {
							varType := "inferred"
							if valueSpec.Type != nil {
								varType = v.inspectType(valueSpec.Type)
							}
							variables = append(variables, Variable{
								Name: ident.Name,
								Type: varType,
							})
						}
					}
				}
			}
			for _, variable := range variables {
				cppType := variable.Type
				if val, ok := typesMap[variable.Type]; ok {
					cppType = val
				}
				err := v.emit(fmt.Sprintf("  %s %s;\n", cppType, variable.Name))
				if err != nil {
					fmt.Println("Error writing to file:", err)
				}
			}
		case *ast.AssignStmt:
			v.emitAssignment(stmt)
		case *ast.ReturnStmt:
			v.emitReturnStmt(stmt)
		case *ast.IfStmt:
			v.emitIfStmt(stmt, indent)

		default:
			fmt.Printf("<Other statement type>\n")
		}
	}
}

func (v *CppBackendVisitor) emitIfStmt(ifStmt *ast.IfStmt, indent int) {
	v.generateIndent(indent)
	v.emit("if")
	v.emitExpression(ifStmt.Cond)
	v.emit(" {\n")
	v.emitBlockStmt(ifStmt.Body, indent+1)
	v.generateIndent(indent)
	v.emit("}\n")
	if ifStmt.Else != nil {
		if elseIf, ok := ifStmt.Else.(*ast.IfStmt); ok {
			v.emitIfStmt(elseIf, indent+1) // Recursive call for else-if
		} else if elseBlock, ok := ifStmt.Else.(*ast.BlockStmt); ok {
			v.generateIndent(indent)
			v.emit("else {\n")
			v.emitBlockStmt(elseBlock, indent+1) // Dump else block
			v.generateIndent(indent)
			v.emit("}\n")
		}
	}
}

type Variable struct {
	Name string
	Type string
}

func (v *CppBackendVisitor) generateFuncDecl(node *ast.FuncDecl) ast.Visitor {
	if node.Type.Results != nil {
		resultArgIndex := 0
		if len(node.Type.Results.List) > 1 {
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
		if len(node.Type.Results.List) > 1 {
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
	v.emitBlockStmt(node.Body, 1)
	err = v.emit("}\n")
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return v
	}
	return v
}

func (v *CppBackendVisitor) buildTypesGraph() map[string][]string {
	typesGraph := make(map[string][]string)
	for _, node := range v.nodes {
		switch node := node.(type) {
		case *ast.TypeSpec:
			if st, ok := node.Type.(*ast.StructType); ok {
				structType := v.pkg.Name + "::" + node.Name.Name
				for _, field := range st.Fields.List {
					switch typ := field.Type.(type) {
					case *ast.Ident:
						if _, ok := primTypes[typ.Name]; !ok {
							fieldType := v.pkg.Name + "::" + typ.Name
							if fieldType != structType {
								typesGraph[fieldType] = append(typesGraph[fieldType], structType)
							}
						}
					case *ast.SelectorExpr: // External struct from another package
						if obj := v.pkg.TypesInfo.Uses[typ.Sel]; obj != nil {
							if named, ok := obj.Type().(*types.Named); ok {
								if _, ok := named.Underlying().(*types.Struct); ok {
									fieldType := named.Obj().Pkg().Name() + "::" + named.Obj().Name()
									if fieldType != structType {
										typesGraph[fieldType] = append(typesGraph[fieldType], structType)
									}
								}
							}
						}
					case *ast.ArrayType:
						switch elt := typ.Elt.(type) {
						case *ast.Ident:
							fieldType := v.pkg.Name + "::" + elt.Name
							if _, ok := primTypes[fieldType]; !ok {
								if fieldType != structType {
									typesGraph[fieldType] = append(typesGraph[fieldType], structType)
								}
							}
						case *ast.SelectorExpr: // Imported types
							if pkgIdent, ok := elt.X.(*ast.Ident); ok {
								fieldType := pkgIdent.Name + "::" + elt.Sel.Name
								if fieldType != structType {
									typesGraph[fieldType] = append(typesGraph[fieldType], structType)
								}
							}
						}
					}
				}
			}
		}
	}
	return typesGraph
}

// TopologicalSort performs a topological sort on the given graph.
// The input graph is a map where keys are nodes and values are slices of their dependencies.
func TopologicalSort(graph map[string][]string) ([]string, error) {
	// Track the state of each node: 0 = unvisited, 1 = visiting, 2 = visited
	visited := make(map[string]int)
	result := []string{}

	// Helper function for depth-first search (DFS)
	var visit func(string) error
	visit = func(node string) error {
		state := visited[node]

		// If the node is already visited, return
		if state == 2 {
			return nil
		}
		// If we find a node in "visiting" state, there is a cycle
		if state == 1 {
			return errors.New("cycle detected in the graph")
		}

		// Mark the node as visiting
		visited[node] = 1

		// Visit all the dependencies of the current node, if any
		if deps, exists := graph[node]; exists {
			for _, dep := range deps {
				if err := visit(dep); err != nil {
					return err // propagate the cycle detection error
				}
			}
		}

		// Mark the node as visited and add it to the result
		visited[node] = 2
		result = append(result, node)

		return nil
	}

	// Visit all nodes in the graph (including those without dependencies)
	for node := range graph {
		if visited[node] == 0 {
			if err := visit(node); err != nil {
				return nil, err
			}
		}
	}

	// Ensure we include nodes without outgoing edges
	// For example, in a graph {A -> B}, if C has no dependencies, it should also be in the result.
	for node := range visited {
		if visited[node] == 0 {
			if err := visit(node); err != nil {
				return nil, err
			}
		}
	}

	// Reverse the result because nodes are added in post-order
	reverse(result)

	return result, nil
}

// reverse reverses a slice of strings in place
func reverse(arr []string) {
	for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
		arr[i], arr[j] = arr[j], arr[i]
	}
}

func SliceToMap(slice []string) map[string]int {
	// Create a map to store the string and its index
	result := make(map[string]int)

	// Loop over the slice and fill the map
	for index, value := range slice {
		result[value] = index
	}

	return result
}

func (v *CppBackendVisitor) gen(precedence map[string]int) {
	structInfos := make([]GenStructInfo, 0)
	for i := 0; i < len(v.nodes); i++ {
		switch node := v.nodes[i].(type) {
		case *ast.TypeSpec:
			if st, ok := node.Type.(*ast.StructType); ok {
				structInfos = append(structInfos, GenStructInfo{
					Name:       node.Name.Name,
					Struct:     st, // We don't have type info for local structs yet
					IsExternal: false,
					Pkg:        v.pkg.Name,
				})
			}
		}
	}
	// Sort structs based on the precedence map
	sort.Slice(structInfos, func(i, j int) bool {
		// If the struct name is in the map, use its precedence value
		// Otherwise, treat it with the highest precedence (e.g., 0 or max int)
		precI := precedence[v.pkg.Name+"::"+structInfos[i].Name]
		precJ := precedence[v.pkg.Name+"::"+structInfos[j].Name]
		return precI < precJ
	})
	for i := 0; i < len(structInfos); i++ {
		err := v.emit(fmt.Sprintf("struct %s\n", structInfos[i].Name))
		if err != nil {
			fmt.Println("Error writing to file:", err)
		}
		err = v.emit("{\n")
		if err != nil {
			fmt.Println("Error writing to file:", err)
		}
		v.generateFields(structInfos[i].Struct)
		err = v.emit("};\n\n")
		if err != nil {
			fmt.Println("Error writing to file:", err)
		}
	}

	for _, node := range v.nodes {
		switch node := node.(type) {
		case *ast.TypeSpec:
			if _, ok := node.Type.(*ast.StructType); !ok {
				if arrayArg, ok := node.Type.(*ast.ArrayType); ok {
					v.generateArrayType(arrayArg, node.Name.Name, ArrayAlias)
				} else {
					err := v.emit(fmt.Sprintf("using %s = %s;\n\n", node.Name.Name, v.inspectType(node.Type)))
					if err != nil {
						fmt.Println("Error writing to file:", err)
					}
				}
			}
		}
	}

	for _, node := range v.nodes {
		switch node := node.(type) {
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

func (v *CppBackendVisitor) complementPrecedenceMap(sortedTypes map[string]int) {
	for _, node := range v.nodes {
		switch node := node.(type) {
		case *ast.TypeSpec:
			if _, ok := node.Type.(*ast.StructType); ok {
				if _, exists := sortedTypes[v.pkg.Name+"::"+node.Name.Name]; !exists {
					sortedTypes[v.pkg.Name+"::"+node.Name.Name] = len(sortedTypes)
				}
			}
		}
	}
}

func (v *CppBackend) PostVisit(visitor ast.Visitor, visited map[string]struct{}) {
	cppVisitor := visitor.(*CppBackendVisitor)
	typesGraph := cppVisitor.buildTypesGraph()
	for name, val := range typesGraph {
		fmt.Println("Type:", name, "Parent:", val)
	}
	typesTopoSorted, err := TopologicalSort(typesGraph)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println("Types Topological Sort:", typesTopoSorted)
	typesPrecedence := SliceToMap(typesTopoSorted)
	cppVisitor.complementPrecedenceMap(typesPrecedence)
	for name, _ := range typesPrecedence {
		if !strings.HasPrefix(name, cppVisitor.pkg.Name) {
			delete(typesPrecedence, name)
		}
	}
	fmt.Println("Types precedence:", typesPrecedence)
	cppVisitor.gen(typesPrecedence)
	if cppVisitor.pkg.Name == "main" {
		return
	}
	err = cppVisitor.emit(fmt.Sprintf("} // namespace %s\n\n", cppVisitor.pkg.Name))
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}

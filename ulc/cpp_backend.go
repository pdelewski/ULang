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

var namespaces = map[string]struct{}{}

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

func (v *CppBackendVisitor) emit(s string, indent int) error {
	_, err := v.pass.file.WriteString(strings.Repeat(" ", indent) + s)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return err
	}
	return nil
}

func (v *CppBackendVisitor) emitToFile(s string) error {
	_, err := v.pass.file.WriteString(s)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return err
	}
	return nil
}

func (v *CppBackendVisitor) emitAsString(s string, indent int) string {
	return strings.Repeat(" ", indent) + s
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
			err = v.emit(fmt.Sprintf("std::vector<%s> %s;\n", cppType, fieldName), 2)
		case ArrayArgument:
			if len(fieldName) == 0 {
				panic("expected field")
			}
			err = v.emit(fmt.Sprintf("std::vector<%s> %s", cppType, fieldName), 0)
		case ArrayReturn:
			err = v.emit(fmt.Sprintf("std::vector<%s>", cppType), 0)
		case ArrayAlias:
			if len(fieldName) == 0 {
				panic("expected field")
			}
			err = v.emit(fmt.Sprintf("using %s = std::vector<%s>;\n\n", fieldName, cppType), 0)
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
				err = v.emit(fmt.Sprintf("std::vector<%s::%s> %s;\n", pkgIdent.Name, cppType, fieldName), 2)
			case ArrayArgument:
				if len(fieldName) == 0 {
					panic("expected field")
				}
				err = v.emit(fmt.Sprintf("std::vector<%s::%s> %s", pkgIdent.Name, cppType, fieldName), 0)
			case ArrayReturn:
				err = v.emit(fmt.Sprintf("std::vector<%s::%s>", pkgIdent.Name, cppType), 0)
			case ArrayAlias:
				if len(fieldName) == 0 {
					panic("expected field")
				}
				err = v.emit(fmt.Sprintf("using %s = std::vector<%s::%s>\n\n", fieldName, pkgIdent.Name, cppType), 0)
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
		err = v.emit(cppType, 0)
	} else if pkg != "" {
		err = v.emit(fmt.Sprintf("%s::%s %s", pkg, cppType, fieldName), 0)
	} else {
		err = v.emit(fmt.Sprintf("%s %s", cppType, fieldName), 0)
	}
	if err != nil {
		fmt.Println("Error writing to file:", err)
	}
}

func (v *CppBackendVisitor) generateFields(st *ast.StructType, indent int) {
	for _, field := range st.Fields.List {
		for _, fieldName := range field.Names {
			switch typ := field.Type.(type) {
			case *ast.Ident:
				cppType := typ.Name
				if val, ok := typesMap[typ.Name]; ok {
					cppType = val
				}
				err := v.emit(fmt.Sprintf("%s %s;\n", cppType, fieldName.Name), indent)
				if err != nil {
					fmt.Println("Error writing to file:", err)
				}
			case *ast.SelectorExpr: // External struct from another package
				if obj := v.pkg.TypesInfo.Uses[typ.Sel]; obj != nil {
					if named, ok := obj.Type().(*types.Named); ok {
						if _, ok := named.Underlying().(*types.Struct); ok {
							v.emit("", 2)
							v.generatePrimType(named.Obj().Name(), named.Obj().Pkg().Name(), fieldName.Name)
							v.emit(";\n", 0)
						}
					}
				}
			case *ast.ArrayType:
				v.generateArrayType(typ, fieldName.Name, ArrayStructField)
			case *ast.FuncType:
				v.emitExpression(typ, 0)
				//v.emit(fmt.Sprintf("%s", v.inspectType(typ)), 0)
				v.emit(fmt.Sprintf("> %s;\n", fieldName.Name), 0)
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
		return "std::vector<" + v.inspectType(t.Elt) + ">"
	case *ast.FuncType:
		return "std::function<>"
	default:
		return fmt.Sprintf("%T", expr)
	}
	panic(fmt.Sprintf("unsupported expression type: %T", expr))
}

func resolveSelector(selExpr *ast.SelectorExpr) string {
	var result string
	switch x := selExpr.X.(type) {
	case *ast.Ident: // Base case: variable name
		result = x.Name
	case *ast.SelectorExpr: // Recursive case: nested selectors
		result = resolveSelector(x)
	default:
		panic(fmt.Sprintf("unsupported expression type: %T", selExpr))
	}
	scopeOperator := "."
	if _, found := namespaces[result]; found {
		scopeOperator = "::"
	}
	return fmt.Sprintf("%s%s%s", result, scopeOperator, selExpr.Sel.Name)
}

func (v *CppBackendVisitor) lowerToBuiltins(selector string) string {
	switch selector {
	case "fmt.Sprintf":
		return "string_format"
	case "fmt.Println":
		return "println"
	case "fmt.Printf":
		return "printf"
	case "fmt.Print":
		return "printf"
	}
	return selector
}

func (v *CppBackendVisitor) emitArgs(node *ast.CallExpr, indent int) string {
	var str string
	str = v.emitAsString("(", 0)
	v.emitToFile(str)
	for i, arg := range node.Args {
		if i > 0 {
			str = v.emitAsString(", ", 0)
			v.emitToFile(str)
		}
		str = v.emitExpression(arg, indent) // Function arguments
	}
	str = v.emitAsString(")", 0)
	v.emitToFile(str)
	return str
}

func (v *CppBackendVisitor) generateCallExpr(node *ast.CallExpr, indent int) error {
	if fun, ok := node.Fun.(*ast.Ident); ok {
		funcName := fun.Name
		if fun.Name == "len" {
			funcName = "std::size"
		}
		str := v.emitAsString(funcName, indent)
		v.emitToFile(str)
		v.emitArgs(node, indent)
	} else if sel, ok := node.Fun.(*ast.SelectorExpr); ok {
		str := v.emitAsString("", indent)
		v.emitToFile(str)
		v.emitExpression(sel, indent)
		v.emitArgs(node, indent)
	} else {
		fmt.Println("<complex call expression>")
	}
	return nil
}

func (v *CppBackendVisitor) emitExpression(expr ast.Expr, indent int) string {
	var str string
	switch e := expr.(type) {
	case *ast.BasicLit:
		if e.Kind == token.STRING {
			e.Value = strings.Replace(e.Value, "\"", "", -1)
			if e.Value[0] == '`' {
				e.Value = strings.Replace(e.Value, "`", "", -1)
				str = v.emitAsString(fmt.Sprintf("R\"(%s)\"", e.Value), 0)
				v.emitToFile(str)
			} else {
				str = v.emitAsString(fmt.Sprintf("\"%s\"", e.Value), 0)
				v.emitToFile(str)
			}
		} else {
			str = v.emitAsString(e.Value, 0)
			v.emitToFile(str)
		}
	case *ast.Ident:
		name := e.Name
		if name == "nil" {
			str = v.emitAsString("{}", 0)
			v.emitToFile(str)
		} else {
			if n, ok := typesMap[e.Name]; ok {
				str = v.emitAsString(n, 0)
				v.emitToFile(str)
			} else {
				str = v.emitAsString(e.Name, 0)
				v.emitToFile(str)
			}
		}
	case *ast.BinaryExpr:
		v.emitExpression(e.X, indent) // Left operand
		str = v.emitAsString(e.Op.String()+" ", 1)
		v.emitToFile(str)
		v.emitExpression(e.Y, indent) // Right operand
	case *ast.CallExpr:
		v.generateCallExpr(e, indent)
	case *ast.ParenExpr:
		str = v.emitAsString("(", 0)
		v.emitToFile(str)
		v.emitExpression(e.X, indent) // Dump inner expression
		str = v.emitAsString(")", 0)
		v.emitToFile(str)
	case *ast.CompositeLit:
		isArray := false
		isLeftBrace := false
		switch t := e.Type.(type) {
		case *ast.Ident:
			str = v.emitAsString(fmt.Sprintf("%s{", t.Name), 0)
			v.emitToFile(str)
			isLeftBrace = true
		case *ast.SelectorExpr:
			if pkgIdent, ok := t.X.(*ast.Ident); ok {
				str = v.emitAsString(fmt.Sprintf("%s::%s{", pkgIdent.Name, t.Sel.Name), 0)
				v.emitToFile(str)
				isLeftBrace = true
			} else {
				panic(fmt.Sprintf("unsupported expression type: %T", expr))
			}
		case *ast.ArrayType:
			str = v.emitAsString("std::vector<", 0)
			v.emitToFile(str)
			v.emitExpression(t.Elt, 0)
			str = v.emitAsString(">{", 0)
			v.emitToFile(str)
			isArray = true
		}
		for i, elt := range e.Elts {
			if i > 0 {
				str = v.emitAsString(", ", 0)
				v.emitToFile(str)
			}
			v.emitExpression(elt, indent)
		}
		if isArray {
			str = v.emitAsString("}", 0)
			v.emitToFile(str)
		}
		if isLeftBrace {
			str = v.emitAsString("}", 0)
			v.emitToFile(str)
		}
	case *ast.SelectorExpr:
		selector := resolveSelector(e)
		selector = v.lowerToBuiltins(selector)
		v.emit(selector, 0)
	case *ast.IndexExpr:
		v.emitExpression(e.X, indent)
		str = v.emitAsString("[", 0)
		v.emitToFile(str)
		v.emitExpression(e.Index, indent)
		str = v.emitAsString("]", 0)
		v.emitToFile(str)
	case *ast.UnaryExpr:
		str = v.emitAsString("(", 0)
		v.emitToFile(str)
		str = v.emitAsString(e.Op.String(), 0)
		v.emitToFile(str)
		v.emitExpression(e.X, 0)
		str = v.emitAsString(")", 0)
		v.emitToFile(str)
	case *ast.SliceExpr:
		str = v.emitAsString("std::vector<std::remove_reference<decltype(", indent)
		v.emitToFile(str)
		v.emitExpression(e.X, 0)
		str = v.emitAsString("[0]", 0)
		v.emitToFile(str)
		str = v.emitAsString(")>::type>(", 0)
		v.emitToFile(str)
		// Check and print Low, High, and Max
		v.emitExpression(e.X, indent)
		str = v.emitAsString(".begin() ", 0)
		v.emitToFile(str)
		if e.Low != nil {
			str = v.emitAsString("+ ", 0)
			v.emitToFile(str)
			v.emitExpression(e.Low, indent)
		} else {
			log.Println("Low index: <nil>")
		}
		str = v.emitAsString(", ", 0)
		v.emitToFile(str)
		v.emitExpression(e.X, indent)
		if e.High != nil {
			str = v.emitAsString(".begin() ", 0)
			v.emitToFile(str)
			str = v.emitAsString("+ ", 0)
			v.emitToFile(str)
			v.emitExpression(e.High, indent)
		} else {
			str = v.emitAsString(".end() ", 0)
			v.emitToFile(str)
		}
		if e.Slice3 && e.Max != nil {
			v.emitExpression(e.Max, indent)
		} else if e.Slice3 {
			log.Println("Max index: <nil>")
		}
		str = v.emitAsString(")", 0)
		v.emitToFile(str)
	case *ast.FuncType:
		str = v.emitAsString("std::function<", indent)
		v.emitToFile(str)
		for i, result := range e.Results.List {
			if i > 0 {
				str = v.emitAsString(", ", 0)
				v.emitToFile(str)
			}
			v.emitExpression(result.Type, indent)
		}
		str = v.emitAsString("(", 0)
		v.emitToFile(str)
		for i, param := range e.Params.List {
			if i > 0 {
				str = v.emitAsString(", ", 0)
				v.emitToFile(str)
			}
			v.emitExpression(param.Type, indent)
		}
		str = v.emitAsString(")", 0)
		v.emitToFile(str)
	case *ast.KeyValueExpr:
		str = v.emitAsString(".", 0)
		v.emitToFile(str)
		v.emitExpression(e.Key, indent)
		str = v.emitAsString("= ", 0)
		v.emitToFile(str)
		v.emitExpression(e.Value, indent)
		str = v.emitAsString("\n", 0)
		v.emitToFile(str)
	case *ast.FuncLit:
		str = v.emitAsString("[&](", indent)
		v.emitToFile(str)
		for i, param := range e.Type.Params.List {
			if i > 0 {
				str = v.emitAsString(", ", 0)
				v.emitToFile(str)
			}
			v.emitExpression(param.Type, indent)
			str = v.emitAsString(" ", 0)
			v.emitToFile(str)
			str = v.emitAsString(param.Names[0].Name, indent)
			v.emitToFile(str)
		}
		str = v.emitAsString(")", 0)
		v.emitToFile(str)
		str = v.emitAsString("->", 0)
		v.emitToFile(str)
		for i, result := range e.Type.Results.List {
			if i > 0 {
				str = v.emitAsString(", ", 0)
				v.emitToFile(str)
			}
			v.emitExpression(result.Type, indent)
		}
		str = v.emitAsString("{\n", 0)
		v.emitToFile(str)
		v.emitBlockStmt(e.Body, indent+2)
		str = v.emitAsString("}", 0)
		v.emitToFile(str)
	case *ast.TypeAssertExpr:
		str = v.emitAsString("std::any_cast<", indent)
		v.emitToFile(str)
		v.emitExpression(e.Type, indent)
		str = v.emitAsString(">(std::any(", 0)
		v.emitToFile(str)
		v.emitExpression(e.X, indent)
		str = v.emitAsString("))", 0)
		v.emitToFile(str)
	default:
		panic(fmt.Sprintf("unsupported expression type: %T", e))
	}
	return str
}

func (v *CppBackendVisitor) emitAssignment(assignStmt *ast.AssignStmt, indent int) {
	assignmentToken := assignStmt.Tok.String()
	if assignmentToken == ":=" && len(assignStmt.Lhs) == 1 {
		v.emit("auto ", indent)
	} else if assignmentToken == ":=" && len(assignStmt.Lhs) > 1 {
		v.emit("auto [", indent)
	} else if assignmentToken == "=" && len(assignStmt.Lhs) > 1 {
		v.emit("std::tie(", indent)
	}
	if assignmentToken != "+=" {
		assignmentToken = "="
	}
	first := true
	for _, lhs := range assignStmt.Lhs {
		if !first {
			v.emit(", ", indent)
		}
		first = false
		if ident, ok := lhs.(*ast.Ident); ok {
			v.emit(ident.Name, indent)
		} else {
			v.emitExpression(lhs, indent)
		}
	}

	if assignStmt.Tok.String() == ":=" && len(assignStmt.Lhs) > 1 {
		v.emit("]", indent)
	} else if assignStmt.Tok.String() == "=" && len(assignStmt.Lhs) > 1 {
		v.emit(")", indent)
	}

	v.emit(assignmentToken+" ", indent+1)

	for _, rhs := range assignStmt.Rhs {
		v.emitExpression(rhs, indent)
	}
}

func (v *CppBackendVisitor) emitReturnStmt(retStmt *ast.ReturnStmt, indent int) {
	v.emit("return ", indent)
	if len(retStmt.Results) > 1 {
		v.emit("std::make_tuple(", 0)
	}
	first := true
	for _, result := range retStmt.Results {
		if !first {
			v.emit(", ", 0)
		}
		first = false
		v.emitExpression(result, indent)
	}
	if len(retStmt.Results) > 1 {
		v.emit(")", 0)
	}
	v.emit(";\n", 0)
}

func (v *CppBackendVisitor) emitStmt(stmt ast.Stmt, indent int) {
	switch stmt := stmt.(type) {
	case *ast.ExprStmt:
		if callExpr, ok := stmt.X.(*ast.CallExpr); ok {
			err := v.generateCallExpr(callExpr, indent)
			str := v.emitAsString(";\n", 0)
			err = v.emitToFile(str)
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
			err := v.emit(fmt.Sprintf("%s %s;\n", cppType, variable.Name), indent)
			if err != nil {
				fmt.Println("Error writing to file:", err)
			}
		}
	case *ast.AssignStmt:
		str := v.emitAsString("", indent)
		v.emitToFile(str)
		v.emitAssignment(stmt, 0)
		str = v.emitAsString(";\n", 0)
		v.emitToFile(str)
	case *ast.ReturnStmt:
		v.emitReturnStmt(stmt, indent)
	case *ast.IfStmt:
		v.emitIfStmt(stmt, indent)
	case *ast.ForStmt:
		str := v.emitAsString("for (", indent)
		v.emitToFile(str)
		if stmt.Init != nil {
			if assignStmt, ok := stmt.Init.(*ast.AssignStmt); ok {
				v.emitAssignment(assignStmt, 0)
				str := v.emitAsString(";", 0)
				v.emitToFile(str)
			} else if exprStmt, ok := stmt.Init.(*ast.ExprStmt); ok {
				v.emitExpression(exprStmt.X, indent)
				str := v.emitAsString("; ", 0)
				v.emitToFile(str)
			} else {
				str := v.emitAsString("; ", 0)
				v.emitToFile(str)
			}
		} else {
			str := v.emitAsString("; ", 0)
			v.emitToFile(str)
		}
		if stmt.Cond != nil {
			v.emitExpression(stmt.Cond, indent)
			str := v.emitAsString("; ", 0)
			v.emitToFile(str)
		} else {
			str := v.emitAsString("; ", 0)
			v.emitToFile(str)
		}
		if stmt.Post != nil {
			if postExpr, ok := stmt.Post.(*ast.ExprStmt); ok {
				v.emitExpression(postExpr.X, indent)
			} else if incDeclStmt, ok := stmt.Post.(*ast.IncDecStmt); ok {
				v.emitExpression(incDeclStmt.X, indent)
				str := v.emitAsString(incDeclStmt.Tok.String(), 0)
				v.emitToFile(str)
			}
		}
		str = v.emitAsString(") {\n", 0)
		v.emitToFile(str)
		v.emitBlockStmt(stmt.Body, indent+2)
		str = v.emitAsString("}\n", indent)
		v.emitToFile(str)
	case *ast.RangeStmt:
		str := v.emitAsString("for (auto ", indent)
		v.emitToFile(str)
		if stmt.Value != nil {
			v.emitExpression(stmt.Value, indent)
		}
		str = v.emitAsString(" : ", 0)
		v.emitToFile(str)
		v.emitExpression(stmt.X, indent)
		str = v.emitAsString(") {\n", 0)
		v.emitToFile(str)
		v.emitBlockStmt(stmt.Body, indent+2)
		str = v.emitAsString("}\n", indent)
		v.emitToFile(str)
	case *ast.SwitchStmt:
		str := v.emitAsString("switch (", indent)
		v.emitToFile(str)
		v.emitExpression(stmt.Tag, indent)
		str = v.emitAsString(") {\n", 0)
		v.emitToFile(str)
		for _, stmt := range stmt.Body.List {
			if caseClause, ok := stmt.(*ast.CaseClause); ok {
				for _, expr := range caseClause.List {
					str := v.emitAsString("case ", indent+2)
					v.emitToFile(str)
					v.emitExpression(expr, indent+2)
					str = v.emitAsString(":\n", 0)
					v.emitToFile(str)
				}
				if len(caseClause.List) == 0 {
					str := v.emitAsString("default:\n", indent+2)
					v.emitToFile(str)
				}
				for _, innerStmt := range caseClause.Body {
					v.emitStmt(innerStmt, indent+4)
				}
				str := v.emitAsString("break;\n", indent+4)
				v.emitToFile(str)
			}
		}
		str = v.emitAsString("}\n", indent)
		v.emitToFile(str)
	case *ast.BranchStmt:
		switch stmt.Tok {
		case token.BREAK:
			str := v.emitAsString("break;\n", indent)
			v.emitToFile(str)
		case token.CONTINUE:
			str := v.emitAsString("continue;\n", indent)
			v.emitToFile(str)
		}
	case *ast.IncDecStmt:
		str := v.emitAsString("", indent)
		v.emitToFile(str)
		v.emitExpression(stmt.X, indent)
		str = v.emitAsString(stmt.Tok.String(), 0)
		str += v.emitAsString(";\n", 0)
		v.emitToFile(str)
	default:
		fmt.Printf("<Other statement type>\n")

	}
}

func (v *CppBackendVisitor) emitBlockStmt(block *ast.BlockStmt, indent int) {
	for _, stmt := range block.List {
		v.emitStmt(stmt, indent)
	}
}

func (v *CppBackendVisitor) emitIfStmt(ifStmt *ast.IfStmt, indent int) {
	str := v.emitAsString("if", indent)
	str += v.emitAsString(" (", 0)
	v.emitToFile(str)
	v.emitExpression(ifStmt.Cond, 1)
	str = v.emitAsString(") ", 0)
	str += v.emitAsString(" {\n", 0)
	v.emitToFile(str)
	v.emitBlockStmt(ifStmt.Body, indent+2)
	str = v.emitAsString("}\n", indent)
	v.emitToFile(str)
	if ifStmt.Else != nil {
		if elseIf, ok := ifStmt.Else.(*ast.IfStmt); ok {
			str = v.emitAsString("else", indent)
			v.emitToFile(str)
			v.emitIfStmt(elseIf, indent) // Recursive call for else-if
		} else if elseBlock, ok := ifStmt.Else.(*ast.BlockStmt); ok {
			str = v.emitAsString("else {\n", indent)
			v.emitToFile(str)
			v.emitBlockStmt(elseBlock, indent+2) // Dump else block
			str = v.emitAsString("}\n", indent)
			v.emitToFile(str)
		}
	}
}

type Variable struct {
	Name string
	Type string
}

func (v *CppBackendVisitor) generateFuncDeclSignature(node *ast.FuncDecl) ast.Visitor {
	if node.Type.Results != nil {
		resultArgIndex := 0
		if len(node.Type.Results.List) > 1 {
			str := v.emitAsString("std::tuple<", 0)
			err := v.emitToFile(str)
			if err != nil {
				fmt.Println("Error writing to file:", err)
				return v
			}
		}
		for _, result := range node.Type.Results.List {
			if resultArgIndex > 0 {
				str := v.emitAsString(",", 0)
				err := v.emitToFile(str)
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
			str := v.emitAsString(">", 0)
			err := v.emitToFile(str)
			if err != nil {
				fmt.Println("Error writing to file:", err)
				return v
			}
		}
	} else if node.Name.Name == "main" {
		str := v.emitAsString("int", 0)
		err := v.emitToFile(str)
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return v
		}
	} else {
		str := v.emitAsString("void", 0)
		err := v.emitToFile(str)
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return v
		}
	}
	str := v.emitAsString("", 1)
	err := v.emitToFile(str)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return v
	}
	str = v.emitAsString(node.Name.Name+"(", 0)
	err = v.emitToFile(str)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return v
	}
	argIndex := 0
	for _, arg := range node.Type.Params.List {
		if argIndex > 0 {
			str = v.emitAsString(", ", 0)
			err = v.emitToFile(str)
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
	str = v.emitAsString(")", 0)
	err = v.emitToFile(str)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return v
	}
	return v
}

func (v *CppBackendVisitor) generateFuncDecl(node *ast.FuncDecl) ast.Visitor {
	v.generateFuncDeclSignature(node)
	str := v.emitAsString("\n", 0)
	str += v.emitAsString("{\n", 0)
	err := v.emitToFile(str)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return v
	}
	v.emitBlockStmt(node.Body, 2)
	str = v.emitAsString("}\n", 0)
	str += v.emitAsString("\n", 0)
	err = v.emitToFile(str)
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
		str := v.emitAsString(fmt.Sprintf("struct %s\n", structInfos[i].Name), 0)
		err := v.emitToFile(str)
		if err != nil {
			fmt.Println("Error writing to file:", err)
		}
		str = v.emitAsString("{\n", 0)
		err = v.emitToFile(str)
		if err != nil {
			fmt.Println("Error writing to file:", err)
		}
		v.generateFields(structInfos[i].Struct, 2)
		str = v.emitAsString("};\n\n", 0)
		err = v.emitToFile(str)
		if err != nil {
			fmt.Println("Error writing to file:", err)
		}
	}
	for _, node := range v.nodes {
		if genDecl, ok := node.(*ast.GenDecl); ok && genDecl.Tok == token.CONST {
			for _, spec := range genDecl.Specs {
				valueSpec := spec.(*ast.ValueSpec)
				for i, name := range valueSpec.Names {
					str := v.emitAsString(fmt.Sprintf("constexpr auto %s = ", name.Name), 0)
					v.emitToFile(str)
					if valueSpec.Values != nil {
						if i < len(valueSpec.Values) {
							value := valueSpec.Values[i]
							switch val := value.(type) {
							case *ast.BasicLit:
								str = v.emitAsString(val.Value, 0)
								str += v.emitAsString(";\n", 0)
								v.emitToFile(str)
							default:
								log.Printf("Unhandled value type: %T", v)
							}
						}
					}
				}
			}
			str := v.emitAsString("\n", 0)
			v.emitToFile(str)
		}

		switch node := node.(type) {
		case *ast.TypeSpec:
			if _, ok := node.Type.(*ast.StructType); !ok {
				if arrayArg, ok := node.Type.(*ast.ArrayType); ok {
					v.generateArrayType(arrayArg, node.Name.Name, ArrayAlias)
				} else {
					str := v.emitAsString(fmt.Sprintf("using %s = %s;\n\n", node.Name.Name, v.inspectType(node.Type)), 0)
					err := v.emitToFile(str)
					if err != nil {
						fmt.Println("Error writing to file:", err)
					}
				}
			}
		}
	}

	// Generate forward function declarations
	str := v.emitAsString("// Forward declarations\n", 0)
	v.emitToFile(str)
	for _, node := range v.nodes {
		switch node := node.(type) {
		case *ast.FuncDecl:
			v.generateFuncDeclSignature(node)
			str = v.emitAsString(";\n", 0)
			v.emitToFile(str)
		}
	}
	str = v.emitAsString("\n", 0)
	v.emitToFile(str)
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
	namespaces = make(map[string]struct{})
	v.outputFile = "./output.cpp"
	var err error
	v.file, err = os.Create(v.outputFile)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	_, err = v.file.WriteString("#include <vector>\n" +
		"#include <string>\n" +
		"#include <tuple>\n" +
		"#include <any>\n" +
		"#include <cstdint>\n" +
		"#include <functional>\n" +
		"#include \"../builtins/builtins.h\"\n\n")
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
	namespaces[cppVisitor.pkg.Name] = struct{}{}
	err := cppVisitor.emit(fmt.Sprintf("namespace %s\n", cppVisitor.pkg.Name), 0)
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
	err = cppVisitor.emit(fmt.Sprintf("} // namespace %s\n\n", cppVisitor.pkg.Name), 0)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}

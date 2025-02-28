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
	emitter    Emitter
}

type CppBackendVisitor struct {
	pkg     *packages.Package
	pass    *CppBackend
	nodes   []ast.Node
	emitter Emitter
}

func (v *CppBackend) Name() string {
	return "CppGen"
}

func (v *CppBackend) Visitors(pkg *packages.Package) []ast.Visitor {
	v.visitor = &CppBackendVisitor{pkg: pkg, emitter: v.emitter}
	v.visitor.pass = v
	return []ast.Visitor{v.visitor}
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

func (v *CppBackendVisitor) generateFields(st *ast.StructType, indent int) {
	for _, field := range st.Fields.List {
		for _, fieldName := range field.Names {
			v.traverseExpression(field.Type, indent)
			v.emitToFile(" ")
			v.traverseExpression(fieldName, 0)
			v.emitToFile(";\n")
		}
	}
}

func (v *CppBackendVisitor) emitArgs(node *ast.CallExpr, indent int) {
	v.emitter.PreVisitCallExprArgs(node.Args, indent)
	for i, arg := range node.Args {
		v.emitter.PreVisitCallExprArg(arg, i, indent)
		v.traverseExpression(arg, 0) // Function arguments
		v.emitter.PostVisitCallExprArg(arg, i, indent)
	}
	v.emitter.PostVisitCallExprArgs(node.Args, indent)
}

func (v *CppBackendVisitor) traverseExpression(expr ast.Expr, indent int) string {
	var str string
	switch e := expr.(type) {
	case *ast.BasicLit:
		v.emitter.PreVisitBasicLit(e, indent)
		v.emitter.PostVisitBasicLit(e, indent)
	case *ast.Ident:
		v.emitter.PreVisitIdent(e, indent)
		v.emitter.PostVisitIdent(e, indent)
	case *ast.BinaryExpr:
		v.emitter.PreVisitBinaryExpr(e, indent)
		v.emitter.PreVisitBinaryExprLeft(e.X, indent)
		v.traverseExpression(e.X, indent) // Left operand
		v.emitter.PostVisitBinaryExprLeft(e.X, indent)
		v.emitter.PreVisitBinaryExprOperator(e.Op, indent)
		v.emitter.PostVisitBinaryExprOperator(e.Op, indent)
		v.emitter.PreVisitBinaryExprRight(e.Y, indent)
		v.traverseExpression(e.Y, indent) // Right operand
		v.emitter.PostVisitBinaryExprRight(e.Y, indent)
		v.emitter.PostVisitBinaryExpr(e, indent)
	case *ast.CallExpr:
		v.emitter.PreVisitCallExpr(e, indent)
		v.emitter.PreVisitCallExprFun(e.Fun, indent)
		v.traverseExpression(e.Fun, indent)
		v.emitter.PostVisitCallExprFun(e.Fun, indent)
		v.emitArgs(e, indent)
		v.emitter.PostVisitCallExpr(e, indent)
	case *ast.ParenExpr:
		v.emitter.PreVisitParenExpr(e, indent)
		v.traverseExpression(e.X, indent) // Dump inner expression
		v.emitter.PostVisitParenExpr(e, indent)
	case *ast.CompositeLit:
		v.emitter.PreVisitCompositeLit(e, indent)
		v.emitter.PreVisitCompositeLitType(e.Type, indent)
		v.traverseExpression(e.Type, indent)
		v.emitter.PostVisitCompositeLitType(e.Type, indent)
		v.emitter.PreVisitCompositeLitElts(e.Elts, indent)
		for i, elt := range e.Elts {
			v.emitter.PreVisitCompositeLitElt(elt, i, indent)
			v.traverseExpression(elt, 0) // Function arguments
			v.emitter.PostVisitCompositeLitElt(elt, i, indent)
		}
		v.emitter.PostVisitCompositeLitElts(e.Elts, indent)
		v.emitter.PostVisitCompositeLit(e, indent)
	case *ast.ArrayType:
		v.emitter.PreVisitArrayType(*e, indent)
		v.traverseExpression(e.Elt, 0)
		v.emitter.PostVisitArrayType(*e, indent)
	case *ast.SelectorExpr:
		v.emitter.PreVisitSelectorExpr(e, indent)
		v.emitter.PreVisitSelectorExprX(e.X, indent)
		v.traverseExpression(e.X, indent)
		v.emitter.PostVisitSelectorExprX(e.X, indent)
		oldIndent := indent
		v.traverseExpression(e.Sel, 0)
		indent = oldIndent
		v.emitter.PostVisitSelectorExpr(e, indent)
	case *ast.IndexExpr:
		v.emitter.PreVisitIndexExpr(e, indent)
		v.emitter.PreVisitIndexExprX(e, indent)
		v.traverseExpression(e.X, indent)
		v.emitter.PostVisitIndexExprX(e, indent)
		v.emitter.PreVisitIndexExprIndex(e, indent)
		v.traverseExpression(e.Index, indent)
		v.emitter.PostVisitIndexExprIndex(e, indent)
		v.emitter.PostVisitIndexExpr(e, indent)
	case *ast.UnaryExpr:
		v.emitter.PreVisitUnaryExpr(e, indent)
		v.traverseExpression(e.X, 0)
		v.emitter.PostVisitUnaryExpr(e, indent)
	case *ast.SliceExpr:
		v.emitter.PreVisitSliceExpr(e, indent)
		v.emitter.PreVisitSliceExprX(e.X, indent)
		v.traverseExpression(e.X, 0)
		v.emitter.PostVisitSliceExprX(e.X, indent)
		// Check and print Low, High, and Max
		v.emitter.PreVisitSliceExprXBegin(e.X, indent)
		v.traverseExpression(e.X, indent)
		v.emitter.PostVisitSliceExprXBegin(e, indent)
		v.emitter.PreVisitSliceExprLow(e.Low, indent)
		if e.Low != nil {
			v.traverseExpression(e.Low, indent)
		}
		v.emitter.PostVisitSliceExprLow(e.Low, indent)
		v.emitter.PreVisitSliceExprXEnd(e, indent)
		v.traverseExpression(e.X, indent)
		v.emitter.PostVisitSliceExprXEnd(e, indent)
		v.emitter.PreVisitSliceExprHigh(e.High, indent)
		if e.High != nil {
			v.traverseExpression(e.High, indent)
		}
		v.emitter.PostVisitSliceExprHigh(e.High, indent)
		if e.Slice3 && e.Max != nil {
			v.traverseExpression(e.Max, indent)
		} else if e.Slice3 {
			log.Println("Max index: <nil>")
		}
		v.emitter.PostVisitSliceExpr(e, indent)
	case *ast.FuncType:
		v.emitter.PreVisitFuncType(e, indent)
		v.emitter.PreVisitFuncTypeResults(e.Results, indent)
		if e.Results != nil {
			for i, result := range e.Results.List {
				v.emitter.PreVisitFuncTypeResult(result, i, indent)
				v.traverseExpression(result.Type, 0)
				v.emitter.PostVisitFuncTypeResult(result, i, indent)
			}
		}
		v.emitter.PostVisitFuncTypeResults(e.Results, indent)
		v.emitter.PreVisitFuncTypeParams(e.Params, indent)
		for i, param := range e.Params.List {
			v.emitter.PreVisitFuncTypeParam(param, i, indent)
			v.traverseExpression(param.Type, 0)
			v.emitter.PostVisitFuncTypeParam(param, i, indent)
		}
		v.emitter.PostVisitFuncTypeParams(e.Params, indent)
		v.emitter.PostVisitFuncType(e, indent)
	case *ast.KeyValueExpr:
		str = v.emitAsString(".", 0)
		v.emitToFile(str)
		v.traverseExpression(e.Key, indent)
		str = v.emitAsString("= ", 0)
		v.emitToFile(str)
		v.traverseExpression(e.Value, indent)
		str = v.emitAsString("\n", 0)
		v.emitToFile(str)
	case *ast.FuncLit:
		str = v.emitAsString("[&](", indent)
		for i, param := range e.Type.Params.List {
			if i > 0 {
				str += v.emitAsString(", ", 0)
			}
			v.emitToFile(str)
			v.traverseExpression(param.Type, indent)
			str = v.emitAsString(" ", 0)
			str += v.emitAsString(param.Names[0].Name, indent)
		}
		str += v.emitAsString(")", 0)
		str += v.emitAsString("->", 0)
		v.emitToFile(str)
		if e.Type.Results != nil {
			for i, result := range e.Type.Results.List {
				if i > 0 {
					str = v.emitAsString(", ", 0)
					v.emitToFile(str)
				}
				v.traverseExpression(result.Type, indent)
			}
		} else {
			str = v.emitAsString("void", 0)
			v.emitToFile(str)
		}
		str = v.emitAsString("{\n", 0)
		v.emitToFile(str)
		v.traverseBlockStmt(e.Body, indent+2)
		str = v.emitAsString("}", 0)
		v.emitToFile(str)
	case *ast.TypeAssertExpr:
		str = v.emitAsString("std::any_cast<", indent)
		v.emitToFile(str)
		v.traverseExpression(e.Type, indent)
		str = v.emitAsString(">(std::any(", 0)
		v.emitToFile(str)
		v.traverseExpression(e.X, indent)
		str = v.emitAsString("))", 0)
		v.emitToFile(str)
	case *ast.StarExpr:
		str = v.emitAsString("*", 0)
		v.emitToFile(str)
		v.traverseExpression(e.X, indent)
	case *ast.InterfaceType:
		str = v.emitAsString("std::any", indent)
		v.emitToFile(str)
	default:
		panic(fmt.Sprintf("unsupported expression type: %T", e))
	}
	return str
}

func (v *CppBackendVisitor) traverseAssignment(assignStmt *ast.AssignStmt, indent int) {
	assignmentToken := assignStmt.Tok.String()
	if assignmentToken == ":=" && len(assignStmt.Lhs) == 1 {
		str := v.emitAsString("auto ", indent)
		v.emitToFile(str)
	} else if assignmentToken == ":=" && len(assignStmt.Lhs) > 1 {
		str := v.emitAsString("auto [", indent)
		v.emitToFile(str)
	} else if assignmentToken == "=" && len(assignStmt.Lhs) > 1 {
		str := v.emitAsString("std::tie(", indent)
		v.emitToFile(str)
	}
	if assignmentToken != "+=" {
		assignmentToken = "="
	}
	first := true
	for _, lhs := range assignStmt.Lhs {
		if !first {
			str := v.emitAsString(", ", indent)
			v.emitToFile(str)
		}
		first = false
		if ident, ok := lhs.(*ast.Ident); ok {
			str := v.emitAsString(ident.Name, indent)
			v.emitToFile(str)
		} else {
			v.traverseExpression(lhs, indent)
		}
	}

	if assignStmt.Tok.String() == ":=" && len(assignStmt.Lhs) > 1 {
		str := v.emitAsString("]", indent)
		v.emitToFile(str)
	} else if assignStmt.Tok.String() == "=" && len(assignStmt.Lhs) > 1 {
		str := v.emitAsString(")", indent)
		v.emitToFile(str)
	}

	str := v.emitAsString(assignmentToken+" ", indent+1)
	v.emitToFile(str)
	for _, rhs := range assignStmt.Rhs {
		v.traverseExpression(rhs, indent)
	}
}

func (v *CppBackendVisitor) traverseReturnStmt(retStmt *ast.ReturnStmt, indent int) {
	str := v.emitAsString("return ", indent)
	v.emitToFile(str)
	if len(retStmt.Results) > 1 {
		str := v.emitAsString("std::make_tuple(", 0)
		v.emitToFile(str)
	}
	first := true
	for _, result := range retStmt.Results {
		if !first {
			str := v.emitAsString(", ", 0)
			v.emitToFile(str)
		}
		first = false
		v.traverseExpression(result, 0)
	}
	if len(retStmt.Results) > 1 {
		str := v.emitAsString(")", 0)
		v.emitToFile(str)
	}
	str = v.emitAsString(";\n", 0)
	v.emitToFile(str)
}

func (v *CppBackendVisitor) traverseStmt(stmt ast.Stmt, indent int) {
	switch stmt := stmt.(type) {
	case *ast.ExprStmt:
		v.traverseExpression(stmt.X, indent)
		str := v.emitAsString(";\n", 0)
		err := v.emitToFile(str)
		if err != nil {
			fmt.Println("Error writing to file:", err)
		}
	case *ast.DeclStmt:
		if genDecl, ok := stmt.Decl.(*ast.GenDecl); ok && genDecl.Tok == token.VAR {
			for _, spec := range genDecl.Specs {
				if valueSpec, ok := spec.(*ast.ValueSpec); ok {
					// Iterate through all variables declared
					for _, ident := range valueSpec.Names {
						v.traverseExpression(valueSpec.Type, indent)
						str := v.emitAsString(" ", 0)
						v.emitToFile(str)
						v.traverseExpression(ident, 0)
						v.emitToFile(";\n")
					}
				}
			}
		}
	case *ast.AssignStmt:
		str := v.emitAsString("", indent)
		v.emitToFile(str)
		v.traverseAssignment(stmt, 0)
		str = v.emitAsString(";\n", 0)
		v.emitToFile(str)
	case *ast.ReturnStmt:
		v.traverseReturnStmt(stmt, indent)
	case *ast.IfStmt:
		v.traverseIfStmt(stmt, indent, false)
	case *ast.ForStmt:
		str := v.emitAsString("for (", indent)
		v.emitToFile(str)
		if stmt.Init != nil {
			if assignStmt, ok := stmt.Init.(*ast.AssignStmt); ok {
				v.traverseAssignment(assignStmt, 0)
				str := v.emitAsString(";", 0)
				v.emitToFile(str)
			} else if exprStmt, ok := stmt.Init.(*ast.ExprStmt); ok {
				v.traverseExpression(exprStmt.X, 0)
				str := v.emitAsString(";", 0)
				v.emitToFile(str)
			} else {
				str := v.emitAsString(";", 0)
				v.emitToFile(str)
			}
		} else {
			str := v.emitAsString(";", 0)
			v.emitToFile(str)
		}
		if stmt.Cond != nil {
			v.traverseExpression(stmt.Cond, 0)
			str := v.emitAsString(";", 0)
			v.emitToFile(str)
		} else {
			str := v.emitAsString(";", 0)
			v.emitToFile(str)
		}
		if stmt.Post != nil {
			if postExpr, ok := stmt.Post.(*ast.ExprStmt); ok {
				v.traverseExpression(postExpr.X, 0)
			} else if incDeclStmt, ok := stmt.Post.(*ast.IncDecStmt); ok {
				v.traverseExpression(incDeclStmt.X, 0)
				str := v.emitAsString(incDeclStmt.Tok.String(), 0)
				v.emitToFile(str)
			}
		}
		str = v.emitAsString(") {\n", 0)
		v.emitToFile(str)
		v.traverseBlockStmt(stmt.Body, indent+2)
		str = v.emitAsString("}\n", indent)
		v.emitToFile(str)
	case *ast.RangeStmt:
		str := v.emitAsString("for (auto ", indent)
		v.emitToFile(str)
		if stmt.Value != nil {
			v.traverseExpression(stmt.Value, 0)
		}
		str = v.emitAsString(" : ", 0)
		v.emitToFile(str)
		v.traverseExpression(stmt.X, 0)
		str = v.emitAsString(") {\n", 0)
		v.emitToFile(str)
		v.traverseBlockStmt(stmt.Body, indent+2)
		str = v.emitAsString("}\n", indent)
		v.emitToFile(str)
	case *ast.SwitchStmt:
		str := v.emitAsString("switch (", indent)
		v.emitToFile(str)
		v.traverseExpression(stmt.Tag, 0)
		str = v.emitAsString(") {\n", 0)
		v.emitToFile(str)
		for _, stmt := range stmt.Body.List {
			if caseClause, ok := stmt.(*ast.CaseClause); ok {
				for _, expr := range caseClause.List {
					str := v.emitAsString("case ", indent+2)
					v.emitToFile(str)
					v.traverseExpression(expr, 0)
					str = v.emitAsString(":\n", 0)
					v.emitToFile(str)
				}
				if len(caseClause.List) == 0 {
					str := v.emitAsString("default:\n", indent+2)
					v.emitToFile(str)
				}
				for _, innerStmt := range caseClause.Body {
					v.traverseStmt(innerStmt, indent+4)
				}
				str := v.emitAsString("break;\n", indent+4)
				v.emitToFile(str)
			}
		}
		str = v.emitAsString("}\n", indent)
		v.emitToFile(str)
	case *ast.BranchStmt:
		var str string
		switch stmt.Tok {
		case token.BREAK:
			str = v.emitAsString("break;\n", indent)
		case token.CONTINUE:
			str = v.emitAsString("continue;\n", indent)
		}
		v.emitToFile(str)
	case *ast.IncDecStmt:
		str := v.emitAsString("", indent)
		v.emitToFile(str)
		v.traverseExpression(stmt.X, indent)
		str = v.emitAsString(stmt.Tok.String(), 0)
		str += v.emitAsString(";\n", 0)
		v.emitToFile(str)
	default:
		fmt.Printf("<Other statement type>\n")

	}
}

func (v *CppBackendVisitor) traverseBlockStmt(block *ast.BlockStmt, indent int) {
	for _, stmt := range block.List {
		v.traverseStmt(stmt, indent)
	}
}

func (v *CppBackendVisitor) traverseIfStmt(ifStmt *ast.IfStmt, indent int, innerif bool) {
	var str string
	if innerif {
		str = v.emitAsString("if", 1)
	} else {
		str = v.emitAsString("if", indent)
	}
	str += v.emitAsString(" (", 0)
	v.emitToFile(str)
	v.traverseExpression(ifStmt.Cond, 0)
	str = v.emitAsString(") ", 0)
	str += v.emitAsString("{\n", 0)
	v.emitToFile(str)
	v.traverseBlockStmt(ifStmt.Body, indent+2)
	str = v.emitAsString("}\n", indent)
	v.emitToFile(str)
	if ifStmt.Else != nil {
		if elseIf, ok := ifStmt.Else.(*ast.IfStmt); ok {
			str = v.emitAsString("else", indent)
			v.emitToFile(str)
			v.traverseIfStmt(elseIf, indent, true) // Recursive call for else-if
		} else if elseBlock, ok := ifStmt.Else.(*ast.BlockStmt); ok {
			str = v.emitAsString("else {\n", indent)
			v.emitToFile(str)
			v.traverseBlockStmt(elseBlock, indent+2) // Dump else block
			str = v.emitAsString("}\n", indent)
			v.emitToFile(str)
		}
	}
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
				v.traverseExpression(arrayArg, 0)
			} else {
				v.traverseExpression(result.Type, 0)
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
				v.traverseExpression(arrayArg, 0)
				v.emitToFile(" ")
				v.traverseExpression(argName, 0)
			} else {
				v.traverseExpression(arg.Type, 0)
				v.emitToFile(" ")
				v.emitToFile(argName.Name)
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
	v.traverseBlockStmt(node.Body, 2)
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
					v.emitToFile(fmt.Sprintf("using "))
					v.traverseExpression(node.Name, 0)
					v.emitToFile(" = ")
					v.traverseExpression(arrayArg, 0)
					v.emitToFile(";\n\n")
				} else {
					str := v.emitAsString(fmt.Sprintf("using %s = ", node.Name.Name), 0)
					err := v.emitToFile(str)
					v.traverseExpression(node.Type, 0)
					str = v.emitAsString(";\n\n", 0)
					err = v.emitToFile(str)
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
	v.emitter.SetFile(v.file)
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
	str := cppVisitor.emitAsString(fmt.Sprintf("namespace %s\n", cppVisitor.pkg.Name), 0)
	err := cppVisitor.emitToFile(str)
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
	str := cppVisitor.emitAsString(fmt.Sprintf("} // namespace %s\n\n", cppVisitor.pkg.Name), 0)
	err = cppVisitor.emitToFile(str)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}

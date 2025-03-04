package main

import (
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

type BasePass struct {
	outputFile string
	file       *os.File
	visitor    *BasePassVisitor
	emitter    Emitter
}

type BasePassVisitor struct {
	pkg     *packages.Package
	pass    *BasePass
	nodes   []ast.Node
	emitter Emitter
}

func (v *BasePass) Name() string {
	return "CppGen"
}

func (v *BasePass) Visitors(pkg *packages.Package) []ast.Visitor {
	v.visitor = &BasePassVisitor{pkg: pkg, emitter: v.emitter}
	v.visitor.pass = v
	return []ast.Visitor{v.visitor}
}

func (v *BasePassVisitor) emitToFile(s string) error {
	_, err := v.pass.file.WriteString(s)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return err
	}
	return nil
}

func (v *BasePassVisitor) emitAsString(s string, indent int) string {
	return strings.Repeat(" ", indent) + s
}

func (v *BasePassVisitor) generateFields(st *ast.StructType, indent int) {
	for _, field := range st.Fields.List {
		for _, fieldName := range field.Names {
			v.traverseExpression(field.Type, indent)
			v.emitToFile(" ")
			v.traverseExpression(fieldName, 0)
			v.emitToFile(";\n")
		}
	}
}

func (v *BasePassVisitor) emitArgs(node *ast.CallExpr, indent int) {
	v.emitter.PreVisitCallExprArgs(node.Args, indent)
	for i, arg := range node.Args {
		v.emitter.PreVisitCallExprArg(arg, i, indent)
		v.traverseExpression(arg, 0) // Function arguments
		v.emitter.PostVisitCallExprArg(arg, i, indent)
	}
	v.emitter.PostVisitCallExprArgs(node.Args, indent)
}

func (v *BasePassVisitor) traverseExpression(expr ast.Expr, indent int) string {
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
				v.traverseExpression(result.Type, indent)
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
		v.emitter.PreVisitKeyValueExpr(e, indent)
		v.emitter.PreVisitKeyValueExprKey(e.Key, indent)
		v.traverseExpression(e.Key, indent)
		v.emitter.PostVisitKeyValueExprKey(e.Key, indent)
		v.emitter.PreVisitKeyValueExprValue(e.Value, indent)
		v.traverseExpression(e.Value, indent)
		v.emitter.PostVisitKeyValueExprValue(e.Value, indent)
		v.emitter.PostVisitKeyValueExpr(e, indent)
	case *ast.FuncLit:
		v.emitter.PreVisitFuncLit(e, indent)
		v.emitter.PreVisitFuncLitTypeParams(e.Type.Params, indent)
		for i, param := range e.Type.Params.List {
			v.emitter.PreVisitFuncLitTypeParam(param, i, indent)
			v.traverseExpression(param.Type, indent)
			v.emitter.PostVisitFuncLitTypeParam(param, i, indent)
		}
		v.emitter.PostVisitFuncLitTypeParams(e.Type.Params, indent)
		v.emitter.PreVisitFuncLitTypeResults(e.Type.Results, indent)

		if e.Type.Results != nil {
			for i, result := range e.Type.Results.List {
				v.emitter.PreVisitFuncLitTypeResult(result, i, indent)
				v.traverseExpression(result.Type, indent)
				v.emitter.PostVisitFuncLitTypeResult(result, i, indent)
			}
		}
		v.emitter.PostVisitFuncLitTypeResults(e.Type.Results, indent)
		v.emitter.PreVisitFuncLitBody(e.Body, indent)
		v.traverseStmt(e.Body, indent+4)
		v.emitter.PostVisitFuncLitBody(e.Body, indent)
		v.emitter.PostVisitFuncLit(e, indent)
	case *ast.TypeAssertExpr:
		v.emitter.PreVisitTypeAssertExpr(e, indent)
		v.emitter.PreVisitTypeAssertExprType(e.Type, indent)
		v.traverseExpression(e.Type, indent)
		v.emitter.PostVisitTypeAssertExprType(e.Type, indent)
		v.emitter.PreVisitTypeAssertExprX(e.X, indent)
		v.traverseExpression(e.X, indent)
		v.emitter.PostVisitTypeAssertExprX(e.X, indent)
		v.emitter.PostVisitTypeAssertExpr(e, indent)
	case *ast.StarExpr:
		v.emitter.PreVisitStarExpr(e, indent)
		v.emitter.PreVisitStarExprX(e, indent)
		v.traverseExpression(e.X, indent)
		v.emitter.PostVisitStarExprX(e, indent)
		v.emitter.PostVisitStarExpr(e, indent)
	case *ast.InterfaceType:
		v.emitter.PreVisitInterfaceType(e, indent)
		v.emitter.PostVisitInterfaceType(e, indent)
	default:
		panic(fmt.Sprintf("unsupported expression type: %T", e))
	}
	return str
}

func (v *BasePassVisitor) traverseAssignment(assignStmt *ast.AssignStmt, indent int) {
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
	for i := 0; i < len(assignStmt.Lhs); i++ {
		if i > 0 {
			str := v.emitAsString(", ", indent)
			v.emitToFile(str)
		}
		v.traverseExpression(assignStmt.Lhs[i], indent)
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

func (v *BasePassVisitor) traverseReturnStmt(retStmt *ast.ReturnStmt, indent int) {
	str := v.emitAsString("return ", indent)
	v.emitToFile(str)
	if len(retStmt.Results) > 1 {
		str := v.emitAsString("std::make_tuple(", 0)
		v.emitToFile(str)
	}
	for i := 0; i < len(retStmt.Results); i++ {
		if i > 0 {
			str := v.emitAsString(", ", 0)
			v.emitToFile(str)
		}
		v.traverseExpression(retStmt.Results[i], 0)
	}
	if len(retStmt.Results) > 1 {
		str := v.emitAsString(")", 0)
		v.emitToFile(str)
	}
	str = v.emitAsString(";", 0)
	v.emitToFile(str)
}

func (v *BasePassVisitor) traverseStmt(stmt ast.Stmt, indent int) {
	switch stmt := stmt.(type) {
	case *ast.ExprStmt:
		v.emitter.PreVisitExprStmt(stmt, indent)
		v.emitter.PreVisitExprStmtX(stmt.X, indent)
		v.traverseExpression(stmt.X, indent)
		v.emitter.PostVisitExprStmtX(stmt.X, indent)
		v.emitter.PostVisitExprStmt(stmt, indent)
	case *ast.DeclStmt:
		v.emitter.PreVisitDeclStmt(stmt, indent)
		if genDecl, ok := stmt.Decl.(*ast.GenDecl); ok && genDecl.Tok == token.VAR {
			for _, spec := range genDecl.Specs {
				if valueSpec, ok := spec.(*ast.ValueSpec); ok {
					// Iterate through all variables declared
					for i := 0; i < len(valueSpec.Names); i++ {
						v.emitter.PreVisitDeclStmtValueSpecType(valueSpec, i, indent)
						v.traverseExpression(valueSpec.Type, indent)
						v.emitter.PostVisitDeclStmtValueSpecType(valueSpec, i, indent)
						v.emitter.PreVisitDeclStmtValueSpecNames(valueSpec.Names[i], i, indent)
						v.traverseExpression(valueSpec.Names[i], 0)
						v.emitter.PostVisitDeclStmtValueSpecNames(valueSpec.Names[i], i, indent)
					}
				}
			}
		}
		v.emitter.PostVisitDeclStmt(stmt, indent)
	case *ast.AssignStmt:
		v.emitter.PreVisitAssignStmt(stmt, indent)
		v.traverseAssignment(stmt, 0)
		v.emitter.PostVisitAssignStmt(stmt, indent)
	case *ast.ReturnStmt:
		v.traverseReturnStmt(stmt, indent)
	case *ast.IfStmt:
		str := v.emitAsString("if (", indent)
		v.emitToFile(str)
		v.traverseExpression(stmt.Cond, 0)
		str = v.emitAsString(")\n", 0)
		v.emitToFile(str)
		v.traverseStmt(stmt.Body, indent)
		if stmt.Else != nil {
			str = v.emitAsString("else", 1)
			v.emitToFile(str)
			v.traverseStmt(stmt.Else, indent)
		}
	case *ast.ForStmt:
		v.emitter.PreVisitForStmt(stmt, indent)
		str := v.emitAsString("for (", indent)
		v.emitToFile(str)
		if stmt.Init != nil {
			v.traverseStmt(stmt.Init, 0)
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
			v.emitter.PreVisitForStmtPost(stmt.Post, indent)
			v.traverseStmt(stmt.Post, 0)
			v.emitter.PostVisitForStmtPost(stmt.Post, indent)
		}
		str = v.emitAsString(")\n", 0)
		v.emitToFile(str)
		v.traverseStmt(stmt.Body, indent)
		v.emitter.PostVisitForStmt(stmt, indent)
	case *ast.RangeStmt:
		str := v.emitAsString("for (auto ", indent)
		v.emitToFile(str)
		if stmt.Value != nil {
			v.traverseExpression(stmt.Value, 0)
		}
		str = v.emitAsString(" : ", 0)
		v.emitToFile(str)
		v.traverseExpression(stmt.X, 0)
		str = v.emitAsString(")\n", 0)
		v.emitToFile(str)
		v.traverseStmt(stmt.Body, indent)
	case *ast.SwitchStmt:
		str := v.emitAsString("switch (", indent)
		v.emitToFile(str)
		v.traverseExpression(stmt.Tag, 0)
		str = v.emitAsString(") {\n", 0)
		v.emitToFile(str)
		for _, stmt := range stmt.Body.List {
			v.traverseStmt(stmt, indent+2)
		}
		str = v.emitAsString("}", indent)
		v.emitToFile(str)
	case *ast.BranchStmt:
		v.emitter.PreVisitBranchStmt(stmt, indent)
		v.emitter.PostVisitBranchStmt(stmt, indent)
	case *ast.IncDecStmt:
		v.emitter.PreVisitIncDecStmt(stmt, indent)
		v.traverseExpression(stmt.X, indent)
		v.emitter.PostVisitIncDecStmt(stmt, indent)
	case *ast.CaseClause:
		for i := 0; i < len(stmt.List); i++ {
			str := v.emitAsString("case ", indent+2)
			v.emitToFile(str)
			v.traverseExpression(stmt.List[i], 0)
			str = v.emitAsString(":\n", 0)
			v.emitToFile(str)
		}
		if len(stmt.List) == 0 {
			str := v.emitAsString("default:\n", indent+2)
			v.emitToFile(str)
		}
		for i := 0; i < len(stmt.Body); i++ {
			v.traverseStmt(stmt.Body[i], indent+4)
		}
		v.emitToFile("\n")
		str := v.emitAsString("break;\n", indent+4)
		v.emitToFile(str)
	case *ast.BlockStmt:
		str := v.emitAsString("{\n", indent)
		v.emitToFile(str)
		v.traverseBlockStmt(stmt, indent+2)
		str = v.emitAsString("}", indent)
		v.emitToFile(str)
	default:
		fmt.Printf("<Other statement type>\n")
	}
}

func (v *BasePassVisitor) traverseBlockStmt(block *ast.BlockStmt, indent int) {
	for i := 0; i < len(block.List); i++ {
		stmt := block.List[i]
		v.traverseStmt(stmt, indent)
		str := v.emitAsString("\n", indent)
		v.emitToFile(str)
	}
}

func (v *BasePassVisitor) generateFuncDeclSignature(node *ast.FuncDecl) ast.Visitor {
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
			v.traverseExpression(result.Type, 0)
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
			v.traverseExpression(arg.Type, 0)
			v.emitToFile(" ")
			v.traverseExpression(argName, 0)
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

func (v *BasePassVisitor) generateFuncDecl(node *ast.FuncDecl) ast.Visitor {
	v.generateFuncDeclSignature(node)
	str := v.emitAsString("\n", 0)
	v.emitToFile(str)
	v.traverseStmt(node.Body, 0)
	str = v.emitAsString("\n\n", 0)
	v.emitToFile(str)
	return v
}

func (v *BasePassVisitor) buildTypesGraph() map[string][]string {
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

func (v *BasePassVisitor) gen(precedence map[string]int) {
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
							v.traverseExpression(valueSpec.Values[i], 0)
							str = v.emitAsString(";\n", 0)
							v.emitToFile(str)
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
				v.emitToFile(fmt.Sprintf("using "))
				v.traverseExpression(node.Name, 0)
				v.emitToFile(" = ")
				v.traverseExpression(node.Type, 0)
				v.emitToFile(";\n\n")
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

func (v *BasePassVisitor) Visit(node ast.Node) ast.Visitor {
	v.nodes = append(v.nodes, node)
	return v
}

func (v *BasePass) ProLog() {
	namespaces = make(map[string]struct{})
	v.emitter.PreVisitProgram(0)
	v.file = v.emitter.GetFile()
}

func (v *BasePass) EpiLog() {
	v.emitter.PostVisitProgram(0)
}

func (v *BasePass) PreVisit(visitor ast.Visitor) {
	cppVisitor := visitor.(*BasePassVisitor)
	namespaces[cppVisitor.pkg.Name] = struct{}{}
	v.emitter.PreVisitPackage(cppVisitor.pkg.Name, 0)
}

func (v *BasePassVisitor) complementPrecedenceMap(sortedTypes map[string]int) {
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

func (v *BasePass) PostVisit(visitor ast.Visitor, visited map[string]struct{}) {
	cppVisitor := visitor.(*BasePassVisitor)
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
	v.emitter.PostVisitPackage(cppVisitor.pkg.Name, 0)
}

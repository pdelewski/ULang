package main

import (
    "fmt"
    "go/ast"
    "go/parser"
    "go/token"
    "log"
    "os"
    "strings"
)

func main() {
    if len(os.Args) < 2 {
	log.Fatalf("Usage: %s <filename.go>", os.Args[0])
    }
    filename := os.Args[1]

    fset := token.NewFileSet()
    node, err := parser.ParseFile(fset, filename, nil, parser.AllErrors)
    if err != nil {
	log.Fatalf("Failed to parse file: %v", err)
    }

    for _, decl := range node.Decls {
	fn, ok := decl.(*ast.FuncDecl)
	if !ok || fn.Body == nil {
	    continue
	}
	generated := transformFunction(fn)
	fmt.Println(generated)
    }
}

func transformFunction(fn *ast.FuncDecl) string {
    var output strings.Builder
    newFuncName := fn.Name.Name + "_with_allocator"
    output.WriteString(fmt.Sprintf("func %s(alloc *Allocator", newFuncName))
    for _, param := range fn.Type.Params.List {
	for _, name := range param.Names {
	    output.WriteString(", " + name.Name + " " + exprToString(param.Type))
	}
    }
    output.WriteString(") {\n")

    for _, stmt := range fn.Body.List {
	switch s := stmt.(type) {

	case *ast.AssignStmt:
	    // Match: x := new(int)
	    if len(s.Rhs) == 1 {
		if call, ok := s.Rhs[0].(*ast.CallExpr); ok {
		    if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "new" {
			varName := s.Lhs[0].(*ast.Ident).Name
			output.WriteString(fmt.Sprintf("\t%s_slot, %s := alloc.AllocAndArena()\n", varName, varName))
			output.WriteString(fmt.Sprintf("\tif %s_slot == -1 {\n\t\treturn\n\t}\n", varName))
			continue
		    }
		}
	    }
	    // Match: *x = ...
	    if len(s.Lhs) == 1 {
		if starExpr, ok := s.Lhs[0].(*ast.StarExpr); ok {
		    if ident, ok := starExpr.X.(*ast.Ident); ok {
			rhs := exprToString(s.Rhs[0])
			output.WriteString(fmt.Sprintf("\t%s[%s_slot] = %s\n", ident.Name, ident.Name, rhs))
			continue
		    }
		}
	    }

	case *ast.ExprStmt:
	    // Match: fmt.Println(*x)
	    if call, ok := s.X.(*ast.CallExpr); ok {
		args := []string{}
		for _, arg := range call.Args {
		    if starArg, ok := arg.(*ast.StarExpr); ok {
			if ident, ok := starArg.X.(*ast.Ident); ok {
			    args = append(args, fmt.Sprintf("%s[%s_slot]", ident.Name, ident.Name))
			    continue
			}
		    }
		    args = append(args, exprToString(arg))
		}
		funcName := exprToString(call.Fun)
		output.WriteString(fmt.Sprintf("\t%s(%s)\n", funcName, strings.Join(args, ", ")))
		continue
	    }
	default:
	    // Fallback to a comment
	    output.WriteString("\t// [unhandled statement]\n")
	}
    }

    output.WriteString("}\n")
    return output.String()
}

// exprToString returns a simplified string representation of an expression
func exprToString(expr ast.Expr) string {
    switch v := expr.(type) {
    case *ast.BasicLit:
	return v.Value
    case *ast.Ident:
	return v.Name
    case *ast.SelectorExpr:
	return exprToString(v.X) + "." + v.Sel.Name
    case *ast.StarExpr:
	return "*" + exprToString(v.X)
    case *ast.CallExpr:
	return exprToString(v.Fun) + "(...)"
    case *ast.ArrayType:
	return "[]" + exprToString(v.Elt)
    case *ast.FuncType:
	return "func(...)"
    case *ast.IndexExpr:
	return exprToString(v.X) + "[" + exprToString(v.Index) + "]"
    default:
	return "<expr>"
    }
}

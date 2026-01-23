package compiler

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/packages"
	"os"
)

// SemaChecker performs semantic analysis to detect unsupported Go constructs.
// Unsupported constructs (checked at compile time):
// 1. iota - constant enumeration
// 2. for _, x := range []T{...} - range over inline composite literal
// 3. if slice == nil / if slice != nil - nil comparison for slices
// 4. type switch statements
// 5. string variable reuse after concatenation (Rust move semantics)
// 6. struct field initialization out of declaration order (C++ designated initializers)
// 7. same variable multiple times in expression (Rust ownership)
// 8. slice self-assignment (Rust borrow checker)
// 9. multiple closures capturing same variable (Rust borrow checker)
//
// Supported (with limitations):
// - interface{} / any - maps to std::any (C++), Box<dyn Any> (Rust), object (C#)
//   Note: type assertions x.(T) supported in C++ only for now
type SemaChecker struct {
	Emitter
	pkg      *packages.Package
	constCtx bool
	// Track string variables consumed by concatenation (for Rust compatibility)
	consumedStringVars map[string]token.Pos
	// Track variables used in closures for multiple closure detection
	closureVars map[string]token.Pos
}

func (sema *SemaChecker) PreVisitPackage(pkg *packages.Package, indent int) {
	sema.pkg = pkg
	// Reset consumed variables map for each package
	sema.consumedStringVars = make(map[string]token.Pos)
	// Reset closure variables map for each package
	sema.closureVars = make(map[string]token.Pos)
}

// PreVisitFuncDecl resets closure tracking for each function
func (sema *SemaChecker) PreVisitFuncDecl(node *ast.FuncDecl, indent int) {
	// Reset closure variables for each function to avoid false positives
	// between closures in different functions
	sema.closureVars = make(map[string]token.Pos)
}

func (sema *SemaChecker) PreVisitGenDeclConstName(node *ast.Ident, indent int) {
	sema.constCtx = true

	// Check if the constant is declared without an explicit type
	if sema.pkg != nil && sema.pkg.TypesInfo != nil {
		if obj := sema.pkg.TypesInfo.Defs[node]; obj != nil {
			if constObj, ok := obj.(*types.Const); ok {
				if basic, ok := constObj.Type().(*types.Basic); ok {
					if basic.Info()&types.IsUntyped != 0 {
						fmt.Printf("\033[33m\033[1mWarning: constant '%s' declared without explicit type\033[0m\n", node.Name)
						fmt.Println("  For cross-platform compatibility, constants should have explicit types.")
						fmt.Println()
						fmt.Println("  \033[33mInstead of:\033[0m")
						fmt.Printf("    const %s = value\n", node.Name)
						fmt.Println()
						fmt.Println("  \033[32mUse explicit type:\033[0m")
						fmt.Printf("    const %s int = value\n", node.Name)
						fmt.Println()
					}
				}
			}
		}
	}
}

func (sema *SemaChecker) PreVisitIdent(node *ast.Ident, indent int) {
	if sema.constCtx {
		if node.String() == "iota" {
			fmt.Println("\033[31m\033[1mCompilation error : iota is not allowed for now\033[0m")
			os.Exit(-1)
		}
	}

	// Check if this identifier was consumed by string concatenation
	if sema.consumedStringVars != nil {
		if consumedPos, wasConsumed := sema.consumedStringVars[node.Name]; wasConsumed {
			// Only error if this use is after the consumption point
			if node.Pos() > consumedPos {
				fmt.Println("\033[31m\033[1mCompilation error: string variable reuse after concatenation\033[0m")
				fmt.Printf("  Variable '%s' was consumed by '+' and cannot be reused.\n", node.Name)
				fmt.Println("  This pattern fails in Rust due to move semantics.")
				fmt.Println()
				fmt.Println("  \033[33mInstead of:\033[0m")
				fmt.Printf("    y = %s + a\n", node.Name)
				fmt.Printf("    z = %s + b  // error: %s was moved\n", node.Name, node.Name)
				fmt.Println()
				fmt.Println("  \033[32mUse separate += statements:\033[0m")
				fmt.Println("    y += a")
				fmt.Println("    y += b")
				os.Exit(-1)
			}
		}
	}

	// Note: "whole struct use after field access" check removed
	// This pattern is valid in Go, and Rust handles it via .clone()
}

func (sema *SemaChecker) PostVisitGenDeclConstName(node *ast.Ident, indent int) {
	sema.constCtx = false
}

func (sema *SemaChecker) PreVisitRangeStmt(node *ast.RangeStmt, indent int) {
	// Handle for _, v := range (value-only): set Key to nil so emitters work correctly
	// for i, v := range (key-value) is now allowed and handled by emitters
	// for i := range (index-only) is allowed (Value is nil)
	if node.Key != nil && node.Value != nil {
		if node.Key.(*ast.Ident).Name == "_" {
			// For value-only range (for _, v := range), set Key to nil so emitters work correctly
			node.Key = nil
		}
		// Otherwise, keep both Key and Value for key-value range loops
	}

	// Check for range over inline composite literal (e.g., for _, x := range []int{1,2,3})
	if _, ok := node.X.(*ast.CompositeLit); ok {
		fmt.Println("\033[31m\033[1mCompilation error : range over inline slice literal (e.g., for _, x := range []int{1,2,3}) is not allowed for now\033[0m")
		os.Exit(-1)
	}
}

// PreVisitBinaryExpr checks for nil comparisons which are not supported for slices
// and tracks string variable consumption for Rust compatibility
func (sema *SemaChecker) PreVisitBinaryExpr(node *ast.BinaryExpr, indent int) {
	// Check for == nil or != nil comparisons
	if node.Op == token.EQL || node.Op == token.NEQ {
		if isNilIdent(node.Y) || isNilIdent(node.X) {
			fmt.Println("\033[31m\033[1mCompilation error : nil comparison (== nil or != nil) is not allowed for now\033[0m")
			os.Exit(-1)
		}
	}

	// Check for string concatenation that consumes variables (Rust move semantics)
	// Pattern: stringVar + "literal" or stringVar + otherVar
	// This pattern causes issues in Rust because the left operand is moved
	if node.Op == token.ADD {
		// Check if left operand is a string type identifier
		if ident, ok := node.X.(*ast.Ident); ok {
			if sema.pkg != nil && sema.pkg.TypesInfo != nil {
				if tv, exists := sema.pkg.TypesInfo.Types[node.X]; exists {
					if tv.Type != nil && tv.Type.String() == "string" {
						// Initialize map if needed
						if sema.consumedStringVars == nil {
							sema.consumedStringVars = make(map[string]token.Pos)
						}
						// Check if this variable was already consumed
						if consumedPos, wasConsumed := sema.consumedStringVars[ident.Name]; wasConsumed {
							if ident.Pos() > consumedPos {
								fmt.Println("\033[31m\033[1mCompilation error: string variable reuse after concatenation\033[0m")
								fmt.Printf("  Variable '%s' was consumed by '+' and cannot be reused.\n", ident.Name)
								fmt.Println("  This pattern fails in Rust due to move semantics.")
								fmt.Println()
								fmt.Println("  \033[33mInstead of:\033[0m")
								fmt.Printf("    y = %s + a\n", ident.Name)
								fmt.Printf("    z = %s + b  // error: %s was moved\n", ident.Name, ident.Name)
								fmt.Println()
								fmt.Println("  \033[32mUse separate += statements:\033[0m")
								fmt.Println("    y += a")
								fmt.Println("    y += b")
								os.Exit(-1)
							}
						}
						// Mark this variable as consumed at this position
						sema.consumedStringVars[ident.Name] = ident.Pos()
					}
				}
			}
		}
	}
}

// PreVisitAssignStmt checks for problematic patterns like: x += x + a
// where x is both borrowed (for +=) and moved (in x + a) in the same statement
// Also checks for slice self-assignment: slice[i] = slice[j]
func (sema *SemaChecker) PreVisitAssignStmt(node *ast.AssignStmt, indent int) {
	// Check for slice self-assignment pattern
	sema.checkSliceSelfAssignment(node)

	// Check for += with string concatenation on RHS that uses the same variable
	if node.Tok == token.ADD_ASSIGN {
		for _, lhs := range node.Lhs {
			if lhsIdent, ok := lhs.(*ast.Ident); ok {
				// Check if this is a string type
				if sema.pkg != nil && sema.pkg.TypesInfo != nil {
					if tv, exists := sema.pkg.TypesInfo.Types[lhs]; exists {
						if tv.Type != nil && tv.Type.String() == "string" {
							// Check if RHS contains a binary + with this variable on the left
							if sema.rhsContainsStringConcatWithVar(node.Rhs[0], lhsIdent.Name) {
								fmt.Println("\033[31m\033[1mCompilation error: self-referencing string concatenation\033[0m")
								fmt.Printf("  Pattern '%s += %s + ...' is not allowed.\n", lhsIdent.Name, lhsIdent.Name)
								fmt.Printf("  Variable '%s' is both borrowed (+=) and moved (+) in the same statement.\n", lhsIdent.Name)
								fmt.Println()
								fmt.Println("  \033[33mInstead of:\033[0m")
								fmt.Printf("    %s += %s + other\n", lhsIdent.Name, lhsIdent.Name)
								fmt.Println()
								fmt.Println("  \033[32mUse separate statements:\033[0m")
								fmt.Printf("    %s += other\n", lhsIdent.Name)
								os.Exit(-1)
							}
						}
					}
				}
			}
		}
	}

	// Check for same non-Copy variable in binary expression: f(x) + g(x)
	for _, rhs := range node.Rhs {
		sema.checkBinaryExprWithSameVar(rhs)
	}
}

// rhsContainsStringConcatWithVar checks if an expression contains a binary + with varName on the left
func (sema *SemaChecker) rhsContainsStringConcatWithVar(expr ast.Expr, varName string) bool {
	switch e := expr.(type) {
	case *ast.BinaryExpr:
		if e.Op == token.ADD {
			if ident, ok := e.X.(*ast.Ident); ok {
				if ident.Name == varName {
					return true
				}
			}
		}
		// Recursively check both sides
		return sema.rhsContainsStringConcatWithVar(e.X, varName) || sema.rhsContainsStringConcatWithVar(e.Y, varName)
	case *ast.ParenExpr:
		return sema.rhsContainsStringConcatWithVar(e.X, varName)
	}
	return false
}

// PostVisitAssignStmt clears consumed state for variables that are reassigned
// This handles patterns like: x = x + a; y = x + b (which should be valid)
func (sema *SemaChecker) PostVisitAssignStmt(node *ast.AssignStmt, indent int) {
	if sema.consumedStringVars == nil {
		return
	}
	// Clear consumed state for any string variables on the LHS
	for _, lhs := range node.Lhs {
		if ident, ok := lhs.(*ast.Ident); ok {
			delete(sema.consumedStringVars, ident.Name)
		}
	}
}

// PreVisitInterfaceType checks for interface{} / any type usage
func (sema *SemaChecker) PreVisitInterfaceType(node *ast.InterfaceType, indent int) {
	// Empty interface (interface{} / any) is now supported
	// Maps to: C++ std::any, Rust Box<dyn Any>, C# object
}

// PreVisitTypeSwitchStmt checks for type switch statements (not supported)
func (sema *SemaChecker) PreVisitTypeSwitchStmt(node *ast.TypeSwitchStmt, indent int) {
	fmt.Println("\033[31m\033[1mCompilation error : type switch statement is not allowed for now\033[0m")
	os.Exit(-1)
}

// isNilIdent checks if an expression is the nil identifier
func isNilIdent(expr ast.Expr) bool {
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name == "nil"
	}
	return false
}

// PreVisitCompositeLit checks for struct field initialization order
// C++ designated initializers require fields to be in declaration order
func (sema *SemaChecker) PreVisitCompositeLit(node *ast.CompositeLit, indent int) {
	if sema.pkg == nil || sema.pkg.TypesInfo == nil {
		return
	}

	// Get the type of the composite literal
	tv, ok := sema.pkg.TypesInfo.Types[node]
	if !ok || tv.Type == nil {
		return
	}

	// Check if it's a struct type
	structType, ok := tv.Type.Underlying().(*types.Struct)
	if !ok {
		return
	}

	// Get field names from the struct declaration (in order)
	declaredFields := make([]string, structType.NumFields())
	fieldIndex := make(map[string]int)
	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		declaredFields[i] = field.Name()
		fieldIndex[field.Name()] = i
	}

	// Get field names from the initialization (in order)
	var initFields []string
	for _, elt := range node.Elts {
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			if ident, ok := kv.Key.(*ast.Ident); ok {
				initFields = append(initFields, ident.Name)
			}
		}
	}

	// If no keyed fields, skip the order check (positional initialization)
	if len(initFields) == 0 {
		return
	}

	// Check if initialization order matches declaration order
	lastIndex := -1
	for _, fieldName := range initFields {
		idx, exists := fieldIndex[fieldName]
		if !exists {
			continue // Unknown field, skip
		}
		if idx < lastIndex {
			// Fields are out of order
			fmt.Println("\033[33m\033[1mWarning: struct field initialization order does not match declaration order\033[0m")
			fmt.Printf("  Field '%s' appears before a previously initialized field.\n", fieldName)
			fmt.Println("  C++ designated initializers require fields in declaration order.")
			fmt.Println()
			fmt.Println("  \033[36mDeclared order:\033[0m")
			for _, f := range declaredFields {
				fmt.Printf("    - %s\n", f)
			}
			fmt.Println()
			fmt.Println("  \033[36mInitialization order:\033[0m")
			for _, f := range initFields {
				fmt.Printf("    - %s\n", f)
			}
			fmt.Println()
			fmt.Println("  \033[32mPlease reorder the initializers to match the struct declaration.\033[0m")
			os.Exit(-1)
		}
		lastIndex = idx
	}
}

// isNonCopyType checks if a type requires cloning in Rust (non-Copy types)
func (sema *SemaChecker) isNonCopyType(t types.Type) bool {
	if t == nil {
		return false
	}
	typeStr := t.String()
	// Strings are non-Copy
	if typeStr == "string" {
		return true
	}
	// Slices are non-Copy
	if _, ok := t.Underlying().(*types.Slice); ok {
		return true
	}
	// Named struct types are non-Copy
	if named, ok := t.(*types.Named); ok {
		if _, isStruct := named.Underlying().(*types.Struct); isStruct {
			return true
		}
	}
	return false
}

// collectIdentifiers collects identifiers that are actually "consumed" (moved) in an expression
// It excludes identifiers used as the base of selector expressions (field access doesn't move)
// and identifiers that are the Sel part of selector expressions (field names)
func (sema *SemaChecker) collectIdentifiers(node ast.Node) []*ast.Ident {
	var idents []*ast.Ident
	// Track identifiers that are used as selector bases (field access doesn't move)
	selectorBases := make(map[*ast.Ident]bool)
	// Track identifiers that are field names (the Sel part of SelectorExpr)
	selectorFields := make(map[*ast.Ident]bool)

	// First pass: find all selector bases and field names
	ast.Inspect(node, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok {
			if ident, ok := sel.X.(*ast.Ident); ok {
				selectorBases[ident] = true
			}
			// The Sel part is always the field name, not a variable
			selectorFields[sel.Sel] = true
		}
		return true
	})

	// Second pass: collect identifiers that are not selector bases or field names
	ast.Inspect(node, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok {
			if !selectorBases[ident] && !selectorFields[ident] {
				idents = append(idents, ident)
			}
		}
		return true
	})
	return idents
}

// getDirectFunctionArgs returns variable names passed directly to a function call
// Returns empty if the expression is not a direct function call
func (sema *SemaChecker) getDirectFunctionArgs(expr ast.Expr) map[string]bool {
	args := make(map[string]bool)
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return args
	}

	// Skip built-in functions
	if ident, ok := call.Fun.(*ast.Ident); ok {
		builtins := map[string]bool{
			"len": true, "cap": true, "append": true, "copy": true,
			"make": true, "new": true, "delete": true, "close": true,
			"panic": true, "recover": true, "print": true, "println": true,
			"complex": true, "real": true, "imag": true,
		}
		if builtins[ident.Name] {
			return args
		}
	}

	// Collect direct identifier arguments
	for _, arg := range call.Args {
		if ident, ok := arg.(*ast.Ident); ok {
			args[ident.Name] = true
		}
	}
	return args
}

// checkBinaryExprWithSameVar checks for binary expressions like f(x) + g(x)
// where the same non-Copy variable is passed to function calls on both sides
func (sema *SemaChecker) checkBinaryExprWithSameVar(node ast.Node) {
	if sema.pkg == nil || sema.pkg.TypesInfo == nil {
		return
	}

	ast.Inspect(node, func(n ast.Node) bool {
		// Skip closure bodies
		if _, ok := n.(*ast.FuncLit); ok {
			return false
		}

		binExpr, ok := n.(*ast.BinaryExpr)
		if !ok {
			return true
		}

		// Get args from left and right sides (only if they are direct function calls)
		leftArgs := sema.getDirectFunctionArgs(binExpr.X)
		rightArgs := sema.getDirectFunctionArgs(binExpr.Y)

		// Check for same variable in both sides
		for varName := range leftArgs {
			if rightArgs[varName] {
				// Check if it's a non-Copy type
				// We need to find the identifier to get type info
				var foundIdent *ast.Ident
				ast.Inspect(binExpr.X, func(inner ast.Node) bool {
					if ident, ok := inner.(*ast.Ident); ok && ident.Name == varName {
						foundIdent = ident
						return false
					}
					return true
				})

				if foundIdent != nil {
					if obj := sema.pkg.TypesInfo.Uses[foundIdent]; obj != nil {
						if _, isConst := obj.(*types.Const); !isConst {
							if _, isFunc := obj.(*types.Func); !isFunc {
								if sema.isNonCopyType(obj.Type()) {
									fmt.Println("\033[31m\033[1mCompilation error: same variable used multiple times in expression\033[0m")
									fmt.Printf("  Variable '%s' (non-Copy type) appears in both sides of a binary expression.\n", varName)
									fmt.Println("  This pattern fails in Rust due to move semantics.")
									fmt.Println()
									fmt.Println("  \033[33mInstead of:\033[0m")
									fmt.Printf("    foo(%s) + bar(%s)\n", varName, varName)
									fmt.Println()
									fmt.Println("  \033[32mUse separate statements:\033[0m")
									fmt.Printf("    a := foo(%s)\n", varName)
									fmt.Printf("    b := bar(%s.clone())  // or redesign to avoid multiple uses\n", varName)
									os.Exit(-1)
								}
							}
						}
					}
				}
			}
		}
		return true
	})
}

// checkSliceSelfAssignment checks for slice[i] = slice[j] pattern
// This causes Rust borrow checker issues (mutable + immutable borrow)
func (sema *SemaChecker) checkSliceSelfAssignment(node *ast.AssignStmt) {
	if sema.pkg == nil || sema.pkg.TypesInfo == nil {
		return
	}

	// Check if LHS is an index expression on a slice
	for _, lhs := range node.Lhs {
		lhsIndex, ok := lhs.(*ast.IndexExpr)
		if !ok {
			continue
		}

		// Get the slice name from LHS
		lhsSlice, ok := lhsIndex.X.(*ast.Ident)
		if !ok {
			continue
		}

		// Check if it's a slice type
		if tv, exists := sema.pkg.TypesInfo.Types[lhsIndex.X]; exists {
			if _, isSlice := tv.Type.Underlying().(*types.Slice); !isSlice {
				continue
			}
		} else {
			continue
		}

		// Check RHS for index expression on the same slice
		for _, rhs := range node.Rhs {
			rhsIndex, ok := rhs.(*ast.IndexExpr)
			if !ok {
				continue
			}

			rhsSlice, ok := rhsIndex.X.(*ast.Ident)
			if !ok {
				continue
			}

			// Check if same slice
			if lhsSlice.Name == rhsSlice.Name {
				fmt.Println("\033[31m\033[1mCompilation error: slice self-assignment pattern\033[0m")
				fmt.Printf("  Pattern '%s[i] = %s[j]' is not allowed.\n", lhsSlice.Name, rhsSlice.Name)
				fmt.Println("  This causes Rust borrow checker issues (simultaneous mutable and immutable borrow).")
				fmt.Println()
				fmt.Println("  \033[33mInstead of:\033[0m")
				fmt.Printf("    %s[i] = %s[j]\n", lhsSlice.Name, rhsSlice.Name)
				fmt.Println()
				fmt.Println("  \033[32mUse a temporary variable:\033[0m")
				fmt.Printf("    tmp := %s[j]\n", rhsSlice.Name)
				fmt.Printf("    %s[i] = tmp\n", lhsSlice.Name)
				os.Exit(-1)
			}
		}
	}
}

// PreVisitFuncLit checks for multiple closures capturing the same non-Copy variable
func (sema *SemaChecker) PreVisitFuncLit(node *ast.FuncLit, indent int) {
	if sema.pkg == nil || sema.pkg.TypesInfo == nil {
		return
	}

	// Collect all identifiers used in the closure body
	idents := sema.collectIdentifiers(node.Body)

	// Dedupe: track which variables we've already processed for this closure
	processedInThisClosure := make(map[string]bool)

	for _, ident := range idents {
		// Skip if we've already processed this variable name in this closure
		if processedInThisClosure[ident.Name] {
			continue
		}

		// Get the object this identifier refers to
		if obj := sema.pkg.TypesInfo.Uses[ident]; obj != nil {
			// Skip constants, functions, and type names
			if _, isConst := obj.(*types.Const); isConst {
				continue
			}
			if _, isFunc := obj.(*types.Func); isFunc {
				continue
			}
			if _, isTypeName := obj.(*types.TypeName); isTypeName {
				continue
			}

			// Check if it's a non-Copy type
			if !sema.isNonCopyType(obj.Type()) {
				continue
			}

			// Check if this is a variable from an outer scope (closure capture)
			// by verifying it was declared before this function literal
			if obj.Pos() < node.Pos() {
				// Mark as processed for this closure
				processedInThisClosure[ident.Name] = true

				// Check if another closure already captured this variable
				if prevPos, exists := sema.closureVars[ident.Name]; exists {
					fmt.Println("\033[31m\033[1mCompilation error: multiple closures capture same variable\033[0m")
					fmt.Printf("  Variable '%s' (non-Copy type) is captured by multiple closures.\n", ident.Name)
					fmt.Println("  This causes Rust borrow checker issues.")
					fmt.Println()
					fmt.Println("  \033[33mInstead of:\033[0m")
					fmt.Printf("    fn1 := func() { use(%s) }\n", ident.Name)
					fmt.Printf("    fn2 := func() { use(%s) }  // error: already captured\n", ident.Name)
					fmt.Println()
					fmt.Println("  \033[32mUse Arc/Rc for shared ownership or redesign:\033[0m")
					fmt.Printf("    %s1 := %s.clone()\n", ident.Name, ident.Name)
					fmt.Printf("    fn1 := func() { use(%s) }\n", ident.Name)
					fmt.Printf("    fn2 := func() { use(%s1) }\n", ident.Name)
					_ = prevPos // suppress unused variable warning
					os.Exit(-1)
				}
				// Mark this variable as captured by a closure
				sema.closureVars[ident.Name] = node.Pos()
			}
		}
	}
}


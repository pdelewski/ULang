package compiler

import (
	"fmt"
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/packages"
	"os"
)

// SemaChecker performs semantic analysis to detect unsupported Go constructs.
// Unsupported constructs (checked at compile time):
// 1. iota - constant enumeration
// 2. for key, value := range - range with both index and value (only for _, value allowed)
// 3. for _, x := range []T{...} - range over inline composite literal
// 4. if slice == nil / if slice != nil - nil comparison for slices
// 5. type switch statements
// 6. string variable reuse after concatenation (Rust move semantics)
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
}

func (sema *SemaChecker) PreVisitPackage(pkg *packages.Package, indent int) {
	sema.pkg = pkg
	// Reset consumed variables map for each package
	sema.consumedStringVars = make(map[string]token.Pos)
}

func (sema *SemaChecker) PreVisitGenDeclConstName(node *ast.Ident, indent int) {
	sema.constCtx = true
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
				fmt.Printf("\033[31m\033[1mCompilation error : string variable '%s' was consumed by concatenation and cannot be reused (Rust compatibility). Use separate += statements instead of 'a + b' patterns.\033[0m\n", node.Name)
				os.Exit(-1)
			}
		}
	}
}

func (sema *SemaChecker) PostVisitGenDeclConstName(node *ast.Ident, indent int) {
	sema.constCtx = false
}

func (sema *SemaChecker) PreVisitRangeStmt(node *ast.RangeStmt, indent int) {
	// Check for for key, value := range (both key and value)
	// Allowed: for _, v := range slice (value-only)
	// Allowed: for i := range slice (index-only, Value is nil)
	// Not allowed: for i, v := range slice (both key and value)
	if node.Key != nil && node.Value != nil {
		if node.Key.(*ast.Ident).Name != "_" {
			fmt.Println("\033[31m\033[1mCompilation error : for key, value := range is not allowed for now\033[0m")
			os.Exit(-1)
		}
		// For value-only range (for _, v := range), set Key to nil so emitters work correctly
		node.Key = nil
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
								fmt.Printf("\033[31m\033[1mCompilation error : string variable '%s' was consumed by concatenation and cannot be reused (Rust compatibility). Use separate += statements instead of 'a + b' patterns.\033[0m\n", ident.Name)
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
func (sema *SemaChecker) PreVisitAssignStmt(node *ast.AssignStmt, indent int) {
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
								fmt.Printf("\033[31m\033[1mCompilation error : cannot use '%s += %s + ...' pattern (Rust compatibility). Variable '%s' is both borrowed and moved. Use separate statements: '%s += %s; %s += ...' instead.\033[0m\n", lhsIdent.Name, lhsIdent.Name, lhsIdent.Name, lhsIdent.Name, lhsIdent.Name, lhsIdent.Name)
								os.Exit(-1)
							}
						}
					}
				}
			}
		}
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

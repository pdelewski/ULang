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
//
// Supported (with limitations):
// - interface{} / any - maps to std::any (C++), Box<dyn Any> (Rust), object (C#)
//   Note: type assertions x.(T) supported in C++ only for now
type SemaChecker struct {
	Emitter
	pkg      *packages.Package
	constCtx bool
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
func (sema *SemaChecker) PreVisitBinaryExpr(node *ast.BinaryExpr, indent int) {
	// Check for == nil or != nil comparisons
	if node.Op == token.EQL || node.Op == token.NEQ {
		if isNilIdent(node.Y) || isNilIdent(node.X) {
			fmt.Println("\033[31m\033[1mCompilation error : nil comparison (== nil or != nil) is not allowed for now\033[0m")
			os.Exit(-1)
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

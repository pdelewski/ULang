package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"golang.org/x/tools/go/packages"
	"os"
)

type Visitor struct{}

func (v *Visitor) Visit(node ast.Node) ast.Visitor {
	if fn, ok := node.(*ast.FuncDecl); ok {
		fmt.Printf("  Found function: %s\n", fn.Name.Name)
	}
	return v
}

func main() {
	var sourceDir string
	flag.StringVar(&sourceDir, "source", "", "./../uql")
	flag.Parse()
	if sourceDir == "" {
		fmt.Println("Please provide a source directory")
		return
	}
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedSyntax |
			packages.NeedTypes,
		Dir: sourceDir,
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		fmt.Println("Error loading packages:", err)
		return
	}

	if len(pkgs) == 0 {
		fmt.Println("No packages found")
		return
	}

	for _, pkg := range pkgs {
		fmt.Printf("Package: %s\n", pkg.Name)
		fmt.Println("Files:")
		for _, file := range pkg.GoFiles {
			fmt.Printf("  %s\n", file)
			fset := token.NewFileSet()

			parsedFile, err := parser.ParseFile(fset, file, nil, parser.AllErrors)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing file %s: %v\n", file, err)
				continue
			}
			v := &Visitor{}
			ast.Walk(v, parsedFile)
		}
	}
}

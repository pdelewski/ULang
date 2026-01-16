package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/tools/go/callgraph/static"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa/ssautil"
)

// edge represents a caller->callee relation for printing/sorting
type edge struct {
	caller string
	callee string
}

func main() {
	var sourceDir string
	var outputBase string
	var format string
	flag.StringVar(&sourceDir, "source", ".", "Source directory to analyze (module or package root)")
	flag.StringVar(&outputBase, "output", "callgraph", "Output file base name (without extension)")
	flag.StringVar(&format, "format", "both", "Output format: text|dot|both")
	flag.Parse()

	absSource, err := filepath.Abs(sourceDir)
	if err != nil {
		log.Fatalf("failed to resolve source directory: %v", err)
	}

	cfg := &packages.Config{
		Mode:  packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedImports | packages.NeedModule | packages.NeedDeps,
		Dir:   absSource,
		Tests: false,
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		log.Fatalf("failed to load packages: %v", err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		// keep going but warn
		log.Printf("warning: some packages reported errors, call graph may be incomplete")
	}
	if len(pkgs) == 0 {
		log.Fatalf("no packages found in %s", absSource)
	}

	prog, _ := ssautil.AllPackages(pkgs, 0)
	prog.Build()

	g := static.CallGraph(prog)

	// Collect edges
	edges := make([]edge, 0, 1024)
	seen := make(map[string]struct{})
	for _, n := range g.Nodes {
		if n == nil || n.Func == nil || n.Func.Pkg == nil {
			continue
		}
		caller := qualified(n.Func.String())
		for _, e := range n.Out {
			if e == nil || e.Callee == nil || e.Callee.Func == nil || e.Callee.Func.Pkg == nil {
				continue
			}
			callee := qualified(e.Callee.Func.String())
			key := caller + "->" + callee
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			edges = append(edges, edge{caller: caller, callee: callee})
		}
	}

	sort.Slice(edges, func(i, j int) bool {
		if edges[i].caller == edges[j].caller {
			return edges[i].callee < edges[j].callee
		}
		return edges[i].caller < edges[j].caller
	})

	// Write outputs
	if format == "text" || format == "both" {
		if err := writeText(edges, outputBase+".txt"); err != nil {
			log.Fatalf("failed writing text callgraph: %v", err)
		}
	}
	if format == "dot" || format == "both" {
		if err := writeDOT(edges, outputBase+".dot"); err != nil {
			log.Fatalf("failed writing dot callgraph: %v", err)
		}
	}
}

func qualified(name string) string {
	// ssa.Function.String() often returns "pkg.fn" or "(*T).m" forms; normalize for graph readability
	// Keep as-is but strip package pointers noise
	s := strings.ReplaceAll(name, "\n", " ")
	s = strings.TrimSpace(s)
	return s
}

func writeText(edges []edge, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, e := range edges {
		if _, err := fmt.Fprintf(f, "%s -> %s\n", e.caller, e.callee); err != nil {
			return err
		}
	}
	return nil
}

func writeDOT(edges []edge, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := fmt.Fprintln(f, "digraph CallGraph {"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(f, "  rankdir=LR;"); err != nil {
		return err
	}
	// Escape helper
	esc := func(s string) string {
		s = strings.ReplaceAll(s, "\\", "\\\\")
		s = strings.ReplaceAll(s, "\"", "\\\"")
		return s
	}
	for _, e := range edges {
		if _, err := fmt.Fprintf(f, "  \"%s\" -> \"%s\";\n", esc(e.caller), esc(e.callee)); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(f, "}"); err != nil {
		return err
	}
	return nil
}

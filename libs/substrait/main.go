package main

import "fmt"

// ---------- Core Types ----------

const (
	ExprLiteral  ExprKind = 0
	ExprColumn   ExprKind = 1
	ExprFunction ExprKind = 2
)

type ExprKind int

type ExprNode struct {
	Kind     ExprKind
	Children []int
	ValueIdx int
	Function string
}

const (
	RelOpScan    int = 0
	RelOpFilter  int = 1
	RelOpProject int = 2
)

type RelNode struct {
	Kind     int
	InputIdx int
	ExprIdxs []int
	NameIdx  int
}

type Plan struct {
	RelNodes []RelNode
	Exprs    []ExprNode
	Literals []string
	Root     int
}

// ---------- Pure Value-Based Helpers ----------

func AddLiteral(plan Plan, value string) (Plan, int) {
	i := 0
	for _, lit := range plan.Literals {
		if lit == value {
			return plan, i
		}
		i++
	}
	plan.Literals = append(plan.Literals, value)
	return plan, len(plan.Literals) - 1
}

func AddExpr(plan Plan, expr ExprNode) (Plan, int) {
	plan.Exprs = append(plan.Exprs, expr)
	return plan, len(plan.Exprs) - 1
}

func AddRel(plan Plan, rel RelNode) (Plan, int) {
	plan.RelNodes = append(plan.RelNodes, rel)
	return plan, len(plan.RelNodes) - 1
}

func formatIntSlice(s []int) string {
	if len(s) == 0 {
		return "[]"
	}
	var result string
	result = "["
	i := 0
	for _, val := range s {
		if i > 0 {
			result += " "
		}
		result += fmt.Sprintf("%d", val)
		i++
	}
	result += "]"
	return result
}

func printPlan(plan Plan) {
	fmt.Printf("Root RelNode Index: %d\n", plan.Root)

	fmt.Println("\nRelNodes:")
	i := 0
	for _, r := range plan.RelNodes {
		fmt.Printf("  [%d] Kind=%d, InputIdx=%d, NameIdx=%d\n",
			i, r.Kind, r.InputIdx, r.NameIdx)
		fmt.Println(formatIntSlice(r.ExprIdxs))
		i++
	}

	fmt.Println("\nExprs:")
	i = 0
	for _, e := range plan.Exprs {
		fmt.Printf("  [%d] Kind=%d, ValueIdx=%d\n",
			i, e.Kind, e.ValueIdx)
		fmt.Println(formatIntSlice(e.Children))
		i++
	}

	fmt.Println("\nLiterals:")
}

// ---------- Main ----------

func main() {
	var plan Plan
	plan.Literals = []string{}
	// Add literal "my_table" for scan
	idx := 0
	plan, idx = AddLiteral(plan, "my_table")
	scan := RelNode{
		Kind:    RelOpScan,
		NameIdx: idx,
	}
	scanIdx := 0
	plan, scanIdx = AddRel(plan, scan)

	xIdx := 0
	// Add column expression "x"
	plan, xIdx = AddLiteral(plan, "x")
	colExpr := ExprNode{
		Kind:     ExprColumn,
		ValueIdx: xIdx,
	}
	colExprIdx := 0
	plan, colExprIdx = AddExpr(plan, colExpr)

	// Add filter node
	filter := RelNode{
		Kind:     RelOpFilter,
		InputIdx: scanIdx,
		ExprIdxs: []int{colExprIdx},
	}
	filterIdx := 0
	plan, filterIdx = AddRel(plan, filter)

	plan.Root = filterIdx

	// Output for demonstration
	printPlan(plan)
}

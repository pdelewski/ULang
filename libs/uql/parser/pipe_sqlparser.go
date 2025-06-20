package parser

import (
	"uql/lexer"
)

type Node struct {
	Type     int8
	Tok      lexer.Token
	Children []int16
	Nodes    []Node
}

/*
func ParsePipeSql(statement string) Node {
	fmt.Println("Parsing SQL statement")
	var Nodes []Node

	Nodes = append(Nodes, Node{Type: 1, Children: []int16{1, 2}})
	Nodes = append(Nodes, Node{Type: 2, Children: []int16{}})
	Nodes = append(Nodes, Node{Type: 3, Children: []int16{}})

	var root Node
	root = Nodes[0]
	root.Nodes = Nodes
	return root
}

func TraverseTree(root Node) {
	traverseTree(root.Nodes, 0)
}

func traverseTree(nodes []Node, index int) {
	node := nodes[index]
	fmt.Print("Visiting Node: Type=")
	fmt.Print(node.Type)
	fmt.Print(" Token:")
	fmt.Println(lexer.DumpTokenString(node.Tok))
	for _, childIdx := range node.Children {
		traverseTree(nodes, int(childIdx)) // Recurse for each child
	}
}
*/

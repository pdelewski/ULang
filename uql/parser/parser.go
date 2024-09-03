package parser

import (
	"ULang/lexer"
	"ULang/uql/ast"
	uqllexer "ULang/uql/lexer"
	"fmt"
)

func sliceToInt64(slice []int8) int64 {
	var result int64
	for i, b := range slice {
		result |= int64(b) << (8 * i)
	}
	return result
}

var precedence = map[int64]int8{
	sliceToInt64([]int8{'&', '&'}): 1,
	sliceToInt64([]int8{'|', '|'}): 1,
	sliceToInt64([]int8{'>'}):      2,
	sliceToInt64([]int8{'<'}):      2,
	sliceToInt64([]int8{'>', '='}): 2,
	sliceToInt64([]int8{'<', '='}): 2,
	sliceToInt64([]int8{'=', '='}): 2,
	sliceToInt64([]int8{'!', '='}): 2,
}
var associativity = map[int64]int8{
	sliceToInt64([]int8{'&', '&'}): 'L',
	sliceToInt64([]int8{'|', '|'}): 'L',
	sliceToInt64([]int8{'>'}):      'L',
	sliceToInt64([]int8{'<'}):      'L',
	sliceToInt64([]int8{'>', '='}): 'L',
	sliceToInt64([]int8{'<', '='}): 'L',
	sliceToInt64([]int8{'=', '='}): 'L',
	sliceToInt64([]int8{'!', '='}): 'L',
}

func parseExpression(tokens []lexer.Token, minPrecedence int8) (ast.LogicalExpr, int) {
	lhs := ast.LogicalExpr{
		Value: tokens[0],
	}

	i := 1
	for i < len(tokens) {
		token := tokens[i]
		tokenPrecedence, exists := precedence[sliceToInt64(token.Representation)]

		if !exists || tokenPrecedence < minPrecedence {
			break
		}

		// Handle right-associative operators
		nextPrecedence := tokenPrecedence
		if associativity[sliceToInt64(token.Representation)] == 'L' {
			nextPrecedence += 1
		}

		rhsExpr, nextPos := parseExpression(tokens[i+1:], nextPrecedence)
		rhsIndex := uint16(len(lhs.Expressions) + 1)

		lhs = ast.LogicalExpr{
			Value: token,
			Left:  0,
			Right: rhsIndex,
			Expressions: []ast.LogicalExpr{
				lhs,
				rhsExpr,
			},
		}

		i += nextPos + 1
	}

	return lhs, i
}

func printLogicalExpr(expr ast.LogicalExpr, depth int) {
	indent := ' '
	fmt.Printf("%sValue: %s\n", indent, expr.Value.Representation)

	if expr.Left != 0 || expr.Right != 0 {
		fmt.Printf("%sLeft:\n", indent)
		printLogicalExpr(expr.Expressions[0], depth+1)
		fmt.Printf("%sRight:\n", indent)
		printLogicalExpr(expr.Expressions[1], depth+1)
	}
}

func Parse(text string) []lexer.Token {
	tokens := lexer.GetTokens(uqllexer.StringToToken(text))
	lexer.DumpTokens(tokens)
	return tokens
}

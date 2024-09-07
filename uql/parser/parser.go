package parser

import (
	"ULang/lexer"
	"ULang/uql/ast"
	uqllexer "ULang/uql/lexer"
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

func ParseExpression(tokens []lexer.Token) (ast.LogicalExpr, error) {
	expr, _ := parseExpression(tokens, 0)
	return expr, nil
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

func WalkLogicalExpr(expr ast.LogicalExpr,
	state any,
	preVisit func(state any, expr ast.LogicalExpr) any,
	postVisit func(state any, expr ast.LogicalExpr) any,
) {
	walkLogicalExpr(expr, 0, state, preVisit, postVisit)
}

func walkLogicalExpr(
	expr ast.LogicalExpr,
	depth int,
	state any,
	preVisit func(state any, expr ast.LogicalExpr) any,
	postVisit func(state any, expr ast.LogicalExpr) any,
) {
	state = preVisit(state, expr)
	if expr.Left != 0 || expr.Right != 0 {
		walkLogicalExpr(expr.Expressions[0], depth+1, state, preVisit, postVisit)
		walkLogicalExpr(expr.Expressions[1], depth+1, state, preVisit, postVisit)
	}
	state = postVisit(state, expr)
}

func parseFrom(tokens []lexer.Token, lhs lexer.Token) (ast.From, int8) {
	return ast.From{ResultTableExpr: lhs}, 0
}

func parseWhere(tokens []lexer.Token, lhs lexer.Token) (ast.Where, int8) {
	return ast.Where{ResultTableExpr: lhs}, 0
}

func parseSelect(tokens []lexer.Token, lhs lexer.Token) (ast.Select, int8) {
	return ast.Select{ResultTableExpr: lhs}, 0
}

func Parse(text string) (ast.AST, int8) {
	tokens := lexer.GetTokens(uqllexer.StringToToken(text))
	token, tokens := lexer.GetNextToken(tokens)

	if !lexer.IsAlpha(token.Representation[0]) {
		return nil, -1
	}
	lhs := token
	token, tokens = lexer.GetNextToken(tokens)
	if !lexer.IsEqual(token.Representation[0]) {
		return nil, -1
	}

	token, tokens = lexer.GetNextToken(tokens)
	if !lexer.IsFrom(token) && !lexer.IsWhere(token) && !lexer.IsSelect(token) {
		return nil, -1
	}

	if lexer.IsFrom(token) {
		parseFrom(tokens, lhs)
	}

	if lexer.IsWhere(token) {
		parseWhere(tokens, lhs)
	}

	if lexer.IsSelect(token) {
		parseSelect(tokens, lhs)
	}

	return ast.AST{}, 0
}

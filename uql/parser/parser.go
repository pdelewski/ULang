package parser

import (
	"uql/ast"
	"uql/lexer"
)

func sliceToInt64(slice []int8) int64 {
	var result int64
	for i, b := range slice {
		result |= int64(b) << (8 * i)
	}
	return result
}

func precedence(operator []int8) int8 {
	if operator[0] == '&' && operator[1] == '&' {
		return 1
	}
	if operator[0] == '|' && operator[1] == '|' {
		return 1
	}
	if operator[0] == '>' {
		return 2
	}
	if operator[0] == '<' {
		return 2
	}
	if operator[0] == '>' && operator[1] == '=' {
		return 2
	}
	if operator[0] == '<' && operator[1] == '=' {
		return 2
	}
	if operator[0] == '=' && operator[1] == '=' {
		return 2
	}
	if operator[0] == '!' && operator[1] == '=' {
		return 2
	}
	return -1
}

func associativity(operator []int8) int8 {
	if operator[0] == '&' && operator[1] == '&' {
		return 'L'
	}
	if operator[0] == '|' && operator[1] == '|' {
		return 'L'
	}
	if operator[0] == '>' {
		return 'L'
	}
	if operator[0] == '<' {
		return 'L'
	}
	if operator[0] == '>' && operator[1] == '=' {
		return 'L'
	}
	if operator[0] == '<' && operator[1] == '=' {
		return 'L'
	}
	if operator[0] == '=' && operator[1] == '=' {
		return 'L'
	}
	if operator[0] == '!' && operator[1] == '=' {
		return 'L'
	}
	return 'L'
}

func ParseExpression(tokens []lexer.Token) (ast.LogicalExpr, int) {
	expr, index := parseExpression(tokens, 0)
	return expr, index
}

func parseExpression(tokens []lexer.Token, minPrecedence int8) (ast.LogicalExpr, int) {
	lhs := ast.LogicalExpr{
		Value: tokens[0],
	}

	i := 1
	for i < len(tokens) {
		token := tokens[i]
		tokenPrecedence := precedence(token.Representation)

		if tokenPrecedence == -1 || tokenPrecedence < minPrecedence {
			break
		}

		// Handle right-associative operators
		nextPrecedence := tokenPrecedence
		if associativity(token.Representation) == 'L' {
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

func parseFrom(tokens []lexer.Token, lhs lexer.Token) (ast.From, []lexer.Token) {
	from := ast.From{ResultTableExpr: lhs}
	for {
		var token lexer.Token
		token, tokens = lexer.GetNextToken(tokens)
		// TODO handle more than one table
		from.TableExpr = append(from.TableExpr, token)
		if lexer.IsSemicolon(token.Representation[0]) {
			break
		}
	}
	return from, tokens
}

func parseWhere(tokens []lexer.Token, lhs lexer.Token) (ast.Where, []lexer.Token) {
	expr, i := ParseExpression(tokens)
	tokens = tokens[i:]
	for {
		var token lexer.Token
		token, tokens = lexer.GetNextToken(tokens)
		if lexer.IsSemicolon(token.Representation[0]) {
			break
		}
	}
	return ast.Where{ResultTableExpr: lhs, Expr: expr}, tokens
}

func parseSelect(tokens []lexer.Token, lhs lexer.Token) (ast.Select, []lexer.Token) {
	project := ast.Select{ResultTableExpr: lhs}
	for {
		var token lexer.Token
		token, tokens = lexer.GetNextToken(tokens)
		// TODO handle more than one field
		project.Fields = append(project.Fields, token)
		if lexer.IsSemicolon(token.Representation[0]) {
			break
		}
	}
	return project, tokens
}

func Parse(text string) (ast.AST, int8) {
	var resultAst ast.AST
	tokens := lexer.GetTokens(lexer.StringToToken(text))

	for len(tokens) > 0 {
		var token lexer.Token
		token, tokens = lexer.GetNextToken(tokens)

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
			var from ast.From
			from, tokens = parseFrom(tokens, lhs)
			resultAst = append(resultAst, ast.Statement{Type: ast.StatementTypeFrom, From: from})
			continue
		}

		if lexer.IsWhere(token) {
			var where ast.Where
			where, tokens = parseWhere(tokens, lhs)
			resultAst = append(resultAst, ast.Statement{Type: ast.StatementTypeWhere, Where: where})
			continue
		}

		if lexer.IsSelect(token) {
			var project ast.Select
			project, tokens = parseSelect(tokens, lhs)
			resultAst = append(resultAst, ast.Statement{Type: ast.StatementTypeSelect, Select: project})
			token, tokens = lexer.GetNextToken(tokens)
			continue
		}
	}
	return resultAst, 0
}

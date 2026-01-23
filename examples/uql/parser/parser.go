package parser

import (
	"uql/ast"
	"uql/lexer"
)

func precedence(op []int8) int8 {
	if op[0] == '&' && op[1] == '&' {
		return 1
	}
	if op[0] == '|' && op[1] == '|' {
		return 1
	}
	if op[0] == '>' {
		return 2
	}
	if op[0] == '<' {
		return 2
	}
	if op[0] == '>' && op[1] == '=' {
		return 2
	}
	if op[0] == '<' && op[1] == '=' {
		return 2
	}
	if op[0] == '=' && op[1] == '=' {
		return 2
	}
	if op[0] == '!' && op[1] == '=' {
		return 2
	}
	return -1
}

func associativity(op []int8) int8 {
	if op[0] == '&' && op[1] == '&' {
		return 'L'
	}
	if op[0] == '|' && op[1] == '|' {
		return 'L'
	}
	if op[0] == '>' {
		return 'L'
	}
	if op[0] == '<' {
		return 'L'
	}
	if op[0] == '>' && op[1] == '=' {
		return 'L'
	}
	if op[0] == '<' && op[1] == '=' {
		return 'L'
	}
	if op[0] == '=' && op[1] == '=' {
		return 'L'
	}
	if op[0] == '!' && op[1] == '=' {
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
	return ast.Where{Expr: expr, ResultTableExpr: lhs}, tokens
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

func parseJoin(tokens []lexer.Token, lhs lexer.Token) (ast.Join, []lexer.Token) {
	join := ast.Join{ResultTableExpr: lhs}
	var token lexer.Token

	// Get left table
	token, tokens = lexer.GetNextToken(tokens)
	join.LeftTable = token

	// Get right table
	token, tokens = lexer.GetNextToken(tokens)
	join.RightTable = token

	// Expect 'on' keyword
	token, tokens = lexer.GetNextToken(tokens)
	if !lexer.IsOn(token) {
		return join, tokens
	}

	// Parse the ON condition
	expr, i := ParseExpression(tokens)
	join.OnCondition = expr
	tokens = tokens[i:]

	// Skip to semicolon
	for {
		token, tokens = lexer.GetNextToken(tokens)
		if lexer.IsSemicolon(token.Representation[0]) {
			break
		}
	}
	return join, tokens
}

func parseOrderBy(tokens []lexer.Token, lhs lexer.Token) (ast.OrderBy, []lexer.Token) {
	orderBy := ast.OrderBy{ResultTableExpr: lhs}
	var token lexer.Token

	// Get source table
	token, tokens = lexer.GetNextToken(tokens)
	orderBy.SourceTable = token

	// Parse fields (skip asc/desc keywords for now)
	for {
		token, tokens = lexer.GetNextToken(tokens)
		if lexer.IsSemicolon(token.Representation[0]) {
			break
		}

		// Skip asc/desc keywords
		if lexer.IsAsc(token) || lexer.IsDesc(token) {
			continue
		}

		orderBy.Fields = append(orderBy.Fields, token)
	}
	return orderBy, tokens
}

func parseLimit(tokens []lexer.Token, lhs lexer.Token) (ast.Limit, []lexer.Token) {
	limit := ast.Limit{ResultTableExpr: lhs}
	var token lexer.Token

	// Get source table
	token, tokens = lexer.GetNextToken(tokens)
	limit.SourceTable = token

	// Get limit count
	token, tokens = lexer.GetNextToken(tokens)
	limit.Count = token

	// Skip to semicolon
	for {
		token, tokens = lexer.GetNextToken(tokens)
		if lexer.IsSemicolon(token.Representation[0]) {
			break
		}
	}
	return limit, tokens
}

func parseGroupBy(tokens []lexer.Token, lhs lexer.Token) (ast.GroupBy, []lexer.Token) {
	groupBy := ast.GroupBy{ResultTableExpr: lhs}
	var token lexer.Token

	// Get source table
	token, tokens = lexer.GetNextToken(tokens)
	groupBy.SourceTable = token

	// Parse fields
	for {
		token, tokens = lexer.GetNextToken(tokens)
		if lexer.IsSemicolon(token.Representation[0]) {
			break
		}

		// Check for aggregate functions
		if lexer.IsCount(token) || lexer.IsSum(token) || lexer.IsAvg(token) ||
			lexer.IsMin(token) || lexer.IsMax(token) {
			agg := ast.Aggregate{Function: token}
			// Get field for aggregate
			token, tokens = lexer.GetNextToken(tokens)
			agg.Field = token
			groupBy.Aggregates = append(groupBy.Aggregates, agg)
			continue
		}

		groupBy.Fields = append(groupBy.Fields, token)
	}
	return groupBy, tokens
}

func Parse(text string) (ast.AST, int8) {
	var resultAst ast.AST
	tokens := lexer.GetTokens(lexer.StringToToken(text))

	for len(tokens) > 0 {
		var token lexer.Token
		token, tokens = lexer.GetNextToken(tokens)

		if !lexer.IsAlpha(token.Representation[0]) {
			return ast.AST{}, -1
		}
		lhs := token
		token, tokens = lexer.GetNextToken(tokens)
		if !lexer.IsEqual(token.Representation[0]) {
			return ast.AST{}, -1
		}

		token, tokens = lexer.GetNextToken(tokens)
		if !lexer.IsFrom(token) && !lexer.IsWhere(token) && !lexer.IsSelect(token) &&
			!lexer.IsJoin(token) && !lexer.IsOrderBy(token) && !lexer.IsLimit(token) &&
			!lexer.IsGroupBy(token) {
			return ast.AST{}, -1
		}

		if lexer.IsFrom(token) {
			var from ast.From
			from, tokens = parseFrom(tokens, lhs)
			resultAst = append(resultAst, ast.Statement{Type: int8(ast.StatementTypeFrom), FromF: from})
			continue
		}

		if lexer.IsWhere(token) {
			var where ast.Where
			where, tokens = parseWhere(tokens, lhs)
			resultAst = append(resultAst, ast.Statement{Type: int8(ast.StatementTypeWhere), WhereF: where})
			continue
		}

		if lexer.IsSelect(token) {
			var project ast.Select
			project, tokens = parseSelect(tokens, lhs)
			resultAst = append(resultAst, ast.Statement{Type: int8(ast.StatementTypeSelect), SelectF: project})
			token, tokens = lexer.GetNextToken(tokens)
			continue
		}

		if lexer.IsJoin(token) {
			var join ast.Join
			join, tokens = parseJoin(tokens, lhs)
			resultAst = append(resultAst, ast.Statement{Type: int8(ast.StatementTypeJoin), JoinF: join})
			continue
		}

		if lexer.IsOrderBy(token) {
			var orderBy ast.OrderBy
			orderBy, tokens = parseOrderBy(tokens, lhs)
			resultAst = append(resultAst, ast.Statement{Type: int8(ast.StatementTypeOrderBy), OrderByF: orderBy})
			continue
		}

		if lexer.IsLimit(token) {
			var limit ast.Limit
			limit, tokens = parseLimit(tokens, lhs)
			resultAst = append(resultAst, ast.Statement{Type: int8(ast.StatementTypeLimit), LimitF: limit})
			continue
		}

		if lexer.IsGroupBy(token) {
			var groupBy ast.GroupBy
			groupBy, tokens = parseGroupBy(tokens, lhs)
			resultAst = append(resultAst, ast.Statement{Type: int8(ast.StatementTypeGroupBy), GroupByF: groupBy})
			continue
		}
	}
	return resultAst, 0
}

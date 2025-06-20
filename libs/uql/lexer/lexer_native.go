package lexer

// That's helper function to convert string to lexer.Token
func StringToToken(s string) Token {
	var token Token
	for _, r := range s {
		token.Representation = append(token.Representation, int8(r))
	}
	return token
}

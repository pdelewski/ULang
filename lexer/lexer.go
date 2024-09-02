package lexer

import "fmt"

const (
	TokenTypeIdentifier = iota
	TokenTypeOperator
	TokenTypeNumber
	TokenTypeWhitespace
)

type Token struct {
	Type           int
	Representation []int8
}

func IsDigit(b int8) bool {
	return b >= '0' && b <= '9'
}

func DumpTokens(tokens []Token) {
	for _, token := range tokens {
		for _, b := range token.Representation {
			if b == ' ' {
				fmt.Print(" ")
			} else if b == '\t' {
				fmt.Print("\t")
			} else if b == '\n' {
				fmt.Println()
			} else {
				fmt.Printf("%c", b)
			}
		}
	}
}

func GetTokens(token Token) []Token {
	var tokens []Token
	var currentToken Token

	for _, b := range token.Representation {
		if b == ' ' || b == '\t' || b == '\n' { // If the character is a space, tab, or newline
			if len(currentToken.Representation) > 0 {
				tokens = append(tokens, currentToken)
				currentToken.Representation = nil
			}
			tokens = append(tokens, Token{Representation: []int8{b}}) // Add the whitespace character as a separate token
		} else {
			currentToken.Representation = append(currentToken.Representation, b)
		}
	}

	// Add the last token if it exists
	if len(currentToken.Representation) > 0 {
		tokens = append(tokens, currentToken)
	}

	return tokens
}

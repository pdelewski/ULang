package lexer

import "fmt"

type Token = []int8

func IsDigit(b int8) bool {
	return b >= '0' && b <= '9'
}

func DumpTokens(tokens []Token) {
	for _, token := range tokens {
		for _, b := range token {
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

func GetTokens(buf []int8) []Token {
	var tokens []Token
	var currentToken Token

	for _, b := range buf {
		if b == ' ' || b == '\t' || b == '\n' { // If the character is a space, tab, or newline
			if len(currentToken) > 0 {
				tokens = append(tokens, currentToken)
				currentToken = nil
			}
			tokens = append(tokens, Token{b}) // Add the whitespace character as a separate token
		} else {
			currentToken = append(currentToken, b)
		}
	}

	// Add the last token if it exists
	if len(currentToken) > 0 {
		tokens = append(tokens, currentToken)
	}

	return tokens
}

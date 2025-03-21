package lexer

import (
	"fmt"
)

const (
	TokenTypeIdentifier = 1
	TokenTypeOperator   = 2
	TokenTypeNumber     = 3
	TokenTypeWhitespace = 4
	TokenTypeDot        = 5 // Added for the dot operator
	TokenTypeSemicolon  = 6
)

type Token struct {
	Type           int8
	Representation []int8
}

func IsDigit(b int8) bool {
	return b >= '0' && b <= '9'
}

func IsAlpha(b int8) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z' || b == '_')
}

func IsEqual(b int8) bool {
	return b == '='
}

func IsSemicolon(b int8) bool {
	return b == ';'
}

func IsFrom(token Token) bool {
	return len(token.Representation) == 4 &&
		token.Representation[0] == 'f' &&
		token.Representation[1] == 'r' &&
		token.Representation[2] == 'o' &&
		token.Representation[3] == 'm'
}

func IsSelect(token Token) bool {
	return len(token.Representation) == 6 &&
		token.Representation[0] == 's' &&
		token.Representation[1] == 'e' &&
		token.Representation[2] == 'l' &&
		token.Representation[3] == 'e' &&
		token.Representation[4] == 'c' &&
		token.Representation[5] == 't'
}

func IsWhere(token Token) bool {
	return len(token.Representation) == 5 &&
		token.Representation[0] == 'w' &&
		token.Representation[1] == 'h' &&
		token.Representation[2] == 'e' &&
		token.Representation[3] == 'r' &&
		token.Representation[4] == 'e'
}

func DumpToken(token Token) {
	fmt.Printf("Token type: %d ", token.Type)
	for _, b := range token.Representation {
		if b == ' ' {
			fmt.Print(" ")
		} else if b == '\t' {
			fmt.Print("\t")
		} else if b == '\n' {
			fmt.Print("\n")
		} else if b == '.' {
			fmt.Print(".")
		} else {
			fmt.Printf("%c", b)
		}
	}
	fmt.Println()
}

func DumpTokens(tokens []Token) {
	for _, token := range tokens {
		DumpToken(token)
	}
}

func DumpTokenString(token Token) string {
	var result string

	// Append token type to the string
	result += fmt.Sprintf("Token type: %d ", token.Type)

	// Append representation to the string
	for _, b := range token.Representation {
		switch b {
		case ' ':
			result += " "
		case '\t':
			result += "\t"
		case '\n':
			result += "\n"
		case '.':
			result += "."
		default:
			result += fmt.Sprintf("%c", b)
		}
	}

	result += "\n"
	// Return the constructed string
	return result
}

func DumpTokensString(tokens []Token) string {
	var result string
	for _, token := range tokens {
		result += fmt.Sprintf("Token type: %d ", token.Type)
		// Append representation to the string
		for _, b := range token.Representation {
			switch b {
			case ' ':
				result += " "
			case '\t':
				result += "\t"
			case '\n':
				result += "\n"
			case '.':
				result += "."
			default:
				result += fmt.Sprintf("%c", b)
			}
		}

		result += "\n"
	}

	return result
}

func GetTokens(token Token) []Token {
	var tokens []Token
	var currentToken Token

	for _, b := range token.Representation {
		if b == ';' {
			if len(currentToken.Representation) > 0 {
				tokens = append(tokens, currentToken)
				currentToken.Representation = nil
			}
			tokens = append(tokens, Token{Type: TokenTypeSemicolon, Representation: []int8{b}})
			continue
		}
		if b == ' ' || b == '\t' || b == '\n' /*|| b == '.'*/ { // If the character is a space, tab, or newline
			if len(currentToken.Representation) > 0 {
				tokens = append(tokens, currentToken)
				currentToken.Representation = nil
			}
			//tokens = append(tokens, Token{Type: TokenTypeWhitespace, Representation: []int8{b}}) // Add the whitespace character as a separate token
		} else {
			// TODO build correct token type
			currentToken.Type = TokenTypeIdentifier
			currentToken.Representation = append(currentToken.Representation, b)
		}
	}

	// Add the last token if it exists
	if len(currentToken.Representation) > 0 {
		tokens = append(tokens, currentToken)
	}

	return tokens
}

func GetNextToken(tokens []Token) (Token, []Token) {
	if len(tokens) == 0 {
		return Token{}, []Token{}
	}
	return tokens[0], tokens[1:]
}

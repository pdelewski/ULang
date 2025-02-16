package lexer

import "fmt"

// Token types as constants
const (
	TokenLetter           = 0
	TokenDigit            = 1
	TokenSpace            = 2
	TokenSymbol           = 3
	TokenLeftParenthesis  = 4
	TokenRightParenthesis = 5
	TokenPipe             = 6
	TokenGreater          = 7
	TokenLess             = 8
)

func IsLetter(b int8) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func IsAlphaNumeric(b int8) bool {
	return IsLetter(b) || IsDigit(b) || b == '_'
}

func IsSpace(b int8) bool {
	return b == ' '
}

func IsLeftParenthesis(b int8) bool {
	return b == '('
}

func IsRightParenthesis(b int8) bool {
	return b == ')'
}

func IsPipe(b int8) bool {
	return b == '|'
}

func IsGreater(b int8) bool {
	return b == '>'
}

func IsLess(b int8) bool {
	return b == '<'
}

// Tokenize splits the input text into categorized tokens.
func Tokenize(text string) []Token {
	var tokens []Token
	var currentToken []int8
	var currentType int8 = -1

	// Helper function to add a token to the list
	addToken := func() {
		if len(currentToken) > 0 {
			tokens = append(tokens, Token{
				Type:           currentType,
				Representation: currentToken,
			})
			currentToken = nil
		}
	}

	// Iterate through the input string
	for i := 0; i < len(text); i++ {
		c := int8(text[i])
		var tokenType int8

		// Determine the type of character
		if IsAlphaNumeric(c) {
			tokenType = TokenLetter
		} else if IsDigit(c) {
			tokenType = TokenDigit
		} else if IsSpace(c) {
			tokenType = TokenSpace
		} else if IsLeftParenthesis(c) {
			tokenType = TokenLeftParenthesis
		} else if IsRightParenthesis(c) {
			tokenType = TokenRightParenthesis
		} else if IsPipe(c) {
			tokenType = TokenPipe
		} else if IsGreater(c) {
			tokenType = TokenGreater
		} else if IsLess(c) {
			tokenType = TokenLess
		} else {
			tokenType = TokenSymbol
		}

		// If the type changes, finalize the previous token
		if tokenType != currentType {
			addToken()
			currentType = tokenType
		}

		// Append character to current token as int8
		currentToken = append(currentToken, c)
	}

	// Add the last token if any
	addToken()

	return tokens
}

func TokenizeTest() {
	tokens1 := Tokenize("Select * from table1 where field1 > 10;")
	for _, token := range tokens1 {
		fmt.Println(DumpTokenString(token))
	}
	tokens2 := Tokenize("(Select * from table1 where field1 > 10)")
	for _, token := range tokens2 {
		fmt.Println(DumpTokenString(token))
	}
}

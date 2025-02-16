package lexer

// Token types as constants
const (
	TokenLetter = 0
	TokenDigit  = 1
	TokenSpace  = 2
	TokenSymbol = 3
)

func IsLetter(b int8) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func IsSpace(b int8) bool {
	return b == ' '
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
		if IsLetter(c) {
			tokenType = TokenLetter
		} else if IsDigit(c) {
			tokenType = TokenDigit
		} else if IsSpace(c) {
			tokenType = TokenSpace
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

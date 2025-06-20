package lexer

import "testing"
import "github.com/stretchr/testify/assert"

func TestIsDigit(t *testing.T) {
	digits := []int8{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}
	nonDigits := []int8{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'}
	for _, d := range digits {
		assert.Equal(t, true, IsDigit(d))
	}
	for _, d := range nonDigits {
		assert.Equal(t, false, IsDigit(d))
	}
}

func TestGetTokens(t *testing.T) {
	token := Token{Representation: []int8{'a', 'b', ' ', 'c', ' ', 'd'}}

	tokens := GetTokens(token)
	assert.Equal(t, 5, len(tokens))
	assert.Equal(t, []int8{'a', 'b'}, tokens[0].Representation)
	assert.Equal(t, []int8{'c'}, tokens[2].Representation)
	assert.Equal(t, []int8{'d'}, tokens[4].Representation)
	DumpTokens(tokens)
}

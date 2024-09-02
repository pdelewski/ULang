package parser

import (
	"ULang/lexer"
	uqllexer "ULang/uql/lexer"
)

func Parse(text string) []lexer.Token {
	tokens := lexer.GetTokens(uqllexer.StringToToken(text))
	lexer.DumpTokens(tokens)
	return tokens
}

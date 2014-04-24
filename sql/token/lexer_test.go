package token

import (
	"testing"
)

func TestQuotedId(t *testing.T) {
	lex := NewLexer("`abc`")

	assertNextToken(t, 0, ID, "`abc`", lex)
	assertEnd(t, lex)
}

func TestQuotedIdNegative(t *testing.T) {
	lex := NewLexer("`0bc`")

	pos, tok, _ := lex.Next()
	if tok != ILLEGAL || pos != 2 {
		t.Fatal(pos, tok)
	}
	t.Log(lex.Error())
}

func TestDequotedId(t *testing.T) {
	lex := NewLexer("abc def")

	assertNextToken(t, 0, ID, "abc", lex)
	assertNextToken(t, 4, ID, "def", lex)
	assertEnd(t, lex)
}

func TestKeyword(t *testing.T) {
	lex := NewLexer("SELECT From WhErE")

	assertNextToken(t, 0, SELECT, "SELECT", lex)
	assertNextToken(t, 7, FROM, "From", lex)
	assertNextToken(t, 12, WHERE, "WhErE", lex)
	assertEnd(t, lex)
}

func TestQuotedKeyword(t *testing.T) {
	lex := NewLexer("`SELECT`\n`From`\r\n`WhErE`")

	assertNextToken(t, 0, ID, "`SELECT`", lex)
	assertNextToken(t, 9, ID, "`From`", lex)
	assertNextToken(t, 17, ID, "`WhErE`", lex)
	assertEnd(t, lex)
}

func TestInt(t *testing.T) {
	lex := NewLexer("190 255 31415926 65535")

	assertNextToken(t, 0, INT_LITERAL, "190", lex)
	assertNextToken(t, 4, INT_LITERAL, "255", lex)
	assertNextToken(t, 8, INT_LITERAL, "31415926", lex)
	assertNextToken(t, 17, INT_LITERAL, "65535", lex)
	assertEnd(t, lex)
}

func TestIntNegative(t *testing.T) {
	lex := NewLexer("190a")

	pos, tok, _ := lex.Next()
	if tok != ILLEGAL || pos != 3 {
		t.Fatal(pos, tok)
	}
	t.Log(lex.Error())
}

func TestComparison(t *testing.T) {
	lex := NewLexer("> >= < <= <>")

	assertNextToken(t, 0, GT, ">", lex)
	assertNextToken(t, 2, GE, ">=", lex)
	assertNextToken(t, 5, LT, "<", lex)
	assertNextToken(t, 7, LE, "<=", lex)
	assertNextToken(t, 10, NE, "<>", lex)
	assertEnd(t, lex)
}

func TestEatSpaceRunes(t *testing.T) {
	lex := NewLexer(" `abc` ")

	assertNextToken(t, 1, ID, "`abc`", lex)
	assertEnd(t, lex)
}

func TestOperator(t *testing.T) {
	lex := NewLexer("+ - * / () . , ;")

	assertNextToken(t, 0, PLUS, "+", lex)
	assertNextToken(t, 2, MINUS, "-", lex)
	assertNextToken(t, 4, STAR, "*", lex)
	assertNextToken(t, 6, SLASH, "/", lex)
	assertNextToken(t, 8, LPAREN, "(", lex)
	assertNextToken(t, 9, RPAREN, ")", lex)
	assertNextToken(t, 11, DOT, ".", lex)
	assertNextToken(t, 13, COMMA, ",", lex)
	assertNextToken(t, 15, SEMI, ";", lex)
	assertEnd(t, lex)
}

func TestDotExpr(t *testing.T) {
	lex := NewLexer("db.name")

	assertNextToken(t, 0, ID, "db", lex)
	assertNextToken(t, 2, DOT, ".", lex)
	assertNextToken(t, 3, ID, "name", lex)
	assertEnd(t, lex)
}

func TestParenExpr(t *testing.T) {
	lex := NewLexer("(1 + 2) * id")

	assertNextToken(t, 0, LPAREN, "(", lex)
	assertNextToken(t, 1, INT_LITERAL, "1", lex)
	assertNextToken(t, 3, PLUS, "+", lex)
	assertNextToken(t, 5, INT_LITERAL, "2", lex)
	assertNextToken(t, 6, RPAREN, ")", lex)
	assertNextToken(t, 8, STAR, "*", lex)
	assertNextToken(t, 10, ID, "id", lex)
	assertEnd(t, lex)
}

func TestStringLiteral(t *testing.T) {
	lex := NewLexer(`'literal'"Hello, World" "yes"`)

	assertNextToken(t, 0, STRING_LITERAL, `'literal'`, lex)
	assertNextToken(t, 9, STRING_LITERAL, `"Hello, World"`, lex)
	assertNextToken(t, 24, STRING_LITERAL, `"yes"`, lex)
	assertEnd(t, lex)
}

func assertEnd(t *testing.T, lex *Lexer) {
	if _, tok, _ := lex.Next(); tok != EOF {
		t.Fatal("Not lexer end! ", tok)
	}
}

func assertNextToken(t *testing.T, pos int, tok Token, lit string, lex *Lexer) {
	l_pos, l_tok, l_lit := lex.Next()
	if pos != l_pos {
		t.Fatal("Unexpected pos: ", lex.Error(), pos, l_pos)
	}
	if tok != l_tok {
		t.Fatal("Unexpected tok("+lit+"): ", lex.Error(), tok, l_tok)
	}
	if lit != l_lit {
		t.Fatal("Unexpected lit: ", lex.Error(), lit, l_lit)
	}
}

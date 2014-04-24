package ast

import (
	"encoding/json"
	"testing"
	"yui/sql/token"
)

func TestLiteral(t *testing.T) {
	var node Node = &Literal{
		ValuePos: 0,
		Value:    "3.14",
		Kind:     token.FLOAT_LITERAL,
	}
	if node.Pos() != 0 {
		t.Fatal("fail")
	}
	if node.End() != 4 {
		t.Fatal("fail")
	}
}

func TestQuotedIdentifer(t *testing.T) {
	id := &Identifier{NamePos: 0, Name: "`name`"}
	if id.Pos() != 0 {
		t.Fatal("fail")
	}
	if id.End() != 6 {
		t.Fatal("fail")
	}
	if has_quote, quote := id.Quote(); !has_quote || quote != "`" {
		t.Fatal("fail")
	}
	if has_quote, name := id.Dequote(); !has_quote || name != "name" {
		t.Fatal("fail")
	}
}

func TestDequotedIdentifier(t *testing.T) {
	id := &Identifier{NamePos: 0, Name: "name"}
	if id.Pos() != 0 {
		t.Fatal("fail")
	}
	if id.End() != 4 {
		t.Fatal("fail")
	}
	if has_quote, _ := id.Quote(); has_quote {
		t.Fatal("fail")
	}
	if has_quote, _ := id.Dequote(); has_quote {
		t.Fatal("fail")
	}
}

func TestUnaryExpr(t *testing.T) {
	// 3.14 IS NULL
	expr := &UnaryExpr{
		OpPos:   8,
		Op:      token.IS_NULL,
		Operand: nil,
	}
	expr.Operand = &Literal{
		ValuePos: 0,
		Value:    "3.14",
		Kind:     token.FLOAT_LITERAL,
	}
	if expr.Operand == nil {
		t.Fatal("fail")
	}
}

func TestWhereInExpr(t *testing.T) {
	// id IN (1, 2)
	expr := &BinaryExpr{
		OpPos: 0,
		Op:    token.IN,
		Lhs: &Literal{
			ValuePos: 3,
			Value:    "id",
			Kind:     token.ID,
		},
	}
	list := make([]Expr, 2)
	list = append(list, &Literal{
		ValuePos: 7,
		Value:    "1",
		Kind:     token.INT_LITERAL,
	}, &Literal{
		ValuePos: 9,
		Value:    "2",
		Kind:     token.INT_LITERAL,
	})
	expr.Rhs = ExprList(list)

	if _, ok := expr.Lhs.(*Literal); !ok {
		t.Fatal("fail")
	}
	if _, ok := expr.Rhs.(ExprList); !ok {
		t.Fatal("fail")
	}
}

func TestAstDump(t *testing.T) {
	expr := &UnaryExpr{
		OpPos:   8,
		Op:      token.IS_NULL,
		Operand: nil,
	}
	expr.Operand = &Literal{
		ValuePos: 0,
		Value:    "3.14",
		Kind:     token.FLOAT_LITERAL,
	}

	out, err := json.Marshal(expr)
	if err != nil {
		t.Fatal(err)
	}
	// if string(out) != `{"OpPos":8,"Op":53,"Operand":{"ValuePos":0,"Value":"3.14","Kind":48}}` {
	// 	t.Fatal("Bad dump")
	// }
	t.Log(string(out))
}

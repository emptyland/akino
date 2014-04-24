package ast

import (
	"strings"

	"github.com/emptyland/akino/sql/token"
)

//------------------------------------------------------------------------------
type ExprList []Expr

func (self ExprList) Pos() int {
	return self[0].Pos()
}

func (self ExprList) End() int {
	return self[len(self)-1].End()
}

//------------------------------------------------------------------------------
type Identifier struct {
	NamePos int
	Name    string
}

func (self *Identifier) Pos() int {
	return self.NamePos
}

func (self *Identifier) End() int {
	return self.Pos() + len(self.Name)
}

func (self *Identifier) Quote() (bool, string) {
	if len(self.Name) > 2 {
		return strings.HasPrefix(self.Name, "`") &&
			strings.HasSuffix(self.Name, "`"), "`"
	} else {
		return false, ""
	}
}

func (self *Identifier) Dequote() (bool, string) {
	if has_quote, quote := self.Quote(); has_quote {
		return true, strings.Trim(self.Name, quote)
	} else {
		return false, self.Name
	}
}

//------------------------------------------------------------------------------
type Literal struct {
	ValuePos int
	Value    string
	Kind     token.Token
}

func (self *Literal) Pos() int {
	return self.ValuePos
}

func (self *Literal) End() int {
	return self.Pos() + len(self.Value)
}

//------------------------------------------------------------------------------
type UnaryExpr struct {
	OpPos   int
	Op      token.Token
	Operand Expr
}

func (self *UnaryExpr) Pos() int {
	return self.OpPos
}

func (self *UnaryExpr) End() int {
	return self.Operand.End()
}

//------------------------------------------------------------------------------
type BinaryExpr struct {
	OpPos int
	Op    token.Token
	Lhs   Expr
	Rhs   Expr
}

func (self *BinaryExpr) Pos() int {
	return self.OpPos
}

func (self *BinaryExpr) End() int {
	return self.Rhs.End()
}

//------------------------------------------------------------------------------
type CallExpr struct {
	Func     Identifier
	Args     []Expr
	Distinct bool
}

func (self *CallExpr) Pos() int {
	return self.Func.Pos()
}

func (self *CallExpr) End() int {
	return self.Args[len(self.Args)-1].End()
}

//------------------------------------------------------------------------------
type Condition struct {
	OpPos  int
	Case   Expr
	Blocks []ConditionBlock
	Else   Expr
}

func (self *Condition) Pos() int {
	return self.OpPos
}

func (self *Condition) End() int {
	if self.Else == nil {
		return self.Blocks[len(self.Blocks)-1].Then.End()
	} else {
		return self.Else.End()
	}
}

type ConditionBlock struct {
	When Expr
	Then Expr
}

//------------------------------------------------------------------------------
type CastExpr struct {
	OpPos   int
	Operand Expr
	To      Type
}

func (self *CastExpr) Pos() int {
	return self.OpPos
}

func (self *CastExpr) End() int {
	return self.To.End()
}

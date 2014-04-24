package ast

import (
	"log"
	"strings"

	"github.com/emptyland/akino/sql/token"
)

type Node interface {
	Pos() int
	End() int
}

type Command interface {
	Node
}

type Expr interface {
	Node
}

type NameRef struct {
	First  string
	Second string
}

func (self *NameRef) Table() string {
	if self.Second == "" {
		return self.First
	} else {
		return self.Second
	}
}

func (self *NameRef) Database() string {
	if self.Second == "" {
		return ""
	} else {
		return self.First
	}
}

func (self *NameRef) Full() string {
	if self.Second == "" {
		return self.First
	} else {
		return self.First + "." + self.Second
	}
}

//------------------------------------------------------------------------------
type Select struct {
	SelectPos  int
	SelectEnd  int
	Op         token.Token
	Prior      *Select
	Distinct   bool
	Limit      Expr
	Offset     Expr
	SelColList []SelectColumn
	From       []Source
	Where      Expr
	Having     Expr
	GroupBy    []Expr
	OrderBy    []OrderByItem
}

func (self *Select) Pos() int {
	return self.SelectPos
}

func (self *Select) End() int {
	return self.SelectEnd
}

type SelectColumn struct {
	SelectExpr Expr
	Alias      string
}

type OrderByItem struct {
	Item Expr
	Desc bool
}

//------------------------------------------------------------------------------
type Source struct {
	SourcePos int
	SourceEnd int
	JoinType  int
	Table     *NameRef
	Subquery  *Select
	Alias     string
	Indexed   string
	On        Expr
	Using     []Identifier
}

func (self *Source) Pos() int {
	return self.SourcePos
}

func (self *Source) End() int {
	return self.SourceEnd
}

func (self *Source) IsSubquery() (bool, *Select) {
	return self.Subquery != nil, self.Subquery
}

func (self *Source) IsTable() (bool, *NameRef) {
	return self.Table != nil, self.Table
}

const (
	JT_INNER = (1 << iota)
	JT_CROSS
	JT_NATURAL
	JT_LEFT
	JT_RIGHT
	JT_OUTER
)

//------------------------------------------------------------------------------
type Transaction struct {
	TransactionPos int
	Op             token.Token // token.BEGIN | START | COMMIT | ROLLBACK
	Type           token.Token // token.DEFERREF | IMMEDIATE | EXCLUSIVE
}

func (self *Transaction) Pos() int {
	return self.TransactionPos
}

func (self *Transaction) End() int {
	return self.Pos()
}

//------------------------------------------------------------------------------
type Show struct {
	ShowPos int
	Dest    token.Token // token.TABLES | DATABASES
}

func (self *Show) Pos() int {
	return self.ShowPos
}

func (self *Show) End() int {
	return self.Pos()
}

//------------------------------------------------------------------------------
type Comment struct {
	CommentPos int
	Text       string
}

func (self *Comment) Pos() int {
	return self.CommentPos
}

func (self *Comment) End() int {
	return self.Pos() + len(self.Text)
}

func (self *Comment) Block() bool {
	switch {
	case strings.HasPrefix(self.Text, "--") && strings.HasSuffix(self.Text, "--"):
		return true

	case strings.HasPrefix(self.Text, "/*") && strings.HasSuffix(self.Text, "*/"):
		return false

	default:
		log.Fatal("Bad comment prefix and suffix!")
		panic("fatal")
	}
}

func (self *Comment) Content() string {
	if self.Block() {
		return strings.Trim(self.Text, "--")
	} else {
		return "" // FIXME
	}
}

//------------------------------------------------------------------------------
type CreateTable struct {
	CreatePos       int
	CreateEnd       int
	Temp            bool
	IfNotExists     bool
	Table           NameRef
	Scheme          []ColumnDefine
	Template        *Select
	CheckConstraint []Expr
}

func (self *CreateTable) Pos() int {
	return self.CreatePos
}

func (self *CreateTable) End() int {
	return self.CreateEnd
}

/*
 * On Conf:
 *	token.IGNORE
 *	token.DEFAULT
 *	token.REPLACE
 *	token.ROLLBACK
 *	token.ABORT
 *	token.FAIL
 */
type ColumnDefine struct {
	Name           string
	ColumnType     Type
	Default        Expr
	NotNull        bool
	NotNullOn      token.Token
	PrimaryKey     bool
	PrimaryKeyOn   token.Token
	PrimaryKeyDesc bool
	Unique         bool
	UniqueOn       token.Token
	AutoIncr       bool
	Collate        string
}

//------------------------------------------------------------------------------
type CreateIndex struct {
	CreatePos   int
	CreateEnd   int
	Unique      bool
	IfNotExists bool
	Name        NameRef // Index name
	Table       string  // For table name
	Index       []IndexDefine
}

func (self *CreateIndex) Pos() int {
	return self.CreatePos
}

func (self *CreateIndex) End() int {
	return self.CreateEnd
}

type IndexDefine struct {
	Name    string
	Collate string
	Desc    bool
}

//------------------------------------------------------------------------------
type Insert struct {
	InsertPos int
	InsertEnd int
	Op        token.Token
	Dest      NameRef
	Column    []Identifier
	Item      []Expr
	From      *Select
}

func (self *Insert) Pos() int {
	return self.InsertPos
}

func (self *Insert) End() int {
	return self.InsertEnd
}

func (self *Insert) DefaultValues() bool {
	return len(self.Item) == 0 && self.From == nil
}

//------------------------------------------------------------------------------
type Update struct {
	UpdatePos int
	UpdateEnd int
	Op        token.Token
	Dest      NameRef
	Indexed   string
	Set       []SetDefine
	Where     Expr
	OrderBy   []OrderByItem
	Limit     Expr
	Offset    Expr
}

func (self *Update) Pos() int {
	return self.UpdatePos
}

func (self *Update) End() int {
	return self.UpdateEnd
}

type SetDefine struct {
	Column string
	Value  Expr
}

//------------------------------------------------------------------------------
type Delete struct {
	DeletePos int
	DeleteEnd int
	Dest      NameRef
	Indexed   string
	Where     Expr
	OrderBy   []OrderByItem
	Limit     Expr
	Offset    Expr
}

func (self *Delete) Pos() int {
	return self.DeletePos
}

func (self *Delete) End() int {
	return self.DeleteEnd
}

//------------------------------------------------------------------------------
type Type struct {
	TokenPos int
	Kind     token.Token
	Width    *Literal
	Decimal  *Literal
	Unsigned bool
}

func (self *Type) Pos() int {
	return self.TokenPos
}

func (self *Type) End() int {
	return self.Pos()
}

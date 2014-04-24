package parser

import (
	"fmt"
	"strings"

	"github.com/emptyland/akino/sql/ast"
	"github.com/emptyland/akino/sql/token"
)

func ParseCommand(cmd string) (ast.Command, error) {
	var p Parser
	return p.Init(cmd).NextStatement()
}

func ParseExpression(expr string) (ast.Expr, error) {
	var p Parser
	return p.Init(expr).NextExpr()
}

type Parser struct {
	cmd string
	lah tokeniton // look a head
	lex *token.Lexer
}

func (self *Parser) Init(cmd string) *Parser {
	self.cmd = cmd
	self.lex = token.NewLexer(cmd)
	self.skip()
	return self
}

func (self *Parser) NextStatement() (ast.Command, error) {
	var cmd ast.Command
	var err error
	if cmd, err = self.Next(); err != nil {
		return nil, err
	}
	if self.peek() != token.EOF {
		if _, err = self.match(token.SEMI); err != nil {
			return nil, err
		}
	}
	return cmd, nil
}

func (self *Parser) Next() (ast.Command, error) {

	switch self.peek() {
	case token.BEGIN, token.START, token.COMMIT, token.ROLLBACK, token.END:
		return self.parseTransaction(self.peek())

	case token.ILLEGAL:
		return nil, self.errorf("Token") // TODO

	case token.SHOW:
		return self.parseShow()

	case token.SELECT:
		return self.parseSelect()

	case token.CREATE:
		return self.parseCreate()

	case token.INSERT, token.REPLACE:
		return self.parseInsert()

	case token.UPDATE:
		return self.parseUpdate()

	case token.DELETE:
		return self.parseDelete()

	default:
		return nil, self.errorf(`Unknown command: "%s"`, self.peekLiteral())
	}
}

func (self *Parser) parseTransaction(op token.Token) (ast.Command, error) {
	cmd := &ast.Transaction{
		TransactionPos: self.peekPos(),
		Op:             op,
		Type:           token.ILLEGAL,
	}
	self.skip()

	// Parse transaction type
	if cmd.Op == token.BEGIN || cmd.Op == token.START {
		cmd.Type = self.parseTransactionType()
	}

	// Parse transaction option
	switch self.peek() {
	case token.EOF:
		return cmd, nil

	case token.TRANSACTION:
		self.skip()
		return cmd, nil

	default:
		return cmd, nil
	}
}

func (self *Parser) parseTransactionType() token.Token {
	rv := token.DEFERRED

	switch self.peek() {
	case token.DEFERRED, token.IMMEDIATE, token.EXCLUSIVE:
		rv = self.peek()
		self.skip()

	default:
		rv = token.DEFERRED
	}
	return rv
}

func (self *Parser) parseShow() (ast.Command, error) {
	cmd := &ast.Show{
		ShowPos: self.peekPos(),
	}
	self.skip() // skip "SHOW"

	switch self.peek() {
	case token.DATABASES, token.TABLES:
		cmd.Dest = self.peek()
		self.skip()
		return cmd, nil

	default:
		return nil, self.errorf(`Bad show command, unexpected: "%s"`, self.peekLiteral())
	}
}

func (self *Parser) parseCreate() (ast.Command, error) {
	self.skip() // skip `CREATE'
	switch self.peek() {
	case token.TABLE:
		return self.parseCreateTable(false)

	case token.TEMP:
		self.skip()
		return self.parseCreateTable(true)

	case token.INDEX:
		return self.parseCreateIndex(false)

	case token.UNIQUE:
		self.skip()
		return self.parseCreateIndex(true)

	default:
		return nil, self.errorf(`Bad create statement, unexpected "%s"`, self.peek().String())
	}
}

//------------------------------------------------------------------------------
// Create Table Actions:
//------------------------------------------------------------------------------
//
// CreateTable ::= `CREATE' Temp `TABLE' IfNotExists NameRef CreateTableArgs
//
// Temp        ::= `TEMP'
//               |
//
// IfNotExists ::= `IF' `NOT' `EXISTS'
//               |
//
//
func (self *Parser) parseCreateTable(temp bool) (*ast.CreateTable, error) {
	var err error

	cmd := &ast.CreateTable{
		CreatePos:   self.peekPos(),
		Temp:        temp,
		IfNotExists: false,
	}

	if _, err = self.match(token.TABLE); err != nil {
		return nil, err
	}

	if self.test(token.IF) {
		if err = self.batchMatch(token.NOT, token.EXISTS); err != nil {
			return nil, err
		}
		cmd.IfNotExists = true
	}

	if cmd.Table, err = self.parseNameRef(); err != nil {
		return nil, err
	}

	switch self.peek() {
	case token.LPAREN:
		self.skip()
		if cmd.Scheme, err = self.parseColumnScheme(cmd); err != nil {
			return nil, err
		}
		if _, err = self.match(token.RPAREN); err != nil {
			return nil, err
		}

	case token.AS:
		self.skip()
		if cmd.Template, err = self.parseSelect(); err != nil {
			return nil, err
		}

	default:
		return nil, self.errorf(`No table scheme be specified, unexpected "%s"`, self.peek().String())
	}
	cmd.CreateEnd = self.peekPos()
	return cmd, nil
}

//
// ColumnScheme     ::= ColumnScheme `,' ColumnDefine
//
// ColumnDefine     ::= Identifer TypeDecl ColumnOptionList
//
// ColumnOptionList ::= ColumnOptionList ColumnOption
//                    | ColumnOption
//                    |
//
func (self *Parser) parseColumnScheme(cmd *ast.CreateTable) ([]ast.ColumnDefine, error) {
	scheme := make([]ast.ColumnDefine, 0)

	for {
		def := ast.ColumnDefine{
			NotNullOn:    token.DEFAULT,
			UniqueOn:     token.DEFAULT,
			PrimaryKeyOn: token.DEFAULT,
		}
		var err error

		if def.Name, err = self.parseName(); err != nil {
			return scheme, err
		}

		var decl *ast.Type
		if decl, err = self.parseType(); err != nil {
			return scheme, err
		} else {
			def.ColumnType = *decl
		}

		var ok bool
		if ok, err = self.parseColumnOption(cmd, &def); err != nil {
			return scheme, err
		}
		for ok {
			if ok, err = self.parseColumnOption(cmd, &def); err != nil {
				return scheme, err
			}
		}

		scheme = append(scheme, def)
		if !self.test(token.COMMA) {
			break
		}
		if ok, err = self.parseColDefOption(cmd, scheme); err != nil {
			return scheme, err
		}
		if ok {
			for self.test(token.COMMA) {
				if ok, err = self.parseColDefOption(cmd, scheme); err != nil {
					return scheme, err
				}
				if !ok {
					return scheme, self.errorf(`Bad column define, unexpected "%s"`, self.peek().String())
				}
			}
			break
		}
	}
	return scheme, nil
}

//
// ColumnOption ::= `DEFAULT' Literal
//                | `DEFAULT' `(' Expr `)'
//                | `DEFAULT' Identifier
//                | `NULL' OnConf
//                | `NOT' `NULL' OnConf
//                | `PRIMARY' `KEY' SortOrder OnConf AutoIncr
//                | `UNIQUE' OnConf
//                | `CHECK' `(' Expr `)'
//                | `COLLATE' Identifier
//                |
//
// AutoIncr     ::= `AUTOINCR'
//                |
//
func (self *Parser) parseColumnOption(cmd *ast.CreateTable, def *ast.ColumnDefine) (bool, error) {
	var err error

	switch self.peek() {
	case token.DEFAULT:
		self.skip()
		if self.peek() == token.LPAREN {
			self.skip()
			if def.Default, err = self.NextExpr(); err != nil {
				return false, err
			}
			if _, err = self.match(token.RPAREN); err != nil {
				return false, err
			}
		} else {
			if def.Default, err = self.NextExpr(); err != nil {
				return false, err
			}
		}
		return true, nil

	case token.NULL:
		self.skip()
		return true, nil

	case token.NOT:
		self.skip()
		if _, err = self.match(token.NULL); err != nil {
			return false, err
		}
		if def.NotNullOn, err = self.parseOnConf(); err != nil {
			return false, err
		}
		def.NotNull = true
		return true, nil

	case token.PRIMARY:
		self.skip()
		if _, err = self.match(token.KEY); err != nil {
			return false, err
		}
		if self.test(token.ASC) {
			def.PrimaryKeyDesc = false
		} else if self.test(token.DESC) {
			def.PrimaryKeyDesc = true
		}
		if def.PrimaryKeyOn, err = self.parseOnConf(); err != nil {
			return false, err
		}
		if self.test(token.AUTOINCR) {
			def.AutoIncr = true
		}
		def.PrimaryKey = true
		return true, nil

	case token.UNIQUE:
		self.skip()
		if def.UniqueOn, err = self.parseOnConf(); err != nil {
			return false, err
		}
		def.Unique = true
		return true, nil

	case token.CHECK:
		self.skip()
		if _, err = self.match(token.LPAREN); err != nil {
			return false, nil
		}
		var expr ast.Expr
		if expr, err = self.NextExpr(); err != nil {
			return false, nil
		}
		if _, err = self.match(token.RPAREN); err != nil {
			return false, nil
		}
		cmd.CheckConstraint = append(cmd.CheckConstraint, expr)
		return true, nil

	case token.COLLATE:
		self.skip()
		if def.Collate, err = self.parseName(); err != nil {
			return false, nil
		} else {
			return true, nil
		}

	default:
		return false, nil
	}
}

//
// ColDefOption ::= `PRIMARY' `KEY' `(' IdxDefList AutoIncr `)' OnConf
//                | `UNIQUE' `(' IdxDefList `)' OnConf
//                | `CHECK' `(' Expr `)'
//
//
func (self *Parser) parseColDefOption(cmd *ast.CreateTable, scheme []ast.ColumnDefine) (bool, error) {
	var err error
	switch self.peek() {
	case token.PRIMARY:
		self.skip()
		if err = self.batchMatch(token.KEY, token.LPAREN); err != nil {
			return false, err
		}

		var list []ast.IndexDefine
		if list, err = self.parseIdxDefList(); err != nil {
			return false, err
		}

		var autoincr bool
		if self.test(token.AUTOINCR) {
			autoincr = true
		} else {
			autoincr = false
		}

		if _, err = self.match(token.RPAREN); err != nil {
			return false, err
		}

		var onconf token.Token
		if onconf, err = self.parseOnConf(); err != nil {
			return false, err
		}
		travelColumnDefine(list, scheme, func(idx *ast.IndexDefine, def *ast.ColumnDefine) {
			def.AutoIncr = autoincr
			def.PrimaryKeyDesc = idx.Desc
			def.PrimaryKeyOn = onconf
			def.Collate = idx.Collate
		})
		return true, nil

	case token.UNIQUE:
		self.skip()
		if _, err = self.match(token.LPAREN); err != nil {
			return false, err
		}

		var list []ast.IndexDefine
		if list, err = self.parseIdxDefList(); err != nil {
			return false, err
		}

		if _, err = self.match(token.RPAREN); err != nil {
			return false, err
		}

		var onconf token.Token
		if onconf, err = self.parseOnConf(); err != nil {
			return false, err
		}

		travelColumnDefine(list, scheme, func(idx *ast.IndexDefine, def *ast.ColumnDefine) {
			def.Unique = true
			def.UniqueOn = onconf
		})
		return true, nil

	case token.CHECK:
		self.skip()
		if _, err = self.match(token.LPAREN); err != nil {
			return false, err
		}

		var expr ast.Expr
		if expr, err = self.NextExpr(); err != nil {
			return false, err
		}

		if _, err = self.match(token.RPAREN); err != nil {
			return false, err
		}

		cmd.CheckConstraint = append(cmd.CheckConstraint, expr)
		return true, nil

	default:
		return false, nil
	}
}

func travelColumnDefine(idx []ast.IndexDefine, scheme []ast.ColumnDefine,
	closure func(idx *ast.IndexDefine, def *ast.ColumnDefine)) {
	for i := 0; i < len(idx); i++ {
		for j := 0; j < len(scheme); j++ {
			if idx[i].Name == scheme[j].Name {
				closure(&idx[i], &scheme[j])
			}
		}
	}
}

//
// OnConf  ::= `ON' `CONFLICT' Resolve
//           |
//
// Resolve ::= Raise
//           | `IGNORE'
//           | `DEFAULT'
//           | `REPLACE'
//
// Raise   ::= `ROLLBACK'
//           | `ABORT'
//           | `FAIL'
//
func (self *Parser) parseOnConf() (token.Token, error) {
	if !self.test(token.ON) {
		return token.DEFAULT, nil
	}

	var err error
	if _, err = self.match(token.CONFLICT); err != nil {
		return token.ILLEGAL, err
	}

	conf := self.peek()
	switch conf {
	case token.IGNORE, token.DEFAULT, token.REPLACE, token.ROLLBACK, token.ABORT,
		token.FAIL:
		self.skip()
		return conf, nil

	default:
		return token.ILLEGAL, self.errorf(`Bad "ON CONFLICT" option`)
	}
}

//------------------------------------------------------------------------------
// Create Index Actions:
//------------------------------------------------------------------------------
//
// CreateIndex ::= `CREATE' UniqueFlag `INDEX' IfNotExists NameRef `ON' `(' IdxDefList `)'
//
// UniqueFlag  ::= `UNIQUE'
//               |
//
func (self *Parser) parseCreateIndex(unique bool) (*ast.CreateIndex, error) {
	cmd := &ast.CreateIndex{
		CreatePos: self.peekPos(),
		Unique:    unique,
	}

	var err error
	if _, err = self.match(token.INDEX); err != nil {
		return nil, err
	}

	if self.test(token.IF) {
		if err = self.batchMatch(token.NOT, token.EXISTS); err != nil {
			return nil, err
		}
		cmd.IfNotExists = true
	}

	if cmd.Name, err = self.parseNameRef(); err != nil {
		return nil, err
	}

	if _, err = self.match(token.ON); err != nil {
		return nil, err
	}
	if cmd.Table, err = self.parseName(); err != nil {
		return nil, err
	}

	if _, err = self.match(token.LPAREN); err != nil {
		return nil, err
	}
	if cmd.Index, err = self.parseIdxDefList(); err != nil {
		return nil, err
	}
	if _, err = self.match(token.RPAREN); err != nil {
		return nil, err
	}

	cmd.CreateEnd = self.peekPos()
	return cmd, nil
}

func (self *Parser) parseIdxDefList() ([]ast.IndexDefine, error) {
	idx := make([]ast.IndexDefine, 0)

	parse_idx_def := func(p *Parser) (ast.IndexDefine, error) {
		var err error
		var def ast.IndexDefine
		if def.Name, err = p.parseName(); err != nil {
			return def, err
		}
		if p.test(token.COLLATE) {
			if def.Collate, err = p.parseName(); err != nil {
				return def, err
			}
		}
		if p.test(token.ASC) {
			def.Desc = false
		} else if p.test(token.DESC) {
			def.Desc = true
		}
		return def, nil
	}

	if elem, err := parse_idx_def(self); err != nil {
		return idx, err
	} else {
		idx = append(idx, elem)
	}
	for self.test(token.COMMA) {
		if elem, err := parse_idx_def(self); err != nil {
			return idx, err
		} else {
			idx = append(idx, elem)
		}
	}
	return idx, nil
}

//------------------------------------------------------------------------------
// Insert Actions:
//------------------------------------------------------------------------------
//
// Insert       ::= InsertPrefix `VALUES' `(' ExprList `)'
//                | InsertPrefix Select
//                | InsertPrefix `DEFAULT' `VALUES'
//
// InsertPrefix ::= InsCmd `INTO' NameRef InsColList
//
// InsCmd       ::= `INSERT' OrConf
//                | `REPLACE'
//
// InsColList   ::= `(' IdentifierList `)'
//                |
//
func (self *Parser) parseInsert() (*ast.Insert, error) {
	cmd := &ast.Insert{
		InsertPos: self.peekPos(),
		Op:        token.DEFAULT,
		Column:    make([]ast.Identifier, 0),
		Item:      make([]ast.Expr, 0),
	}

	var err error
	switch self.peek() {
	case token.INSERT:
		self.skip()
		if cmd.Op, err = self.parseOrConf(); err != nil {
			return nil, err
		}

	case token.REPLACE:
		self.skip()
		cmd.Op = token.REPLACE

	default:
		panic("No reached!")
	}

	if _, err = self.match(token.INTO); err != nil {
		return nil, err
	}
	if cmd.Dest, err = self.parseNameRef(); err != nil {
		return nil, err
	}

	if self.test(token.LPAREN) {
		if cmd.Column, err = self.parseIdentifierList(); err != nil {
			return nil, err
		}
		if _, err = self.match(token.RPAREN); err != nil {
			return nil, err
		}
	}

	switch self.peek() {
	case token.SELECT:
		if cmd.From, err = self.parseSelect(); err != nil {
			return nil, err
		}

	case token.VALUES:
		self.skip()
		if _, err = self.match(token.LPAREN); err != nil {
			return nil, err
		}
		if cmd.Item, err = self.parseExprList(); err != nil {
			return nil, err
		}
		if _, err = self.match(token.RPAREN); err != nil {
			return nil, err
		}

	case token.DEFAULT:
		self.skip()
		if _, err = self.match(token.VALUES); err != nil {
			return nil, err
		}

	default:
		return nil, self.errorf(`Insert statement need values, unexpected "%s"`, self.peek().String())
	}

	cmd.InsertEnd = self.peekPos()
	return cmd, nil
}

func (self *Parser) parseOrConf() (token.Token, error) {
	if !self.test(token.OR) {
		return token.DEFAULT, nil
	}

	conf := self.peek()
	switch conf {
	case token.IGNORE, token.DEFAULT, token.REPLACE, token.ROLLBACK, token.ABORT,
		token.FAIL:
		self.skip()
		return conf, nil

	default:
		return token.ILLEGAL, self.errorf(`Bad "OR" option`)
	}
}

//------------------------------------------------------------------------------
// Update Actions:
//------------------------------------------------------------------------------
//
// Update  ::= `UPDATE' OrConf NameRef Indexed `SET' SetList Where OrderBy Limit
//
// SetList ::= SetList `,' SetDefine
//           | SetDefine
//
func (self *Parser) parseUpdate() (*ast.Update, error) {
	cmd := &ast.Update{
		UpdatePos: self.peekPos(),
		Set:       make([]ast.SetDefine, 0),
	}
	self.skip() // skip `UPDATE'

	var err error
	if cmd.Op, err = self.parseOrConf(); err != nil {
		return nil, err
	}

	if cmd.Dest, err = self.parseNameRef(); err != nil {
		return nil, err
	}

	if cmd.Indexed, err = self.parseIndexed(); err != nil {
		return nil, err
	}

	if _, err = self.match(token.SET); err != nil {
		return nil, err
	}

	var def ast.SetDefine
	if def, err = self.parseSetDefine(); err != nil {
		return nil, err
	} else {
		cmd.Set = append(cmd.Set, def)
	}
	for self.test(token.COMMA) {
		if def, err = self.parseSetDefine(); err != nil {
			return nil, err
		} else {
			cmd.Set = append(cmd.Set, def)
		}
	}

	if self.test(token.WHERE) {
		if cmd.Where, err = self.NextExpr(); err != nil {
			return nil, err
		}
	}

	if self.test(token.ORDER) {
		if _, err = self.match(token.BY); err != nil {
			return nil, err
		}
		if cmd.OrderBy, err = self.parseOrderBy(); err != nil {
			return nil, err
		}
	}

	if self.test(token.LIMIT) {
		if cmd.Limit, cmd.Offset, err = self.parseLimitOffset(); err != nil {
			return nil, err
		}
	}

	cmd.UpdateEnd = self.peekPos()
	return cmd, nil
}

//
// SetDefine ::= Identifier `=' Expr
//
func (self *Parser) parseSetDefine() (ast.SetDefine, error) {
	var def ast.SetDefine
	var err error

	if def.Column, err = self.parseName(); err != nil {
		return def, err
	}
	if _, err = self.match(token.EQ); err != nil {
		return def, err
	}
	if def.Value, err = self.NextExpr(); err != nil {
		return def, err
	}
	return def, nil
}

//------------------------------------------------------------------------------
// Delete Actions:
//------------------------------------------------------------------------------
//
// Delete ::= `DELETE' `FROM' NameRef Indexed Where OrderBy Limit
//
func (self *Parser) parseDelete() (*ast.Delete, error) {
	cmd := &ast.Delete{
		DeletePos: self.peekPos(),
	}

	var err error
	if err = self.batchMatch(token.DELETE, token.FROM); err != nil {
		return nil, err
	}

	if cmd.Dest, err = self.parseNameRef(); err != nil {
		return nil, err
	}

	if cmd.Indexed, err = self.parseIndexed(); err != nil {
		return nil, err
	}

	if self.test(token.WHERE) {
		if cmd.Where, err = self.NextExpr(); err != nil {
			return nil, err
		}
	}

	if self.test(token.ORDER) {
		if _, err = self.match(token.BY); err != nil {
			return nil, err
		}
		if cmd.OrderBy, err = self.parseOrderBy(); err != nil {
			return nil, err
		}
	}

	if self.test(token.LIMIT) {
		if cmd.Limit, cmd.Offset, err = self.parseLimitOffset(); err != nil {
			return nil, err
		}
	}

	cmd.DeleteEnd = self.peekPos()
	return cmd, nil
}

//------------------------------------------------------------------------------
// Select Statement Actions:
//------------------------------------------------------------------------------
//
// Select       ::= Select SetOp SingleSelect
//                | SingleSelect
//
// SetOp        ::= `UNION'
//                | `UNION' `ALL'
//                | `EXCEPT'
//                | `INTERSECT'
//
// SingleSelect ::= `SELECT' Distinct SelColList From Where GroupBy Having OrderBy Limit
//
// Distinct     ::= `DISTINCT'
//                | `ALL'
//                |
//
func (self *Parser) parseSelect() (*ast.Select, error) {
	cmd := &ast.Select{
		SelectPos: self.peekPos(),
	}
	self.skip()

	if self.peek() == token.DISTINCT {
		self.skip()
		cmd.Distinct = true
	} else if self.peek() == token.ALL {
		self.skip()
		cmd.Distinct = false
	}

	var err error
	if cmd.SelColList, err = self.parseSelColList(); err != nil {
		return nil, err
	}

	if self.test(token.FROM) {
		if cmd.From, err = self.parseSelTabList(); err != nil {
			return nil, err
		}
	}

	if self.test(token.WHERE) {
		if cmd.Where, err = self.NextExpr(); err != nil {
			return nil, err
		}
	}

	if self.test(token.GROUP) {
		if _, err = self.match(token.BY); err != nil {
			return nil, err
		}
		if cmd.GroupBy, err = self.parseExprList(); err != nil {
			return nil, err
		}
	}

	if self.test(token.HAVING) {
		if cmd.Having, err = self.NextExpr(); err != nil {
			return nil, err
		}
	}

	if self.test(token.ORDER) {
		if _, err = self.match(token.BY); err != nil {
			return nil, err
		}
		if cmd.OrderBy, err = self.parseOrderBy(); err != nil {
			return nil, err
		}
	}

	if self.test(token.LIMIT) {
		if cmd.Limit, cmd.Offset, err = self.parseLimitOffset(); err != nil {
			return nil, err
		}
	}

	// End of select statement
	cmd.SelectEnd = self.peekPos()

	switch self.peek() {
	case token.UNION:
		self.skip()
		if self.test(token.ALL) {
			cmd.Op = token.UNION_ALL
		} else {
			cmd.Op = token.UNION
		}

	case token.EXCEPT:
		self.skip()
		cmd.Op = token.EXCEPT

	case token.INTERSECT:
		self.skip()
		cmd.Op = token.INTERSECT
	}
	if cmd.Op != 0 {
		if cmd.Prior, err = self.parseSelect(); err != nil {
			return nil, err
		}
	}
	return cmd, nil
}

//
// SelColList   ::= SelColList `,' SelectColumn
//                | SelectColumn
//
// SelectColumn ::= Expr
//                | Expr `AS' Identifer
//                | `*'
//
func (self *Parser) parseSelColList() ([]ast.SelectColumn, error) {
	column := make([]ast.SelectColumn, 0)

	for {
		var elem ast.SelectColumn

		if self.peek() == token.STAR {
			expr := &ast.Literal{
				ValuePos: self.peekPos(),
				Value:    self.peekLiteral(),
				Kind:     self.peek(),
			}
			self.skip()
			elem.SelectExpr = expr
			elem.Alias = ""
		} else {
			if expr, err := self.NextExpr(); err != nil {
				return column, err
			} else {
				elem.SelectExpr = expr
			}
			if self.peek() == token.AS {
				var err error
				if elem.Alias, err = self.parseAliasName(); err != nil {
					return column, err
				}
			}
		}
		column = append(column, elem)
		if !self.test(token.COMMA) {
			break
		}
	}
	return column, nil
}

//
// AliasName ::= `AS' Identifer
//
func (self *Parser) parseAliasName() (string, error) {
	if _, err := self.match(token.AS); err != nil {
		return "", err
	}
	return self.parseName()
}

func (self *Parser) parseName() (string, error) {
	if lah, err := self.match(token.ID); err != nil {
		return "", err
	} else {
		return strings.Trim(lah.Literal, "`"), nil
	}
}

//
// SelTabList ::= SelTabList JoinOp Source
//              | Source
//
// JoinOP     ::= `,'
//              | `LEFT' `OUTER' `JOIN'
//              | `LEFT' `JOIN'
//              | `RIGHT' `OUTER' `JOIN'
//              | `RIGHT' `JOIN'
//              | `FULL' `OUTER' `JOIN'
//              | `JOIN'
//              | `INNER' `JOIN'
//              | `CROSS' `JOIN'
//              | `NATURAL' `JOIN'
//
// Source     ::= Name Alias Indexed On Using
//              | `(' Select `)'' Alias On Using
//
func (self *Parser) parseSelTabList() ([]ast.Source, error) {
	source := make([]ast.Source, 0)

	for {
		var err error
		var elem ast.Source

		elem.SourcePos = self.peekPos()
		if self.test(token.LPAREN) {
			if elem.Subquery, err = self.parseSelect(); err != nil {
				return source, err
			}

			if _, err = self.match(token.RPAREN); err != nil {
				return source, err
			}
		} else {
			var name ast.NameRef
			if name, err = self.parseNameRef(); err != nil {
				return source, nil
			}
			elem.Table = &name
		}
		if self.test(token.AS) || self.peek() == token.ID {
			if elem.Alias, err = self.parseName(); err != nil {
				return source, err
			}
		}

		if elem.Table != nil {
			if elem.Indexed, err = self.parseIndexed(); err != nil {
				return source, err
			}
		}

		if self.test(token.USING) {
			if _, err = self.match(token.LPAREN); err != nil {
				return source, err
			}
			if elem.Using, err = self.parseIdentifierList(); err != nil {
				return source, err
			}
			if _, err = self.match(token.RPAREN); err != nil {
				return source, err
			}
		}

		elem.SourceEnd = self.peekPos()
		if elem.JoinType, err = self.parseJoinType(); err != nil {
			return source, err
		}

		if self.test(token.ON) {
			if _, err = self.match(token.LPAREN); err != nil {
				return source, err
			}
			if elem.On, err = self.NextExpr(); err != nil {
				return source, err
			}
			if _, err = self.match(token.RPAREN); err != nil {
				return source, err
			}
		}
		source = append(source, elem)

		if elem.JoinType == 0 {
			break
		}
	}
	return source, nil
}

func (self *Parser) parseIndexed() (string, error) {
	var err error
	indexed := ""

	switch self.peek() {
	case token.INDEXED:
		self.skip()
		if _, err = self.match(token.BY); err != nil {
			return "", err
		}
		if indexed, err = self.parseName(); err != nil {
			return "", err
		}

	case token.NOT:
		self.skip()
		if _, err = self.match(token.INDEXED); err != nil {
			return "", err
		}
	}
	return indexed, nil
}

func (self *Parser) parseJoinType() (int, error) {
	if self.test(token.COMMA) {
		return ast.JT_INNER, nil
	}

	jt := 0
	for {
		switch self.peek() {
		case token.INNER:
			self.skip()
			jt |= ast.JT_INNER

		case token.CROSS:
			self.skip()
			jt |= ast.JT_CROSS

		case token.NATURAL:
			self.skip()
			jt |= ast.JT_NATURAL

		case token.LEFT:
			self.skip()
			jt |= ast.JT_LEFT

		case token.RIGHT:
			self.skip()
			jt |= ast.JT_RIGHT

		case token.OUTER:
			self.skip()
			jt |= ast.JT_OUTER

		case token.JOIN:
			if jt == 0 {
				jt = ast.JT_INNER
			}
			self.skip()
			return jt, nil

		default:
			return 0, nil
		}
	}
}

func (self *Parser) parseOrderBy() ([]ast.OrderByItem, error) {
	item := make([]ast.OrderByItem, 0)

	for {
		var elem ast.OrderByItem

		if expr, err := self.NextExpr(); err != nil {
			return item, err
		} else {
			elem.Item = expr
		}

		if self.test(token.ASC) {
			elem.Desc = false
		} else if self.test(token.DESC) {
			elem.Desc = true
		}
		item = append(item, elem)
		if !self.test(token.COMMA) {
			return item, nil
		}
	}
}

//
// LimitOffset ::= `LIMIT' IntLiteral
//               | `LIMIT' IntLiteral `,' IntLiteral
//               | `LIMIT' IntLiteral `OFFSET' IntLiteral
//
func (self *Parser) parseLimitOffset() (ast.Expr, ast.Expr, error) {
	var limit, offset ast.Expr
	var err error

	if limit, err = self.parseIntLiteral(); err != nil {
		return nil, nil, err
	}

	switch self.peek() {
	case token.COMMA:
		self.skip()
		offset = limit
		limit, err = self.parseIntLiteral()
		return limit, offset, err

	case token.OFFSET:
		self.skip()
		offset, err = self.parseIntLiteral()
		return limit, offset, err

	default:
		return limit, nil, nil
	}
}

//------------------------------------------------------------------------------
// Expression Actions:
//------------------------------------------------------------------------------
func (self *Parser) NextExpr() (ast.Expr, error) {
	_, expr, err := self.parseExpr(0)
	return expr, err
}

func (self *Parser) parseExpr(limit int) (token.Token, ast.Expr, error) {
	var expr ast.Expr
	var err error

	if self.peek().Prefix() {
		unary := &ast.UnaryExpr{
			OpPos: self.peekPos(),
			Op:    self.peek(),
		}
		self.skip()

		if _, unary.Operand, err = self.parseExpr(kPrioPrefix); err != nil {
			return token.ILLEGAL, nil, err
		}
		expr = unary
	} else {
		if expr, err = self.parseSimple(); err != nil {
			return token.ILLEGAL, nil, err
		}
	}

next:
	op := self.peek()
	for op.Binary() && priority(op).Lhs > limit {
		binary := &ast.BinaryExpr{
			OpPos: self.peekPos(),
			Op:    op,
			Lhs:   expr,
		}
		self.skip()

		switch op {
		case token.IN:
			if binary.Rhs, err = self.parseWhereInSet(); err != nil {
				return token.ILLEGAL, nil, err
			}
			op = self.peek()

		case token.LIKE:
			if self.peek() != token.STRING_LITERAL {
				return token.ILLEGAL, nil, self.errorf("LIKE operator need string pattern")
			}
			binary.Rhs = &ast.Literal{
				ValuePos: self.peekPos(),
				Value:    self.peekLiteral(),
				Kind:     self.peek(),
			}
			self.skip()
			op = self.peek()

		default:
			if op, binary.Rhs, err = self.parseExpr(priority(op).Rhs); err != nil {
				return token.ILLEGAL, nil, err
			}
		}
		expr = binary
	}

	if op.Postfix() && kPrioPostfix > limit {
		if expr, err = self.parsePostfix(expr); err != nil {
			return token.ILLEGAL, nil, err
		}
		goto next
	}
	return op, expr, err
}

//
// WhereInSet ::= `(' SelectStatement `)'
//              | `(' ExprList `)'
func (self *Parser) parseWhereInSet() (ast.Expr, error) {
	_, err := self.match(token.LPAREN)
	if err != nil {
		return nil, err
	}
	var expr ast.Expr
	if self.peek() == token.SELECT {
		if expr, err = self.parseSelect(); err != nil {
			return nil, err
		}
	} else {
		var list []ast.Expr
		if list, err = self.parseExprList(); err != nil {
			return nil, err
		}
		expr = ast.ExprList(list)
	}
	if _, err = self.match(token.RPAREN); err != nil {
		return nil, err
	}
	return expr, nil
}

func (self *Parser) parseSimple() (ast.Expr, error) {
	switch self.peek() {
	case token.NULL, token.INT_LITERAL, token.FLOAT_LITERAL, token.STRING_LITERAL:
		expr := &ast.Literal{
			ValuePos: self.peekPos(),
			Value:    self.peekLiteral(),
			Kind:     self.peek(),
		}
		self.skip()
		return expr, nil

	case token.CASE:
		return self.parseCondition()

	case token.CAST:
		return self.parseCast()

	default:
		return self.parseSuffixed()
	}
}

func (self *Parser) parseCondition() (ast.Expr, error) {
	cond := &ast.Condition{
		OpPos:  self.peekPos(),
		Blocks: make([]ast.ConditionBlock, 0),
	}
	self.skip()

	var err error
	if self.peek() != token.WHEN {
		if cond.Case, err = self.NextExpr(); err != nil {
			return nil, err
		}
	}

	var block ast.ConditionBlock
	for self.test(token.WHEN) {
		if block.When, err = self.NextExpr(); err != nil {
			return nil, err
		}
		if _, err = self.match(token.THEN); err != nil {
			return nil, err
		}
		if block.Then, err = self.NextExpr(); err != nil {
			return nil, err
		}
		cond.Blocks = append(cond.Blocks, block)
	}
	if len(cond.Blocks) == 0 {
		return nil, self.errorf(`WHEN ... THEN ... block not found`)
	}

	if self.peek() == token.ELSE {
		self.skip()
		if cond.Else, err = self.NextExpr(); err != nil {
			return nil, err
		}
	}
	return cond, nil
}

//
// CastExpr ::= `CAST' `(' Expr `AS' Type `)'
func (self *Parser) parseCast() (*ast.CastExpr, error) {
	cast := &ast.CastExpr{
		OpPos: self.peekPos(),
	}
	self.skip() // skip `CAST'

	var err error
	if _, err = self.match(token.LPAREN); err != nil {
		return nil, err
	}

	if cast.Operand, err = self.NextExpr(); err != nil {
		return nil, err
	}

	if _, err = self.match(token.AS); err != nil {
		return nil, err
	}

	var ty *ast.Type
	if ty, err = self.parseType(); err != nil {
		return nil, err
	}
	cast.To = *ty

	if _, err = self.match(token.RPAREN); err != nil {
		return nil, err
	}
	return cast, nil
}

//
// TypeDecl ::= Type Sign
//            | Type `(' IntLiteral `)' Sign
//            | Type `(' IntLiteral `,' IntLiteral `)' Sign
//
// Sign     ::= `UNSIGNED'
//            |
//
// Type     ::= `TINYINT'
//            | `SMALLINT'
//            | `INT'
//            | ...
func (self *Parser) parseType() (*ast.Type, error) {
	if self.peek().Kind() != token.TT_KEYWORD {
		return nil, self.errorf(`"%s" not type!`, self.peek().String())
	}

	decl := &ast.Type{
		TokenPos: self.peekPos(),
		Kind:     self.peek(),
		Unsigned: false,
	}
	self.skip()

	if self.peek() == token.LPAREN {
		self.skip()

		var err error
		if decl.Width, err = self.parseIntLiteral(); err != nil {
			return nil, err
		}

		if self.peek() == token.COMMA {
			self.skip()
			if decl.Decimal, err = self.parseIntLiteral(); err != nil {
				return nil, err
			}
		}

		if _, err = self.match(token.RPAREN); err != nil {
			return nil, err
		}
	}

	if self.peek() == token.UNSIGNED {
		self.skip()

		decl.Unsigned = true
	}
	return decl, nil
}

func (self *Parser) parseIntLiteral() (*ast.Literal, error) {
	lah, err := self.match(token.INT_LITERAL)
	if err != nil {
		return nil, err
	}
	return &ast.Literal{
		ValuePos: lah.Pos,
		Value:    lah.Literal,
		Kind:     lah.Token,
	}, nil
}

func (self *Parser) parseSuffixed() (ast.Expr, error) {
	id, err := self.parsePrimary()
	if err != nil {
		return nil, err
	}
	if isDot(id) {
		return id, nil
	}

	if self.peek() == token.LPAREN {
		self.skip()

		call := &ast.CallExpr{
			Func:     *(id.(*ast.Identifier)),
			Args:     make([]ast.Expr, 0),
			Distinct: false,
		}
		if self.peek() == token.STAR {
			star := &ast.Literal{
				ValuePos: self.peekPos(),
				Value:    self.peekLiteral(),
				Kind:     self.peek(),
			}
			self.skip()
			call.Args = append(call.Args, star)
			if _, err = self.match(token.RPAREN); err != nil {
				return nil, err
			}
			return call, err
		}

		if self.peek() == token.DISTINCT {
			self.skip()
			call.Distinct = true
		}
		if self.peek() != token.RPAREN {
			if call.Args, err = self.parseExprList(); err != nil {
				return nil, err
			}
		}
		if _, err = self.match(token.RPAREN); err != nil {
			return nil, err
		}
		return call, nil
	}
	return id, nil
}

func (self *Parser) parsePrimary() (ast.Expr, error) {
	var expr ast.Expr
	var err error

	switch self.peek() {
	case token.LPAREN:
		self.skip()
		if expr, err = self.NextExpr(); err != nil {
			return nil, err
		} else if _, err = self.match(token.RPAREN); err != nil {
			return nil, err
		} else {
			return expr, nil
		}

	case token.ID:
		if expr, err = self.parseIdentifier(); err != nil {
			return nil, err
		}
		if self.peek() == token.DOT {
			bin := &ast.BinaryExpr{
				OpPos: self.peekPos(),
				Op:    token.DOT,
				Lhs:   expr,
			}
			self.skip()
			if expr, err = self.parseIdentifier(); err != nil {
				return nil, err
			}
			bin.Rhs = expr
			expr = bin
		}
		return expr, nil

	default:
		return nil, self.errorf(`Unexpected expression, expected "%s"`, self.peekLiteral())
	}
}

func (self *Parser) parseExprList() ([]ast.Expr, error) {
	list := make([]ast.Expr, 0)
	expr, err := self.NextExpr()
	if err != nil {
		return list, err
	}
	list = append(list, expr)
	for self.test(token.COMMA) {
		expr, err = self.NextExpr()
		if err != nil {
			return list, err
		}
		list = append(list, expr)
	}
	return list, nil
}

func (self *Parser) parseIdentifier() (*ast.Identifier, error) {
	tok, err := self.match(token.ID)
	if err != nil {
		return nil, err
	} else {
		return &ast.Identifier{
			NamePos: tok.Pos,
			Name:    tok.Literal,
		}, nil
	}
}

func (self *Parser) parsePostfix(expr ast.Expr) (ast.Expr, error) {
	switch self.peek() {
	case token.IS:
		unary := &ast.UnaryExpr{
			OpPos:   self.peekPos(),
			Operand: expr,
		}
		self.skip()
		if self.peek() == token.NOT {
			self.skip()
			unary.Op = token.IS_NOT_NULL
		} else {
			unary.Op = token.IS_NULL
		}
		if _, err := self.match(token.NULL); err != nil {
			return nil, err
		}
		return unary, nil

	default:
		return expr, nil
	}
}

func (self *Parser) parseIdentifierList() ([]ast.Identifier, error) {
	id := make([]ast.Identifier, 0)

	for {
		var elem ast.Identifier

		if lah, err := self.match(token.ID); err != nil {
			return id, err
		} else {
			elem.NamePos = lah.Pos
			elem.Name = lah.Literal
		}
		id = append(id, elem)
		if !self.test(token.COMMA) {
			break
		}
	}
	return id, nil
}

func (self *Parser) parseNameRef() (ast.NameRef, error) {
	var name ast.NameRef
	if lah, err := self.match(token.ID); err != nil {
		return name, err
	} else {
		name.First = strings.Trim(lah.Literal, "`")
	}

	if self.test(token.DOT) {
		if lah, err := self.match(token.ID); err != nil {
			return name, err
		} else {
			name.Second = strings.Trim(lah.Literal, "`")
		}
	}
	return name, nil
}

func (self *Parser) errorf(s string, a ...interface{}) error {
	switch self.peek() {
	case token.ILLEGAL:
		return fmt.Errorf("[%d] Illegal token: %v", self.peekPos(), self.lex.Error())

	case token.EOF:
		return fmt.Errorf("Command already end")

	default:
		return fmt.Errorf(`[%d] %s`, self.peekPos(), fmt.Sprintf(s, a...))
	}
}

func (self *Parser) peek() token.Token {
	return self.lah.Token
}

func (self *Parser) peekPos() int {
	return self.lah.Pos
}

func (self *Parser) peekLiteral() string {
	return self.lah.Literal
}

func (self *Parser) test(exp token.Token) bool {
	if self.peek() == exp {
		self.skip()
		return true
	} else {
		return false
	}
}

func (self *Parser) skip() {
	self.lah.Pos, self.lah.Token, self.lah.Literal = self.lex.Next()
}

func (self *Parser) batchMatch(list ...token.Token) error {
	for _, elem := range list {
		if _, err := self.match(elem); err != nil {
			return err
		}
	}
	return nil
}

func (self *Parser) match(exp token.Token) (tokeniton, error) {
	var prev tokeniton
	if self.peek() != exp {
		return prev, self.errorf(`Unexpected "%s", expected "%s"`, exp, self.peekLiteral())
	}
	prev = self.lah
	self.skip()
	return prev, nil
}

func priority(op token.Token) priorition {
	prio, found := prio[op]
	if !found {
		panic(fmt.Sprintf("Op(%s) not found", op))
	}
	return prio
}

func isDot(expr ast.Expr) bool {
	bin, ok := expr.(*ast.BinaryExpr)
	if !ok {
		return false
	}
	return bin.Op == token.DOT
}

type tokeniton struct {
	Token   token.Token
	Pos     int
	Literal string
}

type priorition struct {
	Lhs int
	Rhs int
}

const (
	kPrioPrefix  = 9
	kPrioPostfix = 1
)

var prio = map[token.Token]priorition{
	token.LIKE: priorition{8, 8},

	token.STAR:  priorition{7, 7},
	token.SLASH: priorition{7, 7},

	token.PLUS:  priorition{6, 6},
	token.MINUS: priorition{6, 6},

	token.IN: priorition{5, 5},

	token.NE: priorition{4, 4},
	token.EQ: priorition{4, 4},

	token.LT: priorition{3, 3},
	token.LE: priorition{3, 3},
	token.GT: priorition{3, 3},
	token.GE: priorition{3, 3},

	token.AND: priorition{2, 2},
	token.OR:  priorition{1, 1},
}

package token

import (
//"strings"
)

type Token int

const (
	ILLEGAL Token = iota
	EOF
	COMMENT // -- This is statement -- /* This is statement */

	SELECT
	INSERT
	UPDATE
	CREATE
	DELETE
	DROP
	FROM
	WHERE
	GROUP
	ASC
	DESC
	ORDER
	HAVING
	LIMIT
	OFFSET
	TABLE
	DISTINCT
	ALL
	UNION
	UNION_ALL
	EXCEPT
	INTERSECT
	TEMP
	IF
	EXISTS
	PRIMARY
	KEY
	UNIQUE
	CHECK
	AUTOINCR
	COLLATE
	INDEX
	INTO
	VALUES
	SET

	// Source
	USING
	INDEXED
	BY
	INNER
	CROSS
	NATURAL
	LEFT
	RIGHT
	OUTER
	JOIN
	ON

	// Misc Command
	SHOW
	DATABASES
	TABLES
	START
	BEGIN
	TRANSACTION
	COMMIT
	END
	ROLLBACK
	DEFERRED
	IMMEDIATE
	EXCLUSIVE
	IGNORE
	DEFAULT
	REPLACE
	//ROLLBACK
	ABORT
	FAIL
	CONFLICT

	ID             // name `name`
	INT_LITERAL    // 1024
	FLOAT_LITERAL  // 1.24
	STRING_LITERAL // '1.24' "1234"
	NULL           // NULL

	IN          // field IN (1, 2, 3)
	IS          // IS
	IS_NULL     // IS NULL
	IS_NOT_NULL // IS NOT NULL

	EQ // = ==
	NE // <> !=
	LT // <
	LE // <=
	GT // >
	GE // >=

	SLASH // a / b
	STAR  // a * b
	PLUS  // a + b
	MINUS // a - b, -a

	COMMA  // ,
	DOT    // .
	SEMI   // ;
	LPAREN // (
	RPAREN // )

	// Logic operator
	AND
	OR
	NOT

	// Condition
	CASE
	WHEN
	THEN
	ELSE

	// Other operators
	LIKE
	CAST
	AS

	// Row types
	BIT        // (M) 1 <= M <= 64 BIT = BIT(1)
	TINYINT    // (M) [UNSIGNED] [ZEROFILL]
	BOOL       // = TINYINT(1)
	BOOLEAN    // = BOOL
	SMALLINT   // (M) [UNSIGNED] [ZEROFILL] 16 bits int
	MEDIUMINT  // (M) [UNSIGNED] [ZEROFILL] 24 bits int
	INT        // (M) [UNSIGNED] [ZEROFILL] 32 bits int
	INTEGER    // = INT
	BIGINT     // (M) [UNSIGNED] [ZEROFILL] 64 bits int
	FLOAT      // (M, D) [UNSIGNED] [ZEROFILL]
	DOUBLE     // PRECISION(M, D) [UNSIGNED] [ZEROFILL], REAL(M, D) [UNSIGNED] [ZEROFILL]
	DECIMAL    // (M, D) [UNSIGNED] [ZEROFILL]
	DATE       // '0000-00-00'
	DATETIME   // '0000-00-00 00:00:00'
	TIMESTAMP  // 00000000000000
	TIME       // '00:00:00'
	YEAR       // (2|4) 0000
	CHAR       // (M) [BINARY|ASCII|UNICODE] CHAR = CHAR(1)
	VARCHAR    // (M) [BINARY] 0 <= M <= 16K
	BINARY     // Like CHAR
	VARBINARY  // (M) 0 <= M <= 16K
	TINYBLOB   // 255B BLOB
	TINYTEXT   // 255B TEXT
	BLOB       // (M) 16KB BLOB
	TEXT       // (M) 16KB TEXT
	MEDIUMBLOB // 16MB BLOB
	MEDIUMTEXT // 16MB TEXT
	LONGBLOB   // 4GB BLOB
	LONGTEXT   // 4GB TEXT
	UNSIGNED
)

type Type int

const (
	TT_OPERATOR Type = iota
	TT_KEYWORD
	TT_LITERAL
	TT_MARK
)

var Keyword = map[string]Token{}

func init() {
	for k, v := range tokenMetadata {
		if v.Kind == TT_KEYWORD {
			Keyword[v.Text] = Token(k)
		}
	}
}

func (self Token) Prefix() bool {
	switch self {
	case MINUS, NOT:
		return true

	default:
		return false
	}
}

func (self Token) Postfix() bool {
	return self == IS
}

func (self Token) Binary() bool {
	switch self {
	case EQ, NE, LT, LE, GT, GE, SLASH, STAR, PLUS, MINUS, AND, OR, NOT, DOT, IN, LIKE:
		return true

	default:
		return false
	}
}

func (self Token) String() string {
	return tokenMetadata[self].Text
}

func (self Token) Kind() Type {
	return tokenMetadata[self].Kind
}

type tokeniton struct {
	Text string
	Kind Type
}

var tokenMetadata = []tokeniton{
	tokeniton{"illegal", TT_MARK},    // ILLEGAL Token = iota
	tokeniton{"EOF", TT_MARK},        // EOF
	tokeniton{"comment", TT_LITERAL}, // COMMENT

	tokeniton{"SELECT", TT_KEYWORD},
	tokeniton{"INSERT", TT_KEYWORD},
	tokeniton{"UPDATE", TT_KEYWORD},
	tokeniton{"CREATE", TT_KEYWORD},
	tokeniton{"DELETE", TT_KEYWORD},
	tokeniton{"DROP", TT_KEYWORD},
	tokeniton{"FROM", TT_KEYWORD},
	tokeniton{"WHERE", TT_KEYWORD},
	tokeniton{"GROUP", TT_KEYWORD},
	tokeniton{"ASC", TT_KEYWORD},
	tokeniton{"DESC", TT_KEYWORD},
	tokeniton{"ORDER", TT_KEYWORD},
	tokeniton{"HAVING", TT_KEYWORD},
	tokeniton{"LIMIT", TT_KEYWORD},
	tokeniton{"OFFSET", TT_KEYWORD},
	tokeniton{"TABLE", TT_KEYWORD},
	tokeniton{"DISTINCT", TT_KEYWORD},
	tokeniton{"ALL", TT_KEYWORD},
	tokeniton{"UNION", TT_KEYWORD},
	tokeniton{"UNION ALL", TT_OPERATOR},
	tokeniton{"EXCEPT", TT_KEYWORD},
	tokeniton{"INTERSECT", TT_KEYWORD},
	tokeniton{"TEMP", TT_KEYWORD},
	tokeniton{"IF", TT_KEYWORD},
	tokeniton{"EXISTS", TT_KEYWORD},
	tokeniton{"PRIMARY", TT_KEYWORD},
	tokeniton{"KEY", TT_KEYWORD},
	tokeniton{"UNIQUE", TT_KEYWORD},
	tokeniton{"CHECK", TT_KEYWORD},
	tokeniton{"AUTOINCR", TT_KEYWORD},
	tokeniton{"COLLATE", TT_KEYWORD},
	tokeniton{"INDEX", TT_KEYWORD},
	tokeniton{"INTO", TT_KEYWORD},
	tokeniton{"VALUES", TT_KEYWORD},
	tokeniton{"SET", TT_KEYWORD},

	// Source
	tokeniton{"USING", TT_KEYWORD},
	tokeniton{"INDEXED", TT_KEYWORD},
	tokeniton{"BY", TT_KEYWORD},
	tokeniton{"INNER", TT_KEYWORD},
	tokeniton{"CROSS", TT_KEYWORD},
	tokeniton{"NATURAL", TT_KEYWORD},
	tokeniton{"LEFT", TT_KEYWORD},
	tokeniton{"RIGHT", TT_KEYWORD},
	tokeniton{"OUTER", TT_KEYWORD},
	tokeniton{"JOIN", TT_KEYWORD},
	tokeniton{"ON", TT_KEYWORD},

	// Misc Command
	tokeniton{"SHOW", TT_KEYWORD},
	tokeniton{"DATABASES", TT_KEYWORD},
	tokeniton{"TABLES", TT_KEYWORD},
	tokeniton{"START", TT_KEYWORD},
	tokeniton{"BEGIN", TT_KEYWORD},
	tokeniton{"TRANSACTION", TT_KEYWORD},
	tokeniton{"COMMIT", TT_KEYWORD},
	tokeniton{"END", TT_KEYWORD},
	tokeniton{"ROLLBACK", TT_KEYWORD},
	tokeniton{"DEFERRED", TT_KEYWORD},
	tokeniton{"IMMEDIATE", TT_KEYWORD},
	tokeniton{"EXCLUSIVE", TT_KEYWORD},
	tokeniton{"IGNORE", TT_KEYWORD},
	tokeniton{"DEFAULT", TT_KEYWORD},
	tokeniton{"REPLACE", TT_KEYWORD},
	tokeniton{"ABORT", TT_KEYWORD},
	tokeniton{"FAIL", TT_KEYWORD},
	tokeniton{"CONFLICT", TT_KEYWORD},

	tokeniton{"identifier", TT_LITERAL}, // ID
	tokeniton{"integer", TT_LITERAL},    // INT_LITERAL
	tokeniton{"float", TT_LITERAL},      // FLOAT_LITERAL
	tokeniton{"string", TT_LITERAL},     // STRING_LITERAL
	tokeniton{"NULL", TT_KEYWORD},

	tokeniton{"IN", TT_KEYWORD},
	tokeniton{"IS", TT_KEYWORD},
	tokeniton{"IS NULL", TT_OPERATOR},     // IS_NULL
	tokeniton{"IS NOT NULL", TT_OPERATOR}, // IS_NOT_NULL

	tokeniton{"=", TT_OPERATOR},  // EQ
	tokeniton{"<>", TT_OPERATOR}, // NE
	tokeniton{"<", TT_OPERATOR},  // LT
	tokeniton{"<=", TT_OPERATOR}, // LE
	tokeniton{">", TT_OPERATOR},  // GT
	tokeniton{">=", TT_OPERATOR}, // GE

	tokeniton{"/", TT_OPERATOR}, // SLASH
	tokeniton{"*", TT_OPERATOR}, // STAR
	tokeniton{"+", TT_OPERATOR}, // PLUS
	tokeniton{"-", TT_OPERATOR}, // MINUS

	tokeniton{",", TT_OPERATOR}, // COMMA
	tokeniton{".", TT_OPERATOR}, // DOT
	tokeniton{";", TT_OPERATOR}, // SEMI
	tokeniton{"(", TT_OPERATOR}, // LPAREN
	tokeniton{")", TT_OPERATOR}, // RPAREN

	tokeniton{"AND", TT_KEYWORD},
	tokeniton{"OR", TT_KEYWORD},
	tokeniton{"NOT", TT_KEYWORD},

	// Condition
	tokeniton{"CASE", TT_KEYWORD},
	tokeniton{"WHEN", TT_KEYWORD},
	tokeniton{"THEN", TT_KEYWORD},
	tokeniton{"ELSE", TT_KEYWORD},

	// Other operators
	tokeniton{"LIKE", TT_KEYWORD},
	tokeniton{"CAST", TT_KEYWORD},
	tokeniton{"AS", TT_KEYWORD},

	// Row types
	tokeniton{"BIT", TT_KEYWORD},
	tokeniton{"TINYINT", TT_KEYWORD},
	tokeniton{"BOOL", TT_KEYWORD},
	tokeniton{"BOOLEAN", TT_KEYWORD},
	tokeniton{"SMALLINT", TT_KEYWORD},
	tokeniton{"MEDIUMINT", TT_KEYWORD},
	tokeniton{"INT", TT_KEYWORD},
	tokeniton{"INTEGER", TT_KEYWORD},
	tokeniton{"BIGINT", TT_KEYWORD},
	tokeniton{"FLOAT", TT_KEYWORD},
	tokeniton{"DOUBLE", TT_KEYWORD},
	tokeniton{"DECIMAL", TT_KEYWORD},
	tokeniton{"DATE", TT_KEYWORD},
	tokeniton{"DATETIME", TT_KEYWORD},
	tokeniton{"TIMESTAMP", TT_KEYWORD},
	tokeniton{"TIME", TT_KEYWORD},
	tokeniton{"YEAR", TT_KEYWORD},
	tokeniton{"CHAR", TT_KEYWORD},
	tokeniton{"VARCHAR", TT_KEYWORD},
	tokeniton{"BINARY", TT_KEYWORD},
	tokeniton{"VARBINARY", TT_KEYWORD},
	tokeniton{"TINYBLOB", TT_KEYWORD},
	tokeniton{"TINYTEXT", TT_KEYWORD},
	tokeniton{"BLOB", TT_KEYWORD},
	tokeniton{"TEXT", TT_KEYWORD},
	tokeniton{"MEDIUMBLOB", TT_KEYWORD},
	tokeniton{"MEDIUMTEXT", TT_KEYWORD},
	tokeniton{"LONGBLOB", TT_KEYWORD},
	tokeniton{"LONGTEXT", TT_KEYWORD},
	tokeniton{"UNSIGNED", TT_KEYWORD},
}

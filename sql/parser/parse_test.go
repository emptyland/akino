package parser

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/emptyland/akino/sql/ast"
	"github.com/emptyland/akino/sql/token"
)

func TestSanity(t *testing.T) {
	cmd, err := ParseCommand("BEGIN TRANSACTION")
	if err != nil {
		t.Fatal(err)
	}

	rv := cmd.(*ast.Transaction)
	if rv.TransactionPos != 0 {
		t.Fatal("Bad pos")
	}
	if rv.Type != token.DEFERRED {
		t.Fatal("Bad type")
	}
	if rv.Op != token.BEGIN {
		t.Fatal("Bad op")
	}
}

func TestDotIdExpr(t *testing.T) {
	assertExpr(t, "db.name", "dot_id")
	assertExpr(t, "`db`.`name`", "quoted_dot_id")
	assertExpr(t, "`DATABASE`.`INDEX`", "quoted_kw_id")
}

func TestArithExpr(t *testing.T) {
	assertExpr(t, "(1 + 2) * id", "aright_0")
	assertExpr(t, " 1 + 2  / id", "aright_1")
}

func TestCondExpr(t *testing.T) {
	expr := `CASE WHEN 1 THEN -1 WHEN 2 THEN -2 ELSE NULL`
	assertExpr(t, expr, "cond_0")

	expr = `CASE id WHEN 1 THEN "first" WHEN 2 THEN "second" ELSE NULL`
	assertExpr(t, expr, "cond_1")

	expr = `CASE id
WHEN 100 THEN
	CASE name
		WHEN 'Jack' THEN 1
		WHEN 'Tom' THEN 2
		ELSE COUNT(DISTINCT name)
WHEN 200 THEN -200
WHEN 300 THEN -300`
	assertExpr(t, expr, "cond_recrsion")
}

func TestFuncCall(t *testing.T) {
	assertExpr(t, "MAX(amt)", "call_func_max")
	assertExpr(t, "POW(1, 2)", "call_func_pow")
	assertExpr(t, "SUM(DISTINCT amt)", "call_func_sum")
	assertExpr(t, "COUNT(*)", "call_func_count")
}

func TestIsOrNotNull(t *testing.T) {
	assertExpr(t, "amt IS NOT NULL", "is_not_null_0")
	assertExpr(t, "1 + amt IS NOT NULL", "is_not_null_1")
	assertExpr(t, "amt IS NOT NULL + 1", "is_not_null_2")
	assertExpr(t, "(1 + 2) * d + COUNT(*) IS NULL", "is_null_3")
	assertExpr(t, "(1 + 2) IS NULL * d + COUNT(*) IS NOT NULL", "is_or_not_null_4")
}

func TestIsOrNotNullRecursion(t *testing.T) {
	assertExpr(t, "amt IS NOT NULL IS NOT NULL IS NOT NULL", "is_or_not_null_recursion_0")
	assertExpr(t, "amt IS NOT NULL IS NULL IS NOT NULL", "is_or_not_null_recursion_1")
}

func TestLogicOperator(t *testing.T) {
	assertExpr(t, "NOT amt", "not")
	assertExpr(t, "NOT NOT NOT amt", "not_recursion")
	assertExpr(t, "1 AND 0 OR 1", "and_or_0")
	assertExpr(t, "1 AND (0 OR 1)", "and_or_1")
}

func TestWhereInSet(t *testing.T) {
	assertExpr(t, "id IN (1, 2, 3, 4)", "where_in_list_0")
	assertExpr(t, "num IN (COUNT(*), 4 + id * 2)", "where_in_list_1")
	assertExpr(t, "1 + id IN (1, 2)", "where_in_list_2")
	assertExpr(t, "1 > id IN (1, 2)", "where_in_list_3")
	assertExpr(t, "id IN (1, 2) < 1", "where_in_list_4")
}

func TestWhereInSubquery(t *testing.T) {
	assertExpr(t, "a IN (SELECT a FROM t)", "where_in_subquery")
}

func TestLike(t *testing.T) {
	assertExpr(t, `"name1234" LIKE "name%"`, "like_0")
	assertExpr(t, `name LIKE "name%"`, "like_1")
	assertExpr(t, `db.name LIKE "name%"`, "like_2")
}

func TestCast(t *testing.T) {
	assertExpr(t, "CAST (1 AS INT)", "cast_int")
	assertExpr(t, "CAST (1 AS INT(4) UNSIGNED)", "cast_uint")
	assertExpr(t, "CAST (1 AS DOUBLE(6, 2))", "cast_double")
	assertExpr(t, `CAST ("hello" AS VARCHAR(8))`, "cast_varchar")
}

func TestSelectSanity(t *testing.T) {
	assertCmd(t, "SELECT * FROM t;", "select_sanity")
	assertCmd(t, "SELECT DISTINCT * FROM t;", "select_distinct")
	assertCmd(t, "SELECT ALL * FROM t;", "select_all_sanity")
}

func TestSelectColumnList(t *testing.T) {
	assertCmd(t, "SELECT *", "selcollist_star")
	assertCmd(t, "SELECT id, t.id, t.name AS name", "selcollist_row_name")
	assertCmd(t, "SELECT `DATE`()", "selcollist_func_call")
}

func TestSelectTableList(t *testing.T) {
	assertCmd(t, "SELECT * FROM db.t", "seltablist_db_tab")
	assertCmd(t, "SELECT * FROM db.t AS dt", "seltablist_alias")
	assertCmd(t, "SELECT * FROM db.t dt", "seltablist_space_alias")
}

func TestSelectFromSubquery(t *testing.T) {
	assertCmd(t, "SELECT * FROM (SELECT a FROM t)", "subquery")
	assertCmd(t, "SELECT * FROM (SELECT a FROM t) at JOIN t", "subquery_join")
}

func TestSelectIndexedBy(t *testing.T) {
	assertCmd(t, "SELECT * FROM t INDEXED BY a", "indexed_by")
	assertCmd(t, "SELECT * FROM t NOT INDEXED", "not_indexed")
}

func TestSelectUsing(t *testing.T) {
	assertCmd(t, "SELECT * FROM t USING (a, b)", "using")
}

func TestSelectJoin(t *testing.T) {
	assertCmd(t, "SELECT * FROM db.t AS dt JOIN db.t AS td", "join")
	assertCmd(t, "SELECT * FROM db.t ldt, db.t rdt", "comma_join_2")
	assertCmd(t, "SELECT * FROM t1, t2, t3", "comma_join_3")
	assertCmd(t, "SELECT * FROM t1 LEFT OUTER JOIN t2 ON (t1.a = t2.a)", "left_outer_join_with_on")
}

func TestSelectGroupBy(t *testing.T) {
	assertCmd(t, "SELECT * FROM t GROUP BY t.a, t.b, 1 + t.c, func(t.d)", "group_by")
}

func TestSelectOrderBy(t *testing.T) {
	assertCmd(t, "SELECT * FROM t ORDER BY t.a", "order_by")
	assertCmd(t, "SELECT * FROM t ORDER BY t.b DESC", "order_by_desc")
	assertCmd(t, "SELECT * FROM t ORDER BY t.a DESC, t.b ASC, t.c", "order_by_list")
}

func TestSelectHaving(t *testing.T) {
	assertCmd(t, "SELECT * FROM t HAVING t.a", "having")
}

func TestSelectLimit(t *testing.T) {
	assertCmd(t, "SELECT * FROM t LIMIT 1", "limit")
	assertCmd(t, "SELECT * FROM t LIMIT 100, 25", "limit_comma")
	assertCmd(t, "SELECT * FROM t LIMIT 100 OFFSET 25", "limit_offset")
}

func TestSelectUnion(t *testing.T) {
	assertCmd(t, "SELECT * FROM t UNION SELECT * FROM u", "select_union_2")
	assertCmd(t, "SELECT * FROM t UNION SELECT * FROM u UNION ALL SELECT * FROM v", "select_union_3")
	assertCmd(t, "SELECT * FROM t UNION ALL SELECT * FROM u", "select_union_all")
	assertCmd(t, "SELECT * FROM t EXCEPT SELECT * FROM u", "select_except")
	assertCmd(t, "SELECT * FROM t INTERSECT SELECT * FROM u", "select_intersect")
}

func TestCreateTableSanity(t *testing.T) {
	assertCmd(t, "CREATE TABLE db.t (id INT, name VARCHAR(16))", "create_table_sanity")
	assertCmd(t, "CREATE TEMP TABLE db.t (id SMALLINT)", "create_table_temp")
	assertCmd(t, "CREATE TABLE IF NOT EXISTS db.t (id SMALLINT)", "create_table_if_not_exists")
}

func TestCreateTablePrimaryKey(t *testing.T) {
	assertCmd(t, "CREATE TABLE db.t (id INT PRIMARY KEY DESC ON CONFLICT IGNORE AUTOINCR, name VARCHAR(16))", "create_table_primary")
}

func TestCreateTableDefault(t *testing.T) {
	assertCmd(t, "CREATE TABLE db.t (id INT DEFAULT 0)", "create_table_default_0")
	assertCmd(t, "CREATE TABLE db.t (id INT DEFAULT (sin(100)))", "create_table_default_1")
	assertCmd(t, "CREATE TABLE db.t (name VARCHAR(16) DEFAULT john)", "create_table_default_2")
}

func TestCreateTableNullOrNotNull(t *testing.T) {
	assertCmd(t, "CREATE TABLE t (id INT NULL)", "create_table_null")
	assertCmd(t, "CREATE TABLE t (id INT NOT NULL)", "create_table_not_null")
	assertCmd(t, "CREATE TABLE t (id INT NOT NULL ON CONFLICT FAIL)", "create_table_not_null_on_conflict")
}

func TestCreateTableCheck(t *testing.T) {
	assertCmd(t, "CREATE TABLE t (id INT CHECK (name <> 'john'), name VARCHAR(16))", "create_table_check")
}

func TestCreateTableUnique(t *testing.T) {
	assertCmd(t, "CREATE TABLE t (id INT UNIQUE, name VARCHAR(16))", "create_table_unique")
	assertCmd(t, "CREATE TABLE t (id INT UNIQUE ON CONFLICT IGNORE, name VARCHAR(16))", "create_table_unique_on_conflict")
}

func TestCreateTableAsSelect(t *testing.T) {
	assertCmd(t, "CREATE TABLE t AS SELECT * FROM u", "create_table_as_select")
}

func TestCreateTablePostfix(t *testing.T) {
	assertCmd(t, "CREATE TABLE t (id INT, PRIMARY KEY (id DESC AUTOINCR))", "create_table_post_primary_key")
	assertCmd(t, "CREATE TABLE t (id INT, name VARCHAR(16), UNIQUE (id, name) ON CONFLICT IGNORE)", "create_table_post_unique")
	assertCmd(t, "CREATE TABLE t (id INT, name VARCHAR(16), CHECK (id <> name))", "create_table_post_check")
}

func TestCreateIndex(t *testing.T) {
	assertCmd(t, "CREATE INDEX db.idx ON t (id DESC, name ASC)", "create_index_sanity_0")
	assertCmd(t, "CREATE UNIQUE INDEX IF NOT EXISTS db.idx ON t (id, name)", "create_index_sanity_1")
}

func TestInsertSanity(t *testing.T) {
	assertCmd(t, "INSERT INTO db.t VALUES (1, 2, 'john')", "insert_sanity")
}

func TestInsertOrConf(t *testing.T) {
	assertCmd(t, "INSERT OR REPLACE INTO db.t VALUES(1)", "insert_or_replace")
	assertCmd(t, "REPLACE INTO db.t VALUES(1)", "insert_as_replace")
	assertCmd(t, "INSERT OR IGNORE INTO db.t VALUES(1)", "insert_or_ignore")
	assertCmd(t, "INSERT OR ABORT INTO db.t VALUES(1)", "insert_or_abort")
}

func TestInsertSource(t *testing.T) {
	assertCmd(t, "INSERT INTO db.t (id, name) VALUES(1, 'john')", "insert_collist_values")
	assertCmd(t, "INSERT INTO db.t (id, name) SELECT * FROM db.u", "insert_collist_select")
	assertCmd(t, "INSERT INTO db.t (id, name) DEFAULT VALUES", "insert_collist_default")
}

func TestUpdateSanity(t *testing.T) {
	assertCmd(t, "UPDATE db.t SET id = 1, name = 'john'", "update_sanity")
	assertCmd(t, "UPDATE db.t SET name = 'john' WHERE id = 1", "update_sanity_where")
}

func TestUpdateOrConf(t *testing.T) {
	assertCmd(t, "UPDATE OR REPLACE db.t SET id = 1, name = 'john'", "update_or_replace")
	assertCmd(t, "UPDATE OR IGNORE db.t SET name = 'john' WHERE id = 1", "update_or_ignore")
}

func TestUpdateOrderBy(t *testing.T) {
	assertCmd(t, "UPDATE db.t SET name = 'john' WHERE id = 1 ORDER BY id", "update_order_by")
}

func TestUpdateLimit(t *testing.T) {
	assertCmd(t, "UPDATE db.t SET name = 'john' WHERE id = 1 LIMIT 1", "update_limit")
	assertCmd(t, "UPDATE db.t SET name = 'john' WHERE id = 1 LIMIT 2, 1", "update_limit_comma")
	assertCmd(t, "UPDATE db.t SET name = 'john' WHERE id = 1 LIMIT 2 OFFSET 1", "update_limit_offset")
}

func TestDeleteSanity(t *testing.T) {
	assertCmd(t, "DELETE FROM db.t INDEXED BY a WHERE id = 1 ORDER BY b LIMIT 1", "delete_sanity")
}

const (
	kTestingPrefix = "testing/"
)

func assertExpr(t *testing.T, input, case_name string) {
	expr, err := ParseExpression(input)
	if err != nil {
		t.Fatalf("%s: %v", input, err)
	}
	assertAst(t, expr, case_name)
}

func assertCmd(t *testing.T, input, case_name string) {
	cmd, err := ParseCommand(input)
	if err != nil {
		t.Fatalf("%s: %v", input, err)
	}
	assertAst(t, cmd, case_name)
}

func assertAst(t *testing.T, root ast.Node, case_name string) {
	var out []byte
	var err error
	if out, err = json.Marshal(root); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err = json.Indent(&buf, out, "", "\t"); err != nil {
		t.Fatal(err)
	}

	var file *os.File
	jsn := buf.String()
	if file, err = os.Open(kTestingPrefix + case_name + ".json"); err != nil {
		t.Logf("[%s] not found: %v, now dump: \n%s", case_name, err, jsn)
		if file, err = os.Create(kTestingPrefix + case_name + ".json_"); err != nil {
			panic(err)
		}
		file.Write(buf.Bytes())
		file.Close()
		return
	}
	defer file.Close()

	buf.Reset()
	if _, err = buf.ReadFrom(file); err != nil {
		t.Fatal(err)
	}
	if buf.String() != jsn {
		t.Fatalf("[%s] Assert failed, Expected: ----------\n%s\nValue is: ----------\n%s",
			case_name, buf.String(), jsn)
	}
}

package opentick

import (
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Query(t *testing.T) {
	fdb.MustAPIVersion(FdbVersion)
	var db = fdb.MustOpenDefault()
	DropDatabase(db, "test")
	CreateDatabase(db, "test")
	ast, _ := Parse("create table test.test(a int, b int, b2 boolean, c int, d double, e bigint, primary key(a, b, b2, c))")
	err := CreateTable(db, "", ast.Create.Table)
	ast, _ = Parse("select a, b, b from test.test where a=1")
	_, err = resolveSelect(db, "", ast.Select)
	assert.Equal(t, "Duplicate column name b", err.Error())
	ast, _ = Parse("select * from test.test where a=1")
	stmt1, err1 := resolveSelect(db, "", ast.Select)
	assert.Equal(t, nil, err1)
	assert.Equal(t, 6, len(stmt1.Cols))
	ast, err1 = Parse("select a, b from test.test where a=2 and b=1 and b2=true limit -2")
	assert.Equal(t, nil, err1)
	stmt1, err1 = resolveSelect(db, "", ast.Select)
	assert.Equal(t, "b", stmt1.Cols[1].Name)
	assert.Equal(t, 2, len(stmt1.Cols))
	assert.Equal(t, 2, stmt1.Limit)
	assert.Equal(t, true, stmt1.Reverse)
	ast, _ = Parse("insert into test.test(a) values(1)")
	_, err = resolveInsert(db, "", ast.Insert)
	assert.Equal(t, "Some primary keys are missing: b, b2, c", err.Error())
	ast, _ = Parse("insert into test.test(a, a, c) values(1, 1, 1)")
	_, err = resolveInsert(db, "", ast.Insert)
	assert.Equal(t, "Duplicate column name a", err.Error())
	ast, _ = Parse("insert into test.test(a, a, c) values(1, 1)")
	_, err = resolveInsert(db, "", ast.Insert)
	assert.Equal(t, "Unmatched column names/values", err.Error())
	ast, _ = Parse("insert into test.test(a, b, b2, c, d) values(1, 1, false, 1, 1)")
	_, err = resolveInsert(db, "", ast.Insert)
	assert.Equal(t, nil, err)
	ast, _ = Parse("delete from test.test where d=1")
	_, err = resolveDelete(db, "", ast.Delete)
	assert.Equal(t, "Invalid column d in where clause, only primary key can be used", err.Error())
	ast, _ = Parse("delete from test.test where a<2.2")
	_, err = resolveDelete(db, "", ast.Delete)
	assert.Equal(t, "Invalid float64 value (2.2) for \"a\" of Int", err.Error())
	ast, _ = Parse("delete from test.test where b2<true")
	_, err = resolveDelete(db, "", ast.Delete)
	assert.Equal(t, "Invalid operator (<) for \"b2\" of type Boolean", err.Error())
	ast, _ = Parse("delete from test.test where a<2.2")
	_, err = resolveDelete(db, "", ast.Delete)
	assert.Equal(t, "Invalid float64 value (2.2) for \"a\" of Int", err.Error())
	ast, _ = Parse("delete from test.test where a=1 and a<1")
	_, err = resolveDelete(db, "", ast.Delete)
	assert.Equal(t, "a cannot be restricted by more than one relation if it includes an Equal", err.Error())
	ast, _ = Parse("delete from test.test where a<=1 and a<1")
	_, err = resolveDelete(db, "", ast.Delete)
	assert.Equal(t, "More than one restriction was found for the end bound on a", err.Error())
	ast, _ = Parse("delete from test.test where a>=1 and a>1")
	_, err = resolveDelete(db, "", ast.Delete)
	assert.Equal(t, "More than one restriction was found for the start bound on a", err.Error())
	ast, _ = Parse("delete from test.test where b=2")
	_, err = resolveDelete(db, "", ast.Delete)
	assert.Equal(t, "Cannot execute this query as it might involve data filtering and thus may have unpredictable performance", err.Error())
	ast, _ = Parse("delete from test.test where a<2 and b=2")
	_, err = resolveDelete(db, "", ast.Delete)
	assert.Equal(t, "Cannot execute this query as it might involve data filtering and thus may have unpredictable performance", err.Error())
	ast, _ = Parse("delete from test.test where a=2 and b=2 and b2=?")
	stmt2, err2 := resolveDelete(db, "", ast.Delete)
	assert.Equal(t, nil, err2)
	assert.Equal(t, 1, stmt2.NumPlaceholders)
	assert.Equal(t, 4, len(stmt2.Schema.Keys))
	Execute(db, "", "drop database test", nil)
	_, err = Execute(db, "", "drop database test", nil)
	assert.Equal(t, "Database test does not exist", err.Error())
	_, err = Execute(db, "", "create table test.test(a int, b int, b2 boolean, c int, d double, e bigint, primary key(a, b, b2, c))", nil)
	assert.Equal(t, "Database test does not exist", err.Error())
	Execute(db, "", "create database test", nil)
	_, err = Execute(db, "", "drop table test.test", nil)
	assert.Equal(t, "Table test.test does not exists", err.Error())
	_, err = Execute(db, "", "create table test.test(a int, b int, b2 boolean, c int, d double, e bigint, primary key(a, b, b2, c))", nil)
	assert.Equal(t, nil, err)
	_, err = Execute(db, "", "create table test.test(a int, b int, b2 boolean, c int, d double, e bigint, primary key(a, b, b2, c))", nil)
	assert.Equal(t, "Table test.test already exists", err.Error())
	_, err = Execute(db, "", "drop table test.test", nil)
	assert.Equal(t, nil, err)
	_, err = Execute(db, "", "create table test.test(a int, b int, b2 boolean, c int, d double, e bigint, primary key(a, b, b2, c))", nil)
	assert.Equal(t, nil, err)
	_, err = Execute(db, "", "insert into test.test(a, b, b2, c, d) values(1, 1, ?, ?, 1)", []interface{}{1})
	assert.Equal(t, "Expected 2 arguments, got 1", err.Error())
	_, err = Execute(db, "", "insert into test.test(a, b, b2, c, d) values(1, 1, ?, ?, 1)", []interface{}{1, 1})
	assert.Equal(t, "Invalid int value (1) for \"b2\" of Boolean", err.Error())
	_, err = Execute(db, "", "insert into test.test(a, b, b2, c, d) values(1, 1, ?, ?, 1)", []interface{}{true, true})
	assert.Equal(t, "Invalid bool value (true) for \"c\" of Int", err.Error())
	_, err = Execute(db, "", "insert into test.test(a, b2) values(1, ?)", []interface{}{true})
	assert.Equal(t, "Some primary keys are missing: b, c", err.Error())
	_, err = Execute(db, "", "select * from test.test where a=1 and b=2 and b2=? and c<?", []interface{}{true, 1})
	assert.Equal(t, nil, err)
	_, err = Execute(db, "", "delete from test.test where a=1 and b=2 and b2=? and c<?", []interface{}{true, 1})
	assert.Equal(t, nil, err)
	_, err = Execute(db, "", "insert into test.test(a, b, b2, c, d, e) values(2, 1, true, 42, 2.2, 102)", nil)
	assert.Equal(t, nil, err)
	_, err = Execute(db, "", "insert into test.test(a, b, b2, c, d, e) values(2, 1, true, 41, 2.2, 104)", nil)
	assert.Equal(t, nil, err)
	_, err = Execute(db, "", "insert into test.test(a, b, b2, c, d, e) values(2, 1, true, 39, 2.2, 105)", nil)
	assert.Equal(t, nil, err)
	res, err1 := Execute(db, "", "select * from test.test where a=2 and b=1 and b2=? and c=?", []interface{}{true, 42})
	assert.Equal(t, nil, err1)
	assert.Equal(t, []interface{}{int64(2), int64(1), true, int64(42), 2.2, int64(102)}, res[0])
	res, err1 = Execute(db, "", "select * from test.test where a=2 and b=1 and b2=true", nil)
	assert.Equal(t, nil, err1)
	assert.Equal(t, 3, len(res))
	assert.Equal(t, []interface{}{int64(2), int64(1), true, int64(39), 2.2, int64(105)}, res[0])
	res, err1 = Execute(db, "", "select * from test.test where a=2 and b=1 and b2=true limit -2", nil)
	assert.Equal(t, nil, err1)
	assert.Equal(t, 2, len(res))
	assert.Equal(t, []interface{}{int64(2), int64(1), true, int64(42), 2.2, int64(102)}, res[0])
	res, err1 = Execute(db, "", "select * from test.test where a=2 and b=1 and b2=true and c>39 and c<42", nil)
	assert.Equal(t, nil, err1)
	assert.Equal(t, 1, len(res))
	assert.Equal(t, []interface{}{int64(2), int64(1), true, int64(41), 2.2, int64(104)}, res[0])
	res, err1 = Execute(db, "", "select * from test.test where a=2 and b=1 and b2=true and c>=39 and c<=42", nil)
	assert.Equal(t, nil, err1)
	assert.Equal(t, 3, len(res))
	assert.Equal(t, int64(39), res[0][3])
	assert.Equal(t, int64(42), res[2][3])
	_, err = Execute(db, "", "delete from test.test where a=2 and b=1 and b2=true and c>=39 and c<=42", nil)
	assert.Equal(t, nil, err)
	res, err1 = Execute(db, "", "select * from test.test where a=2 and b=1 and b2=true", nil)
	assert.Equal(t, nil, err1)
	assert.Equal(t, 0, len(res))
	_, err = Execute(db, "", "create database test", nil)
	assert.Equal(t, "Database test already exists", err.Error())
	_, err = Execute(db, "", "create database if not exists test", nil)
	assert.Equal(t, nil, err)
	_, err = Execute(db, "", "create table if not exists test.test(x int)", nil)
	assert.Equal(t, nil, err)
	Execute(db, "", "drop table test.test", nil)
}

func Benchmark_resolveDelete(b *testing.B) {
	fdb.MustAPIVersion(FdbVersion)
	var db = fdb.MustOpenDefault()
	DropDatabase(db, "test")
	CreateDatabase(db, "test")
	ast, _ := Parse("create table test.test(a int, b int, c int, d double, e bigint, primary key(a, b, c))")
	CreateTable(db, "", ast.Create.Table)
	ast, _ = Parse("delete from test.test where a=2 and b=2 and c<?")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := resolveDelete(db, "", ast.Delete)
		if err != nil {
			b.Fatal(err)
		}
	}
	Execute(db, "", "drop table test.test", nil)
}

func Benchmark_resolveInsert(b *testing.B) {
	fdb.MustAPIVersion(FdbVersion)
	var db = fdb.MustOpenDefault()
	DropDatabase(db, "test")
	CreateDatabase(db, "test")
	ast, _ := Parse("create table test.test(a int, b int, c int, d double, e bigint, primary key(a, b, c))")
	CreateTable(db, "", ast.Create.Table)
	ast, _ = Parse("insert into test.test(a, b, c, d) values(1, 2, ?, 1.2)")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := resolveInsert(db, "", ast.Insert)
		if err != nil {
			b.Fatal(err)
		}
	}
	Execute(db, "", "drop table test.test", nil)
}

func Benchmark_resolveSelect(b *testing.B) {
	fdb.MustAPIVersion(FdbVersion)
	var db = fdb.MustOpenDefault()
	DropDatabase(db, "test")
	CreateDatabase(db, "test")
	ast, _ := Parse("create table test.test(a int, b int, c int, d double, e bigint, primary key(a, b, c))")
	CreateTable(db, "", ast.Create.Table)
	ast, _ = Parse("select a, b, c, d from test.test where a=1 and b=2 and c<2 and c>1")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := resolveSelect(db, "", ast.Select)
		if err != nil {
			b.Fatal(err)
		}
	}
	Execute(db, "", "drop table test.test", nil)
}
